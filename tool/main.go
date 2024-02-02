package main

import (
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	legend_graphic "tool/legend-graphic"
	tile_proxy "tool/tile-proxy"
)

var cli struct {
	Debug          bool `help:"Enable debug mode." short:"d"`
	GenerateLegend struct {
		Input string `help:"The input schema file." placeholder:"<schema-file>"  arg:""`
	} `cmd:"" help:"Generated a PDF file with the map legend based on the schema file."`
	TileProxy struct {
		TargetUrl   string `help:"The target URL where the actual tiles are." short:"t"`
		Port        string `help:"The port of the proxy on localhost." default:"9000" short:"p"`
		CacheFolder string `help:"A folder in which tiles will be cached." default:".tile-cache" short:"c"`
	} `cmd:"" help:"A proxy converting remote tiles into a given image format."`
}

func main() {
	ctx := readCliArgs()

	switch ctx.Command() {
	case "generate-legend <input>":
		legend_graphic.GenerateLegendGraphic(cli.GenerateLegend.Input)
	case "tile-proxy":
		tile_proxy.StartProxy(cli.TileProxy.Port, cli.TileProxy.TargetUrl, cli.TileProxy.CacheFolder)
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
