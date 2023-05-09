package main

import (
	"fmt"
	"strconv"
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
	a := 2.68435456e+08
	aa := strconv.Itoa(int(a/1024/1024)) + strconv.Itoa(int(a/1024/1024))
	fmt.Println(aa)
}
