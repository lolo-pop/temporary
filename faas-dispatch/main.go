package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/lolo-pop/faas-dispatch/pkg/metrics"
	"github.com/lolo-pop/faas-dispatch/pkg/scaling"
)

var scalingWindows int64

func init() {
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

func main() {

	var p metrics.Provider
	p = &metrics.FaasProvider{}
	scaling.Hello("test")
	rand.Seed(time.Now().UnixNano())
	var accuracy [10]float32
	for i := 0; i < 10; i++ {
		accuracy[i] = rand.Float32()
	}
	functionAccuracy := make(map[string]float32)
	index := 0
	// accuracy 和function name的对应关系需要确定是否是固定的。
	for {

		// var functions []types.Function
		// var nodes []types.Node

		functionNames, err := p.Functions()
		// fmt.Println(functionNames)
		for _, fname := range functionNames {
			if _, ok := functionAccuracy[fname]; !ok {
				functionAccuracy[fname] = accuracy[index]
				index += 1
			}
		}
		fmt.Println("current function-accuracy configration:", functionAccuracy)
		if err != nil {
			log.Fatal(err.Error())
		}

		var sortedFunctionAccuracyMap []scaling.Kv
		sortedFunctionAccuracyMap = scaling.FunctionAccuracyMapSort(functionAccuracy)
		for _, funcAccPair := range sortedFunctionAccuracyMap {
			fmt.Println(funcAccPair)
		}
		// debug here
		// for _, fname := range functionNames {

		batchTimeout := scaling.CalculateTimeout()

		scaling.Handle(batchTimeout)

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
