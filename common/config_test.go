package common

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/adrianodrix/sped-nfe-go/types"
)

func TestValidateConfig(t *testing.T) {
	// Test valid configuration
	validConfig := &Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000190",
		SiglaUF:     "SP",
		Schemes:     "/path/to/schemes",
		Versao:      "4.00",
		Timeout:     30,
	}

	if err := ValidateConfig(validConfig); err != nil {
		t.Errorf("Valid configuration should not return error, got: %v", err)
	}
}

func TestValidateConfigNil(t *testing.T) {
	err := ValidateConfig(nil)
	if err == nil {
		t.Error("Nil configuration should return error")
	}
	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("Expected 'cannot be nil' error, got: %v", err)
	}
}

func TestValidateConfigInvalidEnvironment(t *testing.T) {
	config := &Config{
		TpAmb:       types.Environment(99),
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000190",
		SiglaUF:     "SP",
		Schemes:     "/path/to/schemes",
		Versao:      "4.00",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Invalid environment should return error")
	}
	if !strings.Contains(err.Error(), "invalid environment") {
		t.Errorf("Expected 'invalid environment' error, got: %v", err)
	}
}

func TestValidateConfigEmptyRazaoSocial(t *testing.T) {
	config := &Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "   ",
		CNPJ:        "12345678000190",
		SiglaUF:     "SP",
		Schemes:     "/path/to/schemes",
		Versao:      "4.00",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Empty razao social should return error")
	}
}

func TestValidateConfigInvalidCNPJ(t *testing.T) {
	tests := []struct {
		cnpj        string
		shouldError bool
		description string
	}{
		{"12345678000190", false, "valid CNPJ"},
		{"12345678901", false, "valid CPF"},
		{"123456789", true, "too short"},
		{"123456789012345", true, "too long"},
		{"11111111111111", true, "all same digits"},
		{"abc123def456", true, "contains letters"},
	}

	for _, test := range tests {
		config := &Config{
			TpAmb:       types.Homologation,
			RazaoSocial: "Empresa Teste LTDA",
			CNPJ:        test.cnpj,
			SiglaUF:     "SP",
			Schemes:     "/path/to/schemes",
			Versao:      "4.00",
		}

		err := ValidateConfig(config)
		if test.shouldError && err == nil {
			t.Errorf("CNPJ '%s' (%s) should return error", test.cnpj, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("CNPJ '%s' (%s) should not return error, got: %v", test.cnpj, test.description, err)
		}
	}
}

func TestValidateConfigInvalidUF(t *testing.T) {
	config := &Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000190",
		SiglaUF:     "XX",
		Schemes:     "/path/to/schemes",
		Versao:      "4.00",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Invalid UF should return error")
	}
}

func TestValidateConfigInvalidVersion(t *testing.T) {
	config := &Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000190",
		SiglaUF:     "SP",
		Schemes:     "/path/to/schemes",
		Versao:      "5.00",
	}

	err := ValidateConfig(config)
	if err == nil {
		t.Error("Invalid version should return error")
	}
}

func TestValidateConfigTimeout(t *testing.T) {
	tests := []struct {
		timeout     int
		shouldError bool
		description string
	}{
		{30, false, "valid timeout"},
		{5, false, "minimum timeout"},
		{300, false, "maximum timeout"},
		{4, true, "below minimum"},
		{301, true, "above maximum"},
	}

	for _, test := range tests {
		config := &Config{
			TpAmb:       types.Homologation,
			RazaoSocial: "Empresa Teste LTDA",
			CNPJ:        "12345678000190",
			SiglaUF:     "SP",
			Schemes:     "/path/to/schemes",
			Versao:      "4.00",
			Timeout:     test.timeout,
		}

		err := ValidateConfig(config)
		if test.shouldError && err == nil {
			t.Errorf("Timeout %d (%s) should return error", test.timeout, test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("Timeout %d (%s) should not return error, got: %v", test.timeout, test.description, err)
		}
	}
}

func TestParseConfigJSON(t *testing.T) {
	jsonData := `{
		"tpAmb": 2,
		"razaosocial": "Empresa Teste LTDA",
		"cnpj": "12345678000190",
		"siglaUF": "SP",
		"schemes": "/path/to/schemes",
		"versao": "4.00"
	}`

	config, err := ParseConfigJSON([]byte(jsonData))
	if err != nil {
		t.Errorf("Valid JSON should not return error, got: %v", err)
	}

	if config.TpAmb != types.Homologation {
		t.Errorf("Expected environment Homologation, got %v", config.TpAmb)
	}

	if config.RazaoSocial != "Empresa Teste LTDA" {
		t.Errorf("Expected 'Empresa Teste LTDA', got '%s'", config.RazaoSocial)
	}

	// Check default timeout was applied
	if config.Timeout != types.DefaultTimeoutSeconds {
		t.Errorf("Expected default timeout %d, got %d", types.DefaultTimeoutSeconds, config.Timeout)
	}
}

func TestParseConfigJSONInvalid(t *testing.T) {
	tests := []struct {
		jsonData    string
		description string
	}{
		{"", "empty JSON"},
		{"{invalid json", "malformed JSON"},
		{`{"tpAmb": "invalid"}`, "invalid environment type"},
		{`{"tpAmb": 2}`, "missing required fields"},
	}

	for _, test := range tests {
		_, err := ParseConfigJSON([]byte(test.jsonData))
		if err == nil {
			t.Errorf("Invalid JSON (%s) should return error", test.description)
		}
	}
}

func TestConfigGetUF(t *testing.T) {
	tests := []struct {
		siglaUF  string
		expected types.UF
		hasError bool
	}{
		{"SP", types.SP, false},
		{"RJ", types.RJ, false},
		{"sp", types.SP, false}, // lowercase should work
		{"XX", 0, true},         // invalid UF
	}

	for _, test := range tests {
		config := &Config{SiglaUF: test.siglaUF}
		uf, err := config.GetUF()

		if test.hasError && err == nil {
			t.Errorf("UF '%s' should return error", test.siglaUF)
		}
		if !test.hasError && err != nil {
			t.Errorf("UF '%s' should not return error, got: %v", test.siglaUF, err)
		}
		if !test.hasError && uf != test.expected {
			t.Errorf("UF '%s' expected %v, got %v", test.siglaUF, test.expected, uf)
		}
	}
}

func TestConfigToJSON(t *testing.T) {
	config := &Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000190",
		SiglaUF:     "SP",
		Schemes:     "/path/to/schemes",
		Versao:      "4.00",
		Timeout:     30,
	}

	jsonData, err := config.ToJSON()
	if err != nil {
		t.Errorf("ToJSON should not return error, got: %v", err)
	}

	// Verify it's valid JSON by parsing it back
	var parsed Config
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Errorf("Generated JSON should be valid, got error: %v", err)
	}

	if parsed.RazaoSocial != config.RazaoSocial {
		t.Errorf("Parsed config should match original")
	}
}

func TestNewClientConfig(t *testing.T) {
	config := NewClientConfig()

	if config.Environment != types.Homologation {
		t.Errorf("Expected default environment Homologation, got %v", config.Environment)
	}

	if config.UF != types.SP {
		t.Errorf("Expected default UF SP, got %v", config.UF)
	}

	expectedTimeout := time.Duration(types.DefaultTimeoutSeconds) * time.Second
	if config.Timeout != expectedTimeout {
		t.Errorf("Expected default timeout %v, got %v", expectedTimeout, config.Timeout)
	}
}

func TestValidateClientConfig(t *testing.T) {
	validConfig := &ClientConfig{
		Environment: types.Production,
		UF:          types.RJ,
		Timeout:     30 * time.Second,
	}

	if err := ValidateClientConfig(validConfig); err != nil {
		t.Errorf("Valid client config should not return error, got: %v", err)
	}
}

func TestValidateProxyConfig(t *testing.T) {
	validPort := "8080"
	validIP := "192.168.1.1"
	invalidPort := "99999"
	invalidIP := "invalid.ip"

	tests := []struct {
		proxy       *ProxyConfig
		shouldError bool
		description string
	}{
		{nil, false, "nil proxy config"},
		{&ProxyConfig{}, false, "empty proxy config"},
		{&ProxyConfig{ProxyPort: &validPort, ProxyIP: &validIP}, false, "valid proxy config"},
		{&ProxyConfig{ProxyPort: &invalidPort}, true, "invalid port number"},
		{&ProxyConfig{ProxyIP: &invalidIP}, true, "invalid IP format"},
	}

	for _, test := range tests {
		err := validateProxyConfig(test.proxy)
		if test.shouldError && err == nil {
			t.Errorf("Proxy config (%s) should return error", test.description)
		}
		if !test.shouldError && err != nil {
			t.Errorf("Proxy config (%s) should not return error, got: %v", test.description, err)
		}
	}
}
