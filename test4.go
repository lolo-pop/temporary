package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	nums := make([]int, 20)
	for i := 0; i < 20; i++ {
		nums[i] = rand.Intn(4)
	}

	fmt.Println(nums)
}
