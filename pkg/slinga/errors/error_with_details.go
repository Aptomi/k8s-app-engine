package errors

// Details is a type which defines a map of objects that will be attached to the error
type Details map[string]interface{}

// ErrorWithDetails is a custom error which also stores a context (map of objects) in addition to the error message itself
type ErrorWithDetails struct {
	message string
	details Details
}

// NewErrorWithDetails creates a new ErrorWithDetails
func NewErrorWithDetails(message string, details Details) *ErrorWithDetails {
	return &ErrorWithDetails{message: message, details: details}
}

// Error returns an error message
func (err *ErrorWithDetails) Error() string {
	return err.message
}

// Details returns error context, i.e. map of objects which are attached to the error
func (err *ErrorWithDetails) Details() Details {
	return err.details
}
