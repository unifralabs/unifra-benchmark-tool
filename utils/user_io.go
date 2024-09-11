package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/rs/zerolog/log"
	"github.com/unifralabs/unifra-benchmark-tool/constants"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
)

var styles = map[string]string{
	"title":       "bold #00e100",
	"metavar":     "bold #e5e9f0",
	"description": "#aaaaaa",
	"content":     "#00B400",
	"option":      "bold #e5e9f0",
	"comment":     "#888888",
}

func GetNodesPlotColors(nodes map[string]tooltypes.Node) map[string]string {
	colors := make(map[string]string)
	taken := make(map[string]bool)

	for _, node := range nodes {
		var version string
		if node.ClientVersion != "" {
			version = node.ClientVersion
		}

		var color string
		if (version != "" && strings.Contains(version, "reth")) || strings.Contains(node.Name, "reth") {
			if !taken["orange_shades"] {
				color = "orange_shades"
			}
		} else if (version != "" && strings.Contains(version, "erigon")) || strings.Contains(node.Name, "erigon") {
			if !taken["blue_shades"] {
				color = "blue_shades"
			}
		}

		if color == "" {
			for colorName := range constants.PlotColors {
				if !taken[colorName] {
					color = colorName
					break
				}
			}
		}

		if color == "" {
			panic("out of colors")
		}

		colors[node.Name] = color
		taken[color] = true
	}

	return colors
}

func PrintMetricTables(
	results map[string]interface{},
	metrics []string,
	suffix string,
	decimals *int,
	comparison *bool,
	indent *int,
) {
	if len(results) == 0 {
		log.Info().Msg(strings.Repeat(" ", *indent) + "no results")
		return
	}

	if comparison == nil {
		comp := len(results) == 2
		comparison = &comp
	}

	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}

	rates := results[names[0]].(map[string]interface{})["target_rate"].([]int)

	for _, metric := range metrics {
		var metricSuffix string
		switch metric {
		case "success", "n_invalid_json_errors", "n_rpc_errors":
			metricSuffix = ""
		case "throughput":
			metricSuffix = " (rps)"
		default:
			metricSuffix = " (s)"
		}

		unittedNames := make([]string, len(names))
		for i, name := range names {
			unittedNames[i] = name + metricSuffix
		}

		labels := append([]string{"rate (rps)"}, unittedNames...)
		if *comparison {
			if len(results) != 2 {
				panic("comparison of >2 tests not implemented")
			}
			labels = append(labels, names[0]+" / "+names[1])
		}

		rows := make([][]string, len(rates))
		for i, rate := range rates {
			row := make([]string, 1, len(labels))
			row[0] = fmt.Sprintf("%d", rate)
			rows[i] = row
		}

		values := make([]float64, 0)
		for _, name := range names {
			metricValues := results[name].(map[string]interface{})[metric].([]float64)
			for i, value := range metricValues {
				rows[i] = append(rows[i], fmt.Sprintf("%.6f", value))
				values = append(values, value)
			}
		}

		if *comparison {
			for i := range rows {
				v1, _ := strconv.ParseFloat(rows[i][1], 64)
				v2, _ := strconv.ParseFloat(rows[i][2], 64)
				rows[i] = append(rows[i], fmt.Sprintf("%.1f%%", v1/v2*100))
			}
		}

		// useDecimals := 6
		// if decimals != nil {
		// 	useDecimals = *decimals
		// } else if allGreaterThanOne(values) {
		// 	useDecimals = 1
		// }

		// Print header
		log.Info().Msg(strings.Repeat(" ", *indent) + "+" + strings.Repeat("-", len(metric+" vs load"+suffix)+2) + "+")
		log.Info().Msg(strings.Repeat(" ", *indent) + "| " + color.New(color.FgHiWhite, color.Bold).Sprint(metric+" vs load"+suffix) + " |")
		log.Info().Msg(strings.Repeat(" ", *indent) + "+" + strings.Repeat("-", len(metric+" vs load"+suffix)+2) + "+")

		// Print table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(labels)
		table.SetBorder(false)
		alignment := make([]int, len(labels))
		for i := range labels {
			alignment[i] = tablewriter.ALIGN_RIGHT
		}
		table.SetColumnAlignment(alignment)
		table.AppendBulk(rows)
		table.Render()

		if metric != metrics[len(metrics)-1] {
		}
	}
}

func allGreaterThanOne(values []float64) bool {
	for _, v := range values {
		if v <= 1 && v != 0 {
			return false
		}
	}
	return true
}

func PrintTextBox(text string) {
	log.Info().Msg("+" + strings.Repeat("-", len(text)+2) + "+")
	log.Info().Msg("| " + color.New(color.FgHiWhite, color.Bold).Sprint(text) + " |")
	log.Info().Msg("+" + strings.Repeat("-", len(text)+2) + "+")
}

func PrintHeader(text string) {
	log.Info().Msg(color.New(color.FgHiWhite, color.Bold).Sprint(text))
	log.Info().Msg(strings.Repeat("-", len(text)))
}

func PrintBullet(text string) {
	log.Info().Msg("• " + text)
}

func PrintTable(rows [][]string, headers []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.AppendBulk(rows)
	table.Render()
}

func PrintMultilineTable(rows [][]string, headers []string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(rows)
	table.Render()
}

func PrintTimestamped(message string) {
	now := time.Now().Round(time.Second)
	timestamp := fmt.Sprintf("[%s]", now.Format("2006-01-02 15:04:05"))
	log.Info().Msgf("%s %s", color.New(color.FgHiWhite).Sprint(timestamp), message)
}

func DisableTextColors() {
	color.NoColor = true
}
