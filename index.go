package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
)

type Index map[string][]Result

type Result struct {
	College string
	Count   int
}

func (i Index) Search(word string) []Result {
	return i[strings.ToLower(word)]
}

func NewIndex(urlFile, stopwordFile string, bufSize, concurrency int, cache string) Index {
	if cache != "" {
		if err := os.MkdirAll(cache, 0755); err != nil {
			log.Fatal(err)
		}
	}

	urls := urlsFromFile(urlFile, bufSize)
	bodies := fetchBodies(urls, concurrency, cache)
	stopwords := stopwordsFromFile(stopwordFile)
	index := indexReviews(bodies, stopwords)
	return index
}

func urlsFromFile(urlFile string, bufSize int) <-chan string {
	urls := make(chan string, bufSize)

	go func() {
		defer close(urls)

		file, err := os.Open(urlFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			urls <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	return urls
}

func stopwordsFromFile(stopwordFile string) map[string]struct{} {
	stopwords := map[string]struct{}{}

	file, err := os.Open(stopwordFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		stopwords[scanner.Text()] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return stopwords
}

func fetchBodies(urls <-chan string, concurrency int, cache string) <-chan string {
	http.DefaultTransport.(*http.Transport).MaxIdleConns = concurrency
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = concurrency

	bodies := make(chan string)
	wg := sync.WaitGroup{}
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for url := range urls {
				bodies <- fetchBody(url, cache)
			}
		}()
	}
	go func() {
		defer close(bodies)
		wg.Wait()
	}()
	return bodies
}

func fetchBody(url, cache string) string {
	file := fmt.Sprintf("%s/%s", cache, path.Base(url))
	if cache != "" {
		if body, err := ioutil.ReadFile(file); err == nil {
			verbosePrintf("Using cached file: %s\n", file)
			return string(body)
		}
	}

	verbosePrintf("Downloading file: %s\n", file)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Received bad status code: %d. Request body:\n%s", resp.StatusCode, body)
	}

	if cache != "" {
		verbosePrintf("Saving to cache: %s\n", file)
		if err := ioutil.WriteFile(file, body, 0755); err != nil {
			log.Fatal(err)
		}
	}

	return string(body)
}

func indexReviews(bodies <-chan string, stopwords map[string]struct{}) map[string][]Result {
	index := map[string]map[string]int{}
	for body := range bodies {
		lines := strings.Split(body, "\r\n")
		college := lines[0]

		for _, review := range lines[1:] {
			words := strings.FieldsFunc(review, func(r rune) bool {
				switch r {
				case ' ', '.', ',', '!', '?', ';', ':', '(', ')', '"', '\'', '\t', '\n', '\r':
					return true
				}
				return false
			})

			wordMap := map[string]struct{}{}
			for _, word := range words {
				word = strings.ToLower(word)
				if word == "" {
					continue
				} else if _, exists := stopwords[word]; exists {
					continue
				} else if _, exists := wordMap[word]; !exists {
					wordMap[word] = struct{}{}
					if _, exists := index[word]; !exists {
						index[word] = map[string]int{}
					}
					index[word][college] += 1
				}
			}
		}
	}

	sortedIndex := map[string][]Result{}
	for word, colleges := range index {
		results := []Result{}
		for college, count := range colleges {
			results = append(results, Result{
				College: college,
				Count:   count,
			})
		}

		sort.Slice(results, func(i, j int) bool {
			if results[i].Count == results[j].Count {
				return results[i].College < results[j].College
			}
			return results[i].Count > results[j].Count
		})
		sortedIndex[word] = results
	}

	return sortedIndex
}
