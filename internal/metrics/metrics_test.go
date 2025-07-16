package metrics

import (
	"fmt"
	"testing"
	"time"
)

func TestMetricsAddResult(t *testing.T) {
	m := NewMetrics()

	result1 := Result{
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(100 * time.Millisecond),
		Elapsed:    100 * time.Millisecond,
		StatusCode: 200,
		Status:     "200 OK",
		BytesRead:  1024,
	}

	result2 := Result{
		StartTime: time.Now(),
		EndTime:   time.Now().Add(50 * time.Millisecond),
		Elapsed:   50 * time.Millisecond,
		Error:     fmt.Errorf("connection error"),
	}

	m.AddResult(result1)
	m.AddResult(result2)

	if m.totalRequests != 2 {
		t.Errorf("Expected total requests to be 2, got %d", m.totalRequests)
	}

	if m.successfulReqs != 1 {
		t.Errorf("Expected successful requests to be 1, got %d", m.successfulReqs)
	}

	if m.failedReqs != 1 {
		t.Errorf("Expected failed requests to be 1, got %d", m.failedReqs)
	}

	if m.statusCodeCounts[200] != 1 {
		t.Errorf("Expected status code 200 count to be 1, got %d", m.statusCodeCounts[200])
	}

	if m.errorCounts["connection error"] != 1 {
		t.Errorf("Expected error count to be 1, got %d", m.errorCounts["connection error"])
	}
}

func TestStatisticsCalculation(t *testing.T) {
	m := NewMetrics()
	m.startTime = time.Now()

	// Add some test results
	results := []Result{
		{Elapsed: 100 * time.Millisecond, StatusCode: 200, Status: "200 OK"},
		{Elapsed: 200 * time.Millisecond, StatusCode: 200, Status: "200 OK"},
		{Elapsed: 150 * time.Millisecond, StatusCode: 404, Status: "404 Not Found"},
		{Elapsed: 300 * time.Millisecond, StatusCode: 500, Status: "500 Internal Server Error"},
	}

	for _, result := range results {
		result.StartTime = time.Now()
		result.EndTime = result.StartTime.Add(result.Elapsed)
		m.AddResult(result)
	}

	m.SetEndTime(time.Now().Add(1 * time.Second))

	stats := m.GetStatistics()

	if stats.TotalRequests != 4 {
		t.Errorf("Expected total requests to be 4, got %d", stats.TotalRequests)
	}

	if stats.SuccessfulReqs != 4 {
		t.Errorf("Expected successful requests to be 4, got %d", stats.SuccessfulReqs)
	}

	expectedAvg := (100 + 200 + 150 + 300) / 4
	actualAvg := stats.AvgResponseTime.Milliseconds()
	if actualAvg != int64(expectedAvg) {
		t.Errorf("Expected average response time to be %dms, got %dms", expectedAvg, actualAvg)
	}

	if stats.MinResponseTime != 100*time.Millisecond {
		t.Errorf("Expected min response time to be 100ms, got %v", stats.MinResponseTime)
	}

	if stats.MaxResponseTime != 300*time.Millisecond {
		t.Errorf("Expected max response time to be 300ms, got %v", stats.MaxResponseTime)
	}

	// Test percentiles
	if stats.MedianTime != 175*time.Millisecond {
		t.Errorf("Expected median to be 175ms, got %v", stats.MedianTime)
	}
}

func TestPercentileCalculation(t *testing.T) {
	times := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		300 * time.Millisecond,
		400 * time.Millisecond,
		500 * time.Millisecond,
	}

	// Test different percentiles
	tests := []struct {
		percentile int
		expected   time.Duration
	}{
		{0, 100 * time.Millisecond},
		{50, 300 * time.Millisecond},
		{95, 480 * time.Millisecond},
		{100, 500 * time.Millisecond},
	}

	for _, test := range tests {
		result := percentile(times, test.percentile)
		if result != test.expected {
			t.Errorf("Expected P%d to be %v, got %v", test.percentile, test.expected, result)
		}
	}
}

func TestRealTimeCollector(t *testing.T) {
	rtc := NewRealTimeCollector(100 * time.Millisecond)
	rtc.Start()

	// Add some results
	result := Result{
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(100 * time.Millisecond),
		Elapsed:    100 * time.Millisecond,
		StatusCode: 200,
		Status:     "200 OK",
	}

	rtc.AddResult(result)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	rtc.Stop()

	metrics := rtc.GetMetrics()
	if metrics.totalRequests != 1 {
		t.Errorf("Expected total requests to be 1, got %d", metrics.totalRequests)
	}
}
