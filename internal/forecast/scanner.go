package forecast

// Scanner is the interface that wraps the Scan method.
//
// Scan scans a database query result and stores it into the fields
// provided. It will return any errors encountered when scanning
// the values into the provided fields.
type Scanner interface {
	Scan(...any) error
}
