package voc

type Set map[string][]Object

// Loads all object annotations for all images in set.
func Load(dir, set string) (Set, error) {
	imgs, err := Images(dir, set)
	if err != nil {
		return nil, err
	}
	return load(dir, imgs)
}

// Loads all object annotations for all images in a set.
func LoadClass(dir, set, class string) (Set, error) {
	// Load list of images in this class.
	imgs, err := ImagesClass(dir, set, class)
	if err != nil {
		return nil, err
	}

	imgset, err := load(dir, imgs)
	if err != nil {
		return nil, err
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

// Loads annotations for each in a set of images.
func load(dir string, imgs []string) (Set, error) {
	// Load annotations.
	imgset := make(map[string][]Object, len(imgs))
	for _, img := range imgs {
		objs, err := Objects(dir, img)
		if err != nil {
			return nil, err
		}
		imgset[img] = objs
	}
	return imgset, nil
}
