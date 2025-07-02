// Package common provides common functionality and configuration for sped-nfe-go library.
// This package contains shared configuration and base types used across the library.
package common

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/adrianodrix/sped-nfe-go/types"
)

// Config represents the main configuration for NFe operations
// This struct maps to the JSON schema used by the PHP project
type Config struct {
	// Ambiente de operação (1=Produção, 2=Homologação)
	TpAmb types.Environment `json:"tpAmb" validate:"required,oneof=1 2"`
	
	// Dados da empresa
	RazaoSocial string `json:"razaosocial" validate:"required,min=1"`
	CNPJ        string `json:"cnpj" validate:"required,min=11,max=14,numeric"`
	SiglaUF     string `json:"siglaUF" validate:"required,len=2"`
	
	// Configurações técnicas
	Schemes string `json:"schemes" validate:"required"`
	Versao  string `json:"versao" validate:"required,oneof=3.10 4.00"`
	
	// Configurações opcionais
	Atualizacao *string `json:"atualizacao,omitempty"`
	TokenIBPT   *string `json:"tokenIBPT,omitempty"`
	CSC         *string `json:"CSC,omitempty"`
	CSCId       *string `json:"CSCid,omitempty"`
	
	// Configurações de proxy
	ProxyConf *ProxyConfig `json:"aProxyConf,omitempty"`
	
	// Configurações adicionais (não presentes no schema original)
	Timeout int `json:"timeout,omitempty"`
}

// ProxyConfig represents proxy configuration settings
type ProxyConfig struct {
	ProxyIP   *string `json:"proxyIp,omitempty"`
	ProxyPort *string `json:"proxyPort,omitempty"`
	ProxyUser *string `json:"proxyUser,omitempty"`
	ProxyPass *string `json:"proxyPass,omitempty"`
}

// ClientConfig represents the simplified configuration for client initialization
type ClientConfig struct {
	Environment types.Environment
	UF          types.UF
	Timeout     time.Duration
	ProxyConf   *ProxyConfig
}

// NewClientConfig creates a new ClientConfig with default values
func NewClientConfig() *ClientConfig {
	return &ClientConfig{
		Environment: types.Homologation,
		UF:          types.SP,
		Timeout:     types.DefaultTimeoutSeconds * time.Second,
	}
}

// ValidateConfig validates a configuration struct
func ValidateConfig(config *Config) error {
	if config == nil {
		return errors.NewConfigError("configuration cannot be nil", "", nil)
	}

	// Validate environment
	if !config.TpAmb.IsValid() {
		return errors.NewConfigError("invalid environment", "tpAmb", config.TpAmb)
	}

	// Validate required fields
	if strings.TrimSpace(config.RazaoSocial) == "" {
		return errors.NewConfigError("razao social cannot be empty", "razaosocial", config.RazaoSocial)
	}

	// Validate CNPJ format
	if err := validateCNPJ(config.CNPJ); err != nil {
		return errors.NewConfigError("invalid CNPJ format", "cnpj", config.CNPJ)
	}

	// Validate UF
	if err := validateUF(config.SiglaUF); err != nil {
		return errors.NewConfigError("invalid UF", "siglaUF", config.SiglaUF)
	}

	// Validate version
	if config.Versao != string(types.Versao310) && config.Versao != string(types.Versao400) {
		return errors.NewConfigError("invalid version", "versao", config.Versao)
	}

	// Validate schemes path
	if strings.TrimSpace(config.Schemes) == "" {
		return errors.NewConfigError("schemes path cannot be empty", "schemes", config.Schemes)
	}

	// Validate timeout if provided
	if config.Timeout > 0 {
		if config.Timeout < types.MinTimeoutSeconds || config.Timeout > types.MaxTimeoutSeconds {
			return errors.NewConfigError(
				fmt.Sprintf("timeout must be between %d and %d seconds", types.MinTimeoutSeconds, types.MaxTimeoutSeconds),
				"timeout",
				config.Timeout,
			)
		}
	}

	// Validate proxy configuration if provided
	if config.ProxyConf != nil {
		if err := validateProxyConfig(config.ProxyConf); err != nil {
			return err
		}
	}

	return nil
}

// ValidateClientConfig validates a ClientConfig struct
func ValidateClientConfig(config *ClientConfig) error {
	if config == nil {
		return errors.NewConfigError("client configuration cannot be nil", "", nil)
	}

	// Validate environment
	if !config.Environment.IsValid() {
		return errors.NewConfigError("invalid environment", "environment", config.Environment)
	}

	// Validate UF
	if !config.UF.IsValid() {
		return errors.NewConfigError("invalid UF", "uf", config.UF)
	}

	// Validate timeout
	minTimeout := time.Duration(types.MinTimeoutSeconds) * time.Second
	maxTimeout := time.Duration(types.MaxTimeoutSeconds) * time.Second
	
	if config.Timeout < minTimeout || config.Timeout > maxTimeout {
		return errors.NewConfigError(
			fmt.Sprintf("timeout must be between %v and %v", minTimeout, maxTimeout),
			"timeout",
			config.Timeout,
		)
	}

	// Validate proxy configuration if provided
	if config.ProxyConf != nil {
		if err := validateProxyConfig(config.ProxyConf); err != nil {
			return err
		}
	}

	return nil
}

// ParseConfigJSON parses and validates a JSON configuration
// This function replicates the behavior of Config::validate() from PHP
func ParseConfigJSON(jsonData []byte) (*Config, error) {
	if len(jsonData) == 0 {
		return nil, errors.NewConfigError("JSON data cannot be empty", "", nil)
	}

	var config Config
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, errors.NewConfigError("invalid JSON format", "", err)
	}

	// Apply default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = types.DefaultTimeoutSeconds
	}

	// Validate the parsed configuration
	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ToJSON converts a Config struct to JSON
func (c *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// GetUF returns the UF enum from the string representation
func (c *Config) GetUF() (types.UF, error) {
	switch strings.ToUpper(c.SiglaUF) {
	case "AC":
		return types.AC, nil
	case "AL":
		return types.AL, nil
	case "AP":
		return types.AP, nil
	case "AM":
		return types.AM, nil
	case "BA":
		return types.BA, nil
	case "CE":
		return types.CE, nil
	case "DF":
		return types.DF, nil
	case "ES":
		return types.ES, nil
	case "GO":
		return types.GO, nil
	case "MA":
		return types.MA, nil
	case "MT":
		return types.MT, nil
	case "MS":
		return types.MS, nil
	case "MG":
		return types.MG, nil
	case "PA":
		return types.PA, nil
	case "PB":
		return types.PB, nil
	case "PR":
		return types.PR, nil
	case "PE":
		return types.PE, nil
	case "PI":
		return types.PI, nil
	case "RJ":
		return types.RJ, nil
	case "RN":
		return types.RN, nil
	case "RS":
		return types.RS, nil
	case "RO":
		return types.RO, nil
	case "RR":
		return types.RR, nil
	case "SC":
		return types.SC, nil
	case "SP":
		return types.SP, nil
	case "SE":
		return types.SE, nil
	case "TO":
		return types.TO, nil
	case "EX":
		return types.EX, nil
	default:
		return 0, errors.NewConfigError("invalid UF code", "siglaUF", c.SiglaUF)
	}
}

// validateCNPJ validates CNPJ format (basic format validation, not digit verification)
func validateCNPJ(cnpj string) error {
	// Remove non-numeric characters
	cnpjClean := regexp.MustCompile(`[^0-9]`).ReplaceAllString(cnpj, "")
	
	// Check length (11 for CPF, 14 for CNPJ)
	if len(cnpjClean) != 11 && len(cnpjClean) != 14 {
		return fmt.Errorf("CNPJ/CPF must have 11 or 14 digits, got %d", len(cnpjClean))
	}

	// Check if all digits are the same
	if isAllSameDigits(cnpjClean) {
		return fmt.Errorf("CNPJ/CPF cannot have all same digits")
	}

	return nil
}

// validateUF validates if the UF string is valid
func validateUF(uf string) error {
	if len(uf) != 2 {
		return fmt.Errorf("UF must have exactly 2 characters")
	}
	
	validUFs := []string{
		"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
		"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
		"RS", "RO", "RR", "SC", "SP", "SE", "TO", "EX",
	}
	
	ufUpper := strings.ToUpper(uf)
	for _, validUF := range validUFs {
		if ufUpper == validUF {
			return nil
		}
	}
	
	return fmt.Errorf("invalid UF: %s", uf)
}

// validateProxyConfig validates proxy configuration
func validateProxyConfig(proxy *ProxyConfig) error {
	if proxy == nil {
		return nil
	}

	// Validate proxy port if provided
	if proxy.ProxyPort != nil && *proxy.ProxyPort != "" {
		port, err := strconv.Atoi(*proxy.ProxyPort)
		if err != nil {
			return errors.NewConfigError("invalid proxy port format", "proxyPort", *proxy.ProxyPort)
		}
		if port < 1 || port > 65535 {
			return errors.NewConfigError("proxy port must be between 1 and 65535", "proxyPort", port)
		}
	}

	// Validate proxy IP format if provided
	if proxy.ProxyIP != nil && *proxy.ProxyIP != "" {
		// Basic IP format validation (IPv4)
		ipRegex := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
		if !ipRegex.MatchString(*proxy.ProxyIP) {
			return errors.NewConfigError("invalid proxy IP format", "proxyIp", *proxy.ProxyIP)
		}
	}

	return nil
}

// isAllSameDigits checks if all digits in the string are the same
func isAllSameDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	first := s[0]
	for _, char := range s {
		if byte(char) != first {
			return false
		}
	}
	return true
}