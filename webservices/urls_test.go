package webservices

import (
	"testing"

	"github.com/adrianodrix/sped-nfe-go/types"
)

func TestGetAuthorizer(t *testing.T) {
	tests := []struct {
		uf       types.UF
		modelo   types.ModeloNFe
		expected string
		hasError bool
	}{
		{types.SP, types.ModeloNFe55, "SP", false},
		{types.AM, types.ModeloNFe55, "AM", false},
		{types.AC, types.ModeloNFe55, "SVRS", false},
		{types.SP, types.ModeloNFCe65, "SP", false},
		{types.AC, types.ModeloNFCe65, "SVRS", false},
		{types.AN, types.ModeloNFe55, "AN", false},
		{types.UF(999), types.ModeloNFe55, "", true}, // Invalid UF
		{types.SP, types.ModeloNFe(99), "", true},    // Invalid model
	}

	for _, test := range tests {
		result, err := GetAuthorizer(test.uf, test.modelo)
		if test.hasError {
			if err == nil {
				t.Errorf("GetAuthorizer(%s, %d) should return error", test.uf.String(), int(test.modelo))
			}
		} else {
			if err != nil {
				t.Errorf("GetAuthorizer(%s, %d) should not return error, got: %v", test.uf.String(), int(test.modelo), err)
			}
			if result != test.expected {
				t.Errorf("GetAuthorizer(%s, %d) = %s, expected %s", test.uf.String(), int(test.modelo), result, test.expected)
			}
		}
	}
}

func TestGetWebserviceURL(t *testing.T) {
	tests := []struct {
		uf          types.UF
		ambiente    types.Ambiente
		modelo      types.ModeloNFe
		serviceType ServiceType
		hasError    bool
		description string
	}{
		{types.AM, types.AmbienteHomologacao, types.ModeloNFe55, ServiceStatusServico, false, "AM homologacao status"},
		{types.AM, types.AmbienteProducao, types.ModeloNFe55, ServiceAutorizacao, false, "AM producao autorizacao"},
		{types.AN, types.AmbienteHomologacao, types.ModeloNFe55, ServiceDistribuicaoDFe, false, "AN homologacao distribuicao"},
		{types.SVRS, types.AmbienteHomologacao, types.ModeloNFe55, ServiceStatusServico, false, "SVRS homologacao status"},
		{types.UF(999), types.AmbienteHomologacao, types.ModeloNFe55, ServiceStatusServico, true, "invalid UF"},
		{types.AM, types.AmbienteHomologacao, types.ModeloNFe55, ServiceType("InvalidService"), true, "invalid service"},
	}

	for _, test := range tests {
		service, err := GetWebserviceURL(test.uf, test.ambiente, test.modelo, test.serviceType)
		if test.hasError {
			if err == nil {
				t.Errorf("GetWebserviceURL for %s should return error", test.description)
			}
		} else {
			if err != nil {
				t.Errorf("GetWebserviceURL for %s should not return error, got: %v", test.description, err)
			}
			if service == nil {
				t.Errorf("GetWebserviceURL for %s should return service", test.description)
			} else {
				if service.URL == "" {
					t.Errorf("GetWebserviceURL for %s should return service with URL", test.description)
				}
				if service.Method == "" {
					t.Errorf("GetWebserviceURL for %s should return service with Method", test.description)
				}
				if service.Operation == "" {
					t.Errorf("GetWebserviceURL for %s should return service with Operation", test.description)
				}
			}
		}
	}
}

func TestGetAllServices(t *testing.T) {
	tests := []struct {
		uf       types.UF
		ambiente types.Ambiente
		modelo   types.ModeloNFe
		hasError bool
	}{
		{types.AM, types.AmbienteHomologacao, types.ModeloNFe55, false},
		{types.AM, types.AmbienteProducao, types.ModeloNFe55, false},
		{types.SVRS, types.AmbienteHomologacao, types.ModeloNFe55, false},
		{types.UF(999), types.AmbienteHomologacao, types.ModeloNFe55, true},
	}

	for _, test := range tests {
		env, err := GetAllServices(test.uf, test.ambiente, test.modelo)
		if test.hasError {
			if err == nil {
				t.Errorf("GetAllServices(%s, %s, %d) should return error", test.uf.String(), test.ambiente.String(), int(test.modelo))
			}
		} else {
			if err != nil {
				t.Errorf("GetAllServices(%s, %s, %d) should not return error, got: %v", test.uf.String(), test.ambiente.String(), int(test.modelo), err)
			}
			if env == nil {
				t.Errorf("GetAllServices(%s, %s, %d) should return environment", test.uf.String(), test.ambiente.String(), int(test.modelo))
			}
		}
	}
}

func TestIsServiceAvailable(t *testing.T) {
	tests := []struct {
		uf          types.UF
		ambiente    types.Ambiente
		modelo      types.ModeloNFe
		serviceType ServiceType
		expected    bool
	}{
		{types.AM, types.AmbienteHomologacao, types.ModeloNFe55, ServiceStatusServico, true},
		{types.AM, types.AmbienteHomologacao, types.ModeloNFe55, ServiceDistribuicaoDFe, false}, // AM doesn't have this service
		{types.AN, types.AmbienteHomologacao, types.ModeloNFe55, ServiceDistribuicaoDFe, true},  // AN has this service
		{types.SVRS, types.AmbienteHomologacao, types.ModeloNFe55, ServiceStatusServico, true},
		{types.UF(999), types.AmbienteHomologacao, types.ModeloNFe55, ServiceStatusServico, false}, // Invalid UF
	}

	for _, test := range tests {
		result := IsServiceAvailable(test.uf, test.ambiente, test.modelo, test.serviceType)
		if result != test.expected {
			t.Errorf("IsServiceAvailable(%s, %s, %d, %s) = %t, expected %t", 
				test.uf.String(), test.ambiente.String(), int(test.modelo), test.serviceType, result, test.expected)
		}
	}
}

func TestWebserviceConfigToJSON(t *testing.T) {
	config := WebserviceConfig{
		"TEST": &StateWebservices{
			Homologacao: &Environment{
				NfeStatusServico: &Service{
					Method:    "testMethod",
					Operation: "testOperation",
					Version:   "4.00",
					URL:       "https://test.example.com",
				},
			},
		},
	}

	jsonStr, err := config.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() should not return error, got: %v", err)
	}

	if jsonStr == "" {
		t.Error("ToJSON() should return non-empty string")
	}

	// Check if JSON contains expected values
	if !contains(jsonStr, "testMethod") {
		t.Error("JSON should contain testMethod")
	}
	if !contains(jsonStr, "testOperation") {
		t.Error("JSON should contain testOperation")
	}
	if !contains(jsonStr, "https://test.example.com") {
		t.Error("JSON should contain test URL")
	}
}

func TestAuthorizeMapping(t *testing.T) {
	// Test NFe 55 mappings
	nfe55Mapping := AuthorizeMapping[types.ModeloNFe55]
	if len(nfe55Mapping) == 0 {
		t.Error("NFe55 mapping should not be empty")
	}

	// Test specific mappings
	if nfe55Mapping[types.SP] != "SP" {
		t.Errorf("Expected SP to map to SP, got %s", nfe55Mapping[types.SP])
	}

	if nfe55Mapping[types.AM] != "AM" {
		t.Errorf("Expected AM to map to AM, got %s", nfe55Mapping[types.AM])
	}

	if nfe55Mapping[types.AC] != "SVRS" {
		t.Errorf("Expected AC to map to SVRS, got %s", nfe55Mapping[types.AC])
	}

	// Test NFCe 65 mappings
	nfce65Mapping := AuthorizeMapping[types.ModeloNFCe65]
	if len(nfce65Mapping) == 0 {
		t.Error("NFCe65 mapping should not be empty")
	}

	// NFCe should have fewer entries than NFe (no SVAN, AN, etc.)
	if len(nfce65Mapping) >= len(nfe55Mapping) {
		t.Error("NFCe65 mapping should have fewer entries than NFe55")
	}
}

func TestServiceTypes(t *testing.T) {
	serviceTypes := []ServiceType{
		ServiceStatusServico,
		ServiceAutorizacao,
		ServiceConsultaProtocolo,
		ServiceInutilizacao,
		ServiceRetAutorizacao,
		ServiceRecepcaoEvento,
		ServiceConsultaCadastro,
		ServiceDistribuicaoDFe,
		ServiceConsultaDest,
		ServiceDownloadNF,
		ServiceRecepcaoEPEC,
	}

	for _, serviceType := range serviceTypes {
		if string(serviceType) == "" {
			t.Errorf("Service type should not be empty")
		}
	}
}

func TestNFe55ConfigStructure(t *testing.T) {
	// Test that NFe55Config has required states
	requiredStates := []string{"AM", "AN", "BA", "SVRS"}
	
	for _, state := range requiredStates {
		if _, exists := NFe55Config[state]; !exists {
			t.Errorf("NFe55Config should contain state %s", state)
		}
	}

	// Test AM configuration
	amConfig := NFe55Config["AM"]
	if amConfig == nil {
		t.Fatal("AM configuration should not be nil")
	}

	if amConfig.Homologacao == nil {
		t.Error("AM should have homologacao environment")
	}

	if amConfig.Producao == nil {
		t.Error("AM should have producao environment")
	}

	// Test specific service in AM
	if amConfig.Homologacao.NfeStatusServico == nil {
		t.Error("AM homologacao should have NfeStatusServico")
	}

	statusService := amConfig.Homologacao.NfeStatusServico
	if statusService.URL == "" {
		t.Error("NfeStatusServico should have URL")
	}

	if statusService.Method == "" {
		t.Error("NfeStatusServico should have Method")
	}

	if statusService.Operation == "" {
		t.Error("NfeStatusServico should have Operation")
	}

	if statusService.Version == "" {
		t.Error("NfeStatusServico should have Version")
	}
}

func TestGetServiceFromEnvironment(t *testing.T) {
	env := &Environment{
		NfeStatusServico: &Service{
			Method:    "testMethod",
			Operation: "testOperation",
			Version:   "4.00",
			URL:       "https://test.example.com",
		},
	}

	// Test valid service
	service := getServiceFromEnvironment(env, ServiceStatusServico)
	if service == nil {
		t.Error("Should return service for valid service type")
	}

	// Test invalid service
	service = getServiceFromEnvironment(env, ServiceDistribuicaoDFe)
	if service != nil {
		t.Error("Should return nil for unavailable service")
	}

	// Test with nil environment - should handle gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("getServiceFromEnvironment should not panic with nil environment: %v", r)
		}
	}()
	
	service = getServiceFromEnvironment(nil, ServiceStatusServico)
	if service != nil {
		t.Error("Should return nil for nil environment")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr || 
		     containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}