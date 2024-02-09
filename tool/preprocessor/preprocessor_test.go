package preprocessor

import (
	"github.com/paulmach/osm"
	"testing"
)

func TestWay_removeLink(t *testing.T) {
	way := osm.Way{
		ID: 123,
		Tags: []osm.Tag{
			{Key: "foo", Value: "bar"},
			{Key: "highway", Value: "primary_link"},
		},
	}

	handleWay(&way)

	if len(way.Tags) != 2 {
		t.Errorf("Tag count must be 2 but was %d", len(way.Tags))
	}

	hasCorrectHighwayTag := false
	for _, tag := range way.Tags {
		if tag.Key == "highway" && tag.Value == "primary" {
			hasCorrectHighwayTag = true
			break
		}
	}
	if !hasCorrectHighwayTag {
		t.Errorf("No correct highway tag found: %#v", way.Tags)
	}
}

func TestWay_removeLink_underConstruction(t *testing.T) {
	way := osm.Way{
		ID: 123,
		Tags: []osm.Tag{
			{Key: "foo", Value: "bar"},
			{Key: "highway", Value: "construction"},
			{Key: "construction", Value: "primary_link"},
		},
	}

	handleWay(&way)

	if len(way.Tags) != 4 {
		t.Errorf("Tag count must be 3 but was %d", len(way.Tags))
	}

	hasCorrectHighwayTag := false
	hasCorrectAccessTag := false
	for _, tag := range way.Tags {
		if tag.Key == "highway" && tag.Value == "primary" {
			hasCorrectHighwayTag = true
		}
		if tag.Key == "access" && tag.Value == "no" {
			hasCorrectAccessTag = true
		}
	}
	if !hasCorrectHighwayTag {
		t.Errorf("No correct highway tag found: %#v", way.Tags)
	}
	if !hasCorrectAccessTag {
		t.Errorf("No correct access tag found: %#v", way.Tags)
	}
}
