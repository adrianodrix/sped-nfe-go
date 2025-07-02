// Package certificate provides digital certificate management for NFe signing.
// It supports both A1 (.pfx/.p12) and A3 (PKCS#11) certificates from ICP-Brasil.
package certificate

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
)

// Certificate represents a digital certificate interface for A1 and A3 certificates.
// It provides common operations needed for NFe digital signing.
type Certificate interface {
	// Sign signs the given data using the certificate's private key
	Sign(data []byte, algorithm crypto.Hash) ([]byte, error)
	
	// GetPublicKey returns the certificate's public key
	GetPublicKey() crypto.PublicKey
	
	// GetCertificate returns the X.509 certificate
	GetCertificate() *x509.Certificate
	
	// IsValid checks if the certificate is currently valid (not expired)
	IsValid() bool
	
	// GetSubject returns the certificate subject as a formatted string
	GetSubject() string
	
	// GetIssuer returns the certificate issuer as a formatted string
	GetIssuer() string
	
	// GetSerialNumber returns the certificate serial number as a string
	GetSerialNumber() string
	
	// GetFingerprint returns the SHA-256 fingerprint of the certificate
	GetFingerprint() string
	
	// GetValidityPeriod returns the certificate's not before and not after dates
	GetValidityPeriod() (notBefore, notAfter time.Time)
	
	// Close releases any resources associated with the certificate
	Close() error
}

// CertificateType represents the type of certificate (A1 or A3)
type CertificateType int

const (
	// TypeA1 represents software certificates stored in .pfx/.p12 files
	TypeA1 CertificateType = iota
	// TypeA3 represents hardware certificates in tokens/smart cards
	TypeA3
)

func (ct CertificateType) String() string {
	switch ct {
	case TypeA1:
		return "A1"
	case TypeA3:
		return "A3"
	default:
		return "Unknown"
	}
}

// CertificateInfo holds basic information about a certificate
type CertificateInfo struct {
	Type         CertificateType `json:"type"`
	Subject      string          `json:"subject"`
	Issuer       string          `json:"issuer"`
	SerialNumber string          `json:"serialNumber"`
	Fingerprint  string          `json:"fingerprint"`
	NotBefore    time.Time       `json:"notBefore"`
	NotAfter     time.Time       `json:"notAfter"`
	IsValid      bool            `json:"isValid"`
}

// Config holds configuration for certificate operations
type Config struct {
	// ValidateChain enables certificate chain validation against ICP-Brasil root CAs
	ValidateChain bool `json:"validateChain"`
	
	// AllowExpired allows loading of expired certificates (for testing)
	AllowExpired bool `json:"allowExpired"`
	
	// CacheTimeout defines how long to cache certificate validation results
	CacheTimeout time.Duration `json:"cacheTimeout"`
	
	// RequiredKeyUsage defines required key usage extensions
	RequiredKeyUsage x509.KeyUsage `json:"requiredKeyUsage"`
	
	// RequiredEKU defines required extended key usage extensions
	RequiredEKU []x509.ExtKeyUsage `json:"requiredEKU"`
}

// DefaultConfig returns a default configuration for ICP-Brasil certificates
func DefaultConfig() *Config {
	return &Config{
		ValidateChain:    true,
		AllowExpired:     false,
		CacheTimeout:     5 * time.Minute,
		RequiredKeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		RequiredEKU:      []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}

// ValidateCertificate performs comprehensive validation of an X.509 certificate
func ValidateCertificate(cert *x509.Certificate, config *Config) error {
	if cert == nil {
		return errors.NewValidationError("certificate cannot be nil", "certificate", "")
	}

	if config == nil {
		config = DefaultConfig()
	}

	now := time.Now()

	// Check validity period
	if !config.AllowExpired {
		if now.Before(cert.NotBefore) {
			return errors.NewCertificateError("certificate not yet valid", nil)
		}
		if now.After(cert.NotAfter) {
			return errors.NewCertificateError("certificate has expired", nil)
		}
	}

	// Check key usage
	if config.RequiredKeyUsage != 0 {
		if cert.KeyUsage&config.RequiredKeyUsage == 0 {
			return errors.NewCertificateError("certificate missing required key usage", nil)
		}
	}

	// Check extended key usage
	if len(config.RequiredEKU) > 0 {
		found := false
		for _, required := range config.RequiredEKU {
			for _, usage := range cert.ExtKeyUsage {
				if usage == required {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return errors.NewCertificateError("certificate missing required extended key usage", nil)
		}
	}

	// Validate that it's an RSA certificate with adequate key size
	if rsaPubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		if rsaPubKey.Size() < 256 { // Less than 2048 bits
			return errors.NewCertificateError("RSA key size too small, minimum 2048 bits required", nil)
		}
	} else {
		return errors.NewCertificateError("certificate must use RSA keys", nil)
	}

	return nil
}

// GetCertificateInfo extracts basic information from an X.509 certificate
func GetCertificateInfo(cert *x509.Certificate, certType CertificateType) *CertificateInfo {
	if cert == nil {
		return nil
	}

	return &CertificateInfo{
		Type:         certType,
		Subject:      cert.Subject.String(),
		Issuer:       cert.Issuer.String(),
		SerialNumber: cert.SerialNumber.String(),
		Fingerprint:  getCertificateFingerprint(cert),
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		IsValid:      time.Now().After(cert.NotBefore) && time.Now().Before(cert.NotAfter),
	}
}

// getCertificateFingerprint calculates the SHA-256 fingerprint of a certificate
func getCertificateFingerprint(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	
	// Use the same fingerprint calculation from soap/security.go
	hash := crypto.SHA256.New()
	hash.Write(cert.Raw)
	fingerprint := hash.Sum(nil)
	
	// Format as uppercase hex with colons
	result := make([]byte, 0, len(fingerprint)*3-1)
	for i, b := range fingerprint {
		if i > 0 {
			result = append(result, ':')
		}
		result = append(result, "0123456789ABCDEF"[b>>4])
		result = append(result, "0123456789ABCDEF"[b&15])
	}
	
	return string(result)
}

// IsCertificateExpired checks if a certificate is expired
func IsCertificateExpired(cert *x509.Certificate) bool {
	if cert == nil {
		return true
	}
	return time.Now().After(cert.NotAfter)
}

// IsCertificateValidForSigning checks if a certificate can be used for digital signing
func IsCertificateValidForSigning(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}

	// Check if certificate is valid time-wise
	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return false
	}

	// Check digital signature key usage
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return false
	}

	// Must be RSA key
	if _, ok := cert.PublicKey.(*rsa.PublicKey); !ok {
		return false
	}

	return true
}

// GetCertificateKeySize returns the key size in bits for RSA certificates
func GetCertificateKeySize(cert *x509.Certificate) int {
	if cert == nil {
		return 0
	}

	if rsaPubKey, ok := cert.PublicKey.(*rsa.PublicKey); ok {
		return rsaPubKey.Size() * 8 // Convert bytes to bits
	}

	return 0
}

// ParseCertificateFromDER parses a DER-encoded certificate
func ParseCertificateFromDER(der []byte) (*x509.Certificate, error) {
	if len(der) == 0 {
		return nil, errors.NewValidationError("certificate data cannot be empty", "der", "")
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, errors.NewCertificateError("failed to parse certificate", err)
	}

	return cert, nil
}

// ParseCertificatesFromDER parses multiple DER-encoded certificates
func ParseCertificatesFromDER(der []byte) ([]*x509.Certificate, error) {
	if len(der) == 0 {
		return nil, errors.NewValidationError("certificate data cannot be empty", "der", "")
	}

	certs, err := x509.ParseCertificates(der)
	if err != nil {
		return nil, errors.NewCertificateError("failed to parse certificates", err)
	}

	return certs, nil
}