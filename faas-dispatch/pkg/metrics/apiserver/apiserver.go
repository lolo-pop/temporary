package apiserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/lolo-pop/faas-scaling/pkg/util"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset

func init() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	// creates the clientset
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func ColdStart(functionName string, sinceSeconds int64) (float64, error) {

	scaleLine := fmt.Sprintf(`\[Scale\] function=%s 0 => 1 successful`, functionName)
	scaleRe := regexp.MustCompile(scaleLine)
	var scaleSum, scaleCount float64

	podLogOpts := v1.PodLogOptions{
		Container:    "gateway",
		SinceSeconds: &sinceSeconds,
	}

	listOptions := metav1.ListOptions{
		LabelSelector: "app=gateway",
		FieldSelector: "status.phase=Running",
	}
	pods, err := clientset.CoreV1().Pods("openfaas").List(context.TODO(), listOptions)
	if err != nil {
		return 0, err
	}

	for _, pod := range pods.Items {

		podName := pod.Name

		req := clientset.CoreV1().Pods("openfaas").GetLogs(podName, &podLogOpts)
		podLogs, err := req.Stream(context.TODO())
		if err != nil {
			return 0, err
		}

		scanner := bufio.NewScanner(podLogs)
		for scanner.Scan() {
			line := scanner.Text()
			if scaleRe.MatchString(line) { // match scale line
				val, err := util.ExtractValueBetween(line, `- `, `s`)
				if err != nil {
					return 0, err
				}
				scaleSum += val
				scaleCount++
			}
		}
		err = scanner.Err()
		if err != nil {
			return 0, err
		}

		err = podLogs.Close()
		if err != nil {
			return 0, err
		}
	}

	if scaleCount == 0 {
		return 0, errors.New("no cold starts occurred")
	}

	return scaleSum / scaleCount, nil
}

func FunctionsInNode(nodeName string) ([]string, error) {

	options := metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + nodeName +
			",metadata.namespace=openfaas-fn" +
			",status.phase=Running",
	}
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), options)
	if err != nil {
		return nil, err
	}

	var functions []string
	for _, pod := range pods.Items {
		functions = append(functions, pod.GetName())
	}

	return functions, nil
}

func FunctionNodes(functionName string) ([]string, error) {

	options := metav1.ListOptions{
		LabelSelector: "faas_function=" + functionName,
		FieldSelector: "status.phase=Running",
	}
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), options)
	if err != nil {
		return nil, err
	}

	var nodes []string
	for _, pod := range pods.Items {
		nodes = append(nodes, pod.Spec.NodeName)
	}

	return nodes, nil
}
