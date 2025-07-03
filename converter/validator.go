// Package converter provides validation functionality for TXT files.
package converter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Validator handles validation of TXT content according to layout rules
type Validator struct {
	layoutConfig *LayoutConfig
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Line    int    `json:"line"`
	Tag     string `json:"tag"`
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if ve.Field != "" {
		return fmt.Sprintf("line %d, tag %s, field %s: %s", ve.Line, ve.Tag, ve.Field, ve.Message)
	}
	return fmt.Sprintf("line %d, tag %s: %s", ve.Line, ve.Tag, ve.Message)
}

// ValidationResult holds the result of validation
type ValidationResult struct {
	Valid    bool               `json:"valid"`
	Errors   []*ValidationError `json:"errors,omitempty"`
	Warnings []*ValidationError `json:"warnings,omitempty"`
}

// NewValidator creates a new validator instance
func NewValidator(config *LayoutConfig) *Validator {
	return &Validator{
		layoutConfig: config,
	}
}

// ValidateNFe validates an entire NFe TXT structure
func (v *Validator) ValidateNFe(lines []string) error {
	result := v.ValidateNFeDetailed(lines)

	if !result.Valid {
		// Convert validation errors to a single error message
		var messages []string
		for _, err := range result.Errors {
			messages = append(messages, err.Error())
		}
		return fmt.Errorf("validation errors: %s", strings.Join(messages, "; "))
	}

	return nil
}

// ValidateNFeDetailed performs detailed validation and returns full results
func (v *Validator) ValidateNFeDetailed(lines []string) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []*ValidationError{},
		Warnings: []*ValidationError{},
	}

	// Track required tags
	requiredTags := map[string]bool{
		"A": false, // infNFe
		"B": false, // identification
		"C": false, // issuer
		"I": false, // at least one item
	}

	// Validate each line
	for lineNum, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		lineErrors := v.validateLine(line, lineNum+1)
		result.Errors = append(result.Errors, lineErrors...)

		// Track required tags
		if len(lineErrors) == 0 {
			tag := v.extractTag(line)
			if _, exists := requiredTags[tag]; exists {
				requiredTags[tag] = true
			}
		}
	}

	// Check for missing required tags
	for tag, found := range requiredTags {
		if !found {
			result.Errors = append(result.Errors, &ValidationError{
				Line:    0,
				Tag:     tag,
				Message: fmt.Sprintf("required tag %s not found", tag),
			})
		}
	}

	// Validate tag sequence
	sequenceErrors := v.validateTagSequence(lines)
	result.Errors = append(result.Errors, sequenceErrors...)

	// Set overall validation result
	result.Valid = len(result.Errors) == 0

	return result
}

// validateLine validates a single TXT line
func (v *Validator) validateLine(line string, lineNum int) []*ValidationError {
	var errors []*ValidationError

	// Basic format validation
	if !strings.HasSuffix(line, "|") {
		errors = append(errors, &ValidationError{
			Line:    lineNum,
			Tag:     "",
			Message: "line must end with pipe character (|)",
		})
		return errors // Cannot continue validation
	}

	// Extract tag and validate
	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		errors = append(errors, &ValidationError{
			Line:    lineNum,
			Tag:     "",
			Message: "invalid line format: missing tag or fields",
		})
		return errors
	}

	tag := parts[0]

	// Validate tag exists in layout
	structure, exists := v.layoutConfig.Structure[tag]
	if !exists {
		errors = append(errors, &ValidationError{
			Line:    lineNum,
			Tag:     tag,
			Message: fmt.Sprintf("unknown tag: %s", tag),
		})
		return errors
	}

	// Validate field count
	structParts := strings.Split(structure, "|")
	expectedFields := len(structParts) - 2 // Subtract tag and trailing |
	actualFields := len(parts) - 1         // Subtract tag

	if actualFields != expectedFields {
		errors = append(errors, &ValidationError{
			Line:    lineNum,
			Tag:     tag,
			Message: fmt.Sprintf("field count mismatch: expected %d, got %d", expectedFields, actualFields),
		})
	}

	// Validate individual fields
	fieldErrors := v.validateFields(tag, parts, structParts, lineNum)
	errors = append(errors, fieldErrors...)

	return errors
}

// validateFields validates individual fields within a tag
func (v *Validator) validateFields(tag string, parts, structParts []string, lineNum int) []*ValidationError {
	var errors []*ValidationError

	for i := 1; i < len(structParts)-1 && i < len(parts); i++ {
		fieldName := structParts[i]
		fieldValue := parts[i]

		if fieldName == "" {
			continue // Skip unnamed fields
		}

		// Validate field content
		fieldErrors := v.validateFieldContent(tag, fieldName, fieldValue, lineNum)
		errors = append(errors, fieldErrors...)
	}

	return errors
}

// validateFieldContent validates the content of a specific field
func (v *Validator) validateFieldContent(tag, fieldName, value string, lineNum int) []*ValidationError {
	var errors []*ValidationError

	// Clean the value
	cleanValue := strings.TrimSpace(value)

	// Check for prohibited characters
	if v.hasProhibitedChars(cleanValue) {
		errors = append(errors, &ValidationError{
			Line:    lineNum,
			Tag:     tag,
			Field:   fieldName,
			Message: "contains prohibited characters",
		})
	}

	// Specific field validations
	switch fieldName {
	case "versao":
		if !v.isValidVersion(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid version: %s", cleanValue),
			})
		}

	case "CNPJ":
		if cleanValue != "" && !v.isValidCNPJ(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid CNPJ format: %s", cleanValue),
			})
		}

	case "CPF":
		if cleanValue != "" && !v.isValidCPF(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid CPF format: %s", cleanValue),
			})
		}

	case "cUF":
		if cleanValue != "" && !v.isValidUF(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid UF code: %s", cleanValue),
			})
		}

	case "mod":
		if cleanValue != "" && !v.isValidModel(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid model: %s (must be 55 or 65)", cleanValue),
			})
		}

	case "dhEmi", "dhSaiEnt":
		if cleanValue != "" && !v.isValidDateTime(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid datetime format: %s", cleanValue),
			})
		}

	case "IE":
		if cleanValue != "" && cleanValue != "ISENTO" && !v.isValidIE(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid IE format: %s", cleanValue),
			})
		}

	case "email":
		if cleanValue != "" && !v.isValidEmail(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid email format: %s", cleanValue),
			})
		}

	case "CEP":
		if cleanValue != "" && !v.isValidCEP(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid CEP format: %s", cleanValue),
			})
		}

	case "NCM":
		if cleanValue != "" && !v.isValidNCM(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid NCM format: %s", cleanValue),
			})
		}

	case "CFOP":
		if cleanValue != "" && !v.isValidCFOP(cleanValue) {
			errors = append(errors, &ValidationError{
				Line:    lineNum,
				Tag:     tag,
				Field:   fieldName,
				Message: fmt.Sprintf("invalid CFOP format: %s", cleanValue),
			})
		}
	}

	return errors
}

// validateTagSequence validates the logical sequence of tags
func (v *Validator) validateTagSequence(lines []string) []*ValidationError {
	var errors []*ValidationError

	hasA := false
	hasB := false
	hasC := false
	_ = false // inItem - placeholder for future use

	for lineNum, line := range lines {
		tag := v.extractTag(line)

		switch tag {
		case "A":
			hasA = true
		case "B":
			if !hasA {
				errors = append(errors, &ValidationError{
					Line:    lineNum + 1,
					Tag:     tag,
					Message: "tag B must come after tag A",
				})
			}
			hasB = true
		case "C":
			if !hasB {
				errors = append(errors, &ValidationError{
					Line:    lineNum + 1,
					Tag:     tag,
					Message: "tag C must come after tag B",
				})
			}
			hasC = true
		case "H":
			// inItem = true // commented out for now
		case "I":
			if !hasC {
				errors = append(errors, &ValidationError{
					Line:    lineNum + 1,
					Tag:     tag,
					Message: "tag I must come after issuer information (C)",
				})
			}
		}
	}

	return errors
}

// hasProhibitedChars checks for prohibited characters
func (v *Validator) hasProhibitedChars(value string) bool {
	// Check for prohibited characters: < > " ' \t \r and control characters
	prohibitedPattern := `[<>"'\t\r\x00-\x08\x10\x0B\x0C\x0E-\x19\x7F]`
	matched, _ := regexp.MatchString(prohibitedPattern, value)
	return matched
}

// isValidVersion validates version format
func (v *Validator) isValidVersion(version string) bool {
	validVersions := []string{"3.10", "4.00"}
	for _, valid := range validVersions {
		if version == valid {
			return true
		}
	}
	return false
}

// isValidCNPJ validates CNPJ format (basic check)
func (v *Validator) isValidCNPJ(cnpj string) bool {
	// Remove any formatting
	digits := regexp.MustCompile(`\D`).ReplaceAllString(cnpj, "")

	// Must have exactly 14 digits
	if len(digits) != 14 {
		return false
	}

	// Check for repeated digits (all same digit)
	if len(digits) == 14 {
		allSame := true
		for i := 1; i < 14; i++ {
			if digits[i] != digits[0] {
				allSame = false
				break
			}
		}
		if allSame {
			return false
		}
	}

	return true
}

// isValidCPF validates CPF format (basic check)
func (v *Validator) isValidCPF(cpf string) bool {
	// Remove any formatting
	digits := regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")

	// Must have exactly 11 digits
	if len(digits) != 11 {
		return false
	}

	// Check for repeated digits (all same digit)
	if len(digits) == 11 {
		allSame := true
		for i := 1; i < 11; i++ {
			if digits[i] != digits[0] {
				allSame = false
				break
			}
		}
		if allSame {
			return false
		}
	}

	return true
}

// isValidUF validates UF code
func (v *Validator) isValidUF(uf string) bool {
	// Remove spaces and validate as 2-digit code
	cleanUF := strings.TrimSpace(uf)
	if len(cleanUF) != 2 {
		return false
	}

	// Check if it's a valid numeric UF code
	if code, err := strconv.Atoi(cleanUF); err == nil {
		return code >= 11 && code <= 53
	}

	return false
}

// isValidModel validates document model
func (v *Validator) isValidModel(model string) bool {
	return model == "55" || model == "65"
}

// isValidDateTime validates datetime format
func (v *Validator) isValidDateTime(datetime string) bool {
	// Check for common NFe datetime formats
	patterns := []string{
		`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}$`, // ISO with timezone
		`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}$`,                // ISO without timezone
		`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`,                // Simple datetime
		`^\d{4}-\d{2}-\d{2}$`,                                  // Date only
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, datetime); matched {
			return true
		}
	}

	return false
}

// isValidIE validates state registration format (basic check)
func (v *Validator) isValidIE(ie string) bool {
	// Remove any formatting
	digits := regexp.MustCompile(`\D`).ReplaceAllString(ie, "")

	// Must have between 8 and 14 digits
	return len(digits) >= 8 && len(digits) <= 14
}

// isValidEmail validates email format
func (v *Validator) isValidEmail(email string) bool {
	// Basic email validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

// isValidCEP validates CEP format
func (v *Validator) isValidCEP(cep string) bool {
	// Remove any formatting
	digits := regexp.MustCompile(`\D`).ReplaceAllString(cep, "")

	// Must have exactly 8 digits
	return len(digits) == 8
}

// isValidNCM validates NCM format
func (v *Validator) isValidNCM(ncm string) bool {
	// Remove any formatting
	digits := regexp.MustCompile(`\D`).ReplaceAllString(ncm, "")

	// Must have exactly 8 digits
	return len(digits) == 8
}

// isValidCFOP validates CFOP format
func (v *Validator) isValidCFOP(cfop string) bool {
	// Must have exactly 4 digits
	if len(cfop) != 4 {
		return false
	}

	// Must be all digits
	matched, _ := regexp.MatchString(`^\d{4}$`, cfop)
	return matched
}

// extractTag extracts the tag from a TXT line
func (v *Validator) extractTag(line string) string {
	if idx := strings.Index(line, "|"); idx > 0 {
		return line[:idx]
	}
	return ""
}

// ValidateTXTContent validates entire TXT content (multiple NFes)
func (v *Validator) ValidateTXTContent(content []byte) (*ValidationResult, error) {
	lines := strings.Split(string(content), "\n")

	// Clean empty lines
	var cleanLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			cleanLines = append(cleanLines, strings.TrimSpace(line))
		}
	}

	if len(cleanLines) == 0 {
		return &ValidationResult{
			Valid:  false,
			Errors: []*ValidationError{{Line: 1, Message: "empty content"}},
		}, fmt.Errorf("empty content")
	}

	// Validate header
	if !strings.HasPrefix(cleanLines[0], "NOTAFISCAL|") {
		return &ValidationResult{
			Valid:  false,
			Errors: []*ValidationError{{Line: 1, Message: "missing NOTAFISCAL header"}},
		}, fmt.Errorf("missing NOTAFISCAL header")
	}

	// Extract and validate NFe count
	headerParts := strings.Split(cleanLines[0], "|")
	if len(headerParts) < 2 {
		return &ValidationResult{
			Valid:  false,
			Errors: []*ValidationError{{Line: 1, Message: "invalid NOTAFISCAL header format"}},
		}, fmt.Errorf("invalid header format")
	}

	// Split into NFes and validate each
	nfes := v.splitNFes(cleanLines[1:])

	result := &ValidationResult{
		Valid:    true,
		Errors:   []*ValidationError{},
		Warnings: []*ValidationError{},
	}

	for i, nfeLines := range nfes {
		nfeResult := v.ValidateNFeDetailed(nfeLines)
		if !nfeResult.Valid {
			// Add NFe index to error messages
			for _, err := range nfeResult.Errors {
				err.Message = fmt.Sprintf("NFe %d: %s", i+1, err.Message)
			}
			result.Errors = append(result.Errors, nfeResult.Errors...)
		}
		result.Warnings = append(result.Warnings, nfeResult.Warnings...)
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

// splitNFes splits content into individual NFes
func (v *Validator) splitNFes(lines []string) [][]string {
	var nfes [][]string
	var currentNFe []string

	for _, line := range lines {
		if strings.HasPrefix(line, "A|") && len(currentNFe) > 0 {
			// Start of new NFe
			nfes = append(nfes, currentNFe)
			currentNFe = []string{line}
		} else {
			currentNFe = append(currentNFe, line)
		}
	}

	// Add the last NFe
	if len(currentNFe) > 0 {
		nfes = append(nfes, currentNFe)
	}

	return nfes
}
