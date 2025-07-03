// Package nfe provides utilities for NFe generation and formatting.
package nfe

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// FormatValue formats a float value to string with specified decimal places
func FormatValue(value float64, decimals int) string {
	format := fmt.Sprintf("%%.%df", decimals)
	return fmt.Sprintf(format, value)
}

// FormatCurrency formats a value as currency (2 decimal places)
func FormatCurrency(value float64) string {
	return FormatValue(value, 2)
}

// FormatQuantity formats a quantity value (4 decimal places)
func FormatQuantity(value float64) string {
	return FormatValue(value, 4)
}

// FormatPercentage formats a percentage value (4 decimal places)
func FormatPercentage(value float64) string {
	return FormatValue(value, 4)
}

// ParseValue parses a string value to float64
func ParseValue(value string) (float64, error) {
	if value == "" {
		return 0.0, nil
	}

	// Replace comma with dot for decimal separator
	value = strings.Replace(value, ",", ".", -1)

	return strconv.ParseFloat(value, 64)
}

// FormatCNPJ formats CNPJ with dots and slash
func FormatCNPJ(cnpj string) string {
	if len(cnpj) != 14 {
		return cnpj
	}

	return fmt.Sprintf("%s.%s.%s/%s-%s",
		cnpj[0:2], cnpj[2:5], cnpj[5:8], cnpj[8:12], cnpj[12:14])
}

// FormatCPF formats CPF with dots and dash
func FormatCPF(cpf string) string {
	if len(cpf) != 11 {
		return cpf
	}

	return fmt.Sprintf("%s.%s.%s-%s",
		cpf[0:3], cpf[3:6], cpf[6:9], cpf[9:11])
}

// FormatCEP formats CEP with dash
func FormatCEP(cep string) string {
	if len(cep) != 8 {
		return cep
	}

	return fmt.Sprintf("%s-%s", cep[0:5], cep[5:8])
}

// FormatPhone formats phone number
func FormatPhone(phone string) string {
	phone = OnlyNumbers(phone)

	switch len(phone) {
	case 10:
		return fmt.Sprintf("(%s) %s-%s", phone[0:2], phone[2:6], phone[6:10])
	case 11:
		return fmt.Sprintf("(%s) %s-%s", phone[0:2], phone[2:7], phone[7:11])
	default:
		return phone
	}
}

// OnlyNumbers removes all non-numeric characters
func OnlyNumbers(str string) string {
	reg := regexp.MustCompile(`[^0-9]`)
	return reg.ReplaceAllString(str, "")
}

// OnlyLetters removes all non-letter characters
func OnlyLetters(str string) string {
	reg := regexp.MustCompile(`[^a-zA-ZÀ-ÿ\s]`)
	return reg.ReplaceAllString(str, "")
}

// OnlyAlphanumeric removes all non-alphanumeric characters
func OnlyAlphanumeric(str string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9À-ÿ\s]`)
	return reg.ReplaceAllString(str, "")
}

// RemoveAccents removes accents from text
func RemoveAccents(str string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), norm.NFC)
	result, _, _ := transform.String(t, str)
	return result
}

func isMn(r rune) bool {
	return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
}

// TruncateString truncates string to max length
func TruncateString(str string, maxLength int) string {
	if len(str) <= maxLength {
		return str
	}
	return str[:maxLength]
}

// PadLeft pads string to the left with specified character
func PadLeft(str string, length int, padChar rune) string {
	if len(str) >= length {
		return str
	}

	padding := strings.Repeat(string(padChar), length-len(str))
	return padding + str
}

// PadRight pads string to the right with specified character
func PadRight(str string, length int, padChar rune) string {
	if len(str) >= length {
		return str
	}

	padding := strings.Repeat(string(padChar), length-len(str))
	return str + padding
}

// FormatDateTime formats time to NFe datetime format
func FormatDateTime(t time.Time) string {
	// Format: AAAA-MM-DDTHH:mm:ssTZD
	return t.Format("2006-01-02T15:04:05-07:00")
}

// FormatDate formats time to NFe date format
func FormatDate(t time.Time) string {
	// Format: AAAA-MM-DD
	return t.Format("2006-01-02")
}

// FormatTime formats time to NFe time format
func FormatTime(t time.Time) string {
	// Format: HH:mm:ss
	return t.Format("15:04:05")
}

// FormatYearMonth formats time to year-month format (AAMM)
func FormatYearMonth(t time.Time) string {
	year := t.Year() % 100 // Last 2 digits of year
	month := int(t.Month())
	return fmt.Sprintf("%02d%02d", year, month)
}

// ParseDateTime parses NFe datetime format
func ParseDateTime(dateTime string) (time.Time, error) {
	// Try different formats
	formats := []string{
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateTime); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime format: %s", dateTime)
}

// GenerateRandomCode generates a random numeric code
func GenerateRandomCode(length int) string {
	result := make([]byte, length)

	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(10))
		result[i] = byte('0' + num.Int64())
	}

	return string(result)
}

// CalculateModulo11 calculates modulo 11 check digit
func CalculateModulo11(number string) int {
	sum := 0
	weight := 2

	// Process from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))
		sum += digit * weight
		weight++
		if weight > 9 {
			weight = 2
		}
	}

	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}

// CalculateModulo10 calculates modulo 10 check digit (Luhn algorithm)
func CalculateModulo10(number string) int {
	sum := 0
	alternate := false

	// Process from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit/10 + digit%10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return (10 - (sum % 10)) % 10
}

// ValidateCNPJ validates CNPJ format and check digits
func ValidateCNPJ(cnpj string) bool {
	cnpj = OnlyNumbers(cnpj)

	if len(cnpj) != 14 {
		return false
	}

	// Check for repeated digits
	if cnpj == strings.Repeat(string(cnpj[0]), 14) {
		return false
	}

	// Validate first check digit
	sum := 0
	weights := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	for i := 0; i < 12; i++ {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights[i]
	}

	remainder := sum % 11
	firstDigit := 0
	if remainder >= 2 {
		firstDigit = 11 - remainder
	}

	expectedFirst, _ := strconv.Atoi(string(cnpj[12]))
	if firstDigit != expectedFirst {
		return false
	}

	// Validate second check digit
	sum = 0
	weights = []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}

	for i := 0; i < 13; i++ {
		digit, _ := strconv.Atoi(string(cnpj[i]))
		sum += digit * weights[i]
	}

	remainder = sum % 11
	secondDigit := 0
	if remainder >= 2 {
		secondDigit = 11 - remainder
	}

	expectedSecond, _ := strconv.Atoi(string(cnpj[13]))
	return secondDigit == expectedSecond
}

// ValidateCPF validates CPF format and check digits
func ValidateCPF(cpf string) bool {
	cpf = OnlyNumbers(cpf)

	if len(cpf) != 11 {
		return false
	}

	// Check for repeated digits
	if cpf == strings.Repeat(string(cpf[0]), 11) {
		return false
	}

	// Validate first check digit
	sum := 0
	for i := 0; i < 9; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (10 - i)
	}

	remainder := sum % 11
	firstDigit := 0
	if remainder >= 2 {
		firstDigit = 11 - remainder
	}

	expectedFirst, _ := strconv.Atoi(string(cpf[9]))
	if firstDigit != expectedFirst {
		return false
	}

	// Validate second check digit
	sum = 0
	for i := 0; i < 10; i++ {
		digit, _ := strconv.Atoi(string(cpf[i]))
		sum += digit * (11 - i)
	}

	remainder = sum % 11
	secondDigit := 0
	if remainder >= 2 {
		secondDigit = 11 - remainder
	}

	expectedSecond, _ := strconv.Atoi(string(cpf[10]))
	return secondDigit == expectedSecond
}

// ValidateStateRegistration validates state registration (IE)
func ValidateStateRegistration(ie, state string) bool {
	ie = OnlyNumbers(ie)

	if ie == "" || ie == "ISENTO" {
		return true
	}

	// Basic validation - each state has its own rules
	// This is a simplified version
	switch state {
	case "SP":
		return len(ie) == 12
	case "RJ":
		return len(ie) == 8
	case "MG":
		return len(ie) == 13
	default:
		return len(ie) >= 8 && len(ie) <= 14
	}
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ValidateZipCode validates Brazilian ZIP code (CEP)
func ValidateZipCode(cep string) bool {
	cep = OnlyNumbers(cep)
	return len(cep) == 8
}

// ValidatePhone validates Brazilian phone number
func ValidatePhone(phone string) bool {
	phone = OnlyNumbers(phone)
	return len(phone) >= 10 && len(phone) <= 11
}

// CleanXMLString cleans string for XML usage
func CleanXMLString(str string) string {
	// Remove control characters except tab, newline, and carriage return
	cleanStr := ""
	for _, r := range str {
		if r == '\t' || r == '\n' || r == '\r' || r >= 32 {
			cleanStr += string(r)
		}
	}

	// Replace XML special characters
	cleanStr = strings.ReplaceAll(cleanStr, "&", "&amp;")
	cleanStr = strings.ReplaceAll(cleanStr, "<", "&lt;")
	cleanStr = strings.ReplaceAll(cleanStr, ">", "&gt;")
	cleanStr = strings.ReplaceAll(cleanStr, "\"", "&quot;")
	cleanStr = strings.ReplaceAll(cleanStr, "'", "&apos;")

	return cleanStr
}

// NormalizeString normalizes string for NFe usage
func NormalizeString(str string, removeAccents bool, maxLength int) string {
	// Trim spaces
	str = strings.TrimSpace(str)

	// Remove accents if requested
	if removeAccents {
		str = RemoveAccents(str)
	}

	// Convert to uppercase
	str = strings.ToUpper(str)

	// Remove extra spaces
	str = regexp.MustCompile(`\s+`).ReplaceAllString(str, " ")

	// Truncate if necessary
	if maxLength > 0 {
		str = TruncateString(str, maxLength)
	}

	return str
}

// Round rounds float to specified decimal places
func Round(value float64, decimals int) float64 {
	multiplier := 1.0
	for i := 0; i < decimals; i++ {
		multiplier *= 10
	}

	return float64(int(value*multiplier+0.5)) / multiplier
}

// RoundCurrency rounds value to 2 decimal places
func RoundCurrency(value float64) float64 {
	return Round(value, 2)
}

// RoundQuantity rounds value to 4 decimal places
func RoundQuantity(value float64) float64 {
	return Round(value, 4)
}

// IsValidGTIN validates GTIN/EAN codes
func IsValidGTIN(gtin string) bool {
	if gtin == "" || gtin == "SEM GTIN" {
		return true
	}

	gtin = OnlyNumbers(gtin)

	// GTIN can be 8, 12, 13, or 14 digits
	validLengths := []int{8, 12, 13, 14}
	valid := false
	for _, length := range validLengths {
		if len(gtin) == length {
			valid = true
			break
		}
	}

	if !valid {
		return false
	}

	// Validate check digit using modulo 10
	if len(gtin) < 2 {
		return false
	}

	checkDigit := CalculateModulo10(gtin[:len(gtin)-1])
	expectedDigit, _ := strconv.Atoi(string(gtin[len(gtin)-1]))

	return checkDigit == expectedDigit
}

// FormatDecimal formats decimal value for NFe
func FormatDecimal(value float64, decimals int) string {
	rounded := Round(value, decimals)
	return FormatValue(rounded, decimals)
}

// ZeroPad pads number with leading zeros
func ZeroPad(number int, length int) string {
	return fmt.Sprintf("%0*d", length, number)
}

// IsEmpty checks if string is empty or contains only whitespace
func IsEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}

// CoalesceString returns first non-empty string
func CoalesceString(values ...string) string {
	for _, value := range values {
		if !IsEmpty(value) {
			return value
		}
	}
	return ""
}

// SanitizeXMLValue sanitizes value for XML
func SanitizeXMLValue(value string) string {
	if IsEmpty(value) {
		return ""
	}

	// Clean and normalize
	value = CleanXMLString(value)
	value = strings.TrimSpace(value)

	return value
}
