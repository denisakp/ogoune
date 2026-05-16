package maintenance

import "github.com/hibiken/asynq"

// TaskEnqueuer abstracts asynq.Client.Enqueue for testability.
type TaskEnqueuer interface {
	Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

// PeriodicTaskRegistrar abstracts asynq.Scheduler.Register for testability.
type PeriodicTaskRegistrar interface {
	Register(cronspec string, task *asynq.Task, opts ...asynq.Option) (string, error)
}

// AsynqClientAdapter wraps *asynq.Client to satisfy TaskEnqueuer.
type AsynqClientAdapter struct {
	Client *asynq.Client
}

func (a *AsynqClientAdapter) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return a.Client.Enqueue(task, opts...)
}

// AsynqSchedulerAdapter wraps *asynq.Scheduler to satisfy PeriodicTaskRegistrar.
type AsynqSchedulerAdapter struct {
	Scheduler *asynq.Scheduler
}

func (a *AsynqSchedulerAdapter) Register(cronspec string, task *asynq.Task, opts ...asynq.Option) (string, error) {
	return a.Scheduler.Register(cronspec, task, opts...)
}
