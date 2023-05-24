package main

import "fmt"

func main() {
	t := []int{}
	b := []int{1, 2, 3, 4, 5, 6, 7, 8}
	a := 10
	for a > 0 && len(t) < len(b) {
		t = append(t, a)
		a = a - 1
		fmt.Println(t, a)
	}
}

type SCconfig struct {
	Name      string  `json:"name"`
	BatchSize int     `json:"batchSize"`
	Cpu       float64 `json:"cpu"`
	Mem       float64 `json:"mem"`
	//LowRps    float64 `json:"lowRps"`
	//UpRps     float64 `json:"upRps"`
	Node string `json:"node"`
}
