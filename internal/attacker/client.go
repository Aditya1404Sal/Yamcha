package attacker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"Yamcha/internal/config"
	"Yamcha/internal/metrics"
)

// HTTPClient is an improved HTTP client with connection pooling
type HTTPClient struct {
	client    *http.Client
	transport *http.Transport
	config    *config.Config
}

// NewHTTPClient creates a new improved HTTP client
func NewHTTPClient(cfg *config.Config) *HTTPClient {
	transport := &http.Transport{
		MaxIdleConns:        cfg.HTTP.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.HTTP.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.HTTP.IdleConnTimeout.ToDuration(),
		DisableKeepAlives:   !cfg.HTTP.KeepAlive,
	}

	if cfg.HTTP.DisableTLSVerify {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.HTTP.Timeout.ToDuration(),
	}

	return &HTTPClient{
		client:    client,
		transport: transport,
		config:    cfg,
	}
}

// MakeRequest performs a single HTTP request and returns the result
func (hc *HTTPClient) MakeRequest(ctx context.Context) metrics.Result {
	start := time.Now()
	result := metrics.Result{
		StartTime: start,
	}

	// Prepare request body
	var bodyReader io.Reader
	var bodyBytes []byte
	if hc.config.Target.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(hc.config.Target.Body)
		if err != nil {
			result.Error = fmt.Errorf("failed to marshal request body: %w", err)
			result.EndTime = time.Now()
			result.Elapsed = result.EndTime.Sub(result.StartTime)
			return result
		}
		bodyReader = bytes.NewReader(bodyBytes)
		result.BytesWritten = int64(len(bodyBytes))
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, hc.config.Target.Method, hc.config.Target.URL, bodyReader)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.EndTime = time.Now()
		result.Elapsed = result.EndTime.Sub(result.StartTime)
		return result
	}

	// Set headers
	for key, value := range hc.config.Target.Headers {
		req.Header.Set(key, value)
	}

	// Set default headers if not present
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "Yamcha-LoadTester/2.0")
	}

	// Perform request
	resp, err := hc.client.Do(req)
	end := time.Now()
	result.EndTime = end
	result.Elapsed = end.Sub(start)

	if err != nil {
		result.Error = fmt.Errorf("request failed: %w", err)
		return result
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response body: %w", err)
		return result
	}

	result.StatusCode = resp.StatusCode
	result.Status = resp.Status
	result.BytesRead = int64(len(body))

	return result
}

// Close closes the HTTP client and cleans up resources
func (hc *HTTPClient) Close() {
	if hc.transport != nil {
		hc.transport.CloseIdleConnections()
	}
}

// Worker represents a worker that can execute HTTP requests
type Worker struct {
	id      int
	client  *HTTPClient
	results chan<- metrics.Result
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewWorker creates a new worker
func NewWorker(id int, client *HTTPClient, results chan<- metrics.Result) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		id:      id,
		client:  client,
		results: results,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the worker (not needed for simplified approach)
func (w *Worker) Start() {
	// Not used in simplified approach
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.cancel()
}

// ExecuteRequest executes a single request
func (w *Worker) ExecuteRequest() {
	if w.ctx.Err() != nil {
		return // Worker is stopped
	}

	result := w.client.MakeRequest(w.ctx)

	// Try to send result, but handle channel closure gracefully
	defer func() {
		if r := recover(); r != nil {
			// Channel was closed, ignore the panic
			_ = r
		}
	}()

	select {
	case w.results <- result:
	case <-w.ctx.Done():
		// Worker is stopping, ignore the result
	}
}

// WorkerPool manages a pool of workers
type WorkerPool struct {
	workers []*Worker
	client  *HTTPClient
	results chan metrics.Result
	stopped bool
	mu      sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(size int, cfg *config.Config) *WorkerPool {
	client := NewHTTPClient(cfg)
	results := make(chan metrics.Result, size*10) // Buffered channel
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers: make([]*Worker, size),
		client:  client,
		results: results,
		ctx:     ctx,
		cancel:  cancel,
	}

	for i := 0; i < size; i++ {
		pool.workers[i] = NewWorker(i, client, results)
	}

	return pool
}

// Start starts all workers in the pool
func (wp *WorkerPool) Start() {
	// Workers don't need to be started in the simplified approach
}

// Stop stops all workers in the pool
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return
	}
	wp.stopped = true

	// Cancel the context first to stop new requests
	wp.cancel()

	for _, worker := range wp.workers {
		worker.Stop()
	}
	wp.client.Close()

	// Give a small delay for pending requests to complete
	time.Sleep(100 * time.Millisecond)
	close(wp.results)
}

// SendRequest sends a request to an available worker
func (wp *WorkerPool) SendRequest() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.stopped {
		return
	}

	// Check if the context is done
	select {
	case <-wp.ctx.Done():
		return
	default:
	}

	// Use a simple round-robin approach
	worker := wp.workers[0] // For now, just use the first worker
	go worker.ExecuteRequest()
}

// GetResults returns the results channel
func (wp *WorkerPool) GetResults() <-chan metrics.Result {
	return wp.results
}

// RateLimiter controls the rate of requests
type RateLimiter struct {
	rate     int
	interval time.Duration
	ticker   *time.Ticker
	tokens   chan struct{}
	stopped  bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int) *RateLimiter {
	if rate <= 0 {
		rate = 1
	}

	interval := time.Second / time.Duration(rate)
	rl := &RateLimiter{
		rate:     rate,
		interval: interval,
		ticker:   time.NewTicker(interval),
		tokens:   make(chan struct{}, rate),
	}

	// Pre-fill tokens
	for i := 0; i < rate; i++ {
		rl.tokens <- struct{}{}
	}

	go rl.refillTokens()
	return rl
}

// Wait waits for a token to become available
func (rl *RateLimiter) Wait() {
	if rl.stopped {
		return
	}
	<-rl.tokens
}

// Stop stops the rate limiter
func (rl *RateLimiter) Stop() {
	rl.stopped = true
	if rl.ticker != nil {
		rl.ticker.Stop()
	}
}

// refillTokens refills the token bucket
func (rl *RateLimiter) refillTokens() {
	for range rl.ticker.C {
		if rl.stopped {
			return
		}
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full
		}
	}
}

// ValidateURL validates if the given URL is valid
func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		return fmt.Errorf("URL must include a scheme (http:// or https://)")
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}

// ValidateMethod validates if the given HTTP method is supported
func ValidateMethod(method string) error {
	method = strings.ToUpper(method)
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, valid := range validMethods {
		if method == valid {
			return nil
		}
	}

	return fmt.Errorf("unsupported HTTP method: %s", method)
}
