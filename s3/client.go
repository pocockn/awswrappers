package s3

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	s3Lib "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"log"
)

type (
	// Client wraps the receive and delete functionality of s3.
	Client struct {
		s3iface.S3API
		clientConfig *ClientConfig
	}
)

// NewClient creates a new wrapper based on the environment.
func NewClient(config *ClientConfig, useDevelopmentClient bool, client s3iface.S3API) *Client {
	if client == nil {
		var s3Client *s3Lib.S3

		if useDevelopmentClient {
			log.Println("Creating development S3 client")
			config := &aws.Config{
				Endpoint:         aws.String(config.Endpoint),
				S3ForcePathStyle: aws.Bool(true),
			}
			s3Client = s3Lib.New(session.New(), config)
		} else {
			s3Client = s3Lib.New(session.New())
		}

		return &Client{
			s3Client,
			config,
		}
	}

	return &Client{
		client,
		config,
	}
}

// PresignedURI takes a bucket, key and expiration and returns a presigned URI.
func (c Client) PresignedURI(bucket string, key string, expiration time.Duration) (string, error) {
	request, _ := c.PutObjectRequest(
		&s3Lib.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
	)

	url, err := request.Presign(expiration)
	if err != nil {
		return "", err
	}

	return url, nil
}
