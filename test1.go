package main

import (
	"encoding/json"
	"fmt"
)

type FunctionStats struct {
	FunctionName string `json:"function_name"`
	StartDate    string `json:"start_date"`
	Quantile01   string `json:"quantile0.1"`
	Quantile02   string `json:"quantile0.2"`
	Quantile03   string `json:"quantile0.3"`
	Quantile04   string `json:"quantile0.4"`
	Quantile05   string `json:"quantile0.5"`
	Quantile06   string `json:"quantile0.6"`
	Quantile07   string `json:"quantile0.7"`
	Quantile08   string `json:"quantile0.8"`
	Quantile09   string `json:"quantile0.9"`
	Mean         string `json:"mean"`
}

func main() {
	jsonStr := `{"function_name": "myFunction", "start_date": "1970-01-01 00:02:30", "quantile0.1": "14.689090728759766", "quantile0.2": "22.616291046142578", "quantile0.3": "28.97735595703125", "quantile0.4": "32.225460052490234", "quantile0.5": "31.39090347290039", "quantile0.6": "33.746707916259766", "quantile0.7": "49.32496643066406", "quantile0.8": "46.91023254394531", "quantile0.9": "54.31256103515625", "mean": "31.39090347290039"}`
	var functionStats FunctionStats

	err := json.Unmarshal([]byte(jsonStr), &functionStats)
	if err != nil {
		panic(err)
	}

	fmt.Println(functionStats.FunctionName)
	fmt.Println(functionStats.StartDate)
	fmt.Println(functionStats.Quantile01)
	fmt.Println(functionStats.Quantile02)
	fmt.Println(functionStats.Quantile03)
	fmt.Println(functionStats.Quantile04)
	fmt.Println(functionStats.Quantile05)
	fmt.Println(functionStats.Quantile06)
	fmt.Println(functionStats.Quantile07)
	fmt.Println(functionStats.Quantile08)
	fmt.Println(functionStats.Quantile09)
	fmt.Println(functionStats.Mean)
}
