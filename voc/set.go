package voc

type Set map[string][]Object

// Loads all object annotations for all images in a set.
func Load(dir, class, set string) (Set, error) {
	imgs, err := Images(dir, class, set)
	if err != nil {
		return nil, err
	}

	// Load annotations.
	imgset := make(map[string][]Object, len(imgs))
	for _, img := range imgs {
		objs, err := Objects(dir, img)
		if err != nil {
			return nil, err
		}
		imgset[img] = objs
	}
	if class == "*" {
		return imgset, nil
	}

	// Take subset for this class.
	for img, objs := range imgset {
		var subset []Object
		for _, obj := range objs {
			if obj.Class != class {
				continue
			}
			subset = append(subset, obj)
		}
		imgset[img] = subset
	}
	return imgset, nil
}
