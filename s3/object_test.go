package s3_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/request"
	s3Lib "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
	"github.com/vidsy/awswrappers/s3"
)

type (
	MockS3Client struct {
		s3iface.S3API
		mockGetObject        func(*s3Lib.GetObjectInput) (*s3Lib.GetObjectOutput, error)
		mockHeadObject       func(*s3Lib.HeadObjectInput) (*s3Lib.HeadObjectOutput, error)
		mockPutObject        func(*s3Lib.PutObjectInput) (*s3Lib.PutObjectOutput, error)
		mockPutObjectRequest func(input *s3Lib.PutObjectInput) (*request.Request, *s3Lib.PutObjectOutput)
	}
)

func (m MockS3Client) GetObject(input *s3Lib.GetObjectInput) (*s3Lib.GetObjectOutput, error) {
	if m.mockGetObject != nil {
		return m.mockGetObject(input)
	}

	body := ioutil.NopCloser(bytes.NewReader([]byte("foo")))
	return &s3Lib.GetObjectOutput{Body: body}, nil
}

func (m MockS3Client) HeadObject(input *s3Lib.HeadObjectInput) (*s3Lib.HeadObjectOutput, error) {
	if m.mockHeadObject != nil {
		return m.mockHeadObject(input)
	}

	var fileSize int64 = 1024

	return &s3Lib.HeadObjectOutput{ContentLength: &fileSize}, nil
}

func (m MockS3Client) PutObjectRequest(input *s3Lib.PutObjectInput) (*request.Request, *s3Lib.PutObjectOutput) {
	if m.mockPutObjectRequest != nil {
		return m.mockPutObjectRequest(input)
	}

	return nil, nil
}

func (m MockS3Client) PutObject(input *s3Lib.PutObjectInput) (*s3Lib.PutObjectOutput, error) {
	if m.mockPutObject != nil {
		return m.mockPutObject(input)
	}

	return &s3Lib.PutObjectOutput{}, nil
}

func TestObject(t *testing.T) {
	t.Run(".PresignedPutURI()", func(t *testing.T) {
		t.Run("ReturnsPresignedURI", func(t *testing.T) {
			testUrl, _ := url.Parse("http://www.foo.com")
			mockClient := &MockS3Client{
				mockPutObjectRequest: func(input *s3Lib.PutObjectInput) (*request.Request, *s3Lib.PutObjectOutput) {
					return &request.Request{
						Operation: &request.Operation{},
						HTTPRequest: &http.Request{
							URL: testUrl,
						},
					}, nil
				},
			}
			object := s3.NewObject(
				"foo",
				"bar",
				mockClient,
			)

			uri, err := object.PresignedPutURI(10 * time.Second)
			if err != nil {
				t.Fatalf("Expected .PresignedURI() to not return an error, got: '%s'", err)
			}

			if uri != testUrl.String() {
				t.Fatalf("Expected .PresignedURI() to return '%s', got: '%s'", testUrl.String(), uri)
			}
		})
	})

	t.Run("Put()", func(t *testing.T) {
		body := bytes.NewReader([]byte("test"))

		t.Run("ReturnsNoErrorOnSuccessfulPut", func(t *testing.T) {
			mockClient := &MockS3Client{}
			object := s3.NewObject(
				"foo",
				"bar",
				mockClient,
			)
			err := object.Put(body, "foo/bar")

			assert.NoError(t, err)
		})

		t.Run("ReturnsErrorOnClientError", func(t *testing.T) {
			mockClient := &MockS3Client{
				mockPutObject: func(*s3Lib.PutObjectInput) (*s3Lib.PutObjectOutput, error) {
					return nil, errors.New("Put object error")
				},
			}
			object := s3.NewObject(
				"foo",
				"bar",
				mockClient,
			)
			err := object.Put(body, "foo/bar")

			assert.Error(t, err)
		})
	})

	t.Run("Size()", func(t *testing.T) {
		t.Run("ReturnsSizeValue", func(t *testing.T) {
			mockClient := &MockS3Client{}
			object := s3.NewObject(
				"foo",
				"bar",
				mockClient,
			)

			var expectedSize int64 = 1024

			fileSize, err := object.Size()

			assert.Nil(t, err)
			assert.Equal(t, expectedSize, fileSize)
		})

		t.Run("ReturnsErrorOnClientError", func(t *testing.T) {
			mockClient := &MockS3Client{
				mockHeadObject: func(*s3Lib.HeadObjectInput) (*s3Lib.HeadObjectOutput, error) {
					return nil, errors.New("Head object error")
				},
			}
			object := s3.NewObject(
				"foo",
				"bar",
				mockClient,
			)

			_, err := object.Size()

			assert.NotNil(t, err)
		})
	})

	t.Run("RangeGet()", func(t *testing.T) {
		t.Run("ReturnsBody", func(t *testing.T) {
			mockClient := &MockS3Client{}
			object := s3.NewObject(
				"foo",
				"bar",
				mockClient,
			)

			bodyRaw, err := object.RangeGet("range=0-100")

			assert.Nil(t, err)

			bodyByteData, _ := ioutil.ReadAll(bodyRaw)
			body := string(bodyByteData[:])

			assert.Equal(t, "foo", body)
		})

		t.Run("ReturnsErrorOnClientError", func(t *testing.T) {
			mockClient := &MockS3Client{
				mockGetObject: func(*s3Lib.GetObjectInput) (*s3Lib.GetObjectOutput, error) {
					return nil, errors.New("Get object error")
				},
			}

			object := s3.NewObject(
				"foo",
				"bar",
				mockClient,
			)

			_, err := object.RangeGet("range=0-100")

			assert.NotNil(t, err)
		})
	})
}
