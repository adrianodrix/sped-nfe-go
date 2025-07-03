package validation

import (
	"strings"
	"testing"
)

func TestNewXSDValidator(t *testing.T) {
	validator := NewXSDValidator()
	if validator == nil {
		t.Fatal("Expected validator to be created")
	}

	if validator.schemasPath != "schemas/xsd" {
		t.Errorf("Expected default schemas path to be 'schemas/xsd', got '%s'", validator.schemasPath)
	}
}

func TestNewXSDValidatorWithPath(t *testing.T) {
	customPath := "/custom/path"
	validator := NewXSDValidatorWithPath(customPath)

	if validator.schemasPath != customPath {
		t.Errorf("Expected custom schemas path to be '%s', got '%s'", customPath, validator.schemasPath)
	}
}

func TestValidateAccessKey(t *testing.T) {
	validator := NewXSDValidator()

	tests := []struct {
		name        string
		key         string
		expectValid bool
		expectError string
	}{
		{
			name:        "Valid access key",
			key:         "35250712345678000195550010000000011123456782",
			expectValid: true,
		},
		{
			name:        "Valid access key with formatting",
			key:         "3525 0712 3456 7800 0195 5500 1000 0000 0111 2345 6782",
			expectValid: true,
		},
		{
			name:        "Invalid length - too short",
			key:         "3520011420016600018755001000000001512345678",
			expectValid: false,
			expectError: "Access key must have exactly 44 digits, got 43",
		},
		{
			name:        "Invalid length - too long",
			key:         "352001142001660001875500100000000151234567890",
			expectValid: false,
			expectError: "Access key must have exactly 44 digits, got 45",
		},
		{
			name:        "Invalid character",
			key:         "3520011420016600018755001000000001512345678A",
			expectValid: false,
			expectError: "Invalid character 'A' at position 44",
		},
		{
			name:        "Invalid check digit",
			key:         "35200114200166000187550010000000015123456780", // Wrong check digit
			expectValid: false,
			expectError: "Invalid check digit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateAccessKey(tt.key)

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, result.Valid)
			}

			if tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if strings.Contains(err, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.expectError, result.Errors)
				}
			}
		})
	}
}

func TestCalculateModulo11(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected int
	}{
		{
			name:     "Valid NFe key without DV",
			number:   "3525071234567800019555001000000001112345678",
			expected: 2,
		},
		{
			name:     "Test case with repeated 1s",
			number:   "1111111111111111111111111111111111111111111",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateModulo11(tt.number)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestValidateXML_InvalidXML(t *testing.T) {
	validator := NewXSDValidator()

	invalidXML := []byte("invalid xml content")
	result := validator.ValidateXML(invalidXML, "nfe", "4.00")

	if result.Valid {
		t.Error("Expected validation to fail for invalid XML")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for invalid XML")
	}

	// Check that error message mentions XML parsing
	found := false
	for _, err := range result.Errors {
		if strings.Contains(err, "Failed to parse XML") || strings.Contains(err, "XML document has no root element") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'Failed to parse XML' or 'XML document has no root element' error, got: %v", result.Errors)
	}
}

func TestValidateXML_SchemaNotFound(t *testing.T) {
	validator := NewXSDValidator()

	validXML := []byte(`<?xml version="1.0" encoding="UTF-8"?><NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe></infNFe></NFe>`)
	result := validator.ValidateXML(validXML, "nonexistent", "1.00")

	// Should return valid=true when schema doesn't exist (PHP behavior)
	if !result.Valid {
		t.Error("Expected validation to pass when schema doesn't exist")
	}

	// Should have warning about missing schema
	found := false
	for _, err := range result.Errors {
		if strings.Contains(err, "not found - validation skipped") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected schema not found warning, got: %v", result.Errors)
	}
}

func TestValidate_ConvenienceMethod(t *testing.T) {
	validator := NewXSDValidator()

	validXML := []byte(`<?xml version="1.0" encoding="UTF-8"?><NFe xmlns="http://www.portalfiscal.inf.br/nfe"><infNFe></infNFe></NFe>`)

	tests := []struct {
		docType string
		version string
	}{
		{"nfe", "4.00"},
		{"nfce", "4.00"},
		{"envio", "4.00"},
		{"evento", "1.00"},
		{"cce", "1.00"},
		{"inutilizacao", "4.00"},
	}

	for _, tt := range tests {
		t.Run(tt.docType, func(t *testing.T) {
			result := validator.Validate(validXML, tt.docType, tt.version)
			if result == nil {
				t.Error("Expected result, got nil")
			}
		})
	}

	// Test unknown document type
	result := validator.Validate(validXML, "unknown", "4.00")
	if result.Valid {
		t.Error("Expected validation to fail for unknown document type")
	}
}

func TestGetAvailableSchemas(t *testing.T) {
	validator := NewXSDValidator()

	schemas, err := validator.GetAvailableSchemas()
	if err != nil {
		// It's OK if this fails in test environment without schemas
		t.Logf("GetAvailableSchemas failed (expected in test env): %v", err)
		return
	}

	if len(schemas) == 0 {
		t.Log("No schemas found (expected in test environment)")
	}

	// Check that all returned items are .xsd files
	for _, schema := range schemas {
		if !strings.HasSuffix(schema, ".xsd") {
			t.Errorf("Expected all schemas to end with .xsd, got: %s", schema)
		}
	}
}

func TestClearCache(t *testing.T) {
	validator := NewXSDValidator()

	// Add a schema to cache (this would normally happen during LoadSchema)
	validator.schemas["test"] = &Schema{Name: "test"}

	if len(validator.schemas) == 0 {
		t.Error("Expected schema in cache")
	}

	validator.ClearCache()

	if len(validator.schemas) != 0 {
		t.Error("Expected cache to be cleared")
	}
}

func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid:   true,
		Errors:  []string{"test error"},
		Schema:  "nfe",
		Version: "4.00",
	}

	if !result.Valid {
		t.Error("Expected result to be valid")
	}

	if result.Schema != "nfe" {
		t.Errorf("Expected schema 'nfe', got '%s'", result.Schema)
	}

	if result.Version != "4.00" {
		t.Errorf("Expected version '4.00', got '%s'", result.Version)
	}

	if len(result.Errors) != 1 || result.Errors[0] != "test error" {
		t.Errorf("Expected errors ['test error'], got %v", result.Errors)
	}
}
