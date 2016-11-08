package sqs

import "github.com/vidsy/kmsconfig"

type (
	// ClientConfig store config values for the Client.
	ClientConfig struct {
		QueueEndpoint       string
		QueueName           string
		MaxNumberOfMessages int64
		VisibilityTimeout   int64
		WaitTimeSeconds     int64
	}
)

// NewClientConfigFromKMSConfig Creates new client config based on config values
// for the current environment.
func NewClientConfigFromKMSConfig(config kmsconfig.ConfigInterrogator) (*ClientConfig, error) {
	endpointURL, err := config.String("sqs", "endpoint_url")
	if err != nil {
		return nil, err
	}

	queueName, err := config.String("sqs", "queue_name")
	if err != nil {
		return nil, err
	}

	maxNumberOfmessages, err := config.Integer("sqs", "max_number_of_messages")
	if err != nil {
		return nil, err
	}

	visibilityTimeout, err := config.Integer("sqs", "visibility_timeout")
	if err != nil {
		return nil, err
	}

	waitTimeSeconds, err := config.Integer("sqs", "wait_time_seconds")
	if err != nil {
		return nil, err
	}

	return &ClientConfig{
		endpointURL,
		queueName,
		int64(maxNumberOfmessages),
		int64(visibilityTimeout),
		int64(waitTimeSeconds),
	}, nil
}
