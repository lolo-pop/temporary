package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

const redisAddr = "faas-redis-master.openfaas.svc.cluster.local:6379"
const redisPassword = "Y7MkRCBORP"

type SCconfig struct {
	Name string  `json:"name"`
	Bs   int     `json:"batchSize"`
	Cpu  float64 `json:"cpu"`
	Mem  float64 `json:"mem"`
}

func main() {
	// 创建 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // Redis 无密码
		DB:       0,             // 使用默认数据库
	})
	key := "example_key"
	exists, err := client.Exists(key).Result()
	if err != nil {
		panic(err)
	}

	if exists == 1 {
		fmt.Printf("键 %s 存在于 Redis 中\n", key)
	} else {
		fmt.Printf("键 %s 不存在于 Redis 中\n", key)
	}

	start := time.Now()
	// 创建一个 SCconfig 结构体列表
	configs := []SCconfig{}

	// 将 SCconfig 结构体列表转换为 JSON 字符串
	configsJSON, err := json.Marshal(configs)
	if err != nil {
		panic(err)
	}

	//将 JSON 字符串作为值存储到 Redis 中
	err = client.Set("configurations", string(configsJSON), 0).Err()
	if err != nil {
		panic(err)
	}
	end := time.Since(start)
	fmt.Printf("set key value, time %v", end)
	// 从 Redis 中获取值，并将其转换回结构体列表
	start = time.Now()
	val, err := client.Get("configurations").Result()
	if err != nil {
		panic(err)
	}

	var configsFromRedis []SCconfig
	err = json.Unmarshal([]byte(val), &configsFromRedis)
	if err != nil {
		panic(err)
	}
	end = time.Since(start)
	// 打印从 Redis 中获取的结构体列表
	fmt.Println(configsFromRedis, len(configsFromRedis), end)

}
