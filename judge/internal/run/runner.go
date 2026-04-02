package run

import (
	"context"
	"errors"
	"online-oj/api/proto/judge"
	"online-oj/judge/internal/common"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type RunResult struct { // 运行结果
	Status      string              // 状态：AC(答案正确)/WA(错误)/TLE(超时)/MLE(内存超限)/OLE(输出超限)/RE(运行时错误)
	Stdout      string              // 标准输出：程序打印到标准输出的内容，通常是运行结果
	Stderr      string              // 标准错误：程序打印到标准错误的内容，如错误信息、段错误详情
	TimeReal    time.Duration       // 真实花费的时间：程序从启动到结束的实际耗时
	MemoKiBReal int64               // 真实花费的内存大小：程序运行过程中占用的最大内存（KB）
	CaseRusults []*judge.CaseResult // 测试用例的结果
}

type Runner struct {
	Bin          string            // 可执行文件路径
	CpuLimit     time.Duration     // CPU资源限制
	MemoKiBLimit int64             // 内存资源限制
	TestCases    []*judge.TestCase // 测试用例
}

var priority = map[string]int{"AC": 0, "OLE": 1, "WA": 2, "RE": 3, "MLE": 4, "TLE": 5}

func (r *Runner) RunSandboxed(ctx context.Context) (*RunResult, error) {
	zap.L().Info(r.Bin)
	lastIndex := strings.LastIndex(r.Bin, "/") // 截取可执行程序的分隔符
	dir := r.Bin[:lastIndex]                   // 获取创建的临时目录
	zap.L().Debug("path", zap.String("dir", dir))
	defer func() { os.RemoveAll(dir) }() // 删除该临时目录
	// 准备运行程序
	res := &RunResult{
		Status:      "AC",
		CaseRusults: make([]*judge.CaseResult, 0, len(r.TestCases)),
	}

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
	// 利用fork再次创建子进程，让子进程去执行测试用例

	for idx, testCase := range r.TestCases { // 开始执行每一个测试用例
		zap.L().Debug("", zap.Int("case", idx+1),
			zap.String("input", testCase.Input),
			zap.String("output", testCase.Output))
		caseResult := &judge.CaseResult{}
		ctxRun, cancel := context.WithTimeout(ctx, r.CpuLimit) // 限制每一个测试用例的执行时间
		defer cancel()                                         // 确保无论何种路径都 cancel，防泄漏
		cmd := exec.CommandContext(ctxRun, r.Bin)              // 通过绝对路径找到可执行程序
		cmd.Stdin = strings.NewReader(testCase.Input)          // 设置标准输入(0拷贝)
		stdout := &common.OutputBuffer{}                       // 完成重定向
		stderr := &common.OutputBuffer{}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		// 执行测试用例
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

		// 全局调用状态 对于单个测试用例
		caseStatus := "AC"         // 调用结果
		var timeReal time.Duration // 单个测试用例使用时间
		var memoRealKB int64       // 单个测试的内存使用情况
		select {
		case <-ctxRun.Done(): // 执行程序超时
			zap.L().Debug("execute Done Timeout")
			if cmd.Process != nil { // 发送9号信号杀死进程组
				syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
			}
			<-waitCh              // 为了能够获取正常的ProcessState状态 确保 SysUsage() 不会发生nil指针引用
			caseStatus = "TLE"    //  这个测试用例超时限制时间
			timeReal = r.CpuLimit // 设置测试用例的超时时间
			zap.L().Debug("get processstate")
		case err := <-waitCh: // 执行程序正常退出
			zap.L().Debug("execute bin end")
			timeReal = time.Since(start) // 获取执行的时间
			if err != nil {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) {
					// 被信号杀死才标记 RE（段错误、panic、崩溃）
					// 运行时错误：段错误、panic、被杀死
					if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() {
						caseStatus = "RE"
					} // 记录单次测试用例的状态
				}
			}
		}
		// 先获取内存的使用
		if cmd.ProcessState != nil {
			if rusage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
				memoRealKB = rusage.Maxrss
			} else {
				zap.L().Error("SysUsage not *syscall.Rusage", zap.Int("case", idx+1))
				memoRealKB = 0
			}
		} else {
			zap.L().Error("ProcessState is nil after wait", zap.Int("case", idx+1))
			memoRealKB = 0
		}

		if memoRealKB > r.MemoKiBLimit && caseStatus == "AC" {
			caseStatus = "MLE"
		}

		// 检查输出限制
		if caseStatus == "AC" && (len(stdout.Buffer) >= common.MaxOutPut || len(stderr.Buffer) >= common.MaxOutPut) {
			caseStatus = "OLE" // 超过输出限制
		}
		// 判断是否通过测试
		outStr := stdout.String()    // 拿到标准输出
		stdErrStr := stderr.String() // 拿到标准错误
		outputMatch := strings.TrimSpace(outStr) == strings.TrimSpace(testCase.Output)
		passed := caseStatus == "AC" && outputMatch
		if !passed && caseStatus == "AC" && !outputMatch {
			caseStatus = "WA"
		}
		// 填充结果
		caseResult.Passed = passed
		caseResult.Time = timeReal.Nanoseconds()
		caseResult.Memory = memoRealKB
		caseResult.Status = caseStatus
		caseResult.Output = outStr
		res.CaseRusults = append(res.CaseRusults, caseResult) // 添加结果到集合
		zap.L().Debug("Case Rusults", zap.Int("len", len(res.CaseRusults)))
		if caseStatus != "AC" { // 更新全局状态（取最差情况）
			// 保留最差状态：TLE > MLE > RE > WA > OLE > AC
			if priority[caseStatus] > priority[res.Status] {
				res.Status = caseStatus
				if res.Stderr == "" && stdErrStr != "" {
					res.Stderr = stdErrStr
				}
			}
		}
		// 更新全局资源统计（取最大值）
		if timeReal > res.TimeReal {
			res.TimeReal = timeReal
		}
		if memoRealKB > res.MemoKiBReal {
			res.MemoKiBReal = memoRealKB
		}
		zap.L().Debug("case finished",
			zap.Int("case", idx+1),
			zap.String("status", caseStatus),
			zap.Int64("time_ns", timeReal.Nanoseconds()),
			zap.Int64("memory_kb", memoRealKB),
			zap.String("Output", outStr))
	}
	zap.L().Info("sandbox run finished",
		zap.String("final_status", res.Status),
		zap.Int("total_cases", len(r.TestCases)))
	return res, nil
}
