package utils

import (
	"testing"

	"github.com/adrianodrix/sped-nfe-go/types"
)

func TestValidateCNPJ(t *testing.T) {
	tests := []struct {
		cnpj        string
		shouldError bool
		description string
	}{
		{"11.222.333/0001-81", false, "valid CNPJ with formatting"},
		{"11222333000181", false, "valid CNPJ without formatting"},
		{"11.444.777/0001-61", false, "another valid CNPJ"},
		{"11222333000180", true, "invalid check digit"},
		{"11111111111111", true, "all same digits"},
		{"1122233300018", true, "wrong length"},
		{"abc123def456gh", true, "contains letters"},
		{"", true, "empty CNPJ"},
	}

	for _, test := range tests {
		err := ValidateCNPJ(test.cnpj)
		if test.shouldError && err == nil {
			t.Errorf("CNPJ '%s' (%s) should return error", test.cnpj, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("CNPJ '%s' (%s) should not return error, got: %v", test.cnpj, test.description, err)
		}
	}
}

func TestValidateCPF(t *testing.T) {
	tests := []struct {
		cpf         string
		shouldError bool
		description string
	}{
		{"111.444.777-35", false, "valid CPF with formatting"},
		{"11144477735", false, "valid CPF without formatting"},
		{"123.456.789-09", false, "another valid CPF"},
		{"11144477734", true, "invalid check digit"},
		{"11111111111", true, "all same digits"},
		{"123456789", true, "wrong length"},
		{"abc123def45", true, "contains letters"},
		{"", true, "empty CPF"},
	}

	for _, test := range tests {
		err := ValidateCPF(test.cpf)
		if test.shouldError && err == nil {
			t.Errorf("CPF '%s' (%s) should return error", test.cpf, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("CPF '%s' (%s) should not return error, got: %v", test.cpf, test.description, err)
		}
	}
}

func TestValidateIE(t *testing.T) {
	tests := []struct {
		ie          string
		uf          types.UF
		shouldError bool
		description string
	}{
		{"ISENTO", types.SP, false, "exempt IE"},
		{"123456789012", types.SP, false, "valid SP IE format"},
		{"123456789", types.AL, false, "valid AL IE format"},
		{"12345", types.SP, true, "wrong length for SP"},
		{"1234567A", types.RJ, false, "alphanumeric IE"},
		{"", types.SP, true, "empty IE"},
	}

	for _, test := range tests {
		err := ValidateIE(test.ie, test.uf)
		if test.shouldError && err == nil {
			t.Errorf("IE '%s' for %s (%s) should return error", test.ie, test.uf.String(), test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("IE '%s' for %s (%s) should not return error, got: %v", test.ie, test.uf.String(), test.description, err)
		}
	}
}

func TestFormatCNPJ(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"11222333000181", "11.222.333/0001-81", false},
		{"11.222.333/0001-81", "11.222.333/0001-81", false},
		{"11222333000180", "", true}, // Invalid CNPJ
	}

	for _, test := range tests {
		result, err := FormatCNPJ(test.input)
		if test.hasError && err == nil {
			t.Errorf("FormatCNPJ('%s') should return error", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("FormatCNPJ('%s') should not return error, got: %v", test.input, err)
		}
		if !test.hasError && result != test.expected {
			t.Errorf("FormatCNPJ('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestFormatCPF(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"11144477735", "111.444.777-35", false},
		{"111.444.777-35", "111.444.777-35", false},
		{"11144477734", "", true}, // Invalid CPF
	}

	for _, test := range tests {
		result, err := FormatCPF(test.input)
		if test.hasError && err == nil {
			t.Errorf("FormatCPF('%s') should return error", test.input)
		}
		if !test.hasError && err != nil {
			t.Errorf("FormatCPF('%s') should not return error, got: %v", test.input, err)
		}
		if !test.hasError && result != test.expected {
			t.Errorf("FormatCPF('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestCleanDocument(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"111.444.777-35", "11144477735"},
		{"11.222.333/0001-81", "11222333000181"},
		{"abc123def456", "123456"},
		{"  123  456  ", "123456"},
		{"", ""},
	}

	for _, test := range tests {
		result := CleanDocument(test.input)
		if result != test.expected {
			t.Errorf("CleanDocument('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestIsValidDocument(t *testing.T) {
	tests := []struct {
		document    string
		shouldError bool
		description string
	}{
		{"11144477735", false, "valid CPF"},
		{"11222333000181", false, "valid CNPJ"},
		{"111.444.777-35", false, "valid CPF with formatting"},
		{"11.222.333/0001-81", false, "valid CNPJ with formatting"},
		{"123456789", true, "invalid length"},
		{"11144477734", true, "invalid CPF"},
		{"11222333000180", true, "invalid CNPJ"},
	}

	for _, test := range tests {
		err := IsValidDocument(test.document)
		if test.shouldError && err == nil {
			t.Errorf("IsValidDocument('%s') (%s) should return error", test.document, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("IsValidDocument('%s') (%s) should not return error, got: %v", test.document, test.description, err)
		}
	}
}

func TestCalculateCNPJCheckDigit(t *testing.T) {
	tests := []struct {
		digits   []int
		weights  []int
		expected int
	}{
		{[]int{1, 1, 2, 2, 2, 3, 3, 3, 0, 0, 0, 1}, []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}, 8},
		{[]int{1, 1, 2, 2, 2, 3, 3, 3, 0, 0, 0, 1, 8}, []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}, 1},
	}

	for _, test := range tests {
		result := calculateCNPJCheckDigit(test.digits, test.weights)
		if result != test.expected {
			t.Errorf("calculateCNPJCheckDigit(%v, %v) = %d, expected %d", test.digits, test.weights, result, test.expected)
		}
	}
}

func TestIsAllSameDigits(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"11111111111111", true},
		{"00000000000000", true},
		{"11144477735", false},
		{"", false},
		{"1", true},
	}

	for _, test := range tests {
		result := isAllSameDigits(test.input)
		if result != test.expected {
			t.Errorf("isAllSameDigits('%s') = %t, expected %t", test.input, result, test.expected)
		}
	}
}

func TestGetIEExpectedLengths(t *testing.T) {
	lengths := getIEExpectedLengths()

	// Test a few known values
	if lengths[types.SP] != 12 {
		t.Errorf("Expected SP IE length to be 12, got %d", lengths[types.SP])
	}

	if lengths[types.RJ] != 8 {
		t.Errorf("Expected RJ IE length to be 8, got %d", lengths[types.RJ])
	}

	if lengths[types.MG] != 13 {
		t.Errorf("Expected MG IE length to be 13, got %d", lengths[types.MG])
	}

	// Ensure all valid UFs are covered
	expectedCount := 27 // All states including EX
	if len(lengths) != expectedCount {
		t.Errorf("Expected %d UFs in IE lengths map, got %d", expectedCount, len(lengths))
	}
}
