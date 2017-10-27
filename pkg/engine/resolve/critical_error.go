package resolve

import "github.com/Aptomi/aptomi/pkg/errors"

// CriticalError represents a critical error inside the engine.
// If a critical error occurs, it signals the engine to stop policy processing and fail policy resolution with an error.
// If a non-critical error occurs, then the engine will skip the current declaration of service consumption (the one
// that is currently being processed) and proceed with policy resolution, moving on to the remaining declarations of
// service consumption
type CriticalError struct {
	*errors.ErrorWithDetails
	logged bool
}

// NewCriticalError creates a new critical error
func NewCriticalError(err *errors.ErrorWithDetails) *CriticalError {
	return &CriticalError{ErrorWithDetails: err, logged: false}
}

// SetLoggedFlag sets a flag that an error has been logged. So that when we go up the recursion
// stack in the engine, it doesn't get logged multiple times
func (err *CriticalError) SetLoggedFlag() {
	err.logged = true
}

// IsLogged returns if an error has been already processed and logged
func (err *CriticalError) IsLogged() bool {
	return err.logged
}
