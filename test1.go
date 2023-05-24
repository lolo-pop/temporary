package main

import "fmt"

func main() {
	SCSLO := []float64{}
	fmt.Println(SCSLO, len(SCSLO))
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
