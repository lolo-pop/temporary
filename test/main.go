package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Function struct {
	Name   string            `json:"functionName"`
	Labels map[string]string `json:"labels"`
}

func main() {
	// 构造函数对象

	labels := map[string]string{
		"com.openfaas.scale.min": "1",
		"com.openfaas.scale.max": "1",
		"instance.idle":          "true",
	}
	f := Function{
		Name:   "nats-test",
		Labels: labels,
	}
	fun := map[string]interface{}{
		"service": "nats-test",
		"image":   "lolopop/nats-test:latest",
		"labels":  labels,
	}
	fmt.Println(f)
	fmt.Println(fun)
	// 将对象编码为JSON
	data, err := json.Marshal(fun)
	if err != nil {
		panic(err)
	}

	// 调用OpenFaaS API
	req, err := http.NewRequest("PUT", "http://gateway.openfaas.svc.cluster.local:8080/system/functions", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")
	user := "admin"
	password := "admin"
	req.SetBasicAuth(user, password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)
	time.Sleep(time.Second * 10000)
}

type FunctionDeployment struct {

	// Service is the name of the function deployment
	Service string `json:"service"`

	// Image is a fully-qualified container image
	Image string `json:"image"`

	// Namespace for the function, if supported by the faas-provider
	Namespace string `json:"namespace,omitempty"`

	// EnvProcess overrides the fprocess environment variable and can be used
	// with the watchdog
	EnvProcess string `json:"envProcess,omitempty"`

	// EnvVars can be provided to set environment variables for the function runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Constraints are specific to the faas-provider.
	Constraints []string `json:"constraints,omitempty"`

	// Secrets list of secrets to be made available to function
	Secrets []string `json:"secrets,omitempty"`

	// Labels are metadata for functions which may be used by the
	// faas-provider or the gateway
	Labels *map[string]string `json:"labels,omitempty"`

	// Annotations are metadata for functions which may be used by the
	// faas-provider or the gateway
	Annotations *map[string]string `json:"annotations,omitempty"`

	// Limits for function
	Limits *FunctionResources `json:"limits,omitempty"`

	// Requests of resources requested by function
	Requests *FunctionResources `json:"requests,omitempty"`

	// ReadOnlyRootFilesystem removes write-access from the root filesystem
	// mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem,omitempty"`
}

// FunctionResources Memory and CPU
type FunctionResources struct {
	Memory string `json:"memory,omitempty"`
	CPU    string `json:"cpu,omitempty"`
}
