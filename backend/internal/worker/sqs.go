package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSSource struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSSource(client *sqs.Client, queueURL string) *SQSSource {
	return &SQSSource{client: client, queueURL: queueURL}
}

func (s *SQSSource) Receive(ctx context.Context) ([]RawMessage, error) {
	out, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
	})
	if err != nil {
		return nil, fmt.Errorf("sqs receive: %w", err)
	}
	var msgs []RawMessage
	for _, m := range out.Messages {
		body := ""
		if m.Body != nil {
			body = *m.Body
		}
		handle := ""
		if m.ReceiptHandle != nil {
			handle = *m.ReceiptHandle
		}
		msgs = append(msgs, RawMessage{Body: body, ReceiptHandle: handle})
	}
	return msgs, nil
}

func (s *SQSSource) Delete(ctx context.Context, receiptHandle string) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	return err
}

// SQSPublisher for service layer
type SQSPublisher struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSPublisher(client *sqs.Client, queueURL string) *SQSPublisher {
	return &SQSPublisher{client: client, queueURL: queueURL}
}

func (p *SQSPublisher) Publish(ctx context.Context, msg interface{}) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(b)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"ContentType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("application/json"),
			},
		},
	})
	return err
}
