package factories

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewContingency(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
	}{
		{
			name:        "empty contingency",
			jsonData:    "",
			expectError: false,
		},
		{
			name:        "valid contingency JSON",
			jsonData:    `{"type":"SVCAN","motive":"SEFAZ fora do ar","timestamp":1234567890,"tpEmis":6}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jsonData:    `{"type":"SVCAN","invalid":}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewContingency(tt.jsonData)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && c == nil {
				t.Error("Contingency should not be nil")
			}
		})
	}
}

func TestContingency_Activate(t *testing.T) {
	tests := []struct {
		name        string
		config      ContingencyConfig
		expectError bool
		expectedType ContingencyType
	}{
		{
			name: "valid SP SVCAN",
			config: ContingencyConfig{
				UF:     "SP",
				Motive: "SEFAZ SP fora do ar por problemas técnicos",
			},
			expectError:  false,
			expectedType: ContingencySVCAN,
		},
		{
			name: "valid RS SVCRS with explicit type",
			config: ContingencyConfig{
				UF:     "RS",
				Motive: "SEFAZ RS fora do ar por problemas técnicos",
				Type:   ContingencySVCRS,
			},
			expectError:  false,
			expectedType: ContingencySVCRS,
		},
		{
			name: "motive too short",
			config: ContingencyConfig{
				UF:     "SP",
				Motive: "Curto",
			},
			expectError: true,
		},
		{
			name: "motive too long",
			config: ContingencyConfig{
				UF:     "SP",
				Motive: strings.Repeat("A", 256),
			},
			expectError: true,
		},
		{
			name: "invalid state",
			config: ContingencyConfig{
				UF:     "XX",
				Motive: "SEFAZ XX fora do ar por problemas técnicos",
			},
			expectError: true,
		},
		{
			name: "invalid contingency type",
			config: ContingencyConfig{
				UF:     "SP",
				Motive: "SEFAZ SP fora do ar por problemas técnicos",
				Type:   "INVALID",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := NewContingency()
			jsonData, err := c.Activate(tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if c.Type != tt.expectedType {
					t.Errorf("Expected type %s, got %s", tt.expectedType, c.Type)
				}
				if c.Motive != tt.config.Motive {
					t.Errorf("Expected motive %s, got %s", tt.config.Motive, c.Motive)
				}
				if c.Timestamp == 0 {
					t.Error("Timestamp should be set")
				}
				if !c.IsActive() {
					t.Error("Contingency should be active")
				}
				if jsonData == "" {
					t.Error("JSON data should not be empty")
				}

				// Verify JSON is valid
				var parsed Contingency
				if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
					t.Errorf("Invalid JSON returned: %v", err)
				}
			}
		})
	}
}

func TestContingency_Deactivate(t *testing.T) {
	c, _ := NewContingency()
	
	// First activate
	_, err := c.Activate(ContingencyConfig{
		UF:     "SP",
		Motive: "SEFAZ SP fora do ar por problemas técnicos",
	})
	if err != nil {
		t.Fatalf("Failed to activate contingency: %v", err)
	}

	if !c.IsActive() {
		t.Error("Contingency should be active")
	}

	// Then deactivate
	jsonData, err := c.Deactivate()
	if err != nil {
		t.Errorf("Failed to deactivate contingency: %v", err)
	}

	if c.IsActive() {
		t.Error("Contingency should not be active")
	}

	if c.Type != "" {
		t.Error("Type should be empty")
	}

	if c.Motive != "" {
		t.Error("Motive should be empty")
	}

	if c.Timestamp != 0 {
		t.Error("Timestamp should be zero")
	}

	if c.TpEmis != EmissionNormal {
		t.Error("TpEmis should be normal")
	}

	// Verify JSON
	var parsed Contingency
	if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
		t.Errorf("Invalid JSON returned: %v", err)
	}
}

func TestContingency_Load(t *testing.T) {
	validJSON := `{
		"type": "SVCAN",
		"motive": "SEFAZ fora do ar por problemas técnicos",
		"timestamp": 1234567890,
		"tpEmis": 6
	}`

	c, _ := NewContingency()
	err := c.Load(validJSON)
	if err != nil {
		t.Errorf("Failed to load valid JSON: %v", err)
	}

	if c.Type != ContingencySVCAN {
		t.Errorf("Expected type SVCAN, got %s", c.Type)
	}

	if c.TpEmis != EmissionSVCAN {
		t.Errorf("Expected tpEmis 6, got %d", c.TpEmis)
	}

	if c.Timestamp != 1234567890 {
		t.Errorf("Expected timestamp 1234567890, got %d", c.Timestamp)
	}

	// Test invalid JSON
	invalidJSON := `{"invalid": json}`
	err = c.Load(invalidJSON)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestContingency_GetFormattedDateTime(t *testing.T) {
	c, _ := NewContingency()

	// Test with zero timestamp
	formatted := c.GetFormattedDateTime()
	if formatted != "" {
		t.Errorf("Expected empty string for zero timestamp, got %s", formatted)
	}

	// Test with actual timestamp
	c.Timestamp = 1640995200 // 2022-01-01 00:00:00 UTC
	formatted = c.GetFormattedDateTime()
	if !strings.Contains(formatted, "2022-01-01T00:00:00") {
		t.Errorf("Expected ISO formatted date, got %s", formatted)
	}
}

func TestContingency_GetContingencyInfo(t *testing.T) {
	c, _ := NewContingency()

	// Test inactive contingency
	info := c.GetContingencyInfo()
	if info["tpEmis"] != int(EmissionNormal) {
		t.Errorf("Expected tpEmis 1 for inactive, got %v", info["tpEmis"])
	}
	if info["dhCont"] != nil {
		t.Error("dhCont should be nil for inactive contingency")
	}

	// Test active contingency
	c.Activate(ContingencyConfig{
		UF:     "SP",
		Motive: "SEFAZ SP fora do ar por problemas técnicos",
	})

	info = c.GetContingencyInfo()
	if info["tpEmis"] != int(EmissionSVCAN) {
		t.Errorf("Expected tpEmis 6 for SVCAN, got %v", info["tpEmis"])
	}
	if info["dhCont"] == nil {
		t.Error("dhCont should not be nil for active contingency")
	}
	if info["xJust"] != c.Motive {
		t.Error("xJust should match contingency motive")
	}
}

func TestValidateContingencyData(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectError bool
	}{
		{
			name:        "valid data",
			jsonData:    `{"type":"SVCAN","motive":"SEFAZ fora do ar por problemas técnicos","timestamp":1234567890,"tpEmis":6}`,
			expectError: false,
		},
		{
			name:        "invalid JSON",
			jsonData:    `{"invalid": }`,
			expectError: true,
		},
		{
			name:        "invalid contingency type",
			jsonData:    `{"type":"INVALID","motive":"SEFAZ fora do ar por problemas técnicos","timestamp":1234567890,"tpEmis":6}`,
			expectError: true,
		},
		{
			name:        "motive too short",
			jsonData:    `{"type":"SVCAN","motive":"Short","timestamp":1234567890,"tpEmis":6}`,
			expectError: true,
		},
		{
			name:        "invalid emission type",
			jsonData:    `{"type":"SVCAN","motive":"SEFAZ fora do ar por problemas técnicos","timestamp":1234567890,"tpEmis":999}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateContingencyData(tt.jsonData)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetStateContingencyType(t *testing.T) {
	tests := []struct {
		name        string
		uf          string
		expected    ContingencyType
		expectError bool
	}{
		{
			name:        "SP uses SVCAN",
			uf:          "SP",
			expected:    ContingencySVCAN,
			expectError: false,
		},
		{
			name:        "RS uses SVCAN",
			uf:          "RS",
			expected:    ContingencySVCAN,
			expectError: false,
		},
		{
			name:        "BA uses SVCRS",
			uf:          "BA",
			expected:    ContingencySVCRS,
			expectError: false,
		},
		{
			name:        "lowercase sp",
			uf:          "sp",
			expected:    ContingencySVCAN,
			expectError: false,
		},
		{
			name:        "unknown state",
			uf:          "XX",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetStateContingencyType(tt.uf)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestContingencyBuilder(t *testing.T) {
	// Test fluent interface
	c, jsonData, err := NewContingencyBuilder().
		ForState("SP").
		WithMotive("SEFAZ SP fora do ar por problemas técnicos").
		WithType(ContingencySVCAN).
		Activate()

	if err != nil {
		t.Errorf("Builder should not fail: %v", err)
	}

	if c.Type != ContingencySVCAN {
		t.Errorf("Expected SVCAN, got %s", c.Type)
	}

	if !c.IsActive() {
		t.Error("Contingency should be active")
	}

	if jsonData == "" {
		t.Error("JSON data should not be empty")
	}

	// Test builder without explicit type
	c2, _, err := NewContingencyBuilder().
		ForState("BA").
		WithMotive("SEFAZ BA fora do ar por problemas técnicos").
		Activate()

	if err != nil {
		t.Errorf("Builder should not fail: %v", err)
	}

	if c2.Type != ContingencySVCRS {
		t.Errorf("Expected SVCRS for BA, got %s", c2.Type)
	}
}

func TestCreateContingency(t *testing.T) {
	// Test with default type
	c, jsonData, err := CreateContingency("SP", "SEFAZ SP fora do ar por problemas técnicos")
	if err != nil {
		t.Errorf("CreateContingency should not fail: %v", err)
	}

	if c.Type != ContingencySVCAN {
		t.Errorf("Expected SVCAN for SP, got %s", c.Type)
	}

	if jsonData == "" {
		t.Error("JSON data should not be empty")
	}

	// Test with explicit type
	c2, _, err := CreateContingency("RS", "SEFAZ RS fora do ar", ContingencySVCRS)
	if err != nil {
		t.Errorf("CreateContingency should not fail: %v", err)
	}

	if c2.Type != ContingencySVCRS {
		t.Errorf("Expected SVCRS, got %s", c2.Type)
	}

	// Test error case
	_, _, err = CreateContingency("XX", "Invalid state")
	if err == nil {
		t.Error("Expected error for invalid state")
	}
}

func TestContingency_ToJSON(t *testing.T) {
	c, _ := NewContingency()
	c.Type = ContingencySVCAN
	c.Motive = "Test motive for contingency activation"
	c.Timestamp = time.Now().Unix()
	c.TpEmis = EmissionSVCAN

	jsonData, err := c.ToJSON()
	if err != nil {
		t.Errorf("ToJSON should not fail: %v", err)
	}

	// Verify JSON can be parsed back
	var parsed Contingency
	if err := json.Unmarshal([]byte(jsonData), &parsed); err != nil {
		t.Errorf("Generated JSON should be valid: %v", err)
	}

	if parsed.Type != c.Type {
		t.Error("JSON should preserve type")
	}
	if parsed.Motive != c.Motive {
		t.Error("JSON should preserve motive")
	}
}

func TestContingency_String(t *testing.T) {
	c, _ := NewContingency()
	c.Type = ContingencySVCAN
	c.Motive = "Test motive for contingency activation"

	str := c.String()
	if str == "" {
		t.Error("String should not be empty")
	}

	// Should be valid JSON
	var parsed Contingency
	if err := json.Unmarshal([]byte(str), &parsed); err != nil {
		t.Errorf("String should return valid JSON: %v", err)
	}
}

// Benchmark tests
func BenchmarkContingency_Activate(b *testing.B) {
	config := ContingencyConfig{
		UF:     "SP",
		Motive: "SEFAZ SP fora do ar por problemas técnicos",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := NewContingency()
		c.Activate(config)
	}
}

func BenchmarkCreateContingency(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CreateContingency("SP", "SEFAZ SP fora do ar por problemas técnicos")
	}
}