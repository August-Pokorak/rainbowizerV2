package main

import (
	"github.com/PerformLine/go-stockutil/colorutil"
	"image/color"
	"image/png"
	"math"
	"os"
	//"errors"
	"fmt"
	"runtime"

	"image"
)

var nThreads = runtime.NumCPU()
var nBuckets = 360

func convertToHSL(img image.Image){

}


func rainbowize(img image.Image) image.Image{
	bounds := img.Bounds()
	//fmt.Println(bounds)
	width, height := bounds.Max.X, bounds.Max.Y
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	fmt.Println("Step 1: converting to HSV array")

	totalPx := int64(width) * int64(height)
	totalPxf := float64(totalPx)


	h := make([]float64, totalPx)
	s := make([]float64, totalPx)
	l := make([]float64, totalPx)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			i := (height * x) + y

			r, g, b, _ := img.At(x, y).RGBA()
			h[i], s[i], l[i] = colorutil.RgbToHsl(
				float64(r) / 0xffff * 255,
				float64(g) / 0xffff * 255,
				float64(b) / 0xffff * 255)
		}
	}

	avgHue := runReduce(h, func(f float64, f2 float64) float64 {
		return f + f2
	}) / totalPxf
	fmt.Println(avgHue)

	runMap(h, h, func(f float64) float64 {
		return math.Mod(f + 150, 360)
	})

	avgHue = runReduce(h, func(f float64, f2 float64) float64 {
		return f + f2
	}) / totalPxf
	fmt.Println(avgHue)

	fmt.Println("Step n - 1: creating new image")
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			i := (height * x) + y
			r, g, b := colorutil.HslToRgb(h[i],s[i],l[i])
			newImg.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return newImg
}

func linearizeImage(img image.Image) {

}

func main() {

	//reader := bufio.NewReader(os.Stdin)
	//fmt.Print("Enter the path to the PNG file: ")
	//filePath, _ := reader.ReadString('\n')
	filePath := "23-8-2-Eagle-edited.png\n"

	inFile, err := os.Open(filePath[:len(filePath)-1])
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer inFile.Close()

	img, err := png.Decode(inFile)
	if err != nil {
		fmt.Println("Error decoding PNG:", err)
		return
	}
	inFile.Close()

	rainbow := rainbowize(img)

	fmt.Println("Step n: writing")

	outFile, err := os.Create("output.png")

	defer outFile.Close()

	// Encode the image as PNG and write it to the file
	png.Encode(outFile, rainbow)

	fmt.Println("Done!")
}