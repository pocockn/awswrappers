package dynamodb

import (
	"errors"
	"net"
	"net/url"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	dynamoDBLib "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/vidsy/backoff"
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

var (
	developmentBackoffIntervals = []int{0, 500, 1000, 2000, 4000, 8000, 16000, 32000}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, isDevelopment bool, developmentLogMessageHandler func(string), client dynamodbiface.DynamoDBAPI) (*Client, error) {
	if client == nil {
		client = dynamoDBLib.New(session.New())

		if isDevelopment {
			err := testConnection(config.DynamoDBEndpoint, developmentBackoffIntervals, developmentLogMessageHandler)
			if err != nil {
				return nil, err
			}

			client = dynamoDBLib.New(
				session.New(),
				aws.NewConfig().WithEndpoint(config.DynamoDBEndpoint),
			)
		}
	}

	return &Client{
		client,
		config,
	}, nil
}

func testConnection(endpoint string, backoffIntervals []int, developmentLogMessageHandler func(string)) error {
	bp := backoff.Policy{
		Intervals:         backoffIntervals,
		LogMessageHandler: developmentLogMessageHandler,
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

// PutItem extends the default clients PutItem taking a struct that implements
// the marshaler interface.
func (c Client) PutItem(item Marshaler) (*dynamoDBLib.PutItemOutput, error) {
	putItemInput, err := item.Marshal()
	if err != nil {
		return nil, err
	}

	return c.DynamoDBAPI.PutItem(putItemInput)
}

// Query extends the default clients Query and takes the query params and
// struct to unmarshal the data into.
func (c Client) Query(input *dynamoDBLib.QueryInput, bindModel interface{}) (*dynamoDBLib.QueryOutput, error) {
	output, err := c.DynamoDBAPI.Query(input)
	if err != nil {
		return output, err
	}

	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &bindModel)
	if err != nil {
		return output, err
	}

	return output, nil
}
