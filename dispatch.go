package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	serviceURL      = "http://localhost:8081/processImages"
	imageBufferSize = 4
	timeout         = 500 * time.Millisecond
)

var mu sync.Mutex

func main() {
	http.HandleFunc("/sendImage", receiveImageHandler)
	http.ListenAndServe(":8082", nil)
}

type image struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

var batch []image
var timer *time.Timer

func init() {
	batch = make([]image, 0, imageBufferSize)
	timer = time.NewTimer(timeout)
}

func receiveImageHandler(w http.ResponseWriter, r *http.Request) {
	var imageData image
	err := json.NewDecoder(r.Body).Decode(&imageData)
	// imageData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading image data", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	//batch := make([]image, 0, imageBufferSize)
	//timer := time.NewTimer(timeout)

	select {
	case <-timer.C:
		if len(batch) > 0 {
			fmt.Printf("batch len is %d\n", len(batch))
			fmt.Printf("batch is %v\n", batch)
			go sendBatchToService(batch)
			batch = []image{}
		}
		timer.Reset(timeout)
	default:
		mu.Lock()
		batch = append(batch, imageData)
		mu.Unlock()
		if len(batch) == imageBufferSize {
			fmt.Printf("batch len is %d\n", len(batch))
			fmt.Printf("batch is %v\n", batch)
			go sendBatchToService(batch)
			batch = []image{}
			timer.Stop()
			timer.Reset(timeout)
		} else if !timer.Stop() && len(timer.C) > 0 {
			<-timer.C
		}
		timer.Reset(timeout)
	}
}

func sendBatchToService(batch []image) {

	/*
		var buffer bytes.Buffer
		for _, img := range batch {
			buffer.Write(img)
			buffer.WriteByte('\n')
		}
	*/
	jsonData, err := json.Marshal(batch)
	if err != nil {
		// 处理错误
	}
	resp, err := http.Post(serviceURL, "application/octet-stream", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending batch to service: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Batch sent with status: %s\n", resp.Status)
}
