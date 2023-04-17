package types

type Function struct {
	Name           string             `json:"name"`
	Replicas       int                `json:"replicas"`
	InvocationRate float64            `json:"invocation_rate"`
	ResponseTime   float64            `json:"response_time"`
	ProcessingTime float64            `json:"processing_time"`
	Throughput     float64            `json:"throughput"`
	ColdStart      float64            `json:"cold_start"`
	Nodes          []string           `json:"nodes"`
	Cpu            map[string]float64 `json:"cpu,omitempty"`
	Mem            map[string]float64 `json:"mem,omitempty"`
}

type Node struct {
	Name      string   `json:"name"`
	Cpu       float64  `json:"cpu"`
	Mem       float64  `json:"mem"`
	Functions []string `json:"functions,omitempty"`
}

type Message struct {
	Functions []Function `json:"functions"`
	Nodes     []Node     `json:"nodes"`
	Timestamp int64      `json:"timestamp"`
}
