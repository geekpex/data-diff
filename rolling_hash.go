package main

import "fmt"

const (
	windowSize = 16

	pM = uint64(1000000009)
)

// shiftMultiplier is used to precalculate last byte's multiplier in rolling hash window
func shiftMultiplier() (m uint64) {
	m = 256

	for i := 2; i < windowSize; i++ {
		m = (m % pM) * 256
	}

	return
}

var shiftM = shiftMultiplier()

// SingleHash contains rolling hash and index in data from which the hash was calculated
type SingleHash struct {
	i int
	h uint64
}

// calcRollingHash gets input data and sends rolling hash of it to out channel.
func calcRollingHash(data []byte, out chan<- SingleHash) {
	defer close(out)

	if len(data) < windowSize {
		panic(fmt.Sprintf("data to be read by rolling hash has to have minimum size of %d bytes", windowSize))
	}

	var sh SingleHash
	var hash uint64
	var i int
	var l = len(data)

	hash = uint64(data[0])

	for i = 1; i < windowSize; i++ {
		hash = (hash * 256) % pM

		hash = (hash + uint64(data[i])) % pM
	}

	sh = SingleHash{i: windowSize - 1, h: hash}
	out <- sh

	for i = windowSize; i < l; i++ {
		// For my sake the calculation has been separated to smaller pieces.
		// This might cause extra allocations.
		hash += pM
		hash -= uint64(uint64(data[i-windowSize]) * shiftM % pM)
		hash *= 256
		hash = (hash + uint64(data[i])) % pM

		sh.h = hash
		sh.i = i
		out <- sh
	}
}
