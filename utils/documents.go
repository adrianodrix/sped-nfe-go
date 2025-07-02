// Package utils provides Brazilian-specific utilities for document validation,
// formatting, and other operations specific to the Brazilian tax system.
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/adrianodrix/sped-nfe-go/types"
)

// CNPJ validation and formatting utilities

// ValidateCNPJ validates a Brazilian CNPJ (Cadastro Nacional da Pessoa Jurídica)
// with complete digit verification algorithm
func ValidateCNPJ(cnpj string) error {
	// Remove non-numeric characters
	cnpjClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(cnpj, "")
	
	// Check length
	if len(cnpjClean) != 14 {
		return errors.NewValidationError("CNPJ must have exactly 14 digits", "cnpj", cnpj)
	}
	
	// Check if all digits are the same
	if isAllSameDigits(cnpjClean) {
		return errors.NewValidationError("CNPJ cannot have all same digits", "cnpj", cnpj)
	}
	
	// Convert to int slice for calculation
	digits := make([]int, 14)
	for i, digit := range cnpjClean {
		digits[i] = int(digit - '0')
	}
	
	// Calculate first check digit
	firstCheckDigit := calculateCNPJCheckDigit(digits[:12], []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	if digits[12] != firstCheckDigit {
		return errors.NewValidationError("invalid CNPJ check digits", "cnpj", cnpj)
	}
	
	// Calculate second check digit
	secondCheckDigit := calculateCNPJCheckDigit(digits[:13], []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	if digits[13] != secondCheckDigit {
		return errors.NewValidationError("invalid CNPJ check digits", "cnpj", cnpj)
	}
	
	return nil
}

// ValidateCPF validates a Brazilian CPF (Cadastro de Pessoas Físicas)
// with complete digit verification algorithm
func ValidateCPF(cpf string) error {
	// Remove non-numeric characters
	cpfClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(cpf, "")
	
	// Check length
	if len(cpfClean) != 11 {
		return errors.NewValidationError("CPF must have exactly 11 digits", "cpf", cpf)
	}
	
	// Check if all digits are the same
	if isAllSameDigits(cpfClean) {
		return errors.NewValidationError("CPF cannot have all same digits", "cpf", cpf)
	}
	
	// Convert to int slice for calculation
	digits := make([]int, 11)
	for i, digit := range cpfClean {
		digits[i] = int(digit - '0')
	}
	
	// Calculate first check digit
	sum := 0
	for i := 0; i < 9; i++ {
		sum += digits[i] * (10 - i)
	}
	remainder := sum % 11
	firstCheckDigit := 0
	if remainder >= 2 {
		firstCheckDigit = 11 - remainder
	}
	
	if digits[9] != firstCheckDigit {
		return errors.NewValidationError("invalid CPF check digits", "cpf", cpf)
	}
	
	// Calculate second check digit
	sum = 0
	for i := 0; i < 10; i++ {
		sum += digits[i] * (11 - i)
	}
	remainder = sum % 11
	secondCheckDigit := 0
	if remainder >= 2 {
		secondCheckDigit = 11 - remainder
	}
	
	if digits[10] != secondCheckDigit {
		return errors.NewValidationError("invalid CPF check digits", "cpf", cpf)
	}
	
	return nil
}

// ValidateIE validates Inscrição Estadual (State Registration) by UF
// This is a simplified validation - full IE validation is very complex and UF-specific
func ValidateIE(ie string, uf types.UF) error {
	// Remove non-alphanumeric characters
	ieClean := regexp.MustCompile(`[^A-Za-z0-9]`).ReplaceAllString(ie, "")
	ieClean = strings.ToUpper(ieClean)
	
	// Check for exempt IE
	if ieClean == "ISENTO" {
		return nil
	}
	
	// Basic length validation by UF (simplified)
	expectedLengths := getIEExpectedLengths()
	expectedLength, exists := expectedLengths[uf]
	if !exists {
		return errors.NewValidationError("UF not supported for IE validation", "uf", uf)
	}
	
	if len(ieClean) != expectedLength {
		return errors.NewValidationError(
			fmt.Sprintf("IE for %s must have %d characters", uf.String(), expectedLength),
			"ie", ie)
	}
	
	// For now, just validate format and length
	// Full IE validation would require implementing specific algorithms for each UF
	if !regexp.MustCompile(`^[0-9A-Z]+$`).MatchString(ieClean) {
		return errors.NewValidationError("IE contains invalid characters", "ie", ie)
	}
	
	return nil
}

// FormatCNPJ formats a CNPJ string with standard Brazilian mask (XX.XXX.XXX/XXXX-XX)
func FormatCNPJ(cnpj string) (string, error) {
	// Validate first
	if err := ValidateCNPJ(cnpj); err != nil {
		return "", err
	}
	
	// Clean and format
	cnpjClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(cnpj, "")
	return fmt.Sprintf("%s.%s.%s/%s-%s",
		cnpjClean[0:2], cnpjClean[2:5], cnpjClean[5:8], cnpjClean[8:12], cnpjClean[12:14]), nil
}

// FormatCPF formats a CPF string with standard Brazilian mask (XXX.XXX.XXX-XX)
func FormatCPF(cpf string) (string, error) {
	// Validate first
	if err := ValidateCPF(cpf); err != nil {
		return "", err
	}
	
	// Clean and format
	cpfClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(cpf, "")
	return fmt.Sprintf("%s.%s.%s-%s",
		cpfClean[0:3], cpfClean[3:6], cpfClean[6:9], cpfClean[9:11]), nil
}

// CleanDocument removes all non-numeric characters from a document string
func CleanDocument(document string) string {
	return regexp.MustCompile(`[^0-9]`).ReplaceAllString(document, "")
}

// IsValidDocument validates either CPF or CNPJ automatically based on length
func IsValidDocument(document string) error {
	clean := CleanDocument(document)
	
	switch len(clean) {
	case 11:
		return ValidateCPF(document)
	case 14:
		return ValidateCNPJ(document)
	default:
		return errors.NewValidationError("document must be CPF (11 digits) or CNPJ (14 digits)", "document", document)
	}
}

// Helper functions

// calculateCNPJCheckDigit calculates CNPJ check digit using the provided weights
func calculateCNPJCheckDigit(digits []int, weights []int) int {
	sum := 0
	for i, digit := range digits {
		sum += digit * weights[i]
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}

// isAllSameDigits checks if all digits in the string are the same
func isAllSameDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := s[0]
	for _, char := range s {
		if byte(char) != first {
			return false
		}
	}
	return true
}

// getIEExpectedLengths returns the expected length for IE by UF
// This is a simplified mapping - real IE validation is much more complex
func getIEExpectedLengths() map[types.UF]int {
	return map[types.UF]int{
		types.AC: 13, // Acre
		types.AL: 9,  // Alagoas
		types.AP: 9,  // Amapá
		types.AM: 9,  // Amazonas
		types.BA: 9,  // Bahia - can be 8 or 9
		types.CE: 9,  // Ceará
		types.DF: 13, // Distrito Federal
		types.ES: 9,  // Espírito Santo
		types.GO: 9,  // Goiás
		types.MA: 9,  // Maranhão
		types.MT: 11, // Mato Grosso
		types.MS: 9,  // Mato Grosso do Sul
		types.MG: 13, // Minas Gerais
		types.PA: 9,  // Pará
		types.PB: 9,  // Paraíba
		types.PR: 10, // Paraná
		types.PE: 9,  // Pernambuco - can be 9 or 14
		types.PI: 9,  // Piauí
		types.RJ: 8,  // Rio de Janeiro
		types.RN: 10, // Rio Grande do Norte - can be 9 or 10
		types.RS: 10, // Rio Grande do Sul
		types.RO: 14, // Rondônia
		types.RR: 9,  // Roraima
		types.SC: 9,  // Santa Catarina
		types.SP: 12, // São Paulo
		types.SE: 9,  // Sergipe
		types.TO: 11, // Tocantins
	}
}