package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

func getContainerIP() string {
	// 获取容器内部的IP地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		return ""
	}
	for _, addr := range addrs {
		ipnet, ok := addr.(*net.IPNet)
		if ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return ""
}

type PredictionResponse struct {
	FunctionName string  `json:"function_name"`
	StartDate    string  `json:"start_date"`
	Quantile01   float64 `json:"quantile0.1"`
	Quantile02   float64 `json:"quantile0.2"`
	Quantile03   float64 `json:"quantile0.3"`
	Quantile04   float64 `json:"quantile0.4"`
	Quantile05   float64 `json:"quantile0.5"`
	Quantile06   float64 `json:"quantile0.6"`
	Quantile07   float64 `json:"quantile0.7"`
	Quantile08   float64 `json:"quantile0.8"`
	Quantile09   float64 `json:"quantile0.9"`
	Mean         float64 `json:"mean"`
}
type PredictionRequest struct {
	FunctionName       string    `json:"function_name"`
	MonitoringSequence []float64 `json:"monitoring_sequence"`
}
type ImageData struct {
	Name string `json:"name"`
	From string `json:"from"`
	Data string `json:"data"`
}

func test() {
	name := ImageData{"d", "d", "d"}
	test := []ImageData{name}
	jsonData, err := json.Marshal(test)
	if err != nil {

	}
	client := &http.Client{}
	serviceURL := "http://127.0.0.1:8080/function/service-2-0/"
	// serviceURL := "http://localhost:8081/processImages"
	// resp, err := http.Post(serviceURL, "application/json", bytes.NewBuffer(jsonData))
	req, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		errMsg := fmt.Sprintf("send to service failed: %s", err.Error())
		log.Fatal(errMsg)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending batch to service: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Batch sent with status: %d\n", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Response Body:", string(body))
}
func main() {
	/*
		var wg sync.WaitGroup
		for i := 0; i < 9; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				test()
			}()
		}
		wg.Wait()
		fmt.Println("all goroutines are done")
	*/
	start := time.Now()
	test()
	end := time.Since(start)
	fmt.Printf("time cost: %v\n", end)
}
