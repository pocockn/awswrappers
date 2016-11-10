package dynamodb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	dynamoDBLib "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

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
func NewClient(config *ClientConfig, environment string, client dynamodbiface.DynamoDBAPI) *Client {
	if client == nil {
		var dynamoDBClient *dynamoDBLib.DynamoDB

		if environment == "development" {
			log.Println("Creating development DynamoDB client")
			dynamoDBClient = dynamoDBLib.New(
				session.New(),
				aws.NewConfig().WithEndpoint(config.DynamoDBEndpoint),
			)
		} else {
			dynamoDBClient = dynamoDBLib.New(session.New())
		}

		return &Client{
			dynamoDBClient,
			config,
		}
	}

	return &Client{
		client,
		config,
	}
}
