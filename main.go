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
// File-content gets unmarshalled into this struct
type RequestPayload struct {
	Headers map[string]string `json:"headers"`
	Body    map[string]string `json:"body"`
}

func main() {
	url := flag.String("url", "http://localhost:8080", "Site where you want to attack")
	numReq := flag.Int("requests", 101, "Number of requests to send")
	attacktype := flag.String("attack", "steady", "Type of attack")
	plot := flag.Bool("plot", true, "Do you want to plot the test as a timeseries?")
	numCPUS := flag.Int("cpus", runtime.NumCPU(), "Number of CPUs to use")
	method := flag.String("method", "GET", "HTTP method to use (GET, POST, etc.)")
	rate := flag.Int("rate", 25, "Number of requests per second")
	burst := flag.Int("burst", 5, "Number of bursts for burst load attack")
	stepSize := flag.Int("stepsize", 10, "Step size for ramp-up load")
	spikeInterval := flag.Int("spikeInterval", 10, "Spike interval")
	duration := flag.Duration("duration", 10*time.Second, "Duration for sustained load tests")
	bodyFile := flag.String("body", "", "The request body file in JSON format")
	activeConn := flag.Bool("conn", false, "Number of Active connections the goroutines stay alive for")
	// Parsing the flags
	flag.Parse()

	runtime.GOMAXPROCS(*numCPUS)
	//Initializing empty payload for casting
	var payload RequestPayload
	var bodyContent []byte
	var err error

	if *bodyFile != "" {
		// If body flag is used, check if a file is specified
		if _, err = os.Stat(*bodyFile); os.IsNotExist(err) {
			// If file doesn't exist, use default payload
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
			// Reading and parsing of the JSON file
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
		// If body flag is not used, empty payload is used
		payload = RequestPayload{
			Headers: map[string]string{},
			Body:    map[string]string{},
		}
		bodyContent = nil
	}

	fmt.Println("Request headers and body content:")
	fmt.Println("Headers:", payload.Headers)
	fmt.Println("Body:", string(bodyContent))

	results := make([]Result, *numReq)

	switch strings.ToLower(*attacktype) {
	case "steady":
		results = basicAttack(*url, *numReq, *rate, *method, payload.Headers, string(bodyContent), *activeConn)
	case "random":
		results = randomLoad(*url, *numReq, *rate, *method, payload.Headers, string(bodyContent), *activeConn)
	case "burst":
		results = burstLoad(*url, *numReq, *rate, *method, *burst, payload.Headers, string(bodyContent), *activeConn)
	case "rampup":
		results = rampUpLoad(*url, *numReq, *rate, *method, *stepSize, payload.Headers, string(bodyContent), *activeConn)
	case "spike":
		results = spikeLoad(*url, *numReq, *rate, *method, *spikeInterval, payload.Headers, string(bodyContent), *activeConn)
	case "sustained":
		results = sustainedLoad(*url, *numReq, *rate, *method, *duration, payload.Headers, string(bodyContent), *activeConn)
	default:
		fmt.Println("Unknown attack type:", *attacktype)
		return
	}
	displayMetrics(results)

	if *plot {
		plotResults(results)
	}
}
