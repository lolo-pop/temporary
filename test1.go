package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	// 创建一个 HTTP 客户端
	client := &http.Client{Timeout: 10 * time.Second}

	// 准备要部署的函数的信息
	functionName := "new-function"
	imageName := "lolopop/nats-test:latest"
	gatewayURL := "http://gateway.openfaas.svc.cluster.local:8080"
	handler := "new-function"
	memoryLimit := "256Mi"
	cpuLimit := "1000m"
	timeout := "60s"
	envVars := map[string]string{
		"BATCH_SIZE":   "1",
		"RESOLUTION":   "512x512",
		"NATS_ADDRESS": "nats://10.244.0.105:4222",
		"NATS_SUBJECT": "image-test",
	}

	// 构造要发送的请求的 JSON 数据
	requestData := map[string]interface{}{
		"serviceName": functionName,
		"image":       imageName,
		"handler":     handler,
		"limits": map[string]string{
			"memory": memoryLimit,
			"cpu":    cpuLimit,
		},
		"environment": envVars,
		"annotations": map[string]string{
			"timeout": timeout,
		},
	}

	requestBody, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("Error marshaling JSON request body: %v", err)
		os.Exit(1)
	}

	// 构造要发送的请求
	req, err := http.NewRequest("POST", gatewayURL+"/system/functions", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v", err)
		os.Exit(1)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	user := "admin"
	password := "admin"
	req.SetBasicAuth(user, password)
	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %v", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected response status code: %d", resp.StatusCode)
		os.Exit(1)
	}

	// 解析响应的 JSON 数据
	var responseMap map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseMap)
	if err != nil {
		fmt.Printf("Error decoding JSON response body: %v", err)
		os.Exit(1)
	}

	// 输出部署结果
	fmt.Printf("Function %s deployed successfully with URL: %s\n", functionName, responseMap["url"])
}
