package main

import (
	"C"
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
	dispatchURL = "http://faas-dispatch.openfaas.svc.cluster.local:8080/sendImage"
)

func getResults(ch chan<- image) {
	server := &http.Server{Addr: ":8089"}
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
		ch <- result
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

//export Test
func Test(i C.int, f C.int) {
	start := time.Now()
	ch := make(chan image)
	go sendImage(0, 1)
	go getResults(ch)
	result := <-ch
	fmt.Println("test", result)
	end := time.Since(start)
	fmt.Println(end)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	// bs := 5

	Test(0, 6)

	return
}

func getContainerIP() string {
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
	tmp := fmt.Sprintf("image%d.png", i)
	ip := getContainerIP()
	log.Printf("ip:%s", ip)
	imageData := image{tmp, ip, tmp}
	jsonData, err := json.Marshal(imageData)
	if err != nil {
	}
	resp, err := http.Post(dispatchURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending image: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Image sent with status: %s\n", resp.Status)
}
