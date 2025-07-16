package reporter

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Yamcha/internal/config"
	"Yamcha/internal/metrics"
)

// Reporter interface for different output formats
type Reporter interface {
	Generate(stats metrics.Statistics, cfg *config.Config, results []metrics.Result) error
	Extension() string
}

// ConsoleReporter outputs results to console
type ConsoleReporter struct{}

// NewConsoleReporter creates a new console reporter
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

// Extension returns the file extension
func (cr *ConsoleReporter) Extension() string {
	return ""
}

// Generate outputs results to console
func (cr *ConsoleReporter) Generate(stats metrics.Statistics, cfg *config.Config, results []metrics.Result) error {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("                       YAMCHA LOAD TEST RESULTS")
	fmt.Println(strings.Repeat("=", 80))

	fmt.Printf("Target URL:           %s\n", cfg.Target.URL)
	fmt.Printf("Attack Type:          %s\n", cfg.Load.AttackType)
	fmt.Printf("HTTP Method:          %s\n", cfg.Target.Method)
	fmt.Printf("Test Duration:        %v\n", stats.Duration)

	fmt.Println("\n" + strings.Repeat("-", 40) + " SUMMARY " + strings.Repeat("-", 40))
	fmt.Printf("Total Requests:       %d\n", stats.TotalRequests)
	fmt.Printf("Successful Requests:  %d (%.2f%%)\n", stats.SuccessfulReqs, stats.SuccessRate)
	fmt.Printf("Failed Requests:      %d (%.2f%%)\n", stats.FailedReqs, stats.ErrorRate)
	fmt.Printf("Requests/sec:         %.2f\n", stats.RequestsPerSec)
	fmt.Printf("Data Transferred:     %.2f KB/sec\n", stats.BytesPerSec/1024)

	fmt.Println("\n" + strings.Repeat("-", 35) + " RESPONSE TIMES " + strings.Repeat("-", 35))
	fmt.Printf("Average:              %v\n", stats.AvgResponseTime)
	fmt.Printf("Minimum:              %v\n", stats.MinResponseTime)
	fmt.Printf("Maximum:              %v\n", stats.MaxResponseTime)
	fmt.Printf("Median (P50):         %v\n", stats.MedianTime)
	fmt.Printf("95th Percentile:      %v\n", stats.P95ResponseTime)
	fmt.Printf("99th Percentile:      %v\n", stats.P99ResponseTime)

	if len(stats.StatusCodeCounts) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 35) + " STATUS CODES " + strings.Repeat("-", 37))
		for code, count := range stats.StatusCodeCounts {
			percentage := float64(count) / float64(stats.TotalRequests) * 100
			fmt.Printf("  %d:                  %d (%.2f%%)\n", code, count, percentage)
		}
	}

	if len(stats.ErrorCounts) > 0 {
		fmt.Println("\n" + strings.Repeat("-", 38) + " ERRORS " + strings.Repeat("-", 38))
		for errorMsg, count := range stats.ErrorCounts {
			percentage := float64(count) / float64(stats.TotalRequests) * 100
			fmt.Printf("  %s: %d (%.2f%%)\n", errorMsg, count, percentage)
		}
	}

	fmt.Println(strings.Repeat("=", 80))
	return nil
}

// JSONReporter outputs results to JSON format
type JSONReporter struct {
	filename string
}

// NewJSONReporter creates a new JSON reporter
func NewJSONReporter(filename string) *JSONReporter {
	return &JSONReporter{filename: filename}
}

// Extension returns the file extension
func (jr *JSONReporter) Extension() string {
	return "json"
}

// Generate outputs results to JSON file
func (jr *JSONReporter) Generate(stats metrics.Statistics, cfg *config.Config, results []metrics.Result) error {
	report := map[string]interface{}{
		"timestamp":   time.Now().UTC(),
		"config":      cfg,
		"statistics":  stats,
		"raw_results": results,
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(jr.filename, data, 0644)
}

// HTMLReporter outputs results to HTML format with charts
type HTMLReporter struct {
	filename string
}

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter(filename string) *HTMLReporter {
	return &HTMLReporter{filename: filename}
}

// Extension returns the file extension
func (hr *HTMLReporter) Extension() string {
	return "html"
}

// Generate outputs results to HTML file
func (hr *HTMLReporter) Generate(stats metrics.Statistics, cfg *config.Config, results []metrics.Result) error {
	file, err := os.Create(hr.filename)
	if err != nil {
		return fmt.Errorf("failed to create HTML file: %w", err)
	}
	defer file.Close()

	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	templateData := hr.prepareTemplateData(stats, cfg, results)

	if err := tmpl.Execute(file, templateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// prepareTemplateData prepares data for HTML template
func (hr *HTMLReporter) prepareTemplateData(stats metrics.Statistics, cfg *config.Config, results []metrics.Result) map[string]interface{} {
	labels := make([]int, len(results))
	responseData := make([]int64, len(results))
	statusCodes := make([]int, len(results))
	timestamps := make([]string, len(results))

	for i, result := range results {
		labels[i] = i + 1
		responseData[i] = result.Elapsed.Milliseconds()
		statusCodes[i] = result.StatusCode
		timestamps[i] = result.StartTime.Format("15:04:05.000")
	}

	return map[string]interface{}{
		"Config":       cfg,
		"Statistics":   stats,
		"Labels":       labels,
		"ResponseData": responseData,
		"StatusCodes":  statusCodes,
		"Timestamps":   timestamps,
		"GeneratedAt":  time.Now().Format("2006-01-02 15:04:05"),
	}
}

// CSVReporter outputs results to CSV format
type CSVReporter struct {
	filename string
}

// NewCSVReporter creates a new CSV reporter
func NewCSVReporter(filename string) *CSVReporter {
	return &CSVReporter{filename: filename}
}

// Extension returns the file extension
func (cr *CSVReporter) Extension() string {
	return "csv"
}

// Generate outputs results to CSV file
func (cr *CSVReporter) Generate(stats metrics.Statistics, cfg *config.Config, results []metrics.Result) error {
	file, err := os.Create(cr.filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Write header
	header := "timestamp,elapsed_ms,status_code,status,bytes_read,bytes_written,error\n"
	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data
	for _, result := range results {
		errorStr := ""
		if result.Error != nil {
			errorStr = strings.ReplaceAll(result.Error.Error(), ",", ";")
		}

		line := fmt.Sprintf("%s,%d,%d,%s,%d,%d,%s\n",
			result.StartTime.Format("2006-01-02T15:04:05.000Z"),
			result.Elapsed.Milliseconds(),
			result.StatusCode,
			strings.ReplaceAll(result.Status, ",", ";"),
			result.BytesRead,
			result.BytesWritten,
			errorStr,
		)

		if _, err := file.WriteString(line); err != nil {
			return fmt.Errorf("failed to write CSV line: %w", err)
		}
	}

	return nil
}

// ReporterManager manages multiple reporters
type ReporterManager struct {
	reporters []Reporter
	outputDir string
}

// NewReporterManager creates a new reporter manager
func NewReporterManager(cfg *config.Config) (*ReporterManager, error) {
	if err := os.MkdirAll(cfg.Output.Directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	rm := &ReporterManager{
		reporters: make([]Reporter, 0),
		outputDir: cfg.Output.Directory,
	}

	// Generate filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	baseFilename := fmt.Sprintf("yamcha-results-%s", timestamp)
	if cfg.Output.Filename != "" {
		baseFilename = cfg.Output.Filename
	}

	// Add console reporter (always enabled)
	rm.reporters = append(rm.reporters, NewConsoleReporter())

	// Add file reporters based on configuration
	for _, format := range cfg.Output.Formats {
		switch format {
		case "json":
			filename := filepath.Join(cfg.Output.Directory, baseFilename+".json")
			rm.reporters = append(rm.reporters, NewJSONReporter(filename))
		case "html":
			filename := filepath.Join(cfg.Output.Directory, baseFilename+".html")
			rm.reporters = append(rm.reporters, NewHTMLReporter(filename))
		case "csv":
			filename := filepath.Join(cfg.Output.Directory, baseFilename+".csv")
			rm.reporters = append(rm.reporters, NewCSVReporter(filename))
		}
	}

	return rm, nil
}

// GenerateReports generates all configured reports
func (rm *ReporterManager) GenerateReports(stats metrics.Statistics, cfg *config.Config, results []metrics.Result) error {
	for _, reporter := range rm.reporters {
		if err := reporter.Generate(stats, cfg, results); err != nil {
			fmt.Printf("Warning: Failed to generate %s report: %v\n", reporter.Extension(), err)
		} else if reporter.Extension() != "" {
			fmt.Printf("Report generated: %s\n", reporter.Extension())
		}
	}
	return nil
}

// RealTimeReporter provides real-time statistics updates
type RealTimeReporter struct {
	updateInterval time.Duration
	lastStats      metrics.Statistics
}

// NewRealTimeReporter creates a new real-time reporter
func NewRealTimeReporter(updateInterval time.Duration) *RealTimeReporter {
	return &RealTimeReporter{
		updateInterval: updateInterval,
	}
}

// Start starts real-time reporting
func (rtr *RealTimeReporter) Start(statsChan <-chan metrics.Statistics, stopChan <-chan struct{}) {
	ticker := time.NewTicker(rtr.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case stats := <-statsChan:
			rtr.displayRealtimeStats(stats)
			rtr.lastStats = stats
		case <-ticker.C:
			// Could display periodic updates here
		case <-stopChan:
			return
		}
	}
}

// displayRealtimeStats displays current statistics
func (rtr *RealTimeReporter) displayRealtimeStats(stats metrics.Statistics) {
	// Clear previous lines (simple implementation)
	fmt.Printf("\r\033[K")
	fmt.Printf("Live Stats - Requests: %d | Success: %.1f%% | RPS: %.1f | Avg: %v",
		stats.TotalRequests,
		stats.SuccessRate,
		stats.RequestsPerSec,
		stats.AvgResponseTime,
	)
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Yamcha Load Test Results</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5em;
            font-weight: 300;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
        }
        .content {
            display: grid;
            grid-template-columns: 1fr 2fr;
            gap: 30px;
            padding: 30px;
        }
        .stats-section {
            display: flex;
            flex-direction: column;
            gap: 20px;
        }
        .stat-card {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            border-left: 4px solid #667eea;
        }
        .stat-card h3 {
            margin: 0 0 15px 0;
            color: #333;
            font-size: 1.1em;
        }
        .stat-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 10px;
        }
        .stat-item {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #eee;
        }
        .stat-label {
            color: #666;
            font-weight: 500;
        }
        .stat-value {
            font-weight: bold;
            color: #333;
        }
        .chart-section {
            background: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
        }
        .chart-container {
            position: relative;
            height: 400px;
            margin-bottom: 30px;
        }
        .success { color: #28a745; }
        .warning { color: #ffc107; }
        .error { color: #dc3545; }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            border-top: 1px solid #eee;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Yamcha Load Test Results</h1>
            <p>Generated on {{.GeneratedAt}}</p>
        </div>
        
        <div class="content">
            <div class="stats-section">
                <div class="stat-card">
                    <h3>Test Configuration</h3>
                    <div class="stat-item">
                        <span class="stat-label">Target URL:</span>
                        <span class="stat-value">{{.Config.Target.URL}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Attack Type:</span>
                        <span class="stat-value">{{.Config.Load.AttackType}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">HTTP Method:</span>
                        <span class="stat-value">{{.Config.Target.Method}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Rate (req/s):</span>
                        <span class="stat-value">{{.Config.Load.Rate}}</span>
                    </div>
                </div>

                <div class="stat-card">
                    <h3>Summary</h3>
                    <div class="stat-item">
                        <span class="stat-label">Total Requests:</span>
                        <span class="stat-value">{{.Statistics.TotalRequests}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Successful:</span>
                        <span class="stat-value success">{{.Statistics.SuccessfulReqs}} ({{printf "%.2f" .Statistics.SuccessRate}}%)</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Failed:</span>
                        <span class="stat-value error">{{.Statistics.FailedReqs}} ({{printf "%.2f" .Statistics.ErrorRate}}%)</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Duration:</span>
                        <span class="stat-value">{{.Statistics.Duration}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Throughput:</span>
                        <span class="stat-value">{{printf "%.2f" .Statistics.RequestsPerSec}} req/s</span>
                    </div>
                </div>

                <div class="stat-card">
                    <h3>Response Times</h3>
                    <div class="stat-item">
                        <span class="stat-label">Average:</span>
                        <span class="stat-value">{{.Statistics.AvgResponseTime}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Minimum:</span>
                        <span class="stat-value">{{.Statistics.MinResponseTime}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Maximum:</span>
                        <span class="stat-value">{{.Statistics.MaxResponseTime}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">Median (P50):</span>
                        <span class="stat-value">{{.Statistics.MedianTime}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">95th Percentile:</span>
                        <span class="stat-value">{{.Statistics.P95ResponseTime}}</span>
                    </div>
                    <div class="stat-item">
                        <span class="stat-label">99th Percentile:</span>
                        <span class="stat-value">{{.Statistics.P99ResponseTime}}</span>
                    </div>
                </div>

                {{if .Statistics.StatusCodeCounts}}
                <div class="stat-card">
                    <h3>Status Code Distribution</h3>
                    {{range $code, $count := .Statistics.StatusCodeCounts}}
                    <div class="stat-item">
                        <span class="stat-label">{{$code}}:</span>
                        <span class="stat-value">{{$count}}</span>
                    </div>
                    {{end}}
                </div>
                {{end}}
            </div>

            <div class="chart-section">
                <h3>Response Time Chart</h3>
                <div class="chart-container">
                    <canvas id="responseChart"></canvas>
                </div>
            </div>
        </div>

        <div class="footer">
            <p>Generated by Yamcha Load Testing Tool v2.0</p>
        </div>
    </div>

    <script>
        const ctx = document.getElementById('responseChart').getContext('2d');
        const statusCodes = {{.StatusCodes}};
        const responseData = {{.ResponseData}};
        
        new Chart(ctx, {
            type: 'line',
            data: {
                labels: {{.Labels}},
                datasets: [{
                    label: 'Response Time (ms)',
                    data: responseData,
                    borderColor: '#667eea',
                    backgroundColor: 'rgba(102, 126, 234, 0.1)',
                    borderWidth: 2,
                    pointBackgroundColor: statusCodes.map(code => {
                        if (code >= 200 && code < 300) return '#28a745';
                        if (code >= 300 && code < 400) return '#ffc107';
                        if (code >= 400 && code < 500) return '#fd7e14';
                        if (code >= 500) return '#dc3545';
                        return '#6c757d';
                    }),
                    pointRadius: 3,
                    pointHoverRadius: 6,
                    tension: 0.1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    y: {
                        beginAtZero: true,
                        title: {
                            display: true,
                            text: 'Response Time (ms)'
                        }
                    },
                    x: {
                        title: {
                            display: true,
                            text: 'Request Number'
                        }
                    }
                },
                plugins: {
                    tooltip: {
                        callbacks: {
                            afterLabel: function(context) {
                                return 'Status: ' + statusCodes[context.dataIndex];
                            }
                        }
                    },
                    legend: {
                        display: true
                    }
                }
            }
        });
    </script>
</body>
</html>
`
