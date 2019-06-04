package comprehend_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go/aws"
	comprehendLib "github.com/aws/aws-sdk-go/service/comprehend"
	"github.com/aws/aws-sdk-go/service/comprehend/comprehendiface"
	"github.com/vidsy/awswrappers/comprehend"
)

type (
	MockComprehendClient struct {
		comprehendiface.ComprehendAPI

		mockDetectKeyPhrases func(input *comprehendLib.DetectKeyPhrasesInput) (*comprehendLib.DetectKeyPhrasesOutput, error)
	}
)

func (m MockComprehendClient) DetectKeyPhrases(input *comprehendLib.DetectKeyPhrasesInput) (*comprehendLib.DetectKeyPhrasesOutput, error) {
	if m.mockDetectKeyPhrases != nil {
		return m.mockDetectKeyPhrases(input)
	}

	return nil, nil
}

func NewTestClient(mockClient *MockComprehendClient) *comprehend.Client {
	if mockClient == nil {
		mockClient = &MockComprehendClient{}
	}

	return comprehend.NewClient(mockClient)
}

func TestClient(t *testing.T) {
	keyPhraseInput := comprehendLib.DetectKeyPhrasesInput{
		LanguageCode: aws.String("en"),
		Text:         aws.String("Hello Vidsy"),
	}

	t.Run(".DetectKeyPhrases()", func(t *testing.T) {
		t.Run("ReturnsKeyPhraseOutput", func(t *testing.T) {
			keyPhrase := comprehendLib.KeyPhrase{
				BeginOffset: aws.Int64(0),
				EndOffset:   aws.Int64(1),
				Score:       aws.Float64(1),
				Text:        aws.String("Hello Vidsy"),
			}
			keyPhrases := []*comprehendLib.KeyPhrase{&keyPhrase}

			mockClient := &MockComprehendClient{
				mockDetectKeyPhrases: func(input *comprehendLib.DetectKeyPhrasesInput) (*comprehendLib.DetectKeyPhrasesOutput, error) {
					return &comprehendLib.DetectKeyPhrasesOutput{
						KeyPhrases: keyPhrases,
					}, nil
				},
			}

			client := NewTestClient(mockClient)

			keyPhraseOutput, err := client.DetectKeyPhrases(&keyPhraseInput)
			assert.NoError(t, err)

			if assert.Len(t, keyPhraseOutput.KeyPhrases, 1) {
				assert.Equal(t, keyPhraseOutput.KeyPhrases[0].Text, keyPhrase.Text)
			}
		})

		t.Run("ReturnsOnError", func(t *testing.T) {
			mockClient := &MockComprehendClient{
				mockDetectKeyPhrases: func(input *comprehendLib.DetectKeyPhrasesInput) (*comprehendLib.DetectKeyPhrasesOutput, error) {
					return nil, errors.New("Key Phrase Analysis Error")
				},
			}

			testClient := NewTestClient(mockClient)

			_, err := testClient.DetectKeyPhrases(&keyPhraseInput)
			assert.Error(t, err)
		})
	})
}