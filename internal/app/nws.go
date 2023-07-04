package app

import "fmt"

// NWSAPIStatusCodeError is an error that occurs when the NWS API returns
// a unexpected status code for a request.
//
// The body of a unexpected status code response from the NWS API will
// always be in JSON format and contain a status and detail field. These
// values can be unmarshalled into a NWSAPIStatusCodeError.
type NWSAPIStatusCodeError struct {
	StatusCode int    `json:"status"`
	Detail     string `json:"detail"`
}

func (s *NWSAPIStatusCodeError) Error() string {
	return fmt.Sprintf("statusCode=%d, detail=%s", s.StatusCode, s.Detail)
}
