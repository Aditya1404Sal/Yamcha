package attacker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"Yamcha/internal/config"
	"Yamcha/internal/metrics"

	"github.com/schollz/progressbar/v3"
)

// Attacker interface defines different attack patterns
type Attacker interface {
	Attack(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics) error
	Name() string
}

// BaseAttacker provides common functionality for all attackers
type BaseAttacker struct {
	workerPool  *WorkerPool
	progress    *progressbar.ProgressBar
	rateLimiter *RateLimiter
}

// newBaseAttacker creates a new base attacker
func newBaseAttacker(cfg *config.Config, totalRequests int) *BaseAttacker {
	workerPool := NewWorkerPool(cfg.Load.MaxWorkers, cfg)

	var progress *progressbar.ProgressBar
	if cfg.Reporting.EnableProgress {
		progress = progressbar.NewOptions(totalRequests,
			progressbar.OptionSetDescription(fmt.Sprintf("[cyan]%s attack[reset]", cfg.Load.AttackType)),
			progressbar.OptionSetWidth(50),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetItsString("req"),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionSpinnerType(14),
		)
	}

	rateLimiter := NewRateLimiter(cfg.Load.Rate)

	return &BaseAttacker{
		workerPool:  workerPool,
		progress:    progress,
		rateLimiter: rateLimiter,
	}
}

// cleanup cleans up resources
func (ba *BaseAttacker) cleanup() {
	if ba.progress != nil {
		ba.progress.Finish()
	}
	if ba.rateLimiter != nil {
		ba.rateLimiter.Stop()
	}
	if ba.workerPool != nil {
		ba.workerPool.Stop()
	}
}

// SteadyAttacker implements steady load pattern
type SteadyAttacker struct {
	*BaseAttacker
}

// NewSteadyAttacker creates a new steady attacker
func NewSteadyAttacker() *SteadyAttacker {
	return &SteadyAttacker{}
}

// Name returns the attacker name
func (sa *SteadyAttacker) Name() string {
	return "steady"
}

// Attack executes steady load attack
func (sa *SteadyAttacker) Attack(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics) error {
	sa.BaseAttacker = newBaseAttacker(cfg, cfg.Load.Requests)
	defer sa.cleanup()

	sa.workerPool.Start()

	// Collect results in background
	var wg sync.WaitGroup
	wg.Add(1)
	go sa.collectResults(ctx, metricsCollector, &wg, cfg.Load.Requests)

	// Send requests at steady rate
	for i := 0; i < cfg.Load.Requests; i++ {
		select {
		case <-ctx.Done():
			sa.workerPool.Stop()
			wg.Wait()
			return ctx.Err()
		default:
			sa.rateLimiter.Wait()
			sa.workerPool.SendRequest()
		}
	}

	// Wait a bit for all requests to complete
	time.Sleep(time.Duration(cfg.Load.Requests)*time.Second/time.Duration(cfg.Load.Rate) + 2*time.Second)

	// Stop the worker pool and wait for all results
	sa.workerPool.Stop()
	wg.Wait()

	return nil
}

// collectResults collects results from the worker pool
func (sa *SteadyAttacker) collectResults(ctx context.Context, metricsCollector *metrics.Metrics, wg *sync.WaitGroup, expectedRequests int) {
	defer wg.Done()

	requestCount := 0
	for requestCount < expectedRequests {
		select {
		case result, ok := <-sa.workerPool.GetResults():
			if !ok {
				return
			}
			metricsCollector.AddResult(result)
			requestCount++
			if sa.progress != nil {
				sa.progress.Add(1)
			}
		case <-ctx.Done():
			return
		}
	}
}

// BurstAttacker implements burst load pattern
type BurstAttacker struct {
	*BaseAttacker
}

// NewBurstAttacker creates a new burst attacker
func NewBurstAttacker() *BurstAttacker {
	return &BurstAttacker{}
}

// Name returns the attacker name
func (ba *BurstAttacker) Name() string {
	return "burst"
}

// Attack executes burst load attack
func (ba *BurstAttacker) Attack(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics) error {
	totalRequests := cfg.Load.Requests * cfg.Load.BurstCount
	ba.BaseAttacker = newBaseAttacker(cfg, totalRequests)
	defer ba.cleanup()

	ba.workerPool.Start()

	// Collect results in background
	var wg sync.WaitGroup
	wg.Add(1)
	go ba.collectResults(ctx, metricsCollector, &wg, totalRequests)

	// Execute bursts
	for burst := 0; burst < cfg.Load.BurstCount; burst++ {
		select {
		case <-ctx.Done():
			ba.workerPool.Stop()
			wg.Wait()
			return ctx.Err()
		default:
		}

		// Send burst of requests
		for i := 0; i < cfg.Load.Requests; i++ {
			select {
			case <-ctx.Done():
				ba.workerPool.Stop()
				wg.Wait()
				return ctx.Err()
			default:
				ba.workerPool.SendRequest()
			}
		}

		// Wait between bursts (except for the last one)
		if burst < cfg.Load.BurstCount-1 {
			time.Sleep(time.Second / time.Duration(cfg.Load.Rate))
		}
	}

	// Wait for all requests to complete
	time.Sleep(time.Duration(totalRequests)*time.Second/time.Duration(cfg.Load.Rate) + 2*time.Second)
	ba.workerPool.Stop()
	wg.Wait()

	return nil
}

// collectResults collects results from the worker pool
func (ba *BurstAttacker) collectResults(ctx context.Context, metricsCollector *metrics.Metrics, wg *sync.WaitGroup, expectedRequests int) {
	defer wg.Done()

	requestCount := 0
	for requestCount < expectedRequests {
		select {
		case result, ok := <-ba.workerPool.GetResults():
			if !ok {
				return
			}
			metricsCollector.AddResult(result)
			requestCount++
			if ba.progress != nil {
				ba.progress.Add(1)
			}
		case <-ctx.Done():
			return
		}
	}
}

// RampUpAttacker implements ramp-up load pattern
type RampUpAttacker struct {
	*BaseAttacker
}

// NewRampUpAttacker creates a new ramp-up attacker
func NewRampUpAttacker() *RampUpAttacker {
	return &RampUpAttacker{}
}

// Name returns the attacker name
func (ra *RampUpAttacker) Name() string {
	return "rampup"
}

// Attack executes ramp-up load attack
func (ra *RampUpAttacker) Attack(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics) error {
	ra.BaseAttacker = newBaseAttacker(cfg, cfg.Load.Requests)
	defer ra.cleanup()

	ra.workerPool.Start()

	// Collect results in background
	var wg sync.WaitGroup
	wg.Add(1)
	go ra.collectResults(ctx, metricsCollector, &wg)

	// Gradually increase load
	for i := 0; i < cfg.Load.Requests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ra.rateLimiter.Wait()
		ra.workerPool.SendRequest()

		if ra.progress != nil {
			ra.progress.Add(1)
		}

		// Increase delay every step_size requests
		if (i+1)%cfg.Load.StepSize == 0 {
			time.Sleep(time.Second / time.Duration(cfg.Load.Rate))
		}
	}

	ra.workerPool.Stop()
	wg.Wait()

	return nil
}

// collectResults collects results from the worker pool
func (ra *RampUpAttacker) collectResults(ctx context.Context, metricsCollector *metrics.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case result, ok := <-ra.workerPool.GetResults():
			if !ok {
				return
			}
			metricsCollector.AddResult(result)
		case <-ctx.Done():
			return
		}
	}
}

// SpikeAttacker implements spike load pattern
type SpikeAttacker struct {
	*BaseAttacker
}

// NewSpikeAttacker creates a new spike attacker
func NewSpikeAttacker() *SpikeAttacker {
	return &SpikeAttacker{}
}

// Name returns the attacker name
func (sa *SpikeAttacker) Name() string {
	return "spike"
}

// Attack executes spike load attack
func (sa *SpikeAttacker) Attack(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics) error {
	sa.BaseAttacker = newBaseAttacker(cfg, cfg.Load.Requests)
	defer sa.cleanup()

	sa.workerPool.Start()

	// Collect results in background
	var wg sync.WaitGroup
	wg.Add(1)
	go sa.collectResults(ctx, metricsCollector, &wg)

	// Send requests with random spikes
	for i := 0; i < cfg.Load.Requests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		sa.workerPool.SendRequest()

		if sa.progress != nil {
			sa.progress.Add(1)
		}

		// Create spikes at regular intervals
		if i%cfg.Load.SpikeHeight == 0 && i > 0 {
			// Random pause for spike effect
			pause := time.Duration(rand.Intn(3)+1) * time.Second
			time.Sleep(pause)
		} else {
			sa.rateLimiter.Wait()
		}
	}

	sa.workerPool.Stop()
	wg.Wait()

	return nil
}

// collectResults collects results from the worker pool
func (sa *SpikeAttacker) collectResults(ctx context.Context, metricsCollector *metrics.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case result, ok := <-sa.workerPool.GetResults():
			if !ok {
				return
			}
			metricsCollector.AddResult(result)
		case <-ctx.Done():
			return
		}
	}
}

// RandomAttacker implements random load pattern
type RandomAttacker struct {
	*BaseAttacker
}

// NewRandomAttacker creates a new random attacker
func NewRandomAttacker() *RandomAttacker {
	return &RandomAttacker{}
}

// Name returns the attacker name
func (ra *RandomAttacker) Name() string {
	return "random"
}

// Attack executes random load attack
func (ra *RandomAttacker) Attack(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics) error {
	ra.BaseAttacker = newBaseAttacker(cfg, cfg.Load.Requests)
	defer ra.cleanup()

	ra.workerPool.Start()

	// Collect results in background
	var wg sync.WaitGroup
	wg.Add(1)
	go ra.collectResults(ctx, metricsCollector, &wg)

	// Send requests with random intervals
	for i := 0; i < cfg.Load.Requests; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		ra.workerPool.SendRequest()

		if ra.progress != nil {
			ra.progress.Add(1)
		}

		// Random delay
		randomDelay := time.Duration(rand.Intn(1000)) * time.Millisecond / time.Duration(cfg.Load.Rate)
		time.Sleep(randomDelay)
	}

	ra.workerPool.Stop()
	wg.Wait()

	return nil
}

// collectResults collects results from the worker pool
func (ra *RandomAttacker) collectResults(ctx context.Context, metricsCollector *metrics.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case result, ok := <-ra.workerPool.GetResults():
			if !ok {
				return
			}
			metricsCollector.AddResult(result)
		case <-ctx.Done():
			return
		}
	}
}

// SustainedAttacker implements sustained load pattern
type SustainedAttacker struct {
	*BaseAttacker
}

// NewSustainedAttacker creates a new sustained attacker
func NewSustainedAttacker() *SustainedAttacker {
	return &SustainedAttacker{}
}

// Name returns the attacker name
func (sa *SustainedAttacker) Name() string {
	return "sustained"
}

// Attack executes sustained load attack
func (sa *SustainedAttacker) Attack(ctx context.Context, cfg *config.Config, metricsCollector *metrics.Metrics) error {
	// For sustained attacks, calculate duration from requests and rate if duration is not set
	var duration time.Duration
	if cfg.Load.Duration.ToDuration() > 0 {
		duration = cfg.Load.Duration.ToDuration()
	} else {
		// Calculate duration based on requests and rate
		duration = time.Duration(cfg.Load.Requests/cfg.Load.Rate) * time.Second
		if duration == 0 {
			duration = 30 * time.Second // Default to 30 seconds
		}
	}

	estimatedRequests := int(duration.Seconds()) * cfg.Load.Rate
	sa.BaseAttacker = newBaseAttacker(cfg, estimatedRequests)
	defer sa.cleanup()

	sa.workerPool.Start()

	// Collect results in background
	var wg sync.WaitGroup
	wg.Add(1)
	go sa.collectResults(ctx, metricsCollector, &wg)

	// Create a context with timeout for the duration
	sustainedCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	requestCount := 0
	// Send requests for the specified duration
	for {
		select {
		case <-sustainedCtx.Done():
			sa.workerPool.Stop()
			wg.Wait()
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		sa.rateLimiter.Wait()
		sa.workerPool.SendRequest()
		requestCount++

		if sa.progress != nil {
			sa.progress.Add(1)
		}
	}
}

// collectResults collects results from the worker pool
func (sa *SustainedAttacker) collectResults(ctx context.Context, metricsCollector *metrics.Metrics, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case result, ok := <-sa.workerPool.GetResults():
			if !ok {
				return
			}
			metricsCollector.AddResult(result)
		case <-ctx.Done():
			return
		}
	}
}

// AttackerFactory creates attackers based on configuration
type AttackerFactory struct{}

// NewAttackerFactory creates a new attacker factory
func NewAttackerFactory() *AttackerFactory {
	return &AttackerFactory{}
}

// CreateAttacker creates an attacker based on the attack type
func (af *AttackerFactory) CreateAttacker(attackType string) (Attacker, error) {
	switch attackType {
	case "steady":
		return NewSteadyAttacker(), nil
	case "burst":
		return NewBurstAttacker(), nil
	case "rampup":
		return NewRampUpAttacker(), nil
	case "spike":
		return NewSpikeAttacker(), nil
	case "random":
		return NewRandomAttacker(), nil
	case "sustained":
		return NewSustainedAttacker(), nil
	case "gradual":
		return NewRampUpAttacker(), nil // Map gradual to rampup
	default:
		return nil, fmt.Errorf("unknown attack type: %s", attackType)
	}
}
