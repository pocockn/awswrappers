package dynamodb_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	dynamoDBLib "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/assert"
	"github.com/vidsy/awswrappers/dynamodb"
)

type (
	MockSDKClient struct {
		dynamodbiface.DynamoDBAPI
		mockScanPages func(*dynamoDBLib.ScanInput, func(*dynamoDBLib.ScanOutput, bool) bool) error
	}

	TestModel struct {
		Foo string `dynamodbav:"foo"`
	}
)

func (m MockSDKClient) ScanPages(input *dynamoDBLib.ScanInput, pageFunc func(*dynamoDBLib.ScanOutput, bool) bool) error {
	if m.mockScanPages != nil {
		return m.mockScanPages(input, pageFunc)
	}

	return nil
}

func NewTestClient(mockClient *MockSDKClient) (*dynamodb.Client, error) {
	if mockClient == nil {
		mockClient = &MockSDKClient{}
	}

	config := dynamodb.ClientConfig{
		DynamoDBEndpoint: "http://www.test.com",
	}

	return dynamodb.NewClient(&config, true, nil, mockClient)
}

func TestClient(t *testing.T) {
	t.Run(".Scan()", func(t *testing.T) {
		params := dynamoDBLib.ScanInput{
			ExpressionAttributeValues: map[string]*dynamoDBLib.AttributeValue{
				":sent": {BOOL: aws.Bool(true)},
			},
			FilterExpression: aws.String("sent = :sent"),
			TableName:        aws.String("message_group"),
			TotalSegments:    aws.Int64(4),
		}

		t.Run("GeneratesSegmentedQuery", func(t *testing.T) {
			queryCount := 0
			mockSDKClient := &MockSDKClient{
				mockScanPages: func(input *dynamoDBLib.ScanInput, pageFunc func(*dynamoDBLib.ScanOutput, bool) bool) error {
					queryCount++

					output := &dynamoDBLib.ScanOutput{
						Items: []map[string]*dynamoDBLib.AttributeValue{},
					}

					pageFunc(output, true)
					return nil
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			err = testClient.Scan(params, nil)

			assert.Nil(t, err)
			assert.Equal(t, 4, queryCount)
		})

		t.Run("BindsModelData", func(t *testing.T) {
			mockSDKClient := &MockSDKClient{
				mockScanPages: func(input *dynamoDBLib.ScanInput, pageFunc func(*dynamoDBLib.ScanOutput, bool) bool) error {
					fooValue := fmt.Sprintf("foo_%s", input.Segment)

					output := &dynamoDBLib.ScanOutput{
						Items: []map[string]*dynamoDBLib.AttributeValue{
							map[string]*dynamoDBLib.AttributeValue{
								"foo": &dynamoDBLib.AttributeValue{
									S: &fooValue,
								},
							},
						},
					}

					pageFunc(output, true)
					return nil
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			var testModels []TestModel
			err = testClient.Scan(params, &testModels)

			assert.Nil(t, err)
			assert.Len(t, testModels, 4)
		})

		t.Run("ReturnsOnError", func(t *testing.T) {
			mockSDKClient := &MockSDKClient{
				mockScanPages: func(input *dynamoDBLib.ScanInput, pageFunc func(*dynamoDBLib.ScanOutput, bool) bool) error {
					return errors.New("Query Error")
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			err = testClient.Scan(params, nil)
			assert.NotNil(t, err)
		})
	})
}
