package outputter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/unifralabs/unifra-benchmark-tool/constants"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

func SaveSingleRunResults(
	outputDir string,
	nodes tooltypes.Nodes,
	results map[string]tooltypes.LoadTestOutput,
	figures bool,
	testName string,
	tRunStart int64,
	tRunEnd int64,
) (tooltypes.SingleRunResultsPayload, error) {
	if !isDir(outputDir) {
		if fileExists(outputDir) {
			return tooltypes.SingleRunResultsPayload{}, fmt.Errorf("output must be a directory path")
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return tooltypes.SingleRunResultsPayload{}, err
		}
	}

	path := filepath.Join(outputDir, constants.RPC_OUTPUT_FILE)
	payload := tooltypes.SingleRunResultsPayload{
		CLIArgs:   os.Args,
		Type:      "single_test",
		TRunStart: tRunStart,
		TRunEnd:   tRunEnd,
		Nodes:     nodes,
		Results:   results,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return tooltypes.SingleRunResultsPayload{}, err
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return tooltypes.SingleRunResultsPayload{}, err
	}

	if figures {
		figuresDir := filepath.Join(outputDir, "figures")
		colors := utils.GetNodesPlotColors(nodes)
		if err := PlotLoadTestResults(results, testName, figuresDir, true, colors, "", "", true, true, true); err != nil {
			return tooltypes.SingleRunResultsPayload{}, err
		}
	}

	return payload, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
