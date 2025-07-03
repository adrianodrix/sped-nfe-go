// Package certificate provides A1 certificate support for .pfx/.p12 files.
// A1 certificates are software-based certificates stored in PKCS#12 format.
package certificate

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
	"sync"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"golang.org/x/crypto/pkcs12"
	sslmatePkcs12 "software.sslmate.com/src/go-pkcs12"
)

// A1Certificate represents a software-based certificate (A1 type) loaded from .pfx/.p12 files
type A1Certificate struct {
	certificate *x509.Certificate
	privateKey  *rsa.PrivateKey
	chain       []*x509.Certificate
	config      *Config
	mutex       sync.RWMutex
	
	// Cache for validation results
	lastValidation time.Time
	isValidCached  bool
}

// A1CertificateLoader provides methods to load A1 certificates from various sources
type A1CertificateLoader struct {
	config *Config
}

// NewA1CertificateLoader creates a new A1 certificate loader with the given configuration
func NewA1CertificateLoader(config *Config) *A1CertificateLoader {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &A1CertificateLoader{
		config: config,
	}
}

// LoadFromFile loads an A1 certificate from a .pfx or .p12 file
func (loader *A1CertificateLoader) LoadFromFile(filename, password string) (*A1Certificate, error) {
	if filename == "" {
		return nil, errors.NewValidationError("filename cannot be empty", "filename", "")
	}

	// Read the PKCS#12 file
	p12Data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.NewCertificateError("failed to read certificate file", err)
	}

	return loader.LoadFromBytes(p12Data, password)
}

// LoadFromBytes loads an A1 certificate from PKCS#12 data in memory
func (loader *A1CertificateLoader) LoadFromBytes(p12Data []byte, password string) (*A1Certificate, error) {
	if len(p12Data) == 0 {
		return nil, errors.NewValidationError("certificate data cannot be empty", "p12Data", "")
	}

	// Parse PKCS#12 data - try Decode first, fallback to ToPEM
	privateKey, certificate, err := pkcs12.Decode(p12Data, password)
	var caCerts []*x509.Certificate
	
	if err != nil {
		// If Decode fails, try ToPEM method (works with more PKCS#12 structures)
		blocks, pemErr := pkcs12.ToPEM(p12Data, password)
		if pemErr != nil {
			return nil, errors.NewCertificateError("failed to decode PKCS#12 certificate", err)
		}
		
		// Extract certificate and private key from PEM blocks
		for _, block := range blocks {
			switch block.Type {
			case "CERTIFICATE":
				if certificate == nil { // Use first certificate as main cert
					cert, parseErr := x509.ParseCertificate(block.Bytes)
					if parseErr == nil {
						certificate = cert
					}
				} else {
					// Additional certificates go to CA chain
					cert, parseErr := x509.ParseCertificate(block.Bytes)
					if parseErr == nil {
						caCerts = append(caCerts, cert)
					}
				}
			case "PRIVATE KEY":
				if privateKey == nil { // Use first private key
					key, parseErr := x509.ParsePKCS8PrivateKey(block.Bytes)
					if parseErr != nil {
						// Try PKCS#1 format
						key, parseErr = x509.ParsePKCS1PrivateKey(block.Bytes)
					}
					if parseErr == nil {
						privateKey = key
					}
				}
			}
		}
		
		if certificate == nil || privateKey == nil {
			return nil, errors.NewCertificateError("failed to extract certificate or private key from PKCS#12 data", err)
		}
	}

	// Ensure we have a certificate
	if certificate == nil {
		return nil, errors.NewCertificateError("no certificate found in PKCS#12 data", nil)
	}

	// Ensure we have a private key
	if privateKey == nil {
		return nil, errors.NewCertificateError("no private key found in PKCS#12 data", nil)
	}

	// Ensure it's an RSA private key
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.NewCertificateError("certificate must use RSA private key", nil)
	}

	// Create A1 certificate instance
	a1Cert := &A1Certificate{
		certificate: certificate,
		privateKey:  rsaKey,
		chain:       caCerts,
		config:      loader.config,
	}

	// Validate the certificate if required
	if err := a1Cert.validate(); err != nil {
		return nil, err
	}

	return a1Cert, nil
}

// LoadFromPEM loads an A1 certificate from separate PEM files (certificate and private key)
func (loader *A1CertificateLoader) LoadFromPEM(certPEM, keyPEM []byte, keyPassword string) (*A1Certificate, error) {
	if len(certPEM) == 0 {
		return nil, errors.NewValidationError("certificate PEM data cannot be empty", "certPEM", "")
	}
	if len(keyPEM) == 0 {
		return nil, errors.NewValidationError("private key PEM data cannot be empty", "keyPEM", "")
	}

	// Parse certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, errors.NewCertificateError("failed to decode certificate PEM", nil)
	}

	certificate, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, errors.NewCertificateError("failed to parse certificate", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, errors.NewCertificateError("failed to decode private key PEM", nil)
	}

	var privateKey crypto.PrivateKey
	if keyPassword != "" {
		// Decrypt encrypted private key
		decryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(keyPassword))
		if err != nil {
			return nil, errors.NewCertificateError("failed to decrypt private key", err)
		}
		privateKey, err = x509.ParsePKCS1PrivateKey(decryptedKey)
		if err != nil {
			// Try PKCS#8 format
			privateKey, err = x509.ParsePKCS8PrivateKey(decryptedKey)
			if err != nil {
				return nil, errors.NewCertificateError("failed to parse private key", err)
			}
		}
	} else {
		// Parse unencrypted private key
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			// Try PKCS#8 format
			privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, errors.NewCertificateError("failed to parse private key", err)
			}
		}
	}

	// Ensure it's an RSA private key
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.NewCertificateError("certificate must use RSA private key", nil)
	}

	// Create A1 certificate instance
	a1Cert := &A1Certificate{
		certificate: certificate,
		privateKey:  rsaKey,
		config:      loader.config,
	}

	// Validate the certificate if required
	if err := a1Cert.validate(); err != nil {
		return nil, err
	}

	return a1Cert, nil
}

// Sign signs the given data using the certificate's private key
func (a1 *A1Certificate) Sign(data []byte, algorithm crypto.Hash) ([]byte, error) {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if len(data) == 0 {
		return nil, errors.NewValidationError("data to sign cannot be empty", "data", "")
	}

	if a1.privateKey == nil {
		return nil, errors.NewCertificateError("private key not available", nil)
	}

	// Hash the data
	hasher := algorithm.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, a1.privateKey, algorithm, hashed)
	if err != nil {
		return nil, errors.NewCertificateError("failed to sign data", err)
	}

	return signature, nil
}

// GetPublicKey returns the certificate's public key
func (a1 *A1Certificate) GetPublicKey() crypto.PublicKey {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return nil
	}
	return a1.certificate.PublicKey
}

// GetPrivateKey returns the certificate's private key for TLS authentication
func (a1 *A1Certificate) GetPrivateKey() crypto.PrivateKey {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	return a1.privateKey
}

// GetCertificate returns the X.509 certificate
func (a1 *A1Certificate) GetCertificate() *x509.Certificate {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	return a1.certificate
}

// IsValid checks if the certificate is currently valid (not expired) and uses cache
func (a1 *A1Certificate) IsValid() bool {
	a1.mutex.RLock()
	
	// Check cache validity
	if time.Since(a1.lastValidation) < a1.config.CacheTimeout {
		result := a1.isValidCached
		a1.mutex.RUnlock()
		return result
	}
	
	a1.mutex.RUnlock()
	
	// Update cache
	a1.mutex.Lock()
	defer a1.mutex.Unlock()
	
	a1.isValidCached = IsCertificateValidForSigning(a1.certificate)
	a1.lastValidation = time.Now()
	
	return a1.isValidCached
}

// GetSubject returns the certificate subject as a formatted string
func (a1 *A1Certificate) GetSubject() string {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return ""
	}
	return a1.certificate.Subject.String()
}

// GetIssuer returns the certificate issuer as a formatted string
func (a1 *A1Certificate) GetIssuer() string {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return ""
	}
	return a1.certificate.Issuer.String()
}

// GetSerialNumber returns the certificate serial number as a string
func (a1 *A1Certificate) GetSerialNumber() string {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return ""
	}
	return a1.certificate.SerialNumber.String()
}

// GetFingerprint returns the SHA-256 fingerprint of the certificate
func (a1 *A1Certificate) GetFingerprint() string {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	return getCertificateFingerprint(a1.certificate)
}

// GetValidityPeriod returns the certificate's not before and not after dates
func (a1 *A1Certificate) GetValidityPeriod() (notBefore, notAfter time.Time) {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return time.Time{}, time.Time{}
	}
	return a1.certificate.NotBefore, a1.certificate.NotAfter
}

// GetCertificateChain returns the certificate chain (CA certificates)
func (a1 *A1Certificate) GetCertificateChain() []*x509.Certificate {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	// Return a copy to prevent external modification
	chain := make([]*x509.Certificate, len(a1.chain))
	copy(chain, a1.chain)
	return chain
}

// GetInfo returns basic information about the certificate
func (a1 *A1Certificate) GetInfo() *CertificateInfo {
	return GetCertificateInfo(a1.GetCertificate(), TypeA1)
}

// Close releases any resources associated with the certificate
func (a1 *A1Certificate) Close() error {
	a1.mutex.Lock()
	defer a1.mutex.Unlock()

	// Clear sensitive data
	if a1.privateKey != nil {
		// Zero out private key data (security best practice)
		*a1.privateKey = rsa.PrivateKey{}
		a1.privateKey = nil
	}

	a1.certificate = nil
	a1.chain = nil
	
	return nil
}

// ExportToPKCS12 exports the certificate and private key to PKCS#12 format
// This creates a new PKCS#12 container with the certificate and private key
func (a1 *A1Certificate) ExportToPKCS12(password string) ([]byte, error) {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil || a1.privateKey == nil {
		return nil, errors.NewCertificateError("certificate or private key not available", nil)
	}

	if password == "" {
		return nil, errors.NewValidationError("password cannot be empty for PKCS#12 export", "password", "")
	}

	// Use software.sslmate.com/src/go-pkcs12 for encoding
	// This library supports both decoding and encoding PKCS#12
	
	// Create a chain with the certificate and any intermediate certificates
	var chain []*x509.Certificate
	chain = append(chain, a1.certificate)
	
	// Add intermediate certificates if available
	if a1.chain != nil && len(a1.chain) > 0 {
		chain = append(chain, a1.chain...)
	}

	// Encode to PKCS#12 using sslmate library which supports encoding
	// Create a certificate chain starting from intermediate certificates (excluding the main cert)
	var caCerts []*x509.Certificate
	if len(chain) > 1 {
		caCerts = chain[1:]
	}
	
	p12Data, err := sslmatePkcs12.Encode(rand.Reader, a1.privateKey, a1.certificate, caCerts, password)
	if err != nil {
		return nil, errors.NewCertificateError("failed to encode PKCS#12", err)
	}

	return p12Data, nil
}

// ValidateICPBrasilChain validates the certificate chain against ICP-Brasil root CAs
func (a1 *A1Certificate) ValidateICPBrasilChain() error {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return errors.NewCertificateError("certificate not available", nil)
	}

	// Create certificate chain for validation
	var chain []*x509.Certificate
	chain = append(chain, a1.certificate)
	if a1.chain != nil {
		chain = append(chain, a1.chain...)
	}

	// Validate chain against ICP-Brasil root certificates
	return ValidateICPBrasilCertificateChain(chain)
}


// IsICPBrasilCertificate checks if the certificate is from ICP-Brasil
func (a1 *A1Certificate) IsICPBrasilCertificate() bool {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return false
	}

	return IsICPBrasilCertificate(a1.certificate)
}

// ExportCertificateToPEM exports only the certificate to PEM format
func (a1 *A1Certificate) ExportCertificateToPEM() ([]byte, error) {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return nil, errors.NewCertificateError("certificate not available", nil)
	}

	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: a1.certificate.Raw,
	}

	return pem.EncodeToMemory(pemBlock), nil
}

// validate performs validation of the loaded certificate
func (a1 *A1Certificate) validate() error {
	if a1.certificate == nil {
		return errors.NewCertificateError("certificate is nil", nil)
	}

	if a1.privateKey == nil {
		return errors.NewCertificateError("private key is nil", nil)
	}

	// Validate that private key matches certificate
	certPubKey, ok := a1.certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.NewCertificateError("certificate public key is not RSA", nil)
	}

	if certPubKey.N.Cmp(a1.privateKey.N) != 0 {
		return errors.NewCertificateError("private key does not match certificate", nil)
	}

	// Use common validation
	return ValidateCertificate(a1.certificate, a1.config)
}

// VerifySignature verifies a signature against the certificate's public key
func (a1 *A1Certificate) VerifySignature(data, signature []byte, algorithm crypto.Hash) error {
	a1.mutex.RLock()
	defer a1.mutex.RUnlock()

	if a1.certificate == nil {
		return errors.NewCertificateError("certificate not available", nil)
	}

	pubKey, ok := a1.certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.NewCertificateError("certificate public key is not RSA", nil)
	}

	// Hash the data
	hasher := algorithm.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	// Verify signature
	err := rsa.VerifyPKCS1v15(pubKey, algorithm, hashed, signature)
	if err != nil {
		return errors.NewCertificateError("signature verification failed", err)
	}

	return nil
}

// LoadA1FromFile is a convenience function to load an A1 certificate from file
func LoadA1FromFile(filename, password string) (*A1Certificate, error) {
	loader := NewA1CertificateLoader(DefaultConfig())
	return loader.LoadFromFile(filename, password)
}

// LoadA1FromBytes is a convenience function to load an A1 certificate from bytes
func LoadA1FromBytes(p12Data []byte, password string) (*A1Certificate, error) {
	loader := NewA1CertificateLoader(DefaultConfig())
	return loader.LoadFromBytes(p12Data, password)
}

// LoadA1FromPEM is a convenience function to load an A1 certificate from PEM data
func LoadA1FromPEM(certPEM, keyPEM []byte, keyPassword string) (*A1Certificate, error) {
	loader := NewA1CertificateLoader(DefaultConfig())
	return loader.LoadFromPEM(certPEM, keyPEM, keyPassword)
}

// CreateA1FromReader loads an A1 certificate from any io.Reader
func CreateA1FromReader(reader io.Reader, password string) (*A1Certificate, error) {
	if reader == nil {
		return nil, errors.NewValidationError("reader cannot be nil", "reader", "")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.NewCertificateError("failed to read certificate data", err)
	}

	return LoadA1FromBytes(data, password)
}