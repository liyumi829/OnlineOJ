module online-oj/judge

go 1.25.5

require (
	github.com/redis/go-redis/v9 v9.18.0
	go.uber.org/zap v1.27.1
	golang.org/x/sync v0.20.0
	google.golang.org/grpc v1.79.3
	online-oj/api v0.0.0
	online-oj/pkg v0.0.0
)

replace (
	online-oj/api => ../api
	online-oj/pkg => ../pkg
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
