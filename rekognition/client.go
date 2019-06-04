package rekognition

import (
	"github.com/aws/aws-sdk-go/aws/session"
	rekognitionLib "github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/aws/aws-sdk-go/service/rekognition/rekognitioniface"
)

type (
	// Client wraps the Rekognition API.
	Client struct {
		rekognitioniface.RekognitionAPI
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(client rekognitioniface.RekognitionAPI) *Client {
	if client == nil {
		client = rekognitionLib.New(session.New())
	}

	return &Client{
		client,
	}
}
