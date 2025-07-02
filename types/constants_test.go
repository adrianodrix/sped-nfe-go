package types

import (
	"testing"
)

func TestEnvironmentString(t *testing.T) {
	tests := []struct {
		env      Environment
		expected string
	}{
		{Production, "Production"},
		{Homologation, "Homologation"},
		{Environment(99), "Unknown"},
	}
	
	for _, test := range tests {
		result := test.env.String()
		if result != test.expected {
			t.Errorf("Environment.String() for %v: expected '%s', got '%s'", test.env, test.expected, result)
		}
	}
}

func TestEnvironmentIsValid(t *testing.T) {
	tests := []struct {
		env      Environment
		expected bool
	}{
		{Production, true},
		{Homologation, true},
		{Environment(0), false},
		{Environment(3), false},
		{Environment(99), false},
	}
	
	for _, test := range tests {
		result := test.env.IsValid()
		if result != test.expected {
			t.Errorf("Environment.IsValid() for %v: expected %t, got %t", test.env, test.expected, result)
		}
	}
}

func TestUFString(t *testing.T) {
	tests := []struct {
		uf       UF
		expected string
	}{
		{SP, "SP"},
		{RJ, "RJ"},
		{MG, "MG"},
		{AC, "AC"},
		{EX, "EX"},
		{UF(100), "Unknown"},
	}
	
	for _, test := range tests {
		result := test.uf.String()
		if result != test.expected {
			t.Errorf("UF.String() for %v: expected '%s', got '%s'", test.uf, test.expected, result)
		}
	}
}

func TestUFIsValid(t *testing.T) {
	validUFs := []UF{
		AC, AL, AP, AM, BA, CE, DF, ES, GO, MA, MT, MS, MG,
		PA, PB, PR, PE, PI, RJ, RN, RS, RO, RR, SC, SP, SE, TO, EX,
	}
	
	// Test valid UFs
	for _, uf := range validUFs {
		if !uf.IsValid() {
			t.Errorf("UF %v (%s) should be valid", uf, uf.String())
		}
	}
	
	// Test invalid UFs
	invalidUFs := []UF{UF(0), UF(1), UF(98), UF(100)}
	for _, uf := range invalidUFs {
		if uf.IsValid() {
			t.Errorf("UF %v should be invalid", uf)
		}
	}
}

func TestModeloNFeString(t *testing.T) {
	tests := []struct {
		modelo   ModeloNFe
		expected string
	}{
		{ModeloNFe55, "NFe"},
		{ModeloNFCe65, "NFCe"},
		{ModeloNFe(99), "Unknown"},
	}
	
	for _, test := range tests {
		result := test.modelo.String()
		if result != test.expected {
			t.Errorf("ModeloNFe.String() for %v: expected '%s', got '%s'", test.modelo, test.expected, result)
		}
	}
}

func TestVersaoLayoutConstants(t *testing.T) {
	if Versao310 != "3.10" {
		t.Errorf("Expected Versao310 to be '3.10', got '%s'", Versao310)
	}
	
	if Versao400 != "4.00" {
		t.Errorf("Expected Versao400 to be '4.00', got '%s'", Versao400)
	}
}

func TestEventoConstants(t *testing.T) {
	tests := []struct {
		evento   TipoEvento
		expected int
	}{
		{EvtConfirmacao, 210200},
		{EvtCiencia, 210210},
		{EvtDesconhecimento, 210220},
		{EvtCCe, 110110},
		{EvtCancela, 110111},
		{EvtEPEC, 110140},
	}
	
	for _, test := range tests {
		if int(test.evento) != test.expected {
			t.Errorf("Expected evento %v to be %d, got %d", test.evento, test.expected, int(test.evento))
		}
	}
}

func TestDefaultConstants(t *testing.T) {
	if ChaveAcessoLength != 44 {
		t.Errorf("Expected ChaveAcessoLength to be 44, got %d", ChaveAcessoLength)
	}
	
	if DefaultTimeoutSeconds != 30 {
		t.Errorf("Expected DefaultTimeoutSeconds to be 30, got %d", DefaultTimeoutSeconds)
	}
	
	if MinTimeoutSeconds != 5 {
		t.Errorf("Expected MinTimeoutSeconds to be 5, got %d", MinTimeoutSeconds)
	}
	
	if MaxTimeoutSeconds != 300 {
		t.Errorf("Expected MaxTimeoutSeconds to be 300, got %d", MaxTimeoutSeconds)
	}
}

func TestTipoEmissaoConstants(t *testing.T) {
	if TeNormal != 1 {
		t.Errorf("Expected TeNormal to be 1, got %d", TeNormal)
	}
	
	if TeContingenciaFS != 2 {
		t.Errorf("Expected TeContingenciaFS to be 2, got %d", TeContingenciaFS)
	}
}

func TestTipoAmbienteConstants(t *testing.T) {
	if TaProducao != 1 {
		t.Errorf("Expected TaProducao to be 1, got %d", TaProducao)
	}
	
	if TaHomologacao != 2 {
		t.Errorf("Expected TaHomologacao to be 2, got %d", TaHomologacao)
	}
}

func TestAllStatesAreCovered(t *testing.T) {
	// Test that we have all 26 states + DF + EX
	expectedCount := 28
	actualCount := 0
	
	// Count valid UFs
	for i := 0; i < 100; i++ {
		if UF(i).IsValid() {
			actualCount++
		}
	}
	
	if actualCount != expectedCount {
		t.Errorf("Expected %d valid UFs, found %d", expectedCount, actualCount)
	}
}