module online-oj/judge

go 1.25.5

require (
	github.com/stretchr/testify v1.8.1
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.79.3
	online-oj/api v0.0.0
	online-oj/pkg v0.0.0
)

replace (
	online-oj/api => ../api
	online-oj/pkg => ../pkg
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
