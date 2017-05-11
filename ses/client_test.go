package ses_test

import (
	"testing"

	"errors"
	"github.com/aws/aws-sdk-go/aws"
	sesLib "github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
	"github.com/vidsy/awswrappers/ses"
)

type (
	MockSDKClient struct {
		sesiface.SESAPI

		mockSendEmail func(input *sesLib.SendEmailInput) (*sesLib.SendEmailOutput, error)
	}
)

func (m MockSDKClient) SendEmail(input *sesLib.SendEmailInput) (*sesLib.SendEmailOutput, error) {
	if m.mockSendEmail != nil {
		return m.mockSendEmail(input)
	}

	return nil, nil
}

func NewTestClient(mockClient *MockSDKClient) *ses.Client {
	if mockClient == nil {
		mockClient = &MockSDKClient{}
	}

	return ses.NewClient(mockClient)
}

func TestClient(t *testing.T) {
	t.Run(".SendEmailMessage()", func(t *testing.T) {
		t.Run("ReturnsMessageID", func(t *testing.T) {
			messageID := "123"
			mockClient := &MockSDKClient{
				mockSendEmail: func(input *sesLib.SendEmailInput) (*sesLib.SendEmailOutput, error) {
					return &sesLib.SendEmailOutput{
						MessageId: aws.String(messageID),
					}, nil
				},
			}

			client := NewTestClient(mockClient)
			sendEmailID, err := client.SendEmailMessage(
				[]string{"test@test.com"},
				"hello@test.com",
				"An email",
				"Plain body",
				"<b>Html body</b>",
			)

			if err != nil {
				t.Fatalf("Expected .SendEmailMessage() to not return an error, got: '%s'", err)
			}

			if sendEmailID != messageID {
				t.Fatalf("Expected .SendEmailMessage() to return '%s', got: '%s'", messageID, sendEmailID)
			}
		})

		t.Run("ReturnsError", func(t *testing.T) {
			mockClient := &MockSDKClient{
				mockSendEmail: func(input *sesLib.SendEmailInput) (*sesLib.SendEmailOutput, error) {
					return nil, errors.New("SendEmail error")
				},
			}

			client := NewTestClient(mockClient)
			_, err := client.SendEmailMessage(
				[]string{"test@test.com"},
				"hello@test.com",
				"An email",
				"Plain body",
				"<b>Html body</b>",
			)

			if err == nil {
				t.Fatal("Expected .Publish() to return an error", err)
			}
		})
	})
}
