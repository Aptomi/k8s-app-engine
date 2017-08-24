package errors

type Details map[string]interface{}

type ErrorWithDetails struct {
	message string
	details Details
}

func NewErrorWithDetails(message string, details Details) *ErrorWithDetails {
	return &ErrorWithDetails{message: message, details: details}
}

func (err *ErrorWithDetails) Error() string {
	return err.message
}

func (err *ErrorWithDetails) Details() Details {
	return err.details
}
