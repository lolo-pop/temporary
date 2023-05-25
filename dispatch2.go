package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Image struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

//type Batch struct {
//	Pictures []image `json:"pictures"`
//}

type Result struct {
	ID    int    `json:"id"`
	Error string `json:"error,omitempty"`
	Data  string `json:"data,omitempty"`
}

type Dispatcher struct {
	batches  chan Image
	results  chan Result
	quit     chan bool
	received []Image
	mutex    sync.Mutex
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		batches:  make(chan Image), // Buffer up to 10 batches
		results:  make(chan Result),
		quit:     make(chan bool),
		received: []Image{},
	}
}

func (d *Dispatcher) Start() {
	fmt.Println("Starting dispatcher...")
	for {
		//add 获得batch size代码
		select {
		case batch := <-d.batches:
			fmt.Printf("Received image is %v\n", batch)
			d.mutex.Lock()
			d.received = append(d.received, batch)
			// Send batch to service if there are enough pictures
			if len(d.received) >= 4 {
				pics := d.received[:4]
				d.received = d.received[4:]
				go d.sendToService(pics)
			}
			d.mutex.Unlock()
		case result := <-d.results:
			fmt.Printf("Received result for batch %d\n", result.ID)
			// TODO: Send result to sender
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
}

func main() {
	dispatcher := NewDispatcher()
	go dispatcher.Start()

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

		dispatcher.batches <- batch
	})

	fmt.Println("Listening on port 8082...")
	http.ListenAndServe(":8082", nil)
}
