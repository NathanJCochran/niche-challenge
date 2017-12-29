package main

import (
	"fmt"
	"testing"
)

const (
	urlFile      = "./urls.txt"
	stopwordFile = "./stopWords.txt"
)

var index Index

func BenchmarkNewIndex(b *testing.B) {
	concurrencies := []int{
		25, 50, 75, 100, 150, 200, 250, 300, 400, 500, 750, 1000, 1250,
	}
	bufSizes := []int{
		50, 100, 250, 500, 750, 100, 1250,
	}
	for _, concurrency := range concurrencies {
		for _, bufSize := range bufSizes {
			name := fmt.Sprintf("Concurrency%dBufSize%d", concurrency, bufSize)
			b.Run(name, func(b *testing.B) {
				for n := 0; n < b.N; n++ {
					index = NewIndex(urlFile, stopwordFile, bufSize, concurrency)
				}
			})
		}
	}
}
