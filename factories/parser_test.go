package factories

import (
	"strings"
	"testing"
)

func TestNewParser(t *testing.T) {
	tests := []struct {
		name        string
		config      ParserConfig
		expectError bool
	}{
		{
			name: "default config",
			config: ParserConfig{
				Version: "4.00",
				Layout:  LayoutLocal,
			},
			expectError: false,
		},
		{
			name: "empty config uses defaults",
			config: ParserConfig{},
			expectError: false,
		},
		{
			name: "sebrae layout",
			config: ParserConfig{
				Version: "4.00",
				Layout:  LayoutSebrae,
			},
			expectError: false,
		},
		{
			name: "version 3.10",
			config: ParserConfig{
				Version: "3.10",
				Layout:  LayoutLocal,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := NewParser(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if parser == nil {
					t.Error("Parser should not be nil")
				}
				if parser.version == "" {
					t.Error("Version should not be empty")
				}
				if parser.layout == "" {
					t.Error("Layout should not be empty")
				}
			}
		})
	}
}

func TestParser_ValidateTXT(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	tests := []struct {
		name        string
		txtData     string
		expectError bool
	}{
		{
			name: "valid TXT",
			txtData: `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
`,
			expectError: false,
		},
		{
			name: "missing pipe at end",
			txtData: `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789|
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
`,
			expectError: true,
		},
		{
			name: "invalid characters",
			txtData: `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de "produtos"|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0||
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := parser.ValidateTXT(tt.txtData)
			hasErrors := len(errors) > 0
			
			if tt.expectError && !hasErrors {
				t.Error("Expected validation errors but got none")
			}
			if !tt.expectError && hasErrors {
				t.Errorf("Unexpected validation errors: %v", errors)
			}
		})
	}
}

func TestParser_parseFields(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	tests := []struct {
		name        string
		fields      []string
		structure   string
		expected    map[string]interface{}
		expectError bool
	}{
		{
			name:      "valid A tag",
			fields:    []string{"A", "4.00", "NFe41230714200166000187650010000000051123456789", "", ""},
			structure: "A|versao|Id|pk_nItem|",
			expected: map[string]interface{}{
				"versao": "4.00",
				"Id":     "NFe41230714200166000187650010000000051123456789",
			},
			expectError: false,
		},
		{
			name:        "field count mismatch",
			fields:      []string{"A", "4.00"},
			structure:   "A|versao|Id|pk_nItem|",
			expected:    nil,
			expectError: true,
		},
		{
			name:      "B tag with multiple fields",
			fields:    []string{"B", "41", "00000005", "Venda de produtos", "55", "001", "5", ""},
			structure: "B|cUF|cNF|natOp|mod|serie|nNF|",
			expected: map[string]interface{}{
				"cUF":   "41",
				"cNF":   "00000005",
				"natOp": "Venda de produtos",
				"mod":   "55",
				"serie": "001",
				"nNF":   "5",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseFields(tt.fields, tt.structure)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				for k, v := range tt.expected {
					if result[k] != v {
						t.Errorf("Expected %s=%v, got %v", k, v, result[k])
					}
				}
			}
		})
	}
}

func TestParser_ParseTXT(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	// Simple NFe TXT data
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
C|Empresa Teste LTDA|Empresa Teste|123456789|||||
C02|14200166000187|
C05|Rua Teste|123||Centro|4114902|Porto Alegre|RS|90000000|1058|Brasil||
E|Cliente Teste||123456789|||teste@email.com|
E02|12345678000195|
E05|Rua Cliente|456||Bairro|4314902|Cidade|RS|90001000|1058|Brasil||
I|1||
I02|PROD001||Produto Teste|12345678|999|5102|UN|1.0000|100.00|100.00||UN|1.0000|100.00|||||1||||
M|100.00|18.00||||||||100.00||||||||||100.00|||
`

	result, err := parser.ParseTXT(txtData)
	if err != nil {
		t.Fatalf("Unexpected parsing error: %v", err)
	}

	// Check that main sections exist
	if result["numNFe"] != 1 {
		t.Errorf("Expected numNFe=1, got %v", result["numNFe"])
	}

	if result["infNFe"] == nil {
		t.Error("Missing infNFe section")
	}

	if result["ide"] == nil {
		t.Error("Missing ide section")
	}

	if result["emit"] == nil {
		t.Error("Missing emit section")
	}

	if result["dest"] == nil {
		t.Error("Missing dest section")
	}

	if result["det"] == nil {
		t.Error("Missing det section")
	}

	if result["total"] == nil {
		t.Error("Missing total section")
	}

	// Check specific values
	if ide, ok := result["ide"].(map[string]interface{}); ok {
		if ide["cUF"] != "41" {
			t.Errorf("Expected cUF=41, got %v", ide["cUF"])
		}
		if ide["mod"] != "55" {
			t.Errorf("Expected mod=55, got %v", ide["mod"])
		}
	}

	if emit, ok := result["emit"].(map[string]interface{}); ok {
		if emit["xNome"] != "Empresa Teste LTDA" {
			t.Errorf("Expected xNome=Empresa Teste LTDA, got %v", emit["xNome"])
		}
		if emit["CNPJ"] != "14200166000187" {
			t.Errorf("Expected CNPJ=14200166000187, got %v", emit["CNPJ"])
		}
	}
}

func TestParser_GetXML(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	// Parse some data first
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
C|Empresa Teste LTDA|||||||
C02|14200166000187|
M|100.00|||||||||100.00||||||||||100.00||
`

	_, err := parser.ParseTXT(txtData)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	xml, err := parser.GetXML()
	if err != nil {
		t.Fatalf("XML generation failed: %v", err)
	}

	// Check that XML contains expected elements
	expectedElements := []string{
		"<?xml version=\"1.0\" encoding=\"UTF-8\"?>",
		"<NFe xmlns=\"http://www.portalfiscal.inf.br/nfe\">",
		"<infNFe>",
		"<ide>",
		"<emit>",
		"<total>",
		"</NFe>",
	}

	for _, element := range expectedElements {
		if !strings.Contains(xml, element) {
			t.Errorf("XML should contain %s", element)
		}
	}

	// Check for specific values
	if !strings.Contains(xml, "Empresa Teste LTDA") {
		t.Error("XML should contain company name")
	}
	if !strings.Contains(xml, "14200166000187") {
		t.Error("XML should contain CNPJ")
	}
}

func TestParser_storeTagData(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	// Test A tag (infNFe)
	parser.storeTagData("A", map[string]interface{}{
		"versao": "4.00",
		"Id":     "NFe41230714200166000187650010000000051123456789",
	})

	if parser.currentNFe["infNFe"] == nil {
		t.Error("infNFe should be stored")
	}

	// Test B tag (ide)
	parser.storeTagData("B", map[string]interface{}{
		"cUF": "41",
		"mod": "55",
	})

	if parser.currentNFe["ide"] == nil {
		t.Error("ide should be stored")
	}

	// Test C tag (emit)
	parser.storeTagData("C", map[string]interface{}{
		"xNome": "Empresa Teste",
	})
	parser.storeTagData("C02", map[string]interface{}{
		"CNPJ": "12345678000195",
	})

	if parser.currentNFe["emit"] == nil {
		t.Error("emit should be stored")
	}

	emit := parser.currentNFe["emit"].(map[string]interface{})
	if emit["xNome"] != "Empresa Teste" {
		t.Error("xNome should be stored in emit")
	}
	if emit["CNPJ"] != "12345678000195" {
		t.Error("CNPJ should be stored in emit")
	}

	// Test I tag (det - items)
	parser.storeTagData("I", map[string]interface{}{
		"nItem": "1",
	})
	parser.storeTagData("I02", map[string]interface{}{
		"cProd": "PROD001",
		"xProd": "Produto Teste",
	})

	if parser.currentNFe["det"] == nil {
		t.Error("det should be stored")
	}

	det := parser.currentNFe["det"].([]map[string]interface{})
	if len(det) != 1 {
		t.Errorf("Expected 1 item, got %d", len(det))
	}
	if det[0]["cProd"] != "PROD001" {
		t.Error("cProd should be stored in det")
	}
}

func TestConvertTXTToXML(t *testing.T) {
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
C|Empresa Teste LTDA|||||||
M|100.00|||||||||100.00||||||||||100.00||
`

	xml, err := ConvertTXTToXML(txtData, "4.00", LayoutLocal)
	if err != nil {
		t.Fatalf("Conversion failed: %v", err)
	}

	if !strings.Contains(xml, "<NFe") {
		t.Error("XML should contain NFe element")
	}
	if !strings.Contains(xml, "Empresa Teste LTDA") {
		t.Error("XML should contain company name")
	}
}

func TestParseNFeTXT(t *testing.T) {
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
`

	result, err := ParseNFeTXT(txtData, "4.00", LayoutLocal)
	if err != nil {
		t.Fatalf("Parsing failed: %v", err)
	}

	if result["numNFe"] != 1 {
		t.Errorf("Expected numNFe=1, got %v", result["numNFe"])
	}

	if result["infNFe"] == nil {
		t.Error("Missing infNFe section")
	}

	if result["ide"] == nil {
		t.Error("Missing ide section")
	}
}

// Test error handling
func TestParser_ErrorHandling(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	// Test invalid TXT data
	invalidTxtData := `INVALID|DATA|
UNKNOWN|TAG|
`

	_, err := parser.ParseTXT(invalidTxtData)
	if err == nil {
		t.Error("Expected error for invalid TXT data")
	}

	errors := parser.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected parsing errors")
	}
}

// Test empty data
func TestParser_EmptyData(t *testing.T) {
	parser, _ := NewParser(ParserConfig{})

	// Test GetXML with no data
	_, err := parser.GetXML()
	if err == nil {
		t.Error("Expected error when generating XML with no data")
	}

	// Test empty TXT
	result, err := parser.ParseTXT("")
	if err != nil {
		t.Errorf("Empty TXT should not cause error: %v", err)
	}

	if len(result) != 0 {
		t.Error("Empty TXT should result in empty data")
	}
}

// Benchmark tests
func BenchmarkParser_ParseTXT(b *testing.B) {
	parser, _ := NewParser(ParserConfig{})
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
C|Empresa Teste LTDA|||||||
C02|14200166000187|
E|Cliente Teste||||||| 
E02|12345678000195|
I|1||
I02|PROD001||Produto Teste|12345678|999|5102|UN|1.0000|100.00|100.00||UN|1.0000|100.00|||||1||||
M|100.00|18.00||||||||100.00||||||||||100.00|||
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.ParseTXT(txtData)
	}
}

func BenchmarkConvertTXTToXML(b *testing.B) {
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|2023-12-25T15:30:00-03:00||1|1|4114902|1|1|5|2|1|1|0|Aplicacao Teste|1.0|||
C|Empresa Teste LTDA|||||||
M|100.00|||||||||100.00||||||||||100.00||
`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ConvertTXTToXML(txtData, "4.00", LayoutLocal)
	}
}