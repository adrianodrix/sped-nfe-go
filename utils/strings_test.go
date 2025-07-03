package utils

import (
	"testing"
)

func TestRemoveAccents(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ação", "Acao"},
		{"Cação", "Cacao"},
		{"São Paulo", "Sao Paulo"},
		{"Mañana", "Manana"},
		{"Café", "Cafe"},
		{"Naïve", "Naive"},
		{"Zürich", "Zurich"},
		{"Normal text", "Normal text"},
		{"", ""},
		{"123", "123"},
		{"áéíóúàèìòùâêîôûãõäëïöüç", "aeiouaeiouaeiouaoaeiouc"},
		{"ÁÉÍÓÚÀÈÌÒÙÂÊÎÔÛÃÕÄËÏÖÜÇ", "AEIOUAEIOUAEIOUAOAEIOUC"},
	}

	for _, test := range tests {
		result := RemoveAccents(test.input)
		if result != test.expected {
			t.Errorf("RemoveAccents('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestNormalizeString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Ação Empresarial", "ACAO EMPRESARIAL"},
		{"  Café com Açúcar  ", "CAFE COM ACUCAR"},
		{"São Paulo - SP", "SAO PAULO - SP"},
		{"123 ABC", "123 ABC"},
		{"", ""},
		{"normal", "NORMAL"},
		{"Maçã & Pêra", "MACA & PERA"},
		{"Text\x00With\x1FControl", "TEXTWITHCONTROL"}, // Control characters
	}

	for _, test := range tests {
		result := NormalizeString(test.input)
		if result != test.expected {
			t.Errorf("NormalizeString('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestRemoveInvalidXMLChars(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Normal text", "Normal text"},
		{"Text\x00with\x01invalid", "Textwithinvalid"},
		{"Text\x08\x0B\x0Cmore", "Textmore"},
		{"Valid\x09\x0A\x0Dchars", "Valid\x09\x0A\x0Dchars"}, // Tab, LF, CR are valid
		{"Text with spaces", "Text with spaces"},
		{"", ""},
	}

	for _, test := range tests {
		result := removeInvalidXMLChars(test.input)
		if result != test.expected {
			t.Errorf("removeInvalidXMLChars('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestIsValidXMLChar(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{0x09, true},     // Tab
		{0x0A, true},     // LF
		{0x0D, true},     // CR
		{0x20, true},     // Space
		{0x7F, true},     // DEL
		{0xD7FF, true},   // End of first valid range
		{0xE000, true},   // Start of second valid range
		{0xFFFD, true},   // End of second valid range
		{0x10000, true},  // Start of third valid range
		{0x10FFFF, true}, // End of third valid range
		{0x00, false},    // NULL
		{0x01, false},    // Control char
		{0x08, false},    // Backspace
		{0x0B, false},    // Vertical tab
		{0x0C, false},    // Form feed
		{0x0E, false},    // Shift out
		{0x1F, false},    // Unit separator
		{0xD800, false},  // Surrogate
		{0xDFFF, false},  // Surrogate
		{0xFFFE, false},  // Invalid
		{0xFFFF, false},  // Invalid
	}

	for _, test := range tests {
		result := isValidXMLChar(test.char)
		if result != test.expected {
			t.Errorf("isValidXMLChar(0x%04X) = %t, expected %t", test.char, result, test.expected)
		}
	}
}

func TestFormatMoney(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{123.45, "123.45"},
		{100.0, "100.00"},
		{0.1, "0.10"},
		{0.001, "0.00"},
		{1234.567, "1234.57"},
		{0, "0.00"},
		{-123.45, "-123.45"},
	}

	for _, test := range tests {
		result := FormatMoney(test.input)
		if result != test.expected {
			t.Errorf("FormatMoney(%f) = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestFormatQuantity(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{123.45, "123.45"},
		{100.0, "100"},
		{0.1, "0.1"},
		{0.1000, "0.1"},
		{1234.5670, "1234.567"},
		{0, "0"},
		{-123.45, "-123.45"},
		{1.0000, "1"},
		{0.0001, "0.0001"},
	}

	for _, test := range tests {
		result := FormatQuantity(test.input)
		if result != test.expected {
			t.Errorf("FormatQuantity(%f) = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{12.3456, "12.3456"},
		{100.0, "100.0000"},
		{0.1, "0.1000"},
		{0, "0.0000"},
		{-5.25, "-5.2500"},
	}

	for _, test := range tests {
		result := FormatPercentage(test.input)
		if result != test.expected {
			t.Errorf("FormatPercentage(%f) = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestPadLeft(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		padChar  rune
		expected string
	}{
		{"123", 5, '0', "00123"},
		{"hello", 10, ' ', "     hello"},
		{"test", 3, 'x', "test"}, // Already longer
		{"", 3, '0', "000"},
		{"abc", 3, '0', "abc"}, // Exact length
	}

	for _, test := range tests {
		result := PadLeft(test.input, test.length, test.padChar)
		if result != test.expected {
			t.Errorf("PadLeft('%s', %d, '%c') = '%s', expected '%s'", test.input, test.length, test.padChar, result, test.expected)
		}
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		padChar  rune
		expected string
	}{
		{"123", 5, '0', "12300"},
		{"hello", 10, ' ', "hello     "},
		{"test", 3, 'x', "test"}, // Already longer
		{"", 3, '0', "000"},
		{"abc", 3, '0', "abc"}, // Exact length
	}

	for _, test := range tests {
		result := PadRight(test.input, test.length, test.padChar)
		if result != test.expected {
			t.Errorf("PadRight('%s', %d, '%c') = '%s', expected '%s'", test.input, test.length, test.padChar, result, test.expected)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input     string
		maxLength int
		expected  string
	}{
		{"Hello World", 5, "Hello"},
		{"Test", 10, "Test"},
		{"", 5, ""},
		{"Café", 3, "Caf"},      // Unicode handling
		{"测试", 1, "测"},          // Chinese characters
		{"abcdef", 6, "abcdef"}, // Exact length
	}

	for _, test := range tests {
		result := TruncateString(test.input, test.maxLength)
		if result != test.expected {
			t.Errorf("TruncateString('%s', %d) = '%s', expected '%s'", test.input, test.maxLength, result, test.expected)
		}
	}
}

func TestOnlyNumbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123-456-789", "123456789"},
		{"ABC123DEF456", "123456"},
		{"(11) 98765-4321", "11987654321"},
		{"", ""},
		{"NoNumbers", ""},
		{"123", "123"},
		{"12.34", "1234"},
	}

	for _, test := range tests {
		result := OnlyNumbers(test.input)
		if result != test.expected {
			t.Errorf("OnlyNumbers('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestOnlyAlphaNumeric(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABC123!@#", "ABC123"},
		{"Hello, World!", "HelloWorld"},
		{"", ""},
		{"12-34-56", "123456"},
		{"Test_123", "Test123"},
		{"NoSpecial", "NoSpecial"},
	}

	for _, test := range tests {
		result := OnlyAlphaNumeric(test.input)
		if result != test.expected {
			t.Errorf("OnlyAlphaNumeric('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestCleanFileName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test<file>.txt", "test_file_.txt"},
		{"normal_file.pdf", "normal_file.pdf"},
		{"file:with|special?chars", "file_with_special_chars"},
		{"file\x00with\x1fcontrol.doc", "filewithcontrol.doc"},
		{"   file.txt   ", "file.txt"},
		{".hidden_file.", "hidden_file"},
		{"", ""},
	}

	for _, test := range tests {
		result := CleanFileName(test.input)
		if result != test.expected {
			t.Errorf("CleanFileName('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestNormalizeCEP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01234-567", "01234567"},
		{"1234567", "01234567"},
		{"123", "00000123"},
		{"12345678", "12345678"},
		{"", "00000000"},
		{"ABC12345", "00012345"}, // Only numbers
	}

	for _, test := range tests {
		result := NormalizeCEP(test.input)
		if result != test.expected {
			t.Errorf("NormalizeCEP('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestFormatCEP(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"01234567", "01234-567"},
		{"01234-567", "01234-567"},
		{"1234567", "01234-567"},
		{"123", "00000-123"},
		{"invalid", "00000-000"}, // Invalid input gets normalized
		{"", "00000-000"},
	}

	for _, test := range tests {
		result := FormatCEP(test.input)
		if result != test.expected {
			t.Errorf("FormatCEP('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"(11) 98765-4321", "11987654321"},
		{"5511987654321", "11987654321"}, // Remove country code
		{"11987654321", "11987654321"},
		{"", ""},
		{"ABC123", "123"},
		{"551234567890", "1234567890"}, // Remove country code only if 12 digits total
	}

	for _, test := range tests {
		result := NormalizePhone(test.input)
		if result != test.expected {
			t.Errorf("NormalizePhone('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestFormatPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1134567890", "(11) 3456-7890"},       // Fixed line
		{"11987654321", "(11) 98765-4321"},     // Mobile
		{"123", "123"},                         // Invalid, return original
		{"12345678901", "(12) 34567-8901"},     // 11 digits, format as mobile
		{"(11) 3456-7890", "(11) 3456-7890"},   // Fixed line with formatting
		{"(11) 98765-4321", "(11) 98765-4321"}, // Mobile with formatting
	}

	for _, test := range tests {
		result := FormatPhone(test.input)
		if result != test.expected {
			t.Errorf("FormatPhone('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestRemoveExtraSpaces(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello   world  ", "hello world"},
		{"test\t\n\r  multiple", "test multiple"},
		{"normal text", "normal text"},
		{"", ""},
		{"   ", ""},
		{"a", "a"},
	}

	for _, test := range tests {
		result := RemoveExtraSpaces(test.input)
		if result != test.expected {
			t.Errorf("RemoveExtraSpaces('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestCapitalizeWords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"HELLO WORLD", "Hello World"},
		{"tESt CaSe", "Test Case"},
		{"", ""},
		{"a", "A"},
		{"one two three", "One Two Three"},
	}

	for _, test := range tests {
		result := CapitalizeWords(test.input)
		if result != test.expected {
			t.Errorf("CapitalizeWords('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n\r", true},
		{"hello", false},
		{" hello ", false},
		{"a", false},
	}

	for _, test := range tests {
		result := IsEmpty(test.input)
		if result != test.expected {
			t.Errorf("IsEmpty('%s') = %t, expected %t", test.input, result, test.expected)
		}
	}
}

func TestContainsOnlyDigits(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123456", true},
		{"", false},
		{"123a456", false},
		{"12 34", false},
		{"0", true},
		{"123.456", false},
		{"987654321", true},
	}

	for _, test := range tests {
		result := ContainsOnlyDigits(test.input)
		if result != test.expected {
			t.Errorf("ContainsOnlyDigits('%s') = %t, expected %t", test.input, result, test.expected)
		}
	}
}

func TestToASCII(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Café", "Cafe"},
		{"Açúcar", "Acucar"},
		{"测试", ""}, // Chinese characters removed
		{"Hello World", "Hello World"},
		{"", ""},
		{"Naïve résumé", "Naive resume"},
	}

	for _, test := range tests {
		result := ToASCII(test.input)
		if result != test.expected {
			t.Errorf("ToASCII('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestEscapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello & World", "Hello &amp; World"},
		{"<test>", "&lt;test&gt;"},
		{"\"quoted\"", "&quot;quoted&quot;"},
		{"'single'", "&#39;single&#39;"},
		{"Normal text", "Normal text"},
		{"", ""},
		{"A & B < C > D \"E\" 'F'", "A &amp; B &lt; C &gt; D &quot;E&quot; &#39;F&#39;"},
	}

	for _, test := range tests {
		result := EscapeXML(test.input)
		if result != test.expected {
			t.Errorf("EscapeXML('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestUnescapeXML(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello &amp; World", "Hello & World"},
		{"&lt;test&gt;", "<test>"},
		{"&quot;quoted&quot;", "\"quoted\""},
		{"&#39;single&#39;", "'single'"},
		{"Normal text", "Normal text"},
		{"", ""},
		{"A &amp; B &lt; C &gt; D &quot;E&quot; &#39;F&#39;", "A & B < C > D \"E\" 'F'"},
	}

	for _, test := range tests {
		result := UnescapeXML(test.input)
		if result != test.expected {
			t.Errorf("UnescapeXML('%s') = '%s', expected '%s'", test.input, result, test.expected)
		}
	}
}

func TestFormatForXML(t *testing.T) {
	tests := []struct {
		input     string
		maxLength int
		expected  string
	}{
		{"Café & Açúcar", 20, "CAFE &amp; ACUCAR"},
		{"Test <tag>", 10, "TEST &lt;TAG&gt;"},
		{"Very long text that should be truncated", 10, "VERY LONG "},
		{"", 10, ""},
		{"Normal", 0, "NORMAL"}, // No truncation
	}

	for _, test := range tests {
		result := FormatForXML(test.input, test.maxLength)
		if result != test.expected {
			t.Errorf("FormatForXML('%s', %d) = '%s', expected '%s'", test.input, test.maxLength, result, test.expected)
		}
	}
}

func TestZeroFill(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"123", 5, "00123"},
		{"hello", 3, "hello"}, // Already longer
		{"", 3, "000"},
		{"12345", 5, "12345"}, // Exact length
	}

	for _, test := range tests {
		result := ZeroFill(test.input, test.length)
		if result != test.expected {
			t.Errorf("ZeroFill('%s', %d) = '%s', expected '%s'", test.input, test.length, result, test.expected)
		}
	}
}

func TestSpaceFill(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"123", 5, "123  "},
		{"hello", 3, "hello"}, // Already longer
		{"", 3, "   "},
		{"12345", 5, "12345"}, // Exact length
	}

	for _, test := range tests {
		result := SpaceFill(test.input, test.length)
		if result != test.expected {
			t.Errorf("SpaceFill('%s', %d) = '%s', expected '%s'", test.input, test.length, result, test.expected)
		}
	}
}
