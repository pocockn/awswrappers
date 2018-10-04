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
		mockBatchGetItem func(*dynamoDBLib.BatchGetItemInput) (*dynamoDBLib.BatchGetItemOutput, error)
		mockScanPages    func(*dynamoDBLib.ScanInput, func(*dynamoDBLib.ScanOutput, bool) bool) error
	}

	TestModel struct {
		Foo string `dynamodbav:"foo"`
	}
)

func (m MockSDKClient) BatchGetItem(batchGetItem *dynamoDBLib.BatchGetItemInput) (*dynamoDBLib.BatchGetItemOutput, error) {
	if m.mockBatchGetItem != nil {
		return m.mockBatchGetItem(batchGetItem)
	}

	return nil, nil
}

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
					fooValue := fmt.Sprintf("foo_%d", input.Segment)

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

	t.Run("ReturnsSliceOfAttributeValues", func(t *testing.T) {
		tableName := "foo"

		attributeKeyValues := make(map[string][]interface{})
		attributeKeyValues[tableName] = append(
			attributeKeyValues[tableName],
			"some_value",
			"some_other_value",
		)

		batchItemOutput := buildBatchGetItemOutput(tableName, attributeKeyValues)

		data := dynamodb.BatchGetItem{
			"some_key": []interface{}{
				"some_value",
				"some_other_value",
			},
		}

		mockSDKClient := &MockSDKClient{
			mockBatchGetItem: func(input *dynamoDBLib.BatchGetItemInput) (*dynamoDBLib.BatchGetItemOutput, error) {
				requestItems := input.RequestItems

				assert.Contains(t, requestItems, "foo")
				assert.Len(t, requestItems["foo"].Keys, 2)
				assert.Contains(t, requestItems["foo"].Keys[0], "some_key")
				assert.Contains(t, requestItems["foo"].Keys[1], "some_key")

				firstItem := requestItems["foo"].Keys[0]["some_key"].S
				secondItem := requestItems["foo"].Keys[1]["some_key"].S
				assert.Equal(t, "some_value", *firstItem)
				assert.Equal(t, "some_other_value", *secondItem)

				return &batchItemOutput, nil
			},
		}

		testClient, err := NewTestClient(mockSDKClient)
		assert.Nil(t, err)

		output, err := testClient.BatchGetItem(tableName, data)

		assert.NoError(t, err)
		assert.Equal(t, output.GoString(), batchItemOutput.GoString())
	})
}

func buildBatchGetItemOutput(tableName string, attributeKeyValues map[string][]interface{}) dynamoDBLib.BatchGetItemOutput {
	responses := make(map[string][]map[string]*dynamoDBLib.AttributeValue)
	responses[tableName] = make([]map[string]*dynamoDBLib.AttributeValue, 0)

	for key, values := range attributeKeyValues {
		for _, value := range values {
			responses[tableName] = append(
				responses[tableName],
				map[string]*dynamoDBLib.AttributeValue{
					key: &dynamoDBLib.AttributeValue{
						S: aws.String(value.(string)),
					},
				},
			)
		}
	}
	batchItemOutput := dynamoDBLib.BatchGetItemOutput{
		Responses: responses,
	}

	return batchItemOutput
}
