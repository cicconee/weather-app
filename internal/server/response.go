package server

type Response struct {
	Status int
	Body   any
}

type ErrorResponse struct {
	Status   int    `json:"-"`
	ErrorMsg string `json:"error_msg"`
}

func (e *ErrorResponse) AsResponse() Response {
	return Response{
		Status: e.Status,
		Body:   e,
	}
}
