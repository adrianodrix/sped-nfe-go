// Package utils provides string manipulation utilities for Brazilian NFe system,
// including accent removal, normalization, and formatting functions.
package utils

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// RemoveAccents removes accents and diacritical marks from strings
// This is essential for XML compatibility and SEFAZ compliance
func RemoveAccents(s string) string {
	// Transform using NFD (Canonical Decomposition) and then remove combining marks
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// NormalizeString normalizes a string for NFe XML usage
// Removes accents, converts to uppercase, and removes invalid characters
func NormalizeString(s string) string {
	// Remove accents
	normalized := RemoveAccents(s)
	
	// Convert to uppercase
	normalized = strings.ToUpper(normalized)
	
	// Remove invalid XML characters
	normalized = removeInvalidXMLChars(normalized)
	
	// Trim spaces
	normalized = strings.TrimSpace(normalized)
	
	return normalized
}

// RemoveInvalidXMLChars removes characters that are not valid in XML
func removeInvalidXMLChars(s string) string {
	// Valid XML characters according to XML 1.0 specification:
	// #x9 | #xA | #xD | [#x20-#xD7FF] | [#xE000-#xFFFD] | [#x10000-#x10FFFF]
	var result strings.Builder
	
	for _, r := range s {
		if isValidXMLChar(r) {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// isValidXMLChar checks if a rune is valid in XML 1.0
func isValidXMLChar(r rune) bool {
	return r == 0x09 || r == 0x0A || r == 0x0D ||
		(r >= 0x20 && r <= 0xD7FF) ||
		(r >= 0xE000 && r <= 0xFFFD) ||
		(r >= 0x10000 && r <= 0x10FFFF)
}

// FormatMoney formats a monetary value for NFe XML
// Converts float to string with exactly 2 decimal places
func FormatMoney(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

// FormatQuantity formats a quantity value for NFe XML
// Uses up to 4 decimal places, removing trailing zeros
func FormatQuantity(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', 4, 64)
	
	// Remove trailing zeros after decimal point
	if strings.Contains(formatted, ".") {
		formatted = strings.TrimRight(formatted, "0")
		formatted = strings.TrimRight(formatted, ".")
	}
	
	return formatted
}

// FormatPercentage formats a percentage value for NFe XML
// Uses up to 4 decimal places for tax percentages
func FormatPercentage(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64)
}

// PadLeft pads a string to the left with the specified character
func PadLeft(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	
	padding := strings.Repeat(string(padChar), length-len(s))
	return padding + s
}

// PadRight pads a string to the right with the specified character
func PadRight(s string, length int, padChar rune) string {
	if len(s) >= length {
		return s
	}
	
	padding := strings.Repeat(string(padChar), length-len(s))
	return s + padding
}

// TruncateString truncates a string to the specified maximum length
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	
	// Count runes, not bytes, for proper Unicode handling
	runes := []rune(s)
	if len(runes) <= maxLength {
		return s
	}
	
	return string(runes[:maxLength])
}

// OnlyNumbers extracts only numeric characters from a string
func OnlyNumbers(s string) string {
	re := regexp.MustCompile(`[^0-9]`)
	return re.ReplaceAllString(s, "")
}

// OnlyAlphaNumeric extracts only alphanumeric characters from a string
func OnlyAlphaNumeric(s string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(s, "")
}

// CleanFileName removes invalid characters from a filename
func CleanFileName(filename string) string {
	// Remove invalid filename characters
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	clean := re.ReplaceAllString(filename, "_")
	
	// Remove control characters
	re = regexp.MustCompile(`[\x00-\x1f\x7f]`)
	clean = re.ReplaceAllString(clean, "")
	
	// Trim spaces and dots from the ends
	clean = strings.Trim(clean, " .")
	
	return clean
}

// NormalizeCEP normalizes a Brazilian postal code (CEP)
func NormalizeCEP(cep string) string {
	// Extract only numbers
	numbers := OnlyNumbers(cep)
	
	// Pad with zeros to 8 digits
	return PadLeft(numbers, 8, '0')
}

// FormatCEP formats a CEP with the standard Brazilian mask (XXXXX-XXX)
func FormatCEP(cep string) string {
	normalized := NormalizeCEP(cep)
	if len(normalized) != 8 {
		return cep // Return original if invalid
	}
	
	return normalized[0:5] + "-" + normalized[5:8]
}

// NormalizePhone normalizes a Brazilian phone number
func NormalizePhone(phone string) string {
	// Extract only numbers
	numbers := OnlyNumbers(phone)
	
	// Remove country code if present (55)
	if len(numbers) >= 12 && numbers[0:2] == "55" {
		numbers = numbers[2:]
	}
	
	return numbers
}

// FormatPhone formats a Brazilian phone number with standard mask
func FormatPhone(phone string) string {
	normalized := NormalizePhone(phone)
	
	switch len(normalized) {
	case 10: // Fixed line: (XX) XXXX-XXXX
		return "(" + normalized[0:2] + ") " + normalized[2:6] + "-" + normalized[6:10]
	case 11: // Mobile: (XX) XXXXX-XXXX
		return "(" + normalized[0:2] + ") " + normalized[2:7] + "-" + normalized[7:11]
	default:
		return phone // Return original if invalid
	}
}

// RemoveExtraSpaces removes extra whitespace and normalizes spacing
func RemoveExtraSpaces(s string) string {
	// Replace multiple whitespace characters with single space
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// CapitalizeWords capitalizes the first letter of each word
func CapitalizeWords(s string) string {
	return strings.Title(strings.ToLower(s))
}

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ContainsOnlyDigits checks if string contains only numeric digits
func ContainsOnlyDigits(s string) bool {
	if s == "" {
		return false
	}
	
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	
	return true
}

// ToASCII converts a string to ASCII by removing or replacing non-ASCII characters
func ToASCII(s string) string {
	// First remove accents
	noAccents := RemoveAccents(s)
	
	// Then keep only ASCII characters
	var result strings.Builder
	for _, r := range noAccents {
		if r <= 127 { // ASCII range
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// EscapeXML escapes special XML characters
func EscapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// UnescapeXML unescapes XML-encoded characters
func UnescapeXML(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return s
}

// FormatForXML prepares a string for inclusion in XML
// Removes accents, normalizes, truncates if needed, and escapes
func FormatForXML(s string, maxLength int) string {
	// Normalize the string
	normalized := NormalizeString(s)
	
	// Truncate if necessary
	if maxLength > 0 {
		normalized = TruncateString(normalized, maxLength)
	}
	
	// Escape XML characters
	return EscapeXML(normalized)
}

// ZeroFill pads a numeric string with leading zeros
func ZeroFill(s string, length int) string {
	return PadLeft(s, length, '0')
}

// SpaceFill pads a string with trailing spaces
func SpaceFill(s string, length int) string {
	return PadRight(s, length, ' ')
}