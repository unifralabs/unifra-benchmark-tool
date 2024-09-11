package types

type RandomSeed int64

// Load test types

type VegetaAttack struct {
	Rate       int               `json:"rate"`
	Duration   int               `json:"duration"`
	Calls      []*JsonrpcMessage `json:"calls"`
	VegetaArgs *string           `json:"vegeta_args"` // Can be string or nil
}

type VegetaArgs interface{} // Can be string or nil
type MultiVegetaArgs []VegetaArgs
type VegetaArgsShorthand interface{} // Can be VegetaArgs or MultiVegetaArgs

type TestGenerationParameters struct {
	TestName   string              `json:"test_name"`
	RandomSeed RandomSeed          `json:"random_seed"`
	Rates      []int               `json:"rates"`
	Durations  []int               `json:"durations"`
	VegetaArgs VegetaArgsShorthand `json:"vegeta_args"`
	Network    string              `json:"network"`
}

type LoadTest struct {
	TestParameters TestGenerationParameters `json:"test_parameters"`
	Attacks        []VegetaAttack           `json:"attacks"`
}

type LoadTestColumnWise struct {
	Rates      []int           `json:"rates"`
	Durations  []int           `json:"durations"`
	Calls      [][]interface{} `json:"calls"`
	VegetaArgs []interface{}   `json:"vegeta_args"`
}

type LoadTestMode string

const (
	StressMode LoadTestMode = "stress"
	SpikeMode  LoadTestMode = "spike"
	SoakMode   LoadTestMode = "soak"
)

type LoadTestGenerator func(...interface{}) []VegetaAttack
type MultiLoadTestGenerator func(...interface{}) map[string]LoadTest

type RunType string

const (
	SingleTest RunType = "single_test"
)

type DeepOutput string

const (
	RawDeepOutput     DeepOutput = "raw"
	MetricsDeepOutput DeepOutput = "metrics"
)

type SingleRunTestPayload struct {
	Type           RunType                  `json:"type"`
	Name           string                   `json:"name"`
	TestParameters TestGenerationParameters `json:"test_parameters"`
}

type SingleRunResultsPayload struct {
	DependencyVersions map[string]*string        `json:"dependency_versions"`
	CLIArgs            []string                  `json:"cli_args"`
	Type               RunType                   `json:"type"`
	TRunStart          int64                     `json:"t_run_start"`
	TRunEnd            int64                     `json:"t_run_end"`
	Nodes              Nodes                     `json:"nodes"`
	Results            map[string]LoadTestOutput `json:"results"`
}
