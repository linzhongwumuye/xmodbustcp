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
	保留此方式，当遇到下发控制量等值时可以实现其他函数
*/

//0x03:读寄存器,然后回复相关的值
func (this *XSvrer) ReadHoldingRegisters(s *modbusserver.Server, frame framer.Framer) ([]byte, *framer.Exception) {
	register, numRegs, endRegister := registerAddressAndNumber(frame)
	if endRegister > 65535 {
		log.Error("Request’s Max Register Addr Is Invalid")
		return nil, &framer.IllegalDataAddress
	}

	Rids, framer2 := this.getRidsFromRegs(register, numRegs, endRegister)
	if framer2 != &framer.Success {
		log.Error("Get RID Error", frame)
		return []byte{}, framer2
	}

	values, err := this.datasrc.GetData(Rids)
	if err != nil {
		log.Error("Get Data Error", err, Rids, frame.GetData())
		return []byte{}, &framer.MemoryParityError
	}

	convertValues := GJValTOModbus(values)
	log.Debugf("Request Register Start Addr: %02x\t Register Nums: %02x\t", register, numRegs)
	framer.SetResponseWith0x03(frame, uint16(numRegs), convertValues)

	return frame.GetData(), &framer.Success
}

//TODO:: 负数、>65536的数,平台值转换为Modbus数据
func GJValTOModbus(srcval []string) (dstval []uint16) {
	for _, val := range srcval {
		if strings.ContainsAny(val, ".") {
			val = strings.Replace(val, ".", "", 1)
		}
		i, e := strconv.Atoi(val)
		if e != nil {
			log.Info("Original Value Translating Error", e, "Original Value：", val)
			i = 0
		}
		dstval = append(dstval, uint16(i))
	}
	return
}

//获取所有RIDS
func (this *XSvrer) getRidsFromRegs(register, numRegs, endRegister int) (Rids []string, framer2 *framer.Exception) {
	for i := 0; i < numRegs; i++ {
		Rids = append(Rids, this.mapper[uint16(i+register)])
	}
	if len(Rids) != endRegister-register {
		log.Error("len ", len(Rids), "end - start", endRegister-register)
		return nil, &framer.IllegalDataAddress
	}
	return Rids, &framer.Success
}

//从数据包获取寄存器起始地址、数量、终止寄存器地址
func registerAddressAndNumber(frame framer.Framer) (register int, numRegs int, endRegister int) {
	data := frame.GetData()
	register = int(binary.BigEndian.Uint16(data[0:2]))
	numRegs = int(binary.BigEndian.Uint16(data[2:4]))
	endRegister = register + numRegs
	return register, numRegs, endRegister
}
