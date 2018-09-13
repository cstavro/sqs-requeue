package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/cstavro/sqs-requeue"
)

type options struct {
	SourceQueue string
	TargetQueue string
	Concurrency int
}

var args = options{}

func init() {
	flag.StringVar(&args.SourceQueue, "source-queue", "", "queue to read messages from")
	flag.StringVar(&args.TargetQueue, "target-queue", "", "queue to move messages to")
	flag.IntVar(&args.Concurrency, "concurrency", 10, "number of concurrent workers")
	flag.Parse()
}

func main() {
	o := sqs_requeue.NewOptions(args.SourceQueue, args.TargetQueue)
	o.Concurrency = args.Concurrency

	start := time.Now()
	sqs_requeue.Run(o)
	end := time.Now()
	duration := end.Sub(start)
	fmt.Printf("done in %d seconds\n", int(duration.Seconds()))
}
