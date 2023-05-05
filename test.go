package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type PredictionRequest struct {
	FunctionName       string   `json:"function_name"`
	MonitoringSequence []uint64 `json:"monitoring_sequence"`
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

func main() {
	requestData := PredictionRequest{
		FunctionName:       "myFunction",
		MonitoringSequence: []uint64{10, 20, 30, 40, 50},
	}

	// 将请求数据转换为 JSON 格式
	requestDataJson, err := json.Marshal(requestData)
	if err != nil {
		panic(err)
	}

	// 发送 POST 请求
	response, err := http.Post("http://localhost:5000/predict", "application/json", bytes.NewBuffer(requestDataJson))
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	// 读取响应主体中的 JSON 字符串
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	jsonStr := strings.TrimSpace(string(body))
	fmt.Println(jsonStr)
	var responseData PredictionResponse
	err = json.Unmarshal([]byte(string(jsonStr)), &responseData)
	if err != nil {
		panic(err)
	}

	// 打印预测结果

}
