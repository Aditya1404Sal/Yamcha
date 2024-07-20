package main

import (
	"fmt"
	"html/template"
	"os"
	"time"
)

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <title>Load Test Results</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
        }
        .container {
            display: flex;
            justify-content: space-between;
        }
        .tables {
            width: 50%;
        }
        .chart-container {
            width: 45%;
        }
        .chart-wrapper {
            width: 100%;
            height: 400px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
        }
        th, td {
            border: 1px solid #ddd;
            padding: 8px;
            text-align: left;
        }
        th {
            background-color: #f2f2f2;
        }
    </style>
</head>
<body>
    <h1>Load Test Results</h1>
    <div class="container">
        <div class="tables">
            <h2>Test Configuration</h2>
            <table>
                <tr><th>Parameter</th><th>Value</th></tr>
                <tr><td>URL</td><td>{{.Url}}</td></tr>
                <tr><td>Attack Type</td><td>{{.Attack}}</td></tr>
                <tr><td>Request Count</td><td>{{.Req_count}}</td></tr>
                <tr><td>CPU Count</td><td>{{.Cpu_count}}</td></tr>
                <tr><td>Request Method</td><td>{{.Req_method}}</td></tr>
                <tr><td>Rate</td><td>{{.Rate}}</td></tr>
                <tr><td>Burst Count</td><td>{{.Burst_count}}</td></tr>
                <tr><td>Step Size</td><td>{{.Step_size}}</td></tr>
                <tr><td>Spike Interval</td><td>{{.Spike_interval}}</td></tr>
                <tr><td>Duration</td><td>{{.Duration}}</td></tr>
            </table>
            <h3>Request Headers</h3>
            <table>
                <tr><th>Key</th><th>Value</th></tr>
                {{range $key, $value := .Body.Headers}}
                <tr><td>{{$key}}</td><td>{{$value}}</td></tr>
                {{end}}
            </table>
            <h3>Request Body</h3>
            <table>
                <tr><th>Key</th><th>Value</th></tr>
                {{range $key, $value := .Body.Body}}
                <tr><td>{{$key}}</td><td>{{$value}}</td></tr>
                {{end}}
            </table>
        </div>
        <div class="chart-container">
            <h2>Response Time Chart</h2>
            <div class="chart-wrapper">
                <canvas id="myChart"></canvas>
            </div>
        </div>
    </div>
    <script>
    var ctx = document.getElementById('myChart').getContext('2d');
    var myChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: {{.Labels}},
            datasets: [{
                label: 'Response Time (ms)',
                data: {{.Data}},
                borderColor: 'rgb(75, 192, 192)',
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
    </script>
</body>
</html>
`

func plotResults(results []Result, test_payload TestPayLoad) {
	currTime := time.Now()
	layout := "2006-01-02_15-04-05"
	formattedTime := currTime.Format(layout)
	filename := fmt.Sprintf("./results/results-%s.html", formattedTime)

	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Could not create results.html:", err)
		return
	}
	defer f.Close()

	tmpl, err := template.New("results").Parse(htmlTemplate)
	if err != nil {
		fmt.Println("Could not parse HTML template:", err)
		return
	}

	labels := make([]int, len(results))
	data := make([]int64, len(results))
	for i, result := range results {
		labels[i] = i + 1
		data[i] = result.Elapsed.Milliseconds()
	}

	templateData := struct {
		TestPayLoad
		Labels []int
		Data   []int64
	}{
		TestPayLoad: test_payload,
		Labels:      labels,
		Data:        data,
	}

	if err := tmpl.Execute(f, templateData); err != nil {
		fmt.Println("Could not execute template:", err)
	}

	fmt.Println("Results plotted to", filename)
}
