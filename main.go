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

	hfs.Parse(flag.Args())

	runtime.GOMAXPROCS(*numCPUS)

	results := make([]Result, *numReq)

	switch strings.ToLower(*attacktype) {
	case "basic":
		results = basicAttack(*url, *numReq, *rate, *method)
	//Note to self : add more cases for different attack types like : Distributed, prolonged etc.
	default:
		fmt.Println("Unknown attack type:", *attacktype)
		return
	}

	if *plot {
		plotResults(results)
	}
}
