package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/hauke96/sigolo"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/paulmach/osm/osmxml"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var nodeIdCounter int64 = 1
var inputNodes = map[osm.NodeID]*osm.Node{}
var inputWays = map[osm.WayID]*osm.Way{}

var cli struct {
	Debug  bool   `help:"Enable debug mode." short:"d"`
	Input  string `help:"The input file. Either .osm or .pbf (OSM-PBF format)." short:"i"`
	Output string `help:"The output file, which must be a .osm.pbf (OSM-PBF format) file." short:"o"`
}

func main() {
	kong.Parse(&cli)

	if cli.Debug {
		sigolo.LogLevel = sigolo.LOG_DEBUG
	}

	if !strings.HasSuffix(cli.Input, ".osm") && !strings.HasSuffix(cli.Input, ".pbf") {
		sigolo.Error("Input file must be an .osm or .pbf file")
		os.Exit(1)
	}
	if !strings.HasSuffix(cli.Output, ".osm.pbf") {
		sigolo.Error("Output file must be an .osm.pbf file")
		os.Exit(1)
	}

	f, err := os.Open(cli.Input)
	sigolo.FatalCheck(err)
	defer f.Close()

	var scanner osm.Scanner
	if strings.HasSuffix(cli.Input, ".osm") {
		scanner = osmxml.New(context.Background(), f)
	} else if strings.HasSuffix(cli.Input, ".pbf") {
		scanner = osmpbf.New(context.Background(), f, 1)
	}
	defer scanner.Close()

	sigolo.Info("Determine centroids")
	outputOsm := osm.OSM{
		Version: "0.6",
	}

	for scanner.Scan() {
		obj := scanner.Object()
		switch osmObj := obj.(type) {
		case *osm.Node:
			inputNodes[osmObj.ID] = osmObj
		case *osm.Way:
			inputWays[osmObj.ID] = osmObj
			handleWay(osmObj, &outputOsm)
		case *osm.Relation:
			handleRelation(osmObj, &outputOsm)
		}
		outputOsm.Append(obj)
	}

	err = scanner.Err()
	sigolo.FatalCheck(err)

	sigolo.Debug("Convert result to OSM XML")
	outputXml, err := xml.Marshal(outputOsm)
	sigolo.FatalCheck(err)

	osmXmlOutputFile := path.Base(strings.TrimSuffix(cli.Output, ".osm.pbf")) + ".tmp.osm"
	sigolo.Debug("Write result to temp file %s", osmXmlOutputFile)
	err = os.WriteFile(osmXmlOutputFile, outputXml, 0644)
	sigolo.FatalCheck(err)

	sigolo.Debug("Convert written OSM-XML file to OSM-PBF file %s", cli.Output)
	command := exec.Command("osmium", "cat", osmXmlOutputFile, "-o", cli.Output, "--overwrite")
	sigolo.Debug("Call osmium: %s", command.String())
	err = command.Run()
	sigolo.FatalCheck(err)

	sigolo.Debug("Remove temp file %s", osmXmlOutputFile)
	err = os.Remove(osmXmlOutputFile)
	sigolo.FatalCheck(err)

	sigolo.Info("Done. Result feature written to %s", cli.Output)
}

func handleRelation(relation *osm.Relation, outputOsm *osm.OSM) {
	sigolo.Debug("Handle relation %d", relation.ID)

	// Collection of all ways with role "outer". Connected rings are later used to determine one centroid.
	var outerRingWays []*osm.Way

	for _, member := range relation.Members {
		if member.Type != osm.TypeWay {
			continue
		}

		// Member ways do not contain their nodes here, only a ref-ID to the actual way
		if memberWay, ok := inputWays[osm.WayID(member.Ref)]; ok {
			tags := memberWay.Tags

			// Sometimes, the actual ways do not have tags -> use the relation tags. If both have a value for
			// same key, the general relation key is overridden by the specific way key.
			if member.Role == "outer" {
				tags = relation.Tags
				tags = append(tags, memberWay.Tags...)
				outerRingWays = append(outerRingWays, memberWay)
			}
		}
	}

	// Merge lines to form potential polygons of which the relation might consist of.
	// Like in this case: https://www.openstreetmap.org/relation/14931281

	// 1.) Collect all ways that form a ring (potentially in wrong order but that's not a problem). This is done by
	// taking the last node of the current way A and find a way B that touches this endpoint. Now B=A and the process
	// repeats until no further touching ways exist.
	var ringWays [][]*osm.Way
	for len(outerRingWays) != 0 {
		// Start with any way, so we take the first one
		currentOuterRingWay := outerRingWays[0]
		currentLastNode := currentOuterRingWay.Nodes.NodeIDs()[len(currentOuterRingWay.Nodes)-1]

		// Remove the way to now process it twice
		outerRingWays = outerRingWays[1:]

		// This array will store all ways that belong to the current ring we'll try to find.
		waysOfCurrentRing := []*osm.Way{currentOuterRingWay}

		// Find other way with "currentLastNode" as first or last node
		for j := 0; j < len(outerRingWays); j++ {
			otherRingWay := outerRingWays[j]

			// Determine if and where the other way touches the current way
			touchedFirstOfOther := currentLastNode == otherRingWay.Nodes.NodeIDs()[0]
			touchedLastOfOther := currentLastNode == otherRingWay.Nodes.NodeIDs()[len(otherRingWay.Nodes)-1]

			if touchedFirstOfOther || touchedLastOfOther {
				// Update the current way variables. This means the "other" way is now our new "current" way.
				currentOuterRingWay = otherRingWay
				waysOfCurrentRing = append(waysOfCurrentRing, currentOuterRingWay)

				// Store the opposite node on the way (touches the first node? Store the last. And vice versa)
				if touchedFirstOfOther {
					currentLastNode = currentOuterRingWay.Nodes.NodeIDs()[len(currentOuterRingWay.Nodes)-1]
				} else if touchedLastOfOther {
					currentLastNode = currentOuterRingWay.Nodes.NodeIDs()[0]
				}

				// Remove j-th way ("other way") and compensate the j++ of the loop due to the removed element
				outerRingWays = append(outerRingWays[:j], outerRingWays[j+1:]...)
				j = -1
			}
		}

		ringWays = append(ringWays, waysOfCurrentRing)
	}

	sigolo.Debug("Found %d rings for potential centroids", len(ringWays))

	// 2.) Create a centroid for each ring
	for _, ring := range ringWays {
		var allNodesOfRing []osm.WayNode
		for _, way := range ring {
			allNodesOfRing = append(allNodesOfRing, way.Nodes...)
		}

		// Do not use handleWays, since we have more like a point cloud here because we do not know (and care)
		// about the order of the ways.
		handleNodeCloud(allNodesOfRing, relation.Tags, outputOsm)
	}
}

// handleWay interprets the given nodes as one way. Its nodes are passed to handleNodeCloud and processed accordingly.
func handleWay(way *osm.Way, outputOsm *osm.OSM) {
	sigolo.Debug("Handle way %d", way.ID)

	nodes := way.Nodes
	// If way is not closed -> ignore, since it's not a polygon and not interesting for the current approach
	if len(nodes) < 3 || nodes[0].ID != nodes[len(nodes)-1].ID {
		// Given nodes do not form a polygon -> nothing to do here
		sigolo.Debug("Way %d is not a polygon. No node will be created.", way.ID)
		return
	}

	handleNodeCloud(nodes, way.Tags, outputOsm)
}

// handleNodeCloud processed the given nodes, determines a potential centroid and creates a new node if needed. This
// node is added to the given OSM output.
func handleNodeCloud(nodes osm.WayNodes, originalTags osm.Tags, outputOsm *osm.OSM) {
	sigolo.Debug("Process node cloud with %d nodes", len(nodes))

	// Convert the nodes of the ways (which have NO GEOMETRY!) to a polygon with geometry.
	var points []orb.Point
	for _, wayNode := range nodes {
		points = append(points, [2]float64{inputNodes[wayNode.ID].Lon, inputNodes[wayNode.ID].Lat})
	}
	polygon := orb.MultiPoint(points)

	centroid := determineCentroidFeatureFromOsmObject(polygon, originalTags)
	if centroid == nil {
		sigolo.Debug("Could not determine centroid geometry. No node will be created.")
		return
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

	if node.Tags.Find("text") != "" {
		sigolo.Debug("Created centroid node with ID %d at lat=%f / lon=%f", nodeIdCounter, node.Lat, node.Lon)
		outputOsm.Append(node)
	}

	sigolo.Debug("Node not added. Centroid node with ID %d at lat=%f / lon=%f has no text", nodeIdCounter, node.Lat, node.Lon)
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

	labelCategory := "natural"
	labelType := getValue(originalTags, "natural")
	if labelType == "" {
		labelCategory = "landuse"
		labelType = getValue(originalTags, "landuse")
	}
	if labelType == "" {
		labelCategory = "protect_class"
		protectClass := getValue(originalTags, "protect_class")
		if protectClass != "" {
			labelType = fmt.Sprintf("protect_class_%s", protectClass)
		}
	}
	if labelType == "" {
		labelCategory = "place"
		labelType = getValue(originalTags, "place")
	}

	// No supported label type -> ignore
	if labelType == "" {
		sigolo.Debug("No centroid point feature created since label-type could not be determined from the following tags: %#v", originalTags)
		return nil
	}

	var centroid *geojson.Feature
	// Create different nodes for different label types
	tags := map[string]interface{}{}
	if labelCategory == "place" {
		tags["place"] = labelType
		tags["name"] = getValue(originalTags, "name")
	} else {
		tags["label"] = "yes"
		tags["type"] = labelType
		tags["text"] = getValue(originalTags, "name")
	}

	centroid = &geojson.Feature{
		Type:       geojson.TypePoint,
		Geometry:   centroidLocation,
		Properties: tags,
	}
	return centroid
}

func getValue(tags map[string]interface{}, key string) interface{} {
	if value, ok := tags[key]; ok {
		return value
	}
	return ""
}
