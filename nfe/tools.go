// Package nfe provides the main tools for NFe communication with SEFAZ webservices.
package nfe

import (
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/certificate"
	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/soap"
	"github.com/adrianodrix/sped-nfe-go/types"
)

// Type alias for Certificate interface
type Certificate = certificate.Certificate

// GetStatusServiceInfo returns the webservice info for NFe status service using the resolver interface
func (t *Tools) GetStatusServiceInfo() (common.WebServiceInfo, error) {
	uf := strings.ToUpper(t.config.SiglaUF)
	isProduction := t.config.TpAmb == types.Production

	// Use the resolver interface to get webservice information
	return t.resolver.GetStatusServiceURL(uf, isProduction, t.model)
}

// GetAuthorizationServiceInfo returns the webservice info for NFe authorization service using the resolver interface  
func (t *Tools) GetAuthorizationServiceInfo() (common.WebServiceInfo, error) {
	uf := strings.ToUpper(t.config.SiglaUF)
	isProduction := t.config.TpAmb == types.Production

	// Check if resolver supports authorization service (extended interface)
	if extResolver, ok := t.resolver.(interface{
		GetAuthorizationServiceURL(uf string, isProduction bool, model string) (common.WebServiceInfo, error)
	}); ok {
		return extResolver.GetAuthorizationServiceURL(uf, isProduction, t.model)
	}

	// Fallback to old webservices system for backward compatibility
	env := common.Environment(t.config.TpAmb)
	return t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeAutorizacao, env, t.model)
}

// getInutilizacaoServiceInfo returns the webservice info for NFe inutilization service using the resolver interface
func (t *Tools) getInutilizacaoServiceInfo() (common.WebServiceInfo, error) {
	uf := strings.ToUpper(t.config.SiglaUF)
	isProduction := t.config.TpAmb == types.Production

	// Check if resolver supports inutilização service (extended interface)
	if extResolver, ok := t.resolver.(interface{
		GetInutilizacaoServiceURL(uf string, isProduction bool, model string) (common.WebServiceInfo, error)
	}); ok {
		return extResolver.GetInutilizacaoServiceURL(uf, isProduction, t.model)
	}

	// Fallback to old webservices system for backward compatibility
	env := common.Environment(t.config.TpAmb)
	return t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeInutilizacao, env, t.model)
}

// GetLastRequest returns the last SOAP request sent for debugging
func (t *Tools) GetLastRequest() string {
	return t.lastRequest
}

// GetLastResponse returns the last SOAP response received for debugging
func (t *Tools) GetLastResponse() string {
	return t.lastResponse
}

// Tools provides the main interface for NFe operations with SEFAZ
type Tools struct {
	config       *common.Config
	webservices  *common.WebServiceManager
	resolver     common.WebserviceResolver // Interface for webservice URL resolution
	soapClient   *soap.SOAPClient
	certificate  interface{} // Will be properly typed when certificate package is ready
	model        string      // NFe model (55 or 65)
	lastRequest  string      // Last SOAP request sent
	lastResponse string      // Last SOAP response received
}

// NewTools creates a new Tools instance for NFe operations
func NewTools(config *common.Config, resolver common.WebserviceResolver) (*Tools, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if resolver == nil {
		return nil, fmt.Errorf("webservice resolver cannot be nil")
	}

	if err := common.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	// Create SOAP client
	soapConfig := &soap.SOAPClientConfig{
		Timeout:       time.Duration(config.Timeout) * time.Second,
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		UserAgent:     "sped-nfe-go/1.0",
		EnableLogging: false,
	}

	// Check for unsafe SSL environment variable (for testing only)
	if os.Getenv("SPED_NFE_UNSAFE_SSL") == "true" {
		soapConfig.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	soapClient := soap.NewSOAPClient(soapConfig)

	// Create webservice manager
	wsManager := common.NewWebServiceManager()

	return &Tools{
		config:      config,
		webservices: wsManager,
		resolver:    resolver,
		soapClient:  soapClient,
		model:       "55", // Default to NFe
	}, nil
}

// SetModel sets the document model (55 for NFe, 65 for NFCe)
func (t *Tools) SetModel(model string) error {
	if model != "55" && model != "65" {
		return fmt.Errorf("invalid model: %s (must be 55 or 65)", model)
	}
	t.model = model
	return nil
}

// GetModel returns the current document model
func (t *Tools) GetModel() string {
	return t.model
}

// SetCertificate sets the digital certificate for requests and configures SSL/TLS authentication
func (t *Tools) SetCertificate(certificate interface{}) error {
	t.certificate = certificate

	// Configure SSL/TLS client certificate authentication in SOAP client
	if cert, ok := certificate.(Certificate); ok && cert != nil {
		if err := t.soapClient.LoadCertificate(cert); err != nil {
			return fmt.Errorf("failed to configure SSL certificate in SOAP client: %v", err)
		}
	}

	return nil
}


// Status Service Operations

// SefazStatus checks the status of SEFAZ webservice
func (t *Tools) SefazStatus(ctx context.Context) (*StatusResponse, error) {
	// Build status request
	statusRequest := &StatusRequest{
		Xmlns:  "http://www.portalfiscal.inf.br/nfe",
		Versao: t.config.Versao,
		TpAmb:  int(t.config.TpAmb),
		CUF:    getStateCode(t.config.SiglaUF),
		XServ:  "STATUS",
	}

	// Convert to XML
	requestXML, err := xml.Marshal(statusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status request: %v", err)
	}

	// Get webservice info
	serviceInfo, err := t.GetStatusServiceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL: %v", err)
	}

	// Create SOAP request
	soapReq, err := soap.CreateNFeSOAPRequest(serviceInfo.URL, serviceInfo.Action, string(requestXML))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %v", err)
	}

	// Store request for debugging
	t.lastRequest = soapReq.Body

	// Send request
	soapResp, err := t.soapClient.Call(ctx, soapReq)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %v", err)
	}

	// Store response for debugging
	t.lastResponse = soapResp.Body

	// Extract body content
	bodyContent, err := soap.ExtractBodyContent(soapResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract body content: %v", err)
	}

	// Parse response - handle both direct and wrapped responses
	var statusResponse StatusResponse

	// Try direct parsing first
	if err := xml.Unmarshal([]byte(bodyContent), &statusResponse); err != nil {
		// If that fails, try parsing the wrapped response
		var wrappedResponse struct {
			XMLName xml.Name       `xml:"nfeResultMsg"`
			Result  StatusResponse `xml:"retConsStatServ"`
		}
		if err2 := xml.Unmarshal([]byte(bodyContent), &wrappedResponse); err2 != nil {
			return nil, fmt.Errorf("failed to unmarshal status response: %v (also tried wrapped format: %v)", err, err2)
		}
		statusResponse = wrappedResponse.Result
	}

	return &statusResponse, nil
}

// Authorization Service Operations

// SefazEnviaLote sends a batch of NFe for authorization
func (t *Tools) SefazEnviaLote(ctx context.Context, lote *LoteNFe, sincrono bool) (*EnvioLoteResponse, error) {
	if lote == nil {
		return nil, fmt.Errorf("lote cannot be nil")
	}

	// Set synchronous mode
	indSinc := "0"
	if sincrono {
		indSinc = "1"
	}

	// Build authorization request
	envioLote := &EnvioLoteRequest{
		Xmlns:   "http://www.portalfiscal.inf.br/nfe",
		Versao:  t.config.Versao,
		IdLote:  lote.IdLote,
		IndSinc: indSinc,
		NFes:    lote.NFes,
	}

	// Convert to XML
	requestXML, err := xml.Marshal(envioLote)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal envio lote request: %v", err)
	}
	fmt.Printf("requestXML: %s\n", string(requestXML))

	return t.sefazEnviaLoteInternal(ctx, requestXML)
}

// SefazEnviaLoteSignedXML sends a batch with pre-signed NFe XML strings
func (t *Tools) SefazEnviaLoteSignedXML(ctx context.Context, idLote string, signedNFeXMLs []string, sincrono bool) (*EnvioLoteResponse, error) {
	if len(signedNFeXMLs) == 0 {
		return nil, fmt.Errorf("no NFe XMLs provided")
	}

	// Set synchronous mode per official manual (AP03a field)
	indSinc := "0"
	if sincrono {
		indSinc = "1"
	}

	// Build XML manually to preserve signatures following official manual format
	requestXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>`)
	requestXML += fmt.Sprintf(`<enviNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="%s">`, t.config.Versao)
	requestXML += fmt.Sprintf(`<idLote>%s</idLote>`, idLote)
	requestXML += fmt.Sprintf(`<indSinc>%s</indSinc>`, indSinc)

	// Insert each signed NFe XML directly (remove only XML declaration like PHP does)
	for _, nfeXML := range signedNFeXMLs {
		// Remove XML declaration if present (same as PHP preg_replace("/<\?xml.*?\?>/", "", $xml))
		cleanedXML := nfeXML
		if strings.HasPrefix(cleanedXML, "<?xml") {
			if idx := strings.Index(cleanedXML, "?>"); idx >= 0 {
				cleanedXML = strings.TrimSpace(cleanedXML[idx+2:])
			}
		}

		// Insert the cleaned XML directly (preserving all namespaces like PHP)
		requestXML += cleanedXML
	}

	requestXML += `</enviNFe>`

	fmt.Printf("requestXML: %s\n", requestXML)

	return t.sefazEnviaLoteInternal(ctx, []byte(requestXML))
}

// sefazEnviaLoteInternal handles the actual webservice call
func (t *Tools) sefazEnviaLoteInternal(ctx context.Context, requestXML []byte) (*EnvioLoteResponse, error) {

	// Get webservice info using resolver for consistency with QueryStatus
	serviceInfo, err := t.GetAuthorizationServiceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL: %v", err)
	}

	// Create SOAP request
	soapReq, err := soap.CreateNFeSOAPRequest(serviceInfo.URL, serviceInfo.Action, string(requestXML))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %v", err)
	}

	// Store request for debugging
	t.lastRequest = soapReq.Body

	// Debug SOAP request details
	fmt.Printf("🔍 SOAP Request Details:\n")
	fmt.Printf("   URL: %s\n", serviceInfo.URL)
	fmt.Printf("   Action: %s\n", serviceInfo.Action)
	if len(soapReq.Body) > 500 {
		fmt.Printf("   Body (primeiros 500 chars): %s...\n", string(soapReq.Body)[:500])
	} else {
		fmt.Printf("   Body: %s\n", string(soapReq.Body))
	}

	// Send request
	soapResp, err := t.soapClient.Call(ctx, soapReq)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %v", err)
	}

	// Store response for debugging
	t.lastResponse = soapResp.Body

	fmt.Printf("🔍 SOAP Response Details:\n")
	fmt.Printf("   Status: %d\n", soapResp.StatusCode)
	fmt.Printf("   Headers: %v\n", soapResp.Headers)
	if len(soapResp.Body) > 0 {
		fmt.Printf("   Body: %s\n", string(soapResp.Body))
	} else {
		fmt.Printf("   Body: [VAZIO]\n")
	}

	// Extract body content
	bodyContent, err := soap.ExtractBodyContent(soapResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract body content: %v", err)
	}

	// Parse response - try direct parsing first
	var envioResponse EnvioLoteResponse
	if err := xml.Unmarshal([]byte(bodyContent), &envioResponse); err != nil {
		// If that fails, try parsing the wrapped response (nfeResultMsg format)
		var wrappedResponse struct {
			XMLName xml.Name          `xml:"nfeResultMsg"`
			Result  EnvioLoteResponse `xml:"retEnviNFe"`
		}
		if err2 := xml.Unmarshal([]byte(bodyContent), &wrappedResponse); err2 != nil {
			return nil, fmt.Errorf("failed to unmarshal envio lote response: %v (also tried wrapped format: %v)", err, err2)
		}
		envioResponse = wrappedResponse.Result
	}

	return &envioResponse, nil
}

// SefazConsultaRecibo queries the processing result of a batch by receipt number
func (t *Tools) SefazConsultaRecibo(ctx context.Context, nRec string) (*ConsultaReciboResponse, error) {
	if nRec == "" {
		return nil, fmt.Errorf("receipt number cannot be empty")
	}

	// Build consultation request
	consultaRecibo := &ConsultaReciboRequest{
		Xmlns:  "http://www.portalfiscal.inf.br/nfe",
		Versao: t.config.Versao,
		TpAmb:  int(t.config.TpAmb),
		NRec:   nRec,
	}

	// Convert to XML
	requestXML, err := xml.Marshal(consultaRecibo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal consulta recibo request: %v", err)
	}

	// Get webservice info
	env := common.Environment(t.config.TpAmb)
	serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeRetAutorizacao, env, t.model)
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL: %v", err)
	}

	// Create SOAP request
	soapReq, err := soap.CreateNFeSOAPRequest(serviceInfo.URL, serviceInfo.Action, string(requestXML))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %v", err)
	}

	// Store request for debugging
	t.lastRequest = soapReq.Body

	// Send request
	soapResp, err := t.soapClient.Call(ctx, soapReq)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %v", err)
	}

	// Store response for debugging
	t.lastResponse = soapResp.Body

	// Extract body content
	bodyContent, err := soap.ExtractBodyContent(soapResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract body content: %v", err)
	}

	// Parse response
	var consultaResponse ConsultaReciboResponse
	if err := xml.Unmarshal([]byte(bodyContent), &consultaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consulta recibo response: %v", err)
	}

	return &consultaResponse, nil
}

// Query Service Operations

// SefazConsultaChave queries an NFe by its access key
func (t *Tools) SefazConsultaChave(ctx context.Context, chave string) (*ConsultaChaveResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("access key must have 44 digits")
	}

	// Build consultation request
	consultaChave := &ConsultaChaveRequest{
		Xmlns:  "http://www.portalfiscal.inf.br/nfe",
		Versao: t.config.Versao,
		TpAmb:  int(t.config.TpAmb),
		XServ:  "CONSULTAR",
		ChNFe:  chave,
	}

	// Convert to XML
	requestXML, err := xml.Marshal(consultaChave)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal consulta chave request: %v", err)
	}

	// Get webservice info
	env := common.Environment(t.config.TpAmb)
	serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeConsultaProtocolo, env, t.model)
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL: %v", err)
	}

	// Create SOAP request
	soapReq, err := soap.CreateNFeSOAPRequest(serviceInfo.URL, serviceInfo.Action, string(requestXML))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %v", err)
	}

	// Store request for debugging
	t.lastRequest = soapReq.Body

	// Send request
	soapResp, err := t.soapClient.Call(ctx, soapReq)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %v", err)
	}

	// Store response for debugging
	t.lastResponse = soapResp.Body

	// Extract body content
	bodyContent, err := soap.ExtractBodyContent(soapResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract body content: %v", err)
	}

	// Parse response
	var consultaResponse ConsultaChaveResponse
	if err := xml.Unmarshal([]byte(bodyContent), &consultaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consulta chave response: %v", err)
	}

	return &consultaResponse, nil
}

// Invalidation Service Operations

// SefazInutilizaNumeros invalidates a range of NFe numbers with automatic validation and signing
func (t *Tools) SefazInutilizaNumeros(ctx context.Context, nSerie, nIni, nFin int, xJust string, ano ...string) (*InutilizacaoResponse, error) {
	// 1. Validate input parameters
	if err := ValidateInutilizacaoParams(nSerie, nIni, nFin, xJust); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}
	
	// 2. Determine year (default to current year last 2 digits)
	anoStr := fmt.Sprintf("%02d", time.Now().Year()%100)
	if len(ano) > 0 && ano[0] != "" {
		anoStr = ano[0]
		if len(anoStr) != 2 {
			return nil, fmt.Errorf("ano deve ter 2 dígitos, informado: %s", anoStr)
		}
	}
	
	// 3. Get UF code
	cUF := getStateCode(t.config.SiglaUF)
	if cUF == "" {
		return nil, fmt.Errorf("código UF não encontrado para: %s", t.config.SiglaUF)
	}
	
	// 4. Determine document type (CNPJ vs CPF for MT)
	documento := t.config.CNPJ
	isCPF := t.config.SiglaUF == "MT" && len(documento) == 11
	
	// 5. Generate unique ID
	idInut := GenerateInutilizacaoId(cUF, anoStr, documento, t.model, nSerie, nIni, nFin, isCPF)
	
	// 6. Create request structure
	request := &InutilizacaoRequest{
		XMLName: xml.Name{Local: "inutNFe"},
		Xmlns:   "http://www.portalfiscal.inf.br/nfe",
		Versao:  t.config.Versao,
		InfInut: InfInut{
			Id:     idInut,
			TpAmb:  int(t.config.TpAmb),
			XServ:  "INUTILIZAR",
			CUF:    cUF,
			Ano:    anoStr,
			Mod:    t.model,
			Serie:  fmt.Sprintf("%d", nSerie),
			NNFIni: fmt.Sprintf("%d", nIni),
			NNFFin: fmt.Sprintf("%d", nFin),
			XJust:  xJust,
		},
	}
	
	// 7. Set document (CNPJ or CPF)
	if isCPF {
		request.InfInut.CPF = documento
	} else {
		request.InfInut.CNPJ = documento
	}
	
	// 8. Sign the XML (placeholder - to be implemented)
	// TODO: Implement XML signing with certificate
	
	// 9. Call the base function
	return t.SefazInutiliza(ctx, request)
}

// SefazInutiliza invalidates a range of NFe numbers
func (t *Tools) SefazInutiliza(ctx context.Context, inutilizacao *InutilizacaoRequest) (*InutilizacaoResponse, error) {
	if inutilizacao == nil {
		return nil, fmt.Errorf("inutilizacao request cannot be nil")
	}

	// Set common fields
	inutilizacao.Versao = t.config.Versao
	inutilizacao.InfInut.TpAmb = int(t.config.TpAmb)

	// Convert to XML
	requestXML, err := xml.Marshal(inutilizacao)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inutilizacao request: %v", err)
	}

	// Get webservice info using resolver for consistency with authorization
	serviceInfo, err := t.getInutilizacaoServiceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL: %v", err)
	}

	// Create SOAP request
	soapReq, err := soap.CreateNFeSOAPRequest(serviceInfo.URL, serviceInfo.Action, string(requestXML))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %v", err)
	}

	// Store request for debugging
	t.lastRequest = soapReq.Body

	// Send request
	soapResp, err := t.soapClient.Call(ctx, soapReq)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %v", err)
	}

	// Store response for debugging
	t.lastResponse = soapResp.Body

	// Extract body content
	bodyContent, err := soap.ExtractBodyContent(soapResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract body content: %v", err)
	}

	// Parse response
	var inutResponse InutilizacaoResponse
	if err := xml.Unmarshal([]byte(bodyContent), &inutResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inutilizacao response: %v", err)
	}

	return &inutResponse, nil
}

// SefazCCe sends a correction letter (carta de correção) event to SEFAZ
func (t *Tools) SefazCCe(ctx context.Context, chave string, xCorrecao string, nSeqEvento int, dhEvento *time.Time, lote string) (*EventResponseNFe, error) {
	// Create and validate CCe request
	req := &CCeRequest{
		ChaveNFe:  chave,
		Correcao:  xCorrecao,
		Sequencia: nSeqEvento,
		DhEvento:  dhEvento,
		Lote:      lote,
	}

	if err := ValidarCCe(req); err != nil {
		return nil, fmt.Errorf("CCe validation failed: %v", err)
	}

	// Create additional XML tags for CCe
	tagAdic := CreateCCeTagAdic(req.Correcao, req.XCondUso)

	return t.SefazEvento(ctx, req.ChaveNFe, CCeTypeEvent, req.Sequencia, tagAdic, req.DhEvento, req.Lote)
}

// SefazCancela sends a cancellation event to SEFAZ
func (t *Tools) SefazCancela(ctx context.Context, chave string, xJust string, nProt string, dhEvento *time.Time, lote string) (*EventResponseNFe, error) {
	if chave == "" || xJust == "" || nProt == "" {
		return nil, fmt.Errorf("cancellation: chave, xJust or nProt is empty")
	}

	// Validate and clean justification text
	xJust = strings.TrimSpace(xJust)
	if len(xJust) > 255 {
		xJust = xJust[:255]
	}

	tagAdic := fmt.Sprintf("<nProt>%s</nProt><xJust>%s</xJust>", nProt, xJust)

	return t.SefazEvento(ctx, chave, EVT_CANCELA, 1, tagAdic, dhEvento, lote)
}

// SefazCancelaPorSubstituicao sends a cancellation by substitution event to SEFAZ
func (t *Tools) SefazCancelaPorSubstituicao(ctx context.Context, chave string, xJust string, nProt string, chNFeRef string, verAplic string, dhEvento *time.Time, lote string) (*EventResponseNFe, error) {
	if chave == "" || xJust == "" || nProt == "" || chNFeRef == "" {
		return nil, fmt.Errorf("cancellation by substitution: chave, xJust, nProt or chNFeRef is empty")
	}

	// Validate and clean justification text
	xJust = strings.TrimSpace(xJust)
	if len(xJust) > 255 {
		xJust = xJust[:255]
	}

	tagAdic := fmt.Sprintf("<nProt>%s</nProt><xJust>%s</xJust><chNFeRef>%s</chNFeRef><verAplic>%s</verAplic>", nProt, xJust, chNFeRef, verAplic)

	return t.SefazEvento(ctx, chave, EVT_CANCELASUBSTITUICAO, 1, tagAdic, dhEvento, lote)
}

// SefazManifesta sends a manifestation event to SEFAZ
func (t *Tools) SefazManifesta(ctx context.Context, chave string, tpEvento int, xJust string, nSeqEvento int, dhEvento *time.Time, lote string) (*EventResponseNFe, error) {
	if chave == "" {
		return nil, fmt.Errorf("manifestation: chave is empty")
	}

	// Validate event type for manifestation
	validTypes := []int{EVT_CIENCIA, EVT_CONFIRMACAO, EVT_DESCONHECIMENTO, EVT_NAO_REALIZADA}
	valid := false
	for _, validType := range validTypes {
		if tpEvento == validType {
			valid = true
			break
		}
	}
	if !valid {
		return nil, fmt.Errorf("invalid event type for manifestation: %d", tpEvento)
	}

	// Some manifestation events require justification
	if (tpEvento == EVT_DESCONHECIMENTO || tpEvento == EVT_NAO_REALIZADA) && xJust == "" {
		return nil, fmt.Errorf("manifestation event %d requires justification", tpEvento)
	}

	var tagAdic string
	if xJust != "" {
		// Validate and clean justification text
		xJust = strings.TrimSpace(xJust)
		if len(xJust) > 255 {
			xJust = xJust[:255]
		}
		tagAdic = fmt.Sprintf("<xJust>%s</xJust>", xJust)
	}

	return t.SefazEvento(ctx, chave, tpEvento, nSeqEvento, tagAdic, dhEvento, lote)
}

// SefazEvento is the generic function for sending events to SEFAZ
func (t *Tools) SefazEvento(ctx context.Context, chave string, tpEvento int, nSeqEvento int, tagAdic string, dhEvento *time.Time, lote string) (*EventResponseNFe, error) {
	if chave == "" {
		return nil, fmt.Errorf("chave is required")
	}
	if len(chave) != 44 {
		return nil, fmt.Errorf("chave must be 44 characters long")
	}

	// Validate NFe key and extract UF
	uf, err := t.extractUFFromChave(chave)
	if err != nil {
		return nil, fmt.Errorf("invalid chave: %v", err)
	}

	// Get event info
	eventInfo, err := GetEventInfo(tpEvento)
	if err != nil {
		return nil, err
	}

	// Set default values
	if nSeqEvento == 0 {
		nSeqEvento = 1
	}
	if lote == "" {
		lote = strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Create event parameters
	params := EventParams{
		UF:         uf,
		ChNFe:      chave,
		TpEvento:   tpEvento,
		NSeqEvento: nSeqEvento,
		TagAdic:    tagAdic,
		DhEvento:   dhEvento,
		Lote:       lote,
		CNPJ:       t.config.CNPJ,
		TpAmb:      strconv.Itoa(int(t.config.TpAmb)),
		VerEvento:  eventInfo.Version,
	}

	// Create event XML
	eventXML, err := CreateEventXML(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create event XML: %v", err)
	}

	// Convert to XML
	requestXML, err := xml.Marshal(eventXML)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event request: %v", err)
	}

	// Get webservice info
	env := common.Environment(t.config.TpAmb)
	serviceInfo, err := t.webservices.GetServiceURL(uf, common.NFeRecepcaoEvento, env, t.model)
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL: %v", err)
	}

	// Create SOAP request
	soapReq, err := soap.CreateNFeSOAPRequest(serviceInfo.URL, serviceInfo.Action, string(requestXML))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %v", err)
	}

	// Store request for debugging
	t.lastRequest = soapReq.Body

	// Send request
	soapResp, err := t.soapClient.Call(ctx, soapReq)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %v", err)
	}

	// Store response for debugging
	t.lastResponse = soapResp.Body

	// Extract body content
	bodyContent, err := soap.ExtractBodyContent(soapResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract body content: %v", err)
	}

	// Parse response
	var eventResponse EventResponseNFe
	if err := xml.Unmarshal([]byte(bodyContent), &eventResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event response: %v", err)
	}

	return &eventResponse, nil
}

// extractUFFromChave extracts the UF from the NFe access key
func (t *Tools) extractUFFromChave(chave string) (string, error) {
	if len(chave) != 44 {
		return "", fmt.Errorf("invalid chave length: %d", len(chave))
	}

	// The UF code is in positions 0-1 of the access key
	ufCode := chave[0:2]

	// Map UF codes to UF names
	ufMap := map[string]string{
		"12": "AC", "27": "AL", "16": "AP", "23": "AM", "29": "BA", "85": "CE", "53": "DF",
		"32": "ES", "52": "GO", "21": "MA", "51": "MT", "50": "MS", "31": "MG", "15": "PA",
		"25": "PB", "41": "PR", "26": "PE", "22": "PI", "33": "RJ", "20": "RN", "43": "RS",
		"11": "RO", "14": "RR", "42": "SC", "35": "SP", "28": "SE", "17": "TO",
	}

	if uf, exists := ufMap[ufCode]; exists {
		return uf, nil
	}

	return "", fmt.Errorf("invalid UF code in chave: %s", ufCode)
}

// Registry Service Operations

// SefazConsultaCadastro queries registry information
func (t *Tools) SefazConsultaCadastro(ctx context.Context, documento, uf string) (*ConsultaCadastroResponse, error) {
	if documento == "" {
		return nil, fmt.Errorf("document cannot be empty")
	}

	if uf == "" {
		return nil, fmt.Errorf("UF cannot be empty")
	}

	// Build consultation request
	consultaCadastro := &ConsultaCadastroRequest{
		Versao: "2.00", // Registry consultation always uses version 2.00
		InfCons: InfCons{
			XServ: "CONS-CAD",
			UF:    uf,
		},
	}

	// Set document type based on length
	if len(OnlyNumbers(documento)) == 11 {
		consultaCadastro.InfCons.CPF = OnlyNumbers(documento)
	} else if len(OnlyNumbers(documento)) == 14 {
		consultaCadastro.InfCons.CNPJ = OnlyNumbers(documento)
	} else {
		// Assume it's IE (state registration)
		consultaCadastro.InfCons.IE = OnlyNumbers(documento)
	}

	// Convert to XML
	requestXML, err := xml.Marshal(consultaCadastro)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal consulta cadastro request: %v", err)
	}

	// Get webservice info
	env := common.Environment(t.config.TpAmb)
	serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeConsultaCadastro, env, t.model)
	if err != nil {
		return nil, fmt.Errorf("failed to get service URL: %v", err)
	}

	// Create SOAP request
	soapReq, err := soap.CreateNFeSOAPRequest(serviceInfo.URL, serviceInfo.Action, string(requestXML))
	if err != nil {
		return nil, fmt.Errorf("failed to create SOAP request: %v", err)
	}

	// Store request for debugging
	t.lastRequest = soapReq.Body

	// Send request
	soapResp, err := t.soapClient.Call(ctx, soapReq)
	if err != nil {
		return nil, fmt.Errorf("SOAP call failed: %v", err)
	}

	// Store response for debugging
	t.lastResponse = soapResp.Body

	// Extract body content
	bodyContent, err := soap.ExtractBodyContent(soapResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract body content: %v", err)
	}

	// Parse response
	var consultaResponse ConsultaCadastroResponse
	if err := xml.Unmarshal([]byte(bodyContent), &consultaResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal consulta cadastro response: %v", err)
	}

	return &consultaResponse, nil
}

// Utility Methods

// ValidateConfig validates the current configuration
func (t *Tools) ValidateConfig() error {
	return common.ValidateConfig(t.config)
}

// SetTimeout updates the SOAP client timeout
func (t *Tools) SetTimeout(timeout time.Duration) {
	t.soapClient.SetTimeout(timeout)
}

// EnableDebug enables debug logging
func (t *Tools) EnableDebug(enable bool) {
	t.soapClient.EnableLogging(enable)
}

// Helper functions

// getStateCode returns the IBGE code for a state
func getStateCode(uf string) string {
	stateCodes := map[string]string{
		"AC": "12", "AL": "17", "AP": "16", "AM": "13", "BA": "29",
		"CE": "23", "DF": "53", "ES": "32", "GO": "52", "MA": "21",
		"MT": "51", "MS": "50", "MG": "31", "PA": "15", "PB": "25",
		"PR": "41", "PE": "26", "PI": "22", "RJ": "33", "RN": "24",
		"RS": "43", "RO": "11", "RR": "14", "SC": "42", "SP": "35",
		"SE": "28", "TO": "27",
	}

	if code, exists := stateCodes[strings.ToUpper(uf)]; exists {
		return code
	}
	return "35" // Default to SP
}

// generateLoteId generates a batch ID based on current time
func generateLoteId() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// Define request/response structures

// StatusRequest represents a status service request
type StatusRequest struct {
	XMLName xml.Name `xml:"consStatServ"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	TpAmb   int      `xml:"tpAmb"`
	CUF     string   `xml:"cUF"`
	XServ   string   `xml:"xServ"`
}

// StatusResponse represents a status service response
type StatusResponse struct {
	XMLName      xml.Name `xml:"retConsStatServ"`
	Versao       string   `xml:"versao,attr"`
	TpAmb        int      `xml:"tpAmb"`
	VerAplic     string   `xml:"verAplic"`
	CStat        string   `xml:"cStat"`
	XMotivo      string   `xml:"xMotivo"`
	CUF          string   `xml:"cUF"`
	DhRecbto     string   `xml:"dhRecbto"`
	TMedResposta string   `xml:"tMedResposta,omitempty"`
	DhRetorno    string   `xml:"dhRetorno,omitempty"`
	XObs         string   `xml:"xObs,omitempty"`
}

// LoteNFe represents a batch of NFe documents
type LoteNFe struct {
	IdLote string `xml:"idLote,attr"`
	NFes   []NFe  `xml:"NFe"`
}

// EnvioLoteRequest represents an authorization batch request
type EnvioLoteRequest struct {
	XMLName xml.Name `xml:"enviNFe"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	IdLote  string   `xml:"idLote"`
	IndSinc string   `xml:"indSinc"`
	NFes    []NFe    `xml:"NFe"`
}

// EnvioLoteRequestRaw represents an authorization batch request with raw XML NFes
type EnvioLoteRequestRaw struct {
	XMLName xml.Name `xml:"enviNFe"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	IdLote  string   `xml:"idLote"`
	IndSinc string   `xml:"indSinc"`
	NFesXML []string `xml:"-"` // Raw XML strings will be inserted manually
}

// EnvioLoteResponse represents an authorization batch response
type EnvioLoteResponse struct {
	XMLName  xml.Name  `xml:"retEnviNFe"`
	Versao   string    `xml:"versao,attr"`
	TpAmb    int       `xml:"tpAmb"`
	VerAplic string    `xml:"verAplic"`
	CStat    string    `xml:"cStat"`
	XMotivo  string    `xml:"xMotivo"`
	CUF      string    `xml:"cUF,omitempty"`
	DhRecbto string    `xml:"dhRecbto,omitempty"`
	InfRec   *InfRec   `xml:"infRec,omitempty"`
	ProtNFe  []ProtNFe `xml:"protNFe,omitempty"`
}

// InfRec represents receipt information
type InfRec struct {
	NRec         string `xml:"nRec"`
	TMedResposta string `xml:"tMedResposta,omitempty"`
}

// ProcNFe represents the complete NFe with authorization protocol
type ProcNFe struct {
	XMLName xml.Name `xml:"nfeProc"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	NFe     NFe      `xml:"NFe"`
	ProtNFe ProtNFe  `xml:"protNFe"`
}

// ProtNFe represents the NFe authorization protocol
type ProtNFe struct {
	XMLName xml.Name `xml:"protNFe"`
	Versao  string   `xml:"versao,attr"`
	InfProt InfProt  `xml:"infProt"`
}

// InfProt represents protocol information
type InfProt struct {
	TpAmb    int    `xml:"tpAmb"`
	VerAplic string `xml:"verAplic"`
	ChNFe    string `xml:"chNFe"`
	DhRecbto string `xml:"dhRecbto"`
	NProt    string `xml:"nProt,omitempty"`
	DigVal   string `xml:"digVal,omitempty"`
	CStat    string `xml:"cStat"`
	XMotivo  string `xml:"xMotivo"`
}

// ConsultaReciboRequest represents a receipt consultation request
type ConsultaReciboRequest struct {
	XMLName xml.Name `xml:"consReciNFe"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	TpAmb   int      `xml:"tpAmb"`
	NRec    string   `xml:"nRec"`
}

// ConsultaReciboResponse represents a receipt consultation response
type ConsultaReciboResponse struct {
	XMLName  xml.Name  `xml:"retConsReciNFe"`
	Versao   string    `xml:"versao,attr"`
	TpAmb    int       `xml:"tpAmb"`
	VerAplic string    `xml:"verAplic"`
	NRec     string    `xml:"nRec"`
	CStat    string    `xml:"cStat"`
	XMotivo  string    `xml:"xMotivo"`
	CUF      string    `xml:"cUF,omitempty"`
	ProtNFe  []ProtNFe `xml:"protNFe,omitempty"`
}

// ConsultaChaveRequest represents an access key consultation request
type ConsultaChaveRequest struct {
	XMLName xml.Name `xml:"consSitNFe"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	TpAmb   int      `xml:"tpAmb"`
	XServ   string   `xml:"xServ"`
	ChNFe   string   `xml:"chNFe"`
}

// ConsultaChaveResponse represents an access key consultation response
type ConsultaChaveResponse struct {
	XMLName    xml.Name    `xml:"retConsSitNFe"`
	Versao     string      `xml:"versao,attr"`
	TpAmb      int         `xml:"tpAmb"`
	VerAplic   string      `xml:"verAplic"`
	CStat      string      `xml:"cStat"`
	XMotivo    string      `xml:"xMotivo"`
	CUF        string      `xml:"cUF,omitempty"`
	DhRecbto   string      `xml:"dhRecbto,omitempty"`
	ChNFe      string      `xml:"chNFe,omitempty"`
	ProtNFe    *ProtNFe    `xml:"protNFe,omitempty"`
	RetCancNFe *RetCancNFe `xml:"retCancNFe,omitempty"`
}

// RetCancNFe represents cancellation information
type RetCancNFe struct {
	InfCanc InfCanc `xml:"infCanc"`
}

// InfCanc represents cancellation details
type InfCanc struct {
	TpAmb    int    `xml:"tpAmb"`
	VerAplic string `xml:"verAplic"`
	ChNFe    string `xml:"chNFe"`
	DhRecbto string `xml:"dhRecbto"`
	NProt    string `xml:"nProt"`
	CStat    string `xml:"cStat"`
	XMotivo  string `xml:"xMotivo"`
}

// InutilizacaoRequest represents a number invalidation request
type InutilizacaoRequest struct {
	XMLName xml.Name `xml:"inutNFe"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	InfInut InfInut  `xml:"infInut"`
}

// InfInut represents invalidation information
type InfInut struct {
	Id     string `xml:"Id,attr" validate:"required,len=43"`
	TpAmb  int    `xml:"tpAmb" validate:"required,oneof=1 2"`
	XServ  string `xml:"xServ" validate:"required,eq=INUTILIZAR"`
	CUF    string `xml:"cUF" validate:"required,len=2"`
	Ano    string `xml:"ano" validate:"required,len=2"`
	CNPJ   string `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`
	CPF    string `xml:"CPF,omitempty" validate:"omitempty,len=11"`
	Mod    string `xml:"mod" validate:"required,oneof=55 65"`
	Serie  string `xml:"serie" validate:"required,min=0,max=999"`
	NNFIni string `xml:"nNFIni" validate:"required,min=1,max=999999999"`
	NNFFin string `xml:"nNFFin" validate:"required,min=1,max=999999999"`
	XJust  string `xml:"xJust" validate:"required,min=15,max=255"`
}

// InutilizacaoResponse represents a number invalidation response
type InutilizacaoResponse struct {
	XMLName xml.Name   `xml:"retInutNFe"`
	Versao  string     `xml:"versao,attr"`
	InfInut InfInutRet `xml:"infInut"`
}

// InfInutRet represents invalidation response information
type InfInutRet struct {
	TpAmb    int    `xml:"tpAmb"`
	VerAplic string `xml:"verAplic"`
	CStat    string `xml:"cStat"`
	XMotivo  string `xml:"xMotivo"`
	CUF      string `xml:"cUF"`
	Ano      string `xml:"ano,omitempty"`
	CNPJ     string `xml:"CNPJ,omitempty"`
	CPF      string `xml:"CPF,omitempty"`
	Mod      string `xml:"mod,omitempty"`
	Serie    string `xml:"serie,omitempty"`
	NNFIni   string `xml:"nNFIni,omitempty"`
	NNFFin   string `xml:"nNFFin,omitempty"`
	DhRecbto string `xml:"dhRecbto"`
	NProt    string `xml:"nProt,omitempty"`
}

// IsSuccess returns true if the inutilização was successful
func (r *InfInutRet) IsSuccess() bool {
	return r.CStat == "102"
}

// GetMessage returns a user-friendly message based on the status code
func (r *InfInutRet) GetMessage() string {
	switch r.CStat {
	case "102":
		return "Inutilização de número homologado"
	case "215":
		return "CNPJ do emitente inválido"
	case "216":
		return "CPF do emitente inválido"
	case "217":
		return "Inscrição Estadual do emitente inválida"
	case "252":
		return "Ambiente informado diverge do ambiente solicitado"
	case "401":
		return "CPF do emitente não cadastrado"
	case "402":
		return "CNPJ do emitente não cadastrado"
	default:
		return r.XMotivo
	}
}

// ValidateInutilizacaoParams validates inutilização parameters
func ValidateInutilizacaoParams(nSerie, nIni, nFin int, xJust string) error {
	if nSerie < 0 || nSerie > 999 {
		return fmt.Errorf("série deve estar entre 0 e 999, informado: %d", nSerie)
	}
	
	if nIni <= 0 || nIni > 999999999 {
		return fmt.Errorf("número inicial deve estar entre 1 e 999999999, informado: %d", nIni)
	}
	
	if nFin <= 0 || nFin > 999999999 {
		return fmt.Errorf("número final deve estar entre 1 e 999999999, informado: %d", nFin)
	}
	
	if nFin < nIni {
		return fmt.Errorf("número final (%d) deve ser maior ou igual ao inicial (%d)", nFin, nIni)
	}
	
	if len(xJust) < 15 {
		return fmt.Errorf("justificativa deve ter pelo menos 15 caracteres, informado: %d", len(xJust))
	}
	
	if len(xJust) > 255 {
		return fmt.Errorf("justificativa deve ter no máximo 255 caracteres, informado: %d", len(xJust))
	}
	
	return nil
}

// GenerateInutilizacaoId generates the ID for inutilização request
func GenerateInutilizacaoId(cUF, ano, documento, modelo string, serie, nIni, nFin int, isCPF bool) string {
	var docPadded string
	if isCPF {
		docPadded = fmt.Sprintf("%011s", documento)
	} else {
		docPadded = fmt.Sprintf("%014s", documento)
	}
	
	serieStr := fmt.Sprintf("%03d", serie)
	nIniStr := fmt.Sprintf("%09d", nIni)
	nFinStr := fmt.Sprintf("%09d", nFin)
	
	return fmt.Sprintf("ID%s%s%s%s%s%s%s", 
		cUF, ano, docPadded, modelo, serieStr, nIniStr, nFinStr)
}

// EventoRequest represents an event request
type EventoRequest struct {
	XMLName xml.Name `xml:"envEvento"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	IdLote  string   `xml:"idLote"`
	Evento  []Evento `xml:"evento"`
}

// Evento represents an event
type Evento struct {
	XMLName   xml.Name  `xml:"evento"`
	Versao    string    `xml:"versao,attr"`
	InfEvento InfEvento `xml:"infEvento"`
}

// InfEvento represents event information
type InfEvento struct {
	COrgao     string    `xml:"cOrgao"`
	TpAmb      int       `xml:"tpAmb"`
	CNPJ       string    `xml:"CNPJ,omitempty"`
	CPF        string    `xml:"CPF,omitempty"`
	ChNFe      string    `xml:"chNFe"`
	DhEvento   string    `xml:"dhEvento"`
	TpEvento   string    `xml:"tpEvento"`
	NSeqEvento string    `xml:"nSeqEvento"`
	VerEvento  string    `xml:"verEvento"`
	DetEvento  DetEvento `xml:"detEvento"`
}

// DetEvento represents event details
type DetEvento struct {
	Versao     string `xml:"versao,attr"`
	DescEvento string `xml:"descEvento"`
	NProt      string `xml:"nProt,omitempty"`     // For cancellation
	XJust      string `xml:"xJust,omitempty"`     // For cancellation
	XCorrecao  string `xml:"xCorrecao,omitempty"` // For correction letter
	XCondUso   string `xml:"xCondUso,omitempty"`  // For correction letter
}

// EventoResponse represents an event response
type EventoResponse struct {
	XMLName   xml.Name    `xml:"retEnvEvento"`
	Versao    string      `xml:"versao,attr"`
	IdLote    string      `xml:"idLote"`
	TpAmb     int         `xml:"tpAmb"`
	VerAplic  string      `xml:"verAplic"`
	COrgao    string      `xml:"cOrgao"`
	CStat     string      `xml:"cStat"`
	XMotivo   string      `xml:"xMotivo"`
	RetEvento []RetEvento `xml:"retEvento"`
}

// RetEvento represents event return information
type RetEvento struct {
	InfEvento InfEventoRet `xml:"infEvento"`
}

// InfEventoRet represents event return details
type InfEventoRet struct {
	TpAmb       int    `xml:"tpAmb"`
	VerAplic    string `xml:"verAplic"`
	COrgao      string `xml:"cOrgao"`
	CStat       string `xml:"cStat"`
	XMotivo     string `xml:"xMotivo"`
	ChNFe       string `xml:"chNFe"`
	TpEvento    string `xml:"tpEvento"`
	XEvento     string `xml:"xEvento"`
	NSeqEvento  string `xml:"nSeqEvento"`
	CNPJDest    string `xml:"CNPJDest,omitempty"`
	CPFDest     string `xml:"CPFDest,omitempty"`
	EmailDest   string `xml:"emailDest,omitempty"`
	DhRegEvento string `xml:"dhRegEvento"`
	NProt       string `xml:"nProt"`
}

// ConsultaCadastroRequest represents a registry consultation request
type ConsultaCadastroRequest struct {
	XMLName xml.Name `xml:"ConsCad"`
	Xmlns   string   `xml:"xmlns,attr"`
	Versao  string   `xml:"versao,attr"`
	InfCons InfCons  `xml:"infCons"`
}

// InfCons represents consultation information
type InfCons struct {
	XServ string `xml:"xServ"`
	UF    string `xml:"UF"`
	CNPJ  string `xml:"CNPJ,omitempty"`
	CPF   string `xml:"CPF,omitempty"`
	IE    string `xml:"IE,omitempty"`
}

// ConsultaCadastroResponse represents a registry consultation response
type ConsultaCadastroResponse struct {
	XMLName xml.Name   `xml:"retConsCad"`
	Versao  string     `xml:"versao,attr"`
	InfCons InfConsRet `xml:"infCons"`
	InfCad  []InfCad   `xml:"infCad"`
}

// InfConsRet represents consultation return information
type InfConsRet struct {
	VerAplic string `xml:"verAplic"`
	CStat    string `xml:"cStat"`
	XMotivo  string `xml:"xMotivo"`
	UF       string `xml:"UF"`
	IE       string `xml:"IE,omitempty"`
	CNPJ     string `xml:"CNPJ,omitempty"`
	CPF      string `xml:"CPF,omitempty"`
	DhCons   string `xml:"dhCons"`
	CUF      string `xml:"cUF"`
}

// InfCad represents registry information
type InfCad struct {
	IE         string   `xml:"IE"`
	CNPJ       string   `xml:"CNPJ,omitempty"`
	CPF        string   `xml:"CPF,omitempty"`
	UF         string   `xml:"UF"`
	CSit       string   `xml:"cSit"`
	IndCredNFe string   `xml:"indCredNFe"`
	IndCredCTe string   `xml:"indCredCTe"`
	XNome      string   `xml:"xNome"`
	XFant      string   `xml:"xFant,omitempty"`
	XRegApur   string   `xml:"xRegApur,omitempty"`
	CNAE       string   `xml:"CNAE,omitempty"`
	DExcSit    string   `xml:"dExcSit,omitempty"`
	Ender      Endereco `xml:"ender"`
}
