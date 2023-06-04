package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/lolo-pop/faas-dispatch/pkg/types"
)

var (
	natsUrl       string
	level         string
	gatewayURL    string
	redisUrl      string
	redisPassword string
	redisKey      string
)

func init() {
	var ok bool
	natsUrl, ok = os.LookupEnv("NATS_URL")
	if !ok {
		log.Fatal("$NATS_URL not set")
	}
	level, ok = os.LookupEnv("LEVEL")
	if !ok {
		log.Fatal("$LEVEL not set")
	}
	gatewayURL, ok = os.LookupEnv("GATEWAY_URL")
	if !ok {
		log.Fatal("$scaling windows not set")
	}
	redisUrl, ok = os.LookupEnv("REDIS_URL")
	if !ok {
		log.Fatal("$REDIS_URL not set")
	}
	redisPassword, ok = os.LookupEnv("REDIS_PASS")
	if !ok {
		log.Fatal("$REDIS_PASS not set")
	}
	redisKey, ok = os.LookupEnv("REDIS_KEY")
	if !ok {
		log.Fatal("$REDIS_KEY not set")
	}
}

const (
	timeout = 300 * time.Millisecond // 定义超时时间
)

// 定义一个结构体，用于存储图片数据
type ImageData struct {
	Name string `json:"name"`
	From string `json:"from"`
	Data string `json:"data"`
}

// 定义一个结构体，用于存储batch数据
type BatchData struct {
	Images []ImageData `json:"images"`
}

// 定义一个结构体，用于存储处理结果
type ResultData struct {
	Name string `json:"name"`
	From string `json:"from"`
	Data string `json:"data"`
}
type BatchResult struct {
	resultData []ResultData
}

type Dispatcher struct {
	image    chan ImageData
	result   chan BatchResult
	quit     chan bool
	received []ImageData
	status   chan Functions
	mutex    sync.Mutex
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		image:    make(chan ImageData), // Buffer up to 10 batches
		result:   make(chan BatchResult),
		quit:     make(chan bool),
		status:   make(chan Functions),
		received: []ImageData{},
	}
}

type Functions struct {
	functions []types.ScConfig
}
type FunctionStatus struct {
	Status Functions
}

func NewFunctionStatus() *FunctionStatus {
	return &FunctionStatus{
		Status: Functions{},
	}
}
func (d *Dispatcher) Start() {
	fmt.Println("Starting dispatcher...")
	timeout := time.Millisecond * 500
	timer := time.NewTimer(timeout)
	bs := 2
	index := 0
	functionNum := 100
	functionName := "service-0-0"
	for {
		//add 获得batch size代码
		select {
		case funcStatus := <-d.status: // 这里需要更改
			// bs = batchSize.BatchSize
			fmt.Println("batch size:", funcStatus.functions)
			functionNum = len(funcStatus.functions)
			if functionNum != 0 {
				bs = funcStatus.functions[index].BatchSize
				functionName = funcStatus.functions[index].Name
			} else {
				log.Printf("warming functionStatus is empty")
			}
		case image := <-d.image:
			//fmt.Printf("Received image is %v\n", batch)
			d.mutex.Lock()
			d.received = append(d.received, image)
			// Send batch to service if there are enough pictures
			if len(d.received) == bs {
				// pics := d.received[:bs]
				// d.received = d.received[bs:]
				pics := d.received
				d.received = []ImageData{}
				go d.sendToService(functionName, pics)
				index++
				if index >= functionNum && functionNum != 0 {
					index = index % functionNum
				}
				timer.Stop()
				timer.Reset(timeout)
			} else if !timer.Stop() && len(timer.C) > 0 {
				fmt.Println("here")
				<-timer.C
			}
			d.mutex.Unlock()
			// timer.Reset(timeout)
		case result := <-d.result:
			fmt.Printf("Received result for batch %s\n", result.resultData)
			// TODO: Send result to sender
			for _, returnData := range result.resultData {
				ip := returnData.From
				go d.sendToSender(ip, returnData)
			}
			// fmt.Printf("Received result for batch %v\n", result)
			// case <-time.After(500 * time.Millisecond):
		case <-timer.C:
			if len(d.received) > 0 { // Send remaining pictures to service
				pics := d.received
				d.received = []ImageData{}
				go d.sendToService(functionName, pics)
				index++
				if index >= functionNum && functionNum != 0 {
					index = index % functionNum
				}
				timer.Stop()
			}
			timer.Reset(timeout)

		case <-time.After(timeout):
			if len(d.received) > 0 { // Send remaining pictures to service
				pics := d.received
				d.received = []ImageData{}
				go d.sendToService(functionName, pics)
				index++
				if index >= functionNum && functionNum != 0 {
					index = index % functionNum
				}
			}
			// timer.Reset(timeout)
		case <-d.quit:
			fmt.Println("Stopping dispatcher...")
			return
		}
	}
}

func (d *Dispatcher) sendToService(functionName string, pics []ImageData) {

	client := &http.Client{}
	fmt.Printf("sending %v\n", pics)
	jsonData, err := json.Marshal(pics)
	if err != nil {
		// 处理错误
	}
	serviceURL := fmt.Sprintf("%s/function/%s", gatewayURL, functionName)
	// serviceURL := "http://localhost:8081/processImages"
	// resp, err := http.Post(serviceURL, "application/json", bytes.NewBuffer(jsonData))
	req, err := http.NewRequest("POST", serviceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		errMsg := fmt.Sprintf("send to service failed: %s", err.Error())
		log.Fatal(errMsg)
	}
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	user := "admin"
	password := "admin"
	req.SetBasicAuth(user, password)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending batch to service: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Batch sent with status: %d\n", resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return
	}
	var returnData []ResultData
	err = json.Unmarshal(body, &returnData)
	if err != nil {
	}
	var result BatchResult
	result.resultData = returnData
	d.result <- result
}
func (d *Dispatcher) sendToSender(ip string, returnData ResultData) {
	/*
		url := fmt.Sprintf("http://%s:8080/sendResult", ip)
		jsonData, err := json.Marshal(returnData)
		if err != nil {
			fmt.Printf("error marshalling\n")
		}
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error sending result to sender %s: %v\n", ip, err)
			return
		}
		defer resp.Body.Close()
		//fmt.Printf("Sender %s result sent with status: %s\n", ip, resp.Status)
	*/
	key := returnData.Name
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})
	err := client.Set(key, "1", 0).Err()
	if err != nil {
		fmt.Printf("Error sending result to sender %s: %v\n", ip, err)
		return
	}
	log.Printf("sending %s result to sender %s\n", key, ip)
}

func (f *FunctionStatus) Init() {
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})
	key := fmt.Sprintf("%s-%s", redisKey, level)
	for {
		val, err := client.Get(key).Result()
		if err != nil {
			log.Printf("get key failed %s", err.Error())
		}
		var value []types.ScConfig
		err = json.Unmarshal([]byte(val), &value)
		if err != nil {
			log.Printf("get key failed %s", err.Error())
		}
		f.Status.functions = value
		log.Printf("active service functions: %v", f.Status.functions)
		time.Sleep(time.Second)
	}
}

func main() {
	dispatcher := NewDispatcher()
	status := NewFunctionStatus()
	go dispatcher.Start()
	go status.Init()
	time.Sleep(time.Second)
	http.HandleFunc("/sendImage", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Println("got one message")
		var image ImageData
		err := json.NewDecoder(r.Body).Decode(&image)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		// from := batch.From
		var t Functions

		t = status.Status
		dispatcher.status <- t
		dispatcher.image <- image
		// fmt.Printf("current results is %v\n", dispatcher.rcvResults)
	})

	fmt.Println("Listening on port 5000...")
	http.ListenAndServe(":5000", nil)
}
