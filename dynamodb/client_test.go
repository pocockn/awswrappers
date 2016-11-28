package dynamodb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/vidsy/awswrappers/dynamodb"
)

type (
	MockSDKClient struct {
		dynamodbiface.DynamoDBAPI
	}
)

func NewTestClient(mockClient *MockSDKClient) (*dynamodb.Client, error) {
	if mockClient == nil {
		mockClient = &MockSDKClient{}
	}

	config := dynamodb.ClientConfig{
		DynamoDBEndpoint: "http://www.test.com",
	}

	return dynamodb.NewClient(&config, "test", nil, nil, mockClient)
}

func TestClient(t *testing.T) {
	t.Run("...", func(t *testing.T) {
		// No functions to test.
	})
}
