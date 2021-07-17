package lazy

import "time"

type JobInterface interface {
	isLazyJob()
	LazyQueueName() string
}

type EmptyJob struct {
	// Attempts is the number of times the job was executed by the lazy server.
	// Before execute handlers, Attempts grows up.
	Attempts int `json:"attempts"`

	// CreatedAT is the time the job was created.
	CreatedAT time.Time `json:"created_at"`
}

func (EmptyJob) isLazyJob() {}

type deadJob struct {
	Job    string `json:"job"`
	Reason string `json:"reason"`
}
