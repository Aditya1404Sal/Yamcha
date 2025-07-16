package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a custom type that can unmarshal from JSON strings
type Duration time.Duration

// UnmarshalJSON implements json.Unmarshaler
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		// Try to unmarshal as number (nanoseconds)
		var ns int64
		if err2 := json.Unmarshal(data, &ns); err2 != nil {
			return err // Return original string unmarshal error
		}
		*d = Duration(ns)
		return nil
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// MarshalJSON implements json.Marshaler
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalYAML implements yaml.Unmarshaler
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}

	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// MarshalYAML implements yaml.Marshaler
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// String returns the string representation
func (d Duration) String() string {
	return time.Duration(d).String()
}

// ToDuration converts to time.Duration
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}

// Config represents the complete configuration for a load test
type Config struct {
	Target    TargetConfig    `json:"target" yaml:"target"`
	Load      LoadConfig      `json:"load" yaml:"load"`
	HTTP      HTTPConfig      `json:"http" yaml:"http"`
	System    SystemConfig    `json:"system" yaml:"system"`
	Output    OutputConfig    `json:"output" yaml:"output"`
	Reporting ReportingConfig `json:"reporting" yaml:"reporting"`
}

// TargetConfig holds target server configuration
type TargetConfig struct {
	URL     string            `json:"url" yaml:"url"`
	Method  string            `json:"method" yaml:"method"`
	Headers map[string]string `json:"headers" yaml:"headers"`
	Body    interface{}       `json:"body" yaml:"body"`
}

// LoadConfig holds load testing parameters
type LoadConfig struct {
	AttackType  string   `json:"attack_type" yaml:"attack_type"`
	Requests    int      `json:"requests" yaml:"requests"`
	Rate        int      `json:"rate" yaml:"rate"`
	Duration    Duration `json:"duration" yaml:"duration"`
	BurstCount  int      `json:"burst_count" yaml:"burst_count"`
	StepSize    int      `json:"step_size" yaml:"step_size"`
	SpikeHeight int      `json:"spike_height" yaml:"spike_height"`
	MaxWorkers  int      `json:"max_workers" yaml:"max_workers"`
}

// HTTPConfig holds HTTP client configuration
type HTTPConfig struct {
	Timeout             Duration `json:"timeout" yaml:"timeout"`
	KeepAlive           bool     `json:"keep_alive" yaml:"keep_alive"`
	MaxIdleConns        int      `json:"max_idle_conns" yaml:"max_idle_conns"`
	MaxIdleConnsPerHost int      `json:"max_idle_conns_per_host" yaml:"max_idle_conns_per_host"`
	IdleConnTimeout     Duration `json:"idle_conn_timeout" yaml:"idle_conn_timeout"`
	DisableTLSVerify    bool     `json:"disable_tls_verify" yaml:"disable_tls_verify"`
}

// SystemConfig holds system-level configuration
type SystemConfig struct {
	CPUCount    int  `json:"cpu_count" yaml:"cpu_count"`
	EnablePprof bool `json:"enable_pprof" yaml:"enable_pprof"`
	PprofPort   int  `json:"pprof_port" yaml:"pprof_port"`
}

// OutputConfig holds output configuration
type OutputConfig struct {
	Directory string   `json:"directory" yaml:"directory"`
	Formats   []string `json:"formats" yaml:"formats"`
	Filename  string   `json:"filename" yaml:"filename"`
}

// ReportingConfig holds reporting configuration
type ReportingConfig struct {
	EnablePlot     bool     `json:"enable_plot" yaml:"enable_plot"`
	EnableProgress bool     `json:"enable_progress" yaml:"enable_progress"`
	EnableRealTime bool     `json:"enable_real_time" yaml:"enable_real_time"`
	UpdateInterval Duration `json:"update_interval" yaml:"update_interval"`
}

// RequestPayload represents the structure for HTTP request data
type RequestPayload struct {
	Headers map[string]string `json:"headers" yaml:"headers"`
	Body    interface{}       `json:"body" yaml:"body"`
}

// LoadFromFile loads configuration from a YAML or JSON file
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}

	// Try YAML first, then JSON
	if err := yaml.Unmarshal(data, config); err != nil {
		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file as YAML or JSON: %w", err)
		}
	}

	// Validate and set defaults
	if err := config.SetDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	return config, nil
}

// LoadFromFlags creates configuration from command line flags
func LoadFromFlags() (*Config, error) {
	config := &Config{}

	// Target flags
	flag.StringVar(&config.Target.URL, "url", "http://localhost:8080", "Target URL for load testing")
	flag.StringVar(&config.Target.Method, "method", "GET", "HTTP method to use")

	// Load flags
	flag.StringVar(&config.Load.AttackType, "attack", "steady", "Type of attack (steady, burst, spike, rampup, random, sustained)")
	flag.IntVar(&config.Load.Requests, "req", 100, "Number of requests to send")
	flag.IntVar(&config.Load.Rate, "rate", 20, "Number of requests per second")

	var durationStr string
	flag.StringVar(&durationStr, "dur", "10s", "Duration for sustained load tests")

	flag.IntVar(&config.Load.BurstCount, "burst", 5, "Number of bursts for burst load attack")
	flag.IntVar(&config.Load.StepSize, "ss", 10, "Step size for ramp-up load")
	flag.IntVar(&config.Load.SpikeHeight, "sh", 10, "Spike height for spike load")
	flag.IntVar(&config.Load.MaxWorkers, "workers", 100, "Maximum number of concurrent workers")

	// HTTP flags
	var timeoutStr string
	flag.StringVar(&timeoutStr, "timeout", "30s", "HTTP request timeout")

	flag.BoolVar(&config.HTTP.KeepAlive, "keep-alive", true, "Enable HTTP keep-alive")
	flag.IntVar(&config.HTTP.MaxIdleConns, "max-idle", 100, "Maximum idle connections")
	flag.IntVar(&config.HTTP.MaxIdleConnsPerHost, "max-idle-per-host", 10, "Maximum idle connections per host")

	var idleTimeoutStr string
	flag.StringVar(&idleTimeoutStr, "idle-timeout", "45s", "Idle connection timeout")

	flag.BoolVar(&config.HTTP.DisableTLSVerify, "insecure", false, "Disable TLS certificate verification")

	// System flags
	flag.IntVar(&config.System.CPUCount, "cpu", runtime.NumCPU(), "Number of CPUs to use")
	flag.BoolVar(&config.System.EnablePprof, "pprof", false, "Enable pprof profiling")
	flag.IntVar(&config.System.PprofPort, "pprof-port", 6060, "Port for pprof server")

	// Output flags
	flag.StringVar(&config.Output.Directory, "output-dir", "./results", "Output directory for results")

	// Reporting flags
	flag.BoolVar(&config.Reporting.EnablePlot, "plot", true, "Enable plotting of results")
	flag.BoolVar(&config.Reporting.EnableProgress, "progress", true, "Enable progress bar")
	flag.BoolVar(&config.Reporting.EnableRealTime, "real-time", false, "Enable real-time metrics")

	var updateIntervalStr string
	flag.StringVar(&updateIntervalStr, "update-interval", "1s", "Real-time update interval")

	// Request payload flags
	var bodyFile string
	flag.StringVar(&bodyFile, "body", "", "Path to JSON file containing request headers and body")

	flag.Parse()

	// Parse duration strings
	if durationStr != "" {
		if dur, err := time.ParseDuration(durationStr); err != nil {
			return nil, fmt.Errorf("invalid duration format: %w", err)
		} else {
			config.Load.Duration = Duration(dur)
		}
	}

	if timeoutStr != "" {
		if dur, err := time.ParseDuration(timeoutStr); err != nil {
			return nil, fmt.Errorf("invalid timeout format: %w", err)
		} else {
			config.HTTP.Timeout = Duration(dur)
		}
	}

	if idleTimeoutStr != "" {
		if dur, err := time.ParseDuration(idleTimeoutStr); err != nil {
			return nil, fmt.Errorf("invalid idle timeout format: %w", err)
		} else {
			config.HTTP.IdleConnTimeout = Duration(dur)
		}
	}

	if updateIntervalStr != "" {
		if dur, err := time.ParseDuration(updateIntervalStr); err != nil {
			return nil, fmt.Errorf("invalid update interval format: %w", err)
		} else {
			config.Reporting.UpdateInterval = Duration(dur)
		}
	}

	// Load request payload if specified
	if bodyFile != "" {
		payload, err := loadRequestPayload(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load request payload: %w", err)
		}
		config.Target.Headers = payload.Headers
		config.Target.Body = payload.Body
	}

	// Set defaults
	if err := config.SetDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	return config, nil
}

// SetDefaults sets default values for configuration
func (c *Config) SetDefaults() error {
	// Target defaults
	if c.Target.Method == "" {
		c.Target.Method = "GET"
	}
	if c.Target.Headers == nil {
		c.Target.Headers = make(map[string]string)
	}

	// Load defaults
	if c.Load.AttackType == "" {
		c.Load.AttackType = "steady"
	}
	if c.Load.Requests <= 0 {
		c.Load.Requests = 100
	}
	if c.Load.Rate <= 0 {
		c.Load.Rate = 20
	}
	if c.Load.MaxWorkers <= 0 {
		c.Load.MaxWorkers = 100
	}
	if c.Load.SpikeHeight <= 0 {
		c.Load.SpikeHeight = 10
	}
	if c.Load.BurstCount <= 0 {
		c.Load.BurstCount = 5
	}
	if c.Load.StepSize <= 0 {
		c.Load.StepSize = 1
	}

	// HTTP defaults
	if c.HTTP.Timeout == 0 {
		c.HTTP.Timeout = Duration(30 * time.Second)
	}
	if c.HTTP.MaxIdleConns <= 0 {
		c.HTTP.MaxIdleConns = 100
	}
	if c.HTTP.MaxIdleConnsPerHost <= 0 {
		c.HTTP.MaxIdleConnsPerHost = 10
	}
	if c.HTTP.IdleConnTimeout == 0 {
		c.HTTP.IdleConnTimeout = Duration(45 * time.Second)
	}

	// System defaults
	if c.System.CPUCount <= 0 {
		c.System.CPUCount = runtime.NumCPU()
	}
	if c.System.PprofPort <= 0 {
		c.System.PprofPort = 6060
	}

	// Output defaults
	if c.Output.Directory == "" {
		c.Output.Directory = "./results"
	}
	if len(c.Output.Formats) == 0 {
		c.Output.Formats = []string{"html", "json"}
	}

	// Reporting defaults
	if c.Reporting.UpdateInterval == 0 {
		c.Reporting.UpdateInterval = Duration(1 * time.Second)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Target.URL == "" {
		return fmt.Errorf("target URL is required")
	}

	if c.Load.Requests <= 0 && c.Load.Duration <= 0 {
		return fmt.Errorf("either requests count or duration must be specified")
	}

	if c.Load.Rate <= 0 {
		return fmt.Errorf("rate must be greater than 0")
	}

	validAttacks := map[string]bool{
		"steady": true, "burst": true, "spike": true,
		"rampup": true, "random": true, "sustained": true, "gradual": true,
	}
	if !validAttacks[c.Load.AttackType] {
		return fmt.Errorf("invalid attack type: %s", c.Load.AttackType)
	}

	return nil
}

// loadRequestPayload loads request payload from a JSON file
func loadRequestPayload(filename string) (*RequestPayload, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Return default payload if file doesn't exist
		return &RequestPayload{
			Headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "Yamcha-LoadTester/2.0",
			},
			Body: map[string]interface{}{
				"message": "Hello from Yamcha",
			},
		}, nil
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var payload RequestPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &payload, nil
}

// SaveToFile saves the configuration to a file
func (c *Config) SaveToFile(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
