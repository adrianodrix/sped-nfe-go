// Package certificate provides A3 certificate support for PKCS#11 tokens and smart cards.
// A3 certificates are hardware-based certificates stored in cryptographic tokens.
package certificate

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/ThalesIgnite/crypto11"
)

// A3Certificate represents a hardware-based certificate (A3 type) stored in PKCS#11 tokens
type A3Certificate struct {
	certificate *x509.Certificate
	privateKey  crypto11.Signer
	context     *crypto11.Context
	slot        uint
	tokenLabel  string
	config      *Config
	mutex       sync.RWMutex
	
	// Cache for validation results
	lastValidation time.Time
	isValidCached  bool
}

// A3CertificateLoader provides methods to load A3 certificates from PKCS#11 tokens
type A3CertificateLoader struct {
	config *Config
}

// PKCS11Config holds PKCS#11 specific configuration
type PKCS11Config struct {
	// Path to the PKCS#11 library (.so/.dll file)
	LibraryPath string `json:"libraryPath"`
	
	// Token label to use (empty to use first available)
	TokenLabel string `json:"tokenLabel"`
	
	// Token PIN for authentication
	PIN string `json:"pin"`
	
	// Slot number (optional, will search all slots if not specified)
	Slot *uint `json:"slot,omitempty"`
	
	// Certificate label or ID to search for
	CertificateLabel string `json:"certificateLabel"`
	CertificateID    []byte `json:"certificateId"`
}

// TokenInfo holds information about a PKCS#11 token
type TokenInfo struct {
	Slot        uint   `json:"slot"`
	Label       string `json:"label"`
	Manufacturer string `json:"manufacturer"`
	Model       string `json:"model"`
	SerialNumber string `json:"serialNumber"`
	IsPresent   bool   `json:"isPresent"`
	HasTokens   bool   `json:"hasTokens"`
}

// NewA3CertificateLoader creates a new A3 certificate loader with the given configuration
func NewA3CertificateLoader(config *Config) *A3CertificateLoader {
	if config == nil {
		config = DefaultConfig()
	}
	
	return &A3CertificateLoader{
		config: config,
	}
}

// LoadFromToken loads an A3 certificate from a PKCS#11 token
func (loader *A3CertificateLoader) LoadFromToken(pkcs11Config *PKCS11Config) (*A3Certificate, error) {
	if pkcs11Config == nil {
		return nil, errors.NewValidationError("PKCS11 configuration cannot be nil", "pkcs11Config", "")
	}

	if pkcs11Config.LibraryPath == "" {
		return nil, errors.NewValidationError("PKCS11 library path cannot be empty", "libraryPath", "")
	}

	// Create PKCS#11 configuration for crypto11
	crypto11Config := &crypto11.Config{
		Path:       pkcs11Config.LibraryPath,
		TokenLabel: pkcs11Config.TokenLabel,
		Pin:        pkcs11Config.PIN,
	}

	if pkcs11Config.Slot != nil {
		crypto11Config.SlotNumber = pkcs11Config.Slot
	}

	// Initialize PKCS#11 context
	context, err := crypto11.Configure(crypto11Config)
	if err != nil {
		return nil, errors.NewCertificateError("failed to initialize PKCS#11 context", err)
	}

	// Find certificate in the token
	var certificate *x509.Certificate
	var privateKey crypto11.Signer

	if pkcs11Config.CertificateLabel != "" {
		// Find by label
		certificate, err = context.FindCertificate(nil, []byte(pkcs11Config.CertificateLabel), nil)
		if err != nil {
			context.Close()
			return nil, errors.NewCertificateError("failed to find certificate by label", err)
		}
	} else if len(pkcs11Config.CertificateID) > 0 {
		// Find by ID
		certificate, err = context.FindCertificate(pkcs11Config.CertificateID, nil, nil)
		if err != nil {
			context.Close()
			return nil, errors.NewCertificateError("failed to find certificate by ID", err)
		}
	} else {
		// Find first available certificate
		certificates, err := context.FindAllCertificates()
		if err != nil {
			context.Close()
			return nil, errors.NewCertificateError("failed to enumerate certificates", err)
		}

		if len(certificates) == 0 {
			context.Close()
			return nil, errors.NewCertificateError("no certificates found in token", nil)
		}

		certificate = certificates[0]
	}

	if certificate == nil {
		context.Close()
		return nil, errors.NewCertificateError("certificate not found in token", nil)
	}

	// Find corresponding private key
	if pkcs11Config.CertificateLabel != "" {
		privateKey, err = context.FindKeyPair(nil, []byte(pkcs11Config.CertificateLabel))
	} else if len(pkcs11Config.CertificateID) > 0 {
		privateKey, err = context.FindKeyPair(pkcs11Config.CertificateID, nil)
	} else {
		// Try to find private key using certificate's public key
		privateKey, err = loader.findPrivateKeyForCertificate(context, certificate)
	}

	if err != nil || privateKey == nil {
		context.Close()
		return nil, errors.NewCertificateError("failed to find private key for certificate", err)
	}

	// Create A3 certificate instance
	a3Cert := &A3Certificate{
		certificate: certificate,
		privateKey:  privateKey,
		context:     context,
		tokenLabel:  pkcs11Config.TokenLabel,
		config:      loader.config,
	}

	// Validate the certificate if required
	if err := a3Cert.validate(); err != nil {
		context.Close()
		return nil, err
	}

	return a3Cert, nil
}

// findPrivateKeyForCertificate attempts to find a private key that matches the given certificate
func (loader *A3CertificateLoader) findPrivateKeyForCertificate(context *crypto11.Context, cert *x509.Certificate) (crypto11.Signer, error) {
	// Get all key pairs and try to match with certificate
	keyPairs, err := context.FindAllKeyPairs()
	if err != nil {
		return nil, err
	}

	certPubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.NewCertificateError("certificate public key is not RSA", nil)
	}

	// Try to match public keys
	for _, keyPair := range keyPairs {
		if rsaPubKey, ok := keyPair.Public().(*rsa.PublicKey); ok {
			if rsaPubKey.N.Cmp(certPubKey.N) == 0 {
				return keyPair, nil
			}
		}
	}

	return nil, errors.NewCertificateError("no matching private key found for certificate", nil)
}

// ListTokens returns information about available PKCS#11 tokens
func ListTokens(libraryPath string) ([]TokenInfo, error) {
	if libraryPath == "" {
		return nil, errors.NewValidationError("library path cannot be empty", "libraryPath", "")
	}

	// Create temporary context to list tokens
	config := &crypto11.Config{
		Path: libraryPath,
	}

	context, err := crypto11.Configure(config)
	if err != nil {
		return nil, errors.NewCertificateError("failed to initialize PKCS#11 for token listing", err)
	}
	defer context.Close()

	// Get token information - this is a simplified implementation
	// In a real implementation, you would use the PKCS#11 API directly to get detailed token info
	tokens := []TokenInfo{
		{
			Slot:        0,
			Label:       "Token Slot 0",
			Manufacturer: "Unknown",
			Model:       "Unknown",
			SerialNumber: "Unknown",
			IsPresent:   true,
			HasTokens:   true,
		},
	}

	return tokens, nil
}

// Sign signs the given data using the certificate's private key in the hardware token
func (a3 *A3Certificate) Sign(data []byte, algorithm crypto.Hash) ([]byte, error) {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if len(data) == 0 {
		return nil, errors.NewValidationError("data to sign cannot be empty", "data", "")
	}

	if a3.privateKey == nil {
		return nil, errors.NewCertificateError("private key not available", nil)
	}

	// Hash the data
	hasher := algorithm.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	// Sign using the hardware token
	signature, err := a3.privateKey.Sign(nil, hashed, algorithm)
	if err != nil {
		return nil, errors.NewCertificateError("failed to sign data with hardware token", err)
	}

	return signature, nil
}

// GetPublicKey returns the certificate's public key
func (a3 *A3Certificate) GetPublicKey() crypto.PublicKey {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if a3.certificate == nil {
		return nil
	}
	return a3.certificate.PublicKey
}

// GetCertificate returns the X.509 certificate
func (a3 *A3Certificate) GetCertificate() *x509.Certificate {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	return a3.certificate
}

// IsValid checks if the certificate is currently valid (not expired) and uses cache
func (a3 *A3Certificate) IsValid() bool {
	a3.mutex.RLock()
	
	// Check cache validity
	if time.Since(a3.lastValidation) < a3.config.CacheTimeout {
		result := a3.isValidCached
		a3.mutex.RUnlock()
		return result
	}
	
	a3.mutex.RUnlock()
	
	// Update cache
	a3.mutex.Lock()
	defer a3.mutex.Unlock()
	
	a3.isValidCached = IsCertificateValidForSigning(a3.certificate)
	a3.lastValidation = time.Now()
	
	return a3.isValidCached
}

// GetSubject returns the certificate subject as a formatted string
func (a3 *A3Certificate) GetSubject() string {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if a3.certificate == nil {
		return ""
	}
	return a3.certificate.Subject.String()
}

// GetIssuer returns the certificate issuer as a formatted string
func (a3 *A3Certificate) GetIssuer() string {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if a3.certificate == nil {
		return ""
	}
	return a3.certificate.Issuer.String()
}

// GetSerialNumber returns the certificate serial number as a string
func (a3 *A3Certificate) GetSerialNumber() string {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if a3.certificate == nil {
		return ""
	}
	return a3.certificate.SerialNumber.String()
}

// GetFingerprint returns the SHA-256 fingerprint of the certificate
func (a3 *A3Certificate) GetFingerprint() string {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	return getCertificateFingerprint(a3.certificate)
}

// GetValidityPeriod returns the certificate's not before and not after dates
func (a3 *A3Certificate) GetValidityPeriod() (notBefore, notAfter time.Time) {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if a3.certificate == nil {
		return time.Time{}, time.Time{}
	}
	return a3.certificate.NotBefore, a3.certificate.NotAfter
}

// GetTokenLabel returns the label of the token containing this certificate
func (a3 *A3Certificate) GetTokenLabel() string {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	return a3.tokenLabel
}

// GetSlot returns the slot number of the token containing this certificate
func (a3 *A3Certificate) GetSlot() uint {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	return a3.slot
}

// GetInfo returns basic information about the certificate
func (a3 *A3Certificate) GetInfo() *CertificateInfo {
	return GetCertificateInfo(a3.GetCertificate(), TypeA3)
}

// Close releases any resources associated with the certificate and closes PKCS#11 session
func (a3 *A3Certificate) Close() error {
	a3.mutex.Lock()
	defer a3.mutex.Unlock()

	var err error
	
	// Close PKCS#11 context
	if a3.context != nil {
		err = a3.context.Close()
		a3.context = nil
	}

	// Clear references (private key is handled by crypto11)
	a3.certificate = nil
	a3.privateKey = nil
	
	return err
}

// IsTokenPresent checks if the hardware token is still present and accessible
func (a3 *A3Certificate) IsTokenPresent() bool {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if a3.context == nil {
		return false
	}

	// Try to access the token by finding the certificate again
	_, err := a3.context.FindCertificate(nil, nil, nil)
	return err == nil
}

// TestConnection tests if the certificate can be used for signing (basic connectivity test)
func (a3 *A3Certificate) TestConnection() error {
	testData := []byte("test data for connection check")
	_, err := a3.Sign(testData, crypto.SHA256)
	return err
}

// validate performs validation of the loaded certificate
func (a3 *A3Certificate) validate() error {
	if a3.certificate == nil {
		return errors.NewCertificateError("certificate is nil", nil)
	}

	if a3.privateKey == nil {
		return errors.NewCertificateError("private key is nil", nil)
	}

	if a3.context == nil {
		return errors.NewCertificateError("PKCS#11 context is nil", nil)
	}

	// Validate that private key matches certificate
	certPubKey, ok := a3.certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.NewCertificateError("certificate public key is not RSA", nil)
	}

	tokenPubKey, ok := a3.privateKey.Public().(*rsa.PublicKey)
	if !ok {
		return errors.NewCertificateError("token private key is not RSA", nil)
	}

	if certPubKey.N.Cmp(tokenPubKey.N) != 0 {
		return errors.NewCertificateError("private key does not match certificate", nil)
	}

	// Use common validation
	return ValidateCertificate(a3.certificate, a3.config)
}

// VerifySignature verifies a signature against the certificate's public key
func (a3 *A3Certificate) VerifySignature(data, signature []byte, algorithm crypto.Hash) error {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()

	if a3.certificate == nil {
		return errors.NewCertificateError("certificate not available", nil)
	}

	pubKey, ok := a3.certificate.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.NewCertificateError("certificate public key is not RSA", nil)
	}

	// Hash the data
	hasher := algorithm.New()
	hasher.Write(data)
	hashed := hasher.Sum(nil)

	// Verify signature using standard RSA verification
	err := rsa.VerifyPKCS1v15(pubKey, algorithm, hashed, signature)
	if err != nil {
		return errors.NewCertificateError("signature verification failed", err)
	}

	return nil
}

// LoadA3FromToken is a convenience function to load an A3 certificate from a token
func LoadA3FromToken(pkcs11Config *PKCS11Config) (*A3Certificate, error) {
	loader := NewA3CertificateLoader(DefaultConfig())
	return loader.LoadFromToken(pkcs11Config)
}

// GetAvailableTokens is a convenience function to list available PKCS#11 tokens
func GetAvailableTokens(libraryPath string) ([]TokenInfo, error) {
	return ListTokens(libraryPath)
}

// Common PKCS#11 library paths for different operating systems
var (
	// Windows common paths
	WindowsPKCS11Paths = []string{
		"C:\\Windows\\System32\\eToken.dll",
		"C:\\Windows\\System32\\cryptoki.dll",
		"C:\\Windows\\System32\\sadaptor.dll",
		"C:\\Program Files\\SafeNet\\Authentication\\SAC\\x64\\sadaptor.dll",
	}

	// Linux common paths
	LinuxPKCS11Paths = []string{
		"/usr/lib/x86_64-linux-gnu/opensc-pkcs11.so",
		"/usr/lib/opensc-pkcs11.so",
		"/usr/lib64/opensc-pkcs11.so",
		"/usr/lib/pkcs11/opensc-pkcs11.so",
		"/usr/local/lib/opensc-pkcs11.so",
	}

	// macOS common paths
	MacOSPKCS11Paths = []string{
		"/usr/local/lib/opensc-pkcs11.so",
		"/System/Library/Frameworks/PCSC.framework/Versions/A/Libraries/libykcs11.dylib",
		"/usr/lib/opensc-pkcs11.so",
	}
)

// GetCommonPKCS11Paths returns common PKCS#11 library paths for the current OS
func GetCommonPKCS11Paths() []string {
	// In a real implementation, you would detect the OS and return appropriate paths
	// For now, return all paths
	var allPaths []string
	allPaths = append(allPaths, WindowsPKCS11Paths...)
	allPaths = append(allPaths, LinuxPKCS11Paths...)
	allPaths = append(allPaths, MacOSPKCS11Paths...)
	return allPaths
}

// FindPKCS11Library attempts to find a working PKCS#11 library automatically
func FindPKCS11Library() (string, error) {
	paths := GetCommonPKCS11Paths()
	
	for _, path := range paths {
		// Try to load the library (simplified check)
		config := &crypto11.Config{Path: path}
		context, err := crypto11.Configure(config)
		if err == nil {
			context.Close()
			return path, nil
		}
	}
	
	return "", errors.NewCertificateError("no working PKCS#11 library found", nil)
}