package constants

var PlotColors = map[string][]string{
	"green_shades": {
		"forestgreen",
		"limegreen",
		"chartreuse",
	},
	"red_shades": {
		"firebrick",
		"red",
		"salmon",
	},
	"purple_shades": {
		"rebeccapurple",
		"blueviolet",
		"mediumslateblue",
	},
	"orange_shades": {
		"darkgoldenrod",
		"darkorange",
		"gold",
	},
	"blue_shades": {
		"blue",
		"dodgerblue",
		"lightskyblue",
	},
	"streetlight": {
		"crimson",
		"goldenrod",
		"limegreen",
	},
}

func GetPlotColors() map[string][]string {
	return PlotColors
}
