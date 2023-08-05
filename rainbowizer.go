package main

import (
	"bufio"
	"github.com/PerformLine/go-stockutil/colorutil"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"

	//"errors"
	"fmt"
	"runtime"

	"image"
)

var nThreads = runtime.NumCPU()
var nBuckets = 360

func angleSumP(a float64, b float64) float64 {
	return angleToPositive(a + b)
}

func angleSumC(a float64, b float64) float64 {
	return angleToCentered(a + b)
}

func angleToPositive(a float64) float64 {
	r := math.Mod(a, 360)
	if r < 0 {
		r += 360
	}
	return r
}

func angleToCentered(a float64) float64 {
	r := math.Mod(a, 360)
	if r > 180 {
		r -= 360
	} else if r < -180 {
		r += 360
	}
	return r
}

func v1SpreadFunction(x float64) float64 {
	a := float64(.05)
	b := float64(32)
	return (1 + a*x - math.Exp2(-b*x)) / (a + 1)
}

func f64Sum(f float64, f2 float64) float64 {
	return f + f2
}

func rainbowize(img image.Image) image.Image {
	bounds := img.Bounds()
	//fmt.Println(bounds)
	width, height := bounds.Max.X, bounds.Max.Y
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	fmt.Println("Step 1: converting to HSV array")

	totalPx := int64(width) * int64(height)

	h := make([]float64, totalPx)
	s := make([]float64, totalPx)
	l := make([]float64, totalPx)

	weights := make([]float64, totalPx)

	hueX := make([]float64, totalPx)
	hueY := make([]float64, totalPx)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			i := (height * x) + y

			r, g, b, _ := img.At(x, y).RGBA()
			h[i], s[i], l[i] = colorutil.RgbToHsl(
				float64(r)/0xffff*255,
				float64(g)/0xffff*255,
				float64(b)/0xffff*255)
		}
	}

	fmt.Println("Step 2: calculating statistics")
	// calculate weights based on s, l
	runBinOp[float64, float64, float64](s, l, weights, func(f float64, f2 float64) float64 {
		return math.Pow(f * (1 - 2 * math.Abs(f2 - 0.5)), 3)
	})

	totalWeight := runReduce[float64](weights, f64Sum)

	// calculate and weight hue x, y
	runBinOp[float64, float64, float64](h, weights, hueX, func(f float64, f2 float64) float64 {
		return math.Cos(f * math.Pi / 180) * f2
	})
	runBinOp[float64, float64, float64](h, weights, hueY, func(f float64, f2 float64) float64 {
		return math.Sin(f * math.Pi / 180) * f2
	})

	// average hue x, y
	xAvg := runReduce[float64](hueX, f64Sum) / totalWeight
	yAvg := runReduce[float64](hueY, f64Sum) / totalWeight

	//println(xAvg)
	//println(yAvg)

	avgAngle := math.Atan2(yAvg, xAvg) * 180 / math.Pi

	fmt.Println("Step 3: mapping hues")
	// do the hue adjustment
	runMap[float64, float64](h, h, func(f float64) float64 {
		return angleSumP( // 5 decenter
			v1SpreadFunction( // 3 spread
				angleSumC(f, -avgAngle) / // 1 center
					180.0) * // 2 normalize
				180.0, // 4 denormalize
				avgAngle)
	})

	fmt.Println("Step n - 1: creating new image")
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			i := (height * x) + y
			r, g, b := colorutil.HslToRgb(h[i], s[i], l[i])
			newImg.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return newImg
}


func main() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the path to the PNG file: ")
	filePath, _ := reader.ReadString('\n')
	filePath = strings.TrimSpace(filePath)
	//filePath := "23-8-2-Eagle-edited.png\n"

	println(filePath)

	if !strings.Contains(filePath, ".png") {
		fmt.Println("Invalid file type, please provide png")
	}

	newName := filePath[:len(filePath)-4] + "-rainbow.png"

	inFile, err := os.Open(filePath[:len(filePath)])
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

	outFile, err := os.Create(newName)

	defer outFile.Close()

	// Encode the image as PNG and write it to the file
	png.Encode(outFile, rainbow)

	fmt.Println("Done!")
}
