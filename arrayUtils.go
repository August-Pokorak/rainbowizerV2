package main

import "fmt"

func mapWorker[F any, T any](from []F, to []T, f func(F) T, done chan bool){
	for i := 0; i < len(from); i++ {
		to[i] = f(from[i])
	}
	done <- true
}

func runMap[F any, T any](from []F, to []T, f func(F) T) error{
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
		go mapWorker[F, T](from[start:end], to[start:end], f, c)
	}
	for i := 0; i < nThreads; i++ {
		<- c
	}
	return nil
}

func reduceWorker[T any](data []T, f func(T, T) T, value chan T){
	switch len(data) {
	case 0:
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

func runReduce[T any](data []T, f func(T, T) T) T{
	rows := len(data)
	rowsPerThread := rows / nThreads
	c := make(chan T)
	var value T
	nWorkers := nThreads

	switch len(data) {
	case 0:
		return data[0]
	case 1:
		return data[0]
	case 2:
		return f(data[0], data[1])
	}

	if nWorkers > len(data) {
		nWorkers = len(data)
	}

	if nWorkers == 1 {
		go reduceWorker[T](data, f, c)
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
		go reduceWorker[T](data[start:end], f, c)
	}

	first, second := <-c, <-c

	value = f(first, second)

	for i := 2; i < nWorkers; i++ {
		value = f(value, <- c)
	}
	return value
}

func binOpWorker[A any, B any, T any](a []A, b []B, to []T, f func(A, B) T, done chan bool){
	for i := 0; i < len(a); i++ {
		to[i] = f(a[i], b[i])
	}
	done <- true
}

func runBinOp[A any, B any, T any](a []A, b []B, to []T, f func(A, B) T) error{
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
		go binOpWorker[A, B, T](a[start:end], b[start:end], to[start:end], f, c)
	}
	for i := 0; i < nThreads; i++ {
		<- c
	}
	return nil
}

func histWorker(on []int, to []float64, weights []float64, done chan []float64){
	for i := 0; i < len(on); i++ {
		to[on[i]] += weights[i]
	}
	done <- to
}

func runHist(on []int, to []float64, weights []float64, max int) error{
	if len(on) != len(weights) {
		return fmt.Errorf("Hist failed, on, weights of different length.\non:%d\nweights:%d\nto:%d\n", len(on), len(to))
	}
	rows := len(on)
	rowsPerThread := rows / nThreads
	c := make(chan []float64)
	for i := 0; i < nThreads; i++ {
		start := i * rowsPerThread
		var end int
		if i == nThreads-1 {
			end = rows
		} else {
			end = (i+1) * rowsPerThread
		}
		out := make([]float64, max)
		for j := 0; j < max; j++ {
			out[j] = 0
		}
		//fmt.Printf("starting worker for range %d:%d\n", start, end)
		go histWorker(on[start:end], out, weights[start:end], c)
	}

	for i := 0; i < max; i++ {
		to[i] = 0
	}

	for i := 0; i < nThreads; i++ {
		out := <- c
		for j := 0; j < max; j++ {
			to[j] += out[j]
		}
	}
	return nil
}

