package legend_graphic

import (
	"encoding/json"
	"github.com/hauke96/sigolo"
	"github.com/paulmach/osm"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"tool/common"
)

const (
	latOffset     = 0.001
	lonOffset     = 0.001
	featureHeight = 0.0001
	featureWidth  = 0.000225

	legendFeaturesFileName = "features.osm.pbf"

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
	categoryTemplate = `			<h3>` + placeholderCategoryTitle + `</h3>
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

func GenerateLegendGraphic(schemaFile string) {
	if !strings.HasSuffix(schemaFile, ".json") {
		sigolo.Error("Input file must be an .json file")
		os.Exit(1)
	}

	schema := readSchemaFile(schemaFile)

	outputOsm := osm.OSM{
		Version: "0.6",
	}

	createFeaturesFromSchema(schema, &outputOsm)

	common.WriteOsmToPbf(legendFeaturesFileName, &outputOsm)

	common.GenerateVectorTiles(legendFeaturesFileName)

	generatedHtml := generateLegendHtmlItems(schema)
	additionalInfoHtml := generateAdditionalInfoHtml(schema)
	
	createLegendHtmlFile(generatedHtml, additionalInfoHtml)
}

func readSchemaFile(schemaFile string) Schema {
	schemaFileBytes, err := os.ReadFile(schemaFile)
	sigolo.FatalCheck(err)

	schema := Schema{}
	err = json.Unmarshal(schemaFileBytes, &schema)
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

	nodeLowerLeft := common.ToNode(originLon+offsetLon, originLat+offsetLat, nil, timestamp, outputOsm)

	nodeUpperLeft := common.ToNode(originLon+offsetLon, originLat+featureHeight+offsetLat, nil, timestamp, outputOsm)

	nodeUpperRight := common.ToNode(originLon+featureWidth+offsetLon, originLat+featureHeight+offsetLat, nil, timestamp, outputOsm)

	nodeLowerRight := common.ToNode(originLon+featureWidth+offsetLon, originLat+offsetLat, nil, timestamp, outputOsm)

	nodes := osm.WayNodes{
		common.ToWayNode(nodeLowerLeft),
		common.ToWayNode(nodeUpperLeft),
		common.ToWayNode(nodeUpperRight),
		common.ToWayNode(nodeLowerRight),
		common.ToWayNode(nodeLowerLeft),
	}

	if featurePerTag {
		for _, tag := range tags {
			way := common.ToWay(nodes, []osm.Tag{tag}, timestamp)
			outputOsm.Append(way)
		}
	} else {
		way := common.ToWay(nodes, tags, timestamp)
		outputOsm.Append(way)
	}
}

func addLine(originLon float64, offsetLon float64, originLat float64, offsetLat float64, tags []osm.Tag, featurePerTag bool, outputOsm *osm.OSM) {
	timestamp, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	sigolo.FatalCheck(err)

	nodeLeft := common.ToNode(originLon+offsetLon, originLat+featureHeight/2+offsetLat, nil, timestamp, outputOsm)

	nodeRight := common.ToNode(originLon+featureWidth+offsetLon, originLat+featureHeight/2+offsetLat, nil, timestamp, outputOsm)

	nodes := osm.WayNodes{
		common.ToWayNode(nodeLeft),
		common.ToWayNode(nodeRight),
	}

	if featurePerTag {
		for _, tag := range tags {
			way := common.ToWay(nodes, []osm.Tag{tag}, timestamp)
			outputOsm.Append(way)
		}
	} else {
		way := common.ToWay(nodes, tags, timestamp)
		outputOsm.Append(way)
	}
}

func addPoint(originLon float64, offsetLon float64, originLat float64, offsetLat float64, tags []osm.Tag, featurePerTag bool, outputOsm *osm.OSM) {
	timestamp, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	sigolo.FatalCheck(err)

	if featurePerTag {
		for _, tag := range tags {
			node := common.ToNode(originLon+featureWidth/2+offsetLon, originLat+featureHeight/2+offsetLat, []osm.Tag{tag}, timestamp, outputOsm)
			outputOsm.Append(&node)
		}
	} else {
		node := common.ToNode(originLon+featureWidth/2+offsetLon, originLat+featureHeight/2+offsetLat, tags, timestamp, outputOsm)
		outputOsm.Append(&node)
	}
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

func generateAdditionalInfoHtml(schema Schema) string {
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

	return additionalInfoHtml
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
