package main

import (
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

const redisAddr = "faas-redis-master.openfaas.svc.cluster.local:6379"
const redisPassword = "Y7MkRCBORP"

func main() {
	// Connect to Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // no password set
		DB:       0,             // use default DB
	})
	defer client.Close()

	// Start a goroutine to write key-value pairs

	// Generate a random key and value
	key := "key1"
	value := []string{"value1", "value2", "value3"}

	// Set the key-value pair in Redis
	err := client.Set(key, strings.Join(value, ","), 0).Err()
	if err != nil {
		log.Printf("Failed to set key-value pair in Redis: %v", err)
	}

	log.Printf("Set key %swith value %v", key, value)

	// Wait for a random amount of time before setting the next key-value pair
	time.Sleep(time.Duration(1+time.Now().UnixNano()%5) * time.Second)

	// Start a goroutine to read key-value pairs

	for {
		start := time.Now()
		valueStr, err := client.Get(key).Result()
		if err != nil {
			log.Printf("Failed to get value of key %s from Redis: %v", key, err)

		}

		// Split the value into a list of strings
		value = strings.Split(valueStr, ",")
		end := time.Since(start)
		log.Printf("Got key %s with value %v, time %v", key, value, end)

		// Wait for a random amount of time before getting the next key-value pair
		time.Sleep(time.Duration(1+time.Now().UnixNano()%5) * time.Second)

		// Wait for the goroutines to finish
	}

}
