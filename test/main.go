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

	// 构造要发送的请求的 JSON 数据
	requestData := map[string]interface{}{
		"service": functionName,
		"image":   imageName,
		"envVars": map[string]string{
			"BATCH_SIZE":   "4",
			"NATS_ADDRESS": "nats://10.244.0.105:4222",
			"NATS_SUBJECT": "image-test",
			"RESOLUTION":   "512x512",
		},
		"envProcess": "python3 index.py",
		"limits": map[string]string{
			"memory": "1024Mi",
			"cpu":    "1000m",
		},
		"request": map[string]string{
			"memory": "1024Mi",
			"cpu":    "1000m",
		},
		"labels": map[string]string{
			"com.openfaas.scale.zero": "true",
			"com.openfaas.scale.min":  "1",
			"com.openfaas.scale.max":  "1",
		},
		"constraints": []string{"kubernetes.io/hostname=dragonlab11"},
	}

	requestBody, err := json.Marshal(requestData)
	fmt.Println(requestBody)
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

type FunctionDeployment struct {

	// Service is the name of the function deployment
	Service string `json:"service"`

	// Image is a fully-qualified container image
	Image string `json:"image"`

	// Namespace for the function, if supported by the faas-provider
	Namespace string `json:"namespace,omitempty"`

	// EnvProcess overrides the fprocess environment variable and can be used
	// with the watchdog
	EnvProcess string `json:"envProcess,omitempty"`

	// EnvVars can be provided to set environment variables for the function runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Constraints are specific to the faas-provider.
	Constraints []string `json:"constraints,omitempty"`

	// Secrets list of secrets to be made available to function
	Secrets []string `json:"secrets,omitempty"`

	// Labels are metadata for functions which may be used by the
	// faas-provider or the gateway
	Labels *map[string]string `json:"labels,omitempty"`

	// Annotations are metadata for functions which may be used by the
	// faas-provider or the gateway
	Annotations *map[string]string `json:"annotations,omitempty"`

	// Limits for function
	Limits *FunctionResources `json:"limits,omitempty"`

	// Requests of resources requested by function
	Requests *FunctionResources `json:"requests,omitempty"`

	// ReadOnlyRootFilesystem removes write-access from the root filesystem
	// mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}
