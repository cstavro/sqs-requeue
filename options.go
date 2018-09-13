package sqs_requeue

func NewOptions(sourceQueue string, targetQueue string) options {
	o := options{
		SourceQueue: sourceQueue,
		TargetQueue: targetQueue,

		Concurrency:         10,
		WaitTimeSeconds:     5,
		VisibilityTimeout:   30,
		MaxNumberOfMessages: 10,
	}

	return o
}

type options struct {
	SourceQueue string
	TargetQueue string

	Concurrency         int
	WaitTimeSeconds     int64
	VisibilityTimeout   int64
	MaxNumberOfMessages int64
}
