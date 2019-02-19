package ses

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	sesLib "github.com/aws/aws-sdk-go/service/ses"
	"github.com/aws/aws-sdk-go/service/ses/sesiface"
)

type (
	// Client wraps the receive and delete functionality of ses.
	Client struct {
		sesiface.SESAPI
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(client sesiface.SESAPI) *Client {
	if client == nil {
		var sesClient *sesLib.SES

		sesClient = sesLib.New(session.New())

		return &Client{
			sesClient,
		}
	}

	return &Client{
		client,
	}
}

// SendEmailMessage sends an email to the given recipient(s) and returns the message
// ID.
func (c Client) SendEmailMessage(recipients []string, from string, subject string, plainBody string, htmlBody string, replyTo string) (string, error) {
	destination := &sesLib.Destination{
		ToAddresses: aws.StringSlice(recipients),
	}

	message := &sesLib.Message{
		Subject: &sesLib.Content{
			Charset: aws.String("UTF-8"),
			Data:    aws.String(subject),
		},
		Body: &sesLib.Body{
			Text: &sesLib.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(plainBody),
			},
			Html: &sesLib.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(htmlBody),
			},
		},
	}

	params := &sesLib.SendEmailInput{
		Destination:      destination,
		Message:          message,
		ReplyToAddresses: aws.StringSlice([]string{replyTo}),
		Source:           aws.String(from),
	}

	response, err := c.SendEmail(params)
	if err != nil {
		return "", err
	}

	return *response.MessageId, nil
}
