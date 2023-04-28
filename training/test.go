package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	// 准备POST请求的数据
	postData := []byte(`{"monitoring_sequence": "503, 411, 388, 320, 295, 288, 150, 203, 73,735"}`)

	// 创建一个HTTP POST请求
	req, err := http.NewRequest("POST", "http://localhost:5000/predict", bytes.NewBuffer(postData))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送HTTP POST请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// 读取HTTP响应内容
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 输出HTTP响应内容
	fmt.Println(buf.String())
}
