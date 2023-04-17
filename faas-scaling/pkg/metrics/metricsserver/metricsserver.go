package metricsserver

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/lolo-pop/faas-scaling/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var mc *metrics.Clientset
var clientset *kubernetes.Clientset

const namespace = "openfaas-fn"

func init() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	mc, err = metrics.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func TopPods(functionName string) (map[string]float64, map[string]float64, error) {

	cpu := make(map[string]float64)
	mem := make(map[string]float64)

	listOptions := metav1.ListOptions{
		LabelSelector: "faas_function=" + functionName,
	}
	podMetrics, err := mc.MetricsV1beta1().PodMetricses(namespace).List(context.TODO(), listOptions)
	if err != nil {
		return nil, nil, err
	}

	for _, podMetric := range podMetrics.Items {
		podName := podMetric.Name

		pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}

		podContainers := podMetric.Containers
		if len(podContainers) == 0 {
			// containers still not available
			continue
		}
		cpu[podName] = 0 // initialize cpu counter
		mem[podName] = 0 // initialize mem counter
		containersCount := 0

		for i, container := range podContainers {

			containerCpuLimits := float64(pod.Spec.Containers[i].Resources.Limits.Cpu().MilliValue())
			containerMemLimits := float64(pod.Spec.Containers[i].Resources.Limits.Memory().Value())
			if containerCpuLimits == 0 {
				return nil, nil, errors.New("cpu limits not specified")
			}
			if containerCpuLimits == 0 {
				return nil, nil, errors.New("memory limits not specified")
			}

			containerCpuUsage := float64(container.Usage.Cpu().MilliValue())
			containerMemUsage := float64(container.Usage.Memory().Value())
			cpu[podName] += containerCpuUsage / containerCpuLimits // add container cpu usage percentage
			mem[podName] += containerMemUsage / containerMemLimits // add container memory usage percentage
			containersCount = i + 1
		}
		// get average percentage of usage between pod's containers
		cpu[podName] /= float64(containersCount)
		mem[podName] /= float64(containersCount)
	}

	if len(cpu) == 0 {
		msg := fmt.Sprintf("Function %s not found for resources utilization", functionName)
		return nil, nil, errors.New(msg)
	}

	return cpu, mem, nil
}

func TopNodes() ([]types.Node, error) {

	var nodes []types.Node

	nodeMetrics, err := mc.MetricsV1beta1().NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, nodeMetric := range nodeMetrics.Items {
		nodeName := nodeMetric.Name

		node, err := clientset.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		nodeCpuCapacity := node.Status.Capacity.Cpu().MilliValue()
		nodeMemCapacity := node.Status.Capacity.Memory().Value()

		cpu := float64(nodeMetric.Usage.Cpu().MilliValue()) / float64(nodeCpuCapacity)
		mem := float64(nodeMetric.Usage.Memory().Value()) / float64(nodeMemCapacity)
		nodes = append(nodes, types.Node{Name: nodeName, Cpu: cpu, Mem: mem})
	}

	return nodes, nil
}
