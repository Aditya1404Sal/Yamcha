package main

import (
	"fmt"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func plotResults(results []Result) {
	line := charts.NewLine()

	xData := make([]int, len(results))
	for i := range xData {
		xData[i] = i + 1
	}

	yData := make([]opts.LineData, len(results))
	for i, result := range results {
		yData[i] = opts.LineData{Value: result.Elapsed.Milliseconds()}
	}

	line.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Load Test Results",
		Subtitle: "Response Time per Request",
	}))
	line.SetXAxis(xData).AddSeries("Response Time", yData)

	//Chart is saved to an HTML file
	f, err := os.Create("results.html")
	if err != nil {
		fmt.Println("Could not create results.html:", err)
		return
	}
	defer f.Close()

	if err := line.Render(f); err != nil {
		fmt.Println("Could not render chart:", err)
	}

	fmt.Println("Results plotted to results.html")
}
