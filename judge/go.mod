module online-oj/judge

go 1.25.5

require (
	go.uber.org/zap v1.27.1
	online-oj/pkg v0.0.0
)

replace (
	online-oj/api => ../api
	online-oj/pkg => ../pkg
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
