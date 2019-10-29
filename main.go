package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	urlsFile      = flag.String("urls", "./urls.txt", "Path to urls file")
	stopwordsFile = flag.String("stopwords", "./stopWords.txt", "Path to stopwords file")
	bufSize       = flag.Int("buffer", 1250, "Buffer size")
	concurrency   = flag.Int("concurrency", 250, "Download concurrency")
	cache         = flag.String("cache", "", "Cache directory (cache not used if empty)")
	verbose       = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	verbosePrintf("Indexing...\n")
	start := time.Now()
	index := NewIndex(*urlsFile, *stopwordsFile, *bufSize, *concurrency, *cache)
	verbosePrintf("Elapsed: %s\n", time.Since(start))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Enter a word: ")
		if !scanner.Scan() {
			break
		}
		word := scanner.Text()

		results := index.Search(word)
		fmt.Println("Results:")
		for _, result := range results {
			fmt.Printf("%d - %s\n", result.Count, result.College)
		}
		fmt.Println()
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nGoodbye")
}

func verbosePrintf(format string, a ...interface{}) {
	if *verbose {
		fmt.Printf(format, a...)
	}
}
