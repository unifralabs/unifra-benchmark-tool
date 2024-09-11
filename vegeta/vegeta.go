package vegeta

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/unifralabs/unifra-benchmark-tool/types"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"github.com/unifralabs/unifra-benchmark-tool/utils"
)

func RunVegetaAttack(url string, rate int, calls []*types.JsonrpcMessage, duration int, vegetaArgs *string, verbose bool, includeDeepOutput []tooltypes.DeepOutput) (*tooltypes.LoadTestOutputDatum, error) {
	attack, err := constructVegetaAttack(calls, url, nil, verbose)
	if err != nil {
		return nil, err
	}

	attackOutput, err := vegetaAttack(attack["schedule_dir"], &duration, &rate, nil, nil, nil, nil, vegetaArgs, verbose)
	if err != nil {
		return nil, err
	}

	report, err := createVegetaReport(attackOutput, rate, duration, includeDeepOutput, calls)
	if err != nil {
		return nil, err
	}

	return report, nil
}

func constructVegetaAttack(calls []*types.JsonrpcMessage, url string, scheduleDir *string, verbose bool) (map[string]string, error) {
	headers := map[string]string{"Content-Type": "application/json"}

	if scheduleDir == nil {
		tempDir, err := ioutil.TempDir("", "vegeta_attack")
		if err != nil {
			return nil, err
		}
		scheduleDir = &tempDir
	}

	callPaths := []string{}
	for c, call := range calls {
		vegetaCallsPath := filepath.Join(*scheduleDir, fmt.Sprintf("vegeta_calls_%d.json", c))
		callJSON, err := json.Marshal(call)
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(vegetaCallsPath, callJSON, 0644); err != nil {
			return nil, err
		}
		callPaths = append(callPaths, vegetaCallsPath)
	}

	vegetaTargetsPath := filepath.Join(*scheduleDir, "vegeta_targets")
	f, err := os.Create(vegetaTargetsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	for c, _ := range calls {
		fmt.Fprintf(f, "POST %s\n", url)
		for key, value := range headers {
			fmt.Fprintf(f, "%s: %s\n", key, value)
		}
		fmt.Fprintf(f, "@%s\n", callPaths[c])
	}

	if verbose {
		log.Info().Msg("running vegeta attack...")
		log.Info().Msgf("- targets: %s", vegetaTargetsPath)
		log.Info().Msgf("- calls: %s", callPaths[len(callPaths)-1])
	}

	return map[string]string{
		"schedule_dir":        *scheduleDir,
		"vegeta_calls_path":   callPaths[len(callPaths)-1],
		"vegeta_targets_path": vegetaTargetsPath,
	}, nil
}

func vegetaAttack(scheduleDir string, duration *int, rate *int, maxConnections *int, maxWorkers *int, nCpus *int, reportPath *string, vegetaArgs *string, verbose bool) ([]byte, error) {
	log.Info().Msg("running vegeta attack...")
	cmd := []string{"vegeta", "attack"}
	cmd = append(cmd, "-targets="+filepath.Join(scheduleDir, "vegeta_targets"))

	if rate != nil {
		cmd = append(cmd, fmt.Sprintf("-rate=%d", *rate))
	}
	if duration != nil {
		cmd = append(cmd, fmt.Sprintf("-duration=%ds", *duration))
	}
	if maxConnections != nil {
		cmd = append(cmd, fmt.Sprintf("-max-connections=%d", *maxConnections))
	}
	if maxWorkers != nil {
		cmd = append(cmd, fmt.Sprintf("-max-workers=%d", *maxWorkers))
	}
	if vegetaArgs != nil {
		cmd = append(cmd, strings.Split(*vegetaArgs, " ")...)
	}

	if verbose {
		log.Info().Msgf("- command: %s", strings.Join(cmd, " "))
	}

	return exec.Command(cmd[0], cmd[1:]...).Output()
}

func createVegetaReport(attackOutput []byte, targetRate int, targetDuration int, includeDeepOutput []tooltypes.DeepOutput, calls []*types.JsonrpcMessage) (*tooltypes.LoadTestOutputDatum, error) {
	cmd := exec.Command("vegeta", "report", "-type", "json")
	cmd.Stdin = strings.NewReader(string(attackOutput))
	reportOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("reportOutput: %s", string(reportOutput))

	var report tooltypes.RawLoadTestOutputDatum
	if err := json.Unmarshal(reportOutput, &report); err != nil {
		return nil, err
	}

	var latencyMin *float64
	if min, ok := report.Latencies["min"]; ok {
		minValue := float64(min) / 1e9
		latencyMin = &minValue
	}

	var deepRawOutput *string
	var deepMetrics map[tooltypes.ResponseCategory]tooltypes.LoadTestDeepOutputDatum
	var deepRpcErrorPairs []tooltypes.ErrorPair

	if includeDeepOutput == nil {
		includeDeepOutput = []tooltypes.DeepOutput{}
	}

	for _, output := range includeDeepOutput {
		switch output {
		case "raw":
			encodedOutput := EncodeRawVegetaOutput(attackOutput)
			deepRawOutput = &encodedOutput
		case "metrics":
			var err error
			deepMetrics, deepRpcErrorPairs, err = ComputeDeepDatum(attackOutput, targetRate, targetDuration, calls)
			if err != nil {
				return nil, err
			}
		}
	}

	return &tooltypes.LoadTestOutputDatum{
		TargetRate:            targetRate,
		ActualRate:            utils.NewFloat64(report.Rate),
		TargetDuration:        targetDuration,
		ActualDuration:        utils.NewFloat64(float64(report.Duration) / 1e9),
		Requests:              report.Requests,
		Throughput:            utils.NewFloat64(report.Throughput),
		Success:               utils.NewFloat64(float64(report.Success)),
		Min:                   latencyMin,
		Mean:                  utils.NewFloat64(float64(report.Latencies["mean"]) / 1e9),
		P50:                   utils.NewFloat64(float64(report.Latencies["50th"]) / 1e9),
		P90:                   utils.NewFloat64(float64(report.Latencies["90th"]) / 1e9),
		P95:                   utils.NewFloat64(float64(report.Latencies["95th"]) / 1e9),
		P99:                   utils.NewFloat64(float64(report.Latencies["99th"]) / 1e9),
		Max:                   utils.NewFloat64(float64(report.Latencies["max"]) / 1e9),
		StatusCodes:           report.StatusCodes,
		Errors:                report.Errors,
		FirstRequestTimestamp: &report.Earliest,
		LastRequestTimestamp:  &report.Latest,
		LastResponseTimestamp: &report.End,
		FinalWaitTime:         utils.NewFloat64(float64(report.Wait) / 1e9),
		DeepRawOutput:         deepRawOutput,
		DeepMetrics:           deepMetrics,
		DeepRPCErrorPairs:     deepRpcErrorPairs,
	}, nil
}
