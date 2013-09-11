package main

import (
	"github.com/jackvalmadre/go-pascal-voc/voc"
	"github.com/nfnt/resize"

	"image"
	_ "image/jpeg"
	"image/png"

	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage of %s:\n", os.Args[0])
	fmt.Fprintln(os.Stderr, os.Args[0], "[flags] dir classes sets")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, `dir -- Pascal VOC root directory (usually e.g. VOC2012)`)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, `classes -- Comma-separated list of classes`)
	fmt.Fprintln(os.Stderr, `  e.g. "person,horse,tvmonitor"`)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, `sets -- Comma-separated list of sets`)
	fmt.Fprintln(os.Stderr, `  e.g. "train,val"`)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, `Makes a directory ./<class>-<set>/ containing images.`)
	fmt.Fprintln(os.Stderr, `Creates a file ./<class>-<set>.txt with a list of these images.`)
	fmt.Fprintln(os.Stderr)
}

func main() {
	var (
		numPix  int
		exclude voc.Tags
	)
	flag.IntVar(&numPix , "pixels", 100*100, "Rough number of pixels in window")
	flag.BoolVar(&exclude.Difficult, "exclude-difficult", false, "Exclude objects marked as difficult")
	flag.BoolVar(&exclude.Occluded, "exclude-occluded", false, "Exclude objects marked as occluded")
	flag.BoolVar(&exclude.Truncated, "exclude-truncated", false, "Exclude objects marked as truncated")

	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 3 {
		flag.Usage()
		os.Exit(1)
	}
	var (
		dir        = flag.Arg(0)
		classesStr = flag.Arg(1)
		setsStr    = flag.Arg(2)
	)

	// Extract classes and sets.
	classes := strings.Split(classesStr, ",")
	sets := strings.Split(setsStr, ",")

	for _, class := range classes {
		for _, set := range sets {
			log.Printf("sample: class %s, set %s", class, set)
			sample(dir, class, set, numPix, exclude)
		}
	}
}

func sample(vocDir, class, set string, numPix int, exclude voc.Tags) {
	outDir := class + "-" + set
	// Create empty directory to write images to.
	if err := os.RemoveAll(outDir); err != nil {
		log.Fatalln("could not clear image dir:", err)
	}
	if err := os.Mkdir(outDir, 0755); err != nil {
		log.Fatalln("could not create image dir:", err)
	}

	// Load images containing instances of class and corresponding annotations.
	log.Println("load annotations")
	imgset, err := voc.LoadClass(vocDir, set, class)
	if err != nil {
		log.Fatalln("could not load annotations:", err)
	}

	// Trim difficult, occluded or truncated examples.
	var n int
	imgset, n = removeDifficult(imgset, exclude)
	log.Println("removed", n, "windows: difficult")

	// Get image sizes.
	log.Println("get size of images")
	sizes := make(map[string]image.Point, len(imgset))
	for name := range imgset {
		conf, err := loadImageConfig(voc.ImageFile(vocDir, name))
		if err != nil {
			log.Fatalln("could not load image:", err)
		}
		sizes[name] = image.Pt(conf.Width, conf.Height)
	}

	// Get list of aspect ratios.
	var aspects []float64
	for _, objs := range imgset {
		for _, obj := range objs {
			w, h := obj.Region.Dx(), obj.Region.Dy()
			aspects = append(aspects, float64(w)/float64(h))
		}
	}
	// Get optimal aspect ratio.
	aspect := OptimalAspect(aspects)
	log.Printf("optimal aspect ratio: %g", aspect)

	// Compute reference width and height.
	// w h = A, w = a h, A = h^2 a, h = sqrt(A / a)
	// w^2 / a = A, w = sqrt(A * a)
	width := round(math.Sqrt(float64(numPix ) * aspect))
	height := round(math.Sqrt(float64(numPix ) / aspect))
	log.Printf("base size: %dx%d", width, height)

	// Clone and resize the bounding boxes.
	imgset = resizeImageSet(imgset, aspect)
	// Remove any boxes which don't fit.
	imgset, n = removeNotInside(imgset, sizes)
	log.Println("removed", n, "windows: outside image")
	// Remove boxes which are significantly smaller than the window size.
	imgset, n = removeSmall(imgset, image.Pt(width/2, height/2))
	log.Println("removed", n, "windows: too small")

	log.Printf("sample and save images")
	var imgFiles []string
	// Extract windows from images.
	for imgName, objs := range imgset {
		// Load image from file.
		img, err := loadImage(voc.ImageFile(vocDir, imgName))
		if err != nil {
			log.Fatalln("could not load image:", err)
		}

		var (
			sub subImager
			ok  bool
		)
		if sub, ok = img.(subImager); !ok {
			log.Fatalf("could not call SubImage(): %T", img)
		}
		for i, obj := range objs {
			// Extract rectangle.
			subImg := sub.SubImage(obj.Region)
			// Resize.
			subImg = resize.Resize(uint(width), uint(height), subImg, resize.Bilinear)
			imgFile := fmt.Sprintf("%s_%d.png", imgName, i)
			if err := saveImage(subImg, path.Join(outDir, imgFile)); err != nil {
				log.Println("could not save image:", err)
			}
			imgFiles = append(imgFiles, imgFile)
		}
	}

	// Save list of images files.
	listFile := class + "-" + set + ".txt"
	if err := saveLines(imgFiles, listFile); err != nil {
		log.Fatalln("could not save list of images:", err)
	}
}

type subImager interface {
	SubImage(image.Rectangle) image.Image
}

func loadImageConfig(filename string) (image.Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return image.Config{}, err
	}
	defer file.Close()

	conf, _, err := image.DecodeConfig(file)
	if err != nil {
		return image.Config{}, err
	}
	return conf, nil
}

func loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func saveImage(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

// Map.
func resizeImageSet(set map[string][]voc.Object, aspect float64) map[string][]voc.Object {
	dstSet := make(map[string][]voc.Object, len(set))
	for name, objs := range set {
		var dstObjs []voc.Object
		for _, obj := range objs {
			// obj is a clone of objs[i].
			obj.Region = resizeRect(obj.Region, aspect)
			dstObjs = append(dstObjs, obj)
		}
		dstSet[name] = dstObjs
	}
	return dstSet
}

func resizeRect(rect image.Rectangle, aspect float64) image.Rectangle {
	w, h := float64(rect.Dx()), float64(rect.Dy())
	if w < h*aspect {
		// Increase width to preserve aspect ratio.
		w = h * aspect
	} else {
		// Increase height to preserve aspect ratio.
		h = w / aspect
	}
	// Get center point.
	x := float64(rect.Min.X+rect.Max.X) / 2
	y := float64(rect.Min.Y+rect.Max.Y) / 2
	// Subtract, add half of width, height.
	// Upper bound is not inclusive.
	xmin, xmax := round(x-w/2), round(x+w/2)+1
	ymin, ymax := round(y-h/2), round(y+h/2)+1
	return image.Rect(xmin, ymin, xmax, ymax)
}

func round(x float64) int {
	return int(math.Floor(x + 0.5))
}

// Filter.
func removeNotInside(set map[string][]voc.Object, sizes map[string]image.Point) (map[string][]voc.Object, int) {
	dstSet := make(map[string][]voc.Object, len(set))
	var removed int
	for name, objs := range set {
		img := image.Rectangle{image.ZP, sizes[name]}
		var dstObjs []voc.Object
		for _, obj := range objs {
			if !obj.Region.In(img) {
				// Remove the object if it's not inside.
				removed++
				continue
			}
			dstObjs = append(dstObjs, obj)
		}
		// Remove the image if it no longer has any objects.
		if len(dstObjs) > 0 {
			dstSet[name] = dstObjs
		}
	}
	return dstSet, removed
}

// Filter.
func removeSmall(set map[string][]voc.Object, size image.Point) (map[string][]voc.Object, int) {
	dstSet := make(map[string][]voc.Object, len(set))
	var removed int
	for name, objs := range set {
		var dstObjs []voc.Object
		for _, obj := range objs {
			if obj.Region.Dx() < size.X || obj.Region.Dy() < size.Y {
				removed++
				continue
			}
			dstObjs = append(dstObjs, obj)
		}
		// Remove the image if it no longer has any objects.
		if len(dstObjs) > 0 {
			dstSet[name] = dstObjs
		}
	}
	return dstSet, removed
}

// Filter.
func removeDifficult(set map[string][]voc.Object, exclude voc.Tags) (map[string][]voc.Object, int) {
	dstSet := make(map[string][]voc.Object, len(set))
	var removed int
	for name, objs := range set {
		var dstObjs []voc.Object
		for _, obj := range objs {
			if exclude.Occluded && obj.Occluded != nil && *obj.Occluded {
				removed++
				continue
			}
			if exclude.Truncated && obj.Truncated != nil && *obj.Truncated {
				removed++
				continue
			}
			if exclude.Difficult && obj.Difficult != nil && *obj.Difficult {
				removed++
				continue
			}
			dstObjs = append(dstObjs, obj)
		}
		// Remove the image if it no longer has any objects.
		if len(dstObjs) > 0 {
			dstSet[name] = dstObjs
		}
	}
	return dstSet, removed
}
