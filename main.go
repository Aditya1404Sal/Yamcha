package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

// RequestPayload represents the structure of the JSON file containing headers and body
// File-content gets unmarshalled into this struct
type RequestPayload struct {
	Headers map[string]string `json:"headers"`
	Body    map[string]string `json:"body"`
}

type TestPayLoad struct {
	Url               string
	Req_count         int
	Attack            string
	Cpu_count         int
	Req_method        string
	Rate              int
	Burst_count       int
	Step_size         int
	Spike_Height      int
	Duration          time.Duration
	Req_Body          string
	Active_connection bool
	Req_pkt           RequestPayload
}

func main() {
	url := flag.String("url", "http://localhost:8080", "Site where you want to attack")
	numReq := flag.Int("req", 101, "Number of requests to send")
	attacktype := flag.String("attack", "steady", "Type of attack")
	plot := flag.Bool("plot", true, "Do you want to plot the test as a timeseries?")
	numCPUS := flag.Int("cpu", runtime.NumCPU(), "Number of CPUs to use")
	method := flag.String("method", "GET", "HTTP method to use (GET, POST, etc.)")
	rate := flag.Int("rate", 20, "Number of requests per second")
	burst := flag.Int("burst", 5, "Number of bursts for burst load attack")
	stepSize := flag.Int("ss", 10, "Step size for ramp-up load")
	spikeHeight := flag.Int("sh", 10, "Spike Height")
	duration := flag.Duration("dur", 10*time.Second, "Duration for sustained load tests")
	bodyFile := flag.String("body", "", "The request body file in JSON format")
	activeConn := flag.Bool("conn", false, "Number of Active connections")
	// Parsing the flags
	flag.Parse()

	os.Mkdir("./results", 0755)

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
	testPayload := TestPayLoad{
		Url:               *url,
		Req_count:         *numReq,
		Attack:            *attacktype,
		Rate:              *rate,
		Burst_count:       *burst,
		Step_size:         *stepSize,
		Spike_Height:      *spikeHeight,
		Duration:          *duration,
		Req_Body:          string(bodyContent),
		Req_pkt:           payload,
		Req_method:        *method,
		Cpu_count:         *numCPUS,
		Active_connection: *activeConn,
	}
	if bodyContent != nil {
		fmt.Println("Request headers and body content:")
		fmt.Println("Headers:", payload.Headers)
		fmt.Println("Body:", string(bodyContent))
	}
	fmt.Println("Test Status: ")
	bar := progressbar.Default(int64(*numReq))

	results := make([]Result, *numReq)
	//Track How much time it takes for a test to complete
	startTime := time.Now()

	switch strings.ToLower(*attacktype) {
	case "steady":
		results = basicAttack(testPayload, bar)
	case "random":
		results = randomLoad(testPayload, bar)
	case "burst":
		results = burstLoad(testPayload)
	case "rampup":
		results = rampUpLoad(testPayload, bar)
	case "spike":
		results = spikeLoad(testPayload, bar)
	// case "sustained":
	// 	results = sustainedLoad(testPayload)
	default:
		fmt.Println("Unknown attack type:", *attacktype)
		return
	}
	displayMetrics(results)
	// Display How much time it took to complete this Test
	elapsed := time.Since(startTime)
	fmt.Println("Test Completed in : ", elapsed)

	if *plot {
		plotResults(results, testPayload)
	}
}
