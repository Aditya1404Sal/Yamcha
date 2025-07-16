package metrics

import (
	"sort"
	"sync"
	"time"
)

// Result represents the result of a single HTTP request
type Result struct {
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Elapsed      time.Duration `json:"elapsed"`
	StatusCode   int           `json:"status_code"`
	Status       string        `json:"status"`
	BytesRead    int64         `json:"bytes_read"`
	BytesWritten int64         `json:"bytes_written"`
	Error        error         `json:"error,omitempty"`
}

// Metrics holds comprehensive metrics for load test results
type Metrics struct {
	mu               sync.RWMutex
	results          []Result
	startTime        time.Time
	endTime          time.Time
	totalRequests    int
	successfulReqs   int
	failedReqs       int
	totalBytes       int64
	statusCodeCounts map[int]int
	errorCounts      map[string]int
}

// Statistics holds calculated statistics
type Statistics struct {
	TotalRequests    int             `json:"total_requests"`
	SuccessfulReqs   int             `json:"successful_requests"`
	FailedReqs       int             `json:"failed_requests"`
	SuccessRate      float64         `json:"success_rate"`
	ErrorRate        float64         `json:"error_rate"`
	Duration         time.Duration   `json:"duration"`
	RequestsPerSec   float64         `json:"requests_per_second"`
	BytesPerSec      float64         `json:"bytes_per_second"`
	AvgResponseTime  time.Duration   `json:"avg_response_time"`
	MinResponseTime  time.Duration   `json:"min_response_time"`
	MaxResponseTime  time.Duration   `json:"max_response_time"`
	MedianTime       time.Duration   `json:"median_response_time"`
	P95ResponseTime  time.Duration   `json:"p95_response_time"`
	P99ResponseTime  time.Duration   `json:"p99_response_time"`
	StatusCodeCounts map[int]int     `json:"status_code_counts"`
	ErrorCounts      map[string]int  `json:"error_counts"`
	ResponseTimes    []time.Duration `json:"response_times"`
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		results:          make([]Result, 0),
		statusCodeCounts: make(map[int]int),
		errorCounts:      make(map[string]int),
		startTime:        time.Now(),
	}
}

// AddResult adds a result to the metrics collector
func (m *Metrics) AddResult(result Result) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.results = append(m.results, result)
	m.totalRequests++
	m.totalBytes += result.BytesRead

	if result.Error != nil {
		m.failedReqs++
		errorMsg := result.Error.Error()
		m.errorCounts[errorMsg]++
	} else {
		m.successfulReqs++
		m.statusCodeCounts[result.StatusCode]++
	}
}

// SetEndTime sets the end time for the test
func (m *Metrics) SetEndTime(endTime time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.endTime = endTime
}

// GetStatistics calculates and returns comprehensive statistics
func (m *Metrics) GetStatistics() Statistics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.results) == 0 {
		return Statistics{}
	}

	duration := m.endTime.Sub(m.startTime)
	if duration == 0 {
		duration = time.Since(m.startTime)
	}

	// Calculate response time statistics
	responseTimes := make([]time.Duration, len(m.results))
	var totalResponseTime time.Duration
	var minTime, maxTime time.Duration

	for i, result := range m.results {
		responseTimes[i] = result.Elapsed
		totalResponseTime += result.Elapsed

		if i == 0 || result.Elapsed < minTime {
			minTime = result.Elapsed
		}
		if result.Elapsed > maxTime {
			maxTime = result.Elapsed
		}
	}

	// Sort for percentile calculations
	sort.Slice(responseTimes, func(i, j int) bool {
		return responseTimes[i] < responseTimes[j]
	})

	var avgResponseTime time.Duration
	if len(responseTimes) > 0 {
		avgResponseTime = totalResponseTime / time.Duration(len(responseTimes))
	}

	return Statistics{
		TotalRequests:    m.totalRequests,
		SuccessfulReqs:   m.successfulReqs,
		FailedReqs:       m.failedReqs,
		SuccessRate:      float64(m.successfulReqs) / float64(m.totalRequests) * 100,
		ErrorRate:        float64(m.failedReqs) / float64(m.totalRequests) * 100,
		Duration:         duration,
		RequestsPerSec:   float64(m.totalRequests) / duration.Seconds(),
		BytesPerSec:      float64(m.totalBytes) / duration.Seconds(),
		AvgResponseTime:  avgResponseTime,
		MinResponseTime:  minTime,
		MaxResponseTime:  maxTime,
		MedianTime:       percentile(responseTimes, 50),
		P95ResponseTime:  percentile(responseTimes, 95),
		P99ResponseTime:  percentile(responseTimes, 99),
		StatusCodeCounts: copyMap(m.statusCodeCounts),
		ErrorCounts:      copyStringMap(m.errorCounts),
		ResponseTimes:    responseTimes,
	}
}

// GetResults returns a copy of all results
func (m *Metrics) GetResults() []Result {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]Result, len(m.results))
	copy(results, m.results)
	return results
}

// GetRealtimeStats returns current statistics without locking for too long
func (m *Metrics) GetRealtimeStats() Statistics {
	return m.GetStatistics()
}

// percentile calculates the given percentile from sorted response times
func percentile(sortedTimes []time.Duration, p int) time.Duration {
	if len(sortedTimes) == 0 {
		return 0
	}

	if p <= 0 {
		return sortedTimes[0]
	}
	if p >= 100 {
		return sortedTimes[len(sortedTimes)-1]
	}

	index := float64(p) / 100.0 * float64(len(sortedTimes)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sortedTimes) {
		return sortedTimes[len(sortedTimes)-1]
	}

	weight := index - float64(lower)
	return time.Duration(float64(sortedTimes[lower]) + weight*float64(sortedTimes[upper]-sortedTimes[lower]))
}

// copyMap creates a copy of an int map
func copyMap(original map[int]int) map[int]int {
	copy := make(map[int]int)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// copyStringMap creates a copy of a string map
func copyStringMap(original map[string]int) map[string]int {
	copy := make(map[string]int)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// RealTimeCollector collects metrics in real-time
type RealTimeCollector struct {
	metrics        *Metrics
	updateChan     chan Result
	statsChan      chan Statistics
	stopChan       chan struct{}
	updateInterval time.Duration
}

// NewRealTimeCollector creates a new real-time metrics collector
func NewRealTimeCollector(updateInterval time.Duration) *RealTimeCollector {
	return &RealTimeCollector{
		metrics:        NewMetrics(),
		updateChan:     make(chan Result, 1000), // Buffered channel
		statsChan:      make(chan Statistics, 10),
		stopChan:       make(chan struct{}),
		updateInterval: updateInterval,
	}
}

// Start starts the real-time collector
func (rtc *RealTimeCollector) Start() {
	go rtc.run()
}

// Stop stops the real-time collector
func (rtc *RealTimeCollector) Stop() {
	close(rtc.stopChan)
}

// AddResult adds a result to the real-time collector
func (rtc *RealTimeCollector) AddResult(result Result) {
	select {
	case rtc.updateChan <- result:
	default:
		// Channel is full, handle gracefully
		rtc.metrics.AddResult(result)
	}
}

// GetStatsChan returns the channel for receiving real-time statistics
func (rtc *RealTimeCollector) GetStatsChan() <-chan Statistics {
	return rtc.statsChan
}

// GetMetrics returns the underlying metrics collector
func (rtc *RealTimeCollector) GetMetrics() *Metrics {
	return rtc.metrics
}

// run is the main loop for the real-time collector
func (rtc *RealTimeCollector) run() {
	ticker := time.NewTicker(rtc.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case result := <-rtc.updateChan:
			rtc.metrics.AddResult(result)
		case <-ticker.C:
			stats := rtc.metrics.GetRealtimeStats()
			select {
			case rtc.statsChan <- stats:
			default:
				// Channel is full, skip this update
			}
		case <-rtc.stopChan:
			// Process remaining results
			for {
				select {
				case result := <-rtc.updateChan:
					rtc.metrics.AddResult(result)
				default:
					return
				}
			}
		}
	}
}
