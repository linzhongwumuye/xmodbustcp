package datadrive

import (
	"fmt"
	"github.com/go-redis/redis"

)

//暂时仅使用Redis
type DataerRedisRmq struct {
	redisClient 	*redis.Client  //Redis
	Addr 			string 		   //地址
	PoolSize		int 		   //连接池
}


func NewDataer(Addr string,Poolsize int) (dataer *DataerRedisRmq,err error){
	 dataer = &DataerRedisRmq{
	 	Addr: 			Addr,
	 	PoolSize:		Poolsize,
	 }

	 if err = dataer.init();err != nil {
	 	return nil,fmt.Errorf("dataer初始化失败")
	 }

	 if _,ok  := dataer.redisClient.Ping().Result();ok != nil {
	 	return nil,fmt.Errorf("连接redis失败")
	 }

	 return
}

func (this *DataerRedisRmq)Close()(err error) {
	return  this.redisClient.Close()
}

func(this *DataerRedisRmq)init()error{
	this.redisClient = redis.NewClient(&redis.Options{
		Addr:     	this.Addr,
		Password: 	"",
		DB:       	0,
		PoolSize: 	this.PoolSize,
	})
	return nil
}


//返回值一一对应请求值，如果查询不到的测点，默认值为""
func(this *DataerRedisRmq)GetData(rids []string) (values []string,err error) {
		got,err := this.redisClient.MGet(rids...).Result()
		if err != nil {
			return
		}
		for _, value := range got {
			if value == nil {
				values = append(values, "")
			} else {
				values = append(values, value.(string))
			}
		}
		return
}