package internal

import "errors"

const maxOutPut = 128 * 1024 // 最大容纳的缓冲区：128KB

type outputBuffer struct {
	buffer []byte
}

func (ob *outputBuffer) Write(b []byte) (int, error) {
	if len(ob.buffer)+len(b) > maxOutPut { // 超过缓冲区
		return len(b), errors.New("output too large")
	}
	ob.buffer = append(ob.buffer, b...)
	return len(b), nil
}

func (ob *outputBuffer) String() string {
	return string(ob.buffer)
}
