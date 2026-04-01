package run

import (
	"context"
	"errors"
	"fmt"
	"online-oj/judge/internal/common"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type RunResult struct { // 运行结果
	ExitCode    int           // 退出码：程序正常退出时的返回码
	Status      string        // 状态：OK(正常)/TLE(超时)/MLE(内存超限)/OLE(输出超限)/RE(运行时错误)
	Stdout      string        // 标准输出：程序打印到标准输出的内容，通常是运行结果
	Stderr      string        // 标准错误：程序打印到标准错误的内容，如错误信息、段错误详情
	TimeReal    time.Duration // 真实花费的时间：程序从启动到结束的实际耗时
	MemoKiBReal int64         // 真实花费的内存大小：程序运行过程中占用的最大内存（KB）
}

type Runner struct {
	Bin          string        // 可执行文件路径
	CpuLimit     time.Duration // CPU资源限制
	MemoKiBLimit int64         // 内存资源限制
}

func (r *Runner) RunSandboxed(ctx context.Context) (*RunResult, error) {
	zap.L().Info(r.Bin)
	lastIndex := strings.LastIndex(r.Bin, "/") // 截取可执行程序的分隔符
	dir := r.Bin[:lastIndex]                   // 获取创建的临时目录
	zap.L().Debug("path", zap.String("dir", dir))
	defer func() { os.RemoveAll(dir) }() // 删除该临时目录
	// 准备运行程序
	ctxRun, cancle := context.WithTimeout(ctx, r.CpuLimit) // 使用 WithTimeout 创建带超时的子上下文，超时时间为 CPU 限制时间
	defer cancle()                                         // 确保资源释放
	cmd := exec.CommandContext(ctxRun, r.Bin)              // 通过绝对路径找到可执行程序

	stdout := &common.OutputBuffer{} // 完成重定向
	stderr := &common.OutputBuffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	// 设置命名空间和用户映射 由于没有权限，更优的做法是利用docker
	// cmd.SysProcAttr = &syscall.SysProcAttr{
	// 	// Cloneflags 设置命名空间隔离标志
	// 	Cloneflags: syscall.CLONE_NEWPID | // PID命名空间隔离，子进程有独立PID
	// 		syscall.CLONE_NEWNET | // 网络命名空间隔离，独立网络栈
	// 		syscall.CLONE_NEWIPC | // IPC命名空间隔离，独立进程间通信
	// 		syscall.CLONE_NEWUTS, // UTS命名空间隔离，独立主机名和域名
	// 	Setpgid: true,
	// }
	// 限制使用资源 --> 根据实际使用和限制进行比较
	start := time.Now()                 // 用于记录实际运行的时间
	if err := cmd.Start(); err != nil { // 启动子进程 这里不用Run是因为要区别是正常退出还是时钟到期
		zap.L().Error("execute fail", zap.String("error", err.Error()))
		return nil, err
	}
	zap.L().Debug("execute process begin")
	waitCh := make(chan error, 1) // 防止阻塞
	go func() {
		zap.L().Debug("execute success", zap.Any("value", cmd), zap.Any("chan", len(waitCh)))
		waitCh <- cmd.Wait() // 等待子进程执行任务完成
		close(waitCh)        // 发送者对通道进行关闭
	}() // 等待调用结果
	res := &RunResult{Status: "OK"}
	select {
	case <-ctxRun.Done(): // 执行程序超时
		zap.L().Debug("execute Done Timeout")
		if cmd.Process != nil { // 如果进程对象存在，则终止进程
			if err := syscall.Kill(cmd.Process.Pid, syscall.SIGKILL); err != nil { // 发送9号信号
				if !errors.Is(err, syscall.ESRCH) { // 如果不是"进程不存在"错误，则记录错误日志
					zap.L().Error("kill process group error", zap.String("error", err.Error()))
				}
			}
		}
		res.Status = "TLE"        //  超时说明
		res.TimeReal = r.CpuLimit // 超时
		res.ExitCode = -1
		<-waitCh // 为了能够获取正常的ProcessState状态 确保 SysUsage() 不会发生nil指针引用
		zap.L().Debug("get processstate")
	case err := <-waitCh: // 执行程序正常退出
		zap.L().Debug("execute bin end")
		res.TimeReal = time.Since(start) // 获取执行的时间
		if err != nil {
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				// 1. 记录退出码
				res.ExitCode = exitErr.ExitCode()

				// 2. 只有被信号杀死才标记 RE（段错误、panic、崩溃）
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					if status.Signaled() {
						// 运行时错误：段错误、panic、被杀死
						res.Status = fmt.Sprintf("RE:SIG%s", status.Signal())
					} else { // 可能是自定义退出码
						res.Status = "OK"
					}
				}
			}
		}
	}
	// 下面获取 memory、stdout、stderr
	if cmd.ProcessState == nil {
		zap.L().Error("Get a Process State faile...can't get a memory information")
	} else {
		if rusage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok { // 获取该程序资源使用
			// 获取Process的使用情况
			res.MemoKiBReal = rusage.Maxrss
			zap.L().Debug("real memory", zap.Any("real size(KB)", res.MemoKiBReal), zap.Any("limit size(KB)", r.MemoKiBLimit))
			if res.MemoKiBReal > r.MemoKiBLimit {
				res.Status = "MLE" // 超过内存限制
				res.ExitCode = -1  // 重置退出码
			}
		}
	}
	// 退出说明数据已经写好了
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	if len(stdout.Buffer) >= common.MaxOutPut || len(stderr.Buffer) >= common.MaxOutPut {
		res.Status = "OLE"
	}
	if res.Status == "TLE" {
		res.Stderr = "Time Limit Exceeded" // 记录错误信息
	}
	if res.Status == "MLE" {
		res.Stderr = "Memory Limit Exceeded"
	}
	zap.L().Info("return a RunResult")
	return res, nil
}
