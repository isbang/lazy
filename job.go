package lazy

import "time"

type JobInterface interface {
	isLazyJob()
	LazyQueueName() string
}

type EmptyJob struct{}

func (EmptyJob) isLazyJob() {}

type deadJob struct {
	baseJob

	// Reason is explanation why job is dead. Normally, error string.
	Reason string `json:"reason"`
}

type baseJob struct {
	// Job is the job string.
	Job string `json:"job"`

	// CreatedAT is the time the job was created.
	CreatedAT time.Time `json:"created_at"`

	// Attempts is the number of times the job was executed by the lazy server.
	// Before execute handlers, Attempts grows up.
	Attempts int `json:"attempts"`
}
