package common

import (
	"strconv"

	"github.com/google/uuid"
)

const (
	NextPollAfterMS0 = 0   // 准备就绪/结束 立即轮询
	NextPollAfterMS1 = 200 // 正在判题
	NextPollAfterMS2 = 400 // 正在排队
	NextPollAfterMS3 = 800 // 其他
)

// str 是你要转的字符串，比如 "1001"
func StrToUint64(str string) (uint64, error) {
	num, err := strconv.ParseUint(str, 10, 64)
	return num, err
}

func Uuid() string {
	return uuid.New().String()
}
