package nfe

import (
	"strings"
	"testing"
	"time"
)

func TestNewMake(t *testing.T) {
	make := NewMake()
	
	if make == nil {
		t.Fatal("NewMake should not return nil")
	}
	
	if make.version != LayoutVersion {
		t.Errorf("Expected version %s, got %s", LayoutVersion, make.version)
	}
	
	if make.model != ModelNFe {
		t.Errorf("Expected model %v, got %v", ModelNFe, make.model)
	}
	
	if make.environment != EnvironmentTesting {
		t.Errorf("Expected environment %v, got %v", EnvironmentTesting, make.environment)
	}
	
	if !make.checkGTIN {
		t.Error("Expected checkGTIN to be true by default")
	}
	
	if !make.autoCalculate {
		t.Error("Expected autoCalculate to be true by default")
	}
}

func TestMakeConfiguration(t *testing.T) {
	make := NewMake()
	
	// Test configuration methods
	make.SetVersion("4.00").
		SetModel(ModelNFCe).
		SetEnvironment(EnvironmentProduction).
		SetCheckGTIN(false).
		SetRemoveAccents(true).
		SetRoundValues(false).
		SetAutoCalculate(false)
	
	if make.version != "4.00" {
		t.Errorf("Expected version 4.00, got %s", make.version)
	}
	
	if make.model != ModelNFCe {
		t.Errorf("Expected model NFCe, got %v", make.model)
	}
	
	if make.environment != EnvironmentProduction {
		t.Errorf("Expected production environment, got %v", make.environment)
	}
	
	if make.checkGTIN {
		t.Error("Expected checkGTIN to be false")
	}
	
	if !make.removeAccents {
		t.Error("Expected removeAccents to be true")
	}
	
	if make.roundValues {
		t.Error("Expected roundValues to be false")
	}
	
	if make.autoCalculate {
		t.Error("Expected autoCalculate to be false")
	}
}

func TestTagIde(t *testing.T) {
	make := NewMake()
	
	ide := &Identificacao{
		CUF:      "35",
		NatOp:    "Venda",
		Mod:      "55",
		Serie:    "1",
		NNF:      "123",
		TpNF:     "1",
		IdDest:   "1",
		CMunFG:   "3550308",
		TpImp:    "1",
		TpEmis:   "1",
		FinNFe:   "1",
		IndFinal: "1",
		IndPres:  "1",
		ProcEmi:  "0",
		VerProc:  "SPED-NFE-GO v1.0",
	}
	
	err := make.TagIde(ide)
	if err != nil {
		t.Fatalf("TagIde should not return error: %v", err)
	}
	
	if make.identification == nil {
		t.Fatal("Identification should be set")
	}
	
	if make.identification.CNF == "" {
		t.Error("cNF should be auto-generated")
	}
	
	if len(make.identification.CNF) != 8 {
		t.Errorf("cNF should have 8 digits, got %d", len(make.identification.CNF))
	}
	
	if make.identification.TpAmb == "" {
		t.Error("tpAmb should be set from environment")
	}
}

func TestTagEmit(t *testing.T) {
	make := NewMake()
	
	emit := &Emitente{
		CNPJ:  "12345678000195",
		XNome: "EMPRESA TESTE LTDA",
		EnderEmit: Endereco{
			XLgr:    "RUA TESTE",
			Nro:     "123",
			XBairro: "CENTRO",
			CMun:    "3550308",
			XMun:    "SAO PAULO",
			UF:      "SP",
			CEP:     "01000000",
		},
		IE:  "123456789012",
		CRT: "3",
	}
	
	err := make.TagEmit(emit)
	if err != nil {
		t.Fatalf("TagEmit should not return error: %v", err)
	}
	
	if make.issuer == nil {
		t.Fatal("Issuer should be set")
	}
	
	if make.issuer.XNome != "EMPRESA TESTE LTDA" {
		t.Errorf("Expected company name 'EMPRESA TESTE LTDA', got %s", make.issuer.XNome)
	}
}

func TestTagDet(t *testing.T) {
	make := NewMake()
	
	item := &Item{
		Prod: Produto{
			CProd:      "001",
			CEAN:       "SEM GTIN",
			XProd:      "PRODUTO TESTE",
			NCM:        "12345678",
			CFOP:       "5102",
			UCom:       "UN",
			QCom:       "1.0000",
			VUnCom:     "100.0000",
			VProd:      "100.00",
			CEANTrib:   "SEM GTIN",
			UTrib:      "UN",
			QTrib:      "1.0000",
			VUnTrib:    "100.0000",
			IndTot:     "1",
		},
		Imposto: Imposto{
			ICMS: &ICMS{
				ICMS00: &ICMS00{
					Orig:   "0",
					CST:    "00",
					ModBC:  "0",
					VBC:    "100.00",
					PICMS:  "18.00",
					VICMS:  "18.00",
				},
			},
		},
	}
	
	err := make.TagDet(item)
	if err != nil {
		t.Fatalf("TagDet should not return error: %v", err)
	}
	
	if len(make.items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(make.items))
	}
	
	if make.items[0].NItem != "1" {
		t.Errorf("Expected item number 1, got %s", make.items[0].NItem)
	}
	
	// Check if totals were updated (auto-calculate is enabled by default)
	if make.totals.productValue != 100.00 {
		t.Errorf("Expected product value 100.00, got %f", make.totals.productValue)
	}
}

func TestMakeGenerateAccessKey(t *testing.T) {
	make := NewMake()
	
	// Set identification
	ide := &Identificacao{
		CUF:      "35",
		NatOp:    "Venda",
		Serie:    "1",
		NNF:      "123",
		DhEmi:    "2023-01-15T10:00:00-03:00",
	}
	
	err := make.TagIde(ide)
	if err != nil {
		t.Fatalf("Failed to set identification: %v", err)
	}
	
	// Set issuer
	emit := &Emitente{
		CNPJ:  "12345678000195",
		XNome: "EMPRESA TESTE",
		IE:    "123456789",
		CRT:   "3",
	}
	
	err = make.TagEmit(emit)
	if err != nil {
		t.Fatalf("Failed to set issuer: %v", err)
	}
	
	// Generate access key
	err = make.generateAccessKey()
	if err != nil {
		t.Fatalf("Failed to generate access key: %v", err)
	}
	
	if make.accessKey == nil {
		t.Fatal("Access key should be generated")
	}
	
	key := make.accessKey.GetKey()
	if len(key) != 44 {
		t.Errorf("Access key should have 44 digits, got %d", len(key))
	}
	
	// Validate access key components
	if make.accessKey.State != "35" {
		t.Errorf("Expected state 35, got %s", make.accessKey.State)
	}
	
	if make.accessKey.Document != "12345678000195" {
		t.Errorf("Expected document 12345678000195, got %s", make.accessKey.Document)
	}
	
	if make.accessKey.Model != "55" {
		t.Errorf("Expected model 55, got %s", make.accessKey.Model)
	}
}

func TestCompleteNFeGeneration(t *testing.T) {
	make := NewMake()
	
	// Set identification
	ide := &Identificacao{
		CUF:      "35",
		NatOp:    "Venda",
		Mod:      "55",
		Serie:    "1",
		NNF:      "123",
		TpNF:     "1",
		IdDest:   "1",
		CMunFG:   "3550308",
		TpImp:    "1",
		TpEmis:   "1",
		FinNFe:   "1",
		IndFinal: "1",
		IndPres:  "1",
		ProcEmi:  "0",
		VerProc:  "SPED-NFE-GO v1.0",
	}
	
	err := make.TagIde(ide)
	if err != nil {
		t.Fatalf("Failed to set identification: %v", err)
	}
	
	// Set issuer
	emit := &Emitente{
		CNPJ:  "12345678000195",
		XNome: "EMPRESA TESTE LTDA",
		EnderEmit: Endereco{
			XLgr:    "RUA TESTE",
			Nro:     "123",
			XBairro: "CENTRO",
			CMun:    "3550308",
			XMun:    "SAO PAULO",
			UF:      "SP",
			CEP:     "01000000",
		},
		IE:  "123456789012",
		CRT: "3",
	}
	
	err = make.TagEmit(emit)
	if err != nil {
		t.Fatalf("Failed to set issuer: %v", err)
	}
	
	// Set recipient
	dest := &Destinatario{
		CNPJ:      "98765432000123",
		XNome:     "CLIENTE TESTE",
		IndIEDest: "1",
		IE:        "987654321098",
	}
	
	err = make.TagDest(dest)
	if err != nil {
		t.Fatalf("Failed to set recipient: %v", err)
	}
	
	// Add item
	item := &Item{
		Prod: Produto{
			CProd:      "001",
			CEAN:       "SEM GTIN",
			XProd:      "PRODUTO TESTE",
			NCM:        "12345678",
			CFOP:       "5102",
			UCom:       "UN",
			QCom:       "1.0000",
			VUnCom:     "100.0000",
			VProd:      "100.00",
			CEANTrib:   "SEM GTIN",
			UTrib:      "UN",
			QTrib:      "1.0000",
			VUnTrib:    "100.0000",
			IndTot:     "1",
		},
		Imposto: Imposto{
			ICMS: &ICMS{
				ICMS00: &ICMS00{
					Orig:   "0",
					CST:    "00",
					ModBC:  "0",
					VBC:    "100.00",
					PICMS:  "18.00",
					VICMS:  "18.00",
				},
			},
		},
	}
	
	err = make.TagDet(item)
	if err != nil {
		t.Fatalf("Failed to add item: %v", err)
	}
	
	// Set transport
	transp := &Transporte{
		ModFrete: "0",
	}
	
	err = make.TagTransp(transp)
	if err != nil {
		t.Fatalf("Failed to set transport: %v", err)
	}
	
	// Generate XML
	xml, err := make.GetXML()
	if err != nil {
		t.Fatalf("Failed to generate XML: %v", err)
	}
	
	if xml == "" {
		t.Fatal("Generated XML should not be empty")
	}
	
	// Check XML structure
	if !strings.Contains(xml, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>") {
		t.Error("XML should contain XML declaration")
	}
	
	if !strings.Contains(xml, "<NFe") {
		t.Error("XML should contain NFe element")
	}
	
	if !strings.Contains(xml, "<infNFe") {
		t.Error("XML should contain infNFe element")
	}
	
	if !strings.Contains(xml, "<ide>") {
		t.Error("XML should contain ide element")
	}
	
	if !strings.Contains(xml, "<emit>") {
		t.Error("XML should contain emit element")
	}
	
	if !strings.Contains(xml, "<dest>") {
		t.Error("XML should contain dest element")
	}
	
	if !strings.Contains(xml, "<det") {
		t.Error("XML should contain det element")
	}
	
	if !strings.Contains(xml, "<total>") {
		t.Error("XML should contain total element")
	}
	
	if !strings.Contains(xml, "<transp>") {
		t.Error("XML should contain transp element")
	}
	
	// Get access key
	accessKey := make.GetAccessKey()
	if len(accessKey) != 44 {
		t.Errorf("Access key should have 44 digits, got %d", len(accessKey))
	}
	
	// Check if XML contains the access key
	if !strings.Contains(xml, "NFe"+accessKey) {
		t.Error("XML should contain access key in Id attribute")
	}
}

func TestAccessKeyGenerationUtility(t *testing.T) {
	// Test access key generation
	accessKey, err := GenerateAccessKey("35", "12345678000195", ModelNFe, 1, 123, EmissionNormal)
	if err != nil {
		t.Fatalf("Failed to generate access key: %v", err)
	}
	
	if len(accessKey.GetKey()) != 44 {
		t.Errorf("Access key should have 44 digits, got %d", len(accessKey.GetKey()))
	}
	
	// Test validation
	if !accessKey.IsValid() {
		t.Error("Generated access key should be valid")
	}
	
	// Test parsing
	parsed, err := ParseAccessKey(accessKey.GetKey())
	if err != nil {
		t.Fatalf("Failed to parse access key: %v", err)
	}
	
	if parsed.State != "35" {
		t.Errorf("Expected state 35, got %s", parsed.State)
	}
	
	if parsed.Document != "12345678000195" {
		t.Errorf("Expected document 12345678000195, got %s", parsed.Document)
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test formatting functions
	if FormatCurrency(123.456) != "123.46" {
		t.Errorf("Expected 123.46, got %s", FormatCurrency(123.456))
	}
	
	if FormatQuantity(1.23456) != "1.2346" {
		t.Errorf("Expected 1.2346, got %s", FormatQuantity(1.23456))
	}
	
	// Test validation functions
	if !ValidateCNPJ("11444777000161") {
		t.Error("Valid CNPJ should pass validation")
	}
	
	if ValidateCNPJ("12345678901234") {
		t.Error("Invalid CNPJ should fail validation")
	}
	
	// Test normalization
	normalized := NormalizeString("  Ação com açúcar  ", true, 20)
	expected := "ACAO COM ACUCAR"
	if normalized != expected {
		t.Errorf("Expected '%s', got '%s'", expected, normalized)
	}
	
	// Test date formatting
	testTime := time.Date(2023, 1, 15, 10, 30, 45, 0, time.UTC)
	formatted := FormatDateTime(testTime)
	if !strings.Contains(formatted, "2023-01-15T10:30:45") {
		t.Errorf("Unexpected date format: %s", formatted)
	}
}

func TestErrorHandling(t *testing.T) {
	make := NewMake()
	
	// Test validation errors
	err := make.TagIde(nil)
	if err == nil {
		t.Error("TagIde should return error for nil identification")
	}
	
	err = make.TagEmit(nil)
	if err == nil {
		t.Error("TagEmit should return error for nil issuer")
	}
	
	err = make.TagDet(nil)
	if err == nil {
		t.Error("TagDet should return error for nil item")
	}
	
	// Test incomplete NFe generation
	_, err = make.GetXML()
	if err == nil {
		t.Error("GetXML should return error for incomplete NFe")
	}
}