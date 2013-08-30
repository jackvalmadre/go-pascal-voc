package voc

type Set map[string][]Object

// Loads all object annotations for all images in a set.
func Load(dir, setname string) (Set, error) {
	imgs, err := Images(dir, setname)
	if err != nil {
		return nil, err
	}

	set := make(map[string][]Object, len(imgs))
	for _, img := range imgs {
		objs, err := Objects(dir, img)
		if err != nil {
			return nil, err
		}
		set[img] = objs
	}
	return set, nil
}
