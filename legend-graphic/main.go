package main

import (
	"encoding/json"
	"encoding/xml"
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"github.com/paulmach/osm"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var cli struct {
	Debug bool   `help:"Enable debug mode." short:"d"`
	Input string `help:"The input schema file." short:"i"`
}

const (
	latOffset     = 0.001
	lonOffset     = 0.001
	featureHeight = 0.0001
	featureWidth  = 0.000225

	outputFileName = "features.osm.pbf"

	placeholderItems               = "__ITEMS__"
	placeholderCategoryTitle       = "__CAT__"
	placeholderCategoryId          = "__CAT-ID__"
	placeholderItemDesc            = "__DESC__"
	placeholderItemId              = "__TITLE-ID__"
	placeholderItemLat             = "__LAT__"
	placeholderItemLon             = "__LON__"
	placeholderItemZoom            = "__ZOOM__"
	placeholderAdditionalInfo      = "__ADDITIONAL_INFO__"
	placeholderAdditionalInfoTitle = "__ADDITIONAL_INFO_TITLE__"
	placeholderAdditionalInfoHtml  = "__ADDITIONAL_INFO_HTML__"
)

var (
	idCounter        int64 = 1
	categoryTemplate       = `			<h2>` + placeholderCategoryTitle + `</h2>
			` + placeholderItems + `
`
	itemTemplate = `
			<div class="item">
	        	<div class="map-container">
	        		<div id="map-` + placeholderCategoryId + `-` + placeholderItemId + `" class="map"></div>
	        	</div>
	        	<div class="description">` + placeholderItemDesc + `</div>
	        </div>
	        <script>
	        	createMap("map-` + placeholderCategoryId + `-` + placeholderItemId + `", { lat: ` + placeholderItemLat + `, lon: ` + placeholderItemLon + ` }, ` + placeholderItemZoom + `);
	        </script>
`
	additionalInfoTemplate = `
			<h1>` + placeholderAdditionalInfoTitle + `</h1>
			<div class="additional-info-container">
			` + placeholderAdditionalInfoHtml + `
			</div>
`
)

type Schema struct {
	Title           string           `json:"title"`
	Categories      []Category       `json:"categories"`
	AdditionalInfos []AdditionalInfo `json:"additional-infos"`
}

type Category struct {
	Title string `json:"title"`
	Items []Item `json:"items"`
}

type Item struct {
	Tags          [][2]string `json:"tags"`
	FeaturePerTag bool        `json:"feature-per-tag"`
	FeatureType   string      `json:"type"`
	Description   string      `json:"description"`
	OffsetLat     float64     `json:"offset-lataa"`
	OffsetLon     float64     `json:"offset-lonaa"`
}

type AdditionalInfo struct {
	Title string `json:"title"`
	Html  string `json:"html"`
}

func (i Item) getTags() osm.Tags {
	var tags []osm.Tag
	for _, tag := range i.Tags {
		tags = append(tags, osm.Tag{Key: tag[0], Value: tag[1]})
	}
	return tags
}

func main() {
	readCliArgs()

	schema := readSchemaFile()

	outputOsm := osm.OSM{
		Version: "0.6",
	}

	createFeaturesFromSchema(schema, &outputOsm)

	writeOsmToPbf(&outputOsm)

	generateVectorTiles()

	generatedHtml := generateLegendHtmlItems(schema)

	if len(schema.AdditionalInfos) > 0 {
		sigolo.Debug("Generate HTML for additional information")
	} else {
		sigolo.Debug("No additional information found")
	}
	additionalInfoHtml := ""
	for _, info := range schema.AdditionalInfos {
		sigolo.Debug("Generate HTML for additional info section '%s'", info.Title)
		infoHtml := additionalInfoTemplate
		infoHtml = strings.Replace(infoHtml, placeholderAdditionalInfoTitle, info.Title, 1)
		infoHtml = strings.Replace(infoHtml, placeholderAdditionalInfoHtml, info.Html, 1)
		additionalInfoHtml += infoHtml
	}

	createLegendHtmlFile(generatedHtml, additionalInfoHtml)
}

func readCliArgs() {
	kong.Parse(&cli)

	if cli.Debug {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	}

	if !strings.HasSuffix(cli.Input, ".json") {
		sigolo.Error("Input file must be an .json file")
		os.Exit(1)
	}
}

func readSchemaFile() Schema {
	schemaFile, err := os.ReadFile(cli.Input)
	sigolo.FatalCheck(err)

	schema := Schema{}
	err = json.Unmarshal(schemaFile, &schema)
	sigolo.FatalCheck(err)

	return schema
}

func createFeaturesFromSchema(schema Schema, outputOsm *osm.OSM) {
	for i, category := range schema.Categories {
		sigolo.Info("Process category %d: %s", i, category.Title)

		for j, item := range category.Items {
			sigolo.Debug("Category %d / %d: Process item '%s'", i, j, item.Description)

			if item.FeatureType == "polygon" {
				addPolygon(float64(i)*lonOffset, item.OffsetLon, float64(j)*latOffset, item.OffsetLat, item.getTags(), item.FeaturePerTag, outputOsm)
			} else if item.FeatureType == "line" {
				addLine(float64(i)*lonOffset, item.OffsetLon, float64(j)*latOffset, item.OffsetLat, item.getTags(), item.FeaturePerTag, outputOsm)
			} else if item.FeatureType == "point" {
				addPoint(float64(i)*lonOffset, item.OffsetLon, float64(j)*latOffset, item.OffsetLat, item.getTags(), item.FeaturePerTag, outputOsm)
			} else {
				sigolo.Fatal("Category %d / %d: Unknown item type '%s'", i, j, item.FeatureType)
			}
		}
	}
}

func addPolygon(originLon float64, offsetLon float64, originLat float64, offsetLat float64, tags []osm.Tag, featurePerTag bool, outputOsm *osm.OSM) {
	timestamp, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	sigolo.FatalCheck(err)

	nodeLowerLeft := toNode(originLon+offsetLon, originLat+offsetLat, nil, timestamp, outputOsm)
	idCounter++

	nodeUpperLeft := toNode(originLon+offsetLon, originLat+featureHeight+offsetLat, nil, timestamp, outputOsm)
	idCounter++

	nodeUpperRight := toNode(originLon+featureWidth+offsetLon, originLat+featureHeight+offsetLat, nil, timestamp, outputOsm)
	idCounter++

	nodeLowerRight := toNode(originLon+featureWidth+offsetLon, originLat+offsetLat, nil, timestamp, outputOsm)
	idCounter++

	if featurePerTag {
		for _, tag := range tags {
			way := &osm.Way{
				ID:        osm.WayID(idCounter),
				Version:   1,
				Timestamp: timestamp,
				Nodes: []osm.WayNode{
					toWayNode(nodeLowerLeft),
					toWayNode(nodeUpperLeft),
					toWayNode(nodeUpperRight),
					toWayNode(nodeLowerRight),
					toWayNode(nodeLowerLeft),
				},
				Tags: []osm.Tag{tag},
			}
			idCounter++
			outputOsm.Append(way)
		}
	} else {
		way := &osm.Way{
			ID:        osm.WayID(idCounter),
			Version:   1,
			Timestamp: timestamp,
			Nodes: []osm.WayNode{
				toWayNode(nodeLowerLeft),
				toWayNode(nodeUpperLeft),
				toWayNode(nodeUpperRight),
				toWayNode(nodeLowerRight),
				toWayNode(nodeLowerLeft),
			},
			Tags: tags,
		}
		idCounter++
		outputOsm.Append(way)
	}
}

func addLine(originLon float64, offsetLon float64, originLat float64, offsetLat float64, tags []osm.Tag, featurePerTag bool, outputOsm *osm.OSM) {
	timestamp, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	sigolo.FatalCheck(err)

	nodeLeft := toNode(originLon+offsetLon, originLat+featureHeight/2+offsetLat, nil, timestamp, outputOsm)
	idCounter++

	nodeRight := toNode(originLon+featureWidth+offsetLon, originLat+featureHeight/2+offsetLat, nil, timestamp, outputOsm)
	idCounter++

	if featurePerTag {
		for _, tag := range tags {
			way := &osm.Way{
				ID:        osm.WayID(idCounter),
				Version:   1,
				Timestamp: timestamp,
				Nodes: []osm.WayNode{
					toWayNode(nodeLeft),
					toWayNode(nodeRight),
				},
				Tags: []osm.Tag{tag},
			}
			idCounter++
			outputOsm.Append(way)
		}
	} else {
		way := &osm.Way{
			ID:        osm.WayID(idCounter),
			Version:   1,
			Timestamp: timestamp,
			Nodes: []osm.WayNode{
				toWayNode(nodeLeft),
				toWayNode(nodeRight),
			},
			Tags: tags,
		}
		idCounter++
		outputOsm.Append(way)
	}
}

func addPoint(originLon float64, offsetLon float64, originLat float64, offsetLat float64, tags []osm.Tag, featurePerTag bool, outputOsm *osm.OSM) {
	timestamp, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	sigolo.FatalCheck(err)

	if featurePerTag {
		for _, tag := range tags {
			node := toNode(originLon+featureWidth/2+offsetLon, originLat+featureHeight/2+offsetLat, []osm.Tag{tag}, timestamp, outputOsm)
			idCounter++
			outputOsm.Append(&node)
		}
	} else {
		node := toNode(originLon+featureWidth/2+offsetLon, originLat+featureHeight/2+offsetLat, tags, timestamp, outputOsm)
		idCounter++
		outputOsm.Append(&node)
	}
}

func toNode(originLon float64, originLat float64, tags []osm.Tag, timestamp time.Time, outputOsm *osm.OSM) osm.Node {
	node := osm.Node{
		Version:   1,
		ID:        osm.NodeID(idCounter),
		Timestamp: timestamp,
		Tags:      tags,
		Lon:       originLon,
		Lat:       originLat,
	}
	outputOsm.Append(&node)
	return node
}

func toWayNode(node osm.Node) osm.WayNode {
	return osm.WayNode{
		ID:      node.ID,
		Version: node.Version,
		Lat:     node.Lat,
		Lon:     node.Lon,
	}
}

func writeOsmToPbf(outputOsm *osm.OSM) {
	sigolo.Debug("Convert result to OSM XML")
	outputXml, err := xml.Marshal(outputOsm)
	sigolo.FatalCheck(err)

	osmXmlOutputFile := "features.osm"
	sigolo.Debug("Write result to temp file %s", osmXmlOutputFile)
	err = os.WriteFile(osmXmlOutputFile, outputXml, 0644)
	sigolo.FatalCheck(err)

	sigolo.Debug("Convert written OSM-XML file to OSM-PBF file %s", outputFileName)
	command := exec.Command("osmium", "cat", osmXmlOutputFile, "-o", outputFileName, "--overwrite")
	sigolo.Debug("Call osmium: %s", command.String())
	err = command.Run()
	sigolo.FatalCheck(err)

	sigolo.Debug("Remove temp file %s", osmXmlOutputFile)
	err = os.Remove(osmXmlOutputFile)
	sigolo.FatalCheck(err)

	sigolo.Info("Feature written to %s", outputFileName)
}

func generateVectorTiles() {
	sigolo.Info("Generate tiles from generated OSM-PBF")

	currentWorkingDir, err := os.Getwd()
	sigolo.FatalCheck(err)

	osmPbfFile := currentWorkingDir + "/" + outputFileName
	sigolo.Debug("Build make-tiles command with PBF file %s", osmPbfFile)
	command := exec.Command("bash", "make-tiles.sh", osmPbfFile)
	command.Dir = strings.TrimSuffix(currentWorkingDir, path.Base(currentWorkingDir))

	sigolo.Debug("Call make-tiles.sh script to create tiles of generated features")
	err = command.Run()
	sigolo.FatalCheck(err)
}

func generateLegendHtmlItems(schema Schema) string {
	generatedHtml := ""
	for i, category := range schema.Categories {
		sigolo.Info("Generate HTML for category %d: %s", i, category.Title)

		generatedCategoryHtml := strings.Replace(categoryTemplate, placeholderCategoryTitle, category.Title, 1)
		generatedItemsHtml := ""

		for j, item := range category.Items {
			sigolo.Debug("Category %d / %d: Generate HTML for item '%s'", i, j, item.Description)

			generatedItem := itemTemplate
			generatedItem = strings.ReplaceAll(generatedItem, placeholderCategoryId, textToItemId(category.Title))
			generatedItem = strings.ReplaceAll(generatedItem, placeholderItemId, textToItemId(item.Description))
			generatedItem = strings.ReplaceAll(generatedItem, placeholderItemDesc, item.Description)
			generatedItem = strings.ReplaceAll(generatedItem, placeholderItemLon, strconv.FormatFloat(float64(i)*lonOffset+featureWidth/2.0+item.OffsetLon, 'f', -1, 64))
			generatedItem = strings.ReplaceAll(generatedItem, placeholderItemLat, strconv.FormatFloat(float64(j)*latOffset+featureHeight/2.0+item.OffsetLat, 'f', -1, 64))
			generatedItem = strings.ReplaceAll(generatedItem, placeholderItemZoom, "undefined")

			generatedItemsHtml += generatedItem
		}

		generatedHtml += strings.Replace(generatedCategoryHtml, placeholderItems, generatedItemsHtml, 1)
	}

	// Separate rendering of elevation category since there are no tags or anything. Just a smart map extent needs to be
	// used in order to show some elevation lines with a number
	sigolo.Debug("Generate HTML for elevation category")
	generatedItem := itemTemplate
	generatedItem = strings.ReplaceAll(generatedItem, placeholderCategoryId, "contour")
	generatedItem = strings.ReplaceAll(generatedItem, placeholderItemId, "lines")
	generatedItem = strings.ReplaceAll(generatedItem, placeholderItemDesc, "Contour lines")
	generatedItem = strings.ReplaceAll(generatedItem, placeholderItemLon, "10.9687")
	generatedItem = strings.ReplaceAll(generatedItem, placeholderItemLat, "47.43704")
	generatedItem = strings.ReplaceAll(generatedItem, placeholderItemZoom, "14")

	generatedCategoryHtml := strings.Replace(categoryTemplate, placeholderCategoryTitle, "Elevation", 1)
	generatedHtml += strings.Replace(generatedCategoryHtml, placeholderItems, generatedItem, 1)

	return generatedHtml
}

func createLegendHtmlFile(generatedHtml string, additionalInfoHtml string) {
	sigolo.Debug("Read template file")
	templateFileBytes, err := os.ReadFile("legend-template.html")
	sigolo.FatalCheck(err)

	templateFileContent := string(templateFileBytes)

	sigolo.Debug("Replace placeholder %s from template file with actual content", placeholderItems)
	templateFileContent = strings.Replace(templateFileContent, placeholderItems, generatedHtml, 1)

	sigolo.Debug("Replace placeholder %s from template file with actual content", placeholderAdditionalInfo)
	templateFileContent = strings.ReplaceAll(templateFileContent, placeholderAdditionalInfo, additionalInfoHtml)

	sigolo.Debug("Write legend file")
	err = os.WriteFile("legend.html", []byte(templateFileContent), 0644)
	sigolo.FatalCheck(err)
}

func textToItemId(text string) string {
	newText := ""
	matcher := regexp.MustCompile("\\w")
	whitespaceMatcher := regexp.MustCompile("\\s")

	text = strings.ToLower(text)
	strings.ReplaceAll(text, "ä", "a")
	strings.ReplaceAll(text, "ö", "o")
	strings.ReplaceAll(text, "ü", "u")
	strings.ReplaceAll(text, "ß", "ss")

	for _, c := range []rune(text) {
		character := string(c)

		if matcher.MatchString(character) {
			newText += character
		} else if whitespaceMatcher.MatchString(character) {
			newText += "_"
		}
	}

	return newText
}
