package sqs_test

import (
	"errors"
	"testing"

	sqsLib "github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/vidsy/awswrappers/sqs"
)

type (
	// MockSDKClient used to mock the client used by the wrapper
	MockSDKClient struct {
		sqsiface.SQSAPI

		mockReceiveMessage func(*sqsLib.ReceiveMessageInput) (*sqsLib.ReceiveMessageOutput, error)
		mockDeleteMessge   func(*sqsLib.DeleteMessageInput) (*sqsLib.DeleteMessageOutput, error)
	}
)

func (smc MockSDKClient) ReceiveMessage(input *sqsLib.ReceiveMessageInput) (*sqsLib.ReceiveMessageOutput, error) {
	if smc.mockReceiveMessage != nil {
		return smc.mockReceiveMessage(input)
	}

	return nil, nil
}

func (smc MockSDKClient) DeleteMessage(input *sqsLib.DeleteMessageInput) (*sqsLib.DeleteMessageOutput, error) {
	if smc.mockDeleteMessge != nil {
		return smc.mockDeleteMessge(input)
	}

	return nil, nil
}

func NewTestClient(mockClient *MockSDKClient) *sqs.Client {
	if mockClient == nil {
		mockClient = &MockSDKClient{}
	}

	config := sqs.ClientConfig{
		QueueEndpoint: "http://www.test.com",
		QueueName:     "queue_name",
	}

	return sqs.NewClient(&config, "test", mockClient)
}

func GenerateMessages(count int) []*sqsLib.Message {
	var messages []*sqsLib.Message
	for i := 0; i < count; i++ {
		messages = append(messages, &sqsLib.Message{})
	}

	return messages
}

func TestClient(t *testing.T) {
	t.Run(".QueueURL", func(t *testing.T) {
		mock := MockSDKClient{}
		client := NewTestClient(&mock)

		expected := "http://www.test.com/queue_name"
		actual := client.QueueURL()

		if actual != expected {
			t.Fatalf("Expected QueueURL to be '%s', got: '%s'", expected, actual)
		}
	})

	t.Run(".ReceiveMessage", func(t *testing.T) {
		t.Run("ClientCalledWithoutError", func(t *testing.T) {
			mock := MockSDKClient{
				mockReceiveMessage: func(input *sqsLib.ReceiveMessageInput) (*sqsLib.ReceiveMessageOutput, error) {
					return &sqsLib.ReceiveMessageOutput{
						Messages: GenerateMessages(1),
					}, nil
				},
			}

			client := NewTestClient(&mock)
			_, err := client.ReceiveMessage()

			if err != nil {
				t.Fatalf("Expected no error to occur, got: %v", err)
			}
		})

		t.Run("MessageReturned", func(t *testing.T) {
			mockMessage := &sqsLib.Message{}
			mock := MockSDKClient{
				mockReceiveMessage: func(input *sqsLib.ReceiveMessageInput) (*sqsLib.ReceiveMessageOutput, error) {
					var messages []*sqsLib.Message
					messages = append(messages, mockMessage)

					return &sqsLib.ReceiveMessageOutput{
						Messages: messages,
					}, nil
				},
			}

			client := NewTestClient(&mock)
			message, _ := client.ReceiveMessage()

			if message != mockMessage {
				t.Fatalf("Expected message to be returned, got: %v", message)
			}
		})

		t.Run("ReturnsNoErrorAndMessage", func(t *testing.T) {
			mock := MockSDKClient{
				mockReceiveMessage: func(input *sqsLib.ReceiveMessageInput) (*sqsLib.ReceiveMessageOutput, error) {
					return &sqsLib.ReceiveMessageOutput{
						Messages: GenerateMessages(0),
					}, nil
				},
			}

			client := NewTestClient(&mock)
			message, err := client.ReceiveMessage()

			if message != nil || err != nil {
				t.Fatalf("Expected no error and no message, got: %v and %v", message, err)
			}
		})
	})

	t.Run(".DeleteMessage", func(t *testing.T) {
		handle := "test_recipt_handle"

		t.Run("ClientCalledWithoutError", func(t *testing.T) {
			client := NewTestClient(nil)
			err := client.DeleteMessage(&handle)

			if err != nil {
				t.Fatalf("Expected no error to occur, got: %v", err)
			}
		})

		t.Run("ClientCalledWithError", func(t *testing.T) {
			mock := MockSDKClient{
				mockDeleteMessge: func(input *sqsLib.DeleteMessageInput) (*sqsLib.DeleteMessageOutput, error) {
					return nil, errors.New("SQS client error")
				},
			}

			client := NewTestClient(&mock)
			err := client.DeleteMessage(&handle)

			if err == nil {
				t.Fatalf("Expected error from client, got: %v", err)
			}
		})
	})
}
