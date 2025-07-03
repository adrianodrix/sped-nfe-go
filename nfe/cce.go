// Package nfe provides specific functionality for NFe Carta de Correção Eletrônica (CCe)
package nfe

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// CCe constants
const (
	// CCeTypeEvent is the event type for Carta de Correção Eletrônica
	CCeTypeEvent = 110110

	// CCe sequence limits
	CCeMinSequence = 1
	CCeMaxSequence = 20

	// CCe correction text limits
	CCeMinCorrectionLength = 15
	CCeMaxCorrectionLength = 1000

	// CCe fixed usage conditions text (required by SEFAZ)
	CCeUsageConditions = "A Carta de Correção é disciplinada pelo § 1º-A do art. 7º do Convênio S/N, " +
		"de 15 de dezembro de 1970 e pode ser utilizada para regularização de erro ocorrido na emissão " +
		"de documento fiscal, desde que o erro não esteja relacionado com: I - as variáveis que " +
		"determinam o valor do imposto tais como: base de cálculo, alíquota, diferença de preço, " +
		"quantidade, valor da operação ou da prestação; II - a correção de dados cadastrais que " +
		"implique mudança do remetente ou do destinatário; III - a data de emissão ou de saída."
)

// CCeRequest represents a CCe request structure
type CCeRequest struct {
	ChaveNFe   string     `json:"chaveNFe" validate:"required,len=44"`
	Correcao   string     `json:"correcao" validate:"required,min=15,max=1000"`
	Sequencia  int        `json:"sequencia" validate:"required,min=1,max=20"`
	DhEvento   *time.Time `json:"dhEvento,omitempty"`
	Lote       string     `json:"lote,omitempty"`
	XCondUso   string     `json:"xCondUso,omitempty"` // Optional, will use default if empty
}

// CCeResponse represents a CCe response structure
type CCeResponse struct {
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
	Correction  string            `json:"correction,omitempty"`
}

// DetEventoCCe represents the specific event details for CCe
type DetEventoCCe struct {
	Versao     string `xml:"versao,attr" json:"versao"`
	DescEvento string `xml:"descEvento" json:"descEvento"`
	XCorrecao  string `xml:"xCorrecao" json:"xCorrecao"`
	XCondUso   string `xml:"xCondUso" json:"xCondUso"`
}

// CCeStatus represents the possible status codes for CCe events
type CCeStatus int

const (
	// Event registered and linked to NFe
	CCeStatusRegistered CCeStatus = 135

	// Event registered but not linked to NFe
	CCeStatusRegisteredNotLinked CCeStatus = 136

	// Sequence number greater than allowed
	CCeStatusSequenceExceeded CCeStatus = 594

	// Event already exists for this NFe with same sequence
	CCeStatusAlreadyExists CCeStatus = 573

	// Event rejected - NFe not found
	CCeStatusNFeNotFound CCeStatus = 217

	// Event rejected - NFe not authorized
	CCeStatusNFeNotAuthorized CCeStatus = 218

	// Event rejected - invalid correction text
	CCeStatusInvalidCorrection CCeStatus = 489
)

// ValidarCCe validates CCe request parameters
func ValidarCCe(req *CCeRequest) error {
	if req == nil {
		return fmt.Errorf("CCe request cannot be nil")
	}

	// Validate NFe key
	if err := ValidateNFeKey(req.ChaveNFe); err != nil {
		return fmt.Errorf("invalid NFe key: %v", err)
	}

	// Validate correction text
	if err := ValidateCorrection(req.Correcao); err != nil {
		return fmt.Errorf("invalid correction text: %v", err)
	}

	// Validate sequence
	if err := ValidateSequence(req.Sequencia); err != nil {
		return fmt.Errorf("invalid sequence: %v", err)
	}

	// Set default usage conditions if not provided
	if req.XCondUso == "" {
		req.XCondUso = CCeUsageConditions
	}

	return nil
}

// ValidateCorrection validates the correction text for CCe
func ValidateCorrection(correcao string) error {
	// Clean and trim the correction text
	correcao = strings.TrimSpace(correcao)

	if correcao == "" {
		return fmt.Errorf("correction text cannot be empty")
	}

	if len(correcao) < CCeMinCorrectionLength {
		return fmt.Errorf("correction text must be at least %d characters, got %d", CCeMinCorrectionLength, len(correcao))
	}

	if len(correcao) > CCeMaxCorrectionLength {
		return fmt.Errorf("correction text cannot exceed %d characters, got %d", CCeMaxCorrectionLength, len(correcao))
	}

	// Check for valid content using SEFAZ pattern
	if !isValidCorrectionContent(correcao) {
		return fmt.Errorf("correction text contains invalid characters or format")
	}

	return nil
}

// ValidateSequence validates the CCe sequence number
func ValidateSequence(sequencia int) error {
	if sequencia < CCeMinSequence {
		return fmt.Errorf("sequence must be at least %d, got %d", CCeMinSequence, sequencia)
	}

	if sequencia > CCeMaxSequence {
		return fmt.Errorf("sequence cannot exceed %d, got %d", CCeMaxSequence, sequencia)
	}

	return nil
}

// isValidCorrectionContent validates correction text using SEFAZ pattern
// Pattern: [!-ÿ]{1}[ -ÿ]{0,}[!-ÿ]{1}|[!-ÿ]{1}
func isValidCorrectionContent(text string) bool {
	// SEFAZ pattern for correction text
	pattern := `^[!-ÿ]{1}[ -ÿ]{0,}[!-ÿ]{1}$|^[!-ÿ]{1}$`
	matched, err := regexp.MatchString(pattern, text)
	if err != nil {
		return false
	}
	return matched
}

// SanitizeCorrection cleans and sanitizes the correction text
func SanitizeCorrection(correcao string) string {
	// Trim whitespace
	correcao = strings.TrimSpace(correcao)

	// Replace multiple spaces with single space
	correcao = strings.Join(strings.Fields(correcao), " ")

	// Remove or replace problematic characters that might cause XML issues
	replacements := map[string]string{
		"\n": " ",
		"\r": " ",
		"\t": " ",
		"&":  "e",
		"<":  "",
		">":  "",
	}

	for old, new := range replacements {
		correcao = strings.ReplaceAll(correcao, old, new)
	}

	// Truncate if too long
	if len(correcao) > CCeMaxCorrectionLength {
		correcao = correcao[:CCeMaxCorrectionLength]
	}

	return correcao
}

// CreateCCeRequest creates a properly formatted CCe request
func CreateCCeRequest(chaveNFe, correcao string, sequencia int) (*CCeRequest, error) {
	// Sanitize inputs
	chaveNFe = strings.TrimSpace(chaveNFe)
	correcao = SanitizeCorrection(correcao)

	req := &CCeRequest{
		ChaveNFe:  chaveNFe,
		Correcao:  correcao,
		Sequencia: sequencia,
		DhEvento:  nil, // Will be set to current time when sending
		Lote:      "",  // Will be generated when sending
		XCondUso:  CCeUsageConditions,
	}

	// Validate the request
	if err := ValidarCCe(req); err != nil {
		return nil, err
	}

	return req, nil
}

// GetCCeStatusText returns a human-readable description for CCe status codes
func GetCCeStatusText(status int) string {
	switch CCeStatus(status) {
	case CCeStatusRegistered:
		return "Evento registrado e vinculado à NFe"
	case CCeStatusRegisteredNotLinked:
		return "Evento registrado, mas não vinculado à NFe"
	case CCeStatusSequenceExceeded:
		return "Número de sequência maior que permitido"
	case CCeStatusAlreadyExists:
		return "Evento já existe para esta NFe com a mesma sequência"
	case CCeStatusNFeNotFound:
		return "Evento rejeitado - NFe não encontrada"
	case CCeStatusNFeNotAuthorized:
		return "Evento rejeitado - NFe não autorizada"
	case CCeStatusInvalidCorrection:
		return "Evento rejeitado - texto de correção inválido"
	default:
		return fmt.Sprintf("Status desconhecido: %d", status)
	}
}

// IsCCeSuccessful checks if a CCe status indicates success
func IsCCeSuccessful(status int) bool {
	return status == int(CCeStatusRegistered) ||
		status == int(CCeStatusRegisteredNotLinked)
}

// CanSendCCe checks if an NFe can receive a CCe based on its current status
func CanSendCCe(authorized bool, sequencia int) (bool, error) {
	if !authorized {
		return false, fmt.Errorf("NFe must be authorized before CCe can be sent")
	}

	if sequencia < CCeMinSequence {
		return false, fmt.Errorf("invalid sequence number: %d (must be between %d and %d)", sequencia, CCeMinSequence, CCeMaxSequence)
	}

	if sequencia > CCeMaxSequence {
		return false, fmt.Errorf("maximum CCe sequence exceeded: %d (maximum is %d)", sequencia, CCeMaxSequence)
	}

	return true, nil
}

// GetNextSequence calculates the next sequence number for CCe
func GetNextSequence(lastSequence int) (int, error) {
	nextSequence := lastSequence + 1

	if nextSequence > CCeMaxSequence {
		return 0, fmt.Errorf("maximum CCe sequence exceeded: next would be %d (maximum is %d)", nextSequence, CCeMaxSequence)
	}

	return nextSequence, nil
}

// ValidateSequenceIncrement validates that the new sequence is properly incremented
func ValidateSequenceIncrement(currentSequence, newSequence int) error {
	if newSequence != currentSequence+1 {
		return fmt.Errorf("sequence must be incremental: current=%d, new=%d (expected %d)", currentSequence, newSequence, currentSequence+1)
	}

	if newSequence > CCeMaxSequence {
		return fmt.Errorf("sequence %d exceeds maximum allowed (%d)", newSequence, CCeMaxSequence)
	}

	return nil
}

// CreateCCeTagAdic creates the additional XML tags for CCe event
func CreateCCeTagAdic(correcao, condUso string) string {
	// Sanitize inputs
	correcao = SanitizeCorrection(correcao)
	if condUso == "" {
		condUso = CCeUsageConditions
	}

	return fmt.Sprintf("<xCorrecao>%s</xCorrecao><xCondUso>%s</xCondUso>", correcao, condUso)
}

// ParseCCeSequenceFromKey extracts the sequence number from a CCe event ID
func ParseCCeSequenceFromKey(eventID string) (int, error) {
	// Event ID format: ID{tpEvento}{chNFe}{nSeqEvento}
	// For CCe: ID110110{44-digit-key}{2-digit-sequence}
	if len(eventID) < 54 { // ID + 6 digits (tpEvento) + 44 digits (chNFe) + 2 digits (sequence) = 54
		return 0, fmt.Errorf("invalid event ID format: %s", eventID)
	}

	// Extract sequence from the end (last 2 digits)
	seqStr := eventID[len(eventID)-2:]
	seq, err := strconv.Atoi(seqStr)
	if err != nil {
		return 0, fmt.Errorf("invalid sequence in event ID: %s", seqStr)
	}

	return seq, nil
}

// FormatCCeSequence formats sequence number with leading zero if needed
func FormatCCeSequence(sequencia int) string {
	return fmt.Sprintf("%02d", sequencia)
}

// GetCCeEventName returns the event name for CCe
func GetCCeEventName() string {
	return "Carta de Correção Eletrônica"
}

// GetCCeEventDescription returns a description for CCe events
func GetCCeEventDescription() string {
	return "Carta de Correção Eletrônica"
}