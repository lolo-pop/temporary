package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var scalingWindows int64

func init() {
	env, ok := os.LookupEnv("SCALING_WINDOWS")
	if !ok {
		log.Fatal("$scaling windows not set")
	}
	var err error
	val, err := strconv.Atoi(env)
	if err != nil {
		log.Fatal(err.Error())
	}
	scalingWindows = int64(val)
}

const (
	batchNum = 4                      // 定义每个batch包含的图片数量
	timeout  = 500 * time.Millisecond // 定义超时时间
)

// 定义一个结构体，用于存储图片数据
type ImageData struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

// 定义一个结构体，用于存储batch数据
type BatchData struct {
	Images []ImageData `json:"images"`
}

// 定义一个结构体，用于存储处理结果
type ResultData struct {
	Name   string `json:"name"`
	Result string `json:"result"`
}

// 定义一个channel用于传递batch数据
var batchChan = make(chan BatchData)

// 定义一个channel用于传递处理结果
var resultChan = make(chan ResultData)

// 定义一个WaitGroup，用于等待所有goroutine完成
var wg = sync.WaitGroup{}

// 定义一个函数用于处理batch数据
func processBatch(batch BatchData) {
	// 模拟处理时间
	time.Sleep(5 * time.Second)

	// 构造处理结果
	result := "processed"

	// 将处理结果发送到resultChan
	resultChan <- ResultData{Name: batch.Images[0].Name, Result: result}
}

// 定义一个函数用于处理请求
func handleRequest(w http.ResponseWriter, r *http.Request) {
	// 读取请求体数据
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 解析请求体数据为ImageData结构体
	var imageData ImageData
	err = json.Unmarshal(body, &imageData)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 将ImageData结构体发送到batchChan
	batchChan <- BatchData{Images: []ImageData{imageData}}

	// 返回响应
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "image received")
}

// 定义一个函数用于监听batchChan，并将接收到的batch数据发送给另一个容器处理
func sendBatch() {
	var images []ImageData

	//继续：

	var timeoutTimer *time.Timer

	for {
		select {
		case batch := <-batchChan:
			images = append(images, batch.Images...)
			// 如果batch数量达到batchNum或者超时时间到达，就发送batch数据
			if len(images) == batchNum || timeoutTimer != nil {
				// 停止超时计时器
				if timeoutTimer != nil {
					timeoutTimer.Stop()
					timeoutTimer = nil
				}
				// 构造batch数据
				batchData := BatchData{Images: images}
				// 将batch数据发送到另一个容器处理
				go processBatch(batchData)
				// 重置images
				images = []ImageData{}
			}
		case <-time.After(timeout):
			// 如果超时时间到达，就发送已接收到的图片数据
			if len(images) > 0 {
				// 构造batch数据
				batchData := BatchData{Images: images}
				// 将batch数据发送到另一个容器处理
				go processBatch(batchData)
				// 重置images
				images = []ImageData{}
			}
			// 重置超时计时器
			timeoutTimer = time.NewTimer(timeout)
			// 等待超时计时器到达
			<-timeoutTimer.C
			// 重置timeoutTimer
			timeoutTimer = nil
		}
	}
}

// 定义一个函数用于接收处理结果，并将结果返回给图片发送者
func handleResult() {
	for {
		// 从resultChan接收处理结果
		result := <-resultChan

		// 模拟处理时间
		time.Sleep(1 * time.Second)

		// 将处理结果返回给图片发送者
		client := &http.Client{}
		payload, err := json.Marshal(result)
		if err != nil {
			log.Printf("Error marshaling JSON: %s", err)
			continue
		}
		req, err := http.NewRequest("POST", "http://image-sender/result", bytes.NewReader(payload))
		if err != nil {
			log.Printf("Error creating HTTP request: %s", err)
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending HTTP request: %s", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Unexpected status code: %d", resp.StatusCode)
			continue
		}
	}
}

func main() {
	// 启动两个goroutine分别用于监听batchChan和resultChan
	go sendBatch()
	go handleResult()

	// 注册请求处理函数
	http.HandleFunc("/image", handleRequest)

	// 启动HTTP服务器
	log.Fatal(http.ListenAndServe(":8080", nil))
}
