// Package factories provides utility factories for NFe processing including TXT to XML parsing.
package factories

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LayoutType represents the TXT layout type
type LayoutType string

const (
	LayoutLocal    LayoutType = "LOCAL"
	LayoutLocalV12 LayoutType = "LOCAL_V12"
	LayoutSebrae   LayoutType = "SEBRAE"
)

// Parser converts NFe TXT format to XML
type Parser struct {
	version    string
	layout     LayoutType
	structure  map[string]string
	currentNFe map[string]interface{}
	errors     []string
}

// ParserConfig holds configuration for TXT parsing
type ParserConfig struct {
	Version string     // NFe version (3.10, 4.00)
	Layout  LayoutType // Layout type
}

// NewParser creates a new TXT to XML parser
func NewParser(config ParserConfig) (*Parser, error) {
	if config.Version == "" {
		config.Version = "4.00"
	}
	if config.Layout == "" {
		config.Layout = LayoutLocal
	}

	parser := &Parser{
		version:    config.Version,
		layout:     config.Layout,
		currentNFe: make(map[string]interface{}),
		errors:     []string{},
	}

	if err := parser.loadStructure(); err != nil {
		return nil, fmt.Errorf("failed to load TXT structure: %v", err)
	}

	return parser, nil
}

// loadStructure loads the TXT structure from JSON file
func (p *Parser) loadStructure() error {
	// Determine file name based on version and layout
	ver := strings.Replace(p.version, ".", "", -1)
	suffix := ""

	switch p.layout {
	case LayoutSebrae:
		suffix = "_sebrae"
	case LayoutLocalV12:
		suffix = "_v1.2"
	}

	filename := fmt.Sprintf("txtstructure%s%s.json", ver, suffix)
	structPath := filepath.Join("refs", "sped-nfe", "storage", filename)

	// Try to load the structure file
	data, err := os.ReadFile(structPath)
	if err != nil {
		// Create a basic structure if file doesn't exist
		p.structure = p.createBasicStructure()
		return nil
	}

	if err := json.Unmarshal(data, &p.structure); err != nil {
		return fmt.Errorf("failed to parse structure JSON: %v", err)
	}

	return nil
}

// createBasicStructure creates a basic TXT structure for NFe 4.00
func (p *Parser) createBasicStructure() map[string]string {
	return map[string]string{
		"A":    "A|versao|Id|pk_nItem|",
		"B":    "B|cUF|cNF|natOp|mod|serie|nNF|dhEmi|dhSaiEnt|tpNF|idDest|cMunFG|tpImp|tpEmis|cDV|tpAmb|finNFe|indFinal|indPres|procEmi|verProc|dhCont|xJust|",
		"C":    "C|xNome|xFant|IE|IEST|IM|CNAE|CRT|",
		"C02":  "C02|CNPJ|",
		"C02a": "C02a|CPF|",
		"C05":  "C05|xLgr|nro|xCpl|xBairro|cMun|xMun|UF|CEP|cPais|xPais|fone|",
		"E":    "E|xNome|indIEDest|IE|ISUF|IM|email|",
		"E02":  "E02|CNPJ|",
		"E03":  "E03|CPF|",
		"E03a": "E03a|idEstrangeiro|",
		"E05":  "E05|xLgr|nro|xCpl|xBairro|cMun|xMun|UF|CEP|cPais|xPais|fone|",
		"I":    "I|nItem|infAdProd|",
		"I02":  "I02|cProd|cEAN|xProd|NCM|EXTIPI|CFOP|uCom|qCom|vUnCom|vProd|cEANTrib|uTrib|qTrib|vUnTrib|vFrete|vSeg|vDesc|vOutro|indTot|xPed|nItemPed|nFCI|",
		"M":    "M|vBC|vICMS|vICMSDeson|vFCPUFDest|vICMSUFDest|vICMSUFRemet|vFCP|vBCST|vST|vFCPST|vFCPSTRet|vProd|vFrete|vSeg|vDesc|vII|vIPI|vIPIDevol|vPIS|vCOFINS|vOutro|vNF|vTotTrib|",
		"N":    "N|orig|CST|modBC|vBC|pICMS|vICMS|pFCP|vFCP|",
		"W":    "W|vBC|vDespAdu|vII|vIOF|",
	}
}

// ParseTXT converts TXT data to structured data
func (p *Parser) ParseTXT(txtData string) (map[string]interface{}, error) {
	p.errors = []string{}
	p.currentNFe = make(map[string]interface{})

	lines := strings.Split(strings.ReplaceAll(txtData, "\r", ""), "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if err := p.parseLine(line, lineNum+1); err != nil {
			p.errors = append(p.errors, fmt.Sprintf("Line %d: %v", lineNum+1, err))
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parsing errors: %s", strings.Join(p.errors, "; "))
	}

	return p.currentNFe, nil
}

// parseLine parses a single line of TXT data
func (p *Parser) parseLine(line string, lineNum int) error {
	if !strings.HasSuffix(line, "|") {
		return fmt.Errorf("line must end with '|'")
	}

	fields := strings.Split(line, "|")
	if len(fields) < 2 {
		return fmt.Errorf("invalid line format")
	}

	tag := strings.ToUpper(fields[0])

	// Handle NOTAFISCAL header
	if tag == "NOTAFISCAL" {
		if len(fields) >= 2 {
			numNFe, err := strconv.Atoi(fields[1])
			if err != nil {
				return fmt.Errorf("invalid NFe count: %v", err)
			}
			p.currentNFe["numNFe"] = numNFe
		}
		return nil
	}

	structure, exists := p.structure[tag]
	if !exists {
		return fmt.Errorf("unknown tag: %s", tag)
	}

	// Parse the line according to structure
	data, err := p.parseFields(fields, structure)
	if err != nil {
		return fmt.Errorf("failed to parse tag %s: %v", tag, err)
	}

	// Store the parsed data
	p.storeTagData(tag, data)

	return nil
}

// parseFields parses fields according to structure definition
func (p *Parser) parseFields(fields []string, structure string) (map[string]interface{}, error) {
	structFields := strings.Split(structure, "|")
	data := make(map[string]interface{})

	// Check field count
	expectedCount := len(structFields) - 1 // Exclude last empty field after final |
	actualCount := len(fields) - 1         // Exclude last empty field after final |

	if actualCount != expectedCount {
		return nil, fmt.Errorf("field count mismatch: expected %d, got %d", expectedCount, actualCount)
	}

	// Map fields to structure
	for i := 1; i < len(structFields)-1; i++ {
		fieldName := structFields[i]
		if fieldName != "" && i < len(fields) && fields[i] != "" {
			data[fieldName] = fields[i]
		}
	}

	return data, nil
}

// storeTagData stores parsed tag data in the current NFe structure
func (p *Parser) storeTagData(tag string, data map[string]interface{}) {
	switch tag {
	case "A":
		p.currentNFe["infNFe"] = data
	case "B":
		p.currentNFe["ide"] = data
	case "C", "C02", "C02A", "C05":
		if p.currentNFe["emit"] == nil {
			p.currentNFe["emit"] = make(map[string]interface{})
		}
		emit := p.currentNFe["emit"].(map[string]interface{})
		for k, v := range data {
			emit[k] = v
		}
	case "E", "E02", "E03", "E03A", "E05":
		if p.currentNFe["dest"] == nil {
			p.currentNFe["dest"] = make(map[string]interface{})
		}
		dest := p.currentNFe["dest"].(map[string]interface{})
		for k, v := range data {
			dest[k] = v
		}
	case "I", "I02":
		if p.currentNFe["det"] == nil {
			p.currentNFe["det"] = []map[string]interface{}{}
		}
		det := p.currentNFe["det"].([]map[string]interface{})

		// Create new item or update existing
		itemIndex := len(det) - 1
		if tag == "I" {
			// Start new item
			newItem := make(map[string]interface{})
			for k, v := range data {
				newItem[k] = v
			}
			det = append(det, newItem)
		} else if itemIndex >= 0 {
			// Update current item
			for k, v := range data {
				det[itemIndex][k] = v
			}
		}
		p.currentNFe["det"] = det
	case "M":
		p.currentNFe["total"] = data
	case "N":
		// Handle ICMS data - add to current item
		if det, ok := p.currentNFe["det"].([]map[string]interface{}); ok && len(det) > 0 {
			itemIndex := len(det) - 1
			if det[itemIndex]["imposto"] == nil {
				det[itemIndex]["imposto"] = make(map[string]interface{})
			}
			imposto := det[itemIndex]["imposto"].(map[string]interface{})
			imposto["ICMS"] = data
		}
	case "W":
		// Handle II (Import Tax) data
		if det, ok := p.currentNFe["det"].([]map[string]interface{}); ok && len(det) > 0 {
			itemIndex := len(det) - 1
			if det[itemIndex]["imposto"] == nil {
				det[itemIndex]["imposto"] = make(map[string]interface{})
			}
			imposto := det[itemIndex]["imposto"].(map[string]interface{})
			imposto["II"] = data
		}
	default:
		// Store other tags as-is
		p.currentNFe[strings.ToLower(tag)] = data
	}
}

// GetXML converts the parsed data to XML (simplified version)
func (p *Parser) GetXML() (string, error) {
	if len(p.currentNFe) == 0 {
		return "", fmt.Errorf("no NFe data to convert")
	}

	// This is a simplified XML generation
	// In a real implementation, this would use the nfe.Make class
	xml := `<?xml version="1.0" encoding="UTF-8"?>`
	xml += `<NFe xmlns="http://www.portalfiscal.inf.br/nfe">`
	xml += `<infNFe>`

	// Add IDE section
	if ide, ok := p.currentNFe["ide"].(map[string]interface{}); ok {
		xml += `<ide>`
		for k, v := range ide {
			xml += fmt.Sprintf("<%s>%v</%s>", k, v, k)
		}
		xml += `</ide>`
	}

	// Add EMIT section
	if emit, ok := p.currentNFe["emit"].(map[string]interface{}); ok {
		xml += `<emit>`
		for k, v := range emit {
			xml += fmt.Sprintf("<%s>%v</%s>", k, v, k)
		}
		xml += `</emit>`
	}

	// Add DEST section
	if dest, ok := p.currentNFe["dest"].(map[string]interface{}); ok {
		xml += `<dest>`
		for k, v := range dest {
			xml += fmt.Sprintf("<%s>%v</%s>", k, v, k)
		}
		xml += `</dest>`
	}

	// Add DET sections
	if det, ok := p.currentNFe["det"].([]map[string]interface{}); ok {
		for i, item := range det {
			xml += fmt.Sprintf(`<det nItem="%d">`, i+1)
			for k, v := range item {
				if k != "imposto" {
					xml += fmt.Sprintf("<%s>%v</%s>", k, v, k)
				}
			}

			// Add tax information
			if imposto, ok := item["imposto"].(map[string]interface{}); ok {
				xml += `<imposto>`
				for taxType, taxData := range imposto {
					xml += fmt.Sprintf("<%s>", taxType)
					if taxMap, ok := taxData.(map[string]interface{}); ok {
						for k, v := range taxMap {
							xml += fmt.Sprintf("<%s>%v</%s>", k, v, k)
						}
					}
					xml += fmt.Sprintf("</%s>", taxType)
				}
				xml += `</imposto>`
			}
			xml += `</det>`
		}
	}

	// Add TOTAL section
	if total, ok := p.currentNFe["total"].(map[string]interface{}); ok {
		xml += `<total><ICMSTot>`
		for k, v := range total {
			xml += fmt.Sprintf("<%s>%v</%s>", k, v, k)
		}
		xml += `</ICMSTot></total>`
	}

	xml += `</infNFe>`
	xml += `</NFe>`

	return xml, nil
}

// GetErrors returns parsing errors
func (p *Parser) GetErrors() []string {
	return p.errors
}

// ValidateTXT validates TXT format before parsing
func (p *Parser) ValidateTXT(txtData string) []string {
	errors := []string{}
	lines := strings.Split(strings.ReplaceAll(txtData, "\r", ""), "\n")

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line ends with |
		if !strings.HasSuffix(line, "|") {
			errors = append(errors, fmt.Sprintf("Line %d: must end with '|'", lineNum+1))
			continue
		}

		// Check for invalid characters
		if strings.ContainsAny(line, "<>\"'") {
			errors = append(errors, fmt.Sprintf("Line %d: contains invalid characters", lineNum+1))
		}

		// Check field format
		fields := strings.Split(line, "|")
		if len(fields) < 2 {
			errors = append(errors, fmt.Sprintf("Line %d: invalid format", lineNum+1))
		}
	}

	return errors
}

// ParseNFeTXT is a convenience function to parse NFe TXT data
func ParseNFeTXT(txtData string, version string, layout LayoutType) (map[string]interface{}, error) {
	parser, err := NewParser(ParserConfig{
		Version: version,
		Layout:  layout,
	})
	if err != nil {
		return nil, err
	}

	return parser.ParseTXT(txtData)
}

// ConvertTXTToXML converts TXT data directly to XML
func ConvertTXTToXML(txtData string, version string, layout LayoutType) (string, error) {
	parser, err := NewParser(ParserConfig{
		Version: version,
		Layout:  layout,
	})
	if err != nil {
		return "", err
	}

	_, err = parser.ParseTXT(txtData)
	if err != nil {
		return "", err
	}

	return parser.GetXML()
}
