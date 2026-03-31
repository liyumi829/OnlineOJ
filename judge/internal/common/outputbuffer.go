package common

import "errors"

const MaxOutPut = 128 * 1024 // 最大容纳的缓冲区：128KB

type OutputBuffer struct {
	Buffer []byte
}

func (ob *OutputBuffer) Write(b []byte) (int, error) {
	if len(ob.Buffer)+len(b) > MaxOutPut { // 超过缓冲区
		return len(b), errors.New("output too large")
	}
	ob.Buffer = append(ob.Buffer, b...)
	return len(b), nil
}

func (ob *OutputBuffer) String() string {
	return string(ob.Buffer)
}
