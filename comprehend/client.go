package comprehend

import (
	"github.com/aws/aws-sdk-go/aws/session"
	comprehendLib "github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/aws/aws-sdk-go/service/comprehend/comprehendiface"
)

type (
	// Client wraps the Comprehend API.
	Client struct {
		comprehendiface.ComprehendAPI
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(client comprehendiface.ComprehendAPI) *Client {
	if client == nil {
		client = comprehendLib.New(session.New())
	}

	return &Client{
		client,
	}
}
