package main

import (
	//"encoding/base64"
	"fmt"
	"flag"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	//	"math/cmplx"
	//	"strings"
)

//Box Blur kernel
var boxKernel = [][]float64{
	{.9, .9, .9},
	{.9, .9, .9},
	{.9, .9, .9}}

//Gaussian kernel
var gaussianKernel = [][]float64{
	{1 / 256, 4 / 256, 6 / 256, 4 / 256, 1 / 256},
	{4 / 256, 16 / 256, 24 / 256, 16 / 256, 4 / 256},
	{6 / 256, 24 / 256, 36 / 256, 24 / 256, 6 / 256},
	{4 / 256, 16 / 256, 24 / 256, 16 / 256, 4 / 256},
	{1 / 256, 4 / 256, 6 / 256, 4 / 256, 1 / 256}}

var localTestFile = "testImage.png"
var file = flag.String("f", "testImage.png", "image file path")
var dir = flag.String("d", ".", "directory to load images")
var kernel = flag.String("k", "boxBlur", "kernel filter to apply to image(s)")

func main() {
	flag.Parse()
	if *dir == "." {
		fmt.Println("here")
		reader, err := os.Open(localTestFile)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()
		img, _, err := image.Decode(reader)
		if err != nil {
			log.Fatal(err)
		}
		//go boxKernel(img)
		f, err := os.Create("testImageRun.png")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		nimg := boxKernelFilter(img, boxKernel)
		//fmt.Println(nimg)
		png.Encode(f, nimg)

	}

}

func boxKernelFilter(img image.Image, kernel [][]float64) image.Image {
	kernelOffset := len(kernel) / 2
	bounds := img.Bounds()
	newImg := image.NewRGBA(image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y))
	y := 0
	x := 0
	a := 0
	bb := 0
	for x < bounds.Max.X-kernelOffset {
		for y < bounds.Max.Y-kernelOffset {
			nr, ng, nb,na := 0.0, 0.0, 0.0, 0.0
			for a < len(kernel) {
				for bb < len(kernel[0]) {
					xn := x + a - kernelOffset
					yn := y + bb - kernelOffset
					pix := img.At(xn, yn)
					r, g, b, aa := pix.RGBA()
					nr += float64(r) * kernel[a][bb]
					ng += float64(g) * kernel[a][bb]
					nb += float64(b) * kernel[a][bb]
					na = float64(aa)
					bb++
				}
				bb = 0
				a++
			}
			a = 0
			y++
			newImg.Set(x, y, color.RGBA{uint8(nr), uint8(ng), uint8(nb), uint8(na)})
		}
		y = 0
		x++
	}
	return newImg
}