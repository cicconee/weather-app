package app

// ServerResponseError is returned by service methods that a HTTP
// server uses. These errors are intended to hold the appropriate
// HTTP response body and status code to be returned by the server.
// This data is considered safe and can be seen by external sources.
//
// Use the ServerErrorResponse method to get the data that is safe
// to be displayed to external sources.
type ServerResponseError struct {
	// The wrapped error.
	Err error

	// The HTTP response body.
	Msg string

	// The HTTP status code.
	StatusCode int
}

// NewServerResponseError returns a pointer to a ServerResponseError
// set with the data provided.
func NewServerResponseError(err error, msg string, statusCode int) *ServerResponseError {
	return &ServerResponseError{
		Err:        err,
		Msg:        msg,
		StatusCode: statusCode,
	}
}

// ServerErrorResponse returns the status code and the response body.
func (e *ServerResponseError) ServerErrorResponse() (int, string) {
	return e.StatusCode, e.Msg
}

func (e *ServerResponseError) Unwrap() error {
	return e.Err
}

func (e *ServerResponseError) Error() string {
	return e.Err.Error()
}
