package compile

import (
	"context"
	"fmt"
	"online-oj/judge/internal/common"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go.uber.org/zap"
)

// 实现代码的编译服务
// 实现两种程序的编译 C++/Go

type Type int32

const (
	UnKnownType Type = iota
	GoType
	CppType
)

// 编译者
type Compiler struct {
	CodeType Type   // 代码的类型
	Code     string // 完整代码
}

// 编译结果
type CompileResult struct {
	BinPath string // 编译生成的文件路径（使用完成之后需要删除）
	Stderr  string // 标准错误 -- 空标识没有错误
	Status  string // 状态 OK/CE
}

// Compile 根据代码类型编译Copiler的代码
// 参数:
//
//	ctx 上下文内容。外部可以用于控制Compile函数的退出
//
// 返回值: 编译的文件名/返回空说明编译失败
func (c *Compiler) Compile(ctx context.Context, storagePath string) (*CompileResult, error) {
	switch c.CodeType {
	case GoType:
		return c.goCompile(ctx, storagePath)
	case CppType:
		return c.cppCompile(ctx, storagePath)
	default:
		zap.L().Error("unknown Code type")
		return nil, fmt.Errorf("unknown Code type")
	}
}

// 编译
func (c *Compiler) goCompile(ctx context.Context, storagePath string) (*CompileResult, error) {
	// 创建临时目录和文件，并且进行写入
	tempDir, err := os.MkdirTemp(storagePath, "compile-go-") // 创建了临时文件还没有删除 运行完成bin之后删除保存 tempDir
	if err != nil {
		zap.L().Error("create compile-go-tempdir fail... ", zap.String("err", err.Error()))
		return nil, err
	}
	var compileSuccess bool
	defer func() {
		if !compileSuccess {
			os.RemoveAll(tempDir)
		}
	}()
	src := filepath.Join(tempDir, "main.go") // 源文件
	bin := filepath.Join(tempDir, "main")    // 可执行程序
	if err := os.WriteFile(src, []byte(c.Code), 0644); err != nil {
		zap.L().Error("write file fail...", zap.String("err", err.Error()), zap.String("src", src))
		return nil, err
	}
	// 写入文件成功
	ctxCompile, cancle := context.WithTimeout(ctx, 15*time.Second) // 获取下文取消
	defer cancle()
	cmd := exec.CommandContext(ctxCompile, "go", "build", "-trimpath", "-ldflags", "-s -w", "-o", bin, src) // 初始化子进程
	stderr := &common.OutputBuffer{}                                                                        // 缓冲区，用于重定向
	cmd.Stderr = stderr                                                                                     // 重定向标准错误
	err = cmd.Run()                                                                                         // 创建子进程并且运行，父进程在同步等待子进程完成
	if err != nil {                                                                                         // 编译发生错误
		zap.L().Info("compile fail...", zap.String("src", src), zap.String("error", err.Error()))
		return &CompileResult{Stderr: stderr.String(), Status: "CE"}, nil // 发生错误
	}
	// 编译成功
	//zap.L().Debug("compile success...", zap.String("src", src), zap.String("bin", bin))
	compileSuccess = true
	return &CompileResult{BinPath: bin, Status: "OK"}, nil
}

func (c *Compiler) cppCompile(ctx context.Context, storagePath string) (*CompileResult, error) {
	tempDir, err := os.MkdirTemp(storagePath, "compile-cpp-")
	if err != nil {
		zap.L().Error("create compile-cpp-tempdir fail... ", zap.String("err", err.Error()))
		return nil, err
	}
	var compileSuccess bool
	defer func() {
		if !compileSuccess {
			os.RemoveAll(tempDir)
		}
	}()
	src := filepath.Join(tempDir, "main.cpp") // 源文件
	bin := filepath.Join(tempDir, "main")     // 可执行程序
	if err := os.WriteFile(src, []byte(c.Code), 0644); err != nil {
		zap.L().Error("write file fail...", zap.String("err", err.Error()), zap.String("src", src))
		return nil, err
	}
	// 写入文件成功
	ctxCompile, cancle := context.WithTimeout(ctx, 15*time.Second) // 获取一个子Context运行命令行
	defer cancle()
	// "g++", "g++", "-o", PathSpliceUtil::Exe(filename).c_str(),PathSpliceUtil::Src(filename).c_str(), "-D", "COMPILE_ONLINE" , "-std=c++17"
	cmd := exec.CommandContext(ctxCompile, "g++", "-o", bin, src, "-std=c++11") // 初始化子进程
	stderr := &common.OutputBuffer{}                                            // 缓冲区，用于重定向
	cmd.Stderr = stderr                                                         // 重定向标准错误
	err = cmd.Run()                                                             // 创建子进程并且运行
	if err != nil {                                                             // 编译发生错误
		zap.L().Info("compile fail...", zap.String("src", src), zap.String("error", err.Error()))
		return &CompileResult{Stderr: stderr.String(), Status: "CE"}, nil // 发生错误
	}
	// 编译成功
	//zap.L().Debug("compile success...", zap.String("src", src), zap.String("bin", bin))
	compileSuccess = true
	return &CompileResult{BinPath: bin, Status: "OK"}, nil
}
