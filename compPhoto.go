//Author: Armando Lajara
//Email: alajara@pdx.edu

package main

import (
	//"encoding/base64"
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
	"sync"
	"time"
)

type filter func(int, int, color.Color, *image.RGBA)

func Gauss(x int, y int, p color.Color, out *image.RGBA) {
	k := [][]float64{
		{0.00390625, 0.015625, 0.0234375, 0.015625, 0.00390625},
		{0.015625, 0.0625, 0.09375, 0.0625, 0.015625},
		{0.0234375, 0.09375, 0.140625, 0.09375, 0.0234375},
		{0.015625, 0.0625, 0.09375, 0.0625, 0.015625},
		{0.00390625, 0.015625, 0.0234375, 0.015625, 0.00390625}}
	c := color.RGBAModel.Convert(p).(color.RGBA)
	r := 0.0
	g := 0.0
	b := 0.0
	a := 0.0
	i := len(k) - 1
	j := i
	for i > 0 {
		for j > 0 {
			r += float64(c.R) * k[j][i]
			g += float64(c.G) * k[j][i]
			b += float64(c.B) * k[j][i]
			a += float64(c.A) * k[j][i]
			j--
		}
		j = len(k[0]) - 1
		i--
	}
	out.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)})
}

func Box(x int, y int, p color.Color, out *image.RGBA) {
	k := [][]float64{
		{.9, .9, .9},
		{.9, .9, .9},
		{.9, .9, .9},
	}
	c := color.RGBAModel.Convert(p).(color.RGBA)
	r := 0.0
	g := 0.0
	b := 0.0
	a := 0.0
	i := 0
	j := 0
	for i < len(k) {
		for j < len(k[0]) {
			r += float64(c.R) * k[i][j]
			g += float64(c.G) * k[i][j]
			b += float64(c.B) * k[i][j]
			a += float64(c.A) * k[i][j]
			j++
		}
		j = 0
		i++
	}
	out.Set(x, y, color.RGBA{R: uint8(r / float64(len(k)*2)), G: uint8(g / float64(len(k)*2)), B: uint8(b / float64(len(k)*2)), A: uint8(a / float64(len(k)*2))})
}

func Gray(x int, y int, p color.Color, out *image.RGBA) {
	c := color.RGBAModel.Convert(p).(color.RGBA)
	r := float64(c.R) * 0.92126
	g := float64(c.G) * 0.97152
	b := float64(c.B) * 0.90722
	grey := uint8((r + g + b) / 3)
	col := color.RGBA{R: grey, G: grey, B: grey, A: c.A}
	out.Set(x, y, col)
}

var operationMap = map[string]bool{"bilinearKernel": true, "boxKernel": true, "gaussianKernel": true, "grayscale": true, "simpleblur": true, "sobelkernel": true}
var sema = make(chan struct{}, 20)
var fpath = flag.String("p", ".", "path to single image or directory with images")
var all = flag.Bool("a", false, "apply all kernels")
var operation = flag.String("f", "", "bilinearKernel\nboxKernel\ngaussianKernel\ngrayscale\nsimpleblur\nsobelkernel")

func main() {
	flag.Parse()
	_, ok := operationMap[*operation]
	if !ok {
		flag.Usage()
	}
	fi, err := os.Stat(*fpath)
	if err != nil {
		log.Fatal(err)
	}
	roots := []string{*fpath}
	fileNames := make(chan string)
	t0 := time.Now()
	go spinner(100 * time.Millisecond)
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
	loop:
		for {
			select {
			case name, ok := <-fileNames:
				if !ok {
					break loop
				}
				switch *operation {
				case "grayscale":
					ApplyFilter(name, Gray)
				case "gaussianKernel":
					ApplyFilter(name, Gauss)
				case "boxKernel":
					ApplyFilter(name, Box)
				}
			}
		}

	default:
		if !*all {
			switch *operation {
			case "grayscale":
				ApplyFilter(filepath.Base(*fpath), Gray)
			case "gaussianKernel":
				ApplyFilter(filepath.Base(*fpath), Gauss)
			case "boxKernel":
				ApplyFilter(filepath.Base(*fpath), Box)
			case "sobelkernel":
				//ApplyFilter(filepath.Base(*fpath),Box)
			default:
				fmt.Println("Invalid Kernel Filter selected")
			}
		} else {
			ApplyFilter(filepath.Base(*fpath), Gray)
			ApplyFilter(filepath.Base(*fpath), Gauss)
			ApplyFilter(filepath.Base(*fpath), Box)
			//BoxFilter(filepath.Base(*fpath),format)
		}
	}
	t1 := time.Now()
	fmt.Println("Done!")
	fmt.Printf("Total Elapsed time: %v.\n", t1.Sub(t0))
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

func ApplyFilter(name string, fn filter) {
	reader, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()
	img, format, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	stX := img.Bounds().Size().X
	stY := img.Bounds().Size().Y
	x := 0
	y := 0
	out := image.NewRGBA(image.Rect(0, 0, stX, stY))
	for x < stX {
		for y < stY {
			fn(x, y, img.At(x, y), out)
			y++
		}
		y = 0
		x++
	}
	writeImageFile(out, name, format)
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
		fmt.Println("Create")
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
