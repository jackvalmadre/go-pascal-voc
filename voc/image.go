package voc

import (
	"fmt"
	"path"
)

// Loads list of all image names.
//
// Looks in <dir>/ImageSets/Main/<set>.txt.
//
// The class can be either a class name or "*".
// The set can be either "train", "val" or "trainval".
func Images(dir, class, set string) ([]string, error) {
	var name string
	if class == "*" {
		name = set
	} else {
		name = class + "_" + set
	}

	lines, err := loadLines(path.Join(dir, "ImageSets", "Main", name+".txt"))
	if err != nil {
		return nil, err
	}

	// Extract "name label" from every line.
	for i, line := range lines {
		var (
			name  string
			label int
		)
		if _, err := fmt.Sscanf(line, "%s %d", &name, &label); err != nil {
			return nil, err
		}
		lines[i] = name
	}
	return lines, nil
}

// Returns <dir>/JPEGImages/<img>.jpg.
func ImageFile(dir, img string) string {
	return path.Join(dir, "JPEGImages", img+".jpg")
}

func intPtrToBool(x *int) *bool {
	if x == nil {
		return nil
	}
	y := *x != 0
	return &y
}
