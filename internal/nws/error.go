package nws

import "fmt"

type StatusCodeError struct {
	StatusCode int    `json:"status"`
	Detail     string `json:"detail"`
}

func (s *StatusCodeError) Error() string {
	return fmt.Sprintf("invalid status code (StatusCode: %d, Detail: %s)", s.StatusCode, s.Detail)
}
