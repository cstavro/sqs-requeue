package sqs_requeue

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func Run(options options) {
	var wg sync.WaitGroup
	waitCh := make(chan struct{})
	failedMessages := make(chan map[string]string)

	sess := session.Must(session.NewSession())
	config := &aws.Config{}
	client := sqs.New(sess, config)

	go func() {
		for i := 0; i < options.Concurrency; i++ {
			wg.Add(1)
			go requeue(context.Background(), options, client, &wg, failedMessages)
		}

		wg.Wait()
		close(waitCh)
	}()

	running := true
	for running {
		select {
		case <-waitCh:
			running = false
		case messages := <-failedMessages:
			for messageId := range messages {
				fmt.Printf("Failed to re-send message [%s].\n", messageId)
			}
		}
	}
}

func requeue(ctx context.Context, options options, client *sqs.SQS, wg *sync.WaitGroup, failedCh chan map[string]string) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			receiptHandles := make(map[string]string)
			receiveMessageInput := sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(options.SourceQueue),
				WaitTimeSeconds:     aws.Int64(options.WaitTimeSeconds),
				VisibilityTimeout:   aws.Int64(options.VisibilityTimeout),
				MaxNumberOfMessages: aws.Int64(options.MaxNumberOfMessages),
				AttributeNames: []*string{
					aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
				},
				MessageAttributeNames: []*string{
					aws.String(sqs.QueueAttributeNameAll),
				},
			}

			receiveMessageOutput, err := client.ReceiveMessageWithContext(ctx, &receiveMessageInput)
			if err != nil {
				panic(err)
			}
			if len(receiveMessageOutput.Messages) == 0 {
				return
			}

			var sendMessageBatchRequestEntries []*sqs.SendMessageBatchRequestEntry
			for _, message := range receiveMessageOutput.Messages {
				receiptHandles[*message.MessageId] = *message.ReceiptHandle
				sendMessageBatchRequestEntries = append(sendMessageBatchRequestEntries, &sqs.SendMessageBatchRequestEntry{
					Id:          message.MessageId,
					MessageBody: message.Body})
			}
			sendMessageBatchInput := sqs.SendMessageBatchInput{
				Entries:  sendMessageBatchRequestEntries,
				QueueUrl: aws.String(options.TargetQueue),
			}
			sendMessageBatchOutput, err := client.SendMessageBatchWithContext(ctx, &sendMessageBatchInput)

			var deleteMessageBatchRequestEntries []*sqs.DeleteMessageBatchRequestEntry
			for _, success := range sendMessageBatchOutput.Successful {
				deleteMessageBatchRequestEntry := sqs.DeleteMessageBatchRequestEntry{
					Id:            success.Id,
					ReceiptHandle: aws.String(receiptHandles[*success.Id]),
				}
				deleteMessageBatchRequestEntries = append(deleteMessageBatchRequestEntries, &deleteMessageBatchRequestEntry)
			}
			deleteMessageBatchInput := sqs.DeleteMessageBatchInput{
				QueueUrl: aws.String(options.SourceQueue),
				Entries:  deleteMessageBatchRequestEntries,
			}
			deleteMessageBatchOutput, err := client.DeleteMessageBatchWithContext(ctx, &deleteMessageBatchInput)
			for _, x := range deleteMessageBatchOutput.Successful {
				delete(receiptHandles, *x.Id)
			}

			if len(receiptHandles) > 0 {
				failedCh <- receiptHandles
			}
		}
	}
}
