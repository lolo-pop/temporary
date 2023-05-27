package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"
)

const (
	dispatchURL = "http://localhost:8082/sendImage"
)

func getResults(ch chan<- image) {
	server := &http.Server{Addr: ":8084"}
	http.HandleFunc("/sendResult", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var result image
		err := json.NewDecoder(r.Body).Decode(&result)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		// from := batch.From
		ch <- result

		// fmt.Printf("current results is %v\n", dispatcher.rcvResults)
		w.WriteHeader(http.StatusOK)
		go func() {
			err := server.Shutdown(context.Background())
			if err != nil {
				log.Fatal(err)
			}
		}()
	})
	server.ListenAndServe()
}
func test(i int, f int) {

	ch := make(chan image)
	go sendImage(0, 1)
	go getResults(ch)
	result := <-ch
	fmt.Println("test", result)
}
func main() {
	rand.Seed(time.Now().UnixNano())
	bs := 5
	test(0, bs)
	return
}
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

type image struct {
	Name string `json:"name"`
	From string `json:"from"`
	Data string `json:"data"`
}

func sendImage(i int, f int) {
	// imageData := generateRandomImageData()
	tmp := fmt.Sprintf("image%d-%d.png", i, f)
	imageData := image{tmp, getContainerIP(), "image.png"}
	jsonData, err := json.Marshal(imageData)
	if err != nil {
		// 处理错误
	}
	resp, err := http.Post(dispatchURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending image: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Image sent with status: %s\n", resp.Status)
}
