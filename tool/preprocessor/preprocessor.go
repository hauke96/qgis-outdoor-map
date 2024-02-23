package preprocessor

import (
	"context"
	"github.com/hauke96/sigolo"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
	"github.com/paulmach/osm/osmxml"
	"os"
	"strings"
	"tool/common"
)

var (
	// Keeps track of the OSM IDs to not use them twice. While reading OSM objects, this counter is set to the highest
	// ID found in the input data. It is then later used as ID counter for new object. This ensures that no IDs are
	// used twice.
	osmObjIdCounter int64 = 0
	// A counter for temporary new objects. This is always a negative number, which marks new objects. Later, this
	// temporary ID is turned into an actual ID using the osmObjIdCounter variable.
	tmpOsmObjIdCounter int64 = -1
	inputNodes               = map[osm.NodeID]*osm.Node{}
	inputWays                = map[osm.WayID]*osm.Way{}
	inputRelations           = map[osm.RelationID]*osm.Relation{}
	wayRelationMapping       = map[osm.WayID][]osm.RelationID{}
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
			handleWay(osmObj)
		case *osm.Relation:
			inputRelations[osmObj.ID] = osmObj
			handleRelation(osmObj)
		}
	}

	sigolo.Debug("Add hiking route names to ways")
	addHikingRouteNamesToWays()

	sigolo.Debug("Write %d nodes to output", len(inputNodes))
	for _, node := range inputNodes {
		outputOsm.Append(node)
	}
	sigolo.Debug("Write %d ways to output", len(inputWays))
	for _, way := range inputWays {
		// Some ways are new any not have an ID yet, which is represented as a negative temporary ID.
		if way.ID < 0 {
			osmObjIdCounter++
			way.ID = osm.WayID(osmObjIdCounter)
		}
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

func addHikingRouteNamesToWays() {
	for wayId, relationIds := range wayRelationMapping {
		way := inputWays[wayId]

		newHikingRouteName := ""
		isHikingRoute := false

		for _, relationId := range relationIds {
			relation := inputRelations[relationId]

			if relation.Tags.Find("type") != "route" || relation.Tags.Find("route") != "hiking" {
				// Not a hiking route -> Ignore
				continue
			}
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

			isHikingRoute = true
		}

		if isHikingRoute {
			newHikingRouteNameTag := osm.Tag{
				Key:   "hiking_route_names",
				Value: newHikingRouteName,
			}
			newHikingRouteTag := osm.Tag{
				Key:   "hiking_route",
				Value: "yes",
			}
			way.Tags = append(way.Tags, newHikingRouteNameTag, newHikingRouteTag)
		}
	}
}

func handleNode(node *osm.Node) {
	updateOsmIdCounter(int64(node.ID))
}

// handleWay might add new nodes or tags to the given way to handle them easier in styling. It returns true of the way
// should be kept and false if the line should not be part of the output.
func handleWay(way *osm.Way) {
	updateOsmIdCounter(int64(way.ID))

	// Handling of some special cases where certain ways should be converted into point features
	if way.Tags.Find("ford") == "yes" ||
		way.Tags.HasTag("shop") ||
		way.Tags.HasTag("amenity") ||
		way.Tags.HasTag("historic") ||
		way.Tags.Find("tourism") == "camp_site" ||
		way.Tags.Find("tourism") == "wilderness_hut" ||
		way.Tags.Find("waterway") == "waterfall" {
		centroid, _ := getCentroidOfWay(way)
		addNode(centroid.Lon(), centroid.Lat(), way.Tags)
	}

	// Add highway tags for not accessible ways
	isUnderConstruction := way.Tags.HasTag("construction")
	hasNoRealHighwayTag := !way.Tags.HasTag("highway") || way.Tags.Find("highway") == "construction"
	isNotAccessible := way.Tags.Find("access") == "no" ||
		way.Tags.Find("access") == "private" ||
		way.Tags.Find("foot") == "no" ||
		way.Tags.Find("foot") == "private"

	if isUnderConstruction && hasNoRealHighwayTag {
		// Set highway tag for construction roads without proper one
		way.Tags = setTag(way.Tags, "highway", way.Tags.Find("construction"))
	}

	if isUnderConstruction || isNotAccessible {
		// Override access tag for simplicity
		way.Tags = setTag(way.Tags, "access", "no")
	}

	// Remove "_link" part of highway tags
	if strings.HasSuffix(way.Tags.Find("highway"), "_link") {
		way.Tags = setTag(way.Tags, "highway", strings.TrimSuffix(way.Tags.Find("highway"), "_link"))
	}

	// Add polygonal barriers as linestring to simplify styling
	if barrierTagValue := way.Tags.Find("barrier"); barrierTagValue != "" {
		openBarrierWay := &osm.Way{
			ID:        osm.WayID(tmpOsmObjIdCounter),
			Version:   0,
			Timestamp: common.GetTimestamp(),
			Nodes:     way.Nodes,
			Tags: osm.Tags{
				osm.Tag{
					Key:   "barrier",
					Value: barrierTagValue,
				},
			},
		}
		tmpOsmObjIdCounter--
		inputWays[openBarrierWay.ID] = openBarrierWay
	}
}

func handleRelation(relation *osm.Relation) {
	updateOsmIdCounter(int64(relation.ID))

	// Store each way that is part of a hiking-route separately to tag them later.
	for _, member := range relation.Members {
		if member.Type != osm.TypeWay {
			continue
		}

		// Member ways do not contain their nodes here, only a ref-ID to the actual way
		if memberWay, ok := inputWays[osm.WayID(member.Ref)]; ok {
			wayRelationMapping[memberWay.ID] = append(wayRelationMapping[memberWay.ID], relation.ID)
		}
	}
}

// addNode to the input list of nodes
func addNode(originLon float64, originLat float64, tags []osm.Tag) {
	osmObjIdCounter++
	node := osm.Node{
		Version:   1,
		ID:        osm.NodeID(osmObjIdCounter),
		Timestamp: common.GetTimestamp(),
		Tags:      tags,
		Lon:       originLon,
		Lat:       originLat,
	}
	inputNodes[node.ID] = &node
}

func getCentroidOfWay(w *osm.Way) (orb.Point, float64) {
	geometry := make(orb.LineString, 0, len(w.Nodes))
	for _, n := range w.Nodes {
		nodeWithGeometry := inputNodes[n.ID]
		geometry = append(geometry, nodeWithGeometry.Point())
	}
	return planar.CentroidArea(geometry)
}

func setTag(tags osm.Tags, key string, value string) osm.Tags {
	newTags := osm.Tags{
		osm.Tag{
			Key:   key,
			Value: value,
		},
	}

	for _, t := range tags {
		if t.Key != key {
			newTags = append(newTags, t)
		}
	}

	return newTags
}

func updateOsmIdCounter(osmObjId int64) {
	if osmObjId > osmObjIdCounter {
		osmObjIdCounter = osmObjId
	}
}
