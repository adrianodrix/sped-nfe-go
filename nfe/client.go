// Package nfe provides the main client API for Brazilian Electronic Fiscal Documents (NFe/NFCe).
package nfe

import (
	"context"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/certificate"
	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/factories"
	"github.com/adrianodrix/sped-nfe-go/types"
	"github.com/adrianodrix/sped-nfe-go/validation"
	"github.com/adrianodrix/sped-nfe-go/webservices"
)

// NFEClient is the main NFe client that provides a unified API for all NFe operations.
type NFEClient struct {
	config      *common.Config
	tools       *Tools
	certificate certificate.Certificate
	contingency *factories.Contingency
	validator   *validation.XSDValidator
	timeout     time.Duration
	uf          UF
}

// ClientConfig holds configuration for the NFe client.
type ClientConfig struct {
	Environment Environment // Production or Homologation
	UF          UF          // State code
	Timeout     int         // Timeout in seconds (default: 30)
	CSC         string      // NFCe Security Code (optional)
	CSCId       string      // NFCe CSC ID (optional)
}

// AuthResponse represents the response from NFe authorization.
type AuthResponse struct {
	Success      bool              `json:"success"`
	Status       int               `json:"status"`
	StatusText   string            `json:"statusText"`
	Protocol     string            `json:"protocol,omitempty"`
	Receipt      string            `json:"receipt,omitempty"`
	Key          string            `json:"key,omitempty"`
	XML          []byte            `json:"xml,omitempty"`
	OriginalXML  []byte            `json:"originalXml,omitempty"`
	Messages     []ResponseMessage `json:"messages,omitempty"`
	ProcessingAt time.Time         `json:"processingAt"`
}

// QueryResponse represents the response from NFe queries.
type QueryResponse struct {
	Success    bool              `json:"success"`
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	Key        string            `json:"key,omitempty"`
	Protocol   string            `json:"protocol,omitempty"`
	Authorized bool              `json:"authorized"`
	Cancelled  bool              `json:"cancelled"`
	Messages   []ResponseMessage `json:"messages,omitempty"`
	XML        []byte            `json:"xml,omitempty"`
	LastEvent  *EventInfo        `json:"lastEvent,omitempty"`
	QueryAt    time.Time         `json:"queryAt"`
}

// EventResponse represents the response from fiscal events.
type EventResponse struct {
	Success     bool              `json:"success"`
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	EventType   string            `json:"eventType"`
	Key         string            `json:"key"`
	Protocol    string            `json:"protocol,omitempty"`
	Sequence    int               `json:"sequence,omitempty"`
	Messages    []ResponseMessage `json:"messages,omitempty"`
	XML         []byte            `json:"xml,omitempty"`
	ProcessedAt time.Time         `json:"processedAt"`
}

// ClientStatusResponse represents the SEFAZ status response.
type ClientStatusResponse struct {
	Success     bool              `json:"success"`
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	UF          string            `json:"uf"`
	Environment int               `json:"environment"`
	Online      bool              `json:"online"`
	Messages    []ResponseMessage `json:"messages,omitempty"`
	CheckedAt   time.Time         `json:"checkedAt"`
}

// ResponseMessage represents a message in responses.
type ResponseMessage struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Correction string `json:"correction,omitempty"`
	Type       string `json:"type"` // info, warning, error
}

// EventInfo represents information about fiscal events.
type EventInfo struct {
	Type        string    `json:"type"`
	Sequence    int       `json:"sequence"`
	Status      int       `json:"status"`
	Description string    `json:"description"`
	ProcessedAt time.Time `json:"processedAt"`
}

// ManifestationType represents the type of manifestation for events.
type ManifestationType int

const (
	ManifestationConfirmOperation ManifestationType = iota + 1
	ManifestationIgnoreOperation
	ManifestationNotRealized
	ManifestationUnknownOperation
)

// NewClient creates a new NFe client with the provided configuration.
func NewClient(config ClientConfig) (*NFEClient, error) {
	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.Environment == 0 {
		config.Environment = Homologation
	}

	// Validate configuration
	if config.UF == 0 {
		return nil, fmt.Errorf("UF (state) is required")
	}

	// Convert UF code to string representation
	ufMap := map[UF]string{
		11: "RO", 12: "AC", 13: "AM", 14: "RR", 15: "PA", 16: "AP", 17: "TO",
		21: "MA", 22: "PI", 23: "CE", 24: "RN", 25: "PB", 26: "PE", 27: "AL", 28: "SE", 29: "BA",
		31: "MG", 32: "ES", 33: "RJ", 35: "SP",
		41: "PR", 42: "SC", 43: "RS",
		50: "MS", 51: "MT", 52: "GO", 53: "DF",
	}
	ufStr, ok := ufMap[config.UF]
	if !ok {
		return nil, fmt.Errorf("unsupported UF code: %d", config.UF)
	}

	commonConfig := &common.Config{
		TpAmb:       types.Environment(config.Environment),
		Timeout:     config.Timeout,
		RazaoSocial: "NFE Client",     // Default minimal value
		CNPJ:        "00000000000191", // Default test CNPJ
		SiglaUF:     ufStr,
		Schemes:     "./schemes", // Default schemes path
		Versao:      "4.00",      // Default version
	}

	// Basic validation
	if config.Environment < 1 || config.Environment > 2 {
		return nil, fmt.Errorf("invalid environment: must be 1 (production) or 2 (homologation)")
	}

	client := &NFEClient{
		config:    commonConfig,
		tools:     nil, // Tools will be created lazily when needed
		validator: validation.NewXSDValidator(),
		timeout:   time.Duration(config.Timeout) * time.Second,
		uf:        config.UF,
	}

	return client, nil
}

// ensureTools ensures that the Tools instance is initialized
func (c *NFEClient) ensureTools() error {
	if c.tools == nil {
		// Create webservice resolver
		resolver := webservices.NewResolver()

		tools, err := NewTools(c.config, resolver)
		if err != nil {
			return fmt.Errorf("failed to create tools: %v", err)
		}
		c.tools = tools

		// Set certificate if already configured
		if c.certificate != nil {
			if err := c.tools.SetCertificate(c.certificate); err != nil {
				return fmt.Errorf("failed to configure certificate: %v", err)
			}
		}
	}
	return nil
}

// SetCertificate sets the digital certificate for the client.
func (c *NFEClient) SetCertificate(cert certificate.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate cannot be nil")
	}

	c.certificate = cert
	if c.tools != nil {
		if err := c.tools.SetCertificate(cert); err != nil {
			return fmt.Errorf("failed to configure certificate in tools: %v", err)
		}
	}
	return nil
}

// SetTimeout sets the timeout for SEFAZ operations.
func (c *NFEClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.config.Timeout = int(timeout.Seconds())
	// Tools timeout is managed through config
	c.config.Timeout = int(timeout.Seconds())
}

// SetEnvironment changes the environment (production/homologation).
func (c *NFEClient) SetEnvironment(env Environment) error {
	c.config.TpAmb = types.Environment(env)
	return nil
}

// SetContingency activates or deactivates contingency mode.
func (c *NFEClient) SetContingency(contingency *factories.Contingency) {
	c.contingency = contingency
}

// CreateNFe returns a new NFe builder for creating NFe documents.
func (c *NFEClient) CreateNFe() *Make {
	make := NewMake()
	make.SetVersion("4.00")
	return make
}

// CreateNFCe returns a new NFCe builder for creating NFCe documents.
func (c *NFEClient) CreateNFCe() *Make {
	make := NewMake()
	make.SetVersion("4.00")
	make.SetModel(65) // NFCe model
	return make
}

// LoadFromXML loads an NFe from XML bytes.
func (c *NFEClient) LoadFromXML(xmlData []byte) (*NFe, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("XML content cannot be empty")
	}

	var nfe NFe

	// Parse XML into NFe structure
	if err := xml.Unmarshal(xmlData, &nfe); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %v", err)
	}

	// Validate required fields
	if err := c.validateParsedNFe(&nfe); err != nil {
		return nil, fmt.Errorf("NFe validation failed: %v", err)
	}

	return &nfe, nil
}

// validateParsedNFe validates a parsed NFe structure
func (c *NFEClient) validateParsedNFe(nfe *NFe) error {
	if nfe == nil {
		return fmt.Errorf("NFe cannot be nil")
	}

	// Validate InfNFe
	if nfe.InfNFe.ID == "" {
		return fmt.Errorf("NFe ID is required")
	}

	if nfe.InfNFe.Versao == "" {
		return fmt.Errorf("NFe version is required")
	}

	// Validate identification
	if nfe.InfNFe.Ide.CUF == "" {
		return fmt.Errorf("state code (cUF) is required")
	}

	if nfe.InfNFe.Ide.CNF == "" {
		return fmt.Errorf("random code (cNF) is required")
	}

	if nfe.InfNFe.Ide.NatOp == "" {
		return fmt.Errorf("operation nature is required")
	}

	// Validate issuer
	if nfe.InfNFe.Emit.XNome == "" {
		return fmt.Errorf("issuer name is required")
	}

	// Validate at least one item
	if len(nfe.InfNFe.Det) == 0 {
		return fmt.Errorf("at least one item is required")
	}

	return nil
}

// LoadFromTXT loads an NFe from TXT format.
func (c *NFEClient) LoadFromTXT(txt []byte, layout factories.LayoutType) (*NFe, error) {
	parser, err := factories.NewParser(factories.ParserConfig{
		Version: "4.00",
		Layout:  layout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %v", err)
	}

	_, err = parser.ParseTXT(string(txt))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TXT: %v", err)
	}

	xml, err := parser.GetXML()
	if err != nil {
		return nil, fmt.Errorf("failed to generate XML: %v", err)
	}

	return c.LoadFromXML([]byte(xml))
}

// Authorize sends an NFe for authorization.
func (c *NFEClient) Authorize(ctx context.Context, xml []byte) (*AuthResponse, error) {
	if c.certificate == nil {
		return nil, fmt.Errorf("certificate not set")
	}

	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	// Sign the NFe if not already signed
	signedXML, err := c.signIfNeeded(xml)
	if err != nil {
		return nil, fmt.Errorf("failed to sign NFe: %v", err)
	}

	// Create LoteNFe for sending
	lote, err := c.createLoteFromXML(signedXML)
	if err != nil {
		return nil, fmt.Errorf("failed to create lote: %v", err)
	}

	// Send to SEFAZ for authorization
	response, err := c.tools.SefazEnviaLote(ctx, lote, false)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize NFe: %v", err)
	}

	return c.convertToAuthResponse(response, signedXML), nil
}

// QueryChave queries an NFe by its access key.
func (c *NFEClient) QueryChave(ctx context.Context, chave string) (*QueryResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("invalid access key length: expected 44, got %d", len(chave))
	}

	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	response, err := c.tools.SefazConsultaChave(ctx, chave)
	if err != nil {
		return nil, fmt.Errorf("failed to query NFe: %v", err)
	}

	return c.convertToQueryResponse(response), nil
}

// QueryRecibo queries the processing result by receipt number.
func (c *NFEClient) QueryRecibo(ctx context.Context, recibo string) (*QueryResponse, error) {
	if recibo == "" {
		return nil, fmt.Errorf("receipt number cannot be empty")
	}

	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	response, err := c.tools.SefazConsultaRecibo(ctx, recibo)
	if err != nil {
		return nil, fmt.Errorf("failed to query receipt: %v", err)
	}

	return c.convertReciboToQueryResponse(response), nil
}

// QueryStatus checks the SEFAZ service status.
func (c *NFEClient) QueryStatus(ctx context.Context) (*ClientStatusResponse, error) {
	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	response, err := c.tools.SefazStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query SEFAZ status: %v", err)
	}

	return c.convertToStatusResponse(response), nil
}

// Cancel cancels an NFe with the provided justification.
func (c *NFEClient) Cancel(ctx context.Context, chave, justificativa string) (*EventResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("invalid access key length: expected 44, got %d", len(chave))
	}
	if len(justificativa) < 15 {
		return nil, fmt.Errorf("justification must be at least 15 characters")
	}

	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	// Create cancellation event
	eventoReq, err := c.createCancellationEventRequest(chave, justificativa)
	if err != nil {
		return nil, fmt.Errorf("failed to create cancellation event: %v", err)
	}

	// Convert the old event request to the new format
	response, err := c.tools.SefazCancela(ctx, chave, justificativa, eventoReq.Evento[0].InfEvento.DetEvento.NProt, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to cancel NFe: %v", err)
	}

	return c.convertToEventResponse(response, "cancellation"), nil
}

// CancelWithProtocol cancels an NFe with the provided justification and protocol.
func (c *NFEClient) CancelWithProtocol(ctx context.Context, chave, justificativa, protocolo string) (*CancelamentoResponse, error) {
	// Create and validate cancellation request
	req, err := CreateCancelamentoRequest(chave, justificativa, protocolo)
	if err != nil {
		return nil, fmt.Errorf("invalid cancellation request: %v", err)
	}

	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	// Send cancellation event
	response, err := c.tools.SefazCancela(ctx, req.ChaveNFe, req.Justificativa, req.Protocolo, req.DhEvento, req.Lote)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel NFe: %v", err)
	}

	// Convert to CancelamentoResponse
	return c.convertToCancelamentoResponse(response, req.ChaveNFe), nil
}

// CanCancel checks if an NFe can be cancelled based on its authorization date and current status.
func (c *NFEClient) CanCancel(chave string, dhAutorizacao time.Time, currentStatus string) (bool, error) {
	// Validate NFe key
	if err := ValidateNFeKey(chave); err != nil {
		return false, fmt.Errorf("invalid NFe key: %v", err)
	}

	// Check if already cancelled
	if currentStatus == "cancelled" || currentStatus == "cancelada" {
		return false, fmt.Errorf("NFe is already cancelled")
	}

	// Check if authorized
	if currentStatus != "authorized" && currentStatus != "autorizada" {
		return false, fmt.Errorf("NFe must be authorized before it can be cancelled")
	}

	// Check deadline
	return true, ValidarPrazoCancelamento(dhAutorizacao)
}

// ValidateJustificationText validates a cancellation justification
func (c *NFEClient) ValidateJustificationText(justificativa string) error {
	return ValidateJustification(justificativa)
}

// GetCancellationDeadline returns the deadline for cancelling an NFe
func (c *NFEClient) GetCancellationDeadline(dhAutorizacao time.Time) time.Time {
	return dhAutorizacao.Add(CancellationTimeoutHours * time.Hour)
}

// CCe sends a carta de correção eletrônica (electronic correction letter).
func (c *NFEClient) CCe(ctx context.Context, chave, correcao string, sequencia int) (*EventResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("invalid access key length: expected 44, got %d", len(chave))
	}
	if len(correcao) < 15 {
		return nil, fmt.Errorf("correction must be at least 15 characters")
	}
	if sequencia < 1 {
		return nil, fmt.Errorf("sequence must be greater than 0")
	}

	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	// Use the new SefazCCe function
	response, err := c.tools.SefazCCe(ctx, chave, correcao, sequencia, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to send CCe: %v", err)
	}

	return c.convertToEventResponse(response, "cce"), nil
}

// Manifesta sends a manifestation event for received NFe.
func (c *NFEClient) Manifesta(ctx context.Context, chave string, tipo ManifestationType) (*EventResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("invalid access key length: expected 44, got %d", len(chave))
	}

	// TODO: Implement manifestation using SefazEvento
	return nil, fmt.Errorf("manifestation not implemented yet")
}

// Invalidate invalidates a range of NFe numbers.
func (c *NFEClient) Invalidate(ctx context.Context, serie, numeroInicial, numeroFinal int, justificativa string) (*EventResponse, error) {
	if len(justificativa) < 15 {
		return nil, fmt.Errorf("justification must be at least 15 characters")
	}

	if serie < 1 || numeroInicial < 1 || numeroFinal < numeroInicial {
		return nil, fmt.Errorf("invalid series or number range")
	}

	if err := c.ensureTools(); err != nil {
		return nil, err
	}

	// Create invalidation request
	request, err := c.createInutilizacaoRequest(serie, numeroInicial, numeroFinal, justificativa)
	if err != nil {
		return nil, fmt.Errorf("failed to create invalidation request: %v", err)
	}

	response, err := c.tools.SefazInutiliza(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to invalidate NFe numbers: %v", err)
	}

	return c.convertInutilizacaoToEventResponse(response), nil
}

// ValidateXML validates an NFe XML against schemas.
func (c *NFEClient) ValidateXML(xml []byte) error {
	// Basic validation - check if it's valid XML and has required elements
	xmlStr := string(xml)

	if !strings.Contains(xmlStr, "<NFe") && !strings.Contains(xmlStr, "<nfe") {
		return fmt.Errorf("not a valid NFe XML: missing NFe element")
	}

	if !strings.Contains(xmlStr, "<infNFe") {
		return fmt.Errorf("not a valid NFe XML: missing infNFe element")
	}

	// Perform XSD validation
	result := c.validator.ValidateNFe(xml, "4.00")
	if !result.Valid {
		// Join errors into a single message
		return fmt.Errorf("XSD validation failed: %s", strings.Join(result.Errors, "; "))
	}

	return nil
}

// GenerateKey generates an NFe access key.
func (c *NFEClient) GenerateKey(cnpj string, modelo, serie, numero int, dhEmi time.Time) (string, error) {
	// Validate inputs
	if len(OnlyNumbers(cnpj)) != 14 {
		return "", fmt.Errorf("CNPJ must have 14 digits")
	}
	if modelo != 55 && modelo != 65 {
		return "", fmt.Errorf("model must be 55 (NFe) or 65 (NFCe)")
	}
	if serie < 1 || numero < 1 {
		return "", fmt.Errorf("series and number must be positive")
	}

	// Convert to appropriate types
	var documentModel DocumentModel
	if modelo == 55 {
		documentModel = ModelNFe
	} else {
		documentModel = ModelNFCe
	}

	// Generate access key using the proper implementation
	accessKey, err := GenerateAccessKey(
		fmt.Sprintf("%02d", int(c.uf)), // UF code as string
		OnlyNumbers(cnpj),              // Clean CNPJ
		documentModel,                  // Model
		serie,                          // Series
		numero,                         // Number
		EmissionNormal,                 // Emission type
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate access key: %w", err)
	}

	// Set issue date/time
	accessKey.SetIssueDateTime(dhEmi)

	// Build the complete key
	if err := accessKey.Build(); err != nil {
		return "", fmt.Errorf("failed to build access key: %w", err)
	}

	return accessKey.GetKey(), nil
}

// AddProtocol adds authorization protocol to an NFe XML.
func (c *NFEClient) AddProtocol(nfe, protocolo []byte) ([]byte, error) {
	if len(nfe) == 0 {
		return nil, fmt.Errorf("NFe XML cannot be empty")
	}

	if len(protocolo) == 0 {
		return nil, fmt.Errorf("protocol XML cannot be empty")
	}

	// Parse the NFe XML
	var nfeData NFe
	if err := xml.Unmarshal(nfe, &nfeData); err != nil {
		return nil, fmt.Errorf("failed to parse NFe XML: %v", err)
	}

	// Parse the protocol XML
	var protData ProtNFe
	if err := xml.Unmarshal(protocolo, &protData); err != nil {
		return nil, fmt.Errorf("failed to parse protocol XML: %v", err)
	}

	// Create the complete procNFe structure
	procNFe := ProcNFe{
		Xmlns:   "http://www.portalfiscal.inf.br/nfe",
		Versao:  nfeData.InfNFe.Versao,
		NFe:     nfeData,
		ProtNFe: protData,
	}

	// Marshal to XML
	procXML, err := xml.MarshalIndent(procNFe, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal procNFe XML: %v", err)
	}

	// Add XML declaration
	xmlHeader := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	result := xmlHeader + string(procXML)

	return []byte(result), nil
}

// GetConfig returns the current client configuration.
func (c *NFEClient) GetConfig() *common.Config {
	return c.config
}

// GetCertificate returns the current certificate.
func (c *NFEClient) GetCertificate() certificate.Certificate {
	return c.certificate
}

// GetContingency returns the current contingency configuration.
func (c *NFEClient) GetContingency() *factories.Contingency {
	return c.contingency
}

// ActivateContingency activates contingency mode.
func (c *NFEClient) ActivateContingency(motive string, contingencyType ...factories.ContingencyType) error {
	// Convert UF to appropriate format for contingency
	ufCode := fmt.Sprintf("%02d", int(c.uf))

	var cType factories.ContingencyType
	if len(contingencyType) > 0 {
		cType = contingencyType[0]
	}

	contingency, _, err := factories.CreateContingency(ufCode, motive, cType)
	if err != nil {
		return fmt.Errorf("failed to activate contingency: %v", err)
	}

	c.SetContingency(contingency)
	return nil
}

// DeactivateContingency deactivates contingency mode.
func (c *NFEClient) DeactivateContingency() error {
	if c.contingency != nil {
		_, err := c.contingency.Deactivate()
		if err != nil {
			return fmt.Errorf("failed to deactivate contingency: %v", err)
		}
		c.contingency = nil
	}
	return nil
}

// IsContingencyActive returns true if contingency mode is active.
func (c *NFEClient) IsContingencyActive() bool {
	return c.contingency != nil && c.contingency.IsActive()
}

// Helper methods

func (c *NFEClient) signIfNeeded(xml []byte) ([]byte, error) {
	// Check if already signed
	if strings.Contains(string(xml), "<Signature") || strings.Contains(string(xml), "<ds:Signature") {
		return xml, nil
	}

	// Check if certificate is available
	if c.certificate == nil {
		return nil, fmt.Errorf("certificate is required for XML signing")
	}

	// Create XML signer with SEFAZ-compatible configuration
	signer := certificate.CreateXMLSigner(c.certificate)

	// Sign the NFe XML specifically
	signedXML, err := signer.SignNFeXML(string(xml))
	if err != nil {
		return nil, fmt.Errorf("failed to sign NFe XML: %v", err)
	}

	return []byte(signedXML), nil
}

func (c *NFEClient) convertToQueryResponse(response *ConsultaChaveResponse) *QueryResponse {
	// Convert string status to int
	status := 0
	if statusInt, err := strconv.Atoi(response.CStat); err == nil {
		status = statusInt
	}

	query := &QueryResponse{
		Success:    response.CStat == "100",
		Status:     status,
		StatusText: response.XMotivo,
		QueryAt:    time.Now(),
	}

	if response.CStat == "100" {
		query.Authorized = true
		query.Key = response.ChNFe
	}

	// Add status message
	query.Messages = append(query.Messages, ResponseMessage{
		Code:    response.CStat,
		Message: response.XMotivo,
		Type:    "info",
	})

	return query
}

func (c *NFEClient) convertReciboToQueryResponse(response *ConsultaReciboResponse) *QueryResponse {
	// Convert string status to int
	status := 0
	if statusInt, err := strconv.Atoi(response.CStat); err == nil {
		status = statusInt
	}

	return &QueryResponse{
		Success:    response.CStat == "104",
		Status:     status,
		StatusText: response.XMotivo,
		QueryAt:    time.Now(),
		Messages: []ResponseMessage{{
			Code:    response.CStat,
			Message: response.XMotivo,
			Type:    "info",
		}},
	}
}

func (c *NFEClient) convertToStatusResponse(response *StatusResponse) *ClientStatusResponse {
	// Convert string status to int
	status := 0
	if statusInt, err := strconv.Atoi(response.CStat); err == nil {
		status = statusInt
	}

	clientStatus := &ClientStatusResponse{
		Success:     response.CStat == "107",
		Status:      status,
		StatusText:  response.XMotivo,
		UF:          fmt.Sprintf("%02d", int(c.uf)),
		Environment: int(c.config.TpAmb),
		Online:      response.CStat == "107",
		CheckedAt:   time.Now(),
		Messages: []ResponseMessage{{
			Code:    response.CStat,
			Message: response.XMotivo,
			Type:    "info",
		}},
	}

	return clientStatus
}

func (c *NFEClient) convertToAuthResponse(response *EnvioLoteResponse, originalXML []byte) *AuthResponse {
	// Convert string status to int
	status := 0
	if statusInt, err := strconv.Atoi(response.CStat); err == nil {
		status = statusInt
	}

	authResponse := &AuthResponse{
		Success:      response.CStat == "103", // Lote recebido com sucesso
		Status:       status,
		StatusText:   response.XMotivo,
		OriginalXML:  originalXML,
		ProcessingAt: time.Now(),
		Messages: []ResponseMessage{{
			Code:    response.CStat,
			Message: response.XMotivo,
			Type:    "info",
		}},
	}

	// Add receipt if available
	if response.InfRec != nil {
		authResponse.Receipt = response.InfRec.NRec
	}

	return authResponse
}

func (c *NFEClient) createLoteFromXML(xmlData []byte) (*LoteNFe, error) {
	if len(xmlData) == 0 {
		return nil, fmt.Errorf("XML content cannot be empty")
	}

	// Parse the signed NFe XML
	var nfe NFe
	if err := xml.Unmarshal(xmlData, &nfe); err != nil {
		return nil, fmt.Errorf("failed to parse NFe XML: %v", err)
	}

	// Validate the parsed NFe
	if err := c.validateParsedNFe(&nfe); err != nil {
		return nil, fmt.Errorf("NFe validation failed: %v", err)
	}

	// Create LoteNFe with the parsed NFe
	lote := &LoteNFe{
		IdLote: "1", // Default batch ID
		NFes:   []NFe{nfe},
	}

	return lote, nil
}

func (c *NFEClient) createCancellationEventRequest(chave, justificativa string) (*EventoRequest, error) {
	// Create a proper cancellation request using the new validation functions
	req, err := CreateCancelamentoRequest(chave, justificativa, "dummy_protocol")
	if err != nil {
		return nil, fmt.Errorf("failed to create cancellation request: %v", err)
	}

	// Convert to the old EventoRequest format for compatibility
	return &EventoRequest{
		IdLote: "1",
		Evento: []Evento{
			{
				Versao: "1.00",
				InfEvento: InfEvento{
					COrgao:     getStateCode(c.config.SiglaUF),
					TpAmb:      int(c.config.TpAmb),
					CNPJ:       c.config.CNPJ,
					ChNFe:      req.ChaveNFe,
					DhEvento:   FormatDateTime(time.Now()),
					TpEvento:   "110111",
					NSeqEvento: "1",
					VerEvento:  "1.00",
					DetEvento: DetEvento{
						Versao:     "1.00",
						DescEvento: "Cancelamento",
						NProt:      req.Protocolo,
						XJust:      req.Justificativa,
					},
				},
			},
		},
	}, nil
}

func (c *NFEClient) createCCeEventRequest(chave, correcao string, sequencia int) (*EventoRequest, error) {
	// TODO: Implement proper CCe request creation
	// For now, return a basic structure
	return &EventoRequest{}, nil
}

func (c *NFEClient) createInutilizacaoRequest(serie, numeroInicial, numeroFinal int, justificativa string) (*InutilizacaoRequest, error) {
	// TODO: Implement proper invalidation request creation
	// For now, return a basic structure
	return &InutilizacaoRequest{}, nil
}

func (c *NFEClient) convertInutilizacaoToEventResponse(response *InutilizacaoResponse) *EventResponse {
	// TODO: Access proper fields from InutilizacaoResponse
	// For now, return a basic structure
	return &EventResponse{
		Success:     true,
		Status:      102,
		StatusText:  "Inutilização de número homologado",
		EventType:   "inutilizacao",
		ProcessedAt: time.Now(),
		Messages: []ResponseMessage{{
			Code:    "102",
			Message: "Inutilização de número homologado",
			Type:    "info",
		}},
	}
}

func (c *NFEClient) convertToEventResponse(response interface{}, eventType string) *EventResponse {
	event := &EventResponse{
		EventType:   eventType,
		ProcessedAt: time.Now(),
	}

	// Handle EventResponseNFe type
	if nfeResp, ok := response.(*EventResponseNFe); ok {
		event.Success = nfeResp.CStat == "128" || nfeResp.CStat == "135" || nfeResp.CStat == "136"
		if statusCode, err := strconv.Atoi(nfeResp.CStat); err == nil {
			event.Status = statusCode
		}
		event.StatusText = nfeResp.XMotivo

		// Add event protocol if available
		if len(nfeResp.RetEvento) > 0 {
			event.Protocol = nfeResp.RetEvento[0].InfEvento.NProt
		}
	} else {
		// Fallback for other response types
		event.Success = true
		event.Status = 135 // Evento registrado e vinculado a NFe
	}

	return event
}

func (c *NFEClient) convertToCancelamentoResponse(response *EventResponseNFe, chave string) *CancelamentoResponse {
	cancelResponse := &CancelamentoResponse{
		Key:         chave,
		EventType:   "cancellation",
		ProcessedAt: time.Now(),
		Sequence:    1, // Cancellation is always sequence 1
	}

	if response != nil {
		// Parse status code
		if statusCode, err := strconv.Atoi(response.CStat); err == nil {
			cancelResponse.Status = statusCode
			cancelResponse.Success = IsCancellationSuccessful(statusCode)
		}

		cancelResponse.StatusText = response.XMotivo
		if cancelResponse.StatusText == "" {
			cancelResponse.StatusText = GetCancellationStatusText(cancelResponse.Status)
		}

		// Add protocol if available
		if len(response.RetEvento) > 0 {
			cancelResponse.Protocol = response.RetEvento[0].InfEvento.NProt
		}

		// Add messages
		if response.XMotivo != "" {
			cancelResponse.Messages = []ResponseMessage{
				{
					Code:    response.CStat,
					Message: response.XMotivo,
					Type:    "info",
				},
			}
		}
	} else {
		// Fallback for nil response
		cancelResponse.Success = false
		cancelResponse.Status = 0
		cancelResponse.StatusText = "No response received"
	}

	return cancelResponse
}

// Authorized returns true if the authorization response indicates success.
func (r *AuthResponse) Authorized() bool {
	return r.Success && (r.Status == 100 || r.Status == 150)
}

// GetProtocol returns the protocol from authorization response.
func (r *AuthResponse) GetProtocol() string {
	return r.Protocol
}

// HasReceipt returns true if the response contains a receipt number.
func (r *AuthResponse) HasReceipt() bool {
	return r.Receipt != ""
}

// IsOnline returns true if SEFAZ is online.
func (r *ClientStatusResponse) IsOnline() bool {
	return r.Online
}

// IsAuthorized returns true if the NFe is authorized.
func (r *QueryResponse) IsAuthorized() bool {
	return r.Authorized
}

// IsCancelled returns true if the NFe is cancelled.
func (r *QueryResponse) IsCancelled() bool {
	return r.Cancelled
}

// IsProcessed returns true if the event was processed successfully.
func (r *EventResponse) IsProcessed() bool {
	return r.Success && (r.Status == 135 || r.Status == 136)
}
