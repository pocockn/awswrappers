package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	sqsLib "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"

	"fmt"
)

type (
	// ClientWrapper interface for client wrapping sqs.
	ClientWrapper interface {
		DeleteMessage(queueName string, receiptHandle *string) error
		ReceiveMessage(queueName string) (*sqsLib.Message, error)
		SendNewMessage(queueName string, body []byte) (string, error)
	}

	// Client wraps the receive and delete functionality of SQS.
	Client struct {
		sqsiface.SQSAPI
		clientConfig *ClientConfig
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, useDevelopmentClient bool, client sqsiface.SQSAPI) *Client {
	if client == nil {
		var sqsClient *sqsLib.SQS

		if useDevelopmentClient {
			sqsClient = sqsLib.New(session.New(), aws.NewConfig().WithEndpoint(config.QueueEndpoint))
		} else {
			sqsClient = sqsLib.New(session.New())
		}

		return &Client{
			sqsClient,
			config,
		}
	}

	return &Client{
		client,
		config,
	}
}

// ReceiveMessage Returns a message from SQS.
func (s Client) ReceiveMessage(queueName string) (*sqsLib.Message, error) {
	resp, err := s.SQSAPI.ReceiveMessage(s.receiveMessageParams(queueName))
	if err != nil {
		return nil, err
	}

	if len(resp.Messages) > 0 {
		return resp.Messages[0], nil
	}

	return nil, nil
}

// SendNewMessage sends an SQS message on the given queue.
func (s Client) SendNewMessage(queueName string, body []byte) (string, error) {
	params := &sqsLib.SendMessageInput{
		MessageBody: aws.String(string(body[:])),
		QueueUrl:    aws.String(s.queueURL(queueName)),
	}

	resp, err := s.SQSAPI.SendMessage(params)
	if err != nil {
		return "", err
	}

	return *resp.MessageId, nil
}

// SendNewFIFOMessage sends an SQS message on the given FIFO queue.
func (s Client) SendNewFIFOMessage(queueName string, body []byte, deduplicationID string, groupID string, messageAttributes map[string]*sqsLib.MessageAttributeValue) (string, error) {
	params := &sqsLib.SendMessageInput{
		MessageBody:            aws.String(string(body[:])),
		MessageDeduplicationId: aws.String(deduplicationID),
		MessageGroupId:         aws.String(groupID),
		QueueUrl:               aws.String(s.queueURL(queueName)),
	}

	if messageAttributes != nil {
		params.MessageAttributes = messageAttributes
	}

	resp, err := s.SQSAPI.SendMessage(params)
	if err != nil {
		return "", err
	}

	return *resp.MessageId, nil
}

// DeleteMessage removes a message based on the recipt handle.
func (s Client) DeleteMessage(queueName string, receiptHandle *string) error {
	_, err := s.SQSAPI.DeleteMessage(s.deleteMessageParams(queueName, receiptHandle))
	if err != nil {
		return err
	}

	return nil
}

func (s Client) queueURL(queueName string) string {
	return fmt.Sprintf("%s/%s", s.clientConfig.QueueEndpoint, queueName)
}

func (s Client) deleteMessageParams(queueName string, receiptHandle *string) *sqsLib.DeleteMessageInput {
	return &sqsLib.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL(queueName)),
		ReceiptHandle: receiptHandle,
	}
}

func (s Client) receiveMessageParams(queueName string) *sqsLib.ReceiveMessageInput {
	return &sqsLib.ReceiveMessageInput{
		QueueUrl: aws.String(s.queueURL(queueName)),
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
