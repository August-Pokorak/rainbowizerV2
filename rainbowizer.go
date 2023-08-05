package main

import (
	//"errors"
	"fmt"
	"runtime"

	//"github.com/PerformLine/go-stockutil/colorutil"
	"image"
)

var nThreads = runtime.NumCPU()
var nBuckets = 360

func convertToHSL(img image.Image){

}

func rainbowize(img image.Image){
	//colorutil.Color{}
}

type job struct {
	
	
}


func mapWorker(from []float64, to []float64, f func(float64) float64, done chan bool){
	for i := 0; i < len(from); i++ {
		to[i] = f(from[i])
	}
	done <- true
}

func runMap(from []float64, to []float64, f func(float64) float64) error{
	if len(from) != len(to) {
		return fmt.Errorf("Map failed, from and to fields of different lengths.\nFrom:%d\nTo:%d\n", len(from), len(to))
	}
	rows := len(from)
	rowsPerThread := rows / nThreads
	c := make(chan bool)
	for i := 0; i < nThreads; i++ {
		start := i * rowsPerThread
		var end int
		if i == nThreads-1 {
			end = rows
		} else {
			end = (i+1) * rowsPerThread
		}
		//fmt.Printf("starting worker for range %d:%d\n", start, end)
		go mapWorker(from[start:end], to[start:end], f, c)
	}
	for i := 0; i < nThreads; i++ {
		<- c
	}
	return nil
}

func reduceWorker(data []float64, f func(float64, float64) float64, value chan float64){
	switch len(data) {
	case 0:
		value <- 0	// best guess
		return
	case 1:
		value <- data[0]
		return
	}

	var res = f(data[0], data[1])

	for i := 2; i < len(data); i++ {
		res = f(res, data[i])
	}
	value <- res
}

func runReduce(data []float64, f func(float64, float64) float64) float64{
	rows := len(data)
	rowsPerThread := rows / nThreads
	c := make(chan float64)
	var value float64
	nWorkers := nThreads

	switch len(data) {
	case 0:
		return 0
	case 1:
		return data[0]
	case 2:
		return f(data[0], data[1])
	}

	if nWorkers > len(data) {
		nWorkers = len(data)
	}

	if nWorkers == 1 {
		go reduceWorker(data, f, c)
		return <-c
	}

	for i := 0; i < nWorkers; i++ {
		start := i * rowsPerThread
		var end int
		if i == nWorkers-1 {
			end = rows
		} else {
			end = (i + 1) * rowsPerThread
		}
		//fmt.Printf("starting worker for range %d:%d\n", start, end)
		go reduceWorker(data[start:end], f, c)
	}

	first, second := <-c, <-c

	value = f(first, second)

	for i := 2; i < nWorkers; i++ {
		value = f(value, <- c)
	}
	return value
}

func binOpWorker(a []float64, b []float64, to []float64, f func(float64, float64) float64, done chan bool){
	for i := 0; i < len(a); i++ {
		to[i] = f(a[i], b[i])
	}
	done <- true
}

func runBinOp(a []float64, b []float64, to []float64, f func(float64, float64) float64) error{
	if len(a) != len(to) || len(b) != len(to) {
		return fmt.Errorf("Map failed, a, b, and to fields of different lengths.\na:%d\nb:%d\nto:%d\n", len(a), len(b), len(to))
	}
	rows := len(a)
	rowsPerThread := rows / nThreads
	c := make(chan bool)
	for i := 0; i < nThreads; i++ {
		start := i * rowsPerThread
		var end int
		if i == nThreads-1 {
			end = rows
		} else {
			end = (i+1) * rowsPerThread
		}
		//fmt.Printf("starting worker for range %d:%d\n", start, end)
		go binOpWorker(a[start:end], b[start:end], to[start:end], f, c)
	}
	for i := 0; i < nThreads; i++ {
		<- c
	}
	return nil
}

func linearizeImage(img image.Image) {

}

func main() {
	fmt.Printf("String with %d worker threads\n", nThreads)
	a := make([]float64, 100)
	b := make([]float64, 100)
	//c := make([]float64, 100)
	//d := make([]float64, 100)
	//e := make([]float64, 100)
	for i := 0; i < 100; i++ {
		a[i] = float64(i + 1)
	}

	// copy a to b
	runMap(a, b, func(val float64) float64 {
		return val
	})
	fmt.Println(b[0:10])

	// add a and b and save to b
	runBinOp(a, b, b, func(val1 float64, val2 float64) float64 {
		return val1 + val2
	})
	fmt.Println(b[0:10])


	asum := runReduce(a, func(v1 float64, v2 float64) float64 {
		return v1 + v2
	})
	fmt.Println(asum)

	bsum := runReduce(b, func(v1 float64, v2 float64) float64 {
		return v1 + v2
	})
	fmt.Println(bsum)
}