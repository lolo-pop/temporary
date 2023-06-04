package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	const n = 1200
	const min = 2
	const max = 16

	var list []int

	for i := 0; i < n; i++ {
		// Generate a random integer between min and max
		x := rand.Intn(max-min+1) + min

		// Check if x has the Sporadic property
		if x == 2 || x == 3 || x == 5 || x == 7 || x == 11 || x == 13 {
			list = append(list, x)
		}
	}

	// Output the list as a Go slice
	fmt.Printf("[]int{%d", list[0])
	for _, x := range list[1:] {
		fmt.Printf(", %d", x)
	}
	fmt.Println("}")
}
