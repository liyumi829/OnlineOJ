package common

import (
	"strconv"

	"github.com/google/uuid"
)

// str 是你要转的字符串，比如 "1001"
func StrToUint64(str string) (uint64, error) {
	num, err := strconv.ParseUint(str, 10, 64)
	return num, err
}

func Uuid() string {
	return uuid.New().String()
}
