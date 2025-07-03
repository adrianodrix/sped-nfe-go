package nfe

import (
	"encoding/xml"
	"testing"
)

func TestNFeStructureXMLMarshaling(t *testing.T) {
	// Test basic NFe structure marshaling
	nfe := &NFe{
		Xmlns: "http://www.portalfiscal.inf.br/nfe",
		InfNFe: InfNFe{
			ID:     "NFe35200714200166000187550010000000047000000041",
			Versao: "4.00",
			Ide: Identificacao{
				CUF:      "35",
				CNF:      "00000004",
				NatOp:    "Venda",
				Mod:      "55",
				Serie:    "1",
				NNF:      "47",
				DhEmi:    "2020-07-01T10:00:00-03:00",
				TpNF:     "1",
				IdDest:   "1",
				CMunFG:   "3550308",
				TpImp:    "1",
				TpEmis:   "1",
				CDV:      "1",
				TpAmb:    "2",
				FinNFe:   "1",
				IndFinal: "1",
				IndPres:  "1",
				ProcEmi:  "0",
				VerProc:  "SPED-NFE-GO v1.0",
			},
			Emit: Emitente{
				CNPJ:  "14200166000187",
				XNome: "EMPRESA TESTE LTDA",
				EnderEmit: Endereco{
					XLgr:    "RUA TESTE",
					Nro:     "123",
					XBairro: "CENTRO",
					CMun:    "3550308",
					XMun:    "SAO PAULO",
					UF:      "SP",
					CEP:     "01000000",
					CPais:   "1058",
					XPais:   "BRASIL",
				},
				IE:  "123456789012",
				CRT: "3",
			},
			Det: []Item{
				{
					NItem: "1",
					Prod: Produto{
						CProd:    "001",
						CEAN:     "SEM GTIN",
						XProd:    "PRODUTO TESTE",
						NCM:      "12345678",
						CFOP:     "5102",
						UCom:     "UN",
						QCom:     "1.0000",
						VUnCom:   "100.0000",
						VProd:    "100.00",
						CEANTrib: "SEM GTIN",
						UTrib:    "UN",
						QTrib:    "1.0000",
						VUnTrib:  "100.0000",
						IndTot:   "1",
					},
					Imposto: Imposto{
						ICMS: &ICMS{
							ICMS00: &ICMS00{
								Orig:  "0",
								CST:   "00",
								ModBC: "0",
								VBC:   "100.00",
								PICMS: "18.00",
								VICMS: "18.00",
							},
						},
					},
				},
			},
			Total: Total{
				ICMSTot: ICMSTotal{
					VBC:        "100.00",
					VICMS:      "18.00",
					VICMSDeson: "0.00",
					VFCP:       "0.00",
					VBCST:      "0.00",
					VST:        "0.00",
					VFCPST:     "0.00",
					VFCPSTRet:  "0.00",
					VProd:      "100.00",
					VFrete:     "0.00",
					VSeg:       "0.00",
					VDesc:      "0.00",
					VII:        "0.00",
					VIPI:       "0.00",
					VIPIDevol:  "0.00",
					VPIS:       "0.00",
					VCOFINS:    "0.00",
					VOutro:     "0.00",
					VNF:        "100.00",
				},
			},
			Transp: Transporte{
				ModFrete: "0",
			},
		},
	}

	// Test XML marshaling
	xmlData, err := xml.MarshalIndent(nfe, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal NFe to XML: %v", err)
	}

	// Check if XML contains expected elements
	xmlString := string(xmlData)
	expectedElements := []string{
		"<NFe",
		"<infNFe",
		"<ide>",
		"<emit>",
		"<det",
		"<total>",
		"<transp>",
	}

	for _, element := range expectedElements {
		if !contains(xmlString, element) {
			t.Errorf("Expected XML element %s not found in marshaled XML", element)
		}
	}

	// Test XML unmarshaling
	var unmarshaledNFe NFe
	err = xml.Unmarshal(xmlData, &unmarshaledNFe)
	if err != nil {
		t.Fatalf("Failed to unmarshal XML to NFe: %v", err)
	}

	// Verify some key fields
	if unmarshaledNFe.InfNFe.ID != nfe.InfNFe.ID {
		t.Errorf("Expected ID %s, got %s", nfe.InfNFe.ID, unmarshaledNFe.InfNFe.ID)
	}

	if unmarshaledNFe.InfNFe.Emit.XNome != nfe.InfNFe.Emit.XNome {
		t.Errorf("Expected XNome %s, got %s", nfe.InfNFe.Emit.XNome, unmarshaledNFe.InfNFe.Emit.XNome)
	}
}

func TestPaymentBuilder(t *testing.T) {
	// Test cash payment creation
	cashPayment := NewPaymentBuilder().
		SetType(PaymentTypeMoney).
		SetValue("100.00").
		Build()

	if cashPayment.TPag != "01" {
		t.Errorf("Expected payment type 01, got %s", cashPayment.TPag)
	}

	if cashPayment.VPag != "100.00" {
		t.Errorf("Expected payment value 100.00, got %s", cashPayment.VPag)
	}

	// Test card payment creation
	card := NewCardBuilder().
		SetIntegrationType(IntegrationTypeIntegrated).
		SetBrand(CardBrandVisa).
		SetAuthorizationCode("123456").
		Build()

	cardPayment := NewPaymentBuilder().
		SetType(PaymentTypeCreditCard).
		SetValue("200.00").
		SetCard(card).
		Build()

	if cardPayment.TPag != "03" {
		t.Errorf("Expected payment type 03, got %s", cardPayment.TPag)
	}

	if cardPayment.Card.TBand != "01" {
		t.Errorf("Expected card brand 01, got %s", cardPayment.Card.TBand)
	}
}

func TestTransportBuilder(t *testing.T) {
	// Test simple transport creation
	transport := NewTransportBuilder().
		SetFreightMode(FreightSenderResponsibility).
		Build()

	if transport.ModFrete != "0" {
		t.Errorf("Expected freight mode 0, got %s", transport.ModFrete)
	}

	// Test transport with carrier
	carrier := &Transportador{
		CNPJ:  "12345678000195",
		XNome: "TRANSPORTADORA TESTE",
		UF:    "SP",
	}

	transportWithCarrier := NewTransportBuilder().
		SetFreightMode(FreightReceiverResponsibility).
		SetCarrier(carrier).
		Build()

	if transportWithCarrier.ModFrete != "1" {
		t.Errorf("Expected freight mode 1, got %s", transportWithCarrier.ModFrete)
	}

	if transportWithCarrier.Transporta.XNome != "TRANSPORTADORA TESTE" {
		t.Errorf("Expected carrier name 'TRANSPORTADORA TESTE', got %s", transportWithCarrier.Transporta.XNome)
	}
}

func TestVolumeBuilder(t *testing.T) {
	volume := NewVolumeBuilder().
		SetQuantity("2").
		SetSpecies("CAIXA").
		SetBrand("TESTE").
		SetWeights("10.50", "12.00").
		AddSeal("SEAL001").
		AddSeal("SEAL002").
		Build()

	if volume.QVol != "2" {
		t.Errorf("Expected quantity 2, got %s", volume.QVol)
	}

	if volume.Esp != "CAIXA" {
		t.Errorf("Expected species CAIXA, got %s", volume.Esp)
	}

	if len(volume.Lacres) != 2 {
		t.Errorf("Expected 2 seals, got %d", len(volume.Lacres))
	}

	if volume.Lacres[0].NLacre != "SEAL001" {
		t.Errorf("Expected first seal SEAL001, got %s", volume.Lacres[0].NLacre)
	}
}

func TestTotalCalculator(t *testing.T) {
	calculator := NewTotalCalculator()

	// Add test item
	item := Item{
		Prod: Produto{
			VProd: "100.00",
		},
		Imposto: Imposto{
			ICMS: &ICMS{
				ICMS00: &ICMS00{
					VBC:   "100.00",
					VICMS: "18.00",
				},
			},
		},
	}

	calculator.AddItem(item)

	// Calculate totals
	total := calculator.CalculateICMSTotal()

	if total == nil {
		t.Error("Expected total calculation result, got nil")
	}

	// Note: Actual calculation logic would be implemented in the TODO section
	// This test just verifies the structure works
}

func TestICMSStructures(t *testing.T) {
	// Test ICMS00 structure
	icms00 := &ICMS00{
		Orig:  "0",
		CST:   "00",
		ModBC: "0",
		VBC:   "100.00",
		PICMS: "18.00",
		VICMS: "18.00",
	}

	icms := &ICMS{
		ICMS00: icms00,
	}

	// Test XML marshaling
	xmlData, err := xml.MarshalIndent(icms, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal ICMS to XML: %v", err)
	}

	xmlString := string(xmlData)
	if !contains(xmlString, "<ICMS>") {
		t.Error("Expected <ICMS> element not found")
	}

	if !contains(xmlString, "<ICMS00>") {
		t.Error("Expected <ICMS00> element not found")
	}

	if !contains(xmlString, "<CST>00</CST>") {
		t.Error("Expected CST value not found")
	}
}

func TestPaymentTypes(t *testing.T) {
	// Test payment type descriptions
	if PaymentTypeMoney.Description() != "Dinheiro" {
		t.Errorf("Expected 'Dinheiro', got %s", PaymentTypeMoney.Description())
	}

	if PaymentTypeInstant.Description() != "Pagamento Instantâneo (PIX)" {
		t.Errorf("Expected 'Pagamento Instantâneo (PIX)', got %s", PaymentTypeInstant.Description())
	}

	// Test card brand descriptions
	if CardBrandVisa.Description() != "Visa" {
		t.Errorf("Expected 'Visa', got %s", CardBrandVisa.Description())
	}

	if CardBrandElo.Description() != "Elo" {
		t.Errorf("Expected 'Elo', got %s", CardBrandElo.Description())
	}
}

func TestConvenienceFunctions(t *testing.T) {
	// Test cash payment creation
	cashPayment := CreateCashPayment("50.00")
	if cashPayment.TPag != PaymentTypeMoney.String() {
		t.Errorf("Expected payment type %s, got %s", PaymentTypeMoney.String(), cashPayment.TPag)
	}

	// Test PIX payment creation
	pixPayment := CreatePIXPayment("75.00")
	if pixPayment.TPag != PaymentTypeInstant.String() {
		t.Errorf("Expected payment type %s, got %s", PaymentTypeInstant.String(), pixPayment.TPag)
	}

	// Test simple transport creation
	simpleTransport := CreateSimpleTransport(FreightReceiverResponsibility)
	if simpleTransport.ModFrete != "1" {
		t.Errorf("Expected freight mode 1, got %s", simpleTransport.ModFrete)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
