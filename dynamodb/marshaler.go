package dynamodb

import (
	dynamodbLib "github.com/aws/aws-sdk-go/service/dynamodb"
)

type (
	// Marshaler represents an object that returns a
	// *dynamodb.PutItemInput object for the struct
	// implementing the interface.
	Marshaler interface {
		Marshal() (*dynamodbLib.PutItemInput, error)
	}
)
