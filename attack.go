package main

import (
	"fmt"
	"math/rand"
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

func displayMetrics(results []Result) {
	var totalDuration time.Duration
	var successCount, errorCount int
	for _, result := range results {
		if result.Error != nil {
			fmt.Println("Error:", result.Error)
			errorCount++
		} else {
			fmt.Println("Response Status:", result.Status)
			fmt.Println("Response Time:", result.Elapsed)
			totalDuration += result.Elapsed
			successCount++
		}
	}

	fmt.Printf("\nTotal Requests: %d\n", len(results))
	fmt.Printf("Successful Requests: %d\n", successCount)
	fmt.Printf("Failed Requests: %d\n", errorCount)
	if successCount > 0 {
		fmt.Printf("Average Response Time: %v\n", totalDuration/time.Duration(successCount))
	}
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

	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func burstLoad(url string, numRequests, rate int, method string, bursts int) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)
	// Add burst flag

	for i := 0; i < bursts; i++ {
		for j := 0; j < numRequests; j++ {
			wg.Add(1)
			go makeRequest(url, method, &wg, results)
		}
		time.Sleep(time.Second * time.Duration(rate))
	}

	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func randomLoad(url string, numRequests, rate int, method string) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go makeRequest(url, method, &wg, results)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond / time.Duration(rate))
	}
	wg.Wait()
	close(results)
	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func rampUpLoad(url string, numRequests int, rate int, method string, stepSize int) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)

	for i := 1; i <= numRequests; i++ {
		wg.Add(1)
		go makeRequest(url, method, &wg, results)
		if i%stepSize == 0 {
			time.Sleep(time.Second / time.Duration(rate))
		}
	}
	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func spikeLoad(url string, numRequests int, rate int, method string, spikeHeight int) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go makeRequest(url, method, &wg, results)
		if i%spikeHeight == 0 {
			time.Sleep(time.Second * time.Duration(rand.Intn(20)))
		} else {
			time.Sleep(time.Second / time.Duration(rate))
		}
	}
	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice

}
