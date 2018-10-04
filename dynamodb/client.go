package dynamodb

import (
	"fmt"
	"net"
	"net/url"
	"runtime"
	"sync"

	"github.com/pkg/errors"

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

// BatchGetItem extends the default clients BatchGetItem.
func (c Client) BatchGetItem(tableName string, batchGetItem BatchGetItem) (*dynamoDBLib.BatchGetItemOutput, error) {
	attributeValues := marshalValuesIntoAttributeValues(batchGetItem)

	batchGetItemInput := &dynamoDBLib.BatchGetItemInput{
		RequestItems: map[string]*dynamoDBLib.KeysAndAttributes{
			tableName: {
				Keys: attributeValues,
			},
		},
	}

	return c.DynamoDBAPI.BatchGetItem(batchGetItemInput)
}

// DeleteItem extends the default clients DeleteItem taking a struct that implements
// the Deletable interface.
func (c Client) DeleteItem(item Deletable) (*dynamoDBLib.DeleteItemOutput, error) {
	key, err := dynamodbattribute.MarshalMap(item.Key())
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"Problem marshaling key:%s to AttributeValue.",
			key,
		)
	}

	deleteItemInput := &dynamoDBLib.DeleteItemInput{
		Key:       key,
		TableName: aws.String(item.TableName()),
	}

	return c.DynamoDBAPI.DeleteItem(deleteItemInput)
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

// Scan extends the underlying scan with pages functionality and automatically
// creates a set of parellel requests and binds the result to the given struct
// or returns an error.
func (c Client) Scan(params dynamoDBLib.ScanInput, bindModel interface{}) error {
	errChan := make(chan error)
	itemsChan := make(chan map[string]*dynamoDBLib.AttributeValue)
	items := []map[string]*dynamoDBLib.AttributeValue{}
	var scanQueriesWaitGroup sync.WaitGroup

	if *params.TotalSegments == 0 {
		params.TotalSegments = aws.Int64(int64(runtime.NumCPU()))
	}

	scanQueriesWaitGroup.Add(int(*params.TotalSegments))

	for i := int64(0); i < *params.TotalSegments; i++ {
		go c.scanWorker(params, itemsChan, errChan, i, &scanQueriesWaitGroup)
	}

	go func(errChan chan error) {
		scanQueriesWaitGroup.Wait()
		errChan <- nil
	}(errChan)

	for {
		select {
		case item := <-itemsChan:
			items = append(items, item)
		case err := <-errChan:
			if err != nil {
				return err
			}

			err = dynamodbattribute.UnmarshalListOfMaps(items, &bindModel)
			if err != nil {
				return err
			}

			return nil
		}
	}
}

func (c Client) scanWorker(params dynamoDBLib.ScanInput, itemsChan chan map[string]*dynamoDBLib.AttributeValue, errChan chan error, segment int64, scanQueriesWaitGroup *sync.WaitGroup) {
	defer scanQueriesWaitGroup.Done()
	params.Segment = aws.Int64(segment)

	err := c.DynamoDBAPI.ScanPages(&params, func(result *dynamoDBLib.ScanOutput, lastPage bool) bool {
		for _, item := range result.Items {
			itemsChan <- item
		}

		return lastPage
	})

	if err != nil {
		errChan <- err
	}
}

func marshalValuesIntoAttributeValues(batchGetItem BatchGetItem) []map[string]*dynamoDBLib.AttributeValue {
	var attributeKeyValueSlice []map[string]*dynamoDBLib.AttributeValue

	for primaryKey, attributeValues := range batchGetItem {
		for _, attributeValue := range attributeValues {
			attributeKeyValue := make(map[string]*dynamoDBLib.AttributeValue)
			switch castValue := attributeValue.(type) {
			case int, int8, int16, int32, int64, float32, float64:
				stringNumberValue := fmt.Sprintf("%s", castValue)
				attributeKeyValue[primaryKey] = &dynamoDBLib.AttributeValue{N: aws.String(stringNumberValue)}
			case string:
				attributeKeyValue[primaryKey] = &dynamoDBLib.AttributeValue{S: aws.String(castValue)}
			}

			attributeKeyValueSlice = append(
				attributeKeyValueSlice,
				attributeKeyValue,
			)
		}
	}

	return attributeKeyValueSlice
}
