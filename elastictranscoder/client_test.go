package elastictranscoder_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	elastictranscoderLib "github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/aws/aws-sdk-go/service/elastictranscoder/elastictranscoderiface"
	"github.com/vidsy/awswrappers/elastictranscoder"
)

type (
	MockSDKClient struct {
		elastictranscoderiface.ElasticTranscoderAPI

		mockCreateJob func(input *elastictranscoderLib.CreateJobInput) (*elastictranscoderLib.CreateJobResponse, error)
	}
)

func (m MockSDKClient) CreateJob(input *elastictranscoderLib.CreateJobInput) (*elastictranscoderLib.CreateJobResponse, error) {
	if m.mockCreateJob != nil {
		return m.mockCreateJob(input)
	}

	return nil, nil
}

func NewTestClient(mockClient *MockSDKClient) *elastictranscoder.Client {
	if mockClient == nil {
		mockClient = &MockSDKClient{}
	}

	config := elastictranscoder.ClientConfig{
		Endpoint: "http://www.test.com",
	}

	return elastictranscoder.NewClient(&config, true, mockClient)
}

func TestClient(t *testing.T) {
	t.Run(".CreateJob()", func(t *testing.T) {
		jobID := "123456"

		t.Run("ReturnsJobID", func(t *testing.T) {
			mockClient := createMockClient(jobID)
			client := NewTestClient(mockClient)
			clientJobID, err := client.CreateNewJob("pipeline-abc", "input-key", "output-key", "output-preset-id", "output-key-prefix", "", nil)

			if err != nil {
				t.Fatalf("Expected no error, got: %s", err)
			}

			if clientJobID != jobID {
				t.Fatalf(
					"Expected .CreateJob() to return '%s', got: '%s'",
					jobID,
					clientJobID,
				)
			}
		})

		t.Run("ReturnsError", func(t *testing.T) {
		})
	})
}

func createMockClient(jobID string) *MockSDKClient {
	return &MockSDKClient{
		mockCreateJob: func(input *elastictranscoderLib.CreateJobInput) (*elastictranscoderLib.CreateJobResponse, error) {
			return &elastictranscoderLib.CreateJobResponse{
				Job: &elastictranscoderLib.Job{
					Id: aws.String(jobID),
				},
			}, nil
		},
	}
}
