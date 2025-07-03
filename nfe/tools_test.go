package nfe

import (
	"context"
	"testing"
	"time"

	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/types"
	"github.com/adrianodrix/sped-nfe-go/webservices"
)

func TestNewTools(t *testing.T) {
	// Test with valid config
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, err := NewTools(config, webservices.NewResolver())
	if err != nil {
		t.Fatalf("NewTools should not return error with valid config: %v", err)
	}

	if tools == nil {
		t.Fatal("NewTools should not return nil")
	}

	if tools.config != config {
		t.Error("Tools should store the provided config")
	}

	if tools.model != "55" {
		t.Errorf("Expected default model 55, got %s", tools.model)
	}

	// Test with nil config
	_, err = NewTools(nil, webservices.NewResolver())
	if err == nil {
		t.Error("NewTools should return error with nil config")
	}

	// Test with invalid config
	invalidConfig := &common.Config{
		TpAmb: types.Homologation,
		CNPJ:  "invalid",
	}

	_, err = NewTools(invalidConfig, webservices.NewResolver())
	if err == nil {
		t.Error("NewTools should return error with invalid config")
	}
}

func TestSetModel(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, _ := NewTools(config, webservices.NewResolver())

	// Test valid models
	err := tools.SetModel("55")
	if err != nil {
		t.Errorf("SetModel should not return error for NFe model: %v", err)
	}

	if tools.GetModel() != "55" {
		t.Errorf("Expected model 55, got %s", tools.GetModel())
	}

	err = tools.SetModel("65")
	if err != nil {
		t.Errorf("SetModel should not return error for NFCe model: %v", err)
	}

	if tools.GetModel() != "65" {
		t.Errorf("Expected model 65, got %s", tools.GetModel())
	}

	// Test invalid model
	err = tools.SetModel("99")
	if err == nil {
		t.Error("SetModel should return error for invalid model")
	}
}

func TestSetCertificate(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, _ := NewTools(config, webservices.NewResolver())

	// Test setting certificate
	cert := "test_certificate"
	tools.SetCertificate(cert)

	if tools.certificate != cert {
		t.Error("Certificate should be stored correctly")
	}
}

func TestToolsValidateConfig(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, _ := NewTools(config, webservices.NewResolver())

	err := tools.ValidateConfig()
	if err != nil {
		t.Errorf("ValidateConfig should not return error for valid config: %v", err)
	}
}

func TestSetTimeout(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, _ := NewTools(config, webservices.NewResolver())

	newTimeout := 60 * time.Second
	tools.SetTimeout(newTimeout)

	// This test verifies that SetTimeout doesn't panic
	// Actual timeout verification would require accessing private fields
}

func TestEnableDebug(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, _ := NewTools(config, webservices.NewResolver())

	// Test enabling debug
	tools.EnableDebug(true)
	tools.EnableDebug(false)

	// This test verifies that EnableDebug doesn't panic
	// Actual debug state verification would require accessing private fields
}

func TestGetStateCode(t *testing.T) {
	tests := map[string]string{
		"SP": "35",
		"RJ": "33",
		"MG": "31",
		"RS": "43",
		"PR": "41",
		"SC": "42",
		"XX": "35", // Invalid state should default to SP
	}

	for uf, expectedCode := range tests {
		code := getStateCode(uf)
		if code != expectedCode {
			t.Errorf("getStateCode(%s) = %s, expected %s", uf, code, expectedCode)
		}
	}
}

func TestGenerateLoteId(t *testing.T) {
	id1 := generateLoteId()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateLoteId()

	if id1 == id2 {
		t.Error("generateLoteId should generate different IDs")
	}

	if len(id1) == 0 {
		t.Error("generateLoteId should not return empty string")
	}
}

func TestRequestResponseStructures(t *testing.T) {
	// Test StatusRequest struct
	statusReq := StatusRequest{
		Versao: "4.00",
		TpAmb:  2,
		CUF:    "35",
		XServ:  "STATUS",
	}

	if statusReq.Versao != "4.00" {
		t.Error("StatusRequest should store version correctly")
	}

	// Test LoteNFe struct
	lote := LoteNFe{
		IdLote: "123456789",
		NFes:   make([]NFe, 0),
	}

	if lote.IdLote != "123456789" {
		t.Error("LoteNFe should store ID correctly")
	}

	// Test InutilizacaoRequest struct
	inutReq := InutilizacaoRequest{
		Versao: "4.00",
		InfInut: InfInut{
			TpAmb:  2,
			XServ:  "INUTILIZAR",
			CUF:    "35",
			Ano:    "23",
			CNPJ:   "12345678000195",
			Mod:    "55",
			Serie:  "1",
			NNFIni: "1",
			NNFFin: "10",
			XJust:  "Teste de inutilização",
		},
	}

	if inutReq.Versao != "4.00" {
		t.Error("InutilizacaoRequest should store version correctly")
	}

	if inutReq.InfInut.CNPJ != "12345678000195" {
		t.Error("InutilizacaoRequest should store CNPJ correctly")
	}
}

// Mock test for SefazStatus (would require HTTP mocking in real implementation)
func TestSefazStatusStructure(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, err := NewTools(config, webservices.NewResolver())
	if err != nil {
		t.Fatalf("NewTools failed: %v", err)
	}

	// This test would require HTTP mocking to actually call SefazStatus
	// For now, we just verify the tools instance is ready
	if tools.webservices == nil {
		t.Error("Tools should have webservice manager initialized")
	}

	if tools.soapClient == nil {
		t.Error("Tools should have SOAP client initialized")
	}
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	// Test getStateCode with various inputs
	testCases := []struct {
		input    string
		expected string
	}{
		{"sp", "35"},
		{"SP", "35"},
		{"rj", "33"},
		{"RJ", "33"},
		{"", "35"},        // Default case
		{"INVALID", "35"}, // Default case
	}

	for _, tc := range testCases {
		result := getStateCode(tc.input)
		if result != tc.expected {
			t.Errorf("getStateCode(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

// Test context handling
func TestContextHandling(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, _ := NewTools(config, webservices.NewResolver())

	// Test context creation and cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// This would normally test actual SOAP calls with context
	// For now, we just verify context can be created
	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Verify tools can work with context (structure test)
	if tools == nil {
		t.Error("Tools should be ready to work with context")
	}
}

// Test error handling
func TestToolsErrorHandling(t *testing.T) {
	config := &common.Config{
		TpAmb:       types.Homologation,
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	tools, _ := NewTools(config, webservices.NewResolver())

	// Test SefazConsultaChave with invalid key
	ctx := context.Background()
	_, err := tools.SefazConsultaChave(ctx, "invalid_key")
	if err == nil {
		t.Error("SefazConsultaChave should return error for invalid key")
	}

	// Test SefazConsultaRecibo with empty receipt
	_, err = tools.SefazConsultaRecibo(ctx, "")
	if err == nil {
		t.Error("SefazConsultaRecibo should return error for empty receipt")
	}

	// Test SefazInutiliza with nil request
	_, err = tools.SefazInutiliza(ctx, nil)
	if err == nil {
		t.Error("SefazInutiliza should return error for nil request")
	}

	// Test SefazEvento with invalid chave
	_, err = tools.SefazEvento(ctx, "invalid", EVT_CCE, 1, "", nil, "")
	if err == nil {
		t.Error("SefazEvento should return error for invalid chave")
	}

	// Test SefazCancela with invalid parameters
	_, err = tools.SefazCancela(ctx, "invalid", "", "", nil, "")
	if err == nil {
		t.Error("SefazCancela should return error for invalid parameters")
	}

	// Test SefazCCe with invalid parameters
	_, err = tools.SefazCCe(ctx, "invalid", "", 0, nil, "")
	if err == nil {
		t.Error("SefazCCe should return error for invalid parameters")
	}

	// Test SefazConsultaCadastro with empty parameters
	_, err = tools.SefazConsultaCadastro(ctx, "", "")
	if err == nil {
		t.Error("SefazConsultaCadastro should return error for empty parameters")
	}
}
