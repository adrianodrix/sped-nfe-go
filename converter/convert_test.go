package converter

import (
	"testing"
)

func TestNewConverter(t *testing.T) {
	tests := []struct {
		name    string
		layout  Layout
		wantErr bool
	}{
		{
			name:    "Valid Layout400Local",
			layout:  Layout400Local,
			wantErr: false,
		},
		{
			name:    "Valid Layout400Sebrae",
			layout:  Layout400Sebrae,
			wantErr: false,
		},
		{
			name:    "Valid Layout310Local",
			layout:  Layout310Local,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converter, err := NewConverter(tt.layout)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.wantErr && converter == nil {
				t.Error("Expected converter but got nil")
			}
		})
	}
}

func TestConverter_parseTXTContent(t *testing.T) {
	converter, err := NewConverter(Layout400Local)
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name        string
		content     string
		expectNFes  int
		expectError bool
	}{
		{
			name: "Valid single NFe",
			content: `NOTAFISCAL|1|
A|4.00|NFe35150271780456000160550010000000021800700082||
B|35|80070008|VENDA|55|1|2|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||
C|PLASTFOAM IND COM PLASTICOS LTDA|PLASTFOAM|336546371113||184394|2222600|3|
C02|71780456000160|
I|BOLH-S1252||BE6007550 SACO BOLHA|39232190||5101|MI|0.4000|1060.5000|424.20||MI|0.4000|1060.5000|||1|12345|1|`,
			expectNFes:  1,
			expectError: false,
		},
		{
			name: "Valid multiple NFes",
			content: `NOTAFISCAL|2|
A|4.00|NFe1||
B|35|80070008|VENDA|55|1|1|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||
C|EMPRESA 1|EMPRESA1|123456789||184394|2222600|3|
C02|12345678000195|
I|PROD1||PRODUTO 1|39232190||5101|MI|1.0000|100.0000|100.00||MI|1.0000|100.0000|||1|12345|1|
A|4.00|NFe2||
B|35|80070009|VENDA|55|1|2|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||
C|EMPRESA 2|EMPRESA2|987654321||184394|2222600|3|
C02|98765432000187|
I|PROD2||PRODUTO 2|39232190||5101|MI|1.0000|200.0000|200.00||MI|1.0000|200.0000|||1|12345|1|`,
			expectNFes:  2,
			expectError: false,
		},
		{
			name:        "Empty content",
			content:     "",
			expectNFes:  0,
			expectError: true,
		},
		{
			name:        "Missing NOTAFISCAL header",
			content:     "A|4.00|NFe1||",
			expectNFes:  0,
			expectError: true,
		},
		{
			name: "NFe count mismatch",
			content: `NOTAFISCAL|2|
A|4.00|NFe1||
B|35|80070008|VENDA|55|1|1|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||`,
			expectNFes:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nfes, err := converter.parseTXTContent([]byte(tt.content))

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(nfes) != tt.expectNFes {
				t.Errorf("Expected %d NFes, got %d", tt.expectNFes, len(nfes))
			}
		})
	}
}

func TestConverter_splitLines(t *testing.T) {
	converter, _ := NewConverter(Layout400Local)

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Simple lines",
			content:  "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "Lines with empty lines",
			content:  "line1\n\nline2\n   \nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "Lines with trailing spaces",
			content:  "line1  \n  line2  \nline3   ",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "Empty content",
			content:  "",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.splitLines([]byte(tt.content))

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(result))
				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("Line %d: expected %q, got %q", i, tt.expected[i], line)
				}
			}
		})
	}
}

func TestConverter_splitNFes(t *testing.T) {
	converter, _ := NewConverter(Layout400Local)

	tests := []struct {
		name     string
		lines    []string
		expected int
	}{
		{
			name: "Single NFe",
			lines: []string{
				"A|4.00|NFe1||",
				"B|35|80070008|VENDA|55|1|1|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||",
				"C|EMPRESA|EMPRESA|123||184394|2222600|3|",
			},
			expected: 1,
		},
		{
			name: "Multiple NFes",
			lines: []string{
				"A|4.00|NFe1||",
				"B|35|80070008|VENDA|55|1|1|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||",
				"C|EMPRESA 1|EMPRESA1|123||184394|2222600|3|",
				"A|4.00|NFe2||",
				"B|35|80070009|VENDA|55|1|2|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||",
				"C|EMPRESA 2|EMPRESA2|456||184394|2222600|3|",
			},
			expected: 2,
		},
		{
			name:     "Empty lines",
			lines:    []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := converter.splitNFes(tt.lines)

			if len(result) != tt.expected {
				t.Errorf("Expected %d NFes, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestConverter_ValidateTXT(t *testing.T) {
	converter, err := NewConverter(Layout400Local)
	if err != nil {
		t.Fatalf("Failed to create converter: %v", err)
	}

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name: "Valid TXT",
			content: `NOTAFISCAL|1|
A|4.00|NFe35150271780456000160550010000000021800700082||
B|35|80070008|VENDA|55|1|2|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||
C|PLASTFOAM IND COM PLASTICOS LTDA|PLASTFOAM|336546371113||184394|2222600|3|
C02|71780456000160|
I|BOLH-S1252||BE6007550 SACO BOLHA|39232190||5101|MI|0.4000|1060.5000|424.20||MI|0.4000|1060.5000|||1|12345|1|`,
			expectError: false,
		},
		{
			name: "Invalid field count",
			content: `NOTAFISCAL|1|
A|4.00|NFe35150271780456000160550010000000021800700082|
B|35|VENDA|55|1|2|2015-02-19T13:48:00-02:00|
C|PLASTFOAM IND COM PLASTICOS LTDA|PLASTFOAM|
C02|71780456000160|
I|BOLH-S1252||BE6007550 SACO BOLHA|39232190||5101|MI|0.4000|1060.5000|424.20||MI|0.4000|1060.5000|||1|12345|1|`,
			expectError: true,
		},
		{
			name: "Missing required tags",
			content: `NOTAFISCAL|1|
A|4.00|NFe35150271780456000160550010000000021800700082||
C|PLASTFOAM IND COM PLASTICOS LTDA|PLASTFOAM|336546371113||184394|2222600|3|`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := converter.ValidateTXT([]byte(tt.content))

			if tt.expectError {
				if err == nil && len(errors) == 0 {
					t.Error("Expected validation errors but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
				if len(errors) > 0 {
					t.Errorf("Unexpected validation errors: %v", errors)
				}
			}
		})
	}
}

func TestGetSupportedLayouts(t *testing.T) {
	layouts := GetSupportedLayouts()

	if len(layouts) == 0 {
		t.Error("Expected at least one supported layout")
	}

	for _, layout := range layouts {
		if layout == "" {
			t.Error("Empty layout description")
		}
	}
}

func TestLayoutNames(t *testing.T) {
	tests := []struct {
		layout       Layout
		expectedName string
		expectedVer  string
	}{
		{Layout400Local, "NFe 4.00 Local", "4.00"},
		{Layout400Sebrae, "NFe 4.00 SEBRAE", "4.00"},
		{Layout310Local, "NFe 3.10 Local", "3.10"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			converter := &Converter{layout: tt.layout}

			name := converter.getLayoutName()
			if name != tt.expectedName {
				t.Errorf("Expected name %q, got %q", tt.expectedName, name)
			}

			version := converter.getLayoutVersion()
			if version != tt.expectedVer {
				t.Errorf("Expected version %q, got %q", tt.expectedVer, version)
			}
		})
	}
}

func TestConversionResult(t *testing.T) {
	result := &ConversionResult{
		XMLs:     [][]byte{[]byte("<xml>test</xml>")},
		Count:    1,
		Warnings: []string{"Warning 1"},
		Errors:   []string{"Error 1"},
	}

	if result.Count != 1 {
		t.Errorf("Expected count 1, got %d", result.Count)
	}

	if len(result.XMLs) != 1 {
		t.Errorf("Expected 1 XML, got %d", len(result.XMLs))
	}

	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}

	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}
}

func TestNewConverterWithConfig(t *testing.T) {
	config := &LayoutConfig{
		Name:    "Test Layout",
		Version: "4.00",
		Structure: map[string]string{
			"A": "A|versao|Id|pk_nItem|",
			"B": "B|cUF|cNF|natOp|mod|serie|nNF|",
		},
	}

	converter := NewConverterWithConfig(config)

	if converter == nil {
		t.Error("Expected converter but got nil")
	}

	if converter.layoutConfig.Name != "Test Layout" {
		t.Errorf("Expected name 'Test Layout', got %q", converter.layoutConfig.Name)
	}
}

// Benchmark tests
func BenchmarkConverter_parseTXTContent(b *testing.B) {
	converter, _ := NewConverter(Layout400Local)

	content := `NOTAFISCAL|1|
A|4.00|NFe35150271780456000160550010000000021800700082||
B|35|80070008|VENDA|55|1|2|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||
C|PLASTFOAM IND COM PLASTICOS LTDA|PLASTFOAM|336546371113||184394|2222600|3|
C02|71780456000160|
I|BOLH-S1252||BE6007550 SACO BOLHA|39232190||5101|MI|0.4000|1060.5000|424.20||MI|0.4000|1060.5000|||1|12345|1|`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.parseTXTContent([]byte(content))
		if err != nil {
			b.Fatalf("Parse error: %v", err)
		}
	}
}

func BenchmarkConverter_ValidateTXT(b *testing.B) {
	converter, _ := NewConverter(Layout400Local)

	content := `NOTAFISCAL|1|
A|4.00|NFe35150271780456000160550010000000021800700082||
B|35|80070008|VENDA|55|1|2|2015-02-19T13:48:00-02:00||1|1|3518800|1|1|2|2|1|0|0|3|3.10.31|||
C|PLASTFOAM IND COM PLASTICOS LTDA|PLASTFOAM|336546371113||184394|2222600|3|
C02|71780456000160|
I|BOLH-S1252||BE6007550 SACO BOLHA|39232190||5101|MI|0.4000|1060.5000|424.20||MI|0.4000|1060.5000|||1|12345|1|`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := converter.ValidateTXT([]byte(content))
		if err != nil {
			b.Fatalf("Validation error: %v", err)
		}
	}
}
