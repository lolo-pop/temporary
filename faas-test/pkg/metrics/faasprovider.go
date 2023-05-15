package metrics

import (
	"github.com/lolo-pop/faas-monitor/pkg/metrics/apiserver"
	"github.com/lolo-pop/faas-monitor/pkg/metrics/gateway"
	"github.com/lolo-pop/faas-monitor/pkg/metrics/metricsserver"
	"github.com/lolo-pop/faas-monitor/pkg/metrics/prometheus"
	"github.com/lolo-pop/faas-monitor/pkg/types"
)

type FaasProvider struct{}

func (*FaasProvider) Functions() ([]string, error) {
	return gateway.Functions()
}

func (*FaasProvider) FunctionReplicas(functionName string) (int, error) {
	return prometheus.FunctionReplicas(functionName)
}

func (*FaasProvider) FunctionInvocationCounter(functionName string, sinceSeconds int64) (float64, error) {
	return prometheus.FunctionInvocationCounter(functionName, sinceSeconds)
}

func (*FaasProvider) ResponseTime(functionName string, sinceSeconds int64) (float64, error) {
	return prometheus.ResponseTime(functionName, sinceSeconds)
}

func (*FaasProvider) ProcessingTime(functionName string, sinceSeconds int64) (float64, error) {
	return prometheus.ProcessingTime(functionName, sinceSeconds)
}

func (*FaasProvider) Throughput(functionName string, sinceSeconds int64) (float64, error) {
	return prometheus.Throughput(functionName, sinceSeconds)
}

func (*FaasProvider) ColdStart(functionName string, sinceSeconds int64) (float64, error) {
	return apiserver.ColdStart(functionName, sinceSeconds)
}

func (*FaasProvider) TopPods(functionName string) (map[string][]float64, map[string][]float64, error) {
	return metricsserver.TopPods(functionName)
}

func (*FaasProvider) BatchSize(functionName string) (map[string]int, error) {
	return metricsserver.BatchSize(functionName)
}

func (*FaasProvider) TopNodes() ([]types.Node, error) {
	return metricsserver.TopNodes()
}

func (*FaasProvider) FunctionsInNode(nodeName string) ([]string, error) {
	return apiserver.FunctionsInNode(nodeName)
}

func (*FaasProvider) FunctionNodes(functionName string) ([]string, error) {
	return apiserver.FunctionNodes(functionName)
}
