package business

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/mbserver"
	"testing"
	"xmodbustcp/framer"
)

func getVal()[]string{

	src := GetMapper()
	allval := make([]string,0,3)
	//错误值
	for i := 0; i < len(src); i++ {
		allval = append(allval,src[uint16(i)])
	}

	//连接Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     	"127.0.0.1:6379",
		Password: 	"",
		DB:       	0,
		PoolSize: 	10,
	})
	defer redisClient.Close()
	_, e := redisClient.Ping().Result()
	if e!= nil {
		fmt.Println("Redis连接出错.....")
		return allval
	}
	fmt.Println("Redis连接成功")
	fmt.Println("allval",allval)
	res ,err := redisClient.MGet(allval...).Result()
	if err != nil {
		fmt.Println("MgET",err)
		return allval
	}
	fmt.Println("Res",res)
	allval = allval[0:0]
	for _,val := range res {
		if v,ok :=val.(string); ok {
			allval = append(allval,v)
		}
	}
	return allval
}

func Test_getRids(t *testing.T){
	numRegs := 99;
	endRegister := 99
	register := 0

	var Rids []string
	mapper := GetMapper()
	for i := register; i < numRegs ; i ++ {
		Rids = append(Rids,mapper[uint16(i)])
	}

	fmt.Println(Rids)

	if len(Rids) != endRegister - register {
		fmt.Println("出问题了")
	}
	return
}

func TestGetMapper(t *testing.T) {
	//连接Redis
	fmt.Println(GetMapper())
}

func TestGJValTOModbus(t *testing.T) {
	allval := getVal()

	for index,sigleval :=range GJValTOModbus(allval){
		fmt.Println(allval[index],sigleval)
	}
}

func TestSetDataWithRegisterAndNumberAndValues(t *testing.T) {
	//模拟报文
	packet := []byte{0x19,0x1B,0x00,0x00,0x00,0x06,0x01,0x03,0x00,0x00,0x00,0x63}

	//模拟数据
	allval := getVal()
	fmt.Println(allval)
	frame,err := framer.NewTCPFrame(packet)
	if err != nil {
		fmt.Println("分配TCPFrame错误",err)
		return
	}
	framer.SetResponseWith0x03(frame,2,GJValTOModbus([]string{allval[0],allval[1]}))
	fmt.Printf("数据:\n寄存器: %02x\t 数量:%02x\t 长度:%02x\n",frame.Data[:2],frame.Data[2:4],frame.Data[4])
	for i := 0; i < int(frame.Data[4]) ; i+=2{
		fmt.Println("数值 ",mbserver.BytesToUint16(frame.Data[i+5:i+7]))
	}
	fmt.Println("长度",frame.Length,"设备",frame.Device,"标识",frame.ProtocolIdentifier)
}