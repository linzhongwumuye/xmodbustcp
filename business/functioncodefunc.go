package business

import (
	"encoding/binary"
	"strconv"
	"strings"
	"xlib/log"
	"xmodbustcp/framer"
	"xmodbustcp/modbusserver"
)

/*
	此处为对应的一系列函数：
	需要何种功能码做何种处理时，只需要在Server的函数映射表中进行初始化赋值即可
	例如新加坡CMI对接华为 只有0X03功能码，就仅实现读取数据的功能
	保留此方式，当遇到下发控制量等值时可以实现其他函数
*/

//0x03:读寄存器,然后回复相关的值
func (this *XSvrer)ReadHoldingRegisters(s *modbusserver.Server,frame framer.Framer) ([]byte,*framer.Exception){

	register, numRegs, endRegister := registerAddressAndNumber(frame)
	if endRegister > 65535 {
		log.Error("请求中的寄存器地址最大值错误")
		return nil,&framer.IllegalDataAddress
	}

	Rids, framer2 := this.getRidsFromRegs(register, numRegs, endRegister)
	if framer2 != &framer.Success {
		return []byte{},framer2
	}

	values, err := this.datasrc.GetData(Rids)
	if err != nil {
		return []byte{},&framer.MemoryParityError
	}

	convertValues := GJValTOModbus(values)
	log.Debugf("请求寄存器起始地址: %02x\t 寄存器数量: %02x\t",register,numRegs)
	framer.SetResponseWith0x03(frame,uint16(numRegs),convertValues)

	return frame.GetData(),&framer.Success
}

//TODO:: 负数、>65536的数 需要协商后进行处理
// 平台值转换为Modbus数据
func GJValTOModbus(srcval []string)(dstval []uint16){
	for _,val := range srcval {
		if strings.ContainsAny(val,"."){
			val = strings.Replace(val, ".", "", 1)
		}
		i, e := strconv.Atoi(val)
		if e != nil {
			log.Info("平台值转换过程中出现错误",e,"原始值为：",val)
			i = 0
		}
		dstval = append(dstval,uint16(i))
	}
	return
}

//获取所有RIDS
func (this *XSvrer)getRidsFromRegs(register,numRegs,endRegister int)(Rids []string,framer2 *framer.Exception){
	for i := register; i < numRegs ; i ++ {
		Rids = append(Rids,this.mapper[uint16(i)])
	}

	if len(Rids) != endRegister - register {
		return nil,&framer.IllegalDataAddress
	}
	return Rids,&framer.Success
}


//从数据包获取寄存器起始地址、数量、终止寄存器地址
func registerAddressAndNumber(frame framer.Framer) (register int, numRegs int, endRegister int) {
	data := frame.GetData()
	register = int(binary.BigEndian.Uint16(data[0:2]))
	numRegs = int(binary.BigEndian.Uint16(data[2:4]))
	endRegister = register + numRegs
	return register, numRegs, endRegister
}


