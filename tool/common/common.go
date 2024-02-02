package common

import (
	"encoding/xml"
	"github.com/hauke96/sigolo"
	"github.com/paulmach/osm"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

var (
	osmObjIdCounter int64 = 1
)

func AddNode(originLon float64, originLat float64, tags []osm.Tag, timestamp time.Time, outputOsm *osm.OSM) osm.Node {
	node := osm.Node{
		Version:   1,
		ID:        osm.NodeID(osmObjIdCounter),
		Timestamp: timestamp,
		Tags:      tags,
		Lon:       originLon,
		Lat:       originLat,
	}
	osmObjIdCounter++
	outputOsm.Append(&node)
	return node
}

func AddWay(nodes osm.WayNodes, tags osm.Tags, timestamp time.Time, outputOsm *osm.OSM) *osm.Way {
	way := &osm.Way{
		ID:        osm.WayID(osmObjIdCounter),
		Version:   1,
		Timestamp: timestamp,
		Nodes:     nodes,
		Tags:      tags,
	}
	osmObjIdCounter++
	outputOsm.Append(way)
	return way
}

func ToWayNode(node osm.Node) osm.WayNode {
	return osm.WayNode{
		ID:      node.ID,
		Version: node.Version,
		Lat:     node.Lat,
		Lon:     node.Lon,
	}
}

func WriteOsmToPbf(outputFileName string, outputOsm *osm.OSM) {
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

func GenerateVectorTiles(pbfFileName string) {
	sigolo.Info("Generate tiles from generated OSM-PBF")

	currentWorkingDir, err := os.Getwd()
	sigolo.FatalCheck(err)

	osmPbfFile := currentWorkingDir + "/" + pbfFileName
	sigolo.Debug("Build make-tiles command with PBF file %s", osmPbfFile)
	command := exec.Command("bash", "make-tiles.sh", osmPbfFile)
	command.Dir = strings.TrimSuffix(currentWorkingDir, path.Base(currentWorkingDir))

	sigolo.Debug("Call make-tiles.sh script to create tiles of generated features")
	err = command.Run()
	sigolo.FatalCheck(err)
}

func GetTimestamp() time.Time {
	// Osmium only supports this format, so we basically make the time less accurate so that no millis are serialized
	timestamp, err := time.Parse(time.RFC3339, time.Now().UTC().Format(time.RFC3339))
	sigolo.FatalCheck(err)
	return timestamp
}
