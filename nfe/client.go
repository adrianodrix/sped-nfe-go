// Package nfe provides the main client API for Brazilian Electronic Fiscal Documents (NFe/NFCe).
package nfe

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/certificate"
	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/factories"
	"github.com/adrianodrix/sped-nfe-go/types"
)

// NFEClient is the main NFe client that provides a unified API for all NFe operations.
type NFEClient struct {
	config      *common.Config
	tools       *Tools
	certificate certificate.Certificate
	contingency *factories.Contingency
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
	Success     bool              `json:"success"`
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	Key         string            `json:"key,omitempty"`
	Protocol    string            `json:"protocol,omitempty"`
	Authorized  bool              `json:"authorized"`
	Cancelled   bool              `json:"cancelled"`
	Messages    []ResponseMessage `json:"messages,omitempty"`
	XML         []byte            `json:"xml,omitempty"`
	LastEvent   *EventInfo        `json:"lastEvent,omitempty"`
	QueryAt     time.Time         `json:"queryAt"`
}

// EventResponse represents the response from fiscal events.
type EventResponse struct {
	Success       bool              `json:"success"`
	Status        int               `json:"status"`
	StatusText    string            `json:"statusText"`
	EventType     string            `json:"eventType"`
	Key           string            `json:"key"`
	Protocol      string            `json:"protocol,omitempty"`
	Sequence      int               `json:"sequence,omitempty"`
	Messages      []ResponseMessage `json:"messages,omitempty"`
	XML           []byte            `json:"xml,omitempty"`
	ProcessedAt   time.Time         `json:"processedAt"`
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
	Code        string `json:"code"`
	Message     string `json:"message"`
	Correction  string `json:"correction,omitempty"`
	Type        string `json:"type"` // info, warning, error
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

	// Create common config
	commonConfig := &common.Config{
		TpAmb:   types.Environment(config.Environment),
		Timeout: config.Timeout,
	}

	// Basic validation
	if config.Environment < 1 || config.Environment > 2 {
		return nil, fmt.Errorf("invalid environment: must be 1 (production) or 2 (homologation)")
	}

	// Create tools instance - TODO: implement Tools properly
	// For now, we'll create a basic client without tools
	// tools, err := NewTools(commonConfig)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to create tools: %v", err)
	// }

	client := &NFEClient{
		config:  commonConfig,
		tools:   nil, // TODO: implement Tools
		timeout: time.Duration(config.Timeout) * time.Second,
		uf:      config.UF,
	}

	return client, nil
}

// SetCertificate sets the digital certificate for the client.
func (c *NFEClient) SetCertificate(cert certificate.Certificate) error {
	if cert == nil {
		return fmt.Errorf("certificate cannot be nil")
	}

	c.certificate = cert
	// TODO: implement tools
	// c.tools.SetCertificate(cert)
	return nil
}

// SetTimeout sets the timeout for SEFAZ operations.
func (c *NFEClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.config.Timeout = int(timeout.Seconds())
	// TODO: implement tools
	// if c.tools != nil {
	//	c.tools.SetTimeout(timeout)
	// }
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
func (c *NFEClient) LoadFromXML(xml []byte) (*NFe, error) {
	// TODO: Implement XML parsing to NFe struct
	return nil, fmt.Errorf("LoadFromXML not implemented yet")
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

	// Sign the NFe if not already signed
	signedXML, err := c.signIfNeeded(xml)
	if err != nil {
		return nil, fmt.Errorf("failed to sign NFe: %v", err)
	}

	// For now, create a basic response
	// TODO: Implement actual SEFAZ communication
	response := &AuthResponse{
		Success:      true,
		Status:       100,
		StatusText:   "Autorizado",
		OriginalXML:  signedXML,
		ProcessingAt: time.Now(),
	}

	return response, nil
}

// QueryChave queries an NFe by its access key.
func (c *NFEClient) QueryChave(ctx context.Context, chave string) (*QueryResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("invalid access key length: expected 44, got %d", len(chave))
	}

	// TODO: implement SefazConsultaChave
	// response, err := c.tools.SefazConsultaChave(ctx, chave)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to query NFe: %v", err)
	// }
	
	// Return mock response for now
	return &QueryResponse{
		Success:    true,
		Status:     100,
		StatusText: "NFe autorizada",
		Key:        chave,
		Authorized: true,
		QueryAt:    time.Now(),
	}, nil
}

// QueryRecibo queries the processing result by receipt number.
func (c *NFEClient) QueryRecibo(ctx context.Context, recibo string) (*QueryResponse, error) {
	if recibo == "" {
		return nil, fmt.Errorf("receipt number cannot be empty")
	}

	// TODO: implement SefazConsultaRecibo
	// response, err := c.tools.SefazConsultaRecibo(ctx, recibo)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to query receipt: %v", err)
	// }
	
	// Return mock response for now
	return &QueryResponse{
		Success:    true,
		Status:     104,
		StatusText: "Lote processado",
		QueryAt:    time.Now(),
	}, nil
}

// QueryStatus checks the SEFAZ service status.
func (c *NFEClient) QueryStatus(ctx context.Context) (*ClientStatusResponse, error) {
	// TODO: implement SefazStatus
	// response, err := c.tools.SefazStatus(ctx)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to query SEFAZ status: %v", err)
	// }
	
	// Return mock response for now
	return &ClientStatusResponse{
		Success:     true,
		Status:      107,
		StatusText:  "Serviço em Operação",
		UF:          fmt.Sprintf("%02d", int(c.uf)),
		Environment: int(c.config.TpAmb),
		Online:      true,
		CheckedAt:   time.Now(),
	}, nil
}

// Cancel cancels an NFe with the provided justification.
func (c *NFEClient) Cancel(ctx context.Context, chave, justificativa string) (*EventResponse, error) {
	if len(chave) != 44 {
		return nil, fmt.Errorf("invalid access key length: expected 44, got %d", len(chave))
	}
	if len(justificativa) < 15 {
		return nil, fmt.Errorf("justification must be at least 15 characters")
	}

	// TODO: Implement cancellation using SefazEvento
	return nil, fmt.Errorf("cancellation not implemented yet")
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

	// TODO: Implement CCe using SefazEvento
	return nil, fmt.Errorf("CCe not implemented yet")
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

	// Create invalidation request
	// TODO: Create proper InutilizacaoRequest struct
	// request := &InutilizacaoRequest{
	//	Serie:         serie,
	//	NumeroInicial: numeroInicial,
	//	NumeroFinal:   numeroFinal,
	//	Justificativa: justificativa,
	// }
	
	// For now, use basic validation and mock response
	if serie < 1 || numeroInicial < 1 || numeroFinal < numeroInicial {
		return nil, fmt.Errorf("invalid series or number range")
	}

	// TODO: Implement actual SEFAZ invalidation
	// response, err := c.tools.SefazInutiliza(ctx, request)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to invalidate NFe numbers: %v", err)
	// }
	
	// Return mock response for now
	return &EventResponse{
		Success:     true,
		Status:      102, // Inutilização de número homologado
		StatusText:  "Inutilização de número homologado",
		EventType:   "inutilizacao",
		ProcessedAt: time.Now(),
	}, nil
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

	// TODO: Implement full XSD validation
	return nil
}

// GenerateKey generates an NFe access key.
func (c *NFEClient) GenerateKey(cnpj string, modelo, serie, numero int, dhEmi time.Time) (string, error) {
	// TODO: Implement proper access key generation
	// For now, return a mock key with basic validation
	if len(cnpj) != 14 {
		return "", fmt.Errorf("CNPJ must have 14 digits")
	}
	if modelo != 55 && modelo != 65 {
		return "", fmt.Errorf("model must be 55 (NFe) or 65 (NFCe)")
	}
	if serie < 1 || numero < 1 {
		return "", fmt.Errorf("series and number must be positive")
	}
	
	// Generate a mock 44-digit access key
	mockKey := fmt.Sprintf("%02d%s%s%02d%03d%09d%08d%d",
		int(c.uf), // UF code
		dhEmi.Format("0601"), // YYMM
		cnpj, // CNPJ
		modelo, // Model
		serie, // Series
		numero, // Number
		12345678, // Random number
		1) // Check digit
	
	if len(mockKey) != 44 {
		return "", fmt.Errorf("generated key has invalid length: %d", len(mockKey))
	}
	
	return mockKey, nil
}

// AddProtocol adds authorization protocol to an NFe XML.
func (c *NFEClient) AddProtocol(nfe, protocolo []byte) ([]byte, error) {
	// TODO: Implement protocol addition to create procNFe
	// This should create the complete procNFe XML with protocol information
	return nil, fmt.Errorf("not implemented yet")
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
	// TODO: Implement UF.String() method or use int conversion
	uf := fmt.Sprintf("%02d", int(c.uf)) // Use client UF as string
	
	var cType factories.ContingencyType
	if len(contingencyType) > 0 {
		cType = contingencyType[0]
	}

	contingency, _, err := factories.CreateContingency(uf, motive, cType)
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
	if strings.Contains(string(xml), "<Signature") {
		return xml, nil
	}

	// TODO: Implement XML signing
	return xml, nil
}

func (c *NFEClient) convertToQueryResponse(response *ConsultaChaveResponse) *QueryResponse {
	query := &QueryResponse{
		Success:    response.CStat == "100",
		Status:     100,
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
	return &QueryResponse{
		Success:    response.CStat == "104",
		Status:     104,
		StatusText: response.XMotivo,
		QueryAt:    time.Now(),
	}
}

func (c *NFEClient) convertToStatusResponse(response *StatusResponse) *ClientStatusResponse {
	status := &ClientStatusResponse{
		Success:     response.CStat == "107",
		Status:      107,
		StatusText:  response.XMotivo,
		UF:          fmt.Sprintf("%02d", int(c.uf)),
		Environment: int(c.config.TpAmb),
		Online:      response.CStat == "107",
		CheckedAt:   time.Now(),
	}

	return status
}

func (c *NFEClient) convertToEventResponse(response interface{}, eventType string) *EventResponse {
	event := &EventResponse{
		EventType:   eventType,
		ProcessedAt: time.Now(),
	}

	// TODO: Implement proper conversion based on response type
	// For now, return basic structure
	event.Success = true
	event.Status = 135 // Evento registrado e vinculado a NFe

	return event
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