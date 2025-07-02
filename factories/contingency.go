// Package factories provides utility factories for NFe processing including contingency management.
package factories

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// ContingencyType represents the type of contingency
type ContingencyType string

const (
	// ContingencySVCAN represents SEFAZ Virtual de Contingência do Ambiente Nacional
	ContingencySVCAN ContingencyType = "SVCAN"
	// ContingencySVCRS represents SEFAZ Virtual de Contingência do RS
	ContingencySVCRS ContingencyType = "SVCRS"
)

// EmissionType represents the emission type in contingency
type EmissionType int

const (
	// EmissionNormal represents normal emission (1)
	EmissionNormal EmissionType = 1
	// EmissionEPEC represents EPEC contingency (4)
	EmissionEPEC EmissionType = 4
	// EmissionFSDA represents FS-DA contingency (5)
	EmissionFSDA EmissionType = 5
	// EmissionSVCAN represents SVC-AN contingency (6)
	EmissionSVCAN EmissionType = 6
	// EmissionSVCRS represents SVC-RS contingency (7)
	EmissionSVCRS EmissionType = 7
	// EmissionOffline represents offline contingency for NFCe (9)
	EmissionOffline EmissionType = 9
)

// Contingency manages contingency mode for NFe
type Contingency struct {
	Type      ContingencyType `json:"type"`      // Type of contingency (SVCAN or SVCRS)
	Motive    string          `json:"motive"`    // Reason for entering contingency mode
	Timestamp int64           `json:"timestamp"` // Unix timestamp when contingency was activated
	TpEmis    EmissionType    `json:"tpEmis"`    // Emission type code
}

// ContingencyConfig holds configuration for creating contingency
type ContingencyConfig struct {
	UF     string          // State code (SP, RJ, etc.)
	Motive string          // Reason for contingency (15-255 UTF-8 characters)
	Type   ContingencyType // Optional: force specific contingency type
}

// NewContingency creates a new contingency manager
func NewContingency(jsonData ...string) (*Contingency, error) {
	c := &Contingency{
		Type:      "",
		Motive:    "",
		Timestamp: 0,
		TpEmis:    EmissionNormal,
	}

	if len(jsonData) > 0 && jsonData[0] != "" {
		if err := c.Load(jsonData[0]); err != nil {
			return nil, fmt.Errorf("failed to load contingency data: %v", err)
		}
	}

	return c, nil
}

// Load loads contingency configuration from JSON string
func (c *Contingency) Load(jsonData string) error {
	if err := json.Unmarshal([]byte(jsonData), c); err != nil {
		return fmt.Errorf("failed to unmarshal contingency JSON: %v", err)
	}
	return nil
}

// Activate activates contingency mode for a specific state
func (c *Contingency) Activate(config ContingencyConfig) (string, error) {
	// State to contingency type mapping
	stateMapping := map[string]ContingencyType{
		"AC": ContingencySVCAN, "AL": ContingencySVCAN, "AM": ContingencySVCRS,
		"AP": ContingencySVCAN, "BA": ContingencySVCRS, "CE": ContingencySVCAN,
		"DF": ContingencySVCAN, "ES": ContingencySVCAN, "GO": ContingencySVCRS,
		"MA": ContingencySVCRS, "MG": ContingencySVCAN, "MS": ContingencySVCRS,
		"MT": ContingencySVCRS, "PA": ContingencySVCAN, "PB": ContingencySVCAN,
		"PE": ContingencySVCRS, "PI": ContingencySVCAN, "PR": ContingencySVCRS,
		"RJ": ContingencySVCAN, "RN": ContingencySVCAN, "RO": ContingencySVCAN,
		"RR": ContingencySVCAN, "RS": ContingencySVCAN, "SC": ContingencySVCAN,
		"SE": ContingencySVCAN, "SP": ContingencySVCAN, "TO": ContingencySVCAN,
	}

	// Validate motive length (15-255 UTF-8 characters)
	motive := strings.TrimSpace(config.Motive)
	motiveLen := utf8.RuneCountInString(motive)
	if motiveLen < 15 || motiveLen > 255 {
		return "", fmt.Errorf("justification must be between 15 and 255 UTF-8 characters, got %d", motiveLen)
	}

	// Validate and set contingency type
	var contingencyType ContingencyType
	if config.Type != "" {
		// Validate provided type
		normalizedType := ContingencyType(strings.ToUpper(strings.ReplaceAll(string(config.Type), "-", "")))
		if normalizedType != ContingencySVCAN && normalizedType != ContingencySVCRS {
			return "", fmt.Errorf("invalid contingency type: %s. Use SVCAN or SVCRS", config.Type)
		}
		contingencyType = normalizedType
	} else {
		// Use default type for state
		var exists bool
		contingencyType, exists = stateMapping[strings.ToUpper(config.UF)]
		if !exists {
			return "", fmt.Errorf("unknown state: %s", config.UF)
		}
	}

	// Set contingency data
	c.Type = contingencyType
	c.Motive = motive
	c.Timestamp = time.Now().UTC().Unix()
	
	// Set emission type based on contingency type
	switch contingencyType {
	case ContingencySVCAN:
		c.TpEmis = EmissionSVCAN
	case ContingencySVCRS:
		c.TpEmis = EmissionSVCRS
	default:
		c.TpEmis = EmissionNormal
	}

	return c.ToJSON()
}

// Deactivate deactivates contingency mode
func (c *Contingency) Deactivate() (string, error) {
	c.Type = ""
	c.Motive = ""
	c.Timestamp = 0
	c.TpEmis = EmissionNormal

	return c.ToJSON()
}

// IsActive returns true if contingency mode is active
func (c *Contingency) IsActive() bool {
	return c.Type != "" && c.Timestamp > 0
}

// GetFormattedDateTime returns the contingency activation date/time in ISO format
func (c *Contingency) GetFormattedDateTime() string {
	if c.Timestamp == 0 {
		return ""
	}
	return time.Unix(c.Timestamp, 0).UTC().Format("2006-01-02T15:04:05-07:00")
}

// GetContingencyInfo returns contingency information for XML inclusion
func (c *Contingency) GetContingencyInfo() map[string]interface{} {
	if !c.IsActive() {
		return map[string]interface{}{
			"tpEmis": int(EmissionNormal),
		}
	}

	return map[string]interface{}{
		"tpEmis": int(c.TpEmis),
		"dhCont": c.GetFormattedDateTime(),
		"xJust":  c.Motive,
	}
}

// ToJSON converts contingency to JSON string
func (c *Contingency) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal contingency to JSON: %v", err)
	}
	return string(data), nil
}

// String returns a JSON representation of the contingency
func (c *Contingency) String() string {
	json, _ := c.ToJSON()
	return json
}

// ValidateContingencyData validates contingency data format
func ValidateContingencyData(jsonData string) error {
	var c Contingency
	if err := json.Unmarshal([]byte(jsonData), &c); err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}

	// Validate contingency type if present
	if c.Type != "" {
		if c.Type != ContingencySVCAN && c.Type != ContingencySVCRS {
			return fmt.Errorf("invalid contingency type: %s", c.Type)
		}
	}

	// Validate motive if present
	if c.Motive != "" {
		motiveLen := utf8.RuneCountInString(c.Motive)
		if motiveLen < 15 || motiveLen > 255 {
			return fmt.Errorf("motive must be between 15 and 255 UTF-8 characters")
		}
	}

	// Validate emission type
	validEmissionTypes := []EmissionType{
		EmissionNormal, EmissionEPEC, EmissionFSDA,
		EmissionSVCAN, EmissionSVCRS, EmissionOffline,
	}
	
	validType := false
	for _, validEmType := range validEmissionTypes {
		if c.TpEmis == validEmType {
			validType = true
			break
		}
	}
	
	if !validType {
		return fmt.Errorf("invalid emission type: %d", c.TpEmis)
	}

	return nil
}

// GetStateContingencyType returns the default contingency type for a state
func GetStateContingencyType(uf string) (ContingencyType, error) {
	stateMapping := map[string]ContingencyType{
		"AC": ContingencySVCAN, "AL": ContingencySVCAN, "AM": ContingencySVCRS,
		"AP": ContingencySVCAN, "BA": ContingencySVCRS, "CE": ContingencySVCAN,
		"DF": ContingencySVCAN, "ES": ContingencySVCAN, "GO": ContingencySVCRS,
		"MA": ContingencySVCRS, "MG": ContingencySVCAN, "MS": ContingencySVCRS,
		"MT": ContingencySVCRS, "PA": ContingencySVCAN, "PB": ContingencySVCAN,
		"PE": ContingencySVCRS, "PI": ContingencySVCAN, "PR": ContingencySVCRS,
		"RJ": ContingencySVCAN, "RN": ContingencySVCAN, "RO": ContingencySVCAN,
		"RR": ContingencySVCAN, "RS": ContingencySVCAN, "SC": ContingencySVCAN,
		"SE": ContingencySVCAN, "SP": ContingencySVCAN, "TO": ContingencySVCAN,
	}

	contingencyType, exists := stateMapping[strings.ToUpper(uf)]
	if !exists {
		return "", fmt.Errorf("unknown state: %s", uf)
	}

	return contingencyType, nil
}

// ContingencyBuilder provides a fluent interface for contingency creation
type ContingencyBuilder struct {
	config ContingencyConfig
}

// NewContingencyBuilder creates a new contingency builder
func NewContingencyBuilder() *ContingencyBuilder {
	return &ContingencyBuilder{
		config: ContingencyConfig{},
	}
}

// ForState sets the state for contingency
func (b *ContingencyBuilder) ForState(uf string) *ContingencyBuilder {
	b.config.UF = strings.ToUpper(uf)
	return b
}

// WithMotive sets the motive for contingency
func (b *ContingencyBuilder) WithMotive(motive string) *ContingencyBuilder {
	b.config.Motive = motive
	return b
}

// WithType sets a specific contingency type
func (b *ContingencyBuilder) WithType(contingencyType ContingencyType) *ContingencyBuilder {
	b.config.Type = contingencyType
	return b
}

// Activate activates contingency with the configured parameters
func (b *ContingencyBuilder) Activate() (*Contingency, string, error) {
	c, err := NewContingency()
	if err != nil {
		return nil, "", err
	}

	jsonData, err := c.Activate(b.config)
	if err != nil {
		return nil, "", err
	}

	return c, jsonData, nil
}

// CreateContingency is a convenience function to create and activate contingency
func CreateContingency(uf, motive string, contingencyType ...ContingencyType) (*Contingency, string, error) {
	builder := NewContingencyBuilder().ForState(uf).WithMotive(motive)
	
	if len(contingencyType) > 0 {
		builder = builder.WithType(contingencyType[0])
	}
	
	return builder.Activate()
}