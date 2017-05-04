package sns_test

import (
	"testing"

	"errors"
	"github.com/aws/aws-sdk-go/aws"
	snsLib "github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/vidsy/awswrappers/sns"
)

type (
	MockSDKClient struct {
		snsiface.SNSAPI

		mockPublish func(input *snsLib.PublishInput) (*snsLib.PublishOutput, error)
	}
)

func (m MockSDKClient) Publish(input *snsLib.PublishInput) (*snsLib.PublishOutput, error) {
	if m.mockPublish != nil {
		return m.mockPublish(input)
	}

	return nil, nil
}

func NewTestClient(mockClient *MockSDKClient) *sns.Client {
	if mockClient == nil {
		mockClient = &MockSDKClient{}
	}

	config := sns.ClientConfig{
		Endpoint: "http://www.test.com",
	}

	return sns.NewClient(&config, true, mockClient)
}

func TestClient(t *testing.T) {
	t.Run(".Publish()", func(t *testing.T) {
		t.Run("ReturnsMessageID", func(t *testing.T) {
			messageID := "123"
			mockClient := &MockSDKClient{
				mockPublish: func(input *snsLib.PublishInput) (*snsLib.PublishOutput, error) {
					return &snsLib.PublishOutput{
						MessageId: aws.String(messageID),
					}, nil
				},
			}

			client := NewTestClient(mockClient)
			publishMessageID, err := client.SendSMSMessage("0123456789", "Hey, this is a test message!")

			if err != nil {
				t.Fatalf("Expected .Publish() to not return an error, got: '%s'", err)
			}

			if publishMessageID != messageID {
				t.Fatalf("Expected .Publish() to return '%s', got: '%s'", messageID, publishMessageID)
			}
		})

		t.Run("ReturnsError", func(t *testing.T) {
			mockClient := &MockSDKClient{
				mockPublish: func(input *snsLib.PublishInput) (*snsLib.PublishOutput, error) {
					return nil, errors.New("Publish error")
				},
			}

			client := NewTestClient(mockClient)
			_, err := client.SendSMSMessage("0123456789", "Hey, this is a test message!")

			if err == nil {
				t.Fatal("Expected .Publish() to return an error", err)
			}
		})
	})
}
