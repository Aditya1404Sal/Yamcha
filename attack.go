package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Result struct {
	Status  string
	Elapsed time.Duration
	Error   error
}

func makeRequest(url string, method string, headers map[string]string, body string, wg *sync.WaitGroup, results chan<- Result, activeconn bool, bar *progressbar.ProgressBar) {
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
	if bar != nil {
		bar.Add(1) // Update the progress bar
	}
}

func displayMetrics(results []Result) {
	var totalDuration time.Duration
	var successCount, errorCount int
	for _, result := range results {
		if result.Error != nil {
			fmt.Println("Error:", result.Error)
			errorCount++
		} else {
			// fmt.Println("Response Status:", result.Status)
			// fmt.Println("Response Time:", result.Elapsed)
			totalDuration += result.Elapsed
			successCount++
		}
	}

	fmt.Printf("\nTotal Requests: %d\n", len(results))
	fmt.Printf("Successful Requests: %d | Success Rate: %.2f%%\n", successCount, (float64(successCount)/float64(len(results)))*100)
	fmt.Printf("Failed Requests:       %d | Error Rate:   %.2f%%\n", errorCount, (float64(errorCount)/float64(len(results)))*100)
	if successCount > 0 {
		fmt.Printf("Average Response Time: %v\n", totalDuration/time.Duration(successCount))
	}
}

func basicAttack(tp TestPayLoad, bar *progressbar.ProgressBar) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, tp.Req_count)
	ticker := time.NewTicker(time.Second / time.Duration(tp.Rate))
	defer ticker.Stop()

	for i := 0; i < tp.Req_count; i++ {
		wg.Add(1)
		<-ticker.C
		go makeRequest(tp.Url, tp.Req_method, tp.Req_pkt.Headers, tp.Req_Body, &wg, results, tp.Active_connection, bar)
	}
	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, tp.Req_count)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func burstLoad(tp TestPayLoad) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, tp.Req_count*tp.Burst_count)
	bar := progressbar.Default(int64(tp.Req_count * tp.Burst_count))

	for i := 0; i < tp.Burst_count; i++ {
		for j := 0; j < tp.Req_count; j++ {
			wg.Add(1)
			go makeRequest(tp.Url, tp.Req_method, tp.Req_pkt.Headers, tp.Req_Body, &wg, results, tp.Active_connection, bar)
		}
		time.Sleep(time.Second / time.Duration(tp.Rate))
		wg.Wait()
	}

	close(results)

	resultSlice := make([]Result, 0, tp.Req_count)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func randomLoad(tp TestPayLoad, bar *progressbar.ProgressBar) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, tp.Req_count)

	for i := 0; i < tp.Req_count; i++ {
		wg.Add(1)
		go makeRequest(tp.Url, tp.Req_method, tp.Req_pkt.Headers, tp.Req_Body, &wg, results, tp.Active_connection, bar)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond / time.Duration(tp.Rate))
	}
	wg.Wait()
	close(results)
	resultSlice := make([]Result, 0, tp.Req_count)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func rampUpLoad(tp TestPayLoad, bar *progressbar.ProgressBar) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, tp.Req_count)

	for i := 1; i <= tp.Req_count; i++ {
		wg.Add(1)
		go makeRequest(tp.Url, tp.Req_method, tp.Req_pkt.Headers, tp.Req_Body, &wg, results, tp.Active_connection, bar)
		if i%tp.Step_size == 0 {
			time.Sleep(time.Second / time.Duration(tp.Rate))
		}
	}
	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, tp.Req_count)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

func spikeLoad(tp TestPayLoad, bar *progressbar.ProgressBar) []Result {
	var wg sync.WaitGroup
	results := make(chan Result, tp.Req_count)

	for i := 0; i < tp.Req_count; i++ {
		wg.Add(1)
		go makeRequest(tp.Url, tp.Req_method, tp.Req_pkt.Headers, tp.Req_Body, &wg, results, tp.Active_connection, bar)
		if i%tp.Spike_Height == 0 {
			time.Sleep(time.Second * time.Duration(rand.Intn(20)))
		} else {
			time.Sleep(time.Second / time.Duration(tp.Rate))
		}
	}
	wg.Wait()
	close(results)

	resultSlice := make([]Result, 0, tp.Req_count)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}
	return resultSlice
}

// In Progress
// func sustainedLoad(tp TestPayLoad) []Result {
// 	var wg sync.WaitGroup
// 	results := make(chan Result, tp.Req_count)
// 	end := time.Now().Add(tp.Duration)

// 	for time.Now().Before(end) {
// 		wg.Add(1)
// 		go makeRequest(tp.Url, tp.Req_method, tp.Req_pkt.Headers, tp.Req_Body, &wg, results, tp.Active_connection)
// 		time.Sleep(time.Second / time.Duration(tp.Rate))
// 	}
// 	wg.Wait()
// 	close(results)
// 	resultSlice := make([]Result, 0, tp.Req_count)
// 	for result := range results {
// 		resultSlice = append(resultSlice, result)
// 	}
// 	return resultSlice
// }
