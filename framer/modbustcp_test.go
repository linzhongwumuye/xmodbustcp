package framer_test

import (
	"fmt"
	"net"
	"testing"
	"time"
)

var (
	req1 = []byte{
		0x19,0xB2,//检验码
		0x00,0x00,//协议标志
		0x00,0x06,//数据长度
		0x06,//Slave Addr
		0x03,//功能码
		0x00,0x27,//寄存器起始地址
		0x00,0x02}//寄存器个数
)


func TestModbusTcp(t *testing.T) {
	conn, err := net.Dial("tcp4", "127.0.0.1:17789")
	if err != nil {
		fmt.Println("连接服务端错误",err)
		return
	}

	var readBuf []byte
	for {
		sendCount,err := conn.Write(req1)
		if err != nil {
			fmt.Println("写入socket错误",sendCount)
			return
		}
		time.Sleep(5*time.Second)
		recvCount, err := conn.Read(readBuf)
		if err != nil {
			fmt.Println("接收socket错误",recvCount)
		}
		fmt.Println(recvCount)
	}
}