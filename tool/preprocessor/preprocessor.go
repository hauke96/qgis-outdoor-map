package preprocessor

import (
	"context"
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
	"tool/common"
)

var (
	nodeIdCounter      int64 = 0
	inputNodes               = map[osm.NodeID]*osm.Node{}
	inputWays                = map[osm.WayID]*osm.Way{}
	inputRelations           = map[osm.RelationID]*osm.Relation{}
	hikingRouteMapping       = map[osm.WayID][]osm.RelationID{}
)

func PreprocessData(inputFile string, outputFile string) {
	if !strings.HasSuffix(inputFile, ".osm") && !strings.HasSuffix(inputFile, ".pbf") {
		sigolo.Error("Input file must be an .osm or .pbf file")
		os.Exit(1)
	}
	if !strings.HasSuffix(outputFile, ".osm.pbf") {
		sigolo.Error("Output file must be an .osm.pbf file")
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

	outputOsm := osm.OSM{
		Version: "0.6",
	}

	sigolo.Debug("Start processing input data")
	for scanner.Scan() {
		obj := scanner.Object()
		switch osmObj := obj.(type) {
		case *osm.Node:
			inputNodes[osmObj.ID] = osmObj
			handleNode(osmObj)
		case *osm.Way:
			inputWays[osmObj.ID] = osmObj
			handleWay(osmObj, &outputOsm)
		case *osm.Relation:
			inputRelations[osmObj.ID] = osmObj
			handleRelation(osmObj, &outputOsm)
		}
	}

	// Add name of all hiking routes to given ways
	sigolo.Debug("Add hiking route names to ways")
	for wayId, relationIds := range hikingRouteMapping {
		way := inputWays[wayId]

		newHikingRouteName := ""

		for _, relationId := range relationIds {
			relation := inputRelations[relationId]

			routeName := relation.Tags.Find("name")
			routeRef := relation.Tags.Find("ref")
			var combinedRouteName string
			if routeName != "" {
				combinedRouteName = routeName
				if routeRef != "" {
					combinedRouteName += " (" + routeRef + ")"
				}
			} else if routeRef != "" {
				combinedRouteName = routeRef
			} else {
				// Neither name nor ref on route -> cannot set any name on way
				continue
			}

			if newHikingRouteName == "" {
				newHikingRouteName = combinedRouteName
			} else {
				newHikingRouteName = newHikingRouteName + ", " + combinedRouteName
			}
		}

		newHikingRouteNameTag := &osm.Tag{
			Key:   "hiking_route_names",
			Value: newHikingRouteName,
		}
		newHikingRouteTag := &osm.Tag{
			Key:   "hiking_route",
			Value: "yes",
		}
		way.Tags = append(way.Tags, *newHikingRouteNameTag, *newHikingRouteTag)
	}

	sigolo.Debug("Write %d nodes to output", len(inputNodes))
	for _, node := range inputNodes {
		outputOsm.Append(node)
	}
	sigolo.Debug("Write %d ways to output", len(inputWays))
	for _, way := range inputWays {
		outputOsm.Append(way)
	}
	sigolo.Debug("Write %d relations to output", len(inputRelations))
	for _, relation := range inputRelations {
		outputOsm.Append(relation)
	}

	err = scanner.Err()
	sigolo.FatalCheck(err)

	sigolo.Debug("Write OSM")
	common.WriteOsmToPbf(outputFile, &outputOsm)
}

func handleRelation(relation *osm.Relation, outputOsm *osm.OSM) {
	if relation.Tags.Find("type") != "route" || relation.Tags.Find("route") != "hiking" {
		return
	}

	for _, member := range relation.Members {
		if member.Type != osm.TypeWay {
			continue
		}

		// Member ways do not contain their nodes here, only a ref-ID to the actual way
		if memberWay, ok := inputWays[osm.WayID(member.Ref)]; ok {
			hikingRouteMapping[memberWay.ID] = append(hikingRouteMapping[memberWay.ID], relation.ID)

			//// Sometimes, the actual ways do not have tags -> use the relation tags. If both have a value for
			//// same key, the general relation key is overridden by the specific way key.
			//tags := relation.Tags
			//tags = append(tags, memberWay.Tags...)
			//segments = append(segments, memberWay)
		}
	}

	//// Merge lines to form potential polygons of which the relation might consist of.
	//// Like in this case: https://www.openstreetmap.org/relation/14931281
	//
	//// 1.) Collect all ways that form a ring (potentially in wrong order but that's not a problem). This is done by
	//// taking the last node of the current way A and find a way B that touches this endpoint. Now B=A and the process
	//// repeats until no further touching ways exist.
	//var ringWays [][]*osm.Way
	//for len(segments) != 0 {
	//	// Start with any way, so we take the first one
	//	currentOuterRingWay := segments[0]
	//	currentLastNode := currentOuterRingWay.Nodes.NodeIDs()[len(currentOuterRingWay.Nodes)-1]
	//
	//	// Remove the way to now process it twice
	//	segments = segments[1:]
	//
	//	// This array will store all ways that belong to the current ring we'll try to find.
	//	waysOfCurrentRing := []*osm.Way{currentOuterRingWay}
	//
	//	// Find other way with "currentLastNode" as first or last node
	//	for j := 0; j < len(segments); j++ {
	//		otherRingWay := segments[j]
	//
	//		// Determine if and where the other way touches the current way
	//		touchedFirstOfOther := currentLastNode == otherRingWay.Nodes.NodeIDs()[0]
	//		touchedLastOfOther := currentLastNode == otherRingWay.Nodes.NodeIDs()[len(otherRingWay.Nodes)-1]
	//
	//		if touchedFirstOfOther || touchedLastOfOther {
	//			// Update the current way variables. This means the "other" way is now our new "current" way.
	//			currentOuterRingWay = otherRingWay
	//			waysOfCurrentRing = append(waysOfCurrentRing, currentOuterRingWay)
	//
	//			// Store the opposite node on the way (touches the first node? Store the last. And vice versa)
	//			if touchedFirstOfOther {
	//				currentLastNode = currentOuterRingWay.Nodes.NodeIDs()[len(currentOuterRingWay.Nodes)-1]
	//			} else if touchedLastOfOther {
	//				currentLastNode = currentOuterRingWay.Nodes.NodeIDs()[0]
	//			}
	//
	//			// Remove j-th way ("other way") and compensate the j++ of the loop due to the removed element
	//			segments = append(segments[:j], segments[j+1:]...)
	//			j = -1
	//		}
	//	}
	//
	//	ringWays = append(ringWays, waysOfCurrentRing)
	//}
	//
	//sigolo.Debug("Found %d rings for potential centroids", len(ringWays))
	//
	//// 2.) Create a centroid for each ring
	//for _, ring := range ringWays {
	//	var allNodesOfRing []osm.WayNode
	//	for _, way := range ring {
	//		allNodesOfRing = append(allNodesOfRing, way.Nodes...)
	//	}
	//
	//	// Do not use handleWays, since we have more like a point cloud here because we do not know (and care)
	//	// about the order of the ways.
	//	handleNodeCloud(allNodesOfRing, relation.Tags, outputOsm)
	//}
}

func handleNode(node *osm.Node) {
	if int64(node.ID) > nodeIdCounter {
		nodeIdCounter = int64(node.ID)
	}
	//if node.Tags.Find("natural") == "peak" {
	//	name := node.Tags.Find("name")
	//	ele := node.Tags.Find("ele")
	//
	//	newName := ""
	//	if name != "" {
	//		newName = name
	//	}
	//	if ele != "" {
	//		ele = ele + "m"
	//		if newName == "" {
	//			newName = ele
	//		} else {
	//			newName += " " + ele
	//		}
	//	}
	//
	//	if newName != "" {
	//		nameTag := node.Tags.FindTag("name")
	//		if nameTag != nil {
	//			nameTag.Value = newName
	//		} else {
	//			node.Tags = append(node.Tags, osm.Tag{
	//				Key:   "name",
	//				Value: newName,
	//			})
	//		}
	//	}
	//}
}

// handleWay interprets the given nodes as one way. Its nodes are passed to handleNodeCloud and processed accordingly.
func handleWay(way *osm.Way, outputOsm *osm.OSM) {
	// Handling of some special cases where certain ways should be converted into point features
	if way.Tags.Find("waterway") == "waterfall" || way.Tags.Find("tourism") == "camp_site" {
		centroid, _ := determineCentroidLocation(way.LineString())
		AddNode(centroid.Lon(), centroid.Lat(), way.Tags)
	}
}

func AddNode(originLon float64, originLat float64, tags []osm.Tag) {
	nodeIdCounter++
	node := osm.Node{
		Version:   1,
		ID:        osm.NodeID(nodeIdCounter),
		Timestamp: common.GetTimestamp(),
		Tags:      tags,
		Lon:       originLon,
		Lat:       originLat,
	}
	inputNodes[node.ID] = &node
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

	node := &osm.Node{
		Version:   1,
		ID:        osm.NodeID(nodeIdCounter),
		Timestamp: common.GetTimestamp(),
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
	centroidLocation, _ := determineCentroidLocation(geometry)

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

func determineCentroidLocation(geometry orb.Geometry) (orb.Point, float64) {
	return planar.CentroidArea(geometry)
}

func getValue(tags map[string]interface{}, key string) interface{} {
	if value, ok := tags[key]; ok {
		return value
	}
	return ""
}
