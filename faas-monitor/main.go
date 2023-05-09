package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/lolo-pop/faas-monitor/pkg/metrics"
	"github.com/lolo-pop/faas-monitor/pkg/nats"
	"github.com/lolo-pop/faas-monitor/pkg/types"
)

var scrapePeriod int64

func init() {
	env, ok := os.LookupEnv("SCRAPE_PERIOD")
	if !ok {
		log.Fatal("$SCRAPE_PERIOD not set")
	}
	var err error
	val, err := strconv.Atoi(env)
	if err != nil {
		log.Fatal(err.Error())
	}
	scrapePeriod = int64(val)
}

func main() {

	var p metrics.Provider
	p = &metrics.FaasProvider{}

	for {

		var functions []types.Function
		var nodes []types.Node

		functionNames, err := p.Functions()
		if err != nil {
			log.Fatal(err.Error())
		}

		for _, fname := range functionNames {

			f := types.Function{Name: fname}

			f.Replicas, err = p.FunctionReplicas(f.Name)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s replicas: %d\n", f.Name, f.Replicas)

			f.InvocationRate, err = p.FunctionInvocationCounter(f.Name, scrapePeriod)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s invocation rate: %v\n", f.Name, f.InvocationRate)

			f.ResponseTime, err = p.ResponseTime(f.Name, scrapePeriod)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s response time: %v", f.Name, f.ResponseTime)

			f.ProcessingTime, err = p.ProcessingTime(f.Name, scrapePeriod)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s processing time: %v", f.Name, f.ProcessingTime)

			f.Throughput, err = p.Throughput(f.Name, scrapePeriod)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s Throughput: %v", f.Name, f.Throughput)

			f.ColdStart, err = p.ColdStart(f.Name, scrapePeriod)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s cold start time: %v", f.Name, f.ColdStart) // 后续需要确认 需不需要减去processing time

			f.Nodes, err = p.FunctionNodes(f.Name)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s nodes: %v", f.Name, f.Nodes)

			f.Batch, err = p.BatchSize(f.Name)
			if err != nil {
				log.Printf("WARNING: %s", err.Error())
			}
			log.Printf("%s batch size: %v", f.Name, f.Batch)

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

		time.Sleep(time.Duration(scrapePeriod) * time.Second)
	}
}

func sPrintMap(m map[string][]float64) string {
	s := ""
	for key, val := range m {
		s += fmt.Sprintf("%s: %v", key, val)
	}

	return s
}
