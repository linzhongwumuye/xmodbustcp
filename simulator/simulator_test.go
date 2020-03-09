package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestSimulatorDataSrc(t *testing.T) {

	//连接Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	})

	_, e := redisClient.Ping().Result()
	if e != nil {
		fmt.Println("Redis连接出错.....")
		return
	}
	fmt.Println("Redis连接成功")

	/*
		测点编码：
		模拟量:0_1_2_1(0:本层、1:设备1、2:状态量、1：测点编号)
		状态量:0_1_1_2
	*/
	valkey := make(map[string]string)
	for i := 0; i < 99; i++ {
		var val, key string
		if 0 == i%2 {
			key = strconv.Itoa(int(rand.Int63n(2)))
			val = "0_1_1_" + strconv.Itoa(i)
		} else {
			key = strconv.FormatFloat(rand.Float64(), 'f', 2, 32)
			val = "0_1_2_" + strconv.Itoa(i)
		}
		if _, ok := valkey[val]; !ok {

			valkey[val] = key
		}
		redisClient.Set(val, key, 0)
	}
	err := redisClient.Close()
	if err != nil {
		fmt.Println(err)
	}
}

func TestManyRequest(t *testing.T) {
	con, err := net.Dial("tcp4", "127.0.0.1:12789")
	if err != nil {
		fmt.Println(err)
	}
	data1 := []byte{0x19, 0x18, 0x00, 0x00, 0x00, 0x06, 0x00, 0x03, 0x00, 0x00, 0x00, 0x16}
	data2 := make([]byte, 512, 1024)
	for {
		data1[9] = byte(rand.Intn(10))
		data1[11] = byte(rand.Intn(80))
		_, err := con.Write(data1)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(time.Second * 1)
		_, err = con.Read(data2)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("data2", data2)
		time.Sleep(time.Second * 1)
	}
}
