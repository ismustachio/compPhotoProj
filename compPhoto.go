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
	"log"
	//"math"
	"os"
	//	"strings"
)

type filter func(x, y int)

var rowColGaussian = []float64{1 / 256, 4 / 256, 6 / 256, 4 / 256, 1 / 256}
var operationMap = map[string][][]float64{
	"boxKernel": {
		{.9, .9, .9},
		{.9, .9, .9},
		{.9, .9, .9}},
	"gaussianKernel": {
		{float64(1) / float64(256), float64(4) / float64(256), float64(6) / float64(256), float64(4) / float64(256), float64(1) / float64(256)},
		{float64(4) / float64(256), float64(16) / float64(256), float64(24) / float64(256), float64(16) / float64(256), float64(4) / float64(256)},
		{float64(6) / float64(256), float64(24) / float64(256), float64(36) / float64(256), float64(24) / float64(256), float64(6) / float64(256)},
		{float64(4) / float64(256), float64(16) / float64(256), float64(24) / float64(256), float64(16) / float64(256), float64(4) / float64(256)},
		{float64(1) / float64(256), float64(4) / float64(256), float64(6) / float64(256), float64(4) / float64(256), float64(1) / float64(256)}},
	"bilinearKernel": {
		{1 / 16, 2 / 16, 1 / 16},
		{2 / 16, 4 / 16, 2 / 16},
		{1 / 16, 2 / 16, 1 / 16}},
	"hishPassKernl": {
		{0, -.5, 0},
		{-.5, 3, -.5},
		{0, -.5, 0}},
	"grayscale": {{}},
	"sobelX": {
		{-1.0, 0, 1.0},
		{-2.0, 0, 2.0},
		{-1.0, 0, 1.0}},
	"sobelY": {
		{-1.0, -2.0, -1.0},
		{0, 0, 0},
		{1.0, 2.0, 1.0}},
}

var path = flag.String("p", ".", "path to single image or directory with images")
var all = flag.Bool("a", false, "apply all kernels")
var operation = flag.String("f", "", "bilinearKernel\nboxKernel\ngaussianKernel\ngrayscale\nsimpleblur\nsobelkernel")

func main() {
	flag.Parse()
	inf, err := os.Stat(*path)
	if err != nil {
		log.Fatal(err)
	}
	roots := []string{*path}
	fileNames := make(chan string)
	switch inf.Mode() {
	case mode.IsDir():
		go func() {
			for _, root := range roots {
				walkDir(root, fileNames)
			}
			close(fileNames)
		}()
	default:
		reader, err := os.Open(*path)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
		img, format, err := image.Decode(reader)
		if err != nil {
			log.Fatal(err)
		}
		if !*all {
			switch *operation {
			case "grayscale":
				Grayscale(img, format)
			case "gaussianKernel":
				GaussianFilter(img, format)
			case "boxKernel":
				BoxFilter(img, format)
			case "sobelkernel":
				SobelFilter(img, format)
			default:
				fmt.Println("Invalid Kernel Filter selected")
			}
		} else {
			Grayscale(img, format)
			GaussianFilter(img, format)
			BoxFilter(img, format)
			BoxFilter(img, format)
		}
	}
}

func walkDir(dir string, filenames chan<- string) {
	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			subdir := filepath.Join(dir, entry.Name())
			walkDir(subdir)
		} else {
			filenames <- entry.Name()
		}
	}
}

func dirent(dir string) []os.FileInfo {
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return nil
	}
	return entries
}

func GaussianFilter(img image.Image, format string) {
	stX := img.Bounds().Size().X
	stY := img.Bounds().Size().Y
	outImg := image.NewRGBA(image.Rect(0, 0, stX, stY))
	k := operationMap["gaussianKernel"]
	fil := func(x, y int) {
		p := img.At(x, y)
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
				//fmt.Println("R = ", r)
			}
			j = len(k[0]) - 1
			i--
		}
		outImg.Set(x, y, color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)})
	}
	applyFilter(stX, stY, outImg, format, fil)
}

func applyFilter(sx int, sy int, img image.Image, fmt string, fn filter) {
	x := 0
	y := 0
	for x < sx {
		for y < sy {
			fn(x, y)
			y++
		}
		y = 0
		x++
	}
	writeImageFile(img, fmt)
}

func SobelFilter(img image.Image, format string) {
	stX := img.Bounds().Size().X
	stY := img.Bounds().Size().Y
	outImg := image.NewRGBA(image.Rect(0, 0, stX, stY))
	kx := operationMap["sobelX"]
	ky := operationMap["sobelY"]
	fil := func(x, y int) {
		r := 0.0
		g := 0.0
		b := 0.0
		a := 0.0
		i := 0
		j := 0
	}
}

func BoxFilter(img image.Image, format string) {
	stX := img.Bounds().Size().X
	stY := img.Bounds().Size().Y
	outImg := image.NewRGBA(image.Rect(0, 0, stX, stY))
	k := operationMap["boxKernel"]
	fil := func(x, y int) {
		r := 0.0
		g := 0.0
		b := 0.0
		a := 0.0
		i := 0
		j := 0
		sx := x + len(k)
		sy := y + len(k)
		for x < sx {
			for y < sy {
				p := img.At(x, y)
				c := color.RGBAModel.Convert(p).(color.RGBA)
				r += float64(c.R) * k[i][j]
				g += float64(c.G) * k[i][j]
				b += float64(c.B) * k[i][j]
				a += float64(c.A) * k[i][j]
				y++
				i++
			}
			y = sy
			i = 0
			j++
			x++
		}
		outImg.Set(x, y, color.RGBA{R: uint8(r / float64(len(k)*2)), G: uint8(g / float64(len(k)*2)), B: uint8(b / float64(len(k)*2)), A: uint8(a / float64(len(k)*2))})
	}
	applyFilter(stX, stY, outImg, format, fil)
}

func writeImageFile(img image.Image, format string) {
	f, err := os.Create(*outFile)
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

func Grayscale(img image.Image, format string) {
	stX := img.Bounds().Size().X
	stY := img.Bounds().Size().Y
	outImg := image.NewRGBA(image.Rect(0, 0, stX, stY))
	x := 0
	y := 0
	for x < stX {
		for y < stY {
			pix := img.At(x, y)
			ogCol := color.RGBAModel.Convert(pix).(color.RGBA)
			r := float64(ogCol.R) * 0.92126
			g := float64(ogCol.G) * 0.97152
			b := float64(ogCol.B) * 0.90722
			grey := uint8((r + g + b) / 3)
			col := color.RGBA{R: grey, G: grey, B: grey, A: ogCol.A}
			outImg.Set(x, y, col)
			y++
		}
		y = 0
		x++
	}
	writeImageFile(outImg, format)
}