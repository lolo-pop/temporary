package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
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

func main() {
	sequences := []float64{0}
	requestData := PredictionRequest{
		FunctionName:       "test",
		MonitoringSequence: sequences,
	}
	log.Printf("sequences in PredictFunctionRPS is %v", sequences)
	// 将请求数据转换为 JSON 格式
	requestDataJson, err := json.Marshal(requestData)
	if err != nil {

	}

	// 发送 POST 请求
	response, err := http.Post("http://localhost:5000/predict", "application/json", bytes.NewBuffer(requestDataJson))
	if err != nil {

	}
	defer response.Body.Close()

	// 读取响应主体中的 JSON 字符串
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {

	}
	jsonStr := strings.TrimSpace(string(body))
	fmt.Println(jsonStr)
	var responseData PredictionResponse
	err = json.Unmarshal([]byte(string(jsonStr)), &responseData)
	if err != nil {

	}
	fmt.Println(responseData.Mean)
}
