package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

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

type PredictionResponse struct {
	FunctionName string  `json:"function_name"`
	StartDate    string  `json:"start_date"`
	Quantile01   float64 `json:"quantile0.1"`
	Quantile02   float64 `json:"quantile0.2"`
	Quantile03   float64 `json:"quantile0.3"`
	Quantile04   float64 `json:"quantile0.4"`
	Quantile05   float64 `json:"quantile0.5"`
	Quantile06   float64 `json:"quantile0.6"`
	Quantile07   float64 `json:"quantile0.7"`
	Quantile08   float64 `json:"quantile0.8"`
	Quantile09   float64 `json:"quantile0.9"`
	Mean         float64 `json:"mean"`
}
type PredictionRequest struct {
	FunctionName       string    `json:"function_name"`
	MonitoringSequence []float64 `json:"monitoring_sequence"`
}
type ImageData struct {
	Name string `json:"name"`
	From string `json:"from"`
	Data string `json:"data"`
}

func test(level int) {
	start := time.Now()
	client := &http.Client{}
	serviceURL := fmt.Sprintf("http://127.0.0.1:8080/function/test-%d/", level)
	// serviceURL := "http://localhost:8081/processImages"
	// resp, err := http.Post(serviceURL, "application/json", bytes.NewBuffer(jsonData))
	req, err := http.NewRequest("POST", serviceURL, nil)
	if err != nil {
		errMsg := fmt.Sprintf("send to service failed: %s", err.Error())
		log.Fatal(errMsg)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending batch to service: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Batch sent with status: %d\n", resp.StatusCode)
	end := time.Since(start).Milliseconds()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println("Response Body:", string(body))
	if resp.StatusCode == 200 {
		file, err := os.OpenFile(fmt.Sprintf("data-%d.csv", level), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// 创建 CSV writer
		writer := csv.NewWriter(file)
		defer writer.Flush()

		res := strconv.FormatInt(end, 10)
		err = writer.Write([]string{res})
		if err != nil {
			panic(err)
		}
	} else {
		file, err := os.OpenFile(fmt.Sprintf("data-%d.csv", level), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		// 创建 CSV writer
		writer := csv.NewWriter(file)
		defer writer.Flush()

		res := "null"
		err = writer.Write([]string{res})
		if err != nil {
			panic(err)
		}
	}
}
func main() {
	g, err := strconv.Atoi(os.Args[1])
	if err != nil {

	}
	trace := map[int][]int{
		0: []int{2, 3, 1, 0, 2, 1, 3, 2, 0, 1, 3, 1, 2, 0, 3, 0, 1, 2, 3, 0, 2, 3, 1, 0, 2, 1, 3, 2, 0, 1, 3, 1, 2, 0, 3, 0, 1, 2, 3, 0, 2, 3, 1, 0, 2, 1, 3, 2, 0, 1, 3, 1, 2, 0, 3, 0, 1, 2, 3, 0},
		1: []int{1, 3, 0, 2, 2, 1, 3, 0, 2, 1, 2, 0, 3, 1, 0, 1, 3, 0, 3, 2, 1, 3, 0, 2, 2, 1, 3, 0, 2, 1, 2, 0, 3, 1, 0, 1, 3, 0, 3, 2, 1, 3, 0, 2, 2, 1, 3, 0, 2, 1, 2, 0, 3, 1, 0, 1, 3, 0, 3, 2},
		2: []int{0, 1, 2, 1, 2, 2, 1, 0, 2, 0, 1, 1, 0, 2, 0, 2, 1, 0, 1, 2, 0, 1, 2, 1, 2, 2, 1, 0, 2, 0, 1, 1, 0, 2, 0, 2, 1, 0, 1, 2, 0, 1, 2, 1, 2, 2, 1, 0, 2, 0, 1, 1, 0, 2, 0, 2, 1, 0, 1, 2},
		3: []int{0, 1, 0, 2, 0, 0, 1, 1, 2, 0, 0, 1, 2, 0, 1, 0, 0, 2, 0, 1, 0, 1, 0, 2, 0, 0, 1, 1, 2, 0, 0, 1, 2, 0, 1, 0, 0, 2, 0, 1, 0, 1, 0, 2, 0, 0, 1, 1, 2, 0, 0, 1, 2, 0, 1, 0, 0, 2, 0, 1},
		4: []int{2, 0, 0, 1, 1, 0, 1, 1, 0, 0, 1, 1, 0, 2, 0, 1, 0, 2, 0, 0, 1, 0, 0, 1, 1, 0, 1, 1, 0, 0, 2, 0, 0, 0, 0, 1, 0, 2, 0, 0, 2, 0, 0, 1, 1, 0, 1, 1, 0, 0, 2, 1, 1, 2, 0, 1, 0, 2, 0, 0},
		5: []int{0, 1, 1, 1, 0, 1, 0, 1, 0, 1, 0, 0, 2, 0, 0, 1, 1, 0, 0, 1, 0, 1, 0, 2, 0, 1, 0, 1, 0, 1, 1, 0, 0, 2, 0, 1, 1, 0, 0, 1, 0, 1, 0, 1, 0, 2, 0, 1, 0, 1, 1, 0, 0, 2, 0, 0, 1, 1, 0, 1},
	}
	var wg sync.WaitGroup
	index := 0
	timeLen := 600
	for index < timeLen {
		start := time.Now()
		for i := 0; i < trace[g][index%60]; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				test(g)
			}()
		}
		end := time.Since(start)
		fmt.Println(end)
		time.Sleep(time.Second * 1)
		index++
	}
	wg.Wait()
	fmt.Println("all goroutines are done")

}
