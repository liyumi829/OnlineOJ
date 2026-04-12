package manager

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// isTimeoutError 判断错误是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if st, ok := status.FromError(err); ok && st.Code() == codes.DeadlineExceeded {
		return true
	}

	return false
}
