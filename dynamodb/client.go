package dynamodb

import (
	"errors"
	"net"

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
func NewClient(config *ClientConfig, environment string, backoffIntervals *[]int, logPrefix *string, client dynamodbiface.DynamoDBAPI) (*Client, error) {
	if client == nil {
		var dynamoDBClient *dynamoDBLib.DynamoDB

		if environment == "development" || environment == "docker" {
			log.Println("Creating development/docker DynamoDB client")

			err := testConnection(config.DynamoDBEndpoint, *backoffIntervals, *logPrefix)
			if err != nil {
				return nil, err
			}

			dynamoDBClient = dynamoDBLib.New(
				session.New(),
				aws.NewConfig().WithEndpoint(config.DynamoDBEndpoint),
			)
		} else {
			log.Println("Creating live DynamoDB client")
			dynamoDBClient = dynamoDBLib.New(session.New())
		}

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

func testConnection(endpoint string, backoffIntervals []int, logPrefix string) error {
	bp := backoff.Policy{
		Intervals: backoffIntervals,
		LogPrefix: logPrefix,
	}

	connected := bp.Perform(func() bool {
		_, err := net.Dial("tcp", endpoint)
		if err != nil {
			return false
		}

		return true
	})

	if !connected {
		return errors.New("Unable to connect to DynamoDB after backoff.")
	}

	return nil
}
