package benchmarker

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/unifralabs/unifra-benchmark-tool/config"
	"github.com/unifralabs/unifra-benchmark-tool/outputter"
	"github.com/unifralabs/unifra-benchmark-tool/rpc_builder"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
	"github.com/unifralabs/unifra-benchmark-tool/vegeta"
)

type RpcBenchmarker struct {
	cfg   *config.EnvConfig
	nodes tooltypes.Nodes
}

func NewRpcBenchmarker(cfg *config.EnvConfig, nodes tooltypes.Nodes) (*RpcBenchmarker, error) {
	return &RpcBenchmarker{
		cfg:   cfg,
		nodes: nodes,
	}, nil
}

func (b *RpcBenchmarker) Run() error {

	param := tooltypes.TestGenerationParameters{
		TestName:   b.cfg.TestName,
		RandomSeed: tooltypes.RandomSeed(time.Now().UnixNano()),
		Rates:      []int{100},
		Durations:  []int{5},
		VegetaArgs: nil,
	}
	attacks, err := rpc_builder.GenerateTestEthGetBalance(param)
	if err != nil {
		return fmt.Errorf("error generating test: %w", err)
	}

	loadTest := tooltypes.LoadTest{
		TestParameters: param,
		Attacks:        attacks,
	}
	output, err := RunRpcBenchmarks(b.nodes, loadTest, true, []tooltypes.DeepOutput{})

	if err != nil {
		log.Info().Msgf("Error running vegeta attack: %s", err)
		return err
	}

	// log.Info().Msgf("output: %s", output)

	tStart := time.Now()
	outputter.SaveSingleRunResults(b.cfg.OutputDir, b.nodes, output, true, loadTest.TestParameters.TestName, tStart.Unix(), time.Now().Unix())

	return nil
}

func RunRpcBenchmarks(
	parsedNodes tooltypes.Nodes,
	test tooltypes.LoadTest,
	verbose bool,
	includeDeepOutput []tooltypes.DeepOutput,
) (map[string]tooltypes.LoadTestOutput, error) {

	results := make(map[string]tooltypes.LoadTestOutput)

	for _, parsedNode := range parsedNodes {
		result, err := runLoadTestLocally(parsedNode, test, verbose, includeDeepOutput)
		if err != nil {
			return nil, err
		}
		results[parsedNode.Name] = result
	}

	return results, nil
}

func runLoadTestLocally(
	node tooltypes.Node,
	test tooltypes.LoadTest,
	verbose bool,
	includeDeepOutput []tooltypes.DeepOutput,
) (tooltypes.LoadTestOutput, error) {
	if verbose {
		utils.PrintTimestamped(fmt.Sprintf("Running load test for %s", node.Name))
	}

	results := []*tooltypes.LoadTestOutputDatum{}

	for _, attack := range test.Attacks {
		if verbose {
			utils.PrintTimestamped(fmt.Sprintf("Running attack at rate = %d rps", attack.Rate))
		}

		result, err := vegeta.RunVegetaAttack(
			node.URL,
			attack.Rate,
			attack.Calls,
			attack.Duration,
			attack.VegetaArgs,
			verbose,
			includeDeepOutput,
		)
		if err != nil {
			return tooltypes.LoadTestOutput{}, err
		}

		results = append(results, result)

		if verbose {
		}
	}

	// Format output
	outputData := tooltypes.BuildLoadTestOutput(results)

	// Format deep output
	// if len(includeDeepOutput) == 0 {
	// 	for key := range outputData {
	// 		if len(key) >= 5 && key[:5] == "deep_" {
	// 			outputData[key] = nil
	// 		}
	// 	}
	// } else {
	// 	outputData["deep_rpc_error_pairs"] = make([][]tooltypes.ErrorPair, len(results))
	// 	for i, result := range results {
	// 		outputData["deep_rpc_error_pairs"].([][]tooltypes.ErrorPair)[i] = result.DeepRPCErrorPairs
	// 	}

	// 	categories := []tooltypes.ResponseCategory{tooltypes.AllResponses, tooltypes.SuccessfulResponses, tooltypes.FailedResponses}
	// 	deepMetrics := make(map[tooltypes.ResponseCategory]tooltypes.LoadTestDeepOutput)
	// 	for _, category := range categories {
	// 		categoryMetrics := make([]tooltypes.LoadTestDeepOutputDatum, len(results))
	// 		for i, result := range results {
	// 			categoryMetrics[i] = result.DeepMetrics[category]
	// 		}
	// 		deepMetrics[category] = listOfMapsToMapOfLists(categoryMetrics).(tooltypes.LoadTestDeepOutput)
	// 	}
	// 	outputData["deep_metrics"] = deepMetrics
	// }

	return outputData, nil
}
