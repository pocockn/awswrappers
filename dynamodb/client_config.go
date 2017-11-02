package dynamodb

import "github.com/vidsy/go-kmsconfig/kmsconfig"

type (
	// ClientConfig store config values for the DynamoDB Client.
	ClientConfig struct {
		DynamoDBEndpoint string
	}
)

// NewClientConfigFromKMSConfig creates a new client config based on config
// values for the current environment.
func NewClientConfigFromKMSConfig(config kmsconfig.ConfigInterrogator) (*ClientConfig, error) {
	dynamoDBEndpoint, err := config.String("dynamodb", "dynamo_db_endpoint")
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		dynamoDBEndpoint,
	}, nil
}
