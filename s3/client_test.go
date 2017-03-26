package s3_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/vidsy/awswrappers/s3"
)

type (
	MockSDKClient struct {
		s3iface.S3API
	}
)

func NewTestClient(mockClient *MockSDKClient) *s3.Client {
	if mockClient == nil {
		mockClient = &MockSDKClient{}
	}

	config := s3.ClientConfig{
		Endpoint: "http://www.test.com",
	}

	return s3.NewClient(&config, "development", mockClient)
}

func TestClient(t *testing.T) {
	t.Run("...", func(t *testing.T) {
		// No functions to test.
	})
}
