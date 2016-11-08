package sqs

import (
	"errors"
	"testing"
)

type (
	MockConfig struct {
		mockInterger        func(node string, key string) (int, error)
		mockString          func(node string, key string) (string, error)
		mockEncryptedString func(node string, key string) (string, error)
	}
)

func (mc MockConfig) Integer(node string, key string) (int, error) {
	if mc.mockInterger != nil {
		return mc.mockInterger(node, key)
	}

	return 1, nil
}

func (mc MockConfig) String(node string, key string) (string, error) {
	if mc.mockString != nil {
		return mc.mockString(node, key)
	}
	return "string", nil
}

func (mc MockConfig) EncryptedString(node string, key string) (string, error) {
	if mc.mockEncryptedString != nil {
		return mc.mockEncryptedString(node, key)
	}
	return "encrypted_key", nil
}

func NewErrorConfig(valueType string, configKey string) *MockConfig {
	mockConfig := MockConfig{}
	switch {
	case "string" == valueType:
		mockConfig.mockString = func(node string, key string) (string, error) {
			if configKey == key {
				return "", errors.New("Config error")
			}
			return "", nil
		}

	case "integer" == valueType:
		mockConfig.mockInterger = func(node string, key string) (int, error) {
			if configKey == key {
				return 0, errors.New("Config error")
			}
			return 1, nil
		}

	case "encryptedString" == valueType:
		mockConfig.mockEncryptedString = func(node string, key string) (string, error) {
			if configKey == key {
				return "", errors.New("Config error")
			}
			return "", nil
		}
	}

	return &mockConfig
}

func TestClientConfig(t *testing.T) {
	t.Run("NewClientConfigFromKMSConfig", func(t *testing.T) {
		t.Run("CreatesWithValidConfig", func(t *testing.T) {
			clientConfig, _ := NewClientConfigFromKMSConfig(&MockConfig{})

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
				{"string", "queue_name"},
				{"integer", "max_number_of_messages"},
				{"integer", "visibility_timeout"},
				{"integer", "wait_time_seconds"},
			}

			for _, errorCase := range errorCases {
				mockConfig := NewErrorConfig(errorCase.valueType, errorCase.key)
				_, err := NewClientConfigFromKMSConfig(mockConfig)

				if err == nil {
					t.Errorf("Expected error when config value '%s' is invalid, got:", errorCase.key, err)
				}
			}
		})
	})
}
