package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"
)

func main() {
	url := flag.String("url", "http://localhost:8080", "Site where you want to attack")
	numReq := flag.Int("requests", 110, "Number of requests to send")
	attacktype := flag.String("attack", "steady", "type of attack")
	plot := flag.Bool("plot", true, "Do ya wanna plot da test as a timeseries?")
	numCPUS := flag.Int("cpus", runtime.NumCPU(), "Number of CPUs to use")
	method := flag.String("method", "GET", "HTTP method to use (GET, POST, etc.)")
	rate := flag.Int("rate", 30, "Number of requests per second")
	burst := flag.Int("burst", 5, "Number of bursts for burst load attack")
	stepSize := flag.Int("stepsize", 10, "step size for ramp up load")
	spikeInterval := flag.Int("spikeInterval", 10, "spike interval")
	duration := flag.Duration("duration", 10*time.Second, "Duration for sustained load tests")

	// Parse the flags
	flag.Parse()

	runtime.GOMAXPROCS(*numCPUS)

	results := make([]Result, *numReq)

	switch strings.ToLower(*attacktype) {
	case "steady":
		results = basicAttack(*url, *numReq, *rate, *method)
	case "random":
		results = randomLoad(*url, *numReq, *rate, *method)
	case "burst":
		results = burstLoad(*url, *numReq, *rate, *method, *burst)
	case "rampup":
		results = rampUpLoad(*url, *numReq, *rate, *method, *stepSize)
	case "spike":
		results = spikeLoad(*url, *numReq, *rate, *method, *spikeInterval)
	case "sustained":
		results = sustainedLoad(*url, *numReq, *rate, *method, *duration)
	default:
		fmt.Println("Unknown attack type:", *attacktype)
		return
	}
	displayMetrics(results)

	if *plot {
		plotResults(results)
	}
}
