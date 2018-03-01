package dynamodb

import (
	"errors"
	"net"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	dynamoDBLib "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/vidsy/backoff"

	"log"
)

type (
	// ClientWrapper interface for client wrapping DynamoDB.
	ClientWrapper interface{}

	// Client wraps a selection of functions of the DynamoDB client.
	Client struct {
		dynamodbiface.DynamoDBAPI
		clientConfig *ClientConfig
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, client dynamodbiface.DynamoDBAPI) (*Client, error) {
	log.Println("Creating live DynamoDB client")

	if client == nil {
		var dynamoDBClient *dynamoDBLib.DynamoDB

		dynamoDBClient = dynamoDBLib.New(session.New())

		return &Client{
			dynamoDBClient,
			config,
		}, nil
	}

	return &Client{
		client,
		config,
	}, nil
}

// NewDevelopmentClient acts the same as NewClient() however contains logic to
// backoff the connection to the DynamoDB store.
func NewDevelopmentClient(config *ClientConfig, backoffIntervals *[]int, logPrefix *string) (*Client, error) {
	log.Println("Creating development DynamoDB client")

	err := testConnection(config.DynamoDBEndpoint, *backoffIntervals, *logPrefix)
	if err != nil {
		return nil, err
	}

	dynamoDBClient := dynamoDBLib.New(
		session.New(),
		aws.NewConfig().WithEndpoint(config.DynamoDBEndpoint),
	)

	return &Client{
		dynamoDBClient,
		config,
	}, nil
}

func testConnection(endpoint string, backoffIntervals []int, logPrefix string) error {
	bp := backoff.Policy{
		Intervals: backoffIntervals,
		LogPrefix: logPrefix,
	}

	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	connected, _ := bp.Perform(func() (bool, error) {
		_, err := net.Dial("tcp", parsedURL.Host)
		if err != nil {
			return false, nil
		}

		return true, nil
	})

	if !connected {
		return errors.New("Unable to connect to DynamoDB after backoff")
	}

	return nil
}
