package gateway

// APIError structure
type APIError struct {
	Code           string
	Description    string
	HTTPStatusCode int
}

type errCodeMap = map[int]APIError

// error code to APIError structure, these fields carry respective
// descriptions for all the error responses.
var errCodes = errCodeMap{}
