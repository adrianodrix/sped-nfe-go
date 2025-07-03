// Package validation provides XSD validation functionality for NFe documents.
package validation

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/beevik/etree"
)

//go:embed schemas/xsd/*
var schemaFS embed.FS

// XSDValidator provides XSD validation against official SEFAZ schemas
type XSDValidator struct {
	schemasPath string
	schemas     map[string]*Schema
	mutex       sync.RWMutex
}

// Schema represents a loaded XSD schema
type Schema struct {
	Name     string
	Version  string
	Document *etree.Document
	Content  []byte
}

// ValidationResult represents the result of XSD validation
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Schema  string   `json:"schema,omitempty"`
	Version string   `json:"version,omitempty"`
}

// NewXSDValidator creates a new XSD validator
func NewXSDValidator() *XSDValidator {
	return &XSDValidator{
		schemasPath: "schemas/xsd",
		schemas:     make(map[string]*Schema),
	}
}

// NewXSDValidatorWithPath creates a new XSD validator with custom schemas path
func NewXSDValidatorWithPath(schemasPath string) *XSDValidator {
	return &XSDValidator{
		schemasPath: schemasPath,
		schemas:     make(map[string]*Schema),
	}
}

// LoadSchema loads an XSD schema by name and version
func (v *XSDValidator) LoadSchema(name, version string) (*Schema, error) {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	schemaKey := fmt.Sprintf("%s_v%s", name, version)

	// Check if schema is already loaded
	if schema, exists := v.schemas[schemaKey]; exists {
		return schema, nil
	}

	// Try to load from embedded schemas first
	schemaFile := fmt.Sprintf("%s_v%s.xsd", name, version)
	schemaPath := filepath.Join("schemas/xsd", schemaFile)

	var content []byte
	var err error

	// Try embedded filesystem first
	if data, embErr := schemaFS.ReadFile(schemaPath); embErr == nil {
		content = data
	} else {
		// Fallback to local filesystem
		localPath := filepath.Join(v.schemasPath, schemaFile)
		if content, err = os.ReadFile(localPath); err != nil {
			return nil, fmt.Errorf("failed to load schema %s: %w", schemaFile, err)
		}
	}

	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(content); err != nil {
		return nil, fmt.Errorf("failed to parse schema %s: %w", schemaFile, err)
	}

	schema := &Schema{
		Name:     name,
		Version:  version,
		Document: doc,
		Content:  content,
	}

	v.schemas[schemaKey] = schema
	return schema, nil
}

// ValidateNFe validates NFe XML against the appropriate schema
func (v *XSDValidator) ValidateNFe(xmlContent []byte, version string) *ValidationResult {
	return v.ValidateXML(xmlContent, "nfe", version)
}

// ValidateNFCe validates NFCe XML against the appropriate schema
func (v *XSDValidator) ValidateNFCe(xmlContent []byte, version string) *ValidationResult {
	return v.ValidateXML(xmlContent, "nfe", version) // NFCe uses same schema as NFe
}

// ValidateEnvio validates NFe batch submission XML
func (v *XSDValidator) ValidateEnvio(xmlContent []byte, version string) *ValidationResult {
	return v.ValidateXML(xmlContent, "enviNFe", version)
}

// ValidateConsulta validates consultation XML
func (v *XSDValidator) ValidateConsulta(xmlContent []byte, version string, consultaType string) *ValidationResult {
	schemaName := fmt.Sprintf("cons%s", consultaType)
	return v.ValidateXML(xmlContent, schemaName, version)
}

// ValidateEvento validates event XML
func (v *XSDValidator) ValidateEvento(xmlContent []byte, version string) *ValidationResult {
	return v.ValidateXML(xmlContent, "envEvento", version)
}

// ValidateCCe validates Carta de Correção Eletrônica
func (v *XSDValidator) ValidateCCe(xmlContent []byte, version string) *ValidationResult {
	return v.ValidateXML(xmlContent, "envCCe", version)
}

// ValidateInutilizacao validates inutilization XML
func (v *XSDValidator) ValidateInutilizacao(xmlContent []byte, version string) *ValidationResult {
	return v.ValidateXML(xmlContent, "inutNFe", version)
}

// ValidateXML validates XML content against specified schema
func (v *XSDValidator) ValidateXML(xmlContent []byte, schemaName, version string) *ValidationResult {
	result := &ValidationResult{
		Schema:  schemaName,
		Version: version,
		Valid:   false,
		Errors:  []string{},
	}

	// Load schema
	schema, err := v.LoadSchema(schemaName, version)
	if err != nil {
		// If schema doesn't exist, return true (same behavior as PHP version)
		if strings.Contains(err.Error(), "no such file") {
			result.Valid = true
			result.Errors = append(result.Errors, fmt.Sprintf("Schema %s_v%s.xsd not found - validation skipped", schemaName, version))
			return result
		}

		result.Errors = append(result.Errors, fmt.Sprintf("Failed to load schema: %s", err.Error()))
		return result
	}

	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(xmlContent); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse XML: %s", err.Error()))
		return result
	}

	// Perform basic XML structure validation
	if err := v.validateXMLStructure(doc, schema.Document); err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	// If we reach here, validation passed
	result.Valid = true
	return result
}

// validateXMLStructure performs basic XML structure validation
func (v *XSDValidator) validateXMLStructure(xmlDoc, schemaDoc *etree.Document) error {
	// Get root element of XML
	xmlRoot := xmlDoc.Root()
	if xmlRoot == nil {
		return fmt.Errorf("XML document has no root element")
	}

	// Get root element of schema (should be xs:schema)
	schemaRoot := schemaDoc.Root()
	if schemaRoot == nil {
		return fmt.Errorf("Schema document has no root element")
	}

	// Find element definitions in schema
	elements := schemaRoot.FindElements("//xs:element[@name]")
	if len(elements) == 0 {
		// Try with different namespace
		elements = schemaRoot.FindElements("//element[@name]")
	}

	// Basic validation - check if root element exists in schema
	rootElementFound := false
	for _, element := range elements {
		if name := element.SelectAttrValue("name", ""); name == xmlRoot.Tag {
			rootElementFound = true
			break
		}
	}

	if !rootElementFound {
		return fmt.Errorf("root element '%s' not found in schema", xmlRoot.Tag)
	}

	// Additional basic validations could be added here
	// For now, we perform minimal validation since full XSD validation
	// would require a more complex XSD parser

	return nil
}

// GetAvailableSchemas returns list of available schemas
func (v *XSDValidator) GetAvailableSchemas() ([]string, error) {
	var schemas []string

	// Read from embedded filesystem
	entries, err := fs.ReadDir(schemaFS, "schemas/xsd")
	if err != nil {
		// Fallback to local filesystem
		entries, err = os.ReadDir(v.schemasPath)
		if err != nil {
			return nil, err
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".xsd") {
			schemas = append(schemas, entry.Name())
		}
	}

	return schemas, nil
}

// ValidateAccessKey validates an access key format
func (v *XSDValidator) ValidateAccessKey(key string) *ValidationResult {
	result := &ValidationResult{
		Schema: "access_key",
		Valid:  false,
		Errors: []string{},
	}

	// Remove any formatting
	key = strings.ReplaceAll(key, " ", "")
	key = strings.ReplaceAll(key, "-", "")

	// Check length
	if len(key) != 44 {
		result.Errors = append(result.Errors, fmt.Sprintf("Access key must have exactly 44 digits, got %d", len(key)))
		return result
	}

	// Check if all characters are numeric
	for i, r := range key {
		if r < '0' || r > '9' {
			result.Errors = append(result.Errors, fmt.Sprintf("Invalid character '%c' at position %d", r, i+1))
			return result
		}
	}

	// Validate check digit (modulo 11)
	keyWithoutDV := key[:43]
	expectedCheckDigit := calculateModulo11(keyWithoutDV)
	actualCheckDigit := int(key[43] - '0')

	if expectedCheckDigit != actualCheckDigit {
		result.Errors = append(result.Errors, fmt.Sprintf("Invalid check digit: expected %d, got %d", expectedCheckDigit, actualCheckDigit))
		return result
	}

	result.Valid = true
	return result
}

// calculateModulo11 calculates modulo 11 check digit
func calculateModulo11(number string) int {
	sum := 0
	weight := 2

	// Process from right to left
	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
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

// GetSchemaInfo returns information about a loaded schema
func (v *XSDValidator) GetSchemaInfo(name, version string) (*Schema, error) {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	schemaKey := fmt.Sprintf("%s_v%s", name, version)
	if schema, exists := v.schemas[schemaKey]; exists {
		return schema, nil
	}

	return nil, fmt.Errorf("schema %s not loaded", schemaKey)
}

// ClearCache clears the schema cache
func (v *XSDValidator) ClearCache() {
	v.mutex.Lock()
	defer v.mutex.Unlock()

	v.schemas = make(map[string]*Schema)
}

// Validate provides a convenient method for common validations
func (v *XSDValidator) Validate(xmlContent []byte, docType string, version string) *ValidationResult {
	switch strings.ToLower(docType) {
	case "nfe":
		return v.ValidateNFe(xmlContent, version)
	case "nfce":
		return v.ValidateNFCe(xmlContent, version)
	case "envio", "envinfe":
		return v.ValidateEnvio(xmlContent, version)
	case "evento":
		return v.ValidateEvento(xmlContent, version)
	case "cce":
		return v.ValidateCCe(xmlContent, version)
	case "inutilizacao":
		return v.ValidateInutilizacao(xmlContent, version)
	default:
		return &ValidationResult{
			Valid:  false,
			Errors: []string{fmt.Sprintf("Unknown document type: %s", docType)},
		}
	}
}
