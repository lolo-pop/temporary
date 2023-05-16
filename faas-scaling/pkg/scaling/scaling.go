package scaling

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
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
		return 0, err
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

func zfill(str string, width int) string {
	for len(str) < width {
		str = "0" + str
	}

	return str
}
func Profile(path string) (map[string][]float64, error) {
	file, err := os.Open("/home/rongch05/openfaas/faas-scaling/profiling.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建CSV Reader对象
	reader := csv.NewReader(file)

	// 读取CSV文件中的所有记录
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// 将CSV记录解析为字典
	results := make(map[string][]float64)
	for i, record := range records {
		if i == 0 {
			continue
		} else {
			configuration := record[0] + zfill(record[1], 4) + zfill(record[2], 2) + record[3]
			acc, err := strconv.ParseFloat(record[4], 64)
			if err != nil {
				return nil, err
			}
			lat1, err := strconv.ParseFloat(record[5], 64)
			if err != nil {
				return nil, err
			}
			lat2, err := strconv.ParseFloat(record[6], 64)
			if err != nil {
				return nil, err
			}
			results[configuration] = []float64{acc, lat1, lat2}
		}
	}
	return results, nil
}

func LowRPS(SCProfile map[string][]float64, level int, cpu map[string][]float64, mem map[string][]float64, batch map[string]int, LatSLO float64) (float64, error) {
	totalRPS := 0.0
	for podName := range cpu {
		cpuLimits := int(cpu[podName][1] / 1000)
		memLimits := int(mem[podName][1] / 1024 / 1024)
		batchSize := int(batch[podName])

		// config: model 1个字符，memory4个字符，cpu 2个字符，batch 1个字符
		config := strconv.Itoa(level) + zfill(strconv.Itoa(memLimits), 4) + zfill(strconv.Itoa(cpuLimits), 2) + strconv.Itoa(batchSize)
		log.Printf("config is %s in lowRPS function:", config)
		latency := SCProfile[config][1]
		rps := 1 / (LatSLO - latency) * float64(batchSize)
		totalRPS += rps
	}
	return totalRPS, nil
}

func UpRPS(SCProfile map[string][]float64, level int, cpu map[string][]float64, mem map[string][]float64, batch map[string]int) (float64, error) {
	totalRPS := 0.0
	for podName := range cpu {
		cpuLimits := int(cpu[podName][1] / 1024)
		memLimits := int(mem[podName][1] / 1024 / 1024)
		batchSize := int(batch[podName])
		config := strconv.Itoa(level) + zfill(strconv.Itoa(memLimits), 4) + zfill(strconv.Itoa(cpuLimits), 2) + strconv.Itoa(batchSize)
		log.Printf("config is %s in UpRPS function:", config)
		latency := SCProfile[config][1]
		rps := 1 / latency * float64(batchSize)
		totalRPS += rps
	}
	return totalRPS, nil
}

func WarmupFunction(level int, rps float64, wg *sync.WaitGroup, m *sync.Map, functionName string) (string, error) {
	return functionName, nil
}

func RemoveFunction(level int, rps float64, wg *sync.WaitGroup, m *sync.Map, functionName string) (string, error) {
	return functionName, nil
}
