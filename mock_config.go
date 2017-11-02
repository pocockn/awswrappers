package awswrappers

import "errors"

type (
	// MockConfig ...
	MockConfig struct {
		MockBoolean         func(node string, key string) (bool, error)
		MockEncryptedString func(node string, key string) (string, error)
		MockEnvironment     func() string
		MockInteger         func(node string, key string) (int, error)
		MockString          func(node string, key string) (string, error)
	}
)

// Boolean fetches a boolean value from the mock config.
func (mc MockConfig) Boolean(node string, key string) (bool, error) {
	if mc.MockBoolean != nil {
		return mc.MockBoolean(node, key)
	}

	return false, nil
}

// EncryptedString ...
func (mc MockConfig) EncryptedString(node string, key string) (string, error) {
	if mc.MockEncryptedString != nil {
		return mc.MockEncryptedString(node, key)
	}
	return "encrypted_key", nil
}

// Environment returns the current environment.
func (mc MockConfig) Environment() string {
	if mc.MockEnvironment != nil {
		return mc.MockEnvironment()
	}
	return "test"
}

// Integer ...
func (mc MockConfig) Integer(node string, key string) (int, error) {
	if mc.MockInteger != nil {
		return mc.MockInteger(node, key)
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

// NewErrorConfig ...
func NewErrorConfig(valueType string, configKey string) *MockConfig {
	mockConfig := MockConfig{}
	switch {
	case "boolean" == valueType:
		mockConfig.MockBoolean = func(node string, key string) (bool, error) {
			if configKey == key {
				return false, errors.New("Config error")
			}
			return false, nil
		}

	case "string" == valueType:
		mockConfig.MockString = func(node string, key string) (string, error) {
			if configKey == key {
				return "", errors.New("Config error")
			}
			return "", nil
		}

	case "integer" == valueType:
		mockConfig.MockInteger = func(node string, key string) (int, error) {
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
