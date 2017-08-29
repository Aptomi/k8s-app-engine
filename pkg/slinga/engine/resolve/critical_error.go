package resolve

import "github.com/Aptomi/aptomi/pkg/slinga/errors"

// CriticalError represents a critical error inside the engine
// If an error gets wrapped into a critical error, then it signals the engine to stop policy processing and report an error
// If an error doesn't get wrapped into this critical error, then the engine will just say that current dependency can't be fulfilled and move on to other dependencies
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
