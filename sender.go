package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	dispatchURL = "http://localhost:8082/sendImage"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	currentRPS := 5

	for {
		for i := 0; i < currentRPS; i++ {
			go sendImage(i, currentRPS)
		}
		fmt.Println(currentRPS)
		time.Sleep(20 * time.Second)
		currentRPS++
	}
}

type image struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func sendImage(i int, f int) {
	// imageData := generateRandomImageData()
	tmp := fmt.Sprintf("image%d-%d.png", i, f)
	imageData := image{tmp, "image.png"}
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
