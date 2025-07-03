// Package converter provides TXT to XML conversion functionality for NFe documents.
package converter

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/adrianodrix/sped-nfe-go/nfe"
)

//go:embed layouts/*.json
var layoutFS embed.FS

// Layout types for different TXT formats
type Layout int

const (
	// Layout400Local represents the standard 4.00 layout
	Layout400Local Layout = iota
	// Layout400Sebrae represents the SEBRAE customized 4.00 layout
	Layout400Sebrae
	// Layout310Local represents the legacy 3.10 layout
	Layout310Local
)

// LayoutConfig holds the structure definition for a specific layout
type LayoutConfig struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Structure map[string]string `json:"structure"`
}

// Converter handles conversion from TXT format to NFe XML
type Converter struct {
	layout       Layout
	layoutConfig *LayoutConfig
	parser       *Parser
	validator    *Validator
}

// ConversionResult represents the result of a TXT to XML conversion
type ConversionResult struct {
	XMLs     [][]byte `json:"xmls"`
	Count    int      `json:"count"`
	Warnings []string `json:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty"`
}

// NewConverter creates a new converter instance with the specified layout
func NewConverter(layout Layout) (*Converter, error) {
	converter := &Converter{
		layout: layout,
	}

	// Load layout configuration
	if err := converter.loadLayoutConfig(); err != nil {
		return nil, fmt.Errorf("failed to load layout config: %w", err)
	}

	// Initialize parser and validator
	converter.parser = NewParser(converter.layoutConfig)
	converter.validator = NewValidator(converter.layoutConfig)

	return converter, nil
}

// NewConverterWithConfig creates a converter with custom layout configuration
func NewConverterWithConfig(config *LayoutConfig) *Converter {
	return &Converter{
		layout:       Layout400Local, // Default
		layoutConfig: config,
		parser:       NewParser(config),
		validator:    NewValidator(config),
	}
}

// loadLayoutConfig loads the layout configuration from embedded files
func (c *Converter) loadLayoutConfig() error {
	var filename string

	switch c.layout {
	case Layout400Local:
		filename = "layouts/txtstructure400.json"
	case Layout400Sebrae:
		filename = "layouts/txtstructure400_sebrae.json"
	case Layout310Local:
		filename = "layouts/txtstructure310.json"
	default:
		return fmt.Errorf("unsupported layout: %d", c.layout)
	}

	data, err := layoutFS.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read layout file %s: %w", filename, err)
	}

	// Parse as simple map first, then convert to LayoutConfig
	var structure map[string]string
	if err := json.Unmarshal(data, &structure); err != nil {
		return fmt.Errorf("failed to parse layout JSON: %w", err)
	}

	c.layoutConfig = &LayoutConfig{
		Name:      c.getLayoutName(),
		Version:   c.getLayoutVersion(),
		Structure: structure,
	}

	return nil
}

// getLayoutName returns the human-readable name for the layout
func (c *Converter) getLayoutName() string {
	switch c.layout {
	case Layout400Local:
		return "NFe 4.00 Local"
	case Layout400Sebrae:
		return "NFe 4.00 SEBRAE"
	case Layout310Local:
		return "NFe 3.10 Local"
	default:
		return "Unknown Layout"
	}
}

// getLayoutVersion returns the version string for the layout
func (c *Converter) getLayoutVersion() string {
	switch c.layout {
	case Layout310Local:
		return "3.10"
	default:
		return "4.00"
	}
}

// ConvertTXT converts TXT content to NFe XML
func (c *Converter) ConvertTXT(txtContent []byte) (*ConversionResult, error) {
	result := &ConversionResult{
		XMLs:     [][]byte{},
		Warnings: []string{},
		Errors:   []string{},
	}

	// Parse TXT content
	nfes, err := c.parseTXTContent(txtContent)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Parse error: %s", err.Error()))
		return result, err
	}

	// Convert each NFe to XML
	for i, nfeData := range nfes {
		xml, err := c.convertNFeToXML(nfeData, i)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("NFe %d conversion error: %s", i+1, err.Error()))
			continue
		}

		result.XMLs = append(result.XMLs, xml)
	}

	result.Count = len(result.XMLs)

	if len(result.Errors) > 0 && len(result.XMLs) == 0 {
		return result, fmt.Errorf("all conversions failed")
	}

	return result, nil
}

// parseTXTContent parses the TXT content and separates individual NFes
func (c *Converter) parseTXTContent(txtContent []byte) ([][]string, error) {
	lines := c.splitLines(txtContent)

	// Validate header
	if len(lines) == 0 {
		return nil, fmt.Errorf("empty TXT content")
	}

	if !strings.HasPrefix(lines[0], "NOTAFISCAL|") {
		return nil, fmt.Errorf("invalid TXT format: missing NOTAFISCAL header")
	}

	// Extract NFe count from header
	headerParts := strings.Split(lines[0], "|")
	if len(headerParts) < 2 {
		return nil, fmt.Errorf("invalid NOTAFISCAL header format")
	}

	expectedCount, err := strconv.Atoi(headerParts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid NFe count in header: %s", headerParts[1])
	}

	// Split into individual NFes
	nfes := c.splitNFes(lines[1:]) // Skip header

	if len(nfes) != expectedCount {
		return nil, fmt.Errorf("NFe count mismatch: expected %d, found %d", expectedCount, len(nfes))
	}

	return nfes, nil
}

// splitLines splits content into lines and cleans them
func (c *Converter) splitLines(content []byte) []string {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(content))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	return lines
}

// splitNFes separates multiple NFes based on the 'A|' tag
func (c *Converter) splitNFes(lines []string) [][]string {
	var nfes [][]string
	var currentNFe []string

	for _, line := range lines {
		if strings.HasPrefix(line, "A|") && len(currentNFe) > 0 {
			// Start of new NFe, save previous one
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

// convertNFeToXML converts a single NFe TXT data to XML
func (c *Converter) convertNFeToXML(nfeLines []string, index int) ([]byte, error) {
	// Validate NFe structure
	if err := c.validator.ValidateNFe(nfeLines); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse NFe data
	nfeData, err := c.parser.ParseNFe(nfeLines)
	if err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Generate XML using NFe package
	make := nfe.NewMake()

	// Build NFe XML from parsed data
	if err := c.buildNFeXML(make, nfeData); err != nil {
		return nil, fmt.Errorf("XML generation failed: %w", err)
	}

	// Get generated XML
	xml, err := make.GetXML()
	if err != nil {
		return nil, fmt.Errorf("failed to get XML: %w", err)
	}

	return []byte(xml), nil
}

// buildNFeXML builds the NFe XML using the Make instance
func (c *Converter) buildNFeXML(make *nfe.Make, data *NFEData) error {
	// Process basic NFe information
	if data.InfNFe != nil {
		if err := make.TagInfNFe(data.InfNFe.ID, data.InfNFe.Versao); err != nil {
			return fmt.Errorf("failed to add infNFe: %w", err)
		}
	}

	// Process identification
	if data.Identificacao != nil {
		if err := make.TagIde(data.Identificacao); err != nil {
			return fmt.Errorf("failed to add ide: %w", err)
		}
	}

	// Process issuer
	if data.Emitente != nil {
		if err := make.TagEmit(data.Emitente); err != nil {
			return fmt.Errorf("failed to add emit: %w", err)
		}
	}

	// Process recipient
	if data.Destinatario != nil {
		if err := make.TagDest(data.Destinatario); err != nil {
			return fmt.Errorf("failed to add dest: %w", err)
		}
	}

	// Process items
	for _, item := range data.Itens {
		if err := make.TagDet(item); err != nil {
			return fmt.Errorf("failed to add item: %w", err)
		}
	}

	// Process totals - using auto calculation
	// The totals will be calculated automatically by the Make instance

	// Process transport
	if data.Transporte != nil {
		if err := make.TagTransp(data.Transporte); err != nil {
			return fmt.Errorf("failed to add transp: %w", err)
		}
	}

	// Process additional information
	if data.InfAdic != nil {
		if err := make.TagInfAdic(data.InfAdic); err != nil {
			return fmt.Errorf("failed to add infAdic: %w", err)
		}
	}

	return nil
}

// ConvertTXTToXML is a convenience function for simple conversions
func ConvertTXTToXML(txtContent []byte, layout Layout) ([][]byte, error) {
	converter, err := NewConverter(layout)
	if err != nil {
		return nil, err
	}

	result, err := converter.ConvertTXT(txtContent)
	if err != nil {
		return nil, err
	}

	return result.XMLs, nil
}

// ConvertTXTFile converts a TXT file to XML
func ConvertTXTFile(txtPath string, layout Layout) ([][]byte, error) {
	content, err := readFile(txtPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read TXT file: %w", err)
	}

	return ConvertTXTToXML(content, layout)
}

// readFile reads file content (helper function)
func readFile(path string) ([]byte, error) {
	// This would be implemented using os.ReadFile
	// For now, return an error to indicate it needs implementation
	return nil, fmt.Errorf("file reading not yet implemented")
}

// GetSupportedLayouts returns a list of supported layouts
func GetSupportedLayouts() []string {
	return []string{
		"Layout400Local - NFe 4.00 Standard",
		"Layout400Sebrae - NFe 4.00 SEBRAE",
		"Layout310Local - NFe 3.10 Legacy",
	}
}

// ValidateTXT validates TXT content without converting
func (c *Converter) ValidateTXT(txtContent []byte) ([]string, error) {
	nfes, err := c.parseTXTContent(txtContent)
	if err != nil {
		return nil, err
	}

	var allErrors []string
	for i, nfeLines := range nfes {
		if err := c.validator.ValidateNFe(nfeLines); err != nil {
			allErrors = append(allErrors, fmt.Sprintf("NFe %d: %s", i+1, err.Error()))
		}
	}

	return allErrors, nil
}
