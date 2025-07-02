package errors

import (
	"errors"
	"testing"
)

func TestNewConfigError(t *testing.T) {
	err := NewConfigError("test message", "testField", "testValue")

	if err.Type.Code != "CONFIG" {
		t.Errorf("Expected error type 'CONFIG', got '%s'", err.Type.Code)
	}

	if err.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", err.Message)
	}

	if err.Field != "testField" {
		t.Errorf("Expected field 'testField', got '%s'", err.Field)
	}

	if err.Value != "testValue" {
		t.Errorf("Expected value 'testValue', got '%v'", err.Value)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("validation failed", "email", "invalid@")

	if err.Type.Code != "VALIDATION" {
		t.Errorf("Expected error type 'VALIDATION', got '%s'", err.Type.Code)
	}

	expected := "[VALIDATION] validation failed (field: email, value: invalid@)"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestNewNetworkError(t *testing.T) {
	originalErr := errors.New("connection refused")
	err := NewNetworkError("failed to connect", originalErr)

	if err.Type.Code != "NETWORK" {
		t.Errorf("Expected error type 'NETWORK', got '%s'", err.Type.Code)
	}

	if err.Cause != originalErr {
		t.Errorf("Expected cause to be set")
	}

	// Test Unwrap
	if err.Unwrap() != originalErr {
		t.Errorf("Unwrap should return the original error")
	}
}

func TestErrorIs(t *testing.T) {
	err1 := NewConfigError("test", "field", "value")
	err2 := NewValidationError("test", "field", "value")

	// Test Is method with same error type
	if !err1.Is(err1) {
		t.Errorf("Error should be identified as itself")
	}

	// Test Is method with different error type
	if err1.Is(err2) {
		t.Errorf("Config error should not be identified as validation error")
	}

	// Create another config error to test type matching
	err3 := NewConfigError("different message", "other", "other")
	if !err1.Is(err3) {
		t.Errorf("Config errors should be identified as same type")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := WrapError(originalErr, ErrXML, "XML processing failed")

	if wrappedErr.Type.Code != "XML" {
		t.Errorf("Expected error type 'XML', got '%s'", wrappedErr.Type.Code)
	}

	if wrappedErr.Message != "XML processing failed" {
		t.Errorf("Expected message 'XML processing failed', got '%s'", wrappedErr.Message)
	}

	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected cause to be the original error")
	}
}

func TestErrorWithoutField(t *testing.T) {
	err := NewSEFAZError("SEFAZ service unavailable", 503, nil)

	expected := "[SEFAZ] SEFAZ service unavailable"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestAllErrorTypes(t *testing.T) {
	errorTypes := []*ErrorType{
		ErrConfig,
		ErrValidation,
		ErrNetwork,
		ErrCertificate,
		ErrXML,
		ErrSEFAZ,
	}

	expectedCodes := []string{
		"CONFIG",
		"VALIDATION",
		"NETWORK",
		"CERTIFICATE",
		"XML",
		"SEFAZ",
	}

	for i, errType := range errorTypes {
		if errType.Code != expectedCodes[i] {
			t.Errorf("Expected error type code '%s', got '%s'", expectedCodes[i], errType.Code)
		}
	}
}
