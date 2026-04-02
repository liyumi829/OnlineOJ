package execute

import (
	"context"
	"fmt"
	"net"
	"online-oj/api/proto/judge"
	"online-oj/judge/internal/compile"
	"online-oj/judge/internal/run"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// NewServer 创建CompileAndRun服务实例
func NewServer(storagePath string) (judge.JudgeServiceServer, error) {
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			return nil, fmt.Errorf("create storage path failed: %w", err)
		}
	}
	return &server{storagePath: storagePath}, nil
}

// StartGRPCServer 启动gRPC服务端
func StartGRPCServer(addr string, storagePath string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	srv, err := NewServer(storagePath)
	if err != nil {
		return err
	}
	judge.RegisterJudgeServiceServer(s, srv)

	// 优雅关闭（监听系统信号）
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		zap.L().Info("shutting down gRPC server...")
		s.GracefulStop()
	}()

	zap.L().Info("gRPC server started", zap.String("addr", addr))
	return s.Serve(lis)
}

// 实现grpc调用
type server struct {
	judge.UnimplementedJudgeServiceServer
	storagePath string // 存储路径
}

// ExecuteCode 执行用户代码
func (s *server) Judge(ctx context.Context, req *judge.JudgeRequest) (*judge.JudgeResponse, error) {
	zap.L().Debug("receive a execute code request...", zap.Bool("code not empty", req.GetCode() != ""))

	codeType := getCodeType(req.GetCodeType()) // 获取类型
	if codeType == compile.UnKnownType {
		zap.L().Info("user submit a unknown code type", zap.Int32("type code", req.GetCodeType()))
		return nil, fmt.Errorf("user submit a unknown code type, type code: %d", req.GetCodeType())
	}
	code := req.GetCode()          // 获取代码
	compiler := &compile.Compiler{ // 创建compiler实例
		CodeType: codeType,
		Code:     code,
	}
	cRes, err := compiler.Compile(ctx, s.storagePath) // 编译调用
	if err != nil {
		// 编译内部发生问题，不应该被用户知晓
		zap.L().Error("A compilation error occurred", zap.String("error", err.Error()))
		return nil, err
	}
	rsp := &judge.JudgeResponse{}
	// 正常读取
	if cRes.Status == "CE" { // 发生编译错误
		// 发生编译错误
		rsp.Status = "CE"
		rsp.Stderr = cRes.Stderr
		zap.L().Debug("A compilation error occurred in the code.")
		return rsp, nil
	}
	// 编译成功 下面开始执行代码
	runner := &run.Runner{
		Bin:          cRes.BinPath,
		CpuLimit:     time.Duration(req.CpuLimit),
		MemoKiBLimit: req.MemoLimit,
		TestCases:    req.GetTestCases(),
	}
	rRes, err := runner.RunSandboxed(ctx)
	if err != nil {
		zap.L().Error("An internal error occurred in the program while executing the code.", zap.String("error", err.Error()))
		return nil, err
	}
	// 正常结束
	rsp.Status = rRes.Status
	rsp.Stderr = rRes.Stderr
	rsp.Stdout = summarizeCaseStatus(rRes.CaseRusults)
	zap.L().Debug("", zap.String("stdout", rsp.Stdout))
	rsp.Time = rRes.TimeReal.Nanoseconds()
	rsp.Memory = rRes.MemoKiBReal
	rsp.Results = rRes.CaseRusults
	return rsp, nil
}

func getCodeType(value int32) compile.Type {
	switch value {
	case 1:
		return compile.GoType
	case 2:
		return compile.CppType
	default:
	}
	return compile.UnKnownType
}
