package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
)

func main() {
	hfs := flag.NewFlagSet("yamcha", flag.ExitOnError)

	url := hfs.String("url", " ", "Site where you want to attack")
	numReq := hfs.Int("requests", 10, "Number of requests to send")
	attacktype := hfs.String("attack", "basic", "type of attack")
	plot := hfs.Bool("plot", false, "Do ya wanna plot da test as a timeseries ?")
	numCPUS := hfs.Int("cpus", runtime.NumCPU(), "Number of CPUs to use")
	method := hfs.String("method", "GET", "HTTP method to use (GET, POST, etc.)")
	rate := hfs.Int("rate", 1, "Number of requests per second")
	burst := hfs.Int("burst", 5, "Number of bursts for burst load attack")
	stepSize := hfs.Int("stepsize", 10, "step size for ramp up load")
	spikeInterval := hfs.Int("spikeInterval", 10, "spike interval")
	hfs.Parse(flag.Args())

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
	default:
		fmt.Println("Unknown attack type:", *attacktype)
		return
	}
	displayMetrics(results)

	if *plot {
		plotResults(results)
	}
}
