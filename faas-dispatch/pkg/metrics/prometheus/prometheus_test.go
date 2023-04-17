// Prometheus test module
// Shell script test_environment in this directory can help to execute these tests as a minikube cluster,
// In this case remember to deploy Prometheus as a Service in order to access the exposed metrics.
package prometheus

import (
	"testing"
)

// test with an always running function
func TestFunctionReplicasMicroservice(t *testing.T) {
	name := "nodeinfo"
	want := 1
	got, err := FunctionReplicas(name)
	if err != nil {
		t.Errorf("Error occurred: %v\n", err)
	}
	if got < want {
		t.Errorf("FunctionReplicas(%s) = %d, that is less than %d", name, got, want)
	}
}

// test with a function scaled to zero
func TestFunctionReplicasZero(t *testing.T) {
	name := "figlet"
	want := 0
	got, err := FunctionReplicas(name)
	if got != want || err != nil {
		t.Errorf("FunctionReplicas(%s) = (%d, %v), want (%d, nil)", name, got, err, want)
	}
}

// test with a not existing function
func TestFunctionReplicasBad(t *testing.T) {
	name := "missingFunction"
	got, err := FunctionReplicas(name)
	if got != 0 || err == nil {
		t.Errorf("FunctionReplicas(%s) = (%d, %v), want (0, error)", name, got, err)
	}
}

// test with an empty name
func TestFunctionReplicasEmpty(t *testing.T) {
	name := ""
	got, err := FunctionReplicas(name)
	if got != 0 || err == nil {
		t.Errorf(`FunctionReplicas("") = (%d, %v), want (0, error)`, got, err)
	}
}

// test with a 2 seconds sleeping function
func TestResponseTimeSleep(t *testing.T) {
	name := "sleep"
	sinceSeconds := int64(600)
	minTime := 2.0
	got, err := ResponseTime(name, sinceSeconds)
	if err != nil {
		t.Errorf("Error occurred: %v\n", err)
	}
	if got < minTime {
		t.Errorf("ResponseTime(%s) = %v, that is less than %v", name, got, minTime)
	}
}

// test with a function scaled to zero
func TestResponseTimeZero(t *testing.T) {
	name := "figlet"
	sinceSeconds := int64(600)
	minTime := 0.0
	got, err := ResponseTime(name, sinceSeconds)
	if got < minTime || err != nil {
		t.Errorf("ResponseTime(%s) = (%v, %v), want (<time>, nil)", name, got, err)
	}
}

// test with a not existing function
func TestResponseTimeBad(t *testing.T) {
	name := "missingFunction"
	sinceSeconds := int64(600)
	got, err := ResponseTime(name, sinceSeconds)
	if got != 0 || err == nil {
		t.Errorf("ResponseTime(%s) = (%v, %v), want (0, error)", name, got, err)
	}
}

// test with an empty name
func TestResponseTimeEmpty(t *testing.T) {
	name := ""
	sinceSeconds := int64(600)
	got, err := ResponseTime(name, sinceSeconds)
	if got != 0 || err == nil {
		t.Errorf(`ResponseTime("") = (%v, %v), want (0, error)`, got, err)
	}
}

// test with a 2 seconds sleeping function
func TestProcessingTimeSleep(t *testing.T) {
	name := "sleep"
	sinceSeconds := int64(600)
	minTime := 2.0
	got, err := ProcessingTime(name, sinceSeconds)
	if err != nil {
		t.Errorf("Error occurred: %v\n", err)
	}
	if got < minTime {
		t.Errorf("ProcessingTime(%s) = %v, that is less than %v", name, got, minTime)
	}
}

// test with a function that succeeded at least once
func TestThroughput(t *testing.T) {
	name := "nodeinfo"
	sinceSeconds := int64(600)
	got, err := Throughput(name, sinceSeconds)
	minThr := 1.0 / float64(sinceSeconds)
	if err != nil {
		t.Errorf("Error occurred: %v\n", err)
	}
	if got < minThr {
		// assuming that the invocation succeeded
		t.Errorf("Throughput(%s) = %v, that is less than the minimum value %v", name, got, minThr)
	}
}

// test with a function that has been invoked at least once
func TestFunctionInvocationRate(t *testing.T) {
	name := "nodeinfo"
	sinceSeconds := int64(600)
	got, err := FunctionInvocationRate(name, sinceSeconds)
	minRate := 1.0 / float64(sinceSeconds)
	if err != nil {
		t.Errorf("Error occurred: %v\n", err)
	}
	if got < minRate {
		// assuming that the invocation succeeded
		t.Errorf("FunctionInvocationRate(%s) = %v, that is less than the minimum value %v", name, got, minRate)
	}
}
