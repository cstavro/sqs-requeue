# sqs-requeue

Quickly move messages from one SQS queue to another.

## usage

    sqs-requeue -source-queue <url> -target-queue <url>

## options

    -source-queue <url>        queue to read messages from
    -target-queue <url>        queue to send messages to
    -concurrency <int>         number of concurrent workers (default 10)

## build

    go get github.com/cstavro/sqs-requeue
    go build github.com/cstavro/sqs-requeue/cmd/sqs-requeue