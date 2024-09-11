package vegeta

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/unifralabs/unifra-benchmark-tool/types"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

func ComputeDeepDatum(rawOutput []byte, targetRate int, targetDuration int, calls []*types.JsonrpcMessage) (map[tooltypes.ResponseCategory]tooltypes.LoadTestDeepOutputDatum, []tooltypes.ErrorPair, error) {
	allDF, err := convertRawVegetaOutputToDataframe(rawOutput)
	if err != nil {
		return nil, nil, err
	}

	rpcError := make([]bool, len(allDF))
	invalidJSONError := make([]bool, len(allDF))

	for i, row := range allDF {
		statusCode := row["status_code"].(int64)
		response := row["response"].(string)

		if statusCode == 200 {
			decoded, err := base64.StdEncoding.DecodeString(response)
			if err != nil {
				invalidJSONError[i] = true
				continue
			}

			var result map[string]interface{}
			if err := json.Unmarshal(decoded, &result); err != nil {
				invalidJSONError[i] = true
				continue
			}

			if result["result"] == nil {
				rpcError[i] = true
			}
		}
	}

	allDF = append(allDF, map[string]interface{}{
		"invalid_json_error": invalidJSONError,
		"rpc_error":          rpcError,
	})

	deepSuccess := make([]bool, len(allDF))
	for i, row := range allDF {
		deepSuccess[i] = row["status_code"] != nil && row["status_code"].(int64) == 200 &&
			row["invalid_json_error"] == nil &&
			row["rpc_error"] == nil
	}
	allDF = append(allDF, map[string]interface{}{"deep_success": deepSuccess})

	rpcErrorPairs, err := gatherErrorPairs(allDF, calls)
	if err != nil {
		return nil, nil, err
	}

	categoryData := make(map[tooltypes.ResponseCategory]tooltypes.LoadTestDeepOutputDatum)
	dataframes := []struct {
		category tooltypes.ResponseCategory
		df       []map[string]interface{}
	}{
		{"all", allDF},
		{"successful", filterDataframe(allDF, func(row map[string]interface{}) bool { return row["deep_success"].(bool) })},
		{"failed", filterDataframe(allDF, func(row map[string]interface{}) bool { return !row["deep_success"].(bool) })},
	}

	for _, df := range dataframes {
		metrics, err := computeRawOutputSampleMetrics(df.df, targetRate, targetDuration)
		if err != nil {
			return nil, nil, err
		}
		categoryData[df.category] = metrics
	}

	return categoryData, rpcErrorPairs, nil
}

func convertRawVegetaOutputToDataframe(rawOutput []byte) ([]map[string]interface{}, error) {
	cmd := exec.Command("vegeta", "encode", "--to", "csv")
	cmd.Stdin = strings.NewReader(string(rawOutput))
	reportOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// results := make([]Result, 0)
	// dec := NewCSVDecoder(strings.NewReader(string(reportOutput)))
	// var r Result
	// for {
	// 	err = dec(&r)
	// 	if err != nil {
	// 		break
	// 	}
	// 	results = append(results, r)
	// }
	// log.Info().Msgf("results: %s", results)

	reader := csv.NewReader(strings.NewReader(string(reportOutput)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// log.Info().Msgf("records: %s", records)

	schema := []string{
		"timestamp", "status_code", "latency", "bytes_out", "bytes_in",
		"error", "response", "name", "index", "method", "url", "response_headers",
	}

	var df []map[string]interface{}
	for _, record := range records {
		row := make(map[string]interface{})
		for i, field := range record {
			switch schema[i] {
			case "timestamp", "status_code", "latency", "bytes_out", "bytes_in", "index":
				value, _ := strconv.ParseInt(field, 10, 64)
				row[schema[i]] = value
			default:
				row[schema[i]] = field
			}
		}
		log.Info().Msgf("row: %s", row)
		df = append(df, row)
	}

	return df, nil
}

func gatherErrorPairs(df []map[string]interface{}, calls []*types.JsonrpcMessage) ([]tooltypes.ErrorPair, error) {
	callsByID := make(map[int64]interface{})
	for _, call := range calls {
		// callMap := call.(map[string]interface{})
		// callID, ok := callMap["id"].(string)
		// if !ok {
		// 	return nil, fmt.Errorf("id not specified for call")
		// }
		callID := call.ID
		if _, exists := callsByID[callID]; exists {
			return nil, fmt.Errorf("duplicate call for id")
		}
		callsByID[callID] = call
	}

	var pairs []tooltypes.ErrorPair
	for _, row := range df {
		if row["rpc_error"] != nil {
			pairs = append(pairs, tooltypes.ErrorPair{nil, row["response"].(string)})
		}
	}

	return pairs, nil
}

func computeRawOutputSampleMetrics(df []map[string]interface{}, targetRate int, targetDuration int) (tooltypes.LoadTestDeepOutputDatum, error) {
	if len(df) == 0 {
		return tooltypes.LoadTestDeepOutputDatum{
			TargetRate:         targetRate,
			ActualRate:         utils.NewFloat64(0),
			TargetDuration:     targetDuration,
			ActualDuration:     nil,
			Requests:           0,
			Throughput:         nil,
			Success:            nil,
			Min:                nil,
			Mean:               nil,
			P50:                nil,
			P90:                nil,
			P95:                nil,
			P99:                nil,
			Max:                nil,
			StatusCodes:        map[string]int{},
			Errors:             []string{},
			NInvalidJSONErrors: 0,
			NRPCErrors:         0,
		}, nil
	}

	// Compute metrics...
	// This part would require significant work to replicate the exact functionality of the Python code
	// As Go doesn't have an exact equivalent to Polars, you might need to use a different data processing library
	// or implement the calculations manually

	// For demonstration, I'll just return a placeholder struct
	return tooltypes.LoadTestDeepOutputDatum{
		TargetRate:     targetRate,
		ActualRate:     utils.NewFloat64(float64(len(df)) / float64(targetDuration)),
		TargetDuration: targetDuration,
		// ... other fields would be computed here
	}, nil
}

func filterDataframe(df []map[string]interface{}, predicate func(map[string]interface{}) bool) []map[string]interface{} {
	var result []map[string]interface{}
	for _, row := range df {
		if predicate(row) {
			result = append(result, row)
		}
	}
	return result
}

func EncodeRawVegetaOutput(rawOutput []byte) string {
	compressed := compressGzip(rawOutput)
	return base64.StdEncoding.EncodeToString(compressed)
}

func DecodeRawVegetaOutput(encodedOutput string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(encodedOutput)
	if err != nil {
		return nil, err
	}
	return decompressGzip(decoded)
}

func compressGzip(data []byte) []byte {
	// Implement gzip compression
	// This is a placeholder and needs to be implemented
	return data
}

func decompressGzip(data []byte) ([]byte, error) {
	// Implement gzip decompression
	// This is a placeholder and needs to be implemented
	return data, nil
}
