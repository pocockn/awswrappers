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
		mockBatchGetItemPages func(*dynamoDBLib.BatchGetItemInput, func(page *dynamoDBLib.BatchGetItemOutput, lastPage bool) bool) error
		mockScanPages         func(*dynamoDBLib.ScanInput, func(*dynamoDBLib.ScanOutput, bool) bool) error
	}

	TestModel struct {
		Foo string `dynamodbav:"foo"`
	}

	TestBatchGetModel struct {
		ID string `dynamodbav:"id"`
	}
)

func (m MockSDKClient) BatchGetItemPages(batchGetItem *dynamoDBLib.BatchGetItemInput, fn func(page *dynamoDBLib.BatchGetItemOutput, lastPage bool) bool) error {
	if m.mockBatchGetItemPages != nil {
		return m.mockBatchGetItemPages(batchGetItem, fn)
	}

	return nil
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

	t.Run(".BatchGetItem()", func(t *testing.T) {
		tableName := "test_table_name"

		data := dynamodb.BatchGetItem{
			"id": []interface{}{
				"some_value",
				"some_other_value",
			},
		}

		t.Run("ReturnsSliceOfAttributeValues", func(t *testing.T) {
			attributeKeyValues := make(map[string][]interface{})
			attributeKeyValues["id"] = append(
				attributeKeyValues["id"],
				"some_value",
			)
			attributeKeyValues["id"] = append(
				attributeKeyValues["id"],
				"some_other_value",
			)

			batchItemOutput := buildBatchGetItemOutput(tableName, attributeKeyValues)

			mockSDKClient := &MockSDKClient{
				mockBatchGetItemPages: func(input *dynamoDBLib.BatchGetItemInput, paginationFunction func(page *dynamoDBLib.BatchGetItemOutput, lastPage bool) bool) error {
					requestItems := input.RequestItems

					assert.Contains(t, requestItems, tableName)
					assert.Len(t, requestItems[tableName].Keys, 2)
					assert.Contains(t, requestItems[tableName].Keys[0], "id")
					assert.Contains(t, requestItems[tableName].Keys[1], "id")

					firstItem := requestItems[tableName].Keys[0]["id"].S
					secondItem := requestItems[tableName].Keys[1]["id"].S
					assert.Equal(t, "some_value", *firstItem)
					assert.Equal(t, "some_other_value", *secondItem)

					paginationFunction(&batchItemOutput, true)
					return nil
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			var bindModel []TestBatchGetModel

			err = testClient.BatchGetItem(tableName, data, &bindModel)
			assert.NoError(t, err)

			assert.Len(t, bindModel, 2)
			assert.Equal(t, "some_value", bindModel[0].ID)
			assert.Equal(t, "some_other_value", bindModel[1].ID)
		})

		t.Run("ReturnsSliceOfAttributeValuesGeneratingMultipleRequests", func(t *testing.T) {
			attributeKeyValues := make(map[string][]interface{})
			newData := dynamodb.BatchGetItem{
				"id": []interface{}{},
			}

			for i := 0; i < 105; i++ {
				attributeKeyValues["id"] = append(
					attributeKeyValues["id"],
					fmt.Sprintf("%d_some_value", i),
				)
				newData["id"] = append(newData["id"], fmt.Sprintf("%d_some_value", i))
			}

			batchItemOutput := buildBatchGetItemOutput(tableName, attributeKeyValues)
			clientCallCount := 0

			mockSDKClient := &MockSDKClient{
				mockBatchGetItemPages: func(input *dynamoDBLib.BatchGetItemInput, paginationFunction func(page *dynamoDBLib.BatchGetItemOutput, lastPage bool) bool) error {
					requestItems := input.RequestItems

					assert.Contains(t, requestItems, tableName)
					if clientCallCount == 0 {
						assert.Len(t, requestItems[tableName].Keys, 100)
					} else {
						assert.Len(t, requestItems[tableName].Keys, 5)
					}

					assert.Contains(t, requestItems[tableName].Keys[0], "id")
					assert.Contains(t, requestItems[tableName].Keys[1], "id")

					firstItem := requestItems[tableName].Keys[0]["id"].S
					secondItem := requestItems[tableName].Keys[1]["id"].S
					if clientCallCount == 0 {
						assert.Equal(t, "0_some_value", *firstItem)
						assert.Equal(t, "1_some_value", *secondItem)
					} else {
						assert.Equal(t, "100_some_value", *firstItem)
						assert.Equal(t, "101_some_value", *secondItem)
					}

					paginationFunction(&batchItemOutput, true)
					clientCallCount++
					return nil
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			var bindModel []TestBatchGetModel

			err = testClient.BatchGetItem(tableName, newData, &bindModel)
			assert.NoError(t, err)

			assert.Len(t, bindModel, 105)
			assert.Equal(t, "0_some_value", bindModel[0].ID)
			assert.Equal(t, "1_some_value", bindModel[1].ID)
			assert.Equal(t, 2, clientCallCount)
		})

		t.Run("BatchGetItemPagination", func(t *testing.T) {
			attributeKeyValuesPage1 := make(map[string][]interface{})
			attributeKeyValuesPage1["id"] = append(
				attributeKeyValuesPage1["id"],
				"some_value",
			)
			batchItemOutputPage1 := buildBatchGetItemOutput(tableName, attributeKeyValuesPage1)

			attributeKeyValuesPage2 := make(map[string][]interface{})
			attributeKeyValuesPage2["id"] = append(
				attributeKeyValuesPage2["id"],
				"some_other_value",
			)
			batchItemOutputPage2 := buildBatchGetItemOutput(tableName, attributeKeyValuesPage2)

			var paginationCount int
			mockSDKClient := &MockSDKClient{
				mockBatchGetItemPages: func(input *dynamoDBLib.BatchGetItemInput, paginationFunction func(page *dynamoDBLib.BatchGetItemOutput, lastPage bool) bool) error {
					paginationFunction(&batchItemOutputPage1, false)
					paginationCount++
					paginationFunction(&batchItemOutputPage2, true)
					paginationCount++
					return nil
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			var bindModel []TestBatchGetModel

			err = testClient.BatchGetItem(tableName, data, &bindModel)
			assert.NoError(t, err)
			assert.Equal(t, "some_value", bindModel[0].ID)
			assert.Equal(t, "some_other_value", bindModel[1].ID)

			assert.Equal(t, paginationCount, 2)

		})

		t.Run("BatchGetItemReturnsOnError", func(t *testing.T) {
			mockSDKClient := &MockSDKClient{
				mockBatchGetItemPages: func(input *dynamoDBLib.BatchGetItemInput, paginationFunction func(page *dynamoDBLib.BatchGetItemOutput, lastPage bool) bool) error {
					return errors.New("Batch get item pages error.")
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			err = testClient.BatchGetItem(tableName, data, nil)
			assert.Error(t, err, "Batch get item pages error.")
		})
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
