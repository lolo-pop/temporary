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
	"time"

	"github.com/go-redis/redis"
	"github.com/lolo-pop/faas-scaling/pkg/types"
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
		minlatency[i] = 10000
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
	// log.Printf("sequences in PredictFunctionRPS is %v", sequences)
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
	file, err := os.Open(path) //这里的路径需要修改
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
		config_1 := strconv.Itoa(level) + zfill(strconv.Itoa(memLimits), 4) + zfill(strconv.Itoa(cpuLimits), 2) + strconv.Itoa(1)
		latency_1 := SCProfile[config_1][1]
		// rps := 1 / (LatSLO - latency) * float64(batchSize)
		rps := 1 / (LatSLO - latency + latency_1) * 1
		totalRPS += rps
	}
	return totalRPS, nil
}

func UpRPS(SCProfile map[string][]float64, level int, cpu map[string][]float64, mem map[string][]float64, batch map[string]int) (float64, error) {
	totalRPS := 0.0
	for podName := range cpu {
		cpuLimits := int(cpu[podName][1] / 1000)
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

func feasibleSet(cpu []int, mem []int, bs []int, level int, rps float64, SCProfile map[string][]float64, LatSLO float64) [][]float64 {
	var res [][]float64
	for _, b := range bs {
		for _, c := range cpu {
			for _, m := range mem {
				config := strconv.Itoa(level) + zfill(strconv.Itoa(m), 4) + zfill(strconv.Itoa(c), 2) + strconv.Itoa(b)
				if _, ok := SCProfile[config]; ok {
					profile := SCProfile[config]
					execTime := profile[1]
					waitTime := float64(b) / rps
					lowRps := 1 / (LatSLO - execTime) * float64(b)
					upRps := 1 / execTime * float64(b)
					if execTime+waitTime < LatSLO && rps > lowRps {
						res = append(res, []float64{float64(b), float64(c), float64(m), lowRps, upRps})
					}
				} else {
					continue
				}
			}
		}
	}
	return res
}

type ResourceEfficient struct {
	config    []float64
	efficient float64
}

func resEfficient(alpha float64, config [][]float64) ([]ResourceEfficient, ResourceEfficient) {
	maxEfficient := 0.0
	maxConfig := 0
	var res []ResourceEfficient
	for i, cnf := range config {
		efficient := cnf[4] / (cnf[1]*alpha + cnf[2])
		if efficient > maxEfficient {
			maxEfficient = efficient
			maxConfig = i
		}
		res = append(res, ResourceEfficient{cnf, efficient})
	}
	maxres := ResourceEfficient{config[maxConfig], maxEfficient}
	return res, maxres
}

func instancePlacement(alpha float64, nodes []types.Node, level int, config []float64) (string, []types.Node) {
	// input: 当前集群使用率
	pI := 0
	minIndex := 1000.0
	for i, node := range nodes {
		// nodeName := node.Name
		usageCpu := node.Cpu[0]
		capacityCpu := node.Cpu[1]
		usageMem := node.Mem[0]
		capacityMem := node.Mem[1]
		index := alpha*(usageCpu/capacityCpu) + usageMem/capacityMem // 这里需要添加判定，node的资源能否承载
		if index < minIndex {
			pI = i
			minIndex = index
		}
	}
	nodeName := nodes[pI].Name
	cpu := config[1]
	mem := config[2]
	nodes[pI].Cpu[0] = nodes[pI].Cpu[0] + cpu*1000
	nodes[pI].Mem[0] = nodes[pI].Mem[0] + mem*1024
	return nodeName, nodes
}

type Scheduler struct {
	nodeName string
	config   []float64
}

func Scheduling(alpha float64, cpu []int, mem []int, bs []int, level int, rps float64, nodes []types.Node, SCProfile map[string][]float64, LatSLO float64, labeledRemoveFunction []types.SCconfig) ([]types.SCconfig, []Scheduler, int, error) {
	curRps := rps
	curNodes := nodes
	n := 0
	var schedulerRes []Scheduler
	tmp := []int{}
	for curRps > 0 && len(tmp) < len(labeledRemoveFunction) && len(labeledRemoveFunction) > 0 {
		maxRps := -1.0
		maxI := 0
		for i, removeFunction := range labeledRemoveFunction {
			found := false
			for _, index := range tmp {
				if index == i {
					found = true
					break
				}
			}
			if found {
				continue
			} else {
				if removeFunction.UpRps > maxRps {
					maxRps = removeFunction.UpRps
					maxI = i
				}
			}
		}
		curRps = curRps - maxRps
		tmp = append(tmp, maxI)
	}
	labeledActiveFunction := []types.SCconfig{}
	for _, index := range tmp {
		labeledActiveFunction = append(labeledActiveFunction, labeledRemoveFunction[index])
	}

	if curRps > 0 {
		for curRps > 0 {
			feasibleConfig := feasibleSet(cpu, mem, bs, level, curRps, SCProfile, LatSLO)
			if len(feasibleConfig) == 0 {
				log.Fatalln("feasible config set is null")
			} else {
				_, maxEfficient := resEfficient(alpha, feasibleConfig)
				n = n + 1
				instanceRps := maxEfficient.config[4] // upRps
				curRps = curRps - instanceRps
				nodeName := ""
				nodeName, curNodes = instancePlacement(alpha, curNodes, level, maxEfficient.config)
				schedulerRes = append(schedulerRes, Scheduler{nodeName, maxEfficient.config})
			}
		}
	}

	return labeledActiveFunction, schedulerRes, n, nil
}

func functionNameHash(curSCfunctionName []string, level int) ([]int, error) {
	nameHash := make([]int, 10000)
	for _, item := range curSCfunctionName {
		indexStr := strings.Split(item, "-")[2]
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return nameHash, err
		}
		nameHash[index] = 1
	}
	return nameHash, nil
}

func deployFunction(level int, index int, serviceContainerImage map[int]string, functionConfig Scheduler) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	functionName := fmt.Sprintf("service-%d-%d", level, index)
	imageName := serviceContainerImage[level]
	node := functionConfig.nodeName
	batchSize := strconv.Itoa(int(functionConfig.config[0]))
	cpu := strconv.Itoa(int(functionConfig.config[1])*1000) + "m"
	mem := strconv.Itoa(int(functionConfig.config[2])) + "Mi"
	placementLabel := fmt.Sprintf("kubernetes.io/hostname=%s", node)
	gatewayURL := "http://gateway.openfaas.svc.cluster.local:8080"
	requestData := map[string]interface{}{
		"service": functionName,
		"image":   imageName,
		"envVars": map[string]string{
			"BATCH_SIZE":   batchSize,
			"NATS_ADDRESS": "http://nats.openfaas.svc.cluster.local:4222",
			"NATS_SUBJECT": "image-test",
			"RESOLUTION":   "512x512",
		},
		"envProcess": "python3 index.py",
		"limits": map[string]string{
			"memory": mem,
			"cpu":    cpu,
		},
		"requests": map[string]string{
			"memory": mem,
			"cpu":    cpu,
		},
		"labels": map[string]string{
			"com.openfaas.scale.zero": "true",
			"com.openfaas.scale.min":  "1",
			"com.openfaas.scale.max":  "1",
			"instance.idle":           "false",
		},
		"constraints": []string{placementLabel},
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		fmt.Printf("Error marshaling JSON request body: %v", err)
		return functionName, err
	}

	// 构造要发送的请求
	req, err := http.NewRequest("POST", gatewayURL+"/system/functions", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error creating HTTP request: %v", err)
		return functionName, err
	}

	req.Header.Set("Content-Type", "application/json")
	user := "admin"
	password := "admin"
	req.SetBasicAuth(user, password)
	// 发送请求并获取响应
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending HTTP request: %v", err)
		return functionName, err
	}
	defer resp.Body.Close()

	/*
		//  这里需要debug
		// 检查响应状态码
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Unexpected response status code: %d", resp.StatusCode)
		}

		// 解析响应的 JSON 数据
		var responseMap map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&responseMap)
		if err != nil {
			fmt.Printf("Error decoding JSON response body: %v", err)
			return err
		}
	*/
	fmt.Printf("Function %s deployed, response StatusCode is %d\n", functionName, resp.StatusCode)
	// fmt.Printf("Function %s deployed successfully\n", functionName)
	return functionName, nil
}

func WarmupInstace(schedulerRes []Scheduler, replicaNum int, curSCfunctionName []string, level int, serviceContainerImage map[int]string) ([]types.SCconfig, error) {
	index := 0
	var functionInstanceStatus []types.SCconfig
	nameHash, err := functionNameHash(curSCfunctionName, level)
	if err != nil {
		msg := fmt.Sprintf("functionNameHash failed: %s", err.Error())
		log.Fatal(msg)
	}
	i := 0
	for i < replicaNum {
		if nameHash[index] == 1 {
			index = index + 1
		} else {
			functionName, err := deployFunction(level, index, serviceContainerImage, schedulerRes[i])
			if err != nil {
				msg := fmt.Sprintf("deploy function failed: %s", err.Error())
				log.Fatal(msg)
				return nil, err
			}
			nameHash[index] = 1
			index = index + 1
			i = i + 1
			bs := schedulerRes[i].config[0]
			c := schedulerRes[i].config[1]
			m := schedulerRes[i].config[2]
			lowRps := schedulerRes[i].config[3]
			upRps := schedulerRes[i].config[4]
			node := schedulerRes[i].nodeName
			var config types.SCconfig
			config.BatchSize = int(bs)
			config.Cpu = c
			config.Mem = m
			config.LowRps = lowRps
			config.UpRps = upRps
			config.Node = node
			config.Name = functionName
			functionInstanceStatus = append(functionInstanceStatus, config)
		}
	}
	fmt.Printf("have warmed up %d service-container-%d instance", replicaNum, level)
	return functionInstanceStatus, nil
}

func MaxUsage(nodes []types.Node, alpha float64) (string, float64) {
	maxUsage := 0.0
	maxNode := ""
	for _, node := range nodes {
		nodeName := node.Name
		nodeCpuUsage := node.Cpu[0] / node.Cpu[1]
		nodeMemUsage := node.Mem[0] / node.Mem[1]
		usage := nodeCpuUsage*alpha + nodeMemUsage
		if usage > maxUsage {
			maxUsage = usage
			maxNode = nodeName
		}
	}
	return maxNode, maxUsage
}

func RemoveFunction(index float64, alpha float64, level int, rps float64, nodes []types.Node, SCProfile map[string][]float64, LatSLO float64, nodesActiveFunctionStatus map[string][]types.SCconfig) ([]types.SCconfig, error) {
	var labelRemoveFunction []types.SCconfig
	for rps > 0 {
		minResEffi := 1000.0
		maxNodeName, _ := MaxUsage(nodes, alpha)
		var minSC types.SCconfig
		minLowBound := 0.0
		for _, sc := range nodesActiveFunctionStatus[maxNodeName] {
			m := int(sc.Mem)
			c := int(sc.Cpu)
			b := sc.BatchSize
			config := strconv.Itoa(level) + zfill(strconv.Itoa(m), 4) + zfill(strconv.Itoa(c), 2) + strconv.Itoa(b)
			profile := SCProfile[config]
			execTime := profile[1]
			// waitTime := float64(b) / rps
			lowRps := 1 / (LatSLO - execTime) * float64(b)
			upRps := 1 / execTime * float64(b)
			lowbound := (upRps-lowRps)*index + lowRps
			if lowbound/float64(alpha*sc.Cpu+sc.Mem) < minResEffi {
				minResEffi = lowbound / float64(alpha*sc.Cpu+sc.Mem)
				minSC.Name = sc.Name
				minSC.BatchSize = sc.BatchSize
				minSC.Cpu = sc.Cpu
				minSC.Mem = sc.Mem
				//minSC.LowRps = lowRps
				//minSC.UpRps = upRps
				minSC.Node = maxNodeName
				minLowBound = lowbound
			}
		}
		rps = rps - minLowBound

		// 更新 nodes的资源使用状态
		for _, node := range nodes {
			if node.Name == maxNodeName {
				node.Cpu[0] = node.Cpu[0] - minSC.Cpu*1000
				node.Mem[0] = node.Mem[0] - minSC.Mem*1024
			}
		}
		labelRemoveFunction = append(labelRemoveFunction, minSC)
	}
	return labelRemoveFunction, nil
}

func setRedis(client *redis.Client, key string, value []types.SCconfig) error {
	configsJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = client.Set(key, string(configsJSON), 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func getRedis(client *redis.Client, key string) ([]types.SCconfig, error) {
	val, err := client.Get(key).Result()
	if err != nil {
		return nil, err
	}
	var value []types.SCconfig
	err = json.Unmarshal([]byte(val), &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func GetSCRemoveFunction(key string, redisUrl string, redisPassword string) ([]types.SCconfig, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})
	curRemoveFunction := []types.SCconfig{}
	exists, err := client.Exists(key).Result()
	if err != nil {
		return curRemoveFunction, err
	}
	if exists == 1 {
		curRemoveFunction, err := getRedis(client, key)
		if err != nil {
			return curRemoveFunction, err
		}
	} else {
		err := setRedis(client, key, curRemoveFunction)
		if err != nil {
			return curRemoveFunction, err
		}
	}
	return curRemoveFunction, nil
}

func StoreFunctionInWarmup(curRemoveKey string, nextRemoveKey string, WarmupKey string, nextActiveFunction []types.SCconfig, warmupFunction []types.SCconfig, redisUrl string, redisPassword string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})
	exists, err := client.Exists(nextRemoveKey).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		curRemoveFunction, err := getRedis(client, curRemoveKey)
		if err != nil {
			msg := fmt.Sprintf("get key %s failed: %s", curRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		tmp, err := getRedis(client, nextRemoveKey)
		if err != nil {
			msg := fmt.Sprintf("get key %s failed: %s", nextRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		err = setRedis(client, curRemoveKey, tmp)
		if err != nil {
			msg := fmt.Sprintf("set key %s failed: %s", curRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		if len(nextActiveFunction) > 0 {
			nextRemoveFunction := []types.SCconfig{}
			for _, removefunction := range curRemoveFunction {
				found := false
				for _, activefunction := range nextActiveFunction {
					if activefunction.Name == removefunction.Name {
						found = true
						break
					}
				}
				if found {
					continue
				} else {
					nextRemoveFunction = append(nextRemoveFunction, removefunction)
				}
			}
			err := setRedis(client, nextRemoveKey, nextRemoveFunction)
			if err != nil {
				msg := fmt.Sprintf("set key %s failed: %s", nextRemoveKey, err.Error())
				log.Fatal(msg)
				return err
			}
		}

	} else {
		nextRemoveFunction := []types.SCconfig{}
		err := setRedis(client, nextRemoveKey, nextRemoveFunction)
		if err != nil {
			msg := fmt.Sprintf("set key %s failed: %s", nextRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		err = setRedis(client, curRemoveKey, nextRemoveFunction)
		if err != nil {
			msg := fmt.Sprintf("set key %s failed: %s", curRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
	}
	err = setRedis(client, WarmupKey, warmupFunction)
	if err != nil {
		msg := fmt.Sprintf("set key %s failed: %s", WarmupKey, err.Error())
		log.Fatal(msg)
		return err
	}
	return nil
}

func StoreFunctionInRemove(curRemoveKey string, nextRemoveKey string, WarmupKey string, newRemoveFunction []types.SCconfig, redisUrl string, redisPassword string) error {
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})
	exists, err := client.Exists(nextRemoveKey).Result()
	if err != nil {
		return err
	}
	if exists == 1 {
		tmp, err := getRedis(client, nextRemoveKey)
		if err != nil {
			msg := fmt.Sprintf("get key %s failed: %s", nextRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		nextRemoveFunction, err := getRedis(client, curRemoveKey)
		if err != nil {
			msg := fmt.Sprintf("get key %s failed: %s", curRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		for _, removeFunction := range newRemoveFunction {
			nextRemoveFunction = append(nextRemoveFunction, removeFunction)
		}
		err = setRedis(client, nextRemoveKey, nextRemoveFunction)
		if err != nil {
			msg := fmt.Sprintf("set key %s failed: %s", nextRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		err = setRedis(client, curRemoveKey, tmp)
		if err != nil {
			msg := fmt.Sprintf("set key %s failed: %s", curRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
	} else {
		nextRemoveFunction := []types.SCconfig{}
		err := setRedis(client, nextRemoveKey, nextRemoveFunction)
		if err != nil {
			msg := fmt.Sprintf("set key %s failed: %s", nextRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
		err = setRedis(client, curRemoveKey, nextRemoveFunction)
		if err != nil {
			msg := fmt.Sprintf("set key %s failed: %s", curRemoveKey, err.Error())
			log.Fatal(msg)
			return err
		}
	}
	return nil
}

func SetActiveFunctons(Functions map[int][]types.SCconfig, keyName string, redisUrl string, redisPassword string, levelNum int) error {
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})

	for i := 0; i < levelNum; i++ {
		k := fmt.Sprintf("%s-%d", keyName, i)
		configsJSON, err := json.Marshal(Functions[i])
		if err != nil {
			return err
		}
		err = client.Set(k, string(configsJSON), 0).Err()
		if err != nil {
			return err
		}
	}
	/*
		for key, value := range Functions {
			k := fmt.Sprintf("%s-%d", keyName, key)
			configsJSON, err := json.Marshal(value)
			if err != nil {
				return err
			}
			err = client.Set(k, string(configsJSON), 0).Err()
			if err != nil {
				return err
			}
		}
	*/
	return nil
}
func GetFunctionSLO(functionName string, redisUrl string, redisPassword string) (map[string]float64, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl,
		Password: redisPassword,
		DB:       0,
	})

	val, err := client.Get(functionName).Result()
	if err != nil {
		return nil, err
	}
	var value map[string]float64
	err = json.Unmarshal([]byte(val), &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}
