package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	Status  string
	Elapsed time.Duration
	Error   error
}

func makeRequest(url string, method string, wg *sync.WaitGroup, results chan<- Result) {
	defer wg.Done()
	start := time.Now()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		results <- Result{Error: err}
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		results <- Result{Error: err}
		return
	}
	defer resp.Body.Close()
	results <- Result{Status: resp.Status, Elapsed: elapsed}
}

func basicAttack(url string, numRequests int, rate int, method string) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		<-ticker.C
		go makeRequest(url, method, &wg, results)
	}
	wg.Wait()
	close(results)

	var totalDuration time.Duration
	var successCount, errorCount int
	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		if result.Error != nil {
			fmt.Println("Error:", result.Error)
			errorCount++
		} else {
			fmt.Println("Response Status:", result.Status)
			fmt.Println("Response Time:", result.Elapsed)
			totalDuration += result.Elapsed
			successCount++
		}
		resultSlice = append(resultSlice, result)
	}
	fmt.Printf("\n Total Requests; %d\n", numRequests)
	fmt.Printf("Successful Requests: %d\n", successCount)
	fmt.Printf("Failed Requests: %d\n", errorCount)
	fmt.Printf("Average Response Time: %v\n", totalDuration/time.Duration(successCount))
	return resultSlice
}
