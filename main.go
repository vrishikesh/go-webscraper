package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	var (
		url         = flag.String("url", "http://example.com/%d", "Url to scrape")
		start       = flag.Int("start", 0, "From where to start")
		length      = flag.Int("length", 10, "No of pages to scrape")
		concurrency = flag.Int("workers", 3, "No of workers needed")
	)
	flag.Parse()

	type Site struct {
		Url string
		Id  int
	}
	tasks := make(chan Site)

	go func() {
		for i := *start; i < (*start + *length); i++ {
			tasks <- Site{Url: fmt.Sprintf(*url, i), Id: i}
		}

		close(tasks)
	}()

	results := make(chan []string)
	var wg sync.WaitGroup
	wg.Add(*concurrency)
	go func() {
		wg.Wait()
		close(results)
	}()

	for i := 0; i < *concurrency; i++ {
		log.Println("Worker:", i)
		go func(i int) {
			defer wg.Done()
			for t := range tasks {
				r, err := fetchRequest(t.Url, t.Id, i)
				if err != nil {
					log.Print(err)
					continue
				}

				results <- r
			}
		}(i)
	}

	for r := range results {
		log.Println(r)
	}

	log.Println(*url, *start, *length, *concurrency)
}

func fetchRequest(url string, id, index int) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not get %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("too many requests")
		}

		return nil, fmt.Errorf("bad response from server: %v", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse page: %v", err)
	}

	r := []string{fmt.Sprintf("Url: %s, Id: %d, Worker: %d", url, id, index)}
	r = append(r, strings.TrimSpace(doc.Find(".name").Text()))

	return r, nil
}
