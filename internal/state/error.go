package state

type Error struct {
	error
	msg        string
	statusCode int
}

func (e *Error) ServerErrorResponse() (int, string) {
	return e.statusCode, e.msg
}
