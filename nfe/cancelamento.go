// Package nfe provides specific functionality for NFe cancellation (cancelamento)
package nfe

import (
	"fmt"
	"strings"
	"time"
)

// Cancellation timeout constants
const (
	// CancellationTimeoutHours defines the maximum hours after authorization to cancel NFe
	CancellationTimeoutHours = 24

	// Minimum and maximum justification lengths
	MinJustificationLength = 15
	MaxJustificationLength = 255
)

// CancelamentoRequest represents a cancellation request structure
type CancelamentoRequest struct {
	ChaveNFe      string     `json:"chaveNFe" validate:"required,len=44"`
	Justificativa string     `json:"justificativa" validate:"required,min=15,max=255"`
	Protocolo     string     `json:"protocolo" validate:"required"`
	DhEvento      *time.Time `json:"dhEvento,omitempty"`
	Lote          string     `json:"lote,omitempty"`
}

// CancelamentoResponse represents a cancellation response structure
type CancelamentoResponse struct {
	Success     bool              `json:"success"`
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	Protocol    string            `json:"protocol,omitempty"`
	Key         string            `json:"key"`
	Sequence    int               `json:"sequence"`
	Messages    []ResponseMessage `json:"messages,omitempty"`
	XML         []byte            `json:"xml,omitempty"`
	ProcessedAt time.Time         `json:"processedAt"`
	EventType   string            `json:"eventType"`
}

// DetEventoCancelamento represents the specific event details for cancellation
type DetEventoCancelamento struct {
	Versao     string `xml:"versao,attr" json:"versao"`
	DescEvento string `xml:"descEvento" json:"descEvento"`
	NProt      string `xml:"nProt" json:"nProt"`
	XJust      string `xml:"xJust" json:"xJust"`
}

// CancellationStatus represents the possible status codes for cancellation events
type CancellationStatus int

const (
	// Event registered and linked to NFe
	CancellationStatusRegistered CancellationStatus = 135

	// Event already exists for this NFe
	CancellationStatusAlreadyExists CancellationStatus = 573

	// Cancellation event approved
	CancellationStatusApproved CancellationStatus = 155

	// Event rejected - outside deadline
	CancellationStatusOutsideDeadline CancellationStatus = 218

	// Event rejected - NFe not found
	CancellationStatusNFeNotFound CancellationStatus = 217
)

// ValidarCancelamento validates cancellation request parameters
func ValidarCancelamento(req *CancelamentoRequest) error {
	if req == nil {
		return fmt.Errorf("cancellation request cannot be nil")
	}

	// Validate NFe key
	if err := ValidateNFeKey(req.ChaveNFe); err != nil {
		return fmt.Errorf("invalid NFe key: %v", err)
	}

	// Validate justification
	if err := ValidateJustification(req.Justificativa); err != nil {
		return fmt.Errorf("invalid justification: %v", err)
	}

	// Validate protocol
	if strings.TrimSpace(req.Protocolo) == "" {
		return fmt.Errorf("protocol cannot be empty")
	}

	return nil
}

// ValidateNFeKey validates the NFe access key format and content
func ValidateNFeKey(chave string) error {
	chave = strings.TrimSpace(chave)

	if chave == "" {
		return fmt.Errorf("chave cannot be empty")
	}

	if len(chave) != 44 {
		return fmt.Errorf("chave must be exactly 44 digits, got %d", len(chave))
	}

	// Check if all characters are numeric
	for i, r := range chave {
		if r < '0' || r > '9' {
			return fmt.Errorf("chave must contain only numeric characters, invalid character at position %d", i)
		}
	}

	// Extract UF code and validate
	ufCode := chave[0:2]
	if err := validateUFCode(ufCode); err != nil {
		return fmt.Errorf("invalid UF code in chave: %v", err)
	}

	// Extract and validate date (positions 2-7: AAMM)
	dateStr := chave[2:6]
	if err := validateDateInKey(dateStr); err != nil {
		return fmt.Errorf("invalid date in chave: %v", err)
	}

	return nil
}

// ValidateJustification validates the cancellation justification text
func ValidateJustification(justificativa string) error {
	// Clean and trim the justification
	justificativa = strings.TrimSpace(justificativa)

	if justificativa == "" {
		return fmt.Errorf("justification cannot be empty")
	}

	if len(justificativa) < MinJustificationLength {
		return fmt.Errorf("justification must be at least %d characters, got %d", MinJustificationLength, len(justificativa))
	}

	if len(justificativa) > MaxJustificationLength {
		return fmt.Errorf("justification cannot exceed %d characters, got %d", MaxJustificationLength, len(justificativa))
	}

	// Check for basic content (not just spaces or special characters)
	if !hasValidContent(justificativa) {
		return fmt.Errorf("justification must contain meaningful content")
	}

	return nil
}

// ValidarPrazoCancelamento validates if NFe can still be cancelled within the legal deadline
func ValidarPrazoCancelamento(dhAutorizacao time.Time) error {
	if dhAutorizacao.IsZero() {
		return fmt.Errorf("authorization date cannot be zero")
	}

	deadline := dhAutorizacao.Add(CancellationTimeoutHours * time.Hour)
	now := time.Now()

	if now.After(deadline) {
		return fmt.Errorf("cancellation deadline exceeded: NFe was authorized on %v, deadline was %v, current time is %v",
			dhAutorizacao.Format("2006-01-02 15:04:05"),
			deadline.Format("2006-01-02 15:04:05"),
			now.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// CanBeCancelled checks if an NFe can be cancelled based on its current status and authorization date
func CanBeCancelled(authorized bool, cancelled bool, dhAutorizacao time.Time) (bool, error) {
	if !authorized {
		return false, fmt.Errorf("NFe must be authorized before it can be cancelled")
	}

	if cancelled {
		return false, fmt.Errorf("NFe is already cancelled")
	}

	if err := ValidarPrazoCancelamento(dhAutorizacao); err != nil {
		return false, err
	}

	return true, nil
}

// SanitizeJustification cleans and sanitizes the justification text
func SanitizeJustification(justificativa string) string {
	// Trim whitespace
	justificativa = strings.TrimSpace(justificativa)

	// Replace multiple spaces with single space
	justificativa = strings.Join(strings.Fields(justificativa), " ")

	// Remove or replace problematic characters
	replacements := map[string]string{
		"\n": " ",
		"\r": " ",
		"\t": " ",
		"\"": "'",
		"&":  "e",
		"<":  "",
		">":  "",
	}

	for old, new := range replacements {
		justificativa = strings.ReplaceAll(justificativa, old, new)
	}

	// Truncate if too long
	if len(justificativa) > MaxJustificationLength {
		justificativa = justificativa[:MaxJustificationLength]
	}

	return justificativa
}

// CreateCancelamentoRequest creates a properly formatted cancellation request
func CreateCancelamentoRequest(chaveNFe, justificativa, protocolo string) (*CancelamentoRequest, error) {
	// Sanitize inputs
	chaveNFe = strings.TrimSpace(chaveNFe)
	justificativa = SanitizeJustification(justificativa)
	protocolo = strings.TrimSpace(protocolo)

	req := &CancelamentoRequest{
		ChaveNFe:      chaveNFe,
		Justificativa: justificativa,
		Protocolo:     protocolo,
		DhEvento:      nil, // Will be set to current time when sending
		Lote:          "",  // Will be generated when sending
	}

	// Validate the request
	if err := ValidarCancelamento(req); err != nil {
		return nil, err
	}

	return req, nil
}

// GetCancellationStatusText returns a human-readable description for cancellation status codes
func GetCancellationStatusText(status int) string {
	switch CancellationStatus(status) {
	case CancellationStatusRegistered:
		return "Evento registrado e vinculado à NFe"
	case CancellationStatusAlreadyExists:
		return "Evento de cancelamento já existe para esta NFe"
	case CancellationStatusApproved:
		return "Cancelamento homologado"
	case CancellationStatusOutsideDeadline:
		return "Evento rejeitado - fora do prazo de cancelamento"
	case CancellationStatusNFeNotFound:
		return "Evento rejeitado - NFe não encontrada"
	default:
		return fmt.Sprintf("Status desconhecido: %d", status)
	}
}

// IsCancellationSuccessful checks if a cancellation status indicates success
func IsCancellationSuccessful(status int) bool {
	return status == int(CancellationStatusRegistered) ||
		status == int(CancellationStatusApproved)
}

// Helper functions

// validateUFCode validates if the UF code exists in Brazil
func validateUFCode(ufCode string) error {
	validUFs := map[string]string{
		"12": "AC", "27": "AL", "16": "AP", "23": "AM", "29": "BA",
		"85": "CE", "53": "DF", "32": "ES", "52": "GO", "21": "MA",
		"51": "MT", "50": "MS", "31": "MG", "15": "PA", "25": "PB",
		"41": "PR", "26": "PE", "22": "PI", "33": "RJ", "20": "RN",
		"43": "RS", "11": "RO", "14": "RR", "42": "SC", "35": "SP",
		"28": "SE", "17": "TO",
	}

	if _, exists := validUFs[ufCode]; !exists {
		return fmt.Errorf("invalid UF code: %s", ufCode)
	}

	return nil
}

// validateDateInKey validates the date portion of the NFe key (AAMM format)
func validateDateInKey(dateStr string) error {
	if len(dateStr) != 4 {
		return fmt.Errorf("date must be 4 digits (AAMM), got %d", len(dateStr))
	}

	// Check if all characters are numeric
	for _, r := range dateStr {
		if r < '0' || r > '9' {
			return fmt.Errorf("date must contain only numeric characters")
		}
	}

	// Extract year and month
	month := dateStr[2:4]

	// Validate month (01-12)
	if month < "01" || month > "12" {
		return fmt.Errorf("invalid month in date: %s", month)
	}

	return nil
}

// hasValidContent checks if the justification contains meaningful content
func hasValidContent(text string) bool {
	// Count alphanumeric characters
	alphanumericCount := 0
	for _, r := range text {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			alphanumericCount++
		}
	}

	// Must have at least 10 alphanumeric characters for meaningful content
	return alphanumericCount >= 10
}
