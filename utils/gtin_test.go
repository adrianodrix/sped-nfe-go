package utils

import (
	"testing"
)

func TestValidateGTIN(t *testing.T) {
	tests := []struct {
		gtin        string
		shouldError bool
		description string
	}{
		{"", false, "empty GTIN"},
		{"SEM GTIN", false, "SEM GTIN"},
		{"sem gtin", false, "sem gtin lowercase"},
		{"7891000100103", false, "valid EAN-13"},
		{"036000291452", false, "valid UPC-A"},
		{"12345670", false, "valid EAN-8"},
		{"00036000291452", false, "valid GTIN-14"},
		{"7891000100104", true, "invalid EAN-13 check digit"},
		{"12345678901", true, "invalid length (11 digits)"},
		{"1234567890123456", true, "invalid length (16 digits)"},
		{"1111111111111", true, "all same digits"},
		{"0000000000000", false, "all zeros (valid)"},
		{"abc123def456", true, "contains letters"},
		{"123-456-789-012", false, "UPC-A with separators"},
		{"7891000-100103", false, "EAN-13 with separator"},
	}

	for _, test := range tests {
		err := ValidateGTIN(test.gtin)
		if test.shouldError && err == nil {
			t.Errorf("ValidateGTIN('%s') (%s) should return error", test.gtin, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("ValidateGTIN('%s') (%s) should not return error, got: %v", test.gtin, test.description, err)
		}
	}
}

func TestFormatGTIN(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldError bool
		description string
	}{
		{"12345670", "1234-5670", false, "EAN-8 formatting"},
		{"036000291452", "0-36000-29145-2", false, "UPC-A formatting"},
		{"7891000100103", "7-891000-10010-3", false, "EAN-13 formatting"},
		{"00036000291452", "00-036000-29145-2", false, "GTIN-14 formatting"},
		{"SEM GTIN", "SEM GTIN", false, "SEM GTIN passthrough"},
		{"", "", false, "empty GTIN"},
		{"123-456-789-012", "1-23456-78901-2", false, "UPC-A with existing separators"},
		{"1234567890123", "", true, "invalid check digit"},
	}

	for _, test := range tests {
		result, err := FormatGTIN(test.input)
		if test.shouldError && err == nil {
			t.Errorf("FormatGTIN('%s') (%s) should return error", test.input, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("FormatGTIN('%s') (%s) should not return error, got: %v", test.input, test.description, err)
		}
		if !test.shouldError && result != test.expected {
			t.Errorf("FormatGTIN('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestIsGTINEmpty(t *testing.T) {
	tests := []struct {
		gtin     string
		expected bool
	}{
		{"", true},
		{"SEM GTIN", true},
		{"sem gtin", true},
		{"  SEM GTIN  ", true},
		{"7891000100103", false},
		{"123456789012", false},
		{"SEMGTIN", false}, // Without space
		{"OTHER", false},
	}

	for _, test := range tests {
		result := IsGTINEmpty(test.gtin)
		if result != test.expected {
			t.Errorf("IsGTINEmpty('%s') = %t, expected %t", test.gtin, result, test.expected)
		}
	}
}

func TestGetGTINType(t *testing.T) {
	tests := []struct {
		gtin        string
		expected    string
		shouldError bool
	}{
		{"", "EMPTY", false},
		{"SEM GTIN", "EMPTY", false},
		{"12345670", "EAN-8", false},
		{"036000291452", "UPC-A", false},
		{"7891000100103", "EAN-13", false},
		{"00036000291452", "GTIN-14", false},
		{"123-456-789-012", "UPC-A", false}, // With separators
		{"12345", "", true}, // Invalid length
		{"123456789012345", "", true}, // Invalid length
	}

	for _, test := range tests {
		result, err := GetGTINType(test.gtin)
		if test.shouldError && err == nil {
			t.Errorf("GetGTINType('%s') should return error", test.gtin)
		}
		if !test.shouldError && err != nil {
			t.Errorf("GetGTINType('%s') should not return error, got: %v", test.gtin, err)
		}
		if !test.shouldError && result != test.expected {
			t.Errorf("GetGTINType('%s') = '%s', expected '%s'", test.gtin, result, test.expected)
		}
	}
}

func TestIsValidGTINCheckDigit(t *testing.T) {
	tests := []struct {
		gtin     string
		expected bool
	}{
		{"7891000100103", true},  // Valid EAN-13
		{"036000291452", true},   // Valid UPC-A
		{"12345670", true},       // Valid EAN-8
		{"00036000291452", true}, // Valid GTIN-14
		{"7891000100104", false}, // Invalid EAN-13
		{"123456789013", false},  // Invalid UPC-A
		{"12345671", false},      // Invalid EAN-8
		{"01234567890129", false}, // Invalid GTIN-14
		{"1234567", false},       // Too short
	}

	for _, test := range tests {
		result := isValidGTINCheckDigit(test.gtin)
		if result != test.expected {
			t.Errorf("isValidGTINCheckDigit('%s') = %t, expected %t", test.gtin, result, test.expected)
		}
	}
}

func TestCalculateGTINCheckDigit(t *testing.T) {
	tests := []struct {
		partial  string
		expected int
	}{
		{"789100010010", 3},  // EAN-13 partial
		{"03600029145", 2},   // UPC-A partial
		{"1234567", 0},      // EAN-8 partial
		{"0003600029145", 2}, // GTIN-14 partial
	}

	for _, test := range tests {
		result := calculateGTINCheckDigit(test.partial)
		if result != test.expected {
			t.Errorf("calculateGTINCheckDigit('%s') = %d, expected %d", test.partial, result, test.expected)
		}
	}
}

func TestConvertToGTIN14(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldError bool
		description string
	}{
		{"12345670", "00000012345670", false, "EAN-8 to GTIN-14"},
		{"036000291452", "00036000291452", false, "UPC-A to GTIN-14"},
		{"7891000100103", "07891000100103", false, "EAN-13 to GTIN-14"},
		{"00036000291452", "00036000291452", false, "GTIN-14 unchanged"},
		{"SEM GTIN", "SEM GTIN", false, "SEM GTIN passthrough"},
		{"", "SEM GTIN", false, "empty to SEM GTIN"},
		{"1234567890123", "", true, "invalid check digit"},
	}

	for _, test := range tests {
		result, err := ConvertToGTIN14(test.input)
		if test.shouldError && err == nil {
			t.Errorf("ConvertToGTIN14('%s') (%s) should return error", test.input, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("ConvertToGTIN14('%s') (%s) should not return error, got: %v", test.input, test.description, err)
		}
		if !test.shouldError && result != test.expected {
			t.Errorf("ConvertToGTIN14('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestConvertFromGTIN14(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldError bool
		description string
	}{
		{"00000012345670", "12345670", false, "GTIN-14 to EAN-8"},
		{"00036000291452", "036000291452", false, "GTIN-14 to UPC-A"},
		{"07891000100103", "7891000100103", false, "GTIN-14 to EAN-13"},
		{"10036000291459", "10036000291459", false, "GTIN-14 unchanged"},
		{"00000000000000", "0", false, "all zeros"},
		{"SEM GTIN", "SEM GTIN", false, "SEM GTIN passthrough"},
		{"123456789012", "", true, "not GTIN-14 length"},
		{"1234567890123", "", true, "invalid check digit"},
	}

	for _, test := range tests {
		result, err := ConvertFromGTIN14(test.input)
		if test.shouldError && err == nil {
			t.Errorf("ConvertFromGTIN14('%s') (%s) should return error", test.input, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("ConvertFromGTIN14('%s') (%s) should not return error, got: %v", test.input, test.description, err)
		}
		if !test.shouldError && result != test.expected {
			t.Errorf("ConvertFromGTIN14('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestGenerateGTINCheckDigit(t *testing.T) {
	tests := []struct {
		partial     string
		expected    string
		shouldError bool
		description string
	}{
		{"789100010010", "7891000100103", false, "EAN-13 partial"},
		{"03600029145", "036000291452", false, "UPC-A partial"},
		{"1234567", "12345670", false, "EAN-8 partial"},
		{"0003600029145", "00036000291452", false, "GTIN-14 partial"},
		{"123456", "", true, "too short"},
		{"12345678901234", "", true, "too long"},
		{"abc123def", "", true, "contains letters"},
	}

	for _, test := range tests {
		result, err := GenerateGTINCheckDigit(test.partial)
		if test.shouldError && err == nil {
			t.Errorf("GenerateGTINCheckDigit('%s') (%s) should return error", test.partial, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("GenerateGTINCheckDigit('%s') (%s) should not return error, got: %v", test.partial, test.description, err)
		}
		if !test.shouldError && result != test.expected {
			t.Errorf("GenerateGTINCheckDigit('%s') = '%s', expected '%s'", test.partial, result, test.expected)
		}
	}
}

func TestFormatEAN8(t *testing.T) {
	result := formatEAN8("12345670")
	expected := "1234-5670"
	if result != expected {
		t.Errorf("formatEAN8('12345670') = '%s', expected '%s'", result, expected)
	}
}

func TestFormatUPCA(t *testing.T) {
	result := formatUPCA("036000291452")
	expected := "0-36000-29145-2"
	if result != expected {
		t.Errorf("formatUPCA('036000291452') = '%s', expected '%s'", result, expected)
	}
}

func TestFormatEAN13(t *testing.T) {
	result := formatEAN13("7891000100103")
	expected := "7-891000-10010-3"
	if result != expected {
		t.Errorf("formatEAN13('7891000100103') = '%s', expected '%s'", result, expected)
	}
}

func TestFormatGTIN14(t *testing.T) {
	result := formatGTIN14("00036000291452")
	expected := "00-036000-29145-2"
	if result != expected {
		t.Errorf("formatGTIN14('00036000291452') = '%s', expected '%s'", result, expected)
	}
}