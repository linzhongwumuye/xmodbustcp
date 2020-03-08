package business

import (
	"fmt"
	"github.com/go-redis/redis"
	"strconv"
	"strings"
	"xlib/log"
	"xmodbustcp/datadrive"
	"xmodbustcp/define"
	"xmodbustcp/modbusserver"
)

var (
	quit = make(chan bool,1)
)


type XSvrer struct {
	*modbusserver.Server             		//ModbusTcp服务
	conf    	define.SvrConfInterface 	//基础配置
	datasrc 	*datadrive.DataerRedisRmq   //数据来源
	mapper  	map[uint16]string       	//映射关系：寄存器-测点ID
}

func StartXSvrer(svconf define.SvrConfInterface) (err error){
	xs := &XSvrer{
		conf:		svconf,
		Server:		modbusserver.NewServer(),
	}

	err = xs.init()
	return
}

func StopXSvrer() {
	log.Info("主动退出")
	modbusserver.StopSvr()
}

func (this *XSvrer)init()(err error){

	//获取Mapper
	this.mapper = GetMapper()

	//初始化数据来源
	this.datasrc,err = datadrive.NewDataer(this.conf.GetRedisAddress(),this.conf.GetRedisConnPool())
	if err != nil {
		return err
	}
	defer this.datasrc.Close()

	//初始化功能码对应函数
	this.RegisterFunctionHandler(0x03,this.ReadHoldingRegisters)

	//开始监听
	for _,ser := range this.conf.GetServers() {
		if err = this.ListenTCP(ser.Addr);err != nil {
			log.Error("监听TCP",err)
		}
	}
	defer this.Close()

	return
}




/*
	此映射关系为新加坡CMI现场制定
	一个KE站点为一个设备，所有测点从{0x00,0x00}开始编号
*/
func GetMapper() map[uint16]string {
	//GetResource拉取所有测点
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
		return nil
	}
	fmt.Println("Redis连接成功")

	all, e := redisClient.Keys("*").Result()
	if e != nil {
		fmt.Println("Keys * error",e)
		return nil
	}
	mapper := make(map[uint16]string)
	for _,rid := range all {
		index := strings.LastIndex(rid, "_")
		reg,err  :=strconv.Atoi(rid[index+1:])
		if err != nil {
			fmt.Println("转换错误",err,reg)
			continue
		}
		if _,ok := mapper[uint16(reg)]; !ok {
			mapper[uint16(reg)] = rid
		}
	}
	return mapper
}