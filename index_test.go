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
	benchmarks := []struct {
		concurrency int
		bufSize     int
	}{
		{
			concurrency: 100,
			bufSize:     1000,
		},
		{
			concurrency: 250,
			bufSize:     1000,
		},
		{
			concurrency: 500,
			bufSize:     1000,
		},
	}
	for _, benchmark := range benchmarks {
		name := fmt.Sprintf("Concurrency%dBufSize%d", benchmark.concurrency, benchmark.bufSize)
		b.Run(name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				index = NewIndex(urlFile, stopwordFile, benchmark.bufSize, benchmark.concurrency)
			}
		})
	}
}
