package worker

import (
	"context"
	"log"

	mhandler "github.com/denisakp/pulseguard/internal/maintenance"
	"github.com/hibiken/asynq"
)

// Processor wraps the Asynq server and provides a unified interface for handling tasks.
type Processor struct {
	server             *asynq.Server
	monitoringHandler  *MonitoringTaskHandler
	maintenanceHandler *mhandler.TaskHandler
}

// NewProcessor creates a new worker processor with the given handlers.
func NewProcessor(
	redisOpt asynq.RedisConnOpt,
	monitoringHandler *MonitoringTaskHandler,
	maintenanceHandler *mhandler.TaskHandler,
) *Processor {
	server := asynq.NewServer(redisOpt, asynq.Config{
		// Configure the server with appropriate settings
		Concurrency: 10,
		Queues: map[string]int{
			"monitoring":  10, // All priority to monitoring tasks
			"maintenance": 5,
		},
	})

	return &Processor{
		server:             server,
		monitoringHandler:  monitoringHandler,
		maintenanceHandler: maintenanceHandler,
	}
}

// Start begins processing tasks from the queues.
func (p *Processor) Start(ctx context.Context) error {
	// Register task handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc("monitoring:check", p.monitoringHandler.ProcessTask)
	mux.HandleFunc("maintenance:start", p.maintenanceHandler.ProcessStart)
	mux.HandleFunc("maintenance:end", p.maintenanceHandler.ProcessEnd)
	// Note: notification:send handler removed - notifications are now sent directly by IncidentService

	// Start the server
	log.Println("Starting Asynq worker...")
	return p.server.Start(mux)
}

// Stop gracefully shuts down the processor.
func (p *Processor) Stop() {
	log.Println("Shutting down Asynq worker...")
	p.server.Stop()
}
