package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

const (
	dispatchURL = "http://localhost:8082/sendImage"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	currentRPS := 1

	for {
		for i := 0; i < currentRPS; i++ {
			go sendImage(i, currentRPS)
		}
		fmt.Println(currentRPS)
		time.Sleep(20 * time.Second)
	}
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
	resp, err := http.Post(dispatchURL, "application/octet-stream", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending image: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Image sent with status: %s\n", resp.Status)
}

func generateRandomImageData() *bytes.Reader {
	imageSize := 1000
	data := make([]byte, imageSize)
	rand.Read(data)
	return bytes.NewReader(data)
}
