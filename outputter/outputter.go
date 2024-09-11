package outputter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/unifralabs/unifra-benchmark-tool/constants"
	"github.com/unifralabs/unifra-benchmark-tool/stats"
)

type outputFormat struct {
	AverageTPS float64           `json:"averageTPS"`
	Blocks     []stats.BlockInfo `json:"blocks"`
}

func OutputData(data *stats.CollectorData, outputDir string) error {
	log.Info().Msg("💾 Saving run results initialized 💾")

	if !isDir(outputDir) {
		if fileExists(outputDir) {
			return fmt.Errorf("output must be a directory path")
		}
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return err
		}
	}

	blocks := make([]stats.BlockInfo, 0, len(data.BlockInfo))
	for _, block := range data.BlockInfo {
		blocks = append(blocks, *block)
	}

	output := outputFormat{
		AverageTPS: data.TPS,
		Blocks:     blocks,
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("unable to marshal output data: %v", err)
	}

	path := filepath.Join(outputDir, constants.EOA_OUTPUT_FILE)
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("unable to write output to file: %v", err)
	}

	log.Info().Msgf("✅ Run results saved to %s", path)
	return nil
}
