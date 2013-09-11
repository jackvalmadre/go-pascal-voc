package voc

import (
	"fmt"
	"path"
	"regexp"
)

// Loads list of all image names.
//
// Looks in <dir>/ImageSets/Main/<set>.txt.
// The set can be either "train", "val" or "trainval".
func Images(dir, set string) ([]string, error) {
	lines, err := loadLines(path.Join(dir, "ImageSets", "Main", set+".txt"))
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`^\d{4}_\d{6}\b`)
	// Extract "name label" from every line.
	for i, line := range lines {
		name := re.FindString(line)
		if name == "" {
			return nil, fmt.Errorf("could not match: %v", line)
		}
		lines[i] = name
	}
	return lines, nil
}

// Loads list of all image names.
//
// Looks in <dir>/ImageSets/Main/<class>_<set>.txt.
func ImagesClass(dir, set, class string) ([]string, error) {
	return Images(dir, class+"_"+set)
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
