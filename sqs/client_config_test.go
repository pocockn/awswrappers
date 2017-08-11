package sqs_test

import (
	"testing"

	"github.com/vidsy/awswrappers"
	"github.com/vidsy/awswrappers/sqs"
)

func TestClientConfig(t *testing.T) {
	t.Run("NewClientConfigFromKMSConfig", func(t *testing.T) {
		t.Run("CreatesWithValidConfig", func(t *testing.T) {
			clientConfig, _ := sqs.NewClientConfigFromKMSConfig(&awswrappers.MockConfig{})

			if clientConfig == nil {
				t.Fatalf("Expected new ClientConfig, got: %v", clientConfig)
			}
		})

		t.Run("ReturnsErrorWhenInvalidConfig", func(t *testing.T) {
			var errorCases = []struct {
				valueType string
				key       string
			}{
				{"string", "endpoint_url"},
				{"integer", "max_number_of_messages"},
				{"integer", "visibility_timeout"},
				{"integer", "wait_time_seconds"},
			}

			for _, errorCase := range errorCases {
				mockConfig := awswrappers.NewErrorConfig(errorCase.valueType, errorCase.key)
				_, err := sqs.NewClientConfigFromKMSConfig(mockConfig)

				if err == nil {
					t.Errorf(
						"Expected error when config value '%s' is invalid, got: %s",
						errorCase.key,
						err,
					)
				}
			}
		})
	})
}
