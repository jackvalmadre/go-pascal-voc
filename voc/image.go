package voc

import (
	"fmt"
	"path"
)

// Loads list of all image names.
//
// Looks in <dir>/ImageSets/Main/<set>.txt.
//
// The set can either be simply "train", "val" or "trainval",
// or it can be "<class>_<set>", for example "horse_val".
func Images(dir, set string) ([]string, error) {
	lines, err := loadLines(path.Join(dir, "ImageSets", "Main", set+".txt"))
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
	return path.Join(dir, "JPEGImages", img+".txt")
}

func intPtrToBool(x *int) *bool {
	if x == nil {
		return nil
	}
	y := *x == 0
	return &y
}
