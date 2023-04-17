package prometheus

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lolo-pop/faas-scaling/pkg/util"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

var v1api v1.API

type IdleError struct {
	Function string
	Period   int64
}

func (e *IdleError) Error() string {
	return fmt.Sprintf(
		"function %s has been idle in the last %v seconds",
		e.Function, e.Period)
}

func init() {
	prometheusUrl, ok := os.LookupEnv("PROMETHEUS_URL")
	if !ok {
		log.Fatal("$PROMETHEUS_URL not set")
	}
	client, err := api.NewClient(api.Config{
		Address: prometheusUrl,
	})
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	v1api = v1.NewAPI(client)
}

func FunctionReplicas(functionName string) (int, error) {

	if len(functionName) == 0 {
		msg := "empty function name"
		return 0, errors.New(msg)
	}

	q := fmt.Sprintf(`sum by (function_name) (gateway_service_count{function_name="%v.openfaas-fn"})`,
		functionName)

	stringResult, err := query(q)
	if err != nil {
		return 0, err
	}

	if len(stringResult) == 0 {
		msg := fmt.Sprintf("function %s not found in the openfaas-fn namespace", functionName)
		return 0, errors.New(msg)
	}

	replicas, err := util.ExtractValueBetween(stringResult, `=> `, ` @`)
	if err != nil {
		return 0, err
	}

	return int(replicas), err
}

func FunctionInvocationRate(functionName string, sinceSeconds int64) (float64, error) {

	q := fmt.Sprintf(`sum by (function_name)`+
		`(rate(gateway_function_invocation_total{function_name="%s.openfaas-fn"}[%ds]) > 0)`,
		functionName, sinceSeconds)

	ir, err := querySince(q, functionName, sinceSeconds)
	if err != nil {
		return 0, err
	}

	return ir, nil
}

func ResponseTime(functionName string, sinceSeconds int64) (float64, error) {

	q := fmt.Sprintf(
		`sum by (function_name)`+
			`(rate(gateway_functions_seconds_sum{function_name="%s.openfaas-fn"}[%ds]) > 0) `+
			`/ `+
			`sum by (function_name)`+
			`(rate(gateway_functions_seconds_count{function_name="%s.openfaas-fn"}[%ds]) > 0)`,
		functionName, sinceSeconds, functionName, sinceSeconds,
	)

	rt, err := querySince(q, functionName, sinceSeconds)
	if err != nil {
		return 0, err
	}

	return rt, nil
}

func ProcessingTime(functionName string, sinceSeconds int64) (float64, error) {

	q := fmt.Sprintf(
		`sum by (faas_function)`+
			`(rate(http_request_duration_seconds_sum{faas_function="%s",code="200"}[%ds])>0)`+
			`/`+
			`sum by (faas_function)`+
			`(rate(http_request_duration_seconds_count{faas_function="%s",code="200"}[%ds])>0)`,
		functionName, sinceSeconds, functionName, sinceSeconds,
	)

	pt, err := querySince(q, functionName, sinceSeconds)
	if err != nil {
		return 0, err
	}

	return pt, nil
}

func Throughput(functionName string, sinceSeconds int64) (float64, error) {

	q := fmt.Sprintf(
		`sum by (function_name)`+
			`(rate(gateway_function_invocation_total{code="200",function_name="%s.openfaas-fn"}[%ds]) > 0)`,
		functionName, sinceSeconds,
	)

	thr, err := querySince(q, functionName, sinceSeconds)
	if err != nil {
		return 0, err
	}

	return thr, nil
}

func querySince(q, functionName string, sinceSeconds int64) (float64, error) {

	if len(functionName) == 0 {
		msg := "empty function name"
		return 0, errors.New(msg)
	}

	stringResult, err := query(q)
	if len(stringResult) == 0 {
		return 0, &IdleError{Function: functionName, Period: sinceSeconds}
	}

	value, err := util.ExtractValueBetween(stringResult, `=> `, ` @`)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func query(q string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.Query(ctx, q, time.Now())
	if err != nil {
		return "", err
	}
	if len(warnings) > 0 {
		log.Printf("Warnings: %v", warnings)
	}

	stringResult := result.String()
	//fmt.Println(stringResult)

	return stringResult, nil
}
