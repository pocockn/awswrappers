package dynamodb_test

import (
	"errors"
	"fmt"
	"math/rand"
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

	TestBatchGetModel struct {
		ID string `dynamodbav:"id"`
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
				mockBatchGetItem: func(input *dynamoDBLib.BatchGetItemInput) (*dynamoDBLib.BatchGetItemOutput, error) {
					return &batchItemOutput, nil
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

		t.Run("BatchGetItemPagination", func(t *testing.T) {
			sizeOfData := 117
			newData := generateBatchGetItemData(sizeOfData)

			attributeKeyValuesPage1 := make(map[string][]interface{})
			for _, v := range newData["id"][0:100] {
				attributeKeyValuesPage1["id"] = append(
					attributeKeyValuesPage1["id"], v,
				)
			}
			batchItemOutputPage1 := buildBatchGetItemOutput(
				tableName, attributeKeyValuesPage1,
			)

			attributeKeyValuesPage2 := make(map[string][]interface{})
			for _, v := range newData["id"][100:sizeOfData] {
				attributeKeyValuesPage2["id"] = append(
					attributeKeyValuesPage2["id"], v,
				)
			}
			batchItemOutputPage2 := buildBatchGetItemOutput(
				tableName, attributeKeyValuesPage2,
			)

			callCount := 0
			mockSDKClient := &MockSDKClient{
				mockBatchGetItem: func(input *dynamoDBLib.BatchGetItemInput) (*dynamoDBLib.BatchGetItemOutput, error) {
					callCount++

					if len(input.RequestItems[tableName].Keys) == 100 {
						return &batchItemOutputPage1, nil
					}

					return &batchItemOutputPage2, nil
				},
			}

			testClient, err := NewTestClient(mockSDKClient)
			assert.Nil(t, err)

			var bindModel []TestBatchGetModel

			err = testClient.BatchGetItem(tableName, newData, &bindModel)
			assert.NoError(t, err)

			assert.Len(t, bindModel, sizeOfData)
			assert.Equal(t, newData["id"][0], bindModel[0].ID)
			assert.Equal(t, newData["id"][1], bindModel[1].ID)

			assert.Equal(t, callCount, 2)
		})

		t.Run("BatchGetItemReturnsOnError", func(t *testing.T) {
			mockSDKClient := &MockSDKClient{
				mockBatchGetItem: func(input *dynamoDBLib.BatchGetItemInput) (*dynamoDBLib.BatchGetItemOutput, error) {
					return nil, errors.New("Batch get item pages error.")
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

func generateBatchGetItemData(length int) dynamodb.BatchGetItem {
	data := dynamodb.BatchGetItem{
		"id": []interface{}{},
	}

	for i := 0; i < length; i++ {
		data["id"] = append(data["id"], randStringRunes(18))
	}

	return data
}

func randStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
