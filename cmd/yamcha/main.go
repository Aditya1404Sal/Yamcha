package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"Yamcha/internal/attacker"
	"Yamcha/internal/config"
	"Yamcha/internal/dashboard"
	"Yamcha/internal/metrics"
	"Yamcha/internal/reporter"
)

const version = "2.0.0"

func main() {
	// Check for help flag
	if checkHelpFlag() {
		printHelp()
		return
	}

	// Print banner
	printBanner()

	// Check for dashboard mode
	dashboardPort := checkDashboardFlag()
	if dashboardPort > 0 {
		startDashboardMode(dashboardPort)
		return
	}

	// Load configuration
	cfg, err := loadConfiguration()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Set CPU usage
	runtime.GOMAXPROCS(cfg.System.CPUCount)

	// Create metrics collector
	metricsCollector := metrics.NewMetrics()

	// Setup signal handling for graceful shutdown
	ctx, cancel := setupSignalHandling()
	defer cancel()

	// Create reporter manager
	reporterManager, err := reporter.NewReporterManager(cfg)
	if err != nil {
		log.Fatalf("Failed to create reporter manager: %v", err)
	}

	// Show test configuration
	displayTestConfig(cfg)

	// Start real-time reporting if enabled
	var realTimeReporter *reporter.RealTimeReporter
	var rtCollector *metrics.RealTimeCollector
	var stopRealTime chan struct{}

	if cfg.Reporting.EnableRealTime {
		rtCollector = metrics.NewRealTimeCollector(cfg.Reporting.UpdateInterval.ToDuration())
		realTimeReporter = reporter.NewRealTimeReporter(cfg.Reporting.UpdateInterval.ToDuration())
		stopRealTime = make(chan struct{})

		rtCollector.Start()
		go realTimeReporter.Start(rtCollector.GetStatsChan(), stopRealTime)
	}

	// Execute the load test
	fmt.Println("\n🚀 Starting load test...")
	startTime := time.Now()

	if err := executeLoadTest(ctx, cfg, metricsCollector, rtCollector); err != nil {
		if rtCollector != nil {
			rtCollector.Stop()
		}
		if stopRealTime != nil {
			close(stopRealTime)
		}
		log.Fatalf("Load test failed: %v", err)
	}

	endTime := time.Now()
	metricsCollector.SetEndTime(endTime)

	// Stop real-time reporting
	if rtCollector != nil {
		rtCollector.Stop()
	}
	if stopRealTime != nil {
		close(stopRealTime)
	}

	// Generate reports
	fmt.Println("\n📊 Generating reports...")
	stats := metricsCollector.GetStatistics()
	results := metricsCollector.GetResults()

	if err := reporterManager.GenerateReports(stats, cfg, results); err != nil {
		log.Printf("Warning: Failed to generate some reports: %v", err)
	}

	fmt.Printf("\n✅ Load test completed in %v\n", endTime.Sub(startTime))
}

// checkHelpFlag checks if help is requested
func checkHelpFlag() bool {
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" || arg == "help" {
			return true
		}
	}
	return false
}

// printHelp displays usage information
func printHelp() {
	fmt.Printf(`
██╗   ██╗ █████╗ ███╗   ███╗ ██████╗██╗  ██╗ █████╗ 
╚██╗ ██╔╝██╔══██╗████╗ ████║██╔════╝██║  ██║██╔══██╗
 ╚████╔╝ ███████║██╔████╔██║██║     ███████║███████║
  ╚██╔╝  ██╔══██║██║╚██╔╝██║██║     ██╔══██║██╔══██║
   ██║   ██║  ██║██║ ╚═╝ ██║╚██████╗██║  ██║██║  ██║
   ╚═╝   ╚═╝  ╚═╝╚═╝     ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝
                                                      
          Load Testing Tool v%s
          Power Level: Over 9000! 🔥

USAGE:
    yamcha [options]
    yamcha -dashboard [port]     # Start web dashboard
    yamcha -config config.yaml   # Load from config file

🌐 WEB DASHBOARD MODE:
    yamcha -dashboard            # Start on default port 8080
    yamcha -dashboard 9000       # Start on custom port
    
    The dashboard provides:
    • 📊 Real-time test monitoring and charts
    • 🔄 Multiple concurrent test sessions
    • 🎯 Easy test creation and management
    • 📈 Live performance metrics
    • 🗂️ Session history and results

⚡ CLI MODE OPTIONS:
    -url string         Target URL to test
    -method string      HTTP method (GET, POST, PUT, DELETE) (default "GET")
    -requests int       Number of requests (default 100)
    -rate int           Rate limit (requests per second) (default 10)
    -attack string      Attack pattern: steady, burst, spike, sustained, gradual (default "steady")
    -timeout duration   Request timeout (default 30s)
    -workers int        Max concurrent workers (default 10)
    -headers string     HTTP headers as JSON string
    -body string        Request body
    -config string      Configuration file (YAML or JSON)
    
📝 EXAMPLES:
    # CLI: Simple GET test
    yamcha -url https://httpbin.org/get -requests 500 -rate 20

    # CLI: POST test with headers and body
    yamcha -url https://httpbin.org/post -method POST \
           -headers '{"Content-Type":"application/json"}' \
           -body '{"test":"data"}' -requests 100

    # CLI: Using config file
    yamcha -config my-test.yaml

    # Dashboard: Start web interface
    yamcha -dashboard

⚙️  ATTACK PATTERNS:
    steady      Constant rate throughout the test
    burst       Send requests in bursts
    spike       Sudden traffic spikes
    sustained   Long-running constant load
    gradual     Gradually increase load

📊 OUTPUT FORMATS:
    • Console summary with key metrics
    • HTML report with charts (via dashboard)
    • JSON/CSV export (configurable)
    • Real-time terminal progress

For more information and examples, visit: https://github.com/your-repo/yamcha
`, version)
}

// checkDashboardFlag checks if dashboard mode is requested
func checkDashboardFlag() int {
	for i, arg := range os.Args {
		if arg == "-dashboard" || arg == "--dashboard" {
			// Default port
			port := 8080
			// Check if next arg is a port number
			if i+1 < len(os.Args) {
				if portNum, err := strconv.Atoi(os.Args[i+1]); err == nil && portNum > 0 && portNum <= 65535 {
					port = portNum
				}
			}
			return port
		}
	}
	return 0
}

// startDashboardMode starts the web dashboard
func startDashboardMode(port int) {
	fmt.Printf("\n🌐 Starting Yamcha Dashboard on port %d...\n", port)
	fmt.Printf("   Access the dashboard at: http://localhost:%d\n", port)
	fmt.Println("   Press Ctrl+C to stop the dashboard")

	// Create dashboard
	dash := dashboard.NewDashboard(port)

	// Setup signal handling for graceful shutdown
	ctx, cancel := setupSignalHandling()
	defer cancel()

	// Start dashboard in a goroutine
	go func() {
		if err := dash.Start(); err != nil {
			log.Printf("Dashboard server error: %v", err)
			cancel()
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	fmt.Println("\n🛑 Shutting down dashboard...")
	if err := dash.Stop(); err != nil {
		log.Printf("Error stopping dashboard: %v", err)
	}
	fmt.Println("✅ Dashboard stopped")
}

// loadConfiguration loads configuration from flags or config file
func loadConfiguration() (*config.Config, error) {
	// Check if config file is specified
	for i, arg := range os.Args {
		if arg == "-config" && i+1 < len(os.Args) {
			configFile := os.Args[i+1]
			fmt.Printf("Loading configuration from: %s\n", configFile)
			return config.LoadFromFile(configFile)
		}
	}

	// Load from command line flags
	return config.LoadFromFlags()
}

// setupSignalHandling sets up graceful shutdown
func setupSignalHandling() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 Received interrupt signal, stopping load test...")
		cancel()
	}()

	return ctx, cancel
}

// displayTestConfig displays the test configuration
func displayTestConfig(cfg *config.Config) {
	fmt.Println("\n📋 Test Configuration:")
	fmt.Printf("  Target URL:       %s\n", cfg.Target.URL)
	fmt.Printf("  HTTP Method:      %s\n", cfg.Target.Method)
	fmt.Printf("  Attack Type:      %s\n", cfg.Load.AttackType)
	fmt.Printf("  Requests:         %d\n", cfg.Load.Requests)
	fmt.Printf("  Rate (req/s):     %d\n", cfg.Load.Rate)
	fmt.Printf("  Max Workers:      %d\n", cfg.Load.MaxWorkers)
	fmt.Printf("  Timeout:          %v\n", cfg.HTTP.Timeout)
	fmt.Printf("  Keep-Alive:       %t\n", cfg.HTTP.KeepAlive)

	if cfg.Load.AttackType == "sustained" {
		fmt.Printf("  Duration:         %v\n", cfg.Load.Duration)
	}
	if cfg.Load.AttackType == "burst" {
		fmt.Printf("  Burst Count:      %d\n", cfg.Load.BurstCount)
	}
	if cfg.Load.AttackType == "rampup" {
		fmt.Printf("  Step Size:        %d\n", cfg.Load.StepSize)
	}
	if cfg.Load.AttackType == "spike" {
		fmt.Printf("  Spike Height:     %d\n", cfg.Load.SpikeHeight)
	}
}

// executeLoadTest executes the load test with the specified configuration
func executeLoadTest(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics, rtCollector *metrics.RealTimeCollector) error {
	// Validate target URL
	if err := attacker.ValidateURL(cfg.Target.URL); err != nil {
		return fmt.Errorf("invalid target URL: %w", err)
	}

	// Validate HTTP method
	if err := attacker.ValidateMethod(cfg.Target.Method); err != nil {
		return fmt.Errorf("invalid HTTP method: %w", err)
	}

	// Create attacker factory
	factory := attacker.NewAttackerFactory()

	// Create the appropriate attacker
	att, err := factory.CreateAttacker(cfg.Load.AttackType)
	if err != nil {
		return fmt.Errorf("failed to create attacker: %w", err)
	}

	// Use real-time collector if available, otherwise use regular collector
	if rtCollector != nil {
		// Execute the attack with real-time collector
		if err := att.Attack(ctx, cfg, rtCollector.GetMetrics()); err != nil {
			return fmt.Errorf("attack failed: %w", err)
		}

		// Copy results to main collector
		rtResults := rtCollector.GetMetrics().GetResults()
		for _, result := range rtResults {
			metricsCollector.AddResult(result)
		}
	} else {
		// Execute the attack with main collector directly
		if err := att.Attack(ctx, cfg, metricsCollector); err != nil {
			return fmt.Errorf("attack failed: %w", err)
		}
	}

	return nil
}

// printBanner prints the application banner
func printBanner() {
	banner := `
██╗   ██╗ █████╗ ███╗   ███╗ ██████╗██╗  ██╗ █████╗ 
╚██╗ ██╔╝██╔══██╗████╗ ████║██╔════╝██║  ██║██╔══██╗
 ╚████╔╝ ███████║██╔████╔██║██║     ███████║███████║
  ╚██╔╝  ██╔══██║██║╚██╔╝██║██║     ██╔══██║██╔══██║
   ██║   ██║  ██║██║ ╚═╝ ██║╚██████╗██║  ██║██║  ██║
   ╚═╝   ╚═╝  ╚═╝╚═╝     ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝
                                                      
          Load Testing Tool v%s
`
	fmt.Printf(banner, version)
}
