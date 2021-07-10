package lazy

type JobInterface interface {
	isLazyJob()
	LazyQueueName() string
}

type EmptyJob struct{}

func (EmptyJob) isLazyJob() {}

type deadJob struct {
	Job    string `json:"job"`
	Reason string `json:"reason"`
}
