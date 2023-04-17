package metrics

import "github.com/smvfal/faas-monitor/pkg/types"

type Provider interface {

	// Functions provides the names of the deployed functions
	Functions() ([]string, error)

	// FunctionReplicas provides the function replicas' number
	FunctionReplicas(functionName string) (int, error)

	// FunctionInvocationRate provides the function invocation rate
	FunctionInvocationRate(functionName string, sinceSeconds int64) (float64, error)

	// ResponseTime provides function's average response time
	ResponseTime(functionName string, sinceSeconds int64) (float64, error)

	// ProcessingTime provides function's average processing time
	ProcessingTime(functionName string, sinceSeconds int64) (float64, error)

	// Throughput provides function's throughput
	Throughput(functionName string, sinceSeconds int64) (float64, error)

	// ColdStart provides function's cold start time
	ColdStart(functionName string, SinceSeconds int64) (float64, error)

	// TopPods provides function's current average CPU and memory percentage usage for each replica
	TopPods(functionName string) (map[string]float64, map[string]float64, error)

	// TopNodes provides nodes current CPU and memory percentage usage
	TopNodes() ([]types.Node, error)

	// FunctionsInNode returns functions' instances of a node
	FunctionsInNode(nodeName string) ([]string, error)

	// FunctionNodes returns the nodes in which the function is deployed
	FunctionNodes(functionName string) ([]string, error)
}
