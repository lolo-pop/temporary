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
	deltaRPS := map[int]float32{
		1: 0.5,
		2: 0.8,
		3: 0.3,
	}

	var wg sync.WaitGroup
	var m sync.Map

	for key, value := range deltaRPS {
		wg.Add(1)
		go warmupFunction(key, value, &wg, &m)
	}

	// Do other things here

	// Wait for goroutines to complete
	for key := range deltaRPS {
		for {
			if _, ok := m.Load(key); ok {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Printf("%d completed", key)
	}

	// Do more things here
}
