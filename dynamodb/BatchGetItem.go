package dynamodb

type (
	// BatchGetItem represents an item to be used within a BatchGetItem dynamo
	// request, it contains the hash key and the value of the thing we are
	// attempting to retrieve.
	BatchGetItem map[string][]interface{}
)
