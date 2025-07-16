## Parody
"Yamcha has evolved! This isn't your typical low-power level load tester anymore. With a complete rewrite, Yamcha now packs some serious heat (Power level: Over 9000! 🔥). Though he's still humble enough to make mistakes, at least now he learns from them!"

![FkxqjP1aYAAyT5X](https://github.com/Aditya1404Sal/Yamcha/assets/91340059/74915949-c768-401e-acdf-4d581c468725)

---
# Yamcha 2.0: A Modern Load Testing Tool with Live Dashboard

Yamcha is a powerful, rewritten command-line and web-based load testing tool written in Go for conducting comprehensive performance tests on HTTP/S applications. Now featuring a modern web dashboard for real-time test management and monitoring!

## 🚀 What's New in v2.0

- **Complete Rewrite**: Modern, modular architecture with proper separation of concerns
- **🌐 Live Web Dashboard**: Real-time test management with beautiful charts and multi-session support
- **Enhanced Configuration**: Support for YAML/JSON config files alongside CLI flags
- **Advanced Metrics**: Comprehensive statistics including percentiles, throughput, and error analysis
- **Improved Connection Management**: Proper HTTP connection pooling and resource management
- **Multiple Output Formats**: HTML, JSON, CSV reports with beautiful visualizations
- **Real-time Monitoring**: Live metrics updates and WebSocket-based dashboard updates
- **Better Error Handling**: Graceful shutdown and proper error reporting
- **Extensive Testing**: Unit tests and validation for reliable operation

## ✨ Features

### 🌐 Web Dashboard (NEW!)
- **Real-time Test Management**: Create, start, stop, and monitor multiple tests from a modern web interface
- **Live Performance Charts**: Real-time response time, throughput, and error rate visualization
- **Multi-Session Support**: Run and monitor multiple load tests simultaneously
- **Interactive Test Creation**: Easy form-based test configuration with all attack patterns
- **WebSocket Updates**: Live test status and metrics updates without page refresh
- **Session History**: View past and current test results with detailed statistics
- **Responsive Design**: Works on desktop, tablet, and mobile devices

### Core Testing Capabilities
- **Multiple Attack Patterns**: 
  - **Steady**: Consistent load at specified rate
  - **Burst**: Multiple bursts of rapid requests
  - **Spike**: Random spikes in load with pauses
  - **Sustained**: Continuous load for specified duration
  - **Gradual**: Gradually increasing load over time
- **HTTP Method Support**: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- **Advanced Request Configuration**: Custom headers, JSON body, authentication
- **Worker Pool**: Efficient concurrent request handling with configurable worker count
- **Rate Limiting**: Precise control over request rate with token bucket algorithm

### Performance & Reliability
- **Connection Pooling**: Reuse HTTP connections for better performance
- **Resource Management**: Proper cleanup and graceful shutdown
- **Timeout Control**: Configurable request timeouts and connection settings
- **TLS Configuration**: Support for custom TLS settings and certificate validation
- **CPU Utilization**: Configurable CPU core usage

### Monitoring & Reporting
- **Real-time Progress**: Beautiful progress bars with live statistics
- **Comprehensive Metrics**:
  - Response time statistics (min, max, average, percentiles)
  - Throughput (requests/second, bytes/second)
  - Success/failure rates and error categorization
  - Status code distribution
- **Multiple Output Formats**: HTML with charts, JSON, CSV
- **Interactive HTML Reports**: Beautiful charts and detailed statistics

### Configuration & Usability
- **Configuration Files**: YAML and JSON config file support
- **Environment-friendly**: Easy installation and Docker support
- **Validation**: Input validation and helpful error messages
- **Extensible**: Plugin-ready architecture for custom attackers

## 📦 Installation

### From Source
```bash
git clone https://github.com/Aditya1404Sal/Yamcha.git
cd Yamcha
make build
```

### Quick Build
```bash
go build -o yamcha ./cmd/yamcha
```

### Install to PATH
```bash
make install
# or manually:
sudo cp yamcha /usr/local/bin/
```

## 🎯 Quick Start

### Basic Usage
```bash
# Simple GET request test
./yamcha -url https://api.example.com -req 100 -rate 20

# POST request with custom headers and body
./yamcha -url https://api.example.com/users -method POST -req 50 -rate 10 -body body.json

# Burst load test
./yamcha -url https://example.com -attack burst -req 25 -burst 5 -rate 15

# Sustained load for 30 seconds
./yamcha -url https://example.com -attack sustained -dur 30s -rate 10
```

### Using Configuration Files
```bash
# Use YAML config
./yamcha -config config.yaml

# Use JSON config  
./yamcha -config config.json
```

### Example Configuration (YAML)
```yaml
target:
  url: "https://api.example.com/users"
  method: "POST"
  headers:
    Content-Type: "application/json"
    Authorization: "Bearer your-token"
  body:
    name: "Test User"
    email: "test@example.com"

load:
  attack_type: "steady"
  requests: 1000
  rate: 50
  max_workers: 100

http:
  timeout: 30s
  keep_alive: true
  max_idle_conns: 100

output:
  directory: "./results"
  formats: ["html", "json", "csv"]

reporting:
  enable_plot: true
  enable_progress: true
  enable_real_time: true
```

## 🔧 Command Line Options

### Target Configuration
- `-url`: Target URL for testing
- `-method`: HTTP method (GET, POST, PUT, DELETE, etc.)
- `-body`: Path to JSON file with headers and body

### Load Configuration  
- `-attack`: Attack pattern (steady, burst, spike, rampup, random, sustained)
- `-req`: Number of requests to send
- `-rate`: Requests per second
- `-dur`: Duration for sustained attacks
- `-burst`: Number of bursts for burst attack
- `-ss`: Step size for ramp-up attack
- `-sh`: Spike height for spike attack
- `-workers`: Maximum concurrent workers

### HTTP Configuration
- `-timeout`: Request timeout
- `-keep-alive`: Enable HTTP keep-alive (default: true)
- `-max-idle`: Maximum idle connections
- `-max-idle-per-host`: Maximum idle connections per host
- `-idle-timeout`: Idle connection timeout
- `-insecure`: Disable TLS certificate verification

### System Configuration
- `-cpu`: Number of CPU cores to use
- `-pprof`: Enable pprof profiling
- `-pprof-port`: Port for pprof server

### Output Configuration
- `-output-dir`: Output directory for results
- `-plot`: Enable HTML plotting (default: true)
- `-progress`: Enable progress bar (default: true)
- `-real-time`: Enable real-time metrics
- `-update-interval`: Real-time update interval

## 📊 Sample Output

```
================================================================================
                       YAMCHA LOAD TEST RESULTS
================================================================================
Target URL:           https://api.example.com/users
Attack Type:          steady
HTTP Method:          POST
Test Duration:        10.523s

---------------------------------------- SUMMARY ----------------------------------------
Total Requests:       200
Successful Requests:  198 (99.00%)
Failed Requests:      2 (1.00%)
Requests/sec:         19.01
Data Transferred:     45.23 KB/sec

----------------------------------- RESPONSE TIMES -----------------------------------
Average:              52.3ms
Minimum:              12.1ms
Maximum:              245.7ms
Median (P50):         48.2ms
95th Percentile:      89.4ms
99th Percentile:      156.8ms

----------------------------------- STATUS CODES -------------------------------------
  200:                  195 (97.50%)
  201:                  3 (1.50%)
  429:                  2 (1.00%)
================================================================================
```

## 🛠️ Development

### Prerequisites
- Go 1.22+
- Make (optional)

### Building
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run with coverage
make test-coverage

# Format and lint
make fmt
make lint

# Setup development environment
make dev-setup
```

### Project Structure
```
Yamcha/
├── cmd/yamcha/          # Main application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── attacker/        # Attack patterns and HTTP client
│   ├── metrics/         # Metrics collection and statistics
│   └── reporter/        # Output generation and reporting
├── pkg/                 # Public packages (future)
├── config.example.yaml  # Example configuration files
├── config.example.json
├── Makefile            # Build automation
└── README.md
```

### Contributing
1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass: `make test`
5. Submit a pull request

## 📈 Performance Tips

1. **Tune Worker Count**: Start with 50-100 workers and adjust based on your system
2. **Connection Pooling**: Enable keep-alive for better performance
3. **Rate Limiting**: Use appropriate rates to avoid overwhelming the target
4. **Monitoring**: Use real-time monitoring for long-running tests
5. **Resource Limits**: Monitor CPU and memory usage during tests

## 🔍 Troubleshooting

### Common Issues
- **Connection Refused**: Check if target URL is accessible
- **High Error Rates**: Reduce rate or increase timeout
- **Memory Usage**: Reduce worker count or request batch size
- **TLS Errors**: Use `-insecure` flag for testing environments

### Debug Mode
```bash
# Enable verbose logging
export YAMCHA_DEBUG=1
./yamcha [options]

# Enable profiling
./yamcha -pprof -pprof-port 6060 [options]
# Then visit http://localhost:6060/debug/pprof/
```

## 📄 License

MIT License - see LICENSE file for details.

## 🤝 Acknowledgments

- Built with Go and modern development practices
- Inspired by tools like Vegeta, Artillery, and Apache Bench
- Uses Chart.js for beautiful HTML reports
- Progress bars powered by progressbar/v3

---

**Remember**: With great power comes great responsibility. Use Yamcha responsibly and always get permission before load testing production systems! 🥋


