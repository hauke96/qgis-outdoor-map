package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/hauke96/sigolo"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/paulmach/osm/osmxml"
	"os"
	"strings"
	"time"
)

var nodeIdCounter int64 = 1
var inputNodes = map[osm.NodeID]*osm.Node{}
var inputWays = map[osm.WayID]*osm.Way{}

func main() {
	if len(os.Args) != 3 {
		sigolo.Error("Expect 2 args, found %d", len(os.Args))
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputFile := os.Args[2]

	if !strings.HasSuffix(inputFile, ".osm") && !strings.HasSuffix(inputFile, ".pbf") {
		sigolo.Error("Input file must be an .osm or .pbf file")
		os.Exit(1)
	}
	if !strings.HasSuffix(outputFile, ".osm") {
		sigolo.Error("Output file must be an .osm file")
		os.Exit(1)
	}

	f, err := os.Open(inputFile)
	sigolo.FatalCheck(err)
	defer f.Close()

	var scanner osm.Scanner
	if strings.HasSuffix(inputFile, ".osm") {
		scanner = osmxml.New(context.Background(), f)
	} else if strings.HasSuffix(inputFile, ".pbf") {
		scanner = osmpbf.New(context.Background(), f, 1)
	}
	defer scanner.Close()

	sigolo.Info("Determine centroids")
	outputOsm := osm.OSM{
		Version: "0.6",
	}

	for scanner.Scan() {
		switch obj := scanner.Object().(type) {
		case *osm.Node:
			inputNodes[obj.ID] = obj
		case *osm.Way:
			inputWays[obj.ID] = obj
			centroidNode := handleWay(obj.Nodes, obj.Tags)
			if centroidNode == nil {
				continue
			}
			outputOsm.Append(centroidNode)
		case *osm.Relation:
			for _, member := range obj.Members {
				if member.Type != osm.TypeWay {
					continue
				}

				// TODO Merge lines to form potential polygons the relation might consist of. Like in this case: https://www.openstreetmap.org/relation/14931281

				// Member ways do not contain their nodes here, only a ref-ID to the actual way
				if memberWay, ok := inputWays[osm.WayID(member.Ref)]; ok {
					tags := memberWay.Tags

					// Sometimes, the actual ways do not have tags -> use the relation tags. If both have a value for
					// same key, the general relation key is overridden by the specific way key.
					if member.Role == "outer" {
						tags = obj.Tags
						tags = append(tags, memberWay.Tags...)
					}

					centroidNode := handleWay(memberWay.Nodes, tags)
					if centroidNode == nil {
						continue
					}
					outputOsm.Append(centroidNode)
				}
			}
		}
	}

	err = scanner.Err()
	sigolo.FatalCheck(err)

	sigolo.Info("Convert result to OSM XML")
	outputXml, err := xml.Marshal(outputOsm)
	sigolo.FatalCheck(err)

	sigolo.Info("Write result to %s", outputFile)
	err = os.WriteFile(outputFile, outputXml, 0644)
	sigolo.FatalCheck(err)

	sigolo.Info("Done")
}

func handleWay(nodes osm.WayNodes, originalTags osm.Tags) *osm.Node {
	// If way is not closed -> ignore, since it's not a polygon and not interesting for the current approach
	if len(nodes) < 3 || nodes[0].ID != nodes[len(nodes)-1].ID {
		return nil
	}
	// first node == last node -> Polygon

	// Convert the nodes of the ways (which have NO GEOMETRY!) to a polygon with geometry.
	var points []orb.Point
	for _, wayNode := range nodes {
		points = append(points, [2]float64{inputNodes[wayNode.ID].Lon, inputNodes[wayNode.ID].Lat})
	}
	polygon := orb.Polygon{points}

	centroid := determineCentroidFeatureFromOsmObject(polygon, originalTags)
	if centroid == nil {
		return nil
	}

	var tags []osm.Tag
	for k, v := range centroid.Properties {
		tags = append(tags, osm.Tag{Key: k, Value: v.(string)})
	}

	// Osmium only supports this format, so we basically make the time less accurate so that no millis are serialized
	timestamp, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	sigolo.FatalCheck(err)

	node := &osm.Node{
		Version:   1,
		ID:        osm.NodeID(nodeIdCounter),
		Timestamp: timestamp,
		Tags:      tags,
		Lon:       centroid.Geometry.(orb.Point).Lon(),
		Lat:       centroid.Geometry.(orb.Point).Lat(),
	}
	nodeIdCounter++

	return node
}

func determineCentroidFeatureFromOsmObject(geometry orb.Geometry, tags osm.Tags) *geojson.Feature {
	properties := tagsToPropertyMap(tags)
	return determineCentroidFeature(geometry, properties)
}

func tagsToPropertyMap(tags osm.Tags) map[string]interface{} {
	properties := map[string]interface{}{}
	for _, tag := range tags {
		properties[tag.Key] = tag.Value
	}
	return properties
}

func determineCentroidFeature(geometry orb.Geometry, originalTags map[string]interface{}) *geojson.Feature {
	centroidLocation, _ := planar.CentroidArea(geometry)

	labelType := getValue(originalTags, "natural")
	if labelType == "" {
		labelType = getValue(originalTags, "landuse")
	}
	if labelType == "" {
		protectClass := getValue(originalTags, "protect_class")
		if protectClass != "" {
			labelType = fmt.Sprintf("protect_class_%s", protectClass)
		}
	}

	var centroid *geojson.Feature
	// No supported label type -> ignore
	if labelType != "" {
		tags := map[string]interface{}{
			"label": "yes",
			"type":  labelType,
			"text":  getValue(originalTags, "name"),
		}

		centroid = &geojson.Feature{
			Type:       geojson.TypePoint,
			Geometry:   centroidLocation,
			Properties: tags,
		}
	}
	return centroid
}

func getValue(tags map[string]interface{}, key string) interface{} {
	if value, ok := tags[key]; ok {
		return value
	}
	return ""
}
