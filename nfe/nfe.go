package nfe

// Package nfe provides functionality for Brazilian Electronic Invoice (NFe) generation, digital signature, and transmission to SEFAZ webservices.
//
// This package is a work in progress and is not yet ready for production use.
// The package is licensed under the MIT license.
// The package is maintained by Adrianodrix.
//
// The package is maintained by Adrianodrix.

import (
	"fmt"
	"time"
)

// Version represents the current version of the sped-nfe-go package
const Version = "0.1.0"

// Environment represents the SEFAZ environment
type Environment int

const (
	// Production environment
	Production Environment = 1
	// Homologation environment for testing
	Homologation Environment = 2
)

// UF represents Brazilian states
type UF int

const (
	SP UF = 35 // São Paulo
	RJ UF = 33 // Rio de Janeiro
	MG UF = 31 // Minas Gerais
	// Add more states as needed
)

// Config holds configuration for NFe operations
type Config struct {
	Environment Environment `json:"environment"`
	UF          UF          `json:"uf"`
	Timeout     int         `json:"timeout"`
}

// Client represents the main NFe client
type Client struct {
	config Config
}

// New creates a new NFe client with the given configuration
func New(config Config) (*Client, error) {
	if config.Timeout <= 0 {
		config.Timeout = 30 // default 30 seconds
	}

	return &Client{
		config: config,
	}, nil
}

// GetVersion returns the current package version
func GetVersion() string {
	return Version
}

// IsProduction returns true if the client is configured for production environment
func (c *Client) IsProduction() bool {
	return c.config.Environment == Production
}

// GetConfig returns the client configuration
func (c *Client) GetConfig() Config {
	return c.config
}

// GenerateAccessKey generates a 44-digit access key for NFe
// This is a simplified version - will be enhanced later
func (c *Client) GenerateAccessKey(cnpj string, modelo, serie, numero int, tipoEmissao int) string {
	// Simplified implementation for now
	uf := int(c.config.UF)
	year := time.Now().Year() % 100 // last 2 digits
	month := int(time.Now().Month())
	
	// Format: UF + YYMM + CNPJ + MODELO + SERIE + NUMERO + TIPO + COD + DV
	// This is a placeholder implementation
	return fmt.Sprintf("%02d%02d%02d%s%02d%03d%09d%d%08d%d",
		uf, year, month, cnpj[:14], modelo, serie, numero, tipoEmissao, 12345678, 9)
}

// ValidateConfig validates the client configuration
func ValidateConfig(config Config) error {
	if config.Environment != Production && config.Environment != Homologation {
		return fmt.Errorf("invalid environment: must be Production (1) or Homologation (2)")
	}
	
	if config.UF <= 0 {
		return fmt.Errorf("invalid UF: must be a valid state code")
	}
	
	return nil
}