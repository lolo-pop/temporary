package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lolo-pop/faas-scaling/pkg/metrics"
	// "github.com/lolo-pop/faas-scaling/pkg/nats"
	"github.com/lolo-pop/faas-scaling/pkg/scaling"
	"github.com/lolo-pop/faas-scaling/pkg/types"
	"github.com/nats-io/nats.go"
)

var (
	natsUrl               string
	metricsSubject        string
	reqSubject            string
	scalingWindows        int64
	redisUrl              string
	redisPassword         string
	serviceContainerImage map[int]string
)

type SLO struct {
	Accuracy float64
	latency  float64
}

func init() {
	serviceContainerImage = map[int]string{
		1: "lolopop/service-container-1:latest",
		2: "lolopop/service-container-2:latest",
		3: "lolopop/service-container-3:latest",
		4: "lolopop/service-container-4:latest",
		5: "lolopop/service-container-5:latest",
	}
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
	env, ok := os.LookupEnv("SCALING_WINDOWS")
	if !ok {
		log.Fatal("$scaling windows not set")
	}
	var err error
	val, err := strconv.Atoi(env)
	if err != nil {
		log.Fatal(err.Error())
	}
	scalingWindows = int64(val)

	redisUrl, ok = os.LookupEnv("REDIS_URL")
	if !ok {
		log.Fatal("$REDIS_URL not set")
	}
	redisPassword, ok = os.LookupEnv("REDIS_PASS")
	if !ok {
		log.Fatal("$REDIS_PASS not set")
	}
}

/*
	func getFunctionAccRequire(sortedFunctionAccuracyMap []scaling.Kv, functionName string) float64 {
		for _, functionPair := range sortedFunctionAccuracyMap {
			if functionPair.Key == functionName {
				return functionPair.Value
			}
		}
		log.Fatalf("Cannot get %s function's accuracy requirement", functionName)
		return 0
	}
*/

func main() {
	scaling.Hello("test")
	rand.Seed(time.Now().UnixNano())
	accuracy := [20]float64{22.226, 29.066, 29.981, 31.939, 25.687,
		31.391, 32.991, 26.094, 26.303, 23.245,
		28.526, 23.302, 28.489, 33.799, 31.171,
		33.15, 30.037, 24.051, 29.817, 27.067}
	latency := [20]float64{0.667, 0.901, 0.676, 0.663, 0.822,
		0.776, 0.720, 0.851, 0.852, 0.662,
		0.759, 0.839, 0.612, 0.801, 0.790,
		0.668, 0.654, 0.760, 0.690, 0.853}

	SCMap := map[int][]int{
		0: []int{22, 24},
		1: []int{24, 26},
		2: []int{26, 28},
		3: []int{28, 30},
		4: []int{30, 32},
		5: []int{32, 34},
	}
	functionAccuracy := make(map[string]float64)
	functionLatency := make(map[string]float64)
	indexLevel := 0
	levelNum := 6
	profilingPath := "profiling.csv"

	// make(map[string][]float64)  f
	SCProfile, err := scaling.Profile(profilingPath) // make(map[string][]float64)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot parse profiling results: %s", err)
		log.Fatal(errMsg)
	}

	// 连接NATS并订阅metrics subject
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot connect to nats: %s", err.Error())
		log.Fatal(errMsg)
	}
	defer nc.Close()
	sub, err := nc.SubscribeSync(metricsSubject)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot subscribe %s subject: %s", metricsSubject, err)
		log.Fatal(errMsg)
	}
	defer sub.Unsubscribe()

	var preFunctionRPS map[string][]float64 // 所有函数历史RPS 监测数据 functionName: RPS slice

	// accuracy 和function name的对应关系需要确定是否是固定的。
	for {
		msg, err := sub.NextMsg(0)
		var p metrics.Provider
		p = &metrics.FaasProvider{}
		functionNames, err := p.Functions()
		if err != nil {
			log.Fatal(err.Error())
		}
		// var upSCRPS map[int]float64               // 当前副本service conatiner 所能承受的最高RPS,
		// var lowSCRPS map[int]float64              // 当前副本service conatiner 所能承受的最低RPS,
		// var predictFunctionRPS map[string]float64 // 下一个时间窗口的所有函数的RPS的预测值 functionName: RPS
		// var predictSCRPS map[int]float64 // 下一个时间窗口的service container的RPS 预测值 accuracyLevel[int]: RPS

		predictFunctionRPS := make(map[string]float64)
		predictSCRPS := make(map[int]float64)
		//初始化SC RPS的预测值，所有值为0
		for i := 0; i < levelNum; i++ {
			predictSCRPS[i] = 0.0
		}

		//为每个函数映射对应的准确度
		for _, fname := range functionNames {
			if strings.Contains(fname, "service") {
				continue
			} else if _, ok := functionAccuracy[fname]; !ok {
				functionAccuracy[fname] = accuracy[indexLevel]
				functionLatency[fname] = latency[indexLevel]
				indexLevel += 1
			}
		}
		fmt.Println("current function-accuracy configration:", functionAccuracy)
		fmt.Println("current function-latency configration:", functionLatency)
		// 对所有非service container 按照准确度要求进行排序
		var sortedFunctionAccuracyMap []scaling.Kv
		sortedFunctionAccuracyMap = scaling.FunctionAccuracyMapSort(functionAccuracy)
		for _, funcAccPair := range sortedFunctionAccuracyMap {
			fmt.Println(funcAccPair)
		}
		// 根据所有function latency and accuracy requirment 计算每个service container的 latency SLO
		serviceContainerSLO := scaling.ServiceContainerSLO(SCMap, functionAccuracy, functionLatency) //  {level: [acc_low, acc_high, latency]}
		fmt.Printf("service container SLO: %v\n", serviceContainerSLO)

		//反序列从NATS获得的metrics
		var metrics types.Message
		err = json.Unmarshal(msg.Data, &metrics)
		if err != nil {
			errMsg := fmt.Sprintf("Cannot unmarshal message: %s", err.Error())
			log.Fatal(errMsg)
		}
		fmt.Printf("Timestamp: %d", metrics.Timestamp)

		// 为每个function预测下一个时间窗口的RPS
		for _, function := range metrics.Functions {
			functionName := function.Name
			if strings.Contains(functionName, "service") {
				continue
			}
			if _, ok := preFunctionRPS[functionName]; ok {
				preFunctionRPS[functionName] = append(preFunctionRPS[functionName], function.InvocationRate)
				fmt.Printf("current RPS monitor sequence of function %s: %v", functionName, preFunctionRPS[functionName]) //输出list可能会出错，需要主语 debug
			} else {
				preFunctionRPS[functionName] = []float64{function.InvocationRate}
			}
			predictFunctionRPS[functionName], err = scaling.PredictFunctionRPS(functionName, preFunctionRPS[functionName]) // 计算function的RPS预测值
			if err != nil {
				errMsg := fmt.Sprintf("Predicting the RPS of function %s fails: %s", functionName, err)
				log.Fatal(errMsg)
			}
			level, ok := scaling.GetLevel(functionAccuracy[functionName], SCMap) //计算function 属于哪个level
			if !ok {
				errMsg := fmt.Sprintf("get function %s accuracy level failed", functionName)
				log.Fatal(errMsg)
			}
			predictSCRPS[level] += predictFunctionRPS[functionName] //计算SC的RPS的预测值
		}

		// 获得当前系统内被标记为remove的function信息
		labeledRemoveFunction := make(map[int][]types.SCconfig) // 从redis中读取到被标记为remove的function的信息
		for i := 0; i < levelNum; i++ {
			key := fmt.Sprintf("curRemoveFunction-%d", i)
			tmp, err := scaling.GetSCRemoveFunction(key, redisUrl, redisPassword)
			if err != nil {
				msg := fmt.Sprintf("get key %s failed in GetSCRemoveFunction: %s", key, err.Error())
				log.Fatal(msg)
			}
			labeledRemoveFunction[i] = tmp
		}

		// 计算当前系统里存在的service container replicas 所能承担的 RPS的上限和下限。
		// 判断某个function 的资源状况需要先判断 function.Replicas是否为0，如果为0即不存在副本，资源状况也是空的
		//每一个准确度level 对应openfaas多个function（service container），为了控制environment，一个SC instance对应一个function
		serviceContainerName := make(map[int][]string)
		upSCRPS := make(map[int]float64)
		lowSCRPS := make(map[int]float64)
		nodesActiveFunctionStatus := make(map[string][]types.SCconfig) //每个node存放的没有被标记为remove的service container的具体信息
		for _, function := range metrics.Functions {
			functionName := function.Name
			if strings.Contains(functionName, "service") {
				fields := strings.Split(functionName, "-") // service-1-num-random
				accuracyLevel, err := strconv.Atoi(fields[1])
				serviceContainerName[accuracyLevel] = append(serviceContainerName[accuracyLevel], functionName) //这个需要包括 被标记为remove的functionName
				if err != nil {
					log.Fatalf("Failed to convert accuracy level to int: %s", fields[1])
				}
				found := false
				for _, status := range labeledRemoveFunction[accuracyLevel] {
					if status.Name == functionName {
						found = true
						break
					}
				}
				if found {
					continue
				} else {
					lowRps := 0.0
					upRps := 0.0

					if function.Replicas == 0 {
						fmt.Printf("service container %s do not have a running replica", functionName)
					} else if function.Replicas == 1 {
						lowRps, err = scaling.LowRPS(SCProfile, accuracyLevel, function.Cpu, function.Mem, function.Batch, serviceContainerSLO[accuracyLevel][2])
						if err != nil {
							errMsg := fmt.Sprintf("get low RPS failed: %s", err)
							log.Fatalf(errMsg)
						}
						upRps, err = scaling.UpRPS(SCProfile, accuracyLevel, function.Cpu, function.Mem, function.Batch)
						if err != nil {
							errMsg := fmt.Sprintf("get up RPS failed: %s", err.Error())
							log.Fatalf(errMsg)
						}

						// 获得每个node中sc function的信息
						node := function.Nodes[0]
						cpuResource := function.Cpu
						bs := 0
						c := 0.0
						m := 0.0
						for podName, cpu := range cpuResource {
							bs = function.Batch[podName]
							c = cpu[1]
							m = function.Mem[podName][1]
						}
						var tmp types.SCconfig
						tmp.Cpu = c
						tmp.Name = functionName
						tmp.Mem = m
						tmp.BatchSize = bs
						tmp.Node = node
						tmp.LowRps = lowRps
						tmp.UpRps = upRps
						nodesActiveFunctionStatus[node] = append(nodesActiveFunctionStatus[node], tmp)
					} else {
						log.Printf("Error, service container %s have %d replicas\n", functionName, function.Replicas)
					}
					lowSCRPS[accuracyLevel] += lowRps
					upSCRPS[accuracyLevel] += upRps
				}
			}
		}
		/*
			for _, function := range metrics.Functions {
				functionName := function.Name
				if strings.Contains(functionName, "service") {
					if function.Replicas == 1 {
						node := function.Nodes[0]
						cpuResource := function.Cpu
						bs := 0
						c := 0.0
						m := 0.0
						for podName, cpu := range cpuResource {
							bs = function.Batch[podName]
							c = cpu[1]
							m = function.Mem[podName][1]
						}
						var tmp types.SCconfig
						tmp.Cpu = c
						tmp.Name = functionName
						tmp.Mem = m
						tmp.BatchSize = bs
						tmp.Node = node
						nodesStatus[node] = append(nodesStatus[node], tmp)
					} else {
						fmt.Printf("Error, service container %s have %d replicas", functionName, function.Replicas)
					}
				}
			}
		*/
		//predictSCRPS 和 UpSCRPS的差值
		index := 0.8
		bs := []int{1, 2, 4, 8}
		cpu := []int{2, 4, 6, 8, 10, 12, 14, 16}
		mem := []int{256, 512, 1024, 2048, 4096, 6144, 8192}
		alpha := float64(128 / 30)
		var wg sync.WaitGroup
		for level, rps := range predictSCRPS {
			lowBound := (upSCRPS[level]-lowSCRPS[level])*index + lowSCRPS[level]
			upBound := upSCRPS[level]
			if rps > upBound {
				wg.Add(1)
				go func(level int, rps float64) {
					defer wg.Done()
					labeledActiveFunction, schedulerRes, replicaNum, err := scaling.Scheduling(alpha, cpu, mem, bs, level, rps-upBound, metrics.Nodes, SCProfile, serviceContainerSLO[level][2], labeledRemoveFunction[level])
					if err != nil {
						errMsg := fmt.Sprintf("scheduling failed: %s", err.Error())
						log.Fatalf(errMsg)
					}
					warmupfunction, err := scaling.WarmupInstace(schedulerRes, replicaNum, serviceContainerName[level], level, serviceContainerImage)
					if err != nil {
						errMsg := fmt.Sprintf("warmupFunction failed: %s", err.Error())
						log.Fatalf(errMsg)
					}
					err = scaling.StoreFunctionInWarmup(fmt.Sprintf("curRemoveFunction-%d", level), fmt.Sprintf("nextRemoveFunction-%d", level), fmt.Sprintf("warmupFunction-%d", level), labeledActiveFunction, warmupfunction, redisUrl, redisPassword)
					if err != nil {
						errMsg := fmt.Sprintf("StoreKeyValue failed in warmupfunction, level %d: %s", level, err.Error())
						log.Fatalf(errMsg)
					}
					log.Printf("warmup service-%d function succeeded, warmed up %v function replicas", level, len(warmupfunction))

				}(level, rps)
				// go scaling.warmupFunction(level, rps-upBound, &wg, &m, serviceContainerName[level])
			} else if rps < lowBound {
				wg.Add(1)
				go func(level int, rps float64) {
					defer wg.Done()
					newRemoveFunction, err := scaling.RemoveFunction(index, alpha, level, rps-upBound, metrics.Nodes, SCProfile, serviceContainerSLO[level][2], nodesActiveFunctionStatus)
					if err != nil {
						errMsg := fmt.Sprintf("removeFunction failed: %s", err.Error())
						log.Fatalf(errMsg)
					}

					// curRemoveFunction-level 存储的是当前windows被标记移出的函数
					err = scaling.StoreFunctionInRemove(fmt.Sprintf("curRemoveFunction-%d", level), fmt.Sprintf("nextRemoveFunction-%d", level), fmt.Sprintf("warmupFunction-%d", level), newRemoveFunction, redisUrl, redisPassword)
					if err != nil {
						errMsg := fmt.Sprintf("StoreKeyValue failed in removefunction, level %d: %s", level, err.Error())
						log.Fatalf(errMsg)
					}
					log.Printf("level %d function, remove %d function replicas", level, len(newRemoveFunction))
				}(level, rps)
				// go scaling.removeFunction(level, lowBound-rps, &wg, &m, serviceContainerName[level])
			}
		}
		wg.Wait()
	}
}
