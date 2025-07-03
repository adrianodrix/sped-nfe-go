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
	AC UF = 12 // Acre
	AL UF = 27 // Alagoas
	AM UF = 13 // Amazonas
	AP UF = 16 // Amapá
	BA UF = 29 // Bahia
	CE UF = 23 // Ceará
	DF UF = 53 // Distrito Federal
	ES UF = 32 // Espírito Santo
	GO UF = 52 // Goiás
	MA UF = 21 // Maranhão
	MG UF = 31 // Minas Gerais
	MS UF = 50 // Mato Grosso do Sul
	MT UF = 51 // Mato Grosso
	PA UF = 15 // Pará
	PB UF = 25 // Paraíba
	PE UF = 26 // Pernambuco
	PI UF = 22 // Piauí
	PR UF = 41 // Paraná
	RJ UF = 33 // Rio de Janeiro
	RN UF = 24 // Rio Grande do Norte
	RO UF = 11 // Rondônia
	RR UF = 14 // Roraima
	RS UF = 43 // Rio Grande do Sul
	SC UF = 42 // Santa Catarina
	SE UF = 28 // Sergipe
	SP UF = 35 // São Paulo
	TO UF = 17 // Tocantins

	// Virtual SEFAZs and National Agencies
	AN    UF = 91 // Ambiente Nacional
	SVAN  UF = 92 // SEFAZ Virtual do Ambiente Nacional
	SVRS  UF = 93 // SEFAZ Virtual do Rio Grande do Sul
	SVCAN UF = 94 // SEFAZ Virtual de Contingência do Ambiente Nacional
	SVCRS UF = 95 // SEFAZ Virtual de Contingência do Rio Grande do Sul
)

// String returns the string representation of UF
func (u UF) String() string {
	switch u {
	case AC:
		return "AC"
	case AL:
		return "AL"
	case AM:
		return "AM"
	case AP:
		return "AP"
	case BA:
		return "BA"
	case CE:
		return "CE"
	case DF:
		return "DF"
	case ES:
		return "ES"
	case GO:
		return "GO"
	case MA:
		return "MA"
	case MG:
		return "MG"
	case MS:
		return "MS"
	case MT:
		return "MT"
	case PA:
		return "PA"
	case PB:
		return "PB"
	case PE:
		return "PE"
	case PI:
		return "PI"
	case PR:
		return "PR"
	case RJ:
		return "RJ"
	case RN:
		return "RN"
	case RO:
		return "RO"
	case RR:
		return "RR"
	case RS:
		return "RS"
	case SC:
		return "SC"
	case SE:
		return "SE"
	case SP:
		return "SP"
	case TO:
		return "TO"
	case AN:
		return "AN"
	case SVAN:
		return "SVAN"
	case SVRS:
		return "SVRS"
	case SVCAN:
		return "SVCAN"
	case SVCRS:
		return "SVCRS"
	default:
		return "UNKNOWN"
	}
}

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
