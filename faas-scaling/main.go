package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/lolo-pop/faas-scaling/pkg/metrics"
	"github.com/lolo-pop/faas-scaling/pkg/nats"
	"github.com/lolo-pop/faas-scaling/pkg/scaling"
	"github.com/lolo-pop/faas-scaling/pkg/types"
)




var (
	natsUrl        string
	metricsSubject string
	reqSubject     string
	scalingWindows int64
)

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
}


func getFunctionAccRequire(sortedFunctionAccuracyMap []scaling.Kv, functionName string) float32 {
	for _, functionPair := range sortedFunctionAccuracyMap {
		if functionPair.Key == functionName {
			return functionPair.Value
		}
	}
	log.Fatalf("Cannot get %s function's accuracy requirement", functionName)
	return 0
}

func main() {

	var p metrics.Provider
	p = &metrics.FaasProvider{}
	scaling.Hello("test")
	rand.Seed(time.Now().UnixNano())
	// var accuracy [20]float32
	// for i := 0; i < 10; i++ {
	// 	accuracy[i] = rand.Float32()
	// }
	accuracy := [20]float32{0.667, 0.901, 0.676, 0.663, 0.822,
		0.776, 0.720, 0.851, 0.852, 0.662,
		0.759, 0.839, 0.612, 0.801, 0.790,
		0.668, 0.654, 0.760, 0.690, 0.853}
	functionAccuracy := make(map[string]float32)
	index := 0
	
	
	// 连接NATS并订阅metrics subject
	nc, err := nats.Connect(natsUrl)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot connect to nats: %s", err)
		log.Fatal(errMsg)
	}
	defer nc.Close()
	sub, err := nc.SubscribeSync(metricsSubject)
	if err != nil {
		errMsg := fmt.Sprintf("Cannot subscribe %s subject: %s", metricsSubject, err)
		log.Fatal(errMsg)
	}
	defer sub.Unsubscribe()
	// accuracy 和function name的对应关系需要确定是否是固定的。
	for {

		// var functions []types.Function
		// var nodes []types.Node
		msg, err := sub.NextMsg(0)


		functionNames, err := p.Functions()
		if err != nil {
			log.Fatal(err.Error())
		}
		// fmt.Println(functionNames)
		for _, fname := range functionNames {
			if strings.Contains(fname, "service") {
				continue
			} else if _, ok := functionAccuracy[fname]; !ok {
				functionAccuracy[fname] = accuracy[index]
				index += 1
			}
		}
		fmt.Println("current function-accuracy configration:", functionAccuracy)
	
		// 对所有非service container 按照准确度要求进行排序
		var sortedFunctionAccuracyMap []scaling.Kv
		sortedFunctionAccuracyMap = scaling.FunctionAccuracyMapSort(functionAccuracy)
		for _, funcAccPair := range sortedFunctionAccuracyMap {
			fmt.Println(funcAccPair)
		}
		ser
		var metrics types.Message
		err = json.Unmarshal(msg.Data, &metrics)
		if err != nil {
			errMsg := fmt.Sprintf("Cannot unmarshal message: %s", err)
			log.Fatal(errMsg)
		}
		functionSCMap
		fmt.Printf("Timestamp: %d", metrics.Timestamp)
 
		batchSize := 4 // 后续可能需要设置成非固定的batch size
		var preInvocationNum [4][]int // 不同准确度等级的service container 历史request 数据量
		var curSCReplicasNum [4]int  // 当前系统中不同准确度等级的service container的副本数
		var lastInvocationNum [4]int // 上一次时间窗口，不同准确度等级的 service container的历史request数据量
		for _, function := range metrics.Functions {
			functionName := function.Name
			// 区分 service container和App container
			if strings.Contains(functionName, "service") {
				fields := strings.Split(functionName, "-")
				accuracyLevel, err := strconv.Atoi(fields[1])
				if err != nil {
					log.Fatalf("Failed to convert accuracy level to int: %s", fields[1])
				curSCReplicasNum[accuracyLevel] = function.Replicas
			} else {
				functionAccRequire := getFunctionAccRequire(sortedFunctionAccuracyMap, functionName)
				accuracyLevel := int((functionAccRequire - 0.65) / 0.1) // 计算准确度level 
				functionInvocationRate := function.InvocationRate
				functionInvocationNum := int(functionInvocationRate / 0.04) //这里后续需要确定是0.04 还是 乘以timewindows

				// preInvocationNum[accuracyLevel] = append(preInvocationNum[accuracyLevel], functionInvocationNum)
				lastInvocationNum[accuracyLevel] += functionInvocationNum
				// 根据非service container的历史吞吐量 预测当前time windows的吞吐量
				// counter +=
			}
		}
		for level, invocationNum := range lastInvocationNum {
			preInvocationNum[level] = append(preInvocationNum[level], invocationNum)
		}

		for level, invocationNumSlice := range preInvocationNum {
			// level 表示准确度的等级
			// PredictInvocationNum 根据历史调用次数，预测下一个窗口调用次数
			// PredictSCReplicas 根据预测值计算预测service container的副本数
			predictNum, ok := scaling.PredictInvocationNum(invocationNumSlice)
			if !ok {
				log.Fatalf("Predict the number of service-%d requests failed", level)
			}
			predictSCReplicasNum, ok := scaling.PredictSCReplicas(predictNum, level, batchSize)
			if !ok {
				log.Fatalf("Predict the number of service-%d replicas failed", level)
			}
			if predictSCReplicasNum > curSCReplicasNum[level] { // 如果预测的副本数量大于当前的SC副本数量，则warm, 否则 label remove 
				ok := scaling.WarmSCReplicas(predictSCReplicasNum-curSCReplicasNum[level], level)
				if !ok {
					log.Fatalf("Warming up service-%d replicas failed: %d", level, predictSCReplicasNum-curSCReplicasNum[level])
				}
			} else {
				ok := scaling.RemoveSCReplicas(curSCReplicasNum[level]-predictSCReplicasNum, level)
				if !ok {
					log.Fatalf("Warming up service-%d replicas failed: %d", level, predictSCReplicasNum-curSCReplicasNum[level])
				}
			}
		}

		batchTimeout := scaling.CalculateTimeout()

		scaling.Handle(batchTimeout)

		for _, funcAccPair := range sortedFunctionAccuracyMap {

		}
		/*
			for _, funcAccPair := range sortedFunctionAccuracyMap {
				fname := funcAccPair.Key
				// funcAcc := funcAccPair.Value
				fmt.Println("main: ", fname)
				f := types.Function{Name: fname}

				f.Replicas, err = p.FunctionReplicas(f.Name)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s replicas: %d\n", f.Name, f.Replicas)

				f.InvocationRate, err = p.FunctionInvocationRate(f.Name, scalingWindows)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s invocation rate: %v\n", f.Name, f.InvocationRate)

				f.ResponseTime, err = p.ResponseTime(f.Name, scalingWindows)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s response time: %v", f.Name, f.ResponseTime)

				f.ProcessingTime, err = p.ProcessingTime(f.Name, scalingWindows)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s processing time: %v", f.Name, f.ProcessingTime)

				f.Throughput, err = p.Throughput(f.Name, scalingWindows)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s Throughput: %v", f.Name, f.Throughput)

				f.ColdStart, err = p.ColdStart(f.Name, scalingWindows)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s cold start time: %v", f.Name, f.ColdStart)

				f.Nodes, err = p.FunctionNodes(f.Name)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s nodes: %v", f.Name, f.Nodes)

				f.Cpu, f.Mem, err = p.TopPods(f.Name)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}
				log.Printf("%s CPU usage: %s", f.Name, sPrintMap(f.Cpu))
				log.Printf("%s memory usage: %s", f.Name, sPrintMap(f.Mem))

				functions = append(functions, f)

			}

			nodes, err = p.TopNodes()
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}

			for i, n := range nodes {
				nodeName := n.Name

				n.Functions, err = p.FunctionsInNode(nodeName)
				if err != nil {
					log.Printf("WARNING: %s", err.Error())
				}

				log.Printf("Node %s functions: %v", nodeName, n.Functions)
				log.Printf("Node %s CPU usage: %v", nodeName, n.Cpu)
				log.Printf("Node %s memory usage: %v", nodeName, n.Mem)

				// update the node
				nodes[i] = n
			}

			msg := types.Message{Functions: functions, Nodes: nodes, Timestamp: time.Now().Unix()}

			jsonMsg, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				log.Fatal(err.Error())
			}

			nats.Publish(jsonMsg)
		*/
		time.Sleep(time.Duration(scalingWindows) * time.Second)
	}
}

func sPrintMap(m map[string]float64) string {
	s := ""
	for key, val := range m {
		s += fmt.Sprintf("\n%s: %v", key, val)
	}
	return s
}
