package scaling

const (
	// DefaultMinReplicas is the minimal amount of replicas for a service.
	// 默认的副本的数量是1，为了支持scaling to 0 修改为1
	DefaultMinReplicas = 1

	// DefaultMaxReplicas is the amount of replicas a service will auto-scale up to.
	DefaultMaxReplicas = 1000

	// DefaultScalingFactor is the defining proportion for the scaling increments.
	DefaultScalingFactor = 10

	DefaultTypeScale = "rps"

	// MinScaleLabel label indicating min scale for a function
	MinScaleLabel = "com.openfaas.scale.min"

	// MaxScaleLabel label indicating max scale for a function
	MaxScaleLabel = "com.openfaas.scale.max"

	// ScalingFactorLabel label indicates the scaling factor for a function
	ScalingFactorLabel = "com.openfaas.scale.factor"
)
