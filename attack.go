package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Status  string
	Elapsed time.Duration
	Error   error
}

func makeRequest(url string, method string, headers map[string]string, body string, wg *sync.WaitGroup, results chan<- Result, activeconn bool) {
	defer wg.Done()
	start := time.Now()

	var req *http.Request
	var err error

	if body == "" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
	}

	if err != nil {
		results <- Result{Error: err}
		return
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	transport := &http.Transport{}
	if activeconn {
		req.Header.Set("Connection", "keep-alive")
		transport.MaxIdleConns = 100
		transport.MaxIdleConnsPerHost = 100
		transport.IdleConnTimeout = 45 * time.Second
	}

	client := &http.Client{Transport: transport}
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

func basicAttack(url string, numRequests int, rate int, method string, headers map[string]string, body string, activeconn bool) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)
	ticker := time.NewTicker(time.Second / time.Duration(rate))
	defer ticker.Stop()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		<-ticker.C
		go makeRequest(url, method, headers, body, &wg, results, activeconn)
	}
	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func burstLoad(url string, numRequests, rate int, method string, bursts int, headers map[string]string, body string, activeconn bool) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)

	for i := 0; i < bursts; i++ {
		for j := 0; j < numRequests; j++ {
			wg.Add(1)
			go makeRequest(url, method, headers, body, &wg, results, activeconn)
		}
		time.Sleep(time.Second / time.Duration(rate))
	}

	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func randomLoad(url string, numRequests, rate int, method string, headers map[string]string, body string, activeconn bool) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go makeRequest(url, method, headers, body, &wg, results, activeconn)
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

func rampUpLoad(url string, numRequests int, rate int, method string, stepSize int, headers map[string]string, body string, activeconn bool) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)

	for i := 1; i <= numRequests; i++ {
		wg.Add(1)
		go makeRequest(url, method, headers, body, &wg, results, activeconn)
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

func spikeLoad(url string, numRequests int, rate int, method string, spikeHeight int, headers map[string]string, body string, activeconn bool) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go makeRequest(url, method, headers, body, &wg, results, activeconn)
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

func sustainedLoad(url string, numRequests int, rate int, method string, duration time.Duration, headers map[string]string, body string, activeconn bool) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, numRequests)
	end := time.Now().Add(duration)

	for time.Now().Before(end) {
		wg.Add(1)
		go makeRequest(url, method, headers, body, &wg, results, activeconn)
		time.Sleep(time.Second / time.Duration(rate))
	}
	wg.Wait()
	close(results)
	resultSlice := make([]Result, 0, numRequests)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}
