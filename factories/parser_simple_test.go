package factories

import (
	"strings"
	"testing"
)

func TestSimpleParser(t *testing.T) {
	parser, err := NewParser(ParserConfig{
		Version: "4.00",
		Layout:  LayoutLocal,
	})
	if err != nil {
		t.Fatalf("Failed to create parser: %v", err)
	}

	// Test with very simple data that matches the structure exactly
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
`

	result, err := parser.ParseTXT(txtData)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	// Check basic structure
	if result["numNFe"] != 1 {
		t.Errorf("Expected numNFe=1, got %v", result["numNFe"])
	}

	if result["infNFe"] == nil {
		t.Error("Missing infNFe section")
	}

	infNFe := result["infNFe"].(map[string]interface{})
	if infNFe["versao"] != "4.00" {
		t.Errorf("Expected versao=4.00, got %v", infNFe["versao"])
	}
}

func TestParserValidation(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	// Test validation
	validTxt := "NOTAFISCAL|1|\nA|4.00|test||\n"
	errors := parser.ValidateTXT(validTxt)
	if len(errors) > 0 {
		t.Errorf("Valid TXT should not have errors: %v", errors)
	}

	// Test invalid TXT (missing pipe at end)
	invalidTxt := "A|4.00|test"
	errors = parser.ValidateTXT(invalidTxt)
	if len(errors) == 0 {
		t.Error("Invalid TXT should have errors")
	}
}

func TestParserXMLGeneration(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	// Parse minimal data
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
`

	_, err := parser.ParseTXT(txtData)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	xml, err := parser.GetXML()
	if err != nil {
		t.Fatalf("XML generation failed: %v", err)
	}

	// Check basic XML structure
	if !strings.Contains(xml, "<NFe") {
		t.Error("XML should contain NFe element")
	}
	if !strings.Contains(xml, "<?xml") {
		t.Error("XML should contain XML declaration")
	}
	if !strings.Contains(xml, "</NFe>") {
		t.Error("XML should close NFe element")
	}
}

func TestConvenienceFunctions(t *testing.T) {
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
`

	// Test ParseNFeTXT
	result, err := ParseNFeTXT(txtData, "4.00", LayoutLocal)
	if err != nil {
		t.Fatalf("ParseNFeTXT failed: %v", err)
	}
	if result["numNFe"] != 1 {
		t.Error("ParseNFeTXT should parse correctly")
	}

	// Test ConvertTXTToXML
	xml, err := ConvertTXTToXML(txtData, "4.00", LayoutLocal)
	if err != nil {
		t.Fatalf("ConvertTXTToXML failed: %v", err)
	}
	if !strings.Contains(xml, "<NFe") {
		t.Error("ConvertTXTToXML should generate valid XML")
	}
}
