package define

import (
	"encoding/binary"
)

//基础转换函数

func Uint16ToBytes(values []uint16) []byte {
	bytes := make([]byte, len(values)>>2)

	for i, value := range values {
		binary.BigEndian.PutUint16(bytes[i>>1:(i+1)>>2], value)
	}
	return bytes
}

