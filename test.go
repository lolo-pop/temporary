package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func main() {
	// Set the Prometheus API endpoint URL
	prometheusURL := "http://10.244.0.32:9090/api/v1/query"

	// Set the query to retrieve the RPS for a function named "my-function" over the last 5 minutes
	query := "rate(gateway_function_invocation_total{function_name=\"hello-python\"}[5m])"

	// Set the query parameters
	params := url.Values{}
	params.Set("query", query)

	// Send the HTTP GET request to the Prometheus API endpoint
	response, err := http.Get(prometheusURL + "?" + params.Encode())
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	// Parse the JSON response
	var responseJSON map[string]interface{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&responseJSON)
	if err != nil {
		panic(err)
	}

	// Extract the RPS value from the query result
	data, ok := responseJSON["data"].(map[string]interface{})
	if !ok {
		panic("Unexpected response format: 'data' field not found")
	}
	result, ok := data["result"].([]interface{})
	if !ok || len(result) == 0 {
		panic("No results found for the query")
	}
	value, ok := result[0].(map[string]interface{})["value"].([]interface{})
	if !ok || len(value) < 2 {
		panic("Unexpected response format: 'value' field not found")
	}
	rpsValue, ok := value[1].(string)
	if !ok {
		panic("Unexpected response format: 'value' is not a string")
	}

	// Print the RPS value
	fmt.Printf("RPS for my-function: %s\n", rpsValue)
}
