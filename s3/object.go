package s3

import (
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	s3Lib "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

type (
	// Object represents an object in s3.
	Object struct {
		Bucket string
		Key    string
		client s3iface.S3API
	}
)

// NewObject creates a new Object struct and returns it.
func NewObject(bucket string, key string, client s3iface.S3API) Object {
	return Object{
		Bucket: bucket,
		Key:    key,
		client: client,
	}
}

// Get returns the data for a given key.
func (s Object) Get() (io.ReadCloser, error) {
	params := &s3Lib.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}

	resp, err := s.client.GetObject(params)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// PresignedPutURI returns a pre signed URI with the
// given expiration.
func (s Object) PresignedPutURI(expiration time.Duration) (string, error) {
	params := &s3Lib.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}

	request, _ := s.client.PutObjectRequest(params)
	url, err := request.Presign(expiration)
	if err != nil {
		return "", err
	}

	return url, nil
}

// Put puts the given data to the given key in S3.
func (s Object) Put(body io.ReadSeeker, contentType string) error {
	params := &s3Lib.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(s.Key),
		ContentType: aws.String(contentType),
		Body:        body,
	}

	_, err := s.client.PutObject(params)
	if err != nil {
		return err
	}

	return nil
}

// RangeGet returns the data for a given byte range.
func (s Object) RangeGet(rangeHeader string) (io.ReadCloser, error) {
	params := &s3Lib.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
		Range:  aws.String(rangeHeader),
	}

	resp, err := s.client.GetObject(params)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// Size returns the size of an S3 object.
func (s Object) Size() (int64, error) {
	params := &s3Lib.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}

	resp, err := s.client.HeadObject(params)
	if err != nil {
		return 0, err
	}

	return *resp.ContentLength, nil
}
