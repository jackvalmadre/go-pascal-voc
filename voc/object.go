package voc

import (
	"encoding/xml"
	"image"
	"os"
	"path"
)

// A window with a name.
type Object struct {
	Class  string
	Region image.Rectangle
	// Optional flags.
	Difficult *bool
	Occluded  *bool
	Truncated *bool
}

// Loads the objects present in an image.
// An image may contain multiple objects.
//
// Looks in <dir>/Annotations/<image>.xml.
func Objects(dir, img string) ([]Object, error) {
	// Open file.
	name := path.Join(dir, "Annotations", img+".xml")
	fi, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	// Parse from XML.
	var data struct {
		XMLName xml.Name `xml:"annotation"`
		Objects []struct {
			Name   string `xml:"name"`
			BndBox struct {
				XMin int `xml:"xmin"`
				YMin int `xml:"ymin"`
				XMax int `xml:"xmax"`
				YMax int `xml:"ymax"`
			} `xml:"bndbox"`
			Difficult *int `xml:"difficult"`
			Occluded  *int `xml:"occluded"`
			Truncated *int `xml:"truncated"`
		} `xml:"object"`
	}
	if err := xml.NewDecoder(fi).Decode(&data); err != nil {
		return nil, err
	}

	// Construct from XML object.
	objs := make([]Object, len(data.Objects))
	for i, raw := range data.Objects {
		box := raw.BndBox
		obj := Object{
			raw.Name,
			image.Rect(box.XMin, box.YMin, box.XMax, box.YMax),
			intPtrToBool(raw.Difficult),
			intPtrToBool(raw.Occluded),
			intPtrToBool(raw.Truncated),
		}
		objs[i] = obj
	}
	return objs, nil
}
