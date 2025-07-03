// Package common provides webservice URL management for Brazilian SEFAZ services.
package common

import (
	"fmt"
	"strings"
)

// WebServiceInfo represents information about a SEFAZ webservice
type WebServiceInfo struct {
	URL       string `json:"url" xml:"url"`             // Service URL
	Method    string `json:"method" xml:"method"`       // SOAP method name
	Operation string `json:"operation" xml:"operation"` // SOAP operation
	Version   string `json:"version" xml:"version"`     // Service version
	Action    string `json:"action" xml:"action"`       // SOAP action header
}

// UFServices represents services available for a specific state
type UFServices struct {
	Sigla       string                    `json:"sigla" xml:"sigla"`
	Homologacao map[string]WebServiceInfo `json:"homologacao" xml:"homologacao"`
	Producao    map[string]WebServiceInfo `json:"producao" xml:"producao"`
}

// WebServiceType represents the type of webservice
type WebServiceType string

const (
	// NFe Model 55 Services
	NFeStatusServico     WebServiceType = "NfeStatusServico"     // Service status check
	NFeAutorizacao       WebServiceType = "NfeAutorizacao"       // Authorization service
	NFeRetAutorizacao    WebServiceType = "NfeRetAutorizacao"    // Authorization return/query
	NFeConsultaProtocolo WebServiceType = "NfeConsultaProtocolo" // Protocol consultation
	NFeInutilizacao      WebServiceType = "NfeInutilizacao"      // Number invalidation
	NFeConsultaCadastro  WebServiceType = "NfeConsultaCadastro"  // Registry consultation
	NFeRecepcaoEvento    WebServiceType = "NfeRecepcaoEvento"    // Event reception
	NFeDistribuicaoDFe   WebServiceType = "NfeDistribuicaoDFe"   // DFe distribution
	NFeDownloadNF        WebServiceType = "NfeDownloadNF"        // NFe download

	// NFCe Model 65 Services
	NFCeStatusServico     WebServiceType = "NfceStatusServico"     // NFCe service status
	NFCeAutorizacao       WebServiceType = "NfceAutorizacao"       // NFCe authorization
	NFCeRetAutorizacao    WebServiceType = "NfceRetAutorizacao"    // NFCe authorization return
	NFCeConsultaProtocolo WebServiceType = "NfceConsultaProtocolo" // NFCe protocol consultation
	NFCeInutilizacao      WebServiceType = "NfceInutilizacao"      // NFCe number invalidation
	NFCeRecepcaoEvento    WebServiceType = "NfceRecepcaoEvento"    // NFCe event reception

	// Special services
	AN                   WebServiceType = "AN"                   // Ambiente Nacional
	EPEC                 WebServiceType = "EPEC"                 // EPEC service
	CadConsultaCadastro2 WebServiceType = "CadConsultaCadastro2" // Registry consultation v2
)

// Environment represents the service environment
type Environment int

const (
	Production Environment = 1 // Production environment
	Testing    Environment = 2 // Testing/homologation environment
)

// AuthorizerType represents the type of authorizer
type AuthorizerType string

const (
	AuthorizerSEFAZ AuthorizerType = "SEFAZ" // State SEFAZ
	AuthorizerSVRS  AuthorizerType = "SVRS"  // Virtual RS
	AuthorizerSVAN  AuthorizerType = "SVAN"  // Virtual AN
	AuthorizerSVCAN AuthorizerType = "SVCAN" // Virtual CAN
	AuthorizerSVCRS AuthorizerType = "SVCRS" // Virtual CRS
)

// WebServiceManager manages SEFAZ webservice URLs
type WebServiceManager struct {
	services       map[string]UFServices        // Services by state
	authorizers    map[string]map[string]string // Authorizers by model and state
	contingencyMap map[string]string            // Contingency mappings
}

// NewWebServiceManager creates a new webservice manager
func NewWebServiceManager() *WebServiceManager {
	wsm := &WebServiceManager{
		services:       make(map[string]UFServices),
		authorizers:    make(map[string]map[string]string),
		contingencyMap: make(map[string]string),
	}

	// Initialize default authorizers
	wsm.initializeAuthorizers()

	// Initialize default services
	wsm.initializeDefaultServices()

	return wsm
}

// GetServiceURL returns the URL for a specific service
func (wsm *WebServiceManager) GetServiceURL(uf string, service WebServiceType, env Environment, model string) (WebServiceInfo, error) {
	// Get authorizer for the state and model
	authorizer, err := wsm.GetAuthorizer(uf, model)
	if err != nil {
		return WebServiceInfo{}, err
	}

	// Get services for the authorizer
	services, exists := wsm.services[authorizer]
	if !exists {
		return WebServiceInfo{}, fmt.Errorf("no services found for authorizer: %s", authorizer)
	}

	// Select environment
	var envServices map[string]WebServiceInfo
	switch env {
	case Production:
		envServices = services.Producao
	case Testing:
		envServices = services.Homologacao
	default:
		return WebServiceInfo{}, fmt.Errorf("invalid environment: %d", env)
	}

	// Get specific service
	serviceInfo, exists := envServices[string(service)]
	if !exists {
		return WebServiceInfo{}, fmt.Errorf("service %s not found for %s in environment %d", service, authorizer, env)
	}

	return serviceInfo, nil
}

// GetAuthorizer returns the authorizer for a given state and model
func (wsm *WebServiceManager) GetAuthorizer(uf string, model string) (string, error) {
	uf = strings.ToUpper(uf)

	modelMap, exists := wsm.authorizers[model]
	if !exists {
		return "", fmt.Errorf("model %s not supported", model)
	}

	authorizer, exists := modelMap[uf]
	if !exists {
		return "", fmt.Errorf("state %s not supported for model %s", uf, model)
	}

	return authorizer, nil
}

// AddService adds a service to a specific state and environment
func (wsm *WebServiceManager) AddService(uf string, env Environment, service WebServiceType, info WebServiceInfo) {
	if _, exists := wsm.services[uf]; !exists {
		wsm.services[uf] = UFServices{
			Sigla:       uf,
			Homologacao: make(map[string]WebServiceInfo),
			Producao:    make(map[string]WebServiceInfo),
		}
	}

	switch env {
	case Production:
		wsm.services[uf].Producao[string(service)] = info
	case Testing:
		wsm.services[uf].Homologacao[string(service)] = info
	}
}

// IsServiceAvailable checks if a service is available for a state and environment
func (wsm *WebServiceManager) IsServiceAvailable(uf string, service WebServiceType, env Environment, model string) bool {
	_, err := wsm.GetServiceURL(uf, service, env, model)
	return err == nil
}

// GetAvailableServices returns all available services for a state and environment
func (wsm *WebServiceManager) GetAvailableServices(uf string, env Environment, model string) ([]WebServiceType, error) {
	authorizer, err := wsm.GetAuthorizer(uf, model)
	if err != nil {
		return nil, err
	}

	services, exists := wsm.services[authorizer]
	if !exists {
		return nil, fmt.Errorf("no services found for authorizer: %s", authorizer)
	}

	var envServices map[string]WebServiceInfo
	switch env {
	case Production:
		envServices = services.Producao
	case Testing:
		envServices = services.Homologacao
	default:
		return nil, fmt.Errorf("invalid environment: %d", env)
	}

	var availableServices []WebServiceType
	for serviceName := range envServices {
		availableServices = append(availableServices, WebServiceType(serviceName))
	}

	return availableServices, nil
}

// initializeAuthorizers sets up the default authorizer mappings
func (wsm *WebServiceManager) initializeAuthorizers() {
	// NFe Model 55 authorizers
	wsm.authorizers["55"] = map[string]string{
		"AC": "SVRS", "AL": "SVRS", "AP": "SVRS", "AM": "AM", "BA": "BA",
		"CE": "CE", "DF": "SVRS", "ES": "SVRS", "GO": "GO", "MA": "SVAN",
		"MT": "MT", "MS": "MS", "MG": "MG", "PA": "SVAN", "PB": "SVRS",
		"PR": "PR", "PE": "PE", "PI": "SVAN", "RJ": "SVRS", "RN": "SVRS",
		"RS": "RS", "RO": "SVRS", "RR": "SVRS", "SC": "SVRS", "SP": "SP",
		"SE": "SVRS", "TO": "TO",
	}

	// NFCe Model 65 authorizers
	wsm.authorizers["65"] = map[string]string{
		"AC": "SVRS", "AL": "SVRS", "AP": "SVRS", "AM": "AM", "BA": "BA",
		"CE": "CE", "DF": "SVRS", "ES": "SVRS", "GO": "GO", "MA": "SVAN",
		"MT": "MT", "MS": "MS", "MG": "MG", "PA": "SVAN", "PB": "SVRS",
		"PR": "PR", "PE": "PE", "PI": "SVAN", "RJ": "SVRS", "RN": "SVRS",
		"RS": "RS", "RO": "SVRS", "RR": "SVRS", "SC": "SVRS", "SP": "SP",
		"SE": "SVRS", "TO": "TO",
	}
}

// initializeDefaultServices sets up default service configurations for major authorizers
func (wsm *WebServiceManager) initializeDefaultServices() {
	// São Paulo (SP) - Production
	wsm.initializeSPServices()

	// Rio de Janeiro / SVRS - Production
	wsm.initializeSVRSServices()

	// Minas Gerais (MG) - Production
	wsm.initializeMGServices()

	// Rio Grande do Sul (RS) - Production
	wsm.initializeRSServices()

	// Ambiente Nacional (SVAN) - Production
	wsm.initializeSVANServices()
}

// initializeSPServices sets up São Paulo webservices
func (wsm *WebServiceManager) initializeSPServices() {
	sp := UFServices{
		Sigla:       "SP",
		Homologacao: make(map[string]WebServiceInfo),
		Producao:    make(map[string]WebServiceInfo),
	}

	// Homologation services
	sp.Homologacao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	sp.Homologacao[string(NFeAutorizacao)] = WebServiceInfo{
		URL:       "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeautorizacao4.asmx",
		Method:    "nfeAutorizacaoLote",
		Operation: "NFeAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
	}

	sp.Homologacao[string(NFeRetAutorizacao)] = WebServiceInfo{
		URL:       "https://homologacao.nfe.fazenda.sp.gov.br/ws/nferetautorizacao4.asmx",
		Method:    "nfeRetAutorizacaoLote",
		Operation: "NFeRetAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
	}

	sp.Homologacao[string(NFeConsultaProtocolo)] = WebServiceInfo{
		URL:       "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeconsultaprotocolo4.asmx",
		Method:    "nfeConsultaNF",
		Operation: "NFeConsultaProtocolo4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4/nfeConsultaNF",
	}

	sp.Homologacao[string(NFeInutilizacao)] = WebServiceInfo{
		URL:       "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeinutilizacao4.asmx",
		Method:    "nfeInutilizacaoNF",
		Operation: "NFeInutilizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4/nfeInutilizacaoNF",
	}

	sp.Homologacao[string(NFeRecepcaoEvento)] = WebServiceInfo{
		URL:       "https://homologacao.nfe.fazenda.sp.gov.br/ws/nferecepcaoevento4.asmx",
		Method:    "nfeRecepcaoEvento",
		Operation: "NFeRecepcaoEvento4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRecepcaoEvento4/nfeRecepcaoEvento",
	}

	sp.Homologacao[string(NFeConsultaCadastro)] = WebServiceInfo{
		URL:       "https://homologacao.nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx",
		Method:    "consultaCadastro",
		Operation: "CadConsultaCadastro4",
		Version:   "2.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/CadConsultaCadastro4/consultaCadastro",
	}

	// Production services (same structure with production URLs)
	sp.Producao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	sp.Producao[string(NFeAutorizacao)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.sp.gov.br/ws/nfeautorizacao4.asmx",
		Method:    "nfeAutorizacaoLote",
		Operation: "NFeAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
	}

	sp.Producao[string(NFeRetAutorizacao)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.sp.gov.br/ws/nferetautorizacao4.asmx",
		Method:    "nfeRetAutorizacaoLote",
		Operation: "NFeRetAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
	}

	sp.Producao[string(NFeConsultaProtocolo)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.sp.gov.br/ws/nfeconsultaprotocolo4.asmx",
		Method:    "nfeConsultaNF",
		Operation: "NFeConsultaProtocolo4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeConsultaProtocolo4/nfeConsultaNF",
	}

	sp.Producao[string(NFeInutilizacao)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.sp.gov.br/ws/nfeinutilizacao4.asmx",
		Method:    "nfeInutilizacaoNF",
		Operation: "NFeInutilizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeInutilizacao4/nfeInutilizacaoNF",
	}

	sp.Producao[string(NFeRecepcaoEvento)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.sp.gov.br/ws/nferecepcaoevento4.asmx",
		Method:    "nfeRecepcaoEvento",
		Operation: "NFeRecepcaoEvento4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRecepcaoEvento4/nfeRecepcaoEvento",
	}

	sp.Producao[string(NFeConsultaCadastro)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx",
		Method:    "consultaCadastro",
		Operation: "CadConsultaCadastro4",
		Version:   "2.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/CadConsultaCadastro4/consultaCadastro",
	}

	wsm.services["SP"] = sp
}

// initializeSVRSServices sets up SVRS (Virtual Rio Grande do Sul) webservices
func (wsm *WebServiceManager) initializeSVRSServices() {
	svrs := UFServices{
		Sigla:       "SVRS",
		Homologacao: make(map[string]WebServiceInfo),
		Producao:    make(map[string]WebServiceInfo),
	}

	// Homologation services
	svrs.Homologacao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	svrs.Homologacao[string(NFeAutorizacao)] = WebServiceInfo{
		URL:       "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
		Method:    "nfeAutorizacaoLote",
		Operation: "NFeAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
	}

	svrs.Homologacao[string(NFeRetAutorizacao)] = WebServiceInfo{
		URL:       "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
		Method:    "nfeRetAutorizacaoLote",
		Operation: "NFeRetAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
	}

	// Production services
	svrs.Producao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	svrs.Producao[string(NFeAutorizacao)] = WebServiceInfo{
		URL:       "https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
		Method:    "nfeAutorizacaoLote",
		Operation: "NFeAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
	}

	svrs.Producao[string(NFeRetAutorizacao)] = WebServiceInfo{
		URL:       "https://nfe.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
		Method:    "nfeRetAutorizacaoLote",
		Operation: "NFeRetAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeRetAutorizacao4/nfeRetAutorizacaoLote",
	}

	wsm.services["SVRS"] = svrs
}

// initializeMGServices sets up Minas Gerais webservices
func (wsm *WebServiceManager) initializeMGServices() {
	mg := UFServices{
		Sigla:       "MG",
		Homologacao: make(map[string]WebServiceInfo),
		Producao:    make(map[string]WebServiceInfo),
	}

	// Homologation services
	mg.Homologacao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeStatusServico4",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	mg.Homologacao[string(NFeAutorizacao)] = WebServiceInfo{
		URL:       "https://hnfe.fazenda.mg.gov.br/nfe2/services/NFeAutorizacao4",
		Method:    "nfeAutorizacaoLote",
		Operation: "NFeAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
	}

	// Production services
	mg.Producao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.mg.gov.br/nfe2/services/NFeStatusServico4",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	mg.Producao[string(NFeAutorizacao)] = WebServiceInfo{
		URL:       "https://nfe.fazenda.mg.gov.br/nfe2/services/NFeAutorizacao4",
		Method:    "nfeAutorizacaoLote",
		Operation: "NFeAutorizacao4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeAutorizacao4/nfeAutorizacaoLote",
	}

	wsm.services["MG"] = mg
}

// initializeRSServices sets up Rio Grande do Sul webservices
func (wsm *WebServiceManager) initializeRSServices() {
	rs := UFServices{
		Sigla:       "RS",
		Homologacao: make(map[string]WebServiceInfo),
		Producao:    make(map[string]WebServiceInfo),
	}

	// Homologation services
	rs.Homologacao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	// Production services
	rs.Producao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://nfe.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	wsm.services["RS"] = rs
}

// initializeSVANServices sets up SVAN (Sistema Virtual do Ambiente Nacional) webservices
func (wsm *WebServiceManager) initializeSVANServices() {
	svan := UFServices{
		Sigla:       "SVAN",
		Homologacao: make(map[string]WebServiceInfo),
		Producao:    make(map[string]WebServiceInfo),
	}

	// Homologation services
	svan.Homologacao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://hom.sefazvirtual.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	// Production services
	svan.Producao[string(NFeStatusServico)] = WebServiceInfo{
		URL:       "https://www.sefazvirtual.fazenda.gov.br/NFeStatusServico4/NFeStatusServico4.asmx",
		Method:    "nfeStatusServicoNF",
		Operation: "NFeStatusServico4",
		Version:   "4.00",
		Action:    "http://www.portalfiscal.inf.br/nfe/wsdl/NFeStatusServico4/nfeStatusServicoNF",
	}

	wsm.services["SVAN"] = svan
}

// GetContingencyServices returns contingency services for a given state
func (wsm *WebServiceManager) GetContingencyServices(uf string, contingencyType string) (string, error) {
	contingencyMap := map[string]map[string]string{
		"EPEC": {
			"SP": "SP",
			"MG": "MG",
			"RS": "RS",
		},
		"SVCRS": {
			"AC": "SVRS", "AL": "SVRS", "AP": "SVRS", "DF": "SVRS",
			"ES": "SVRS", "PB": "SVRS", "RJ": "SVRS", "RN": "SVRS",
			"RO": "SVRS", "RR": "SVRS", "SC": "SVRS", "SE": "SVRS",
		},
		"SVCAN": {
			"MA": "SVAN", "PA": "SVAN", "PI": "SVAN",
		},
	}

	if typeMap, exists := contingencyMap[contingencyType]; exists {
		if authorizer, exists := typeMap[uf]; exists {
			return authorizer, nil
		}
	}

	return "", fmt.Errorf("contingency service %s not available for state %s", contingencyType, uf)
}

// LoadServicesFromXML loads services from XML configuration (similar to PHP implementation)
func (wsm *WebServiceManager) LoadServicesFromXML(xmlData []byte) error {
	// This would parse XML configuration files like wsnfe_4.00_mod55.xml
	// Implementation would use encoding/xml to parse the service definitions
	// For now, returning nil as the default services are hardcoded above
	return nil
}

// LoadServicesFromJSON loads services from JSON configuration
func (wsm *WebServiceManager) LoadServicesFromJSON(jsonData []byte) error {
	// This would parse JSON configuration files
	// Implementation would unmarshal JSON into the services map
	// For now, returning nil as the default services are hardcoded above
	return nil
}
