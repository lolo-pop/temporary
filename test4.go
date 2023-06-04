package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

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

	rps := [][]int{
		[]int{},
		[]int{},
		[]int{},
		[]int{},
		[]int{},
		[]int{},
	}

	var wg sync.WaitGroup
	index := 0
	timeLen := 100
	for index < timeLen {
		start := time.Now()
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				test(0)
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
