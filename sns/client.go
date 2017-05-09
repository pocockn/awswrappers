package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	snsLib "github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"

	"log"
)

const (
	// PromotionalSMSType for SMS messages that are non
	// critical such as promotional.
	PromotionalSMSType = "Promotional"

	// TransactionalSMSType for SMS messages that are
	// critical to user such as MFA tokens.
	TransactionalSMSType = "Transactional"
)

type (
	// Client wraps the receive and delete functionality of sns.
	Client struct {
		snsiface.SNSAPI
		clientConfig *ClientConfig
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, useDevelopmentClient bool, client snsiface.SNSAPI) *Client {
	if client == nil {
		var snsClient *snsLib.SNS

		if useDevelopmentClient {
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
func (c Client) SendSMSMessage(number string, from string, messageType string, message string) (string, error) {
	messageAttributes := map[string]*snsLib.MessageAttributeValue{
		"AWS.SNS.SMS.SenderID": &snsLib.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(from),
		},
		"AWS.SNS.SMS.SMSType": &snsLib.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(messageType),
		},
	}

	params := &snsLib.PublishInput{
		Message:           aws.String(message),
		MessageAttributes: messageAttributes,
		PhoneNumber:       aws.String(number),
	}

	response, err := c.Publish(params)
	if err != nil {
		return "", err
	}

	return *response.MessageId, nil
}
