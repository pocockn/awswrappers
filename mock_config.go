package awswrappers

import "errors"

type (
	// MockConfig ...
	MockConfig struct {
		MockInterger        func(node string, key string) (int, error)
		MockString          func(node string, key string) (string, error)
		MockEncryptedString func(node string, key string) (string, error)
	}
)

// Integer ...
func (mc MockConfig) Integer(node string, key string) (int, error) {
	if mc.MockInterger != nil {
		return mc.MockInterger(node, key)
	}

	return 1, nil
}

// String ...
func (mc MockConfig) String(node string, key string) (string, error) {
	if mc.MockString != nil {
		return mc.MockString(node, key)
	}
	return "string", nil
}

// EncryptedString ...
func (mc MockConfig) EncryptedString(node string, key string) (string, error) {
	if mc.MockEncryptedString != nil {
		return mc.MockEncryptedString(node, key)
	}
	return "encrypted_key", nil
}

// NewErrorConfig ...
func NewErrorConfig(valueType string, configKey string) *MockConfig {
	mockConfig := MockConfig{}
	switch {
	case "string" == valueType:
		mockConfig.MockString = func(node string, key string) (string, error) {
			if configKey == key {
				return "", errors.New("Config error")
			}
			return "", nil
		}

	case "integer" == valueType:
		mockConfig.MockInterger = func(node string, key string) (int, error) {
			if configKey == key {
				return 0, errors.New("Config error")
			}
			return 1, nil
		}

	case "encryptedString" == valueType:
		mockConfig.MockEncryptedString = func(node string, key string) (string, error) {
			if configKey == key {
				return "", errors.New("Config error")
			}
			return "", nil
		}
	}

	return &mockConfig
}
