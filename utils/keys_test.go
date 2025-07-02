package utils

import (
	"testing"
	"time"

	"github.com/adrianodrix/sped-nfe-go/types"
)

func TestGenerateAccessKey(t *testing.T) {
	// Test valid key generation
	components := NFEKeyComponents{
		UF:       types.SP,
		DateTime: time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC),
		CNPJ:     "11222333000181",
		Model:    types.ModeloNFe55,
		Series:   1,
		Number:   123456,
		EmitType: types.TeNormal,
	}

	key, err := GenerateAccessKey(components)
	if err != nil {
		t.Errorf("GenerateAccessKey should not return error for valid components, got: %v", err)
	}

	if len(key) != 44 {
		t.Errorf("Generated key should have 44 digits, got %d", len(key))
	}

	// Test key validation
	if err := ValidateAccessKey(key); err != nil {
		t.Errorf("Generated key should be valid, got error: %v", err)
	}
}

func TestGenerateAccessKeyWithCustomCode(t *testing.T) {
	code := 12345678
	components := NFEKeyComponents{
		UF:       types.RJ,
		DateTime: time.Date(2023, 12, 25, 15, 45, 0, 0, time.UTC),
		CNPJ:     "11444777000161",
		Model:    types.ModeloNFCe65,
		Series:   5,
		Number:   987654,
		EmitType: types.TeNormal,
		Code:     &code,
	}

	key, err := GenerateAccessKey(components)
	if err != nil {
		t.Errorf("GenerateAccessKey with custom code should not return error, got: %v", err)
	}

	// Parse the key and verify the code
	parsed, err := ParseAccessKey(key)
	if err != nil {
		t.Errorf("ParseAccessKey should not return error, got: %v", err)
	}

	if *parsed.Code != code {
		t.Errorf("Parsed code should be %d, got %d", code, *parsed.Code)
	}
}

func TestGenerateAccessKeyInvalidInputs(t *testing.T) {
	tests := []struct {
		components  NFEKeyComponents
		description string
	}{
		{
			NFEKeyComponents{UF: types.UF(999), DateTime: time.Now(), CNPJ: "11222333000181", Model: types.ModeloNFe55, Series: 1, Number: 123},
			"invalid UF",
		},
		{
			NFEKeyComponents{UF: types.SP, DateTime: time.Now(), CNPJ: "11222333000181", Model: types.ModeloNFe(99), Series: 1, Number: 123},
			"invalid model",
		},
		{
			NFEKeyComponents{UF: types.SP, DateTime: time.Now(), CNPJ: "11222333000181", Model: types.ModeloNFe55, Series: 1000, Number: 123},
			"invalid series",
		},
		{
			NFEKeyComponents{UF: types.SP, DateTime: time.Now(), CNPJ: "11222333000181", Model: types.ModeloNFe55, Series: 1, Number: 0},
			"invalid number (zero)",
		},
		{
			NFEKeyComponents{UF: types.SP, DateTime: time.Now(), CNPJ: "123456789", Model: types.ModeloNFe55, Series: 1, Number: 123},
			"invalid CNPJ length",
		},
	}

	for _, test := range tests {
		_, err := GenerateAccessKey(test.components)
		if err == nil {
			t.Errorf("GenerateAccessKey should return error for %s", test.description)
		}
	}
}

func TestValidateAccessKey(t *testing.T) {
	tests := []struct {
		key         string
		shouldError bool
		description string
	}{
		{"35230511222333000181550010000123456781234567", true, "invalid check digit"},
		{"3523051122233300018155001000012345678123456", true, "wrong length (43 digits)"},
		{"352305112223330001815500100001234567812345678", true, "wrong length (45 digits)"},
		{"35230511222333000181550010000123456781234abc", true, "contains letters"},
		{"", true, "empty key"},
		{"  352305112223330001815500100001234567812345678  ", true, "with spaces"},
	}

	for _, test := range tests {
		err := ValidateAccessKey(test.key)
		if test.shouldError && err == nil {
			t.Errorf("ValidateAccessKey('%s') (%s) should return error", test.key, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("ValidateAccessKey('%s') (%s) should not return error, got: %v", test.key, test.description, err)
		}
	}
}

func TestParseAccessKey(t *testing.T) {
	// First generate a valid key
	components := NFEKeyComponents{
		UF:       types.MG,
		DateTime: time.Date(2023, 8, 10, 0, 0, 0, 0, time.UTC),
		CNPJ:     "12345678000195",
		Model:    types.ModeloNFe55,
		Series:   3,
		Number:   555666,
		EmitType: types.TeNormal,
	}

	key, err := GenerateAccessKey(components)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Parse the key
	parsed, err := ParseAccessKey(key)
	if err != nil {
		t.Errorf("ParseAccessKey should not return error, got: %v", err)
	}

	// Verify components
	if parsed.UF != components.UF {
		t.Errorf("Parsed UF should be %v, got %v", components.UF, parsed.UF)
	}

	if parsed.DateTime.Year() != components.DateTime.Year() {
		t.Errorf("Parsed year should be %d, got %d", components.DateTime.Year(), parsed.DateTime.Year())
	}

	if parsed.DateTime.Month() != components.DateTime.Month() {
		t.Errorf("Parsed month should be %d, got %d", components.DateTime.Month(), parsed.DateTime.Month())
	}

	if parsed.Model != components.Model {
		t.Errorf("Parsed model should be %v, got %v", components.Model, parsed.Model)
	}

	if parsed.Series != components.Series {
		t.Errorf("Parsed series should be %d, got %d", components.Series, parsed.Series)
	}

	if parsed.Number != components.Number {
		t.Errorf("Parsed number should be %d, got %d", components.Number, parsed.Number)
	}
}

func TestFormatAccessKey(t *testing.T) {
	// Generate a valid key for testing
	components := NFEKeyComponents{
		UF:       types.SP,
		DateTime: time.Now(),
		CNPJ:     "11222333000181",
		Model:    types.ModeloNFe55,
		Series:   1,
		Number:   123456,
		EmitType: types.TeNormal,
	}

	key, err := GenerateAccessKey(components)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	formatted, err := FormatAccessKey(key)
	if err != nil {
		t.Errorf("FormatAccessKey should not return error, got: %v", err)
	}

	// Check format (should have spaces every 4 digits)
	expectedLength := 54 // 44 digits + 10 spaces
	if len(formatted) != expectedLength {
		t.Errorf("Formatted key should have %d characters, got %d", expectedLength, len(formatted))
	}

	// Remove spaces and verify it's the same key
	unformatted := ""
	for _, char := range formatted {
		if char != ' ' {
			unformatted += string(char)
		}
	}

	if unformatted != key {
		t.Errorf("Unformatted key should match original key")
	}
}

func TestCalculateCheckDigit(t *testing.T) {
	tests := []struct {
		key43    string
		expected string
	}{
		{"3523051122233300018155001000012345678123456", "1"}, // Valid test case
		{"3323051122233300018155001000012345678123456", "7"}, // Different UF
	}

	for _, test := range tests {
		result := calculateCheckDigit(test.key43)
		if result != test.expected {
			t.Errorf("calculateCheckDigit('%s') = '%s', expected '%s'", test.key43, result, test.expected)
		}
	}
}

func TestCalculateCheckDigitInvalidLength(t *testing.T) {
	// Test with invalid length
	result := calculateCheckDigit("123")
	if result != "" {
		t.Errorf("calculateCheckDigit with invalid length should return empty string, got '%s'", result)
	}
}

func TestGenerateRandomCode(t *testing.T) {
	nfeNumber := 123456

	code, err := generateRandomCode(nfeNumber)
	if err != nil {
		t.Errorf("generateRandomCode should not return error, got: %v", err)
	}

	// Check that code is different from NFe number
	if code == nfeNumber {
		t.Errorf("Generated code should be different from NFe number")
	}

	// Check that code is in valid range
	if code < 10000000 || code > 99999999 {
		t.Errorf("Generated code should be 8 digits, got %d", code)
	}
}

func TestIsValidModel(t *testing.T) {
	tests := []struct {
		model    types.ModeloNFe
		expected bool
	}{
		{types.ModeloNFe55, true},
		{types.ModeloNFCe65, true},
		{types.ModeloNFe(99), false},
		{types.ModeloNFe(0), false},
	}

	for _, test := range tests {
		result := isValidModel(test.model)
		if result != test.expected {
			t.Errorf("isValidModel(%v) = %t, expected %t", test.model, result, test.expected)
		}
	}
}

func TestGetKeyComponents(t *testing.T) {
	// Generate a valid key
	originalComponents := NFEKeyComponents{
		UF:       types.BA,
		DateTime: time.Date(2023, 7, 20, 0, 0, 0, 0, time.UTC),
		CNPJ:     "98765432000112",
		Model:    types.ModeloNFCe65,
		Series:   2,
		Number:   789123,
		EmitType: types.TeNormal,
	}

	key, err := GenerateAccessKey(originalComponents)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	// Get components back
	components, err := GetKeyComponents(key)
	if err != nil {
		t.Errorf("GetKeyComponents should not return error, got: %v", err)
	}

	// Verify key components match
	if components.UF != originalComponents.UF {
		t.Errorf("UF should match: expected %v, got %v", originalComponents.UF, components.UF)
	}

	if components.Model != originalComponents.Model {
		t.Errorf("Model should match: expected %v, got %v", originalComponents.Model, components.Model)
	}

	if components.Series != originalComponents.Series {
		t.Errorf("Series should match: expected %d, got %d", originalComponents.Series, components.Series)
	}

	if components.Number != originalComponents.Number {
		t.Errorf("Number should match: expected %d, got %d", originalComponents.Number, components.Number)
	}
}

func TestBuildKey43(t *testing.T) {
	components := NFEKeyComponents{
		UF:       types.SP,
		DateTime: time.Date(2023, 5, 15, 0, 0, 0, 0, time.UTC),
		CNPJ:     "11222333000181",
		Model:    types.ModeloNFe55,
		Series:   1,
		Number:   123456,
		EmitType: types.TeNormal,
	}

	cnpj := "11222333000181"
	code := 12345678

	key43 := buildKey43(components, cnpj, code)

	if len(key43) != 43 {
		t.Errorf("Key43 should have 43 digits, got %d", len(key43))
	}

	// Verify it starts with the correct UF (35 = SP)
	if key43[0:2] != "35" {
		t.Errorf("Key should start with UF code 35, got %s", key43[0:2])
	}

	// Verify year and month
	if key43[2:4] != "23" { // 2023 -> 23
		t.Errorf("Key should contain year 23, got %s", key43[2:4])
	}

	if key43[4:6] != "05" { // May -> 05
		t.Errorf("Key should contain month 05, got %s", key43[4:6])
	}
}