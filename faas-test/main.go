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
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func main() {
	// 设置 OpenFaaS API 网关地址和函数名称
	gateway := "http://gateway.openfaas.svc.cluster.local:8080"
	function := "nats-test"

	for i := 0; i < 5; i++ {
		batchSize := i + 1
		scaleFunction(gateway, function, i+1, batchSize)
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
