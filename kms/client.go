package kms

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	kmsLib "github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

type (
	// Client wraps the AWS KMS client.
	Client struct {
		kmsiface.KMSAPI
		developmentMode bool
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(developmentMode bool, client kmsiface.KMSAPI) *Client {
	if client == nil {
		client = kmsLib.New(session.New())
	}

	return &Client{
		client,
		developmentMode,
	}
}

// EncryptData takes a KMS key arn and data to encrypt and
// returns the encrypted Ciphertext Blob.
func (c Client) EncryptData(keyID string, data []byte) (string, error) {
	if c.developmentMode {
		return string(data[:]), nil
	}

	input := &kmsLib.EncryptInput{
		KeyId:     aws.String(keyID),
		Plaintext: data,
	}

	result, err := c.Encrypt(input)
	if err != nil {
		return "", err
	}

	return string(result.CiphertextBlob[:]), nil
}

// DecryptData takes a blob of encrypted data and attempts to
// decrypt it.
func (c Client) DecryptData(data []byte) (string, error) {
	if c.developmentMode {
		return string(data[:]), nil
	}

	input := &kmsLib.DecryptInput{
		CiphertextBlob: data,
	}

	result, err := c.Decrypt(input)
	if err != nil {
		return "", err
	}

	return string(result.Plaintext[:]), nil
}
