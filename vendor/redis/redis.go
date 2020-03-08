package redis

import (
	"errors"
	"time"
	"github.com/garyburd/redigo/redis"
)

//redis单连接
type RedisConnection struct {
	Network        string        //tcp连接
	Address        string        //ip和端口
	ConnTimeOut    time.Duration //连接超时
	ReadTimeOut    time.Duration //读超时
	WriteTimeOut   time.Duration //写超时
	ReconnTimes    uint          //重连次数
	ReconnInterval uint          //重连间隔
	redisPool      *redis.Pool   //redis连接池
}

//创建redis连接池
func (this *RedisConnection) Connect() (err error) {
	this.redisPool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.DialTimeout(this.Network, this.Address, this.ConnTimeOut, this.ReadTimeOut, this.WriteTimeOut)
		},
	}
	return nil
}

//批量获取接口，once指定一次最多获取多少条数据
func (this *RedisConnection) MultiGet(keys []interface{}, once int) (result []string, err error) {
	if once <= 0 {
		return nil, errors.New("MultiGet Argument once <= 0")
	}

	var keys_len = len(keys)
	result = make([]string, keys_len)
	c := this.redisPool.Get()
	defer c.Close()

	//每次获取once条数据
	for begin, end := 0, 0; begin < keys_len; begin += once {
		end = MinInt(begin+once, keys_len)
		reply, err := c.Do("mget", keys[begin:end]...)
		if err != nil {
			return nil, err
		}

		values_arr, _ := redis.Strings(reply, nil)
		copy(result[begin:], values_arr)
	}

	return result, nil
}

//批量写入接口，once指定一次最多写入多少条数据
func (this *RedisConnection) MultiSet(kvs []interface{}, once int) (err error) {
	if once <= 0 {
		return errors.New("MultiSet Argument once <= 0")
	}

	var keys_len = len(kvs)
	if keys_len%2 != 0 {
		return errors.New("MultiSet Argument kvs sum is not event")
	}
	c := this.redisPool.Get()
	defer c.Close()

	//每次写入once条数据
	for begin, end := 0, 0; begin < keys_len; begin += 2 * once {
		end = MinInt(begin+2*once, keys_len)
		if _, err = c.Do("mset", kvs[begin:end]...); err != nil {
			return err
		}
	}

	return nil
}

//KYES接口
func (this *RedisConnection) Keys(key interface{}) (result []interface{}, err error) {

	c := this.redisPool.Get()
	defer c.Close()

	reply, err := c.Do("keys", key)
	if err != nil {
		return nil, err
	}

	values, err := redis.Strings(reply, nil)
	if err != nil {
		return nil, err
	}

	var ret []interface{}
	for _, v := range values {
		ret = append(ret, v)
	}

	return ret, nil
}

//析构函数
func (this *RedisConnection) Destroy() {
	if this.redisPool != nil {
		this.redisPool.Close()
	}
}

//int最大值
func MaxInt(first int, args ...int) int {
	for _, v := range args {
		if first < v {
			first = v
		}
	}
	return first
}

//uint最大值
func MaxUInt(first uint32, args ...uint32) uint32 {
	for _, v := range args {
		if first < v {
			first = v
		}
	}
	return first
}

//int64最大值
func MaxInt64(first int64, args ...int64) int64 {
	for _, v := range args {
		if first < v {
			first = v
		}
	}
	return first
}

//int最小值
func MinInt(first int, args ...int) int {
	for _, v := range args {
		if first > v {
			first = v
		}
	}
	return first
}