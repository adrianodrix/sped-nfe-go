// Package utils provides GTIN (Global Trade Item Number) validation utilities.
// GTIN is used for EAN-8, EAN-13, UPC-A, UPC-E, and other product codes.
package utils

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/adrianodrix/sped-nfe-go/errors"
)

// ValidateGTIN validates a GTIN code (Global Trade Item Number)
// Supports GTIN-8, GTIN-12, GTIN-13, and GTIN-14 with check digit validation
func ValidateGTIN(gtin string) error {
	// Handle special cases
	if gtin == "" || strings.ToUpper(gtin) == "SEM GTIN" {
		return nil // Empty or "SEM GTIN" is considered valid
	}
	
	// Remove non-numeric characters
	gtinClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(gtin, "")
	
	// Check supported lengths
	validLengths := []int{8, 12, 13, 14}
	isValidLength := false
	for _, length := range validLengths {
		if len(gtinClean) == length {
			isValidLength = true
			break
		}
	}
	
	if !isValidLength {
		return errors.NewValidationError("GTIN must be 8, 12, 13, or 14 digits", "gtin", gtin)
	}
	
	// Check if all digits are the same (invalid), except for all zeros
	if isAllSameDigits(gtinClean) && gtinClean != strings.Repeat("0", len(gtinClean)) {
		return errors.NewValidationError("GTIN cannot have all same digits", "gtin", gtin)
	}
	
	// Validate check digit
	if !isValidGTINCheckDigit(gtinClean) {
		return errors.NewValidationError("invalid GTIN check digit", "gtin", gtin)
	}
	
	return nil
}

// FormatGTIN formats a GTIN code with appropriate separators
func FormatGTIN(gtin string) (string, error) {
	// Validate first
	if err := ValidateGTIN(gtin); err != nil {
		return "", err
	}
	
	// Clean the GTIN
	gtinClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(gtin, "")
	
	// Handle special cases
	if gtin == "" || strings.ToUpper(gtin) == "SEM GTIN" {
		return strings.ToUpper(gtin), nil
	}
	
	// Format based on length
	switch len(gtinClean) {
	case 8: // EAN-8
		return formatEAN8(gtinClean), nil
	case 12: // UPC-A
		return formatUPCA(gtinClean), nil
	case 13: // EAN-13
		return formatEAN13(gtinClean), nil
	case 14: // GTIN-14
		return formatGTIN14(gtinClean), nil
	default:
		return gtinClean, nil
	}
}

// IsGTINEmpty checks if GTIN is empty or "SEM GTIN"
func IsGTINEmpty(gtin string) bool {
	return gtin == "" || strings.ToUpper(strings.TrimSpace(gtin)) == "SEM GTIN"
}

// GetGTINType returns the type of GTIN based on its length
func GetGTINType(gtin string) (string, error) {
	if IsGTINEmpty(gtin) {
		return "EMPTY", nil
	}
	
	gtinClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(gtin, "")
	
	switch len(gtinClean) {
	case 8:
		return "EAN-8", nil
	case 12:
		return "UPC-A", nil
	case 13:
		return "EAN-13", nil
	case 14:
		return "GTIN-14", nil
	default:
		return "", errors.NewValidationError("invalid GTIN length", "gtin", gtin)
	}
}

// Helper functions

// isValidGTINCheckDigit validates the GTIN check digit using the standard algorithm
func isValidGTINCheckDigit(gtin string) bool {
	if len(gtin) < 8 {
		return false
	}
	
	// Calculate check digit for the first n-1 digits
	calculatedCheckDigit := calculateGTINCheckDigit(gtin[:len(gtin)-1])
	
	// Get the actual check digit (last digit)
	actualCheckDigit, err := strconv.Atoi(string(gtin[len(gtin)-1]))
	if err != nil {
		return false
	}
	
	return calculatedCheckDigit == actualCheckDigit
}

// calculateGTINCheckDigit calculates the check digit for a GTIN
func calculateGTINCheckDigit(partialGTIN string) int {
	sum := 0
	
	// Calculate weighted sum from right to left
	// Odd positions (from right) get weight 3, even positions get weight 1
	for i, digit := range partialGTIN {
		digitValue, _ := strconv.Atoi(string(digit))
		
		// Position from right (1-indexed)
		positionFromRight := len(partialGTIN) - i
		
		// Odd positions from right get weight 3, even positions get weight 1
		weight := 1
		if positionFromRight%2 == 1 {
			weight = 3
		}
		
		sum += digitValue * weight
	}
	
	// Calculate check digit
	remainder := sum % 10
	if remainder == 0 {
		return 0
	}
	return 10 - remainder
}

// Formatting functions

// formatEAN8 formats an 8-digit EAN code
func formatEAN8(ean8 string) string {
	// Format as XXXX-XXXX
	return ean8[0:4] + "-" + ean8[4:8]
}

// formatUPCA formats a 12-digit UPC-A code
func formatUPCA(upca string) string {
	// Format as X-XXXXX-XXXXX-X
	return upca[0:1] + "-" + upca[1:6] + "-" + upca[6:11] + "-" + upca[11:12]
}

// formatEAN13 formats a 13-digit EAN code
func formatEAN13(ean13 string) string {
	// Format as X-XXXXXX-XXXXXX-X
	return ean13[0:1] + "-" + ean13[1:7] + "-" + ean13[7:12] + "-" + ean13[12:13]
}

// formatGTIN14 formats a 14-digit GTIN code
func formatGTIN14(gtin14 string) string {
	// Format as XX-XXXXXX-XXXXXX-X
	return gtin14[0:2] + "-" + gtin14[2:8] + "-" + gtin14[8:13] + "-" + gtin14[13:14]
}

// ConvertToGTIN14 converts any valid GTIN to GTIN-14 format by padding with zeros
func ConvertToGTIN14(gtin string) (string, error) {
	if err := ValidateGTIN(gtin); err != nil {
		return "", err
	}
	
	if IsGTINEmpty(gtin) {
		return "SEM GTIN", nil
	}
	
	gtinClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(gtin, "")
	
	// Pad to 14 digits with leading zeros
	gtin14 := strings.Repeat("0", 14-len(gtinClean)) + gtinClean
	
	return gtin14, nil
}

// ConvertFromGTIN14 converts a GTIN-14 to its shorter equivalent if possible
func ConvertFromGTIN14(gtin14 string) (string, error) {
	if err := ValidateGTIN(gtin14); err != nil {
		return "", err
	}
	
	if IsGTINEmpty(gtin14) {
		return gtin14, nil
	}
	
	gtinClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(gtin14, "")
	
	if len(gtinClean) != 14 {
		return "", errors.NewValidationError("input must be a 14-digit GTIN", "gtin14", gtin14)
	}
	
	// Remove leading zeros to find the shortest valid representation
	trimmed := strings.TrimLeft(gtinClean, "0")
	
	// If all zeros, return "0"
	if trimmed == "" {
		return "0", nil
	}
	
	// Determine the appropriate length based on the first digit
	switch len(trimmed) {
	case 1, 2, 3, 4, 5, 6, 7:
		// Pad to 8 digits for EAN-8
		return strings.Repeat("0", 8-len(trimmed)) + trimmed, nil
	case 8:
		return trimmed, nil // Already EAN-8
	case 9, 10, 11:
		// Pad to 12 digits for UPC-A
		return strings.Repeat("0", 12-len(trimmed)) + trimmed, nil
	case 12:
		return trimmed, nil // Already UPC-A
	case 13:
		return trimmed, nil // EAN-13
	case 14:
		return trimmed, nil // GTIN-14
	default:
		return trimmed, nil
	}
}

// GenerateGTINCheckDigit generates the check digit for a partial GTIN
func GenerateGTINCheckDigit(partialGTIN string) (string, error) {
	// Validate input
	gtinClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(partialGTIN, "")
	
	if len(gtinClean) < 7 || len(gtinClean) > 13 {
		return "", errors.NewValidationError("partial GTIN must be 7-13 digits", "partialGTIN", partialGTIN)
	}
	
	checkDigit := calculateGTINCheckDigit(gtinClean)
	return gtinClean + strconv.Itoa(checkDigit), nil
}