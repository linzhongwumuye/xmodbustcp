package framer

import (
	"encoding/binary"
	"fmt"
)
//ModbusTcp报文处理
type TCPFrame struct {
	TransactionIdentifier uint16
	ProtocolIdentifier    uint16
	Length                uint16
	Device                uint8
	Function              uint8
	Data                  []byte
}

//New
func NewTCPFrame(bytes []byte) (*TCPFrame,error) {
	if len(bytes) < 9 {//校验码2 + 标识2 + 长度2 + 设备地址1 + 功能码1  + 寄存器2 + 个数2
		return nil,fmt.Errorf("报文长度小于9不合格")
	}

	tcpFrame := &TCPFrame{
		TransactionIdentifier: 		binary.BigEndian.Uint16(bytes[0:2]),
		ProtocolIdentifier: 		binary.BigEndian.Uint16(bytes[2:4]),
		Length: 					binary.BigEndian.Uint16(bytes[4:6]),
		Device: 					uint8(bytes[6]),
		Function: 					uint8(bytes[7]),
		Data: 						bytes[8:],
	}

	//校验长度
	if int(tcpFrame.Length) != len(tcpFrame.Data)+2 {
		return nil,fmt.Errorf("tcpFrame长度错误,length",tcpFrame.Length,"Datalen",len(tcpFrame.Data))
	}

	return tcpFrame,nil
}

//拷贝
func (this *TCPFrame)Copy() Framer {
	frame := *this
	return &frame
}

//组协议包
func (this *TCPFrame)Bytes() []byte{
	bytes := make([]byte, 8)

	binary.BigEndian.PutUint16(bytes[0:2], this.TransactionIdentifier)
	binary.BigEndian.PutUint16(bytes[2:4], this.ProtocolIdentifier)
	binary.BigEndian.PutUint16(bytes[4:6], uint16(2+len(this.Data)))
	bytes[6] = this.Device
	bytes[7] = this.Function
	bytes = append(bytes, this.Data...)

	return bytes
}

//获取数据
func (this *TCPFrame)GetData() []byte{
	return this.Data
}

//获取功能码
func (this *TCPFrame)GetFunction() uint8{
	return this.Function
}

//设置Exception
func (this *TCPFrame)SetException(exception *Exception){
	this.Function = this.Function | 0x80
	this.Data = []byte{byte(*exception)}
	this.setlenth()
}

//设置数据
func (this *TCPFrame)SetData(data []byte){
	this.Data = data
	this.setlenth()
}

//更新FrameData长度
func (this *TCPFrame)setlenth(){
	this.Length = uint16(len(this.Data) + 2)
}