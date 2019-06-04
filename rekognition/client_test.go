package rekognition_test

import (
	"github.com/aws/aws-sdk-go/service/rekognition/rekognitioniface"
	"github.com/pkg/errors"
	"github.com/pocockn/awswrappers/rekognition"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	rekognitionLib "github.com/aws/aws-sdk-go/service/rekognition"
)

type (
	MockRekognitionClient struct {
		rekognitioniface.RekognitionAPI

		MockCompareFaces func(input *rekognitionLib.CompareFacesInput) (*rekognitionLib.CompareFacesOutput, error)
	}
)

func (m MockRekognitionClient) CompareFaces(input *rekognitionLib.CompareFacesInput) (*rekognitionLib.CompareFacesOutput, error) {
	if m.MockCompareFaces != nil {
		return m.MockCompareFaces(input)
	}

	return nil, nil
}

func NewTestClient(mockClient *MockRekognitionClient) *rekognition.Client {
	if mockClient == nil {
		mockClient = &MockRekognitionClient{}
	}

	return rekognition.NewClient(mockClient)
}

func TestClient(t *testing.T) {
	compareFacesInput := rekognitionLib.CompareFacesInput{
		SimilarityThreshold: aws.Float64(40.000000),
		SourceImage: &rekognitionLib.Image{
			Bytes: []byte("source"),
		},
		TargetImage: &rekognitionLib.Image{
			Bytes: []byte("target"),
		},
	}

	t.Run(".CompareFaces()", func(t *testing.T) {
		face := rekognitionLib.ComparedFace{
			Confidence: aws.Float64(60.0000),
		}
		faceMatches := []*rekognitionLib.CompareFacesMatch{
			{
				Similarity: aws.Float64(60.000000),
				Face:       &face,
			},
		}

		t.Run("ReturnsCompareFacesOutput", func(t *testing.T) {
			mockClient := &MockRekognitionClient{
				MockCompareFaces: func(input *rekognitionLib.CompareFacesInput) (*rekognitionLib.CompareFacesOutput, error) {
					return &rekognitionLib.CompareFacesOutput{
						FaceMatches: faceMatches,
					}, nil
				},
			}

			client := NewTestClient(mockClient)

			compareFaceOuput, err := client.CompareFaces(&compareFacesInput)
			assert.NoError(t, err)

			assert.Equal(t, compareFaceOuput.FaceMatches[0].Similarity, aws.Float64(60.000000))
		})

		t.Run("ReturnsOnError", func(t *testing.T) {
			mockClient := &MockRekognitionClient{
				MockCompareFaces: func(input *rekognitionLib.CompareFacesInput) (*rekognitionLib.CompareFacesOutput, error) {
					return nil, errors.New("Compare faces error.")
				},
			}

			testClient := NewTestClient(mockClient)

			_, err := testClient.CompareFaces(&compareFacesInput)
			assert.Error(t, err)
		})
	})
}
