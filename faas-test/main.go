/*
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
)

type FunctionInstance struct {
	Name      string
	Replicas  int
	EnvLabel  string
	EnvValues map[string]string
}

func main() {
	// Define OpenFaaS API endpoint and function name
	openfaasEndpoint := "http://gateway.openfaas.svc.cluster.local:8080"

	// Get current concurrency level
	functionInstances := []FunctionInstance{
		{
			Name:     "nats-test",
			Replicas: 3,
			EnvLabel: "bs3",
		},
		{
			Name:     "nats-test",
			Replicas: 2,
			EnvLabel: "bs2",
		},
	}
	for _, instance := range functionInstances {
		scaleFunctionInstance(openfaasEndpoint, instance.Name, instance.Replicas, instance.EnvLabel)
	}
}

// scaleFunctionInstance scales the function instance with the given batch size
func scaleFunctionInstance(endpoint string, functionName string, replicaNum int, envLabel string) {
	// Set environment variables for new instance
	labels := map[string]string{
		"env": envLabel,
	}
	log.Println("envlabel:", envLabel)
	// Define scale request payload

	// Send scale request to OpenFaaS API
	jsonLabels, err := json.Marshal(labels)
	if err != nil {
		log.Fatalf("Failed to marshal JSON labels: %v", err)
	}
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("POST", endpoint+"/system/scale-function/"+functionName+"?label_selector=env%3D"+envLabel, bytes.NewBuffer(jsonLabels))
	if err != nil {
		fmt.Println("Failed to create HTTP request:", err)
		return
	}
	user := "admin"
	password := "admin"
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, password)

	q := req.URL.Query()
	q.Add("count", strconv.Itoa(replicaNum))
	req.URL.RawQuery = q.Encode()

	response, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send scale request:", err)
		return
	}
	defer response.Body.Close()
	fmt.Println(response)
	resBody, err := ioutil.ReadAll(response.Body)
	fmt.Printf("resBody: %v", string(resBody))
	if err != nil {
		fmt.Println("Failed to read response body:", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Failed to scale function %s: %d %s", functionName, response.StatusCode, string(resBody))
	}
	log.Printf("Scaled function %s to %d instances with labels %v\n", functionName, replicaNum, labels)
}
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/openfaas/faas-netes/pkg/client/clientset/versioned/scheme"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/http"
	"strings"
	"time"
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

func main() {
	// 设置 OpenFaaS API 网关地址和函数名称
	gateway := "http://gateway.openfaas.svc.cluster.local:8080"
	functionName := "nats-test"
	namespace := "openfaas-fn"
	for i := 0; i < 5; i++ {
		batchSize := i + 1
		// scaleFunction(gateway, functionName, i+1, batchSize)
		scaleFunction(gateway, functionName, i+1, batchSize)
		time.Sleep(time.Second * 20)
	}
}

func scaleFunction(httpURL string, functionName string, replicas int, batchSize int) {
	// 调用 OpenFaaS API 扩展 function
	// 设置环境变量 BATCH_SIZE
	fmt.Printf("Scaling function %s to %d replicas with BATCH_SIZE=%d\n", functionName, replicas, batchSize)

	// 更新 function 的配置
	updateFunction(httpURL, functionName, batchSize)

	// 扩展 function
	client := &http.Client{}
	url := fmt.Sprintf("%s/system/scale-function/%s", httpURL, functionName)
	payload := strings.NewReader(fmt.Sprintf(`{"serviceName": "%s", "replicas": %d}`, functionName, replicas))
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		errMsg := fmt.Sprintf("request Put failed when scaling function: %s", err.Error())
		fmt.Println(errMsg)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	user := "admin"
	password := "admin"
	req.SetBasicAuth(user, password)
	res, err := client.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("request POST failed when scaling function: %s", err.Error())
		fmt.Println(errMsg)
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		errMsg := fmt.Sprintf("read body failed when scaling function: %s", err.Error())
		fmt.Println(errMsg)
		return
	}
	fmt.Println(string(body))
}

func updateFunction(gatewayURL string, functionName string, batchSize int) {
	// 调用 OpenFaaS API 更新 nats-test 函数的环境变量
	yamlConfig, err := getYAMLConfig(gatewayURL, functionName)
	fmt.Println(yamlConfig)
	if err != nil {
		errMsg := fmt.Sprintf("get function config failed when updating function: %s", err.Error())
		fmt.Println(errMsg)
		return
	}
	updatedYamlConfig, err := updateYAMLConfig(yamlConfig, batchSize)

	if err != nil {
		errMsg := fmt.Sprintf("update yaml config failed when updating function: %s", err.Error())
		fmt.Println(errMsg)
		return
	}
	err = updateFunctionConfig(gatewayURL, updatedYamlConfig)
	if err != nil {
		errMsg := fmt.Sprintf("update function config failed when updating function: %s", err.Error())
		fmt.Println(errMsg)
		return
	}

	fmt.Println("Function configuration updated successfully")
}

func getYAMLConfig(gatewayURL string, functionName string) ([]byte, error) {
	httpClient := &http.Client{}
	URL := fmt.Sprintf("%s/system/function/%s", gatewayURL, functionName)
	fmt.Println(URL)
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	user := "admin"
	password := "admin"
	req.SetBasicAuth(user, password)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	/*
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to get YAML config for function %s: %s", functionName, resp.Status)
		}
	*/
	yamlConfig, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return yamlConfig, nil
}

// 修改YAML配置中的`environment`字段
func updateYAMLConfig(yamlConfig []byte, batchSize int) ([]byte, error) {
	var functionConfig map[string]interface{}

	err := yaml.Unmarshal(yamlConfig, &functionConfig)
	if err != nil {
		fmt.Println("error happened 1")
		return nil, err
	}
	fmt.Println(functionConfig)
	environment, ok := functionConfig["envProcess"].(map[string]interface{})
	if !ok {
		environment = make(map[string]interface{})
		functionConfig["environment"] = environment
	}
	fmt.Println(functionConfig["envProcess"])
	environment["MY_ENV_VAR"] = "my_value"

	updatedYamlConfig, err := yaml.Marshal(functionConfig)
	if err != nil {
		return nil, err
	}

	return updatedYamlConfig, nil
}

// 上传修改后的YAML配置
func updateFunctionConfig(gatewayURL string, updatedYamlConfig []byte) error {
	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/system/functions", gatewayURL), bytes.NewBuffer(updatedYamlConfig))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/yaml")
	user := "admin"
	password := "admin"
	req.SetBasicAuth(user, password)
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update function config: %s", resp.Status)
	}

	return nil
}

func updateFunctionK8s(namespace string, functionName string, batchSize int) {
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("faas_function=%s", functionName),
	})
	if err != nil {
		panic(err.Error())
	}
	if len(podList.Items) == 0 {
		panic(fmt.Sprintf("No running pods found for function %s", functionName))
	}
	podName := podList.Items[0].Name

	// 获取OpenFaaS函数的环境变量
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, v1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	// 解码Pod的注释以获取OpenFaaS函数的环境变量
	var functionSpec function
	err = runtime.DecodeInto(serializer.NewCodecFactory(scheme.Scheme).UniversalDeserializer(), []byte(pod.Annotations["com.openfaas.scale.zero.function"]), &functionSpec)
	if err != nil {
		panic(err.Error())
	}

	// 更新OpenFaaS函数的环境变量
	functionSpec.Environment["MY_VAR"] = "my-new-value"

	// 将更新后的函数规范编码为JSON格式
	functionJSON, err := json.Marshal(functionSpec)
	if err != nil {
		panic(err.Error())
	}

	// 更新Pod的注释以更新OpenFaaS函数的环境变量
	pod.Annotations["com.openfaas.scale.zero.function"] = string(functionJSON)
	_, err = clientset.CoreV1().Pods(namespace).Update(context.TODO(), pod, v1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}

}

type function struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	EnvProcess  string            `json:"envProcess"`
	Environment map[string]string `json:"environment"`
}
