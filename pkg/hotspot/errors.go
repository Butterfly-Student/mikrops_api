package hotspot

import "fmt"

// Error definitions for hotspot operations
var (
	// Profile errors
	ErrProfileNotFound = fmt.Errorf("profile not found")
	ErrInvalidProfile  = fmt.Errorf("invalid profile configuration")

	// User errors
	ErrUserNotFound    = fmt.Errorf("user not found")
	ErrInvalidUser     = fmt.Errorf("invalid user data")
	ErrUserExists      = fmt.Errorf("user already exists")

	// Session errors
	ErrSessionNotFound = fmt.Errorf("session not found")
	ErrUserNotActive   = fmt.Errorf("user is not currently active")

	// Sales errors
	ErrSaleNotFound    = fmt.Errorf("sale record not found")
	ErrInvalidSale     = fmt.Errorf("invalid sale data")

	// Voucher errors
	ErrInvalidValidity = fmt.Errorf("invalid validity format")
	ErrInvalidQuantity = fmt.Errorf("quantity must be greater than 0")
	ErrInvalidLength   = fmt.Errorf("length must be between 4 and 12")

	// Scheduler errors
	ErrSchedulerNotFound = fmt.Errorf("scheduler not found")
	ErrExpiryMode       = fmt.Errorf("invalid expiry mode")

	// Connection errors
	ErrNotConnected     = fmt.Errorf("not connected to router")
	ErrCommandFailed    = fmt.Errorf("command execution failed")
)

// HotspotError wraps errors with additional context about the operation
type HotspotError struct {
	// Operation is the name of the operation that failed
	Operation string

	// Err is the underlying error
	Err error

	// Context contains additional context (optional)
	Context map[string]interface{}
}

// Error returns the formatted error message
func (e *HotspotError) Error() string {
	base := fmt.Sprintf("hotspot %s: %v", e.Operation, e.Err)
	if len(e.Context) > 0 {
		base += fmt.Sprintf(" (context: %v)", e.Context)
	}
	return base
}

// Unwrap returns the underlying error for use with errors.Is/As
func (e *HotspotError) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *HotspotError) WithContext(key string, value interface{}) *HotspotError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewError creates a new HotspotError with the given operation and underlying error
func NewError(operation string, err error) *HotspotError {
	return &HotspotError{
		Operation: operation,
		Err:       err,
	}
}

// WrapError wraps an error with additional operation context
// If err is already a HotspotError, it updates the operation
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if hErr, ok := err.(*HotspotError); ok {
		hErr.Operation = operation
		return hErr
	}
	return NewError(operation, err)
}

// IsNotFound checks if an error indicates a resource was not found
func IsNotFound(err error) bool {
	return err == ErrProfileNotFound ||
		err == ErrUserNotFound ||
		err == ErrSessionNotFound ||
		err == ErrSaleNotFound ||
		err == ErrSchedulerNotFound
}

// IsInvalid checks if an error indicates invalid input
func IsInvalid(err error) bool {
	return err == ErrInvalidProfile ||
		err == ErrInvalidUser ||
		err == ErrInvalidSale ||
		err == ErrInvalidValidity ||
		err == ErrInvalidQuantity ||
		err == ErrInvalidLength
}
