package types

// Load tests outputs

type RawLoadTestOutputDatum struct {
	Latencies             map[string]float64 `json:"latencies"`
	BytesIn               map[string]float64 `json:"bytes_in"`
	BytesOut              map[string]float64 `json:"bytes_out"`
	Earliest              string             `json:"earliest"`
	Latest                string             `json:"latest"`
	End                   string             `json:"end"`
	Duration              int                `json:"duration"`
	Wait                  int                `json:"wait"`
	Requests              int                `json:"requests"`
	Rate                  float64            `json:"rate"`
	Throughput            float64            `json:"throughput"`
	Success               float64            `json:"success"`
	StatusCodes           map[string]int     `json:"status_codes"`
	Errors                []string           `json:"errors"`
	FirstRequestTimestamp string             `json:"first_request_timestamp"`
	LastRequestTimestamp  string             `json:"last_request_timestamp"`
	LastResponseTimestamp string             `json:"last_response_timestamp"`
	FinalWaitTime         int                `json:"final_wait_time"`
}

type LoadTestOutputDatum struct {
	TargetRate            int                                          `json:"target_rate"`
	ActualRate            *float64                                     `json:"actual_rate"`
	TargetDuration        int                                          `json:"target_duration"`
	ActualDuration        *float64                                     `json:"actual_duration"`
	Requests              int                                          `json:"requests"`
	Throughput            *float64                                     `json:"throughput"`
	Success               *float64                                     `json:"success"`
	Min                   *float64                                     `json:"min"`
	Mean                  *float64                                     `json:"mean"`
	P50                   *float64                                     `json:"p50"`
	P90                   *float64                                     `json:"p90"`
	P95                   *float64                                     `json:"p95"`
	P99                   *float64                                     `json:"p99"`
	Max                   *float64                                     `json:"max"`
	StatusCodes           map[string]int                               `json:"status_codes"`
	Errors                []string                                     `json:"errors"`
	FirstRequestTimestamp *string                                      `json:"first_request_timestamp"`
	LastRequestTimestamp  *string                                      `json:"last_request_timestamp"`
	LastResponseTimestamp *string                                      `json:"last_response_timestamp"`
	FinalWaitTime         *float64                                     `json:"final_wait_time"`
	DeepRawOutput         *string                                      `json:"deep_raw_output"`
	DeepMetrics           map[ResponseCategory]LoadTestDeepOutputDatum `json:"deep_metrics"`
	DeepRPCErrorPairs     []ErrorPair                                  `json:"deep_rpc_error_pairs"`
}

type ResponseCategory string

const (
	AllResponses        ResponseCategory = "all"
	SuccessfulResponses ResponseCategory = "successful"
	FailedResponses     ResponseCategory = "failed"
)

type ErrorPair [2]interface{}

type LoadTestDeepOutputDatum struct {
	TargetRate            int            `json:"target_rate"`
	ActualRate            *float64       `json:"actual_rate"`
	TargetDuration        int            `json:"target_duration"`
	ActualDuration        *float64       `json:"actual_duration"`
	Requests              int            `json:"requests"`
	Throughput            *float64       `json:"throughput"`
	Success               *float64       `json:"success"`
	Min                   *float64       `json:"min"`
	Mean                  *float64       `json:"mean"`
	P50                   *float64       `json:"p50"`
	P90                   *float64       `json:"p90"`
	P95                   *float64       `json:"p95"`
	P99                   *float64       `json:"p99"`
	Max                   *float64       `json:"max"`
	StatusCodes           map[string]int `json:"status_codes"`
	Errors                []string       `json:"errors"`
	FirstRequestTimestamp *string        `json:"first_request_timestamp"`
	LastRequestTimestamp  *string        `json:"last_request_timestamp"`
	LastResponseTimestamp *string        `json:"last_response_timestamp"`
	FinalWaitTime         *float64       `json:"final_wait_time"`
	NInvalidJSONErrors    int            `json:"n_invalid_json_errors"`
	NRPCErrors            int            `json:"n_rpc_errors"`
}

type LoadTestOutput struct {
	TargetRate            []int                                   `json:"target_rate"`
	ActualRate            []*float64                              `json:"actual_rate"`
	TargetDuration        []int                                   `json:"target_duration"`
	ActualDuration        []*float64                              `json:"actual_duration"`
	Requests              []int                                   `json:"requests"`
	Throughput            []*float64                              `json:"throughput"`
	Success               []*float64                              `json:"success"`
	Min                   []*float64                              `json:"min"`
	Mean                  []*float64                              `json:"mean"`
	P50                   []*float64                              `json:"p50"`
	P90                   []*float64                              `json:"p90"`
	P95                   []*float64                              `json:"p95"`
	P99                   []*float64                              `json:"p99"`
	Max                   []*float64                              `json:"max"`
	StatusCodes           []map[string]int                        `json:"status_codes"`
	Errors                [][]string                              `json:"errors"`
	FirstRequestTimestamp []*string                               `json:"first_request_timestamp"`
	LastRequestTimestamp  []*string                               `json:"last_request_timestamp"`
	LastResponseTimestamp []*string                               `json:"last_response_timestamp"`
	FinalWaitTime         []*float64                              `json:"final_wait_time"`
	DeepRawOutput         []*string                               `json:"deep_raw_output"`
	DeepMetrics           map[ResponseCategory]LoadTestDeepOutput `json:"deep_metrics"`
	DeepRPCErrorPairs     [][]ErrorPair                           `json:"deep_rpc_error_pairs"`
}

type LoadTestDeepOutput struct {
	TargetRate            []int            `json:"target_rate"`
	ActualRate            []*float64       `json:"actual_rate"`
	TargetDuration        []int            `json:"target_duration"`
	ActualDuration        []*float64       `json:"actual_duration"`
	Requests              []int            `json:"requests"`
	Throughput            []*float64       `json:"throughput"`
	Success               []*float64       `json:"success"`
	Min                   []*float64       `json:"min"`
	Mean                  []*float64       `json:"mean"`
	P50                   []*float64       `json:"p50"`
	P90                   []*float64       `json:"p90"`
	P95                   []*float64       `json:"p95"`
	P99                   []*float64       `json:"p99"`
	Max                   []*float64       `json:"max"`
	StatusCodes           []map[string]int `json:"status_codes"`
	Errors                [][]string       `json:"errors"`
	FirstRequestTimestamp []*string        `json:"first_request_timestamp"`
	LastRequestTimestamp  []*string        `json:"last_request_timestamp"`
	LastResponseTimestamp []*string        `json:"last_response_timestamp"`
	FinalWaitTime         []*float64       `json:"final_wait_time"`
	NInvalidJSONErrors    []int            `json:"n_invalid_json_errors"`
	NRPCErrors            []int            `json:"n_rpc_errors"`
}

func BuildLoadTestOutput(listOfMaps []*LoadTestOutputDatum) LoadTestOutput {
	if len(listOfMaps) == 0 {
		return LoadTestOutput{}
	}

	result := LoadTestOutput{}

	result.TargetRate = make([]int, len(listOfMaps))
	for i, m := range listOfMaps {
		result.TargetRate[i] = m.TargetRate
	}

	result.ActualRate = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.ActualRate[i] = m.ActualRate
	}
	result.TargetDuration = make([]int, len(listOfMaps))
	for i, m := range listOfMaps {
		result.TargetDuration[i] = m.TargetDuration
	}
	result.ActualDuration = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.ActualDuration[i] = m.ActualDuration
	}
	result.Requests = make([]int, len(listOfMaps))
	for i, m := range listOfMaps {
		result.Requests[i] = m.Requests
	}
	result.Throughput = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.Throughput[i] = m.Throughput
	}
	result.Success = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.Success[i] = m.Success
	}
	result.Min = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.Min[i] = m.Min
	}
	result.Mean = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.Mean[i] = m.Mean
	}
	result.P50 = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.P50[i] = m.P50
	}
	result.P90 = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.P90[i] = m.P90
	}
	result.P95 = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.P95[i] = m.P95
	}
	result.P99 = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.P99[i] = m.P99
	}
	result.Max = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.Max[i] = m.Max
	}
	result.StatusCodes = make([]map[string]int, len(listOfMaps))
	for i, m := range listOfMaps {
		result.StatusCodes[i] = m.StatusCodes
	}
	result.Errors = make([][]string, len(listOfMaps))
	for i, m := range listOfMaps {
		result.Errors[i] = m.Errors
	}
	result.FirstRequestTimestamp = make([]*string, len(listOfMaps))
	for i, m := range listOfMaps {
		result.FirstRequestTimestamp[i] = m.FirstRequestTimestamp
	}
	result.LastRequestTimestamp = make([]*string, len(listOfMaps))
	for i, m := range listOfMaps {
		result.LastRequestTimestamp[i] = m.LastRequestTimestamp
	}
	result.LastResponseTimestamp = make([]*string, len(listOfMaps))
	for i, m := range listOfMaps {
		result.LastResponseTimestamp[i] = m.LastResponseTimestamp
	}
	result.FinalWaitTime = make([]*float64, len(listOfMaps))
	for i, m := range listOfMaps {
		result.FinalWaitTime[i] = m.FinalWaitTime
	}

	result.DeepRawOutput = make([]*string, len(listOfMaps))
	for i, m := range listOfMaps {
		result.DeepRawOutput[i] = m.DeepRawOutput
	}

	return result
}
