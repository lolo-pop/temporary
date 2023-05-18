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

func main() {
	a := 1000
	b := a
	b = b - 100
	fmt.Println(a, b)
	// Do more things here
}
