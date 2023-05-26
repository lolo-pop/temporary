package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Image struct {
	Name string `json:"name"`
	From string `json:"from"`
	Data string `json:"data"`
}

//type Batch struct {
//	Pictures []image `json:"pictures"`
//}

type Result struct {
	//Name  string `json:"name"`
	//Error string `json:"error,omitempty"`
	//Data  string `json:"data,omitempty"`
	returnData []Image
}

type Dispatcher struct {
	batches  chan Image
	results  chan Result
	quit     chan bool
	received []Image
	status   chan ScConfig
	mutex    sync.Mutex
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		batches:  make(chan Image), // Buffer up to 10 batches
		results:  make(chan Result),
		quit:     make(chan bool),
		status:   make(chan ScConfig),
		received: []Image{},
	}
}

type ScConfig struct {
	BatchSize int `json:"batchSize"`
}
type FunctionStatus struct {
	Functions ScConfig
}

func NewFunctionStatus() *FunctionStatus {
	return &FunctionStatus{
		Functions: ScConfig{},
	}
}
func (d *Dispatcher) Start() {
	fmt.Println("Starting dispatcher...")
	timeout := time.Millisecond * 500
	timer := time.NewTimer(timeout)
	bs := 0
	for {
		//add 获得batch size代码
		select {
		case batchSize := <-d.status:
			bs = batchSize.BatchSize
			// fmt.Println("batch size:", bs)
		case batch := <-d.batches:
			//fmt.Printf("Received image is %v\n", batch)
			d.mutex.Lock()
			d.received = append(d.received, batch)
			// Send batch to service if there are enough pictures
			if len(d.received) == bs {
				// pics := d.received[:bs]
				// d.received = d.received[bs:]
				pics := d.received
				d.received = []Image{}
				go d.sendToService(pics)
				timer.Stop()
				timer.Reset(timeout)
			} else if !timer.Stop() && len(timer.C) > 0 {
				fmt.Println("here")
				<-timer.C
			}
			d.mutex.Unlock()
			// timer.Reset(timeout)
		case result := <-d.results:
			// fmt.Printf("Received result for batch %s\n", result.returnData)
			// TODO: Send result to sender

			// fmt.Printf("Received result for batch %v\n", result)
			// case <-time.After(500 * time.Millisecond):
		case <-timer.C:
			if len(d.received) > 0 { // Send remaining pictures to service
				pics := d.received
				d.received = []Image{}
				go d.sendToService(pics)
				//timer.Stop()
			}
			timer.Reset(timeout)
		case <-time.After(500 * time.Millisecond):
			if len(d.received) > 0 { // Send remaining pictures to service
				pics := d.received
				d.received = []Image{}
				go d.sendToService(pics)
			}
		case <-d.quit:
			fmt.Println("Stopping dispatcher...")
			return
		}
	}
}

func (d *Dispatcher) sendToService(pics []Image) {
	fmt.Printf("sending %v\n", pics)
	// TODO: Call service API and handle response
	jsonData, err := json.Marshal(pics)
	if err != nil {
		// 处理错误
	}
	serviceURL := "http://localhost:8081/processImages"
	resp, err := http.Post(serviceURL, "application/octet-stream", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending batch to service: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Batch sent with status: %s\n", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %s\n", err)
		return
	}
	var returnData []Image
	err = json.Unmarshal(body, &returnData)
	if err != nil {
	}
	var result Result
	fmt.Printf("results %v\n", returnData)
	result.returnData = returnData
	d.results <- result
}

func (f *FunctionStatus) Init() {
	i := 0
	for i < 1000 {
		f.Functions.BatchSize = 3
		i++
		time.Sleep(time.Second * 10)
	}
}
func main() {
	dispatcher := NewDispatcher()
	status := NewFunctionStatus()
	go dispatcher.Start()
	go status.Init()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var batch Image
		err := json.NewDecoder(r.Body).Decode(&batch)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		// from := batch.From
		var t ScConfig

		t = status.Functions
		dispatcher.status <- t
		dispatcher.batches <- batch
		// fmt.Printf("current results is %v\n", dispatcher.rcvResults)
	})

	fmt.Println("Listening on port 8082...")
	http.ListenAndServe(":8082", nil)
}