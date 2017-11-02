package elastictranscoder

import "github.com/vidsy/go-kmsconfig/kmsconfig"

type (
	// ClientConfig store config values for the Client.
	ClientConfig struct {
		Endpoint string
	}
)

// NewClientConfigFromKMSConfig Creates new client config based on config values
// for the current environment.
func NewClientConfigFromKMSConfig(config kmsconfig.ConfigInterrogator) (*ClientConfig, error) {
	endpointURL, err := config.String("elastictranscoder", "endpoint_url")
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		endpointURL,
	}, nil
}
