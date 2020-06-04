//Author: Armando Lajara
//Email: alajara@pdx.edu

// Computational Photography Project
// This project attempts to act as a command-line utility, image filter.
// To maxize this utility, it is best to run against large quantaties of // images. This utility will spawned a thread (goroutine) per each image
// using all available system resources.

package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type kernels int

const (
	Blur kernels = iota
	BSobel
	LSobel
	RSobel
	TSobel
	Emboss
	Identity
	Outline
	Sharp
	GrayScale
	Gaussian
	Custom
)

type filter func(int, int, image.Image, *image.RGBA, [][]float64)

func Filter(x int, y int, img image.Image, out *image.RGBA, k [][]float64) {
	r := 0.0
	g := 0.0
	b := 0.0
	a := 0.0
	i := 0
	j := 0
	off := len(k) / 2
	st := 0
	for st < *runTime {
		for i < len(k) {
			for j < len(k[0]) {
				c := color.RGBAModel.Convert(img.At(x+i-off, y+j-off)).(color.RGBA)
				r += float64(c.R) * k[i][j]
				g += float64(c.G) * k[i][j]
				b += float64(c.B) * k[i][j]
				a += float64(c.A) * k[i][j]
				j++
			}
			j = 0
			i++
		}
		st++
	}
	out.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)})
}

func Gray(img image.Image, f string, format string) int {
	stX := img.Bounds().Size().X
	stY := img.Bounds().Size().Y
	x := 0
	y := 0
	i := 0
	out := image.NewRGBA(image.Rect(0, 0, stX, stY))
	for i < *runTime {
		for x < stX {
			for y < stY {
				c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
				r := float64(c.R) * 0.92126
				g := float64(c.G) * 0.97152
				b := float64(c.B) * 0.90722
				grey := uint8((r + g + b) / 3)
				col := color.RGBA{R: grey, G: grey, B: grey, A: c.A}
				out.Set(x, y, col)
				y++
			}
			y = 0
			x++
		}
		i++
	}
	writeImageFile(out, f, format)
	return 1
}

var sema = make(chan struct{}, 20)
var fpath = flag.String("p", ".", "path to single image or directory with images")
var operation = flag.String("f", "", "Available kernel's: \n[LRTB]Sobel\nBlur\nSharp\nOutline\nIdentity\nEmboss\nAvailable Image Transformation:\nGrayScale")
var runTime = flag.Int("r", 1, "times to run the given convolution kernel")
var custom = flag.String("c", "", "Use custom 3x3 matrix. Ex: -c '3 2 -1 8 9 2 1 -9 1' ")

func main() {
	flag.Parse()
	total := 0
	fi, err := os.Stat(*fpath)
	if err != nil {
		log.Fatal(err)
	}
	roots := []string{*fpath}
	fileNames := make(chan string)
	t0 := time.Now()
	go spinner(100 * time.Millisecond)
	k := Identity
	if *custom != "" {
		if len(strings.Split(*custom, " ")) != 9 {
			flag.Usage()
			return
		}
		k = Custom
	} else {
		switch *operation {
		case "LSobel":
			k = LSobel
		case "RSobel":
			k = RSobel
		case "TSobel":
			k = TSobel
		case "BSobel":
			k = BSobel
		case "Emboss":
			k = Emboss
		case "Identity":
			k = Identity
		case "Outline":
			k = Outline
		case "Sharp":
			k = Sharp
		case "GrayScale":
			k = GrayScale
		case "Gaussian":
			k = Gaussian
		case "Blur":
			k = Blur
		default:
			flag.Usage()
			return
		}
	}
	kk := kernel(k)
	switch mode := fi.Mode(); {
	case mode.IsDir():
		var n sync.WaitGroup
		for _, root := range roots {
			n.Add(1)
			go walkDir(root, &n, fileNames)
		}
		go func() {
			n.Wait()
			close(fileNames)
		}()
		total = ApplyFilter(fileNames, kk, Filter)
	default:
		total = singleImage(*fpath, kk, Filter)
	}
	fmt.Println("Total files: ", total)
	t1 := time.Now()
	fmt.Println("Done!")
	fmt.Printf("Total Elapsed time: %v.\n", t1.Sub(t0))
}

func singleImage(f string, k [][]float64, fn filter) int {
	img, format, err := extractImage(f)
	if err != nil {
		log.Fatal(err)
	}
	if *operation == "GrayScale" {
		Gray(img, f, format)
	} else {
		convolutionSingle(img, k, fn, f, format)
	}
	return 1
}

func kernel(k kernels) [][]float64 {
	out := make([][]float64, 3)
	for k, _ := range out {
		out[k] = make([]float64, 3)
	}
	switch k {
	case BSobel:
		out = [][]float64{{-1, -2, -1}, {0, 0, 0}, {1, 2, 1}}
	case LSobel:
		out = [][]float64{{1, 0, -1}, {2, 0, -2}, {1, 0, -1}}
	case RSobel:
		out = [][]float64{{-1, 0, 1}, {-2, 0, 2}, {-1, 0, 1}}
	case TSobel:
		out = [][]float64{{1, 2, 1}, {0, 0, 0}, {-1, -2, -1}}
	case Emboss:
		out = [][]float64{{-2, -1, 0}, {-1, 1, 1}, {0, 1, 2}}
	case Identity:
		out = [][]float64{{0, 0, 0}, {0, 1, 0}, {0, 0, 0}}
	case Outline:
		out = [][]float64{{-1, -1, -1}, {-1, 8, -1}, {-1, -1, -1}}
	case Sharp:
		out = [][]float64{{0, -.6, 0}, {-.6, 3, -.6}, {0, -.6, 0}}
	case Blur:
		out = [][]float64{{.0625, .125, .0625}, {.125, .25, .125}, {.0625, .125, .0625}}
	case Gaussian:
		out = [][]float64{{0.00390625, 0.015625, 0.0234375, 0.015625, 0.00390625}, {0.015625, 0.0625, 0.09375, 0.0625, 0.015625}, {0.0234375, 0.09375, 0.140625, 0.09375, 0.0234375}, {0.015625, 0.0625, 0.09375, 0.0625, 0.015625},
			{0.00390625, 0.015625, 0.0234375, 0.015625, 0.00390625}}
	case Custom:
		spl := strings.Split(*custom, " ")
		a, err := strconv.ParseFloat(spl[0], 64)
		if err != nil {
			log.Fatal(err)
		}
		b, err := strconv.ParseFloat(spl[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		c, err := strconv.ParseFloat(spl[2], 64)
		if err != nil {
			log.Fatal(err)
		}
		d, err := strconv.ParseFloat(spl[3], 64)
		if err != nil {
			log.Fatal(err)
		}
		e, err := strconv.ParseFloat(spl[4], 64)
		if err != nil {
			log.Fatal(err)
		}
		f, err := strconv.ParseFloat(spl[5], 64)
		if err != nil {
			log.Fatal(err)
		}
		g, err := strconv.ParseFloat(spl[6], 64)
		if err != nil {
			log.Fatal(err)
		}
		h, err := strconv.ParseFloat(spl[7], 64)
		if err != nil {
			log.Fatal(err)
		}
		i, err := strconv.ParseFloat(spl[8], 64)
		if err != nil {
			log.Fatal(err)
		}
		out = [][]float64{{a, b, c}, {d, e, f}, {g, h, i}}
	}
	return out
}

func walkDir(dir string, n *sync.WaitGroup, filenames chan<- string) {
	defer n.Done()
	for _, entry := range dirent(dir) {
		if entry.IsDir() {
			n.Add(1)
			subdir := filepath.Join(dir, entry.Name())
			walkDir(subdir, n, filenames)
		} else {
			filenames <- path.Join(dir, entry.Name())
		}
	}
}

func dirent(dir string) []os.FileInfo {
	sema <- struct{}{}
	defer func() { <-sema }()
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil
	}
	return entries
}

func extractImage(f string) (image.Image, string, error) {
	reader, err := os.Open(f)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	defer reader.Close()
	img, format, err := image.Decode(reader)
	if err != nil {
		log.Println(err)
		return nil, "", err
	}
	return img, format, nil
}

func convolutionChannel(img image.Image, k [][]float64, fn filter, name string, format string, count chan int) {
	x := len(k) / 2
	y := x
	stX := img.Bounds().Size().X - x
	stY := img.Bounds().Size().Y - x
	out := image.NewRGBA(image.Rect(0, 0, stX, stY))
	for x < stX {
		for y < stY {
			fn(x, y, img, out, k)
			y++
		}
		y = 0
		x++
		count <- 1
	}
	writeImageFile(out, name, format)
}

func convolutionSingle(img image.Image, k [][]float64, fn filter, name string, format string) {
	x := len(k) / 2
	y := x
	stX := img.Bounds().Size().X - x
	stY := img.Bounds().Size().Y - x
	out := image.NewRGBA(image.Rect(0, 0, stX, stY))
	for x < stX {
		for y < stY {
			fn(x, y, img, out, k)
			y++
		}
		y = 0
		x++
	}
	writeImageFile(out, name, format)
}

func ApplyFilter(name <-chan string, k [][]float64, fn filter) int {
	count := make(chan int)
	var wg sync.WaitGroup //num of go routines
	for f := range name {
		wg.Add(1)
		//worker
		go func(f string) {
			defer wg.Done()
			img, format, err := extractImage(f)
			if err == nil {
				if *operation == "GrayScale" {
					count <- Gray(img, f, format)
				} else {
					convolutionChannel(img, k, fn, f, format, count)
				}
			}
		}(f)
	}
	//closer
	go func() {
		wg.Wait()
		close(count)
	}()
	var total int
	for tot := range count {
		total += tot
	}
	return total
}

func writeImageFile(img image.Image, name string, format string) {
	p, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	np := path.Join(p, "outImages/")
	_, err = os.Stat(np)
	if err != nil {
		if err = os.Mkdir(np, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	f, err := os.Create(path.Join(np, filepath.Base(name)))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if format == "jpeg" {
		jpeg.Encode(f, img, nil)
	} else {
		png.Encode(f, img)
	}

}

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
