package scheduler

import (
	"context"
	"sync"
	"time"
)

// TimingWheel represents an in-process timing wheel scheduler for monitoring checks.
type TimingWheelScheduler struct {
	config   *Config
	state    State
	ticker   *time.Ticker
	stopChan chan struct{}
	doneChan chan struct{}

	// Schedule state
	mu        sync.RWMutex
	schedules map[string]*ScheduledResource // resourceID -> ScheduledResource
	paused    map[string]bool               // paused resource IDs

	// Dispatch state
	checkQueue chan *CheckJob
	notifQueue chan *NotificationJob

	// Shutdown coordination (T044: notification worker)
	notifWorkerDone chan struct{}

	// Worker pool
	workers chan struct{}
	wg      sync.WaitGroup
}

// ScheduledResource tracks the scheduling state of a resource
type ScheduledResource struct {
	ResourceID string
	Interval   time.Duration
	NextDue    time.Time
}

// CheckJob represents a check to be executed
type CheckJob struct {
	ResourceID string
	Interval   time.Duration
	Context    context.Context
}

// NotificationJob represents a notification to be sent
type NotificationJob struct {
	IncidentID string
	EventType  string
}

// State represents the current state of the timing wheel scheduler.
type State int

const (
	StateStopped State = iota
	StateRunning
	StateStopping
)

// NewTimingWheel creates a new timing wheel scheduler instance.
func NewTimingWheel(cfg *Config) (*TimingWheelScheduler, error) {
	if cfg == nil {
		cfg = &Config{
			Mode: ModeTimingWheel,
			TimingWheel: TimingWheelConfig{
				TickInterval:          1 * time.Second,
				MaxWorkers:            10,
				ShutdownTimeout:       15 * time.Second,
				NotificationQueueSize: 100,
			},
		}
	}

	return &TimingWheelScheduler{
		config:     cfg,
		state:      StateStopped,
		stopChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
		schedules:  make(map[string]*ScheduledResource),
		paused:     make(map[string]bool),
		checkQueue: make(chan *CheckJob, cfg.TimingWheel.MaxWorkers),
		notifQueue: make(chan *NotificationJob, cfg.TimingWheel.NotificationQueueSize),
		workers:    make(chan struct{}, cfg.TimingWheel.MaxWorkers),
	}, nil
}

// Start initializes and starts the timing wheel scheduler.
func (tw *TimingWheelScheduler) Start(ctx context.Context, repo ActiveResourceRepository) error {
	if tw.state != StateStopped {
		return ErrSchedulerAlreadyRunning
	}

	tw.stopChan = make(chan struct{})
	tw.doneChan = make(chan struct{})
	tw.checkQueue = make(chan *CheckJob, tw.config.TimingWheel.MaxWorkers)
	tw.notifQueue = make(chan *NotificationJob, tw.config.TimingWheel.NotificationQueueSize)
	tw.workers = make(chan struct{}, tw.config.TimingWheel.MaxWorkers)
	tw.schedules = make(map[string]*ScheduledResource)
	tw.paused = make(map[string]bool)
	tw.state = StateRunning
	tw.ticker = time.NewTicker(tw.config.TimingWheel.TickInterval)
	tw.notifWorkerDone = make(chan struct{})

	// Initialize worker pool
	for i := 0; i < tw.config.TimingWheel.MaxWorkers; i++ {
		tw.workers <- struct{}{}
	}

	// Load initial schedules from repository
	if repo != nil {
		items, err := repo.FindScheduledResources(ctx)
		if err == nil && items != nil {
			tw.mu.Lock()
			for _, item := range items {
				if !item.Paused && item.Interval > 0 {
					tw.schedules[item.ResourceID] = &ScheduledResource{
						ResourceID: item.ResourceID,
						Interval:   item.Interval,
						NextDue:    time.Now().Add(item.Interval),
					}
				}
			}
			tw.mu.Unlock()
		}
	}

	// Start background goroutines
	tw.wg.Add(2)
	go tw.run()
	go tw.notificationWorker()

	return nil
}

// Stop gracefully shuts down the timing wheel scheduler.
// T045: Enhanced graceful shutdown sequencing with timeout handling
func (tw *TimingWheelScheduler) Stop(ctx context.Context) error {
	if tw.state != StateRunning {
		return ErrSchedulerNotRunning
	}

	tw.state = StateStopping
	close(tw.stopChan)

	// Wait for graceful shutdown with timeout
	// Shutdown sequence:
	// 1. Stop accepting new ticks
	// 2. Drain in-flight check queue
	// 3. Drain notification worker
	// 4. Notify caller on success or return timeout error
	done := make(chan struct{})
	go func() {
		tw.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		tw.state = StateStopped
		return nil
	case <-ctx.Done():
		tw.state = StateStopped
		return ErrShutdownTimeout
	}
}

// Schedule adds or updates a resource's check schedule.
func (tw *TimingWheelScheduler) Schedule(resourceID string, interval time.Duration) error {
	if interval <= 0 {
		return ErrInvalidInterval
	}

	tw.mu.Lock()
	defer tw.mu.Unlock()

	// First dispatch should occur within interval + 1 tick
	tw.schedules[resourceID] = &ScheduledResource{
		ResourceID: resourceID,
		Interval:   interval,
		NextDue:    time.Now().Add(interval).Add(tw.config.TimingWheel.TickInterval),
	}
	tw.paused[resourceID] = false

	return nil
}

// Unschedule removes a resource from the scheduling queue.
func (tw *TimingWheelScheduler) Unschedule(resourceID string) error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	delete(tw.schedules, resourceID)
	delete(tw.paused, resourceID)

	return nil
}

// Pause temporarily stops scheduling for a resource.
func (tw *TimingWheelScheduler) Pause(resourceID string) error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if _, exists := tw.schedules[resourceID]; !exists {
		return nil // No-op if not scheduled
	}

	tw.paused[resourceID] = true

	return nil
}

// Resume resumes scheduling for a paused resource.
func (tw *TimingWheelScheduler) Resume(resourceID string) error {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	if _, exists := tw.schedules[resourceID]; !exists {
		return nil // No-op if not scheduled
	}

	tw.paused[resourceID] = false

	return nil
}

func (tw *TimingWheelScheduler) run() {
	defer tw.wg.Done()
	defer close(tw.doneChan)
	defer tw.ticker.Stop()

	for {
		select {
		case <-tw.stopChan:
			tw.state = StateStopped
			// Cleanup schedules
			tw.cleanup()
			return
		case <-tw.ticker.C:
			// Process tick - check for due resources
			tw.tick()
		}
	}
}

func (tw *TimingWheelScheduler) tick() {
	tw.mu.RLock()
	now := time.Now()
	var dueIds []string

	// Find all resources that are due
	for id, sched := range tw.schedules {
		if !tw.paused[id] && now.After(sched.NextDue) {
			dueIds = append(dueIds, id)
		}
	}
	tw.mu.RUnlock()

	// Dispatch due checks (non-blocking)
	for _, id := range dueIds {
		tw.mu.RLock()
		sched, exists := tw.schedules[id]
		tw.mu.RUnlock()

		if !exists {
			continue
		}

		select {
		case tw.checkQueue <- &CheckJob{
			ResourceID: id,
			Interval:   sched.Interval,
			Context:    context.Background(),
		}:
			// Update next due time
			tw.mu.Lock()
			tw.schedules[id].NextDue = time.Now().Add(sched.Interval)
			tw.mu.Unlock()
		default:
			// T043: Saturation handling - queue is full (non-blocking dispatch)
			// Retain NextDue unchanged so resource remains due for immediate retry
			// on next tick. Worker pool may be busy or queue backlog exists.
			// This ensures no checks are lost due to congestion.
		}
	}
}

func (tw *TimingWheelScheduler) cleanup() {
	// T045: Enhanced cleanup for graceful shutdown
	// Clear all schedules to prevent new dispatches
	tw.mu.Lock()
	tw.schedules = make(map[string]*ScheduledResource)
	tw.paused = make(map[string]bool)
	tw.mu.Unlock()

	// Close check queue so external dispatchers can exit cleanly.
	close(tw.checkQueue)

	// Signal notification worker to exit
	close(tw.notifQueue)
}

// CheckJobs exposes the timingwheel check queue for in-process dispatchers.
func (tw *TimingWheelScheduler) CheckJobs() <-chan *CheckJob {
	return tw.checkQueue
}

// EnqueueNotification adds a notification to the queue (non-blocking)
func (tw *TimingWheelScheduler) EnqueueNotification(incidentID, eventType string) error {
	if err := ValidateNotificationEventType(eventType); err != nil {
		return err
	}

	select {
	case tw.notifQueue <- &NotificationJob{
		IncidentID: incidentID,
		EventType:  eventType,
	}:
		return nil
	default:
		// T044: Queue full - drop notification (non-blocking requirement)
		// Worker may be congested; notification loss is acceptable for
		// rate-limited incident notifications during high-load conditions.
		return nil
	}
}

// notificationWorker processes notifications asynchronously
// T044: Asynchronous notification worker for graceful queue handling
func (tw *TimingWheelScheduler) notificationWorker() {
	defer tw.wg.Done()
	defer close(tw.notifWorkerDone)

	for {
		select {
		case job, ok := <-tw.notifQueue:
			if !ok {
				// Channel closed during shutdown, drain remaining items
				return
			}
			if job != nil {
				// Process notification
				// In production, this would send to incident service,
				// notification channels, etc. For now, just mark processed.
				_ = job
			}
		case <-tw.stopChan:
			// Shutdown signal received, drain queue and exit
			for {
				select {
				case job, ok := <-tw.notifQueue:
					if !ok {
						return
					}
					if job != nil {
						_ = job
					}
				default:
					return
				}
			}
		}
	}
}

// Compatibility aliases for existing code (use original TimingWheel name)
type TimingWheel = TimingWheelScheduler
