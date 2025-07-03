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

// getStatusServiceInfo returns the webservice info for NFe status service using the resolver interface
func (t *Tools) getStatusServiceInfo() (common.WebServiceInfo, error) {
	uf := strings.ToUpper(t.config.SiglaUF)
	isProduction := t.config.TpAmb == types.Production

	// Use the resolver interface to get webservice information
	return t.resolver.GetStatusServiceURL(uf, isProduction, t.model)
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

// GetLastRequest returns the last SOAP request sent
func (t *Tools) GetLastRequest() string {
	return t.lastRequest
}

// GetLastResponse returns the last SOAP response received
func (t *Tools) GetLastResponse() string {
	return t.lastResponse
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
	serviceInfo, err := t.getStatusServiceInfo()
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

	// Get webservice info
	env := common.Environment(t.config.TpAmb)
	serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeAutorizacao, env, t.model)
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

	// Get webservice info
	env := common.Environment(t.config.TpAmb)
	serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeInutilizacao, env, t.model)
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

// Event Service Operations

// SefazEvento sends a generic event to SEFAZ
func (t *Tools) SefazEvento(ctx context.Context, evento *EventoRequest) (*EventoResponse, error) {
	if evento == nil {
		return nil, fmt.Errorf("evento request cannot be nil")
	}

	// Set common fields
	evento.Versao = t.config.Versao

	// Convert to XML
	requestXML, err := xml.Marshal(evento)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal evento request: %v", err)
	}

	// Get webservice info
	env := common.Environment(t.config.TpAmb)
	serviceInfo, err := t.webservices.GetServiceURL(t.config.SiglaUF, common.NFeRecepcaoEvento, env, t.model)
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
	var eventoResponse EventoResponse
	if err := xml.Unmarshal([]byte(bodyContent), &eventoResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal evento response: %v", err)
	}

	return &eventoResponse, nil
}

// SefazCancela cancels an NFe (cancellation event)
func (t *Tools) SefazCancela(ctx context.Context, chave, protocolo, justificativa string) (*EventoResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("access key must have 44 digits")
	}

	if protocolo == "" {
		return nil, fmt.Errorf("protocol cannot be empty")
	}

	if len(justificativa) < 15 {
		return nil, fmt.Errorf("justification must have at least 15 characters")
	}

	// Create cancellation event
	evento := &EventoRequest{
		IdLote: generateLoteId(),
		Evento: []Evento{
			{
				InfEvento: InfEvento{
					COrgao:     getStateCode(t.config.SiglaUF),
					TpAmb:      int(t.config.TpAmb),
					CNPJ:       t.config.CNPJ,
					ChNFe:      chave,
					DhEvento:   FormatDateTime(time.Now()),
					TpEvento:   "110111", // Cancellation event type
					NSeqEvento: "1",
					VerEvento:  "1.00",
					DetEvento: DetEvento{
						Versao:     "1.00",
						DescEvento: "Cancelamento",
						NProt:      protocolo,
						XJust:      justificativa,
					},
				},
			},
		},
	}

	return t.SefazEvento(ctx, evento)
}

// SefazCCe sends a correction letter (carta de correção eletrônica)
func (t *Tools) SefazCCe(ctx context.Context, chave, correcao string, sequencia int) (*EventoResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("access key must have 44 digits")
	}

	if len(correcao) < 15 {
		return nil, fmt.Errorf("correction text must have at least 15 characters")
	}

	if sequencia < 1 {
		return nil, fmt.Errorf("sequence must be greater than 0")
	}

	// Create correction letter event
	evento := &EventoRequest{
		IdLote: generateLoteId(),
		Evento: []Evento{
			{
				InfEvento: InfEvento{
					COrgao:     getStateCode(t.config.SiglaUF),
					TpAmb:      int(t.config.TpAmb),
					CNPJ:       t.config.CNPJ,
					ChNFe:      chave,
					DhEvento:   FormatDateTime(time.Now()),
					TpEvento:   "110110", // Correction letter event type
					NSeqEvento: strconv.Itoa(sequencia),
					VerEvento:  "1.00",
					DetEvento: DetEvento{
						Versao:     "1.00",
						DescEvento: "Carta de Correcao Eletronica",
						XCorrecao:  correcao,
						XCondUso:   "A Carta de Correcao e disciplinada pelo paragrafo 1o-A do art. 7o do Convenio S/N, de 15 de dezembro de 1970 e pode ser utilizada para regularizacao de erro ocorrido na emissao de documento fiscal, desde que o erro nao esteja relacionado com: I - as variaveis que determinam o valor do imposto tais como: base de calculo, aliquota, diferenca de preco, quantidade, valor da operacao ou da prestacao; II - a correcao de dados cadastrais que implique mudanca do remetente ou do destinatario; III - a data de emissao ou de saida.",
					},
				},
			},
		},
	}

	return t.SefazEvento(ctx, evento)
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
	TpAmb  int    `xml:"tpAmb"`
	XServ  string `xml:"xServ"`
	CUF    string `xml:"cUF"`
	Ano    string `xml:"ano"`
	CNPJ   string `xml:"CNPJ"`
	Mod    string `xml:"mod"`
	Serie  string `xml:"serie"`
	NNFIni string `xml:"nNFIni"`
	NNFFin string `xml:"nNFFin"`
	XJust  string `xml:"xJust"`
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
	Ano      string `xml:"ano"`
	CNPJ     string `xml:"CNPJ"`
	Mod      string `xml:"mod"`
	Serie    string `xml:"serie"`
	NNFIni   string `xml:"nNFIni"`
	NNFFin   string `xml:"nNFFin"`
	DhRecbto string `xml:"dhRecbto"`
	NProt    string `xml:"nProt,omitempty"`
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
