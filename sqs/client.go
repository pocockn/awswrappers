package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	sqsLib "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	"fmt"
	"log"
)

type (
	// ClientWrapper interface for client wrapping sqs.
	ClientWrapper interface {
		ReceiveMessage() (*sqsLib.Message, error)
		DeleteMessage(receiptHandle *string) error
	}

	// Client wraps the receive and delete functionality of SQS.
	Client struct {
		sqsiface.SQSAPI
		clientConfig *ClientConfig
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, environment string) *Client {
	var sqsClient *sqsLib.SQS

	if environment == "development" {
		log.Println("Creating development SQS client")
		sqsClient = sqsLib.New(session.New(), aws.NewConfig().WithEndpoint(config.QueueEndpoint))
	} else {
		sqsClient = sqsLib.New(session.New())
	}

	return &Client{
		sqsClient,
		config,
	}
}

// ReceiveMessage Returns a message from SQS.
func (s Client) ReceiveMessage() (*sqsLib.Message, error) {
	resp, err := s.SQSAPI.ReceiveMessage(s.receiveMessageParams())
	if err != nil {
		return nil, err
	}

	if len(resp.Messages) > 0 {
		return resp.Messages[0], nil
	}

	return nil, nil
}

// DeleteMessage removes a message based on the recipt handle.
func (s Client) DeleteMessage(receiptHandle *string) error {
	_, err := s.SQSAPI.DeleteMessage(s.deleteMessageParams(receiptHandle))
	if err != nil {
		return err
	}

	return nil
}

func (s Client) queueURL() string {
	return fmt.Sprintf("%s/%s", s.clientConfig.QueueEndpoint, s.clientConfig.QueueName)
}

func (s Client) deleteMessageParams(receiptHandle *string) *sqsLib.DeleteMessageInput {
	return &sqsLib.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL()),
		ReceiptHandle: receiptHandle,
	}
}

func (s Client) receiveMessageParams() *sqsLib.ReceiveMessageInput {
	return &sqsLib.ReceiveMessageInput{
		QueueUrl: aws.String(s.queueURL()),
		AttributeNames: []*string{
			aws.String("All"),
		},
		MaxNumberOfMessages: aws.Int64(s.clientConfig.MaxNumberOfMessages),
		MessageAttributeNames: []*string{
			aws.String("All"),
		},
		VisibilityTimeout: aws.Int64(s.clientConfig.VisibilityTimeout),
		WaitTimeSeconds:   aws.Int64(s.clientConfig.WaitTimeSeconds),
	}
}
