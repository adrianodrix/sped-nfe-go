// Package utils provides key generation and validation utilities for NFe access keys.
// This implements the exact same algorithm used by the original PHP sped-nfe project.
package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/adrianodrix/sped-nfe-go/types"
)

// NFEKeyComponents represents the components needed to build an NFe access key
type NFEKeyComponents struct {
	UF       types.UF     // Código da UF do emitente
	DateTime time.Time    // Data/hora de emissão
	CNPJ     string       // CNPJ ou CPF do emitente (será preenchido com zeros)
	Model    types.ModeloNFe // Modelo do documento (55=NFe, 65=NFCe)
	Series   int          // Série do documento
	Number   int          // Número da nota fiscal
	EmitType types.TipoEmissao // Tipo de emissão
	Code     *int         // Código numérico (opcional, será gerado se nil)
}

// GenerateAccessKey generates a complete 44-digit NFe access key
// This function replicates the exact logic from Keys::build() in the PHP project
func GenerateAccessKey(components NFEKeyComponents) (string, error) {
	// Validate components
	if !components.UF.IsValid() {
		return "", errors.NewValidationError("invalid UF", "uf", components.UF)
	}
	
	if !isValidModel(components.Model) {
		return "", errors.NewValidationError("invalid model", "model", components.Model)
	}
	
	if components.Series < 0 || components.Series > 999 {
		return "", errors.NewValidationError("series must be between 0 and 999", "series", components.Series)
	}
	
	if components.Number < 1 || components.Number > 999999999 {
		return "", errors.NewValidationError("number must be between 1 and 999999999", "number", components.Number)
	}
	
	// Validate and clean CNPJ/CPF
	cnpjClean := CleanDocument(components.CNPJ)
	if len(cnpjClean) != 11 && len(cnpjClean) != 14 {
		return "", errors.NewValidationError("CNPJ/CPF must have 11 or 14 digits", "cnpj", components.CNPJ)
	}
	
	// Generate random code if not provided
	code := components.Code
	if code == nil {
		randomCode, err := generateRandomCode(components.Number)
		if err != nil {
			return "", errors.NewValidationError("failed to generate random code", "code", err)
		}
		code = &randomCode
	}
	
	// Validate code
	if *code < 0 || *code > 99999999 {
		return "", errors.NewValidationError("code must be between 0 and 99999999", "code", *code)
	}
	
	// Build the 43-digit key
	key43 := buildKey43(components, cnpjClean, *code)
	
	// Calculate check digit
	checkDigit := calculateCheckDigit(key43)
	
	// Return complete 44-digit key
	return key43 + checkDigit, nil
}

// ValidateAccessKey validates a complete 44-digit access key
func ValidateAccessKey(key string) error {
	// Remove any non-numeric characters
	keyClean := strings.ReplaceAll(key, " ", "")
	keyClean = strings.ReplaceAll(keyClean, "-", "")
	
	// Check length
	if len(keyClean) != types.ChaveAcessoLength {
		return errors.NewValidationError(
			fmt.Sprintf("access key must have exactly %d digits", types.ChaveAcessoLength),
			"key", key)
	}
	
	// Check if all characters are numeric
	for _, char := range keyClean {
		if char < '0' || char > '9' {
			return errors.NewValidationError("access key must contain only numeric characters", "key", key)
		}
	}
	
	// Extract components and validate
	components, err := ParseAccessKey(keyClean)
	if err != nil {
		return err
	}
	
	// Validate check digit
	expectedCheckDigit := calculateCheckDigit(keyClean[:43])
	actualCheckDigit := keyClean[43:44]
	
	if expectedCheckDigit != actualCheckDigit {
		return errors.NewValidationError("invalid check digit", "key", key)
	}
	
	// Validate UF
	if !components.UF.IsValid() {
		return errors.NewValidationError("invalid UF in access key", "uf", components.UF)
	}
	
	// Validate model
	if !isValidModel(components.Model) {
		return errors.NewValidationError("invalid model in access key", "model", components.Model)
	}
	
	return nil
}

// ParseAccessKey parses a 44-digit access key and extracts its components
func ParseAccessKey(key string) (*NFEKeyComponents, error) {
	if len(key) != 44 {
		return nil, errors.NewValidationError("access key must have 44 digits", "key", key)
	}
	
	// Extract UF (positions 1-2)
	ufCode, err := strconv.Atoi(key[0:2])
	if err != nil {
		return nil, errors.NewValidationError("invalid UF in access key", "uf", key[0:2])
	}
	uf := types.UF(ufCode)
	
	// Extract year and month (positions 3-6)
	yearStr := key[2:4]
	monthStr := key[4:6]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid year in access key", "year", yearStr)
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid month in access key", "month", monthStr)
	}
	
	// Convert 2-digit year to 4-digit year
	fullYear := 2000 + year
	if year > 50 { // Assume years > 50 are in 1900s
		fullYear = 1900 + year
	}
	
	dateTime := time.Date(fullYear, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	
	// Extract CNPJ (positions 7-20)
	cnpj := key[6:20]
	
	// Extract model (positions 21-22)
	modelCode, err := strconv.Atoi(key[20:22])
	if err != nil {
		return nil, errors.NewValidationError("invalid model in access key", "model", key[20:22])
	}
	model := types.ModeloNFe(modelCode)
	
	// Extract series (positions 23-25)
	series, err := strconv.Atoi(key[22:25])
	if err != nil {
		return nil, errors.NewValidationError("invalid series in access key", "series", key[22:25])
	}
	
	// Extract number (positions 26-34)
	number, err := strconv.Atoi(key[25:34])
	if err != nil {
		return nil, errors.NewValidationError("invalid number in access key", "number", key[25:34])
	}
	
	// Extract emission type (position 35)
	emitType, err := strconv.Atoi(key[34:35])
	if err != nil {
		return nil, errors.NewValidationError("invalid emission type in access key", "emitType", key[34:35])
	}
	
	// Extract code (positions 36-43)
	code, err := strconv.Atoi(key[35:43])
	if err != nil {
		return nil, errors.NewValidationError("invalid code in access key", "code", key[35:43])
	}
	
	return &NFEKeyComponents{
		UF:       uf,
		DateTime: dateTime,
		CNPJ:     cnpj,
		Model:    model,
		Series:   series,
		Number:   number,
		EmitType: types.TipoEmissao(emitType),
		Code:     &code,
	}, nil
}

// FormatAccessKey formats an access key with standard spacing for readability
func FormatAccessKey(key string) (string, error) {
	if err := ValidateAccessKey(key); err != nil {
		return "", err
	}
	
	clean := strings.ReplaceAll(key, " ", "")
	clean = strings.ReplaceAll(clean, "-", "")
	
	// Format as XXXX XXXX XXXX XXXX XXXX XXXX XXXX XXXX XXXX XXXX XXXX
	formatted := ""
	for i := 0; i < len(clean); i += 4 {
		end := i + 4
		if end > len(clean) {
			end = len(clean)
		}
		
		if i > 0 {
			formatted += " "
		}
		formatted += clean[i:end]
	}
	
	return formatted, nil
}

// Helper functions

// buildKey43 builds the 43-digit key (without check digit)
func buildKey43(components NFEKeyComponents, cnpj string, code int) string {
	// Pad CNPJ to 14 digits
	cnpjPadded := fmt.Sprintf("%014s", cnpj)
	
	// Format components according to the specification
	return fmt.Sprintf("%02d%02d%02d%s%02d%03d%09d%01d%08d",
		int(components.UF),                    // UF (2 digits)
		components.DateTime.Year()%100,        // Year (2 digits)
		int(components.DateTime.Month()),      // Month (2 digits)
		cnpjPadded,                           // CNPJ/CPF (14 digits)
		int(components.Model),                // Model (2 digits)
		components.Series,                    // Series (3 digits)
		components.Number,                    // Number (9 digits)
		int(components.EmitType),             // Emission type (1 digit)
		code,                                 // Random code (8 digits)
	)
}

// calculateCheckDigit calculates the check digit using the modulo 11 algorithm
// This replicates the exact logic from Keys::verifyingDigit() in the PHP project
func calculateCheckDigit(key43 string) string {
	if len(key43) != 43 {
		return ""
	}
	
	multipliers := []int{2, 3, 4, 5, 6, 7, 8, 9}
	weightedSum := 0
	position := 42 // Start from the rightmost position (0-indexed)
	
	for position >= 0 {
		for multiplierIndex := 0; multiplierIndex < 8 && position >= 0; multiplierIndex++ {
			digit := int(key43[position] - '0')
			weightedSum += digit * multipliers[multiplierIndex]
			position--
		}
	}
	
	remainder := weightedSum % 11
	checkDigit := 11 - remainder
	if checkDigit > 9 {
		checkDigit = 0
	}
	
	return strconv.Itoa(checkDigit)
}

// generateRandomCode generates a random 8-digit code ensuring it's different from the NFe number
// This replicates the logic from Keys::random() in the PHP project
func generateRandomCode(nfeNumber int) (int, error) {
	maxAttempts := 100
	
	for attempts := 0; attempts < maxAttempts; attempts++ {
		// Generate random number between 10000000 and 99999999
		max := big.NewInt(90000000) // 99999999 - 10000000 + 1
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return 0, err
		}
		
		code := int(n.Int64()) + 10000000
		
		// Ensure code is different from NFe number (NT2019.001 requirement)
		if code != nfeNumber {
			return code, nil
		}
	}
	
	return 0, fmt.Errorf("failed to generate unique code after %d attempts", maxAttempts)
}

// isValidModel checks if the NFe model is valid
func isValidModel(model types.ModeloNFe) bool {
	return model == types.ModeloNFe55 || model == types.ModeloNFCe65
}

// GetKeyComponents is a convenience function to extract key components from a string key
func GetKeyComponents(key string) (*NFEKeyComponents, error) {
	if err := ValidateAccessKey(key); err != nil {
		return nil, err
	}
	
	return ParseAccessKey(key)
}