package framer

import (
	"xlib/log"
	"xmodbustcp/define"
)

//ModbusTCP/RTU报文接口
type Framer interface {
	Bytes() []byte
	Copy() Framer
	GetData() []byte
	GetFunction() uint8
	SetException(exception *Exception)
	SetData(data []byte)
}

//0X03功能码回包数据设置，从Value -> Bytes ->属于协议层
func SetResponseWith0x03(frame Framer,  number uint16, values []uint16) {
	data := make([]byte, 1+len(values)*2)
	data[0] = uint8(number)//0x00 00
	copy(data[1:], define.Uint16ToBytes(values))
	log.Debugf("Values: %02x\n After Translate Data: %02x",values,data)
	frame.SetData(data)
}