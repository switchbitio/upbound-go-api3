package errors

// notFoundError is an error indicating the resource is not found.
type notFoundError struct {
	err error
}

// Error calls the underlying error's Error method.
func (n *notFoundError) Error() string {
	return n.err.Error()
}

// NotFound indicates that this is a not found error.
func (n *notFoundError) NotFound() bool {
	return true
}

// NewNotFound wraps an existing error as a not found error.
func NewNotFound(err error) error {
	return &notFoundError{
		err: err,
	}
}

// notFound indicates a resource is not found.
type notFound interface {
	NotFound() bool
}

// IsNotFound checks whether an error implements the not found interface.
func IsNotFound(err error) bool {
	ne, ok := err.(notFound) //nolint:errorlint // we want to be able to call and return NotFound()
	return ok && ne.NotFound()
}
