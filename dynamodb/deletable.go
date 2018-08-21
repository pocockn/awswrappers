package dynamodb

type (
	// Deletable represents an object that can be deleted within Dynamo.
	Deletable interface {
		Key() map[string]interface{}
		TableName() string
	}
)
