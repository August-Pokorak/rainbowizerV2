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
	r := math.Mod(a, 360.0)
	if r < 0 {
		r += 360.0
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
	a := float64(1.2)
	b := float64(64)
	c := float64(0.05)
	x1 := math.Abs(x)
	v := (1 - math.Exp(-b * math.Pow(x1, a)) + (c * x1)) / (1 + c)
	if math.Signbit(x) {
		return v
	} else {
		return -v
	}

}

func f64Sum(f float64, f2 float64) float64 {
	return f + f2
}


func calcHist(from []float64, to []float64, weights []float64) {
	bucketSize := 360 / float64(nBuckets)
	hBucketed := make([]int, len(from))
	runMap[float64, int](from, hBucketed, func(x float64) int {
		return int(math.Floor(x / bucketSize))
	})
	runHist(hBucketed, to, weights, nBuckets)
}


func lerp (a float64, b float64, t float64) float64 {
	return a + t*(b-a)
}


func spreadIter(from []float64, to []float64, weights []float64, t float64) {
	bucketSize := 360 / float64(nBuckets)

	hist := make([]float64, nBuckets)

	weightsNorm := make([]float64, len(from))
	weightsForDist := make([]float64, len(from))
	runMap[float64](weights, weightsNorm, func(f float64) float64 {
		return f * float64(len(from))
	})

	//println(runReduce[float64](weightsNorm, f64Sum) / float64(len(from)))


	a := float64(1)
	runMap[float64](weightsNorm, weightsForDist, func(f float64) float64 {
		return (f + a) / (a + 1)
	})

	calcHist(from, hist, weights)

	//fmt.Println(runReduce[float64](hist, f64Sum))

	//fmt.Println(hist)

	// calculate slopes (very basic)
	//runMap[float64, float64](histNorm, slopes, func(f float64, f2 float64) {
	//	return
	//})



	// move all values down slope
	runBinOp[float64, float64, float64](from, weightsForDist, to, func(f float64, w float64) float64 {
		bucketf := f / bucketSize
		bucketi := int(math.Floor(bucketf))
		var delta float64
		this := hist[bucketi]
		before := hist[(bucketi + nBuckets - 1) % nBuckets]
		after := hist[(bucketi + 1) % nBuckets]
		dist := (1 + bucketf - float64(bucketi)) / 3.0

		ta := lerp(this, after, dist)
		bt := lerp(before, this, dist)
		delta = t * w * (ta - bt) / (bt + ta)

		return angleToPositive(f - delta)
	})
}


func rainbowize(img image.Image) image.Image {
	bounds := img.Bounds()
	//fmt.Println(bounds)
	width, height := bounds.Max.X, bounds.Max.Y
	newImg := image.NewRGBA(image.Rect(0, 0, width, height))

	fmt.Println("Step 1: converting to HSV array")

	totalPx := width * height

	alpha := make([]uint32, totalPx)

	h := make([]float64, totalPx)
	s := make([]float64, totalPx)
	l := make([]float64, totalPx)

	weights := make([]float64, totalPx)

	hist := make([]float64, nBuckets)


	//hueX := make([]float64, totalPx)
	//hueY := make([]float64, totalPx)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			i := (height * x) + y

			r, g, b, a := img.At(x, y).RGBA()
			alpha[i] = a
			h[i], s[i], l[i] = colorutil.RgbToHsl(
				float64(r)/0xffff*255,
				float64(g)/0xffff*255,
				float64(b)/0xffff*255)
		}
	}

	fmt.Println("Step 2: calculating statistics")
	// calculate weights based on s, l
	runBinOp[float64, float64, float64](s, l, weights, func(f float64, f2 float64) float64 {
		return math.Pow(f * (1 - 2 * math.Abs(f2 - 0.5)), 2)
	})
	weightSum := runReduce[float64](weights, f64Sum)
	// normalize weights
	runMap[float64, float64](weights, weights, func(f float64) float64 {
		return f / weightSum
	})

	fmt.Println(runReduce[float64](weights, f64Sum))


	//println(h)
	fmt.Println("Initial hist:")
	calcHist(h, hist, weights)
	fmt.Println(hist)
	//fmt.Println(runReduce[float64](hist, f64Sum))

	fmt.Println("Step 2: flatting the curve")

	iters := 1000
	for i := 0; i < iters; i++ {
		spreadIter(h, h, weights, 0.2)
		//calcHist(h, hist, weights)
		//fmt.Println(hist)
		//fmt.Println(runReduce[float64](hist, f64Sum))
		if (i + 1) % (iters / 100) == 0 {
			fmt.Print(".")
			if (i + 1) % (iters / 5) == 0 {
				fmt.Print("\n")
			}
		}
	}

	fmt.Println("Final hist:")
	calcHist(h, hist, weights)
	fmt.Println(hist)
	//fmt.Println(runReduce[float64](hist, f64Sum))


	fmt.Println("Step n - 1: creating new image")
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			i := (height * x) + y
			r, g, b := colorutil.HslToRgb(h[i], s[i], l[i])
			newImg.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: uint8(alpha[i])})
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

	//println(filePath)

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
