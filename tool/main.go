package main

import (
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"tool/preprocessor"
	tile_proxy "tool/tile-proxy"
)

var cli struct {
	Debug         bool `help:"Enable debug mode." short:"d"`
	Preprocessing struct {
		Input  string `help:"The input file. Either .osm or .osm..pbf." placeholder:"<input-file>" arg:""`
		Output string `help:"The output file, which must be a .osm.pbf file." placeholder:"<output-file>" arg:""`
	} `cmd:"" help:"Preprocesses the OSM data by adding e.g. label nodes."`
	TileProxy struct {
		Mappings    []string `help:"A list of URL mappings of the following form: \"<endpoint1>:<url1> <endpoint2>:<url2> ...\". Each mapping will result in an API endpoint of the form http://localhost:<port>/<endpoint>/..." arg:""`
		Port        string   `help:"The port of the proxy on localhost." default:"9000" short:"p"`
		CacheFolder string   `help:"A folder in which tiles will be cached." default:".tile-cache" short:"c"`
	} `cmd:"" help:"A proxy converting remote tiles into a given image format."`
}

func main() {
	ctx := readCliArgs()

	switch ctx.Command() {
	case "preprocessing <input> <output>":
		preprocessor.PreprocessData(cli.Preprocessing.Input, cli.Preprocessing.Output)
	case "tile-proxy <mappings>":
		tile_proxy.StartProxy(cli.TileProxy.Port, cli.TileProxy.Mappings, cli.TileProxy.CacheFolder)
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
