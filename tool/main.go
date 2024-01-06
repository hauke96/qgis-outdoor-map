package main

import (
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	legend_graphic "tool/legend-graphic"
	"tool/preprocessor"
)

var cli struct {
	Debug         bool `help:"Enable debug mode." short:"d"`
	Preprocessing struct {
		Input  string `help:"The input file. Either .osm or .osm..pbf." placeholder:"<input-file>"  arg:""`
		Output string `help:"The output file, which must be a .osm.pbf file." placeholder:"<output-file>"  arg:""`
	} `cmd:"" help:"Preprocesses the OSM data by adding e.g. label nodes."`
	GenerateLegend struct {
		Input string `help:"The input schema file." placeholder:"<schema-file>"  arg:""`
	} `cmd:"" help:"Generated a PDF file with the map legend based on the schema file."`
}

func main() {
	ctx := readCliArgs()

	switch ctx.Command() {
	case "preprocessing <input> <output>":
		preprocessor.PreprocessData(cli.Preprocessing.Input, cli.Preprocessing.Output)
	case "generate-legend <input>":
		legend_graphic.GenerateLegendGraphic(cli.GenerateLegend.Input)
	default:
		sigolo.Fatal("Unknown command: %v", ctx.Command())
	}
}

func readCliArgs() *kong.Context {
	ctx := kong.Parse(
		&cli,
		kong.Name("Outdoor and hiking map utility"),
		kong.Description("A CLI tool to process the OSM data of the outdoor map and to generate a legend graphic."),
	)

	if cli.Debug {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	}

	return ctx
}
