package main

import (
	"bytes"
	"fmt"
	"net/http"
)

func main() {
	// OpenFaaS gateway URL
	gatewayURL := "http://127.0.0.1:8080"

	// Name of the function to scale
	functionName := "test-0"

	// New replica count
	replicaCount := 5

	// Create HTTP client
	client := &http.Client{}

	// Build request URL
	url := fmt.Sprintf("%s/system/scale-function/%s", gatewayURL, functionName)

	// Build request body
	requestBody := fmt.Sprintf(`{"serviceName":"%s","replicas":%d}`, functionName, replicaCount)
	requestBodyBytes := []byte(requestBody)

	// Create HTTP request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		panic(err)
	}

	// Set request headers
	request.Header.Set("Content-Type", "application/json")
	user := "admin"
	password := "admin"
	request.SetBasicAuth(user, password)
	// Send HTTP request
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}

	// Check response status code
	if response.StatusCode != http.StatusOK {
		panic(fmt.Sprintf("Unexpected response status code: %d", response.StatusCode))
	}

	// Print success message
	fmt.Printf("Scaled function %s to %d replicas\n", functionName, replicaCount)
}
