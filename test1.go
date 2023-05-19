package main

import (
	"fmt"
	"sync"
	"time"
)

func warmupFunction(key int, value float32, wg *sync.WaitGroup, m *sync.Map) {
	defer wg.Done()
	fmt.Printf("Key: %d, Value: %f\n", key, value)
	time.Sleep(time.Second * 5)
	m.Store(key, true)

}
func test(t []int) []int {
	t[0] = t[0] - 1
	t[1] = t[1] - 1
	return t
}
func main() {
	var x []int
	x = append(x, 1)
	fmt.Println(x)
}
