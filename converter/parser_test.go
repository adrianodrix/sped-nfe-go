package converter

import (
	"testing"

	"github.com/adrianodrix/sped-nfe-go/nfe"
)

func TestNewParser(t *testing.T) {
	config := &LayoutConfig{
		Name:    "Test Layout",
		Version: "4.00",
		Structure: map[string]string{
			"A": "A|versao|Id|pk_nItem|",
			"B": "B|cUF|cNF|natOp|mod|serie|nNF|dhEmi|dhSaiEnt|tpNF|idDest|cMunFG|tpImp|tpEmis|cDV|tpAmb|finNFe|indFinal|indPres|indIntermed|procEmi|verProc|dhCont|xJust|",
		},
	}

	parser := NewParser(config)
	
	if parser == nil {
		t.Error("Expected parser but got nil")
	}
	
	if parser.layoutConfig != config {
		t.Error("Parser config not set correctly")
	}
	
	if parser.currentItem != 0 {
		t.Errorf("Expected currentItem to be 0, got %d", parser.currentItem)
	}
}

func TestParser_parseFields(t *testing.T) {
	config := &LayoutConfig{
		Structure: map[string]string{
			"A": "A|versao|Id|pk_nItem|",
		},
	}
	
	parser := NewParser(config)

	tests := []struct {
		name      string
		parts     []string
		structure string
		expectErr bool
		expected  FieldMap
	}{
		{
			name:      "Valid fields",
			parts:     []string{"A", "4.00", "NFe123", "1"},
			structure: "A|versao|Id|pk_nItem|",
			expectErr: false,
			expected: FieldMap{
				"versao":   "4.00",
				"Id":       "NFe123",
				"pk_nItem": "1",
			},
		},
		{
			name:      "Field count mismatch",
			parts:     []string{"A", "4.00"},
			structure: "A|versao|Id|pk_nItem|",
			expectErr: true,
			expected:  nil,
		},
		{
			name:      "Empty fields",
			parts:     []string{"A", "", "NFe123", ""},
			structure: "A|versao|Id|pk_nItem|",
			expectErr: false,
			expected: FieldMap{
				"Id": "NFe123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseFields(tt.parts, tt.structure)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			for key, expectedValue := range tt.expected {
				if result[key] != expectedValue {
					t.Errorf("Field %s: expected %q, got %q", key, expectedValue, result[key])
				}
			}
		})
	}
}

func TestParser_processTagA(t *testing.T) {
	parser := NewParser(&LayoutConfig{})
	parser.currentNFe = &NFEData{}

	fields := FieldMap{
		"versao": "4.00",
		"Id":     "NFe35150271780456000160550010000000021800700082",
	}

	err := parser.processTagA(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if parser.currentNFe.InfNFe == nil {
		t.Error("InfNFe not set")
		return
	}

	if parser.currentNFe.InfNFe.Versao != "4.00" {
		t.Errorf("Expected version 4.00, got %s", parser.currentNFe.InfNFe.Versao)
	}

	if parser.currentNFe.InfNFe.ID != "NFe35150271780456000160550010000000021800700082" {
		t.Errorf("Expected ID NFe35150271780456000160550010000000021800700082, got %s", parser.currentNFe.InfNFe.ID)
	}
}

func TestParser_processTagB(t *testing.T) {
	parser := NewParser(&LayoutConfig{})
	parser.currentNFe = &NFEData{}

	fields := FieldMap{
		"cUF":    "35",
		"cNF":    "80070008",
		"natOp":  "VENDA",
		"mod":    "55",
		"serie":  "1",
		"nNF":    "2",
		"dhEmi":  "2015-02-19T13:48:00-02:00",
		"tpNF":   "1",
		"idDest": "1",
		"tpAmb":  "2",
	}

	err := parser.processTagB(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if parser.currentNFe.Identificacao == nil {
		t.Error("Identificacao not set")
		return
	}

	ide := parser.currentNFe.Identificacao
	if ide.CUF != "35" {
		t.Errorf("Expected CUF 35, got %s", ide.CUF)
	}
	if ide.CNF != "80070008" {
		t.Errorf("Expected CNF 80070008, got %s", ide.CNF)
	}
	if ide.NatOp != "VENDA" {
		t.Errorf("Expected NatOp VENDA, got %s", ide.NatOp)
	}
}

func TestParser_processTagC(t *testing.T) {
	parser := NewParser(&LayoutConfig{})
	parser.currentNFe = &NFEData{}

	fields := FieldMap{
		"xNome": "PLASTFOAM IND COM PLASTICOS LTDA",
		"xFant": "PLASTFOAM",
		"IE":    "336546371113",
		"CRT":   "3",
	}

	err := parser.processTagC(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if parser.currentNFe.Emitente == nil {
		t.Error("Emitente not set")
		return
	}

	emit := parser.currentNFe.Emitente
	if emit.XNome != "PLASTFOAM IND COM PLASTICOS LTDA" {
		t.Errorf("Expected XNome PLASTFOAM IND COM PLASTICOS LTDA, got %s", emit.XNome)
	}
	if emit.XFant != "PLASTFOAM" {
		t.Errorf("Expected XFant PLASTFOAM, got %s", emit.XFant)
	}
	if emit.IE != "336546371113" {
		t.Errorf("Expected IE 336546371113, got %s", emit.IE)
	}
}

func TestParser_processTagC02(t *testing.T) {
	parser := NewParser(&LayoutConfig{})
	parser.currentNFe = &NFEData{}

	// First process tag C to create Emitente
	fieldsC := FieldMap{"xNome": "Empresa"}
	err := parser.processTagC(fieldsC)
	if err != nil {
		t.Errorf("Unexpected error in processTagC: %v", err)
	}

	// Then process C02
	fieldsC02 := FieldMap{"CNPJ": "71780456000160"}
	err = parser.processTagC02(fieldsC02)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if parser.currentNFe.Emitente.CNPJ != "71780456000160" {
		t.Errorf("Expected CNPJ 71780456000160, got %s", parser.currentNFe.Emitente.CNPJ)
	}
}

func TestParser_processTagI(t *testing.T) {
	parser := NewParser(&LayoutConfig{})
	parser.currentNFe = &NFEData{Itens: []*nfe.Item{}}
	parser.currentItem = 1

	fields := FieldMap{
		"cProd":   "BOLH-S1252",
		"xProd":   "BE6007550 SACO BOLHA SIMPLES",
		"NCM":     "39232190",
		"CFOP":    "5101",
		"uCom":    "MI",
		"qCom":    "0.4000",
		"vUnCom":  "1060.5000",
		"vProd":   "424.20",
	}

	err := parser.processTagI(fields)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(parser.currentNFe.Itens) != 1 {
		t.Errorf("Expected 1 item, got %d", len(parser.currentNFe.Itens))
		return
	}

	item := parser.currentNFe.Itens[0]
	if item.Prod.CProd != "BOLH-S1252" {
		t.Errorf("Expected CProd BOLH-S1252, got %s", item.Prod.CProd)
	}
	if item.Prod.XProd != "BE6007550 SACO BOLHA SIMPLES" {
		t.Errorf("Expected XProd BE6007550 SACO BOLHA SIMPLES, got %s", item.Prod.XProd)
	}
	if item.Prod.NCM != "39232190" {
		t.Errorf("Expected NCM 39232190, got %s", item.Prod.NCM)
	}
}

func TestParser_validateRequiredFields(t *testing.T) {
	tests := []struct {
		name      string
		nfeData   *NFEData
		expectErr bool
	}{
		{
			name: "All required fields present",
			nfeData: &NFEData{
				InfNFe:       &nfe.InfNFe{},
				Identificacao: &nfe.Identificacao{},
				Emitente:     &nfe.Emitente{},
				Itens:        []*nfe.Item{{}},
			},
			expectErr: false,
		},
		{
			name: "Missing InfNFe",
			nfeData: &NFEData{
				Identificacao: &nfe.Identificacao{},
				Emitente:     &nfe.Emitente{},
				Itens:        []*nfe.Item{{}},
			},
			expectErr: true,
		},
		{
			name: "Missing Identificacao",
			nfeData: &NFEData{
				InfNFe:   &nfe.InfNFe{},
				Emitente: &nfe.Emitente{},
				Itens:    []*nfe.Item{{}},
			},
			expectErr: true,
		},
		{
			name: "Missing Emitente",
			nfeData: &NFEData{
				InfNFe:       &nfe.InfNFe{},
				Identificacao: &nfe.Identificacao{},
				Itens:        []*nfe.Item{{}},
			},
			expectErr: true,
		},
		{
			name: "Missing Items",
			nfeData: &NFEData{
				InfNFe:       &nfe.InfNFe{},
				Identificacao: &nfe.Identificacao{},
				Emitente:     &nfe.Emitente{},
				Itens:        []*nfe.Item{},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(&LayoutConfig{})
			parser.currentNFe = tt.nfeData

			err := parser.validateRequiredFields()
			
			if tt.expectErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	fields := FieldMap{
		"existing": "value",
		"empty":    "",
	}

	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "Existing field",
			key:          "existing",
			defaultValue: "default",
			expected:     "value",
		},
		{
			name:         "Missing field",
			key:          "missing",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "Empty field",
			key:          "empty",
			defaultValue: "default",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFieldValue(fields, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	fields := FieldMap{
		"valid":   "123.45",
		"invalid": "not_a_number",
		"empty":   "",
	}

	tests := []struct {
		name      string
		key       string
		expected  float64
		expectErr bool
	}{
		{
			name:      "Valid float",
			key:       "valid",
			expected:  123.45,
			expectErr: false,
		},
		{
			name:      "Invalid float",
			key:       "invalid",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "Empty field",
			key:       "empty",
			expected:  0,
			expectErr: false,
		},
		{
			name:      "Missing field",
			key:       "missing",
			expected:  0,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFloat(fields, tt.key)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	fields := FieldMap{
		"valid":   "123",
		"invalid": "not_a_number",
		"empty":   "",
	}

	tests := []struct {
		name      string
		key       string
		expected  int
		expectErr bool
	}{
		{
			name:      "Valid int",
			key:       "valid",
			expected:  123,
			expectErr: false,
		},
		{
			name:      "Invalid int",
			key:       "invalid",
			expected:  0,
			expectErr: true,
		},
		{
			name:      "Empty field",
			key:       "empty",
			expected:  0,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInt(fields, tt.key)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestParseDateTime(t *testing.T) {
	fields := FieldMap{
		"iso_with_tz":    "2015-02-19T13:48:00-02:00",
		"iso_without_tz": "2015-02-19T13:48:00",
		"simple":         "2015-02-19 13:48:00",
		"date_only":      "2015-02-19",
		"invalid":        "invalid_date",
		"empty":          "",
	}

	tests := []struct {
		name      string
		key       string
		expectErr bool
	}{
		{
			name:      "ISO with timezone",
			key:       "iso_with_tz",
			expectErr: false,
		},
		{
			name:      "ISO without timezone",
			key:       "iso_without_tz",
			expectErr: false,
		},
		{
			name:      "Simple datetime",
			key:       "simple",
			expectErr: false,
		},
		{
			name:      "Date only",
			key:       "date_only",
			expectErr: false,
		},
		{
			name:      "Invalid format",
			key:       "invalid",
			expectErr: true,
		},
		{
			name:      "Empty field",
			key:       "empty",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDateTime(fields, tt.key)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result.IsZero() {
				t.Error("Expected valid time but got zero time")
			}
		})
	}
}

// Benchmark tests
func BenchmarkParser_parseFields(b *testing.B) {
	config := &LayoutConfig{
		Structure: map[string]string{
			"B": "B|cUF|cNF|natOp|mod|serie|nNF|dhEmi|dhSaiEnt|tpNF|idDest|cMunFG|tpImp|tpEmis|cDV|tpAmb|finNFe|indFinal|indPres|indIntermed|procEmi|verProc|dhCont|xJust|",
		},
	}
	
	parser := NewParser(config)
	parts := []string{"B", "35", "80070008", "VENDA", "55", "1", "2", "2015-02-19T13:48:00-02:00", "", "1", "1", "3518800", "1", "1", "2", "2", "1", "0", "0", "3", "3.10.31", "", "", ""}
	structure := config.Structure["B"]

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.parseFields(parts, structure)
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}