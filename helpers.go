package prism

import (
	"runtime"
	"sync"
)

// clamp restricts a value to the [0, 255] range.
func clamp(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v + 0.5)
}

// parallel executes fn for each index in [0, count) using
// runtime.NumCPU() goroutines.
func parallel(start, stop int, fn func(int)) {
	count := stop - start
	if count <= 0 {
		return
	}

	procs := runtime.NumCPU()
	if procs > count {
		procs = count
	}
	if procs <= 1 {
		for i := start; i < stop; i++ {
			fn(i)
		}
		return
	}

	var wg sync.WaitGroup
	wg.Add(procs)
	chunkSize := (count + procs - 1) / procs

	for p := 0; p < procs; p++ {
		lo := start + p*chunkSize
		hi := lo + chunkSize
		if hi > stop {
			hi = stop
		}
		go func(lo, hi int) {
			defer wg.Done()
			for i := lo; i < hi; i++ {
				fn(i)
			}
		}(lo, hi)
	}
	wg.Wait()
}

// absInt returns the absolute value of an integer.
func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

// minInt returns the smaller of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the larger of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
