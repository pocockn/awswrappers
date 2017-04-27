package s3_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	s3Lib "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/vidsy/awswrappers/s3"
)

type (
	MockSDKClient struct {
		s3iface.S3API

		mockPutObjectRequest func(input *s3Lib.PutObjectInput) (*request.Request, *s3Lib.PutObjectOutput)
	}
)

func (m MockSDKClient) PutObjectRequest(input *s3Lib.PutObjectInput) (*request.Request, *s3Lib.PutObjectOutput) {
	if m.mockPutObjectRequest != nil {
		return m.mockPutObjectRequest(input)
	}

	return nil, nil
}

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
	t.Run(".PresignedURI()", func(t *testing.T) {
		t.Run("ReturnsPresignedURI", func(t *testing.T) {
			testUrl, _ := url.Parse("http://www.foo.com")
			mockClient := &MockSDKClient{
				mockPutObjectRequest: func(input *s3Lib.PutObjectInput) (*request.Request, *s3Lib.PutObjectOutput) {
					return &request.Request{
						HTTPRequest: &http.Request{
							URL: testUrl,
						},
					}, nil
				},
			}

			client := NewTestClient(mockClient)
			uri, err := client.PresignedURI("bop", "bip", 10*time.Second)

			if err != nil {
				t.Fatalf("Expected .PresignedURI() to not return an error, got: '%s'", err)
			}

			if uri != testUrl.String() {
				t.Fatalf("Expected .PresignedURI() to return '%s', got: '%s'", testUrl.String(), uri)
			}
		})
	})
}
