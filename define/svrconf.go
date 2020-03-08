package define

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

//入口
type SvrConfInterface interface {
	GetDbpath() 	 string
	GetDbList() 	 string
	GetLogRollType() string
	GetLogDir() 	 string
	GetLogFile() 	 string
	GetLogCount() 	 int32
	GetLogSize() 	 int64
	GetLogUnit() 	 string
	GetLogLevel() 	 string
	GetLogCompress() int64
	GetDbTimeout()   time.Duration
	GetDbOpenMax() 	 int
	GetDbIdelMax() 	 int
	GetServers() 	 []Server
	GetRedisNet() 	 string
	GetRedisAddress  ()string
	GetRedisConnTimeOut() int
	GetRedisReadTimeOut() int
	GetRedisWriteTimeOut() int
	GetRedisReconnTimes() uint
	GetRedisReconnInterval() uint
	GetRedisConnPool() 		int
	GetPid()				string
}


//日志、数据库、网络
type Svrconf struct {
	Redis struct {
		Network        string //redis的网路类型
		Address        string //redis地址
		ConnTimeOut    int    //redis连接超时
		ReadTimeOut    int    //redis读超时
		WriteTimeOut   int    //redis写超时
		ReconnTimes    uint   //redis重连次数
		ReconnInterval uint   //redis重连间隔
		ConnPool       int    //连接池数量
	}
	Db struct{
		Path 			string				//数据库路径
		Listname 		string  			//表名
		Timeout         time.Duration		//超时时间
		IdelMax         int					//空闲连接
		OpenMax         int					//最大连接
	}
	Logger struct {
		RollType string //日志滚动方式
		Dir      string //日志目录
		File     string //日志文件
		Count    int32  //日志文件数量
		Size     int64  //日志文件大小
		Unit     string //日志大小的单位
		Level    string //日志级别
		Compress int64  // 日志是否压缩
	}
	Servers 	 []Server
	Pid 		 string
}

type Server struct {
	Addr 		string
}

func (this *Svrconf) GetPid()string {
	return this.Pid
}

func (this *Svrconf) GetRedisNet() string {
	return this.Redis.Network
}

func (this *Svrconf) GetRedisAddress() string{
	return this.Redis.Address
}

func (this *Svrconf) GetRedisConnTimeOut() int {
	return this.Redis.ConnTimeOut
}

func (this *Svrconf) GetRedisReadTimeOut() int {
	return this.Redis.ReadTimeOut
}

func (this *Svrconf) GetRedisWriteTimeOut() int {
	return this.Redis.WriteTimeOut
}

func (this *Svrconf) GetRedisReconnTimes() uint {
	return this.Redis.ReconnTimes
}

func (this *Svrconf) GetRedisReconnInterval() uint {
	return this.Redis.ReconnInterval
}

func (this *Svrconf) GetRedisConnPool() int{
	return this.Redis.ConnPool
}


func (this *Svrconf) GetServers() []Server{
	return this.Servers
}

func (this *Svrconf) GetDbpath() string{
	return this.Db.Path
}

func (this *Svrconf) GetDbList() string {
	return this.Db.Listname
}

func (this *Svrconf) GetLogRollType() string{
	return this.Logger.RollType
}

func (this *Svrconf) GetLogDir() string {
	return this.Logger.Dir
}

func (this *Svrconf) GetLogFile() string {
	return this.Logger.File
}

func (this *Svrconf) GetLogCount() int32 {
	return this.Logger.Count
}

func (this *Svrconf) GetLogSize() int64 {
	return this.Logger.Size
}

func (this *Svrconf) GetLogUnit() string {
	return this.Logger.Unit
}

func (this *Svrconf) GetLogLevel() string {
	return this.Logger.Level
}

func (this *Svrconf) GetDbTimeout() time.Duration{
	return this.Db.Timeout
}

func (this *Svrconf) GetDbIdelMax() int{
	return this.Db.IdelMax
}

func (this *Svrconf) GetDbOpenMax() int{
	return this.Db.OpenMax
}

func (this *Svrconf) GetLogCompress() int64{
	return this.Logger.Compress
}

func ReadSvrConf(confile string, confInterface SvrConfInterface) error {
	//打开文件
	fd, err := os.Open(confile)
	if err != nil {
		fmt.Println("Open", confile, err.Error())
		return err
	}
	defer  fd.Close()

	jDecoder := json.NewDecoder(fd)

	if err = jDecoder.Decode(confInterface); err != nil {
		fmt.Println("Json Decode", confile, err.Error())
		return err
	}
	fmt.Println("Read ", confile, "Success")
	return nil
}
