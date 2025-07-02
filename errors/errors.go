// Package errors provides custom error types for sped-nfe-go library.
// This package replaces PHP exceptions with Go idiomatic error handling.
package errors

import (
	"fmt"
)

// Error types for different categories of failures
var (
	// ErrConfig represents configuration-related errors
	ErrConfig = &ErrorType{Code: "CONFIG", Message: "Configuration error"}
	
	// ErrValidation represents validation-related errors
	ErrValidation = &ErrorType{Code: "VALIDATION", Message: "Validation error"}
	
	// ErrNetwork represents network/communication errors
	ErrNetwork = &ErrorType{Code: "NETWORK", Message: "Network error"}
	
	// ErrCertificate represents certificate-related errors
	ErrCertificate = &ErrorType{Code: "CERTIFICATE", Message: "Certificate error"}
	
	// ErrXML represents XML processing errors
	ErrXML = &ErrorType{Code: "XML", Message: "XML processing error"}
	
	// ErrSEFAZ represents SEFAZ webservice errors
	ErrSEFAZ = &ErrorType{Code: "SEFAZ", Message: "SEFAZ error"}
)

// ErrorType represents a category of error
type ErrorType struct {
	Code    string
	Message string
}

// NFError represents a structured error with context
type NFError struct {
	Type    *ErrorType
	Message string
	Field   string
	Value   interface{}
	Cause   error
}

// Error implements the error interface
func (e *NFError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("[%s] %s (field: %s, value: %v)", e.Type.Code, e.Message, e.Field, e.Value)
	}
	return fmt.Sprintf("[%s] %s", e.Type.Code, e.Message)
}

// Unwrap returns the underlying cause error
func (e *NFError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches a specific error type
func (e *NFError) Is(target error) bool {
	if t, ok := target.(*NFError); ok {
		return e.Type.Code == t.Type.Code
	}
	return false
}

// NewConfigError creates a new configuration error
func NewConfigError(message string, field string, value interface{}) *NFError {
	return &NFError{
		Type:    ErrConfig,
		Message: message,
		Field:   field,
		Value:   value,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string, field string, value interface{}) *NFError {
	return &NFError{
		Type:    ErrValidation,
		Message: message,
		Field:   field,
		Value:   value,
	}
}

// NewNetworkError creates a new network error
func NewNetworkError(message string, cause error) *NFError {
	return &NFError{
		Type:    ErrNetwork,
		Message: message,
		Cause:   cause,
	}
}

// NewCertificateError creates a new certificate error
func NewCertificateError(message string, cause error) *NFError {
	return &NFError{
		Type:    ErrCertificate,
		Message: message,
		Cause:   cause,
	}
}

// NewXMLError creates a new XML processing error
func NewXMLError(message string, field string, cause error) *NFError {
	return &NFError{
		Type:    ErrXML,
		Message: message,
		Field:   field,
		Cause:   cause,
	}
}

// NewSEFAZError creates a new SEFAZ error
func NewSEFAZError(message string, value interface{}, cause error) *NFError {
	return &NFError{
		Type:    ErrSEFAZ,
		Message: message,
		Value:   value,
		Cause:   cause,
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType *ErrorType, message string) *NFError {
	return &NFError{
		Type:    errorType,
		Message: message,
		Cause:   err,
	}
}