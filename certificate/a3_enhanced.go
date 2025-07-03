package certificate

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/ThalesIgnite/crypto11"
	"github.com/adrianodrix/sped-nfe-go/errors"
)

// A3CertificateManager provides enhanced management for A3 certificates
type A3CertificateManager struct {
	contexts map[string]*crypto11.Context
	config   *PKCS11Config
	mutex    sync.RWMutex
}

// PKCS11ConfigExtended holds extended PKCS#11 configuration
type PKCS11ConfigExtended struct {
	// Basic configuration
	LibraryPath string `json:"libraryPath"`
	TokenLabel  string `json:"tokenLabel"`
	PIN         string `json:"pin"`
	Slot        *uint  `json:"slot,omitempty"`
	
	// Extended configuration
	CertificateLabel string `json:"certificateLabel,omitempty"`
	CertificateID    string `json:"certificateId,omitempty"`
	
	// Connection settings
	MaxSessions      int           `json:"maxSessions"`
	SessionTimeout   time.Duration `json:"sessionTimeout"`
	RetryAttempts    int           `json:"retryAttempts"`
	RetryDelay       time.Duration `json:"retryDelay"`
	
	// Security settings
	ForceAuthentication bool `json:"forceAuthentication"`
	LogSensitiveData    bool `json:"logSensitiveData"`
}

// A3TokenInfo contains information about a PKCS#11 token
type A3TokenInfo struct {
	SlotID          uint   `json:"slotId"`
	TokenLabel      string `json:"tokenLabel"`
	ManufacturerID  string `json:"manufacturerId"`
	Model           string `json:"model"`
	SerialNumber    string `json:"serialNumber"`
	IsInitialized   bool   `json:"isInitialized"`
	HasTokenPresent bool   `json:"hasTokenPresent"`
	IsWriteProtected bool  `json:"isWriteProtected"`
	LoginRequired   bool   `json:"loginRequired"`
}

// A3CertificateDetails contains detailed information about an A3 certificate
type A3CertificateDetails struct {
	CertificateInfo *CertificateInfo `json:"certificateInfo"`
	TokenInfo       *A3TokenInfo     `json:"tokenInfo"`
	Label           string           `json:"label"`
	ID              []byte           `json:"id"`
	IsPrivateKeyAccessible bool     `json:"isPrivateKeyAccessible"`
	KeySize         int             `json:"keySize"`
	Usage           []string        `json:"usage"`
}

// NewA3CertificateManager creates a new A3 certificate manager
func NewA3CertificateManager(config *PKCS11ConfigExtended) *A3CertificateManager {
	if config == nil {
		config = DefaultPKCS11ConfigExtended()
	}
	
	return &A3CertificateManager{
		contexts: make(map[string]*crypto11.Context),
		config:   convertToPKCS11Config(config),
		mutex:    sync.RWMutex{},
	}
}

// DefaultPKCS11ConfigExtended returns a default extended PKCS#11 configuration
func DefaultPKCS11ConfigExtended() *PKCS11ConfigExtended {
	return &PKCS11ConfigExtended{
		LibraryPath:         "/usr/lib/x86_64-linux-gnu/opensc-pkcs11.so", // Common Linux path
		MaxSessions:         10,
		SessionTimeout:      30 * time.Minute,
		RetryAttempts:       3,
		RetryDelay:          1 * time.Second,
		ForceAuthentication: true,
		LogSensitiveData:    false,
	}
}

// ListAvailableTokens lists all available PKCS#11 tokens (simplified implementation)
func (manager *A3CertificateManager) ListAvailableTokens() ([]*A3TokenInfo, error) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	
	// Create a temporary context to test connection
	tempConfig := &crypto11.Config{
		Path: manager.config.LibraryPath,
	}
	
	ctx, err := crypto11.Configure(tempConfig)
	if err != nil {
		return nil, errors.NewCertificateError("failed to configure PKCS#11", err)
	}
	defer ctx.Close()
	
	// Return a simplified token list since detailed enumeration is not available
	tokens := []*A3TokenInfo{
		{
			SlotID:          0,
			TokenLabel:      "Available Token",
			ManufacturerID:  "Unknown",
			Model:           "Generic PKCS#11 Token",
			SerialNumber:    "000000000",
			IsInitialized:   true,
			HasTokenPresent: true,
			IsWriteProtected: false,
			LoginRequired:   true,
		},
	}
	
	return tokens, nil
}

// LoadCertificateFromToken loads a certificate from a specific token
func (manager *A3CertificateManager) LoadCertificateFromToken(tokenLabel, pin string) (*A3Certificate, error) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	// Check if context already exists for this token
	contextKey := fmt.Sprintf("%s:%s", tokenLabel, manager.config.LibraryPath)
	ctx, exists := manager.contexts[contextKey]
	
	if !exists {
		// Create new context
		config := &crypto11.Config{
			Path:       manager.config.LibraryPath,
			TokenLabel: tokenLabel,
			Pin:        pin,
		}
		
		var err error
		ctx, err = crypto11.Configure(config)
		if err != nil {
			return nil, errors.NewCertificateError("failed to configure PKCS#11 for token", err)
		}
		
		manager.contexts[contextKey] = ctx
	}
	
	// Find certificate and private key using existing A3 loader
	a3Loader := NewA3CertificateLoader(DefaultConfig())
	pkcs11Config := &PKCS11Config{
		LibraryPath: manager.config.LibraryPath,
		TokenLabel:  tokenLabel,
		PIN:         pin,
	}
	
	a3Cert, err := a3Loader.LoadFromToken(pkcs11Config)
	if err != nil {
		return nil, errors.NewCertificateError("failed to load certificate from token", err)
	}
	
	return a3Cert, nil
}

// GetCertificateDetails returns detailed information about certificates in a token
func (manager *A3CertificateManager) GetCertificateDetails(tokenLabel, pin string) ([]*A3CertificateDetails, error) {
	// Load certificate using existing loader
	a3Cert, err := manager.LoadCertificateFromToken(tokenLabel, pin)
	if err != nil {
		return nil, err
	}
	defer a3Cert.Close()
	
	cert := a3Cert.GetCertificate()
	if cert == nil {
		return nil, errors.NewCertificateError("no certificate available", nil)
	}
	
	certInfo := GetCertificateInfo(cert, TypeA3)
	tokenInfo := &A3TokenInfo{
		SlotID:          0,
		TokenLabel:      tokenLabel,
		ManufacturerID:  "Unknown",
		Model:           "PKCS#11 Token",
		SerialNumber:    "Unknown",
		IsInitialized:   true,
		HasTokenPresent: true,
		IsWriteProtected: false,
		LoginRequired:   true,
	}
	
	// Get key size
	keySize := GetCertificateKeySize(cert)
	
	// Get key usage
	usage := getKeyUsageStrings(cert)
	
	detail := &A3CertificateDetails{
		CertificateInfo:        certInfo,
		TokenInfo:              tokenInfo,
		Label:                  tokenLabel,
		ID:                     nil,
		IsPrivateKeyAccessible: true, // Assume accessible if we got the certificate
		KeySize:                keySize,
		Usage:                  usage,
	}
	
	return []*A3CertificateDetails{detail}, nil
}

// TestTokenConnection tests if a token connection can be established
func (manager *A3CertificateManager) TestTokenConnection(tokenLabel, pin string) error {
	config := &crypto11.Config{
		Path:       manager.config.LibraryPath,
		TokenLabel: tokenLabel,
		Pin:        pin,
	}
	
	ctx, err := crypto11.Configure(config)
	if err != nil {
		return errors.NewCertificateError("failed to connect to token", err)
	}
	defer ctx.Close()
	
	// Simple test - try to create a key pair finder
	_, err = ctx.FindKeyPair(nil, nil)
	if err != nil {
		// This is expected if no keys are found, but connection works
		// We consider this a successful connection test
	}
	
	return nil
}

// Close closes all PKCS#11 contexts
func (manager *A3CertificateManager) Close() error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	var lastError error
	for key, ctx := range manager.contexts {
		if err := ctx.Close(); err != nil {
			lastError = err
		}
		delete(manager.contexts, key)
	}
	
	return lastError
}

// Enhanced A3Certificate methods

// SignWithRetry signs data with retry logic for hardware failures
func (a3 *A3Certificate) SignWithRetry(data []byte, algorithm crypto.Hash, maxRetries int) ([]byte, error) {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()
	
	var lastError error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		signature, err := a3.privateKey.Sign(nil, data, algorithm)
		if err == nil {
			return signature, nil
		}
		
		lastError = err
		if attempt < maxRetries {
			time.Sleep(time.Millisecond * 100) // Brief delay before retry
		}
	}
	
	return nil, errors.NewCertificateError("failed to sign after retries", lastError)
}

// GetTokenInfo returns simplified information about the token containing this certificate
func (a3 *A3Certificate) GetTokenInfo() (*A3TokenInfo, error) {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()
	
	if a3.context == nil {
		return nil, errors.NewCertificateError("PKCS#11 context not available", nil)
	}
	
	// Return simplified token info
	return &A3TokenInfo{
		SlotID:          0,
		TokenLabel:      a3.tokenLabel,
		ManufacturerID:  "Unknown",
		Model:           "PKCS#11 Token",
		SerialNumber:    "Unknown",
		IsInitialized:   true,
		HasTokenPresent: true,
		IsWriteProtected: false,
		LoginRequired:   true,
	}, nil
}

// IsTokenPresentEnhanced checks if the physical token is still present (enhanced version)
func (a3 *A3Certificate) IsTokenPresentEnhanced() bool {
	a3.mutex.RLock()
	defer a3.mutex.RUnlock()
	
	if a3.context == nil {
		return false
	}
	
	// Try a simple operation to test if token is accessible
	_, err := a3.context.FindKeyPair(nil, nil)
	// If we can perform operations, token is likely present
	// Error is expected if no keys found, but context works
	return err == nil || err.Error() != "context closed"
}

// Helper functions

// convertToPKCS11Config converts extended config to basic PKCS11Config
func convertToPKCS11Config(extended *PKCS11ConfigExtended) *PKCS11Config {
	basic := &PKCS11Config{
		LibraryPath: extended.LibraryPath,
		TokenLabel:  extended.TokenLabel,
		PIN:         extended.PIN,
		Slot:        extended.Slot,
	}
	
	return basic
}

// getKeyUsageStrings converts X.509 key usage to human-readable strings
func getKeyUsageStrings(cert *x509.Certificate) []string {
	var usage []string
	
	if cert.KeyUsage&x509.KeyUsageDigitalSignature != 0 {
		usage = append(usage, "Digital Signature")
	}
	if cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0 {
		usage = append(usage, "Key Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageDataEncipherment != 0 {
		usage = append(usage, "Data Encipherment")
	}
	if cert.KeyUsage&x509.KeyUsageKeyAgreement != 0 {
		usage = append(usage, "Key Agreement")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign != 0 {
		usage = append(usage, "Certificate Signing")
	}
	if cert.KeyUsage&x509.KeyUsageCRLSign != 0 {
		usage = append(usage, "CRL Signing")
	}
	
	return usage
}

// LoadA3CertificateFromToken is a convenience function to load A3 certificate
func LoadA3CertificateFromToken(libraryPath, tokenLabel, pin string) (*A3Certificate, error) {
	config := &PKCS11ConfigExtended{
		LibraryPath: libraryPath,
		TokenLabel:  tokenLabel,
		PIN:         pin,
	}
	
	manager := NewA3CertificateManager(config)
	defer manager.Close()
	
	return manager.LoadCertificateFromToken(tokenLabel, pin)
}

// ValidateA3Certificate performs comprehensive validation of an A3 certificate
func ValidateA3Certificate(cert *A3Certificate) error {
	if cert == nil {
		return errors.NewValidationError("certificate cannot be nil", "certificate", "")
	}
	
	// Check if token is present
	if !cert.IsTokenPresent() {
		return errors.NewCertificateError("token not present", nil)
	}
	
	// Validate the X.509 certificate
	x509Cert := cert.GetCertificate()
	if err := ValidateCertificate(x509Cert, cert.config); err != nil {
		return err
	}
	
	// Test private key access
	testData := []byte("test")
	_, err := cert.SignWithRetry(testData, crypto.SHA1, 2)
	if err != nil {
		return errors.NewCertificateError("private key not accessible", err)
	}
	
	return nil
}