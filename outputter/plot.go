package outputter

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"

	"github.com/unifralabs/unifra-benchmark-tool/constants"
	tooltypes "github.com/unifralabs/unifra-benchmark-tool/types"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/font/liberation"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

func PlotLoadTestResults(
	outputs map[string]tooltypes.LoadTestOutput,
	testName string,
	outputDir string,
	latencyYscaleLog bool,
	colors map[string]string,
	titleSuffix string,
	fileSuffix string,
	plotSuccessRate bool,
	plotThroughput bool,
	plotLatency bool,
) error {
	if titleSuffix == "" {
		titleSuffix = ""
	}
	if fileSuffix == "" {
		fileSuffix = ""
	}

	font.DefaultCache.Add(liberation.Collection())
	// cache := font.NewCache(liberation.Collection())
	face := font.DefaultCache.Lookup(font.Font{Typeface: "Liberation", Variant: "Mono"}, 12)
	f := face.Font

	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	if plotSuccessRate {
		p := plot.New()
		setPlotFont(p, f)
		setPlotFontSize(p)

		if err := PlotLoadTestSuccess(p, outputs, colors, &testName); err != nil {
			return fmt.Errorf("failed to plot success rate: %w", err)
		}
		if outputDir != "" {
			path := filepath.Join(outputDir, "success_rate"+fileSuffix+".png")
			if err := savePlot(p, path); err != nil {
				return fmt.Errorf("failed to save success rate plot: %w", err)
			}
		}
	}

	if plotThroughput {
		p := plot.New()
		setPlotFont(p, f)
		setPlotFontSize(p)
		if err := PlotLoadTestThroughput(p, outputs, colors, &testName); err != nil {
			return fmt.Errorf("failed to plot throughput: %w", err)
		}
		if outputDir != "" {
			path := filepath.Join(outputDir, "throughput"+fileSuffix+".png")
			if err := savePlot(p, path); err != nil {
				return fmt.Errorf("failed to save throughput plot: %w", err)
			}
		}
	}

	if plotLatency {
		p := plot.New()
		setPlotFont(p, f)
		setPlotFontSize(p)
		if err := PlotLoadTestLatencies(p, outputs, nil, &testName, latencyYscaleLog, colors); err != nil {
			return fmt.Errorf("failed to plot latencies: %w", err)
		}
		if outputDir != "" {
			path := filepath.Join(outputDir, "latencies"+fileSuffix+".png")
			if err := savePlot(p, path); err != nil {
				return fmt.Errorf("failed to save latencies plot: %w", err)
			}
		}
	}

	// Deep graphs
	// hasDeepOutputs := false
	// for _, output := range outputs {
	// 	if output.DeepMetrics != nil {
	// 		hasDeepOutputs = true
	// 		break
	// 	}
	// }

	// if hasDeepOutputs {
	// 	successfulOutputs := make(map[string]tooltypes.LoadTestOutput)
	// 	failedOutputs := make(map[string]tooltypes.LoadTestOutput)
	// 	for name, output := range outputs {
	// 		if output.DeepMetrics != nil {
	// 			successfulOutputs[name] = output.DeepMetrics["successful"]
	// 			failedOutputs[name] = output.DeepMetrics["failed"]
	// 		}
	// 	}

	// 	if err := PlotLoadTestResults(
	// 		successfulOutputs,
	// 		testName,
	// 		outputDir,
	// 		latencyYscaleLog,
	// 		colors,
	// 		", successful calls only",
	// 		"_successful_calls",
	// 		false,
	// 		false,
	// 		true,
	// 	); err != nil {
	// 		return fmt.Errorf("failed to plot successful calls: %w", err)
	// 	}

	// 	if err := PlotLoadTestResults(
	// 		failedOutputs,
	// 		"failed "+testName,
	// 		outputDir,
	// 		latencyYscaleLog,
	// 		colors,
	// 		", failed calls only",
	// 		"_failed_calls",
	// 		false,
	// 		false,
	// 		true,
	// 	); err != nil {
	// 		return fmt.Errorf("failed to plot failed calls: %w", err)
	// 	}
	// }

	return nil
}

func savePlot(p *plot.Plot, filename string) error {
	return p.Save(12*vg.Inch, 12*vg.Inch, filename)
}

// PlotLoadTestSuccess plots the success rate
func PlotLoadTestSuccess(
	p *plot.Plot,
	results map[string]tooltypes.LoadTestOutput,
	colors map[string]string,
	testName *string,
) error {
	err := PlotLoadTestResultMetrics(
		p,
		results,
		[]string{"success"},
		PlotOptions{
			Colors:   colors,
			TestName: testName,
			Title:    "Success Rate vs Request Rate\n(higher is better)",
			YLabel:   "success rate",
			YLim:     []float64{-0.03, 1.03},
		},
	)

	p.Legend.Top = true
	p.Legend.Left = false
	p.Legend.YPosition = 0
	return err
}

func PlotLoadTestThroughput(
	p *plot.Plot,
	results map[string]tooltypes.LoadTestOutput,
	colors map[string]string,
	testName *string,
) error {
	zero := 0.0
	err := PlotLoadTestResultMetrics(
		p,
		results,
		[]string{"throughput"},
		PlotOptions{
			Colors:   colors,
			TestName: testName,
			Title:    "Throughput vs Request Rate\n(higher is better)",
			YLabel:   "throughput\n(responses per second)",
			YMin:     &zero,
		},
	)

	p.Legend.Top = true
	p.Legend.Left = true
	return err
}

func PlotLoadTestLatencies(
	p *plot.Plot,
	results map[string]tooltypes.LoadTestOutput,
	metrics []string,
	testName *string,
	yscaleLog bool,
	colors map[string]string,
) error {
	var ymin *float64
	if !yscaleLog {
		zero := 0.0
		ymin = &zero
	}

	if metrics == nil {
		metrics = []string{"p99", "p90", "p50"}
	}

	err := PlotLoadTestResultMetrics(
		p,
		results,
		metrics,
		PlotOptions{
			Colors:    colors,
			TestName:  testName,
			YMin:      ymin,
			Title:     "Latency vs Request Rate\n(lower is better)",
			YLabel:    "latency (seconds)",
			YScaleLog: yscaleLog,
		},
	)

	p.Legend.Top = true
	p.Legend.Left = true
	return err
}

type PlotOptions struct {
	Colors    map[string]string
	TestName  *string
	Title     string
	YLabel    string
	YLim      []float64
	YMin      *float64
	YScaleLog bool
}

func PlotLoadTestResultMetrics(
	p *plot.Plot,
	results map[string]tooltypes.LoadTestOutput,
	metrics []string,
	options PlotOptions,
) error {

	plotColors := constants.GetPlotColors()

	keys := make([]string, 0, len(plotColors))
	for k := range plotColors {
		keys = append(keys, k)
	}

	if options.Colors == nil {
		options.Colors = make(map[string]string)
		index := 0
		for key, _ := range results {
			options.Colors[key] = keys[index]
			index++
		}
	}

	for name, result := range results {
		resultColors, err := determineColors(options.Colors[name], metrics, plotColors)
		if err != nil {
			return err
		}

		for zorder, metric := range metrics {
			label := name
			if len(metrics) > 1 {
				label += " " + metric
			}

			pts := make(plotter.XYs, len(result.TargetRate))
			for i := range result.TargetRate {
				pts[i].X = float64(result.TargetRate[i])
				var y float64
				switch metric {
				case "success":
					y = *result.Success[i]
				case "throughput":
					y = *result.Throughput[i]
				case "p99":
					y = *result.P99[i]
				case "p90":
					y = *result.P90[i]
				case "p50":
					y = *result.P50[i]

				}
				pts[i].Y = y
			}

			line, points, err := plotter.NewLinePoints(pts)
			if err != nil {
				return err
			}

			line.Color = resultColors[zorder]
			points.Color = resultColors[zorder]
			points.Shape = draw.CircleGlyph{}
			points.Radius = vg.Points(5)

			p.Add(line, points)
			p.Legend.Add(label, line, points)
		}
	}

	// Set labels and options
	p.Title.Text = options.Title
	p.X.Label.Text = "requests per second"
	if options.TestName != nil {
		p.X.Label.Text += "\n[" + *options.TestName + "]"
	}
	p.Y.Label.Text = options.YLabel

	if options.YScaleLog {
		p.Y.Scale = plot.LogScale{}
	}

	if len(options.YLim) == 2 {
		p.Y.Min = options.YLim[0]
		p.Y.Max = options.YLim[1]
	}

	if options.YMin != nil {
		p.Y.Min = *options.YMin
	}

	AddTickGrid(p)

	return nil
}

func determineColors(c string, metrics []string, plotColors map[string][]string) ([]color.Color, error) {

	if colors, ok := plotColors[c]; ok {
		if len(metrics) == 1 {
			return []color.Color{parseColor(colors[1])}, nil
		}
		return parseColors(colors), nil
	}
	return []color.Color{parseColor(c)}, nil

}

func parseColor(s string) color.Color {
	return constants.Colors[s]
}

func parseColors(s []string) []color.Color {
	colors := make([]color.Color, len(s))
	for i, c := range s {
		colors[i] = constants.Colors[c]
	}
	return colors
}

func setPlotFontSize(p *plot.Plot) {
	const (
		SMALL_SIZE  = 18
		MEDIUM_SIZE = 20
		BIGGER_SIZE = 24
	)

	// Set the title font size (equivalent to 'axes.titlesize')
	p.Title.TextStyle.Font.Size = vg.Points(BIGGER_SIZE)

	// Set the axis label font sizes (equivalent to 'axes.labelsize')
	p.X.Label.TextStyle.Font.Size = vg.Points(MEDIUM_SIZE)
	p.Y.Label.TextStyle.Font.Size = vg.Points(MEDIUM_SIZE)

	// Set the tick label font sizes (equivalent to 'xtick.labelsize' and 'ytick.labelsize')
	p.X.Tick.Label.Font.Size = vg.Points(SMALL_SIZE)
	p.Y.Tick.Label.Font.Size = vg.Points(SMALL_SIZE)

	// Set the legend font size (equivalent to 'legend.fontsize')
	p.Legend.TextStyle.Font.Size = vg.Points(SMALL_SIZE)
}

func setPlotFont(p *plot.Plot, font font.Font) {
	// Set font for the title
	p.Title.TextStyle.Font = font

	// Set font for axis labels
	p.X.Label.TextStyle.Font = font
	p.Y.Label.TextStyle.Font = font

	// Set font for tick labels
	p.X.Tick.Label.Font = font
	p.Y.Tick.Label.Font = font

	// Set font for legend
	p.Legend.TextStyle.Font = font
}

func AddTickGrid(p *plot.Plot) {
	alpha := 0.1

	color := constants.Colors["black"]
	linestyle := "--"
	linewidth := vg.Points(1)
	xtickGrid := true
	ytickGrid := true

	SanitizeRange(&p.X)
	SanitizeRange(&p.Y)

	xmin, xmax := p.X.Min, p.X.Max
	ymin, ymax := p.Y.Min, p.Y.Max

	// log.Debug().Msgf("xmin: %f, xmax: %f, ymin: %f, ymax: %f", xmin, xmax, ymin, ymax)

	if xtickGrid {
		for _, tick := range p.X.Tick.Marker.Ticks(xmin, xmax) {
			xys := plotter.XYs{
				{X: tick.Value, Y: ymin},
				{X: tick.Value, Y: ymax},
			}
			line, err := plotter.NewLine(xys)
			if err != nil {
				panic(fmt.Errorf("tick.Value: %f, ymin: %f, ymax: %f, %v", tick.Value, ymin, ymax, err))
			}
			line.Color = color2RGBA(color, alpha)
			line.Width = linewidth
			line.Dashes = dashesFromLinestyle(linestyle)
			p.Add(line)
		}
	}

	if ytickGrid {
		for _, tick := range p.Y.Tick.Marker.Ticks(ymin, ymax) {
			xys := plotter.XYs{
				{X: xmin, Y: tick.Value},
				{X: xmax, Y: tick.Value},
			}
			line, err := plotter.NewLine(xys)
			if err != nil {
				panic(fmt.Errorf("tick.Value: %f, xmin: %f, xmax: %f, %v", tick.Value, xmin, xmax, err))
			}
			line.Color = color2RGBA(color, alpha)
			line.Width = linewidth
			line.Dashes = dashesFromLinestyle(linestyle)
			p.Add(line)
		}
	}

	// Restore plot limits
	p.X.Min, p.X.Max = xmin, xmax
	p.Y.Min, p.Y.Max = ymin, ymax
}

func SanitizeRange(a *plot.Axis) {
	if math.IsInf(a.Min, 0) {
		a.Min = 0
	}
	if math.IsInf(a.Max, 0) {
		a.Max = 0
	}
	if a.Min > a.Max {
		a.Min, a.Max = a.Max, a.Min
	}
	if a.Min == a.Max {
		a.Min--
		a.Max++
	}

	if a.AutoRescale {
		marks := a.Tick.Marker.Ticks(a.Min, a.Max)
		for _, t := range marks {
			a.Min = math.Min(a.Min, t.Value)
			a.Max = math.Max(a.Max, t.Value)
		}
	}
}

// Helper function to convert color string and alpha to color.RGBA
func color2RGBA(rgba color.Color, alpha float64) color.RGBA {
	r, g, b, _ := rgba.RGBA()
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(alpha * 255)}
}

// Helper function to convert linestyle string to dashes
func dashesFromLinestyle(linestyle string) []vg.Length {
	switch linestyle {
	case "--":
		return []vg.Length{vg.Points(6), vg.Points(4)}
	case "-":
		return nil
	// Add more cases as needed
	default:
		return nil
	}
}
