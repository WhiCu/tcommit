package template

import "fmt"

// Common template errors
var (
	ErrInvalidTokenSyntax = fmt.Errorf("invalid token syntax")
	ErrNoReplacement      = fmt.Errorf("no replacement for key")
	ErrInvalidValue       = fmt.Errorf("invalid value for key")
)

// Error constructors
func NewNoReplacementError(key string) error {
	return fmt.Errorf("%w %q", ErrNoReplacement, key)
}

func NewInvalidValueError(val, key string, choices []string) error {
	return fmt.Errorf("%w %q - %q; allowed: %v", ErrInvalidValue, key, val, choices)
}

func NewInvalidTokenSyntaxError(token string) error {
	return fmt.Errorf("%w: %q", ErrInvalidTokenSyntax, token)
}
