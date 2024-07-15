package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

// RequestPayload represents the structure of the JSON file containing headers and body
type RequestPayload struct {
	Headers map[string]string `json:"headers"`
	Body    map[string]string `json:"body"`
}

func main() {
	url := flag.String("url", "http://localhost:8080", "Site where you want to attack")
	numReq := flag.Int("requests", 110, "Number of requests to send")
	attacktype := flag.String("attack", "steady", "Type of attack")
	plot := flag.Bool("plot", true, "Do you want to plot the test as a timeseries?")
	numCPUS := flag.Int("cpus", runtime.NumCPU(), "Number of CPUs to use")
	method := flag.String("method", "GET", "HTTP method to use (GET, POST, etc.)")
	rate := flag.Int("rate", 30, "Number of requests per second")
	burst := flag.Int("burst", 5, "Number of bursts for burst load attack")
	stepSize := flag.Int("stepsize", 10, "Step size for ramp-up load")
	spikeInterval := flag.Int("spikeInterval", 10, "Spike interval")
	duration := flag.Duration("duration", 10*time.Second, "Duration for sustained load tests")
	bodyFile := flag.String("body", "", "The request body file in JSON format")

	// Parse the flags
	flag.Parse()

	runtime.GOMAXPROCS(*numCPUS)

	var payload RequestPayload
	var bodyContent []byte
	var err error

	if *bodyFile != "" {
		// If body flag is used, check if a file is specified
		if _, err = os.Stat(*bodyFile); os.IsNotExist(err) {
			// File doesn't exist, use default payload
			payload = RequestPayload{
				Headers: map[string]string{
					"Content-Type": "application/json",
					"Session-ID":   "Put your session-ID here",
				},
				Body: map[string]string{
					"message": "Hello world",
				},
			}
		} else {
			// Read and parse the JSON file
			fileContent, err := os.ReadFile(*bodyFile)
			if err != nil {
				fmt.Println("Error reading body file:", err)
				return
			}
			err = json.Unmarshal(fileContent, &payload)
			if err != nil {
				fmt.Println("Error parsing body file:", err)
				return
			}
		}
		// Convert the body to a JSON string
		bodyContent, err = json.Marshal(payload.Body)
		if err != nil {
			fmt.Println("Error encoding body content:", err)
			return
		}
	} else {
		// No body flag, use empty payload
		payload = RequestPayload{
			Headers: map[string]string{},
			Body:    map[string]string{},
		}
		bodyContent = nil
	}

	// Print the content of the payload (for debugging purposes)
	fmt.Println("Request headers and body content:")
	fmt.Println("Headers:", payload.Headers)
	fmt.Println("Body:", string(bodyContent))

	results := make([]Result, *numReq)

	switch strings.ToLower(*attacktype) {
	case "steady":
		results = basicAttack(*url, *numReq, *rate, *method, payload.Headers, string(bodyContent))
	case "random":
		results = randomLoad(*url, *numReq, *rate, *method, payload.Headers, string(bodyContent))
	case "burst":
		results = burstLoad(*url, *numReq, *rate, *method, *burst, payload.Headers, string(bodyContent))
	case "rampup":
		results = rampUpLoad(*url, *numReq, *rate, *method, *stepSize, payload.Headers, string(bodyContent))
	case "spike":
		results = spikeLoad(*url, *numReq, *rate, *method, *spikeInterval, payload.Headers, string(bodyContent))
	case "sustained":
		results = sustainedLoad(*url, *numReq, *rate, *method, *duration, payload.Headers, string(bodyContent))
	default:
		fmt.Println("Unknown attack type:", *attacktype)
		return
	}
	displayMetrics(results)

	if *plot {
		plotResults(results)
	}
}
