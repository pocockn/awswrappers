package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	snsLib "github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"

	"log"
)

type (
	// Client wraps the receive and delete functionality of sns.
	Client struct {
		snsiface.SNSAPI
		clientConfig *ClientConfig
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, environment string, client snsiface.SNSAPI) *Client {
	if client == nil {
		var snsClient *snsLib.SNS

		if environment == "development" {
			log.Println("Creating development sns client")
			snsClient = snsLib.New(session.New(), aws.NewConfig().WithEndpoint(config.Endpoint))
		} else {
			snsClient = snsLib.New(session.New())
		}

		return &Client{
			snsClient,
			config,
		}
	}

	return &Client{
		client,
		config,
	}
}

// SendSMSMessage sends an SMS message and returns the MessageID.
func (c Client) SendSMSMessage(number string, message string) (string, error) {
	params := &snsLib.PublishInput{
		Message:     aws.String(message),
		PhoneNumber: aws.String(number),
	}

	response, err := c.Publish(params)
	if err != nil {
		return "", err
	}

	return *response.MessageId, nil
}
