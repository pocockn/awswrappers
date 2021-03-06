package sns_test

import (
	"testing"

	"errors"

	"github.com/aws/aws-sdk-go/aws"
	snsLib "github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"github.com/stretchr/testify/assert"
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
	t.Run(".SendSMSMessage()", func(t *testing.T) {
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
			publishMessageID, err := client.SendSMSMessage(
				"0123456789", "Foo", sns.PromotionalSMSType, "Hey, this is a test message!",
			)

			assert.NoError(t, err)
			assert.Equal(t, messageID, publishMessageID)
		})

		t.Run("ReturnsError", func(t *testing.T) {
			mockClient := &MockSDKClient{
				mockPublish: func(input *snsLib.PublishInput) (*snsLib.PublishOutput, error) {
					return nil, errors.New("Publish error")
				},
			}

			client := NewTestClient(mockClient)
			_, err := client.SendSMSMessage(
				"0123456789", "Foo", sns.PromotionalSMSType, "Hey, this is a test message!",
			)

			assert.Error(t, err)
		})
	})

	t.Run(".PublishMessage()", func(t *testing.T) {
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
			publishMessageID, err := client.PublishMessage(
				`{"foo":"bar"}`, "8da92fa4-3913-4300-9f3b-31de66e27a97",
			)

			assert.NoError(t, err)
			assert.Equal(t, messageID, publishMessageID)
		})

		t.Run("ReturnsError", func(t *testing.T) {
			mockClient := &MockSDKClient{
				mockPublish: func(input *snsLib.PublishInput) (*snsLib.PublishOutput, error) {
					return nil, errors.New("Publish error")
				},
			}

			client := NewTestClient(mockClient)
			_, err := client.PublishMessage(
				`{"foo":"bar"}`, "8da92fa4-3913-4300-9f3b-31de66e27a97",
			)

			assert.Error(t, err)
		})
	})
}
