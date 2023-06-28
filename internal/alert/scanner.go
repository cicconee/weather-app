package alert

type Scanner interface {
	Scan(...any) error
}
