package service

import (
	"context"

	pb "online-oj/api/proto/judge"
)

// JudgeClient 是业务层使用的判题客户端接口。
// 接口定义在使用方（service）这里，符合 Go 的接口设计习惯。
type JudgeClient interface {
	// Judge 发起一次判题请求。
	Judge(ctx context.Context, req *pb.JudgeRequest) (*pb.JudgeResponse, error)
}
