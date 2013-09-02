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
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage of %s:\n", os.Args[0])
	fmt.Fprintln(os.Stderr, os.Args[0], "dir class set")
	flag.PrintDefaults()
}

var (
	numPixels int
)

func init() {
	flag.Usage = usage
	flag.IntVar(&numPixels, "pixels", 100*100, "Rough number of pixels in window")
}

func main() {
	flag.Parse()
	if flag.NArg() != 3 {
		flag.Usage()
		os.Exit(1)
	}
	var (
		dir   = flag.Arg(0)
		class = flag.Arg(1)
		set   = flag.Arg(2)
	)

	// Load list of images.
	log.Println("load annotations")
	imgset, err := voc.Load(dir, class, set)
	if err != nil {
		log.Fatalln("could not load list of images:", err)
	}

	// Trim difficult, occluded or truncated examples.
	var n int
	imgset, n = removeDifficult(imgset)
	log.Println("removed", n, "windows: difficult")

	// Get image sizes.
	log.Println("get size of images")
	sizes := make(map[string]image.Point, len(imgset))
	for name := range imgset {
		conf, err := loadImageConfig(voc.ImageFile(dir, name))
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
	width := round(math.Sqrt(float64(numPixels) * aspect))
	height := round(math.Sqrt(float64(numPixels) / aspect))
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
	// Extract windows from images.
	for name, objs := range imgset {
		// Load image from file.
		img, err := loadImage(voc.ImageFile(dir, name))
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
			subimg := sub.SubImage(obj.Region)
			// Resize.
			subimg = resize.Resize(uint(width), uint(height), subimg, resize.Bilinear)
			if err := saveImage(subimg, fmt.Sprintf("%s_%d.png", name, i)); err != nil {
				log.Println("could not save image:", err)
			}
		}
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
func removeDifficult(set map[string][]voc.Object) (map[string][]voc.Object, int) {
	dstSet := make(map[string][]voc.Object, len(set))
	var removed int
	for name, objs := range set {
		var dstObjs []voc.Object
		for _, obj := range objs {
			if obj.Occluded != nil && *obj.Occluded {
				removed++
				continue
			}
			if obj.Truncated != nil && *obj.Truncated {
				removed++
				continue
			}
			if obj.Difficult != nil && *obj.Difficult {
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
