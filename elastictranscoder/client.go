package elastictranscoder

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	elastictranscoderLib "github.com/aws/aws-sdk-go/service/elastictranscoder"
	"github.com/aws/aws-sdk-go/service/elastictranscoder/elastictranscoderiface"
)

type (
	// Client wraps the functionality of elastictranscoder.
	Client struct {
		elastictranscoderiface.ElasticTranscoderAPI
		clientConfig *ClientConfig
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, useDevelopmentClient bool, client elastictranscoderiface.ElasticTranscoderAPI) *Client {
	if client == nil {
		var elastictranscoderClient *elastictranscoderLib.ElasticTranscoder

		if useDevelopmentClient {
			log.Println("Creating development elastictranscoder client")
			elastictranscoderClient = elastictranscoderLib.New(session.New(), aws.NewConfig().WithEndpoint(config.Endpoint))
		} else {
			elastictranscoderClient = elastictranscoderLib.New(session.New())
		}

		return &Client{
			elastictranscoderClient,
			config,
		}
	}

	return &Client{
		client,
		config,
	}
}

// CreateNewJob creates a new elastictranscoder job.
func (c Client) CreateNewJob(pipelineID string, inputKey string, outputKey string, outputPresetID string, outputKeyPrefix string, thumbnailPattern string, metadata map[string]*string) (string, error) {
	params := &elastictranscoderLib.CreateJobInput{
		PipelineId: aws.String(pipelineID),
		Input: &elastictranscoderLib.JobInput{
			Key: aws.String(inputKey),
		},
		Output: &elastictranscoderLib.CreateJobOutput{
			Key:      aws.String(outputKey),
			PresetId: aws.String(outputPresetID),
		},
		OutputKeyPrefix: aws.String(outputKeyPrefix),
		UserMetadata:    metadata,
	}

	if thumbnailPattern != "" {
		params.Output.ThumbnailPattern = aws.String(thumbnailPattern)
	}

	response, err := c.CreateJob(params)
	if err != nil {
		return "", err
	}

	if response.Job == nil {
		return "", fmt.Errorf(
			"No Job returned from c.CreateJob client method: %v",
			response,
		)
	}

	return *response.Job.Id, nil
}
