package scaling

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

type Kv struct {
	Key   string
	Value float64
}

var (
	natsUrl        string
	metricsSubject string
	reqSubject     string
)

type PredictionRequest struct {
	FunctionName       string    `json:"function_name"`
	MonitoringSequence []float64 `json:"monitoring_sequence"`
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

func init() {
	var ok bool
	natsUrl, ok = os.LookupEnv("NATS_URL")
	if !ok {
		log.Fatal("$NATS_URL not set")
	}
	metricsSubject, ok = os.LookupEnv("METRICS_SUBJECT")
	if !ok {
		log.Fatal("$METRICS_SUBJECT not set")
	}
	reqSubject, ok = os.LookupEnv("REQ_SUBJECT")
	if !ok {
		log.Fatal("$REQ_SUBJECT not set")
	}

}

func Hello(name string) {
	fmt.Println("hello scaling:", name)
}

func FunctionAccuracyMapSort(acc map[string]float64) []Kv {

	var result []Kv
	for k, v := range acc {
		result = append(result, Kv{k, v})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Value > result[j].Value
	})
	// for _, kvpair := range result {
	//	fmt.Println(kvpair)
	// }
	return result
}

func GetLevel(acc float64, SCMap map[int][]int) (int, bool) {
	for level, limits := range SCMap {
		if acc < float64(limits[1]) && acc >= float64(limits[0]) {
			return level, true
		}
	}
	return -1, false
}
func ServiceContainerSLO(SCMap map[int][]int, functionAccuracy map[string]float64, functionLatency map[string]float64) map[int][]float64 {
	levelnum := len(SCMap)
	minlatency := make([]float64, levelnum)
	for i := 0; i < levelnum; i++ {
		minlatency[i] = 0
	}
	for fname, acc := range functionAccuracy {
		lat := functionLatency[fname]
		level, ok := GetLevel(acc, SCMap)
		if !ok {
			fmt.Printf("get function %s accuracy level failed", fname)
		}
		if minlatency[level] > lat {
			minlatency[level] = lat
		}
	}
	SCSLO := make(map[int][]float64)
	for i := 0; i < levelnum; i++ {
		SCSLO[i] = append(SCSLO[i], float64(SCMap[i][0]), float64(SCMap[i][1]), float64(minlatency[i]))
	}
	return SCSLO
}

func PredictFunctionRPS(functionName string, sequences []float64) (float64, error) {
	requestData := PredictionRequest{
		FunctionName:       functionName,
		MonitoringSequence: sequences,
	}

	// 将请求数据转换为 JSON 格式
	requestDataJson, err := json.Marshal(requestData)
	if err != nil {
		panic(err)
	}

	// 发送 POST 请求
	response, err := http.Post("http://localhost:5000/predict", "application/json", bytes.NewBuffer(requestDataJson))
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	// 读取响应主体中的 JSON 字符串
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	jsonStr := strings.TrimSpace(string(body))
	fmt.Println(jsonStr)
	var responseData PredictionResponse
	err = json.Unmarshal([]byte(string(jsonStr)), &responseData)
	if err != nil {
		return 0, err
	}
	return responseData.Mean, nil
}
