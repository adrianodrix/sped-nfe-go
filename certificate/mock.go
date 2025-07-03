// Package certificate provides mock certificate for testing
package certificate

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"time"
)

// MockCertificate implements the Certificate interface for testing
type MockCertificate struct {
	privateKey *rsa.PrivateKey
	cert       *x509.Certificate
}

// NewMockCertificate creates a new mock certificate for testing
func NewMockCertificate() *MockCertificate {
	// Generate a real RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("Failed to generate mock private key: " + err.Error())
	}

	// Create a basic X509 certificate for testing
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(123456789),
		Subject: pkix.Name{
			CommonName: "Mock ICP-Brasil Certificate:12345678000195",
		},
		Issuer: pkix.Name{
			CommonName: "AC Mock Issuer",
		},
		NotBefore:   time.Now().Add(-24 * time.Hour),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		PublicKey:   &privateKey.PublicKey,
	}

	return &MockCertificate{
		privateKey: privateKey,
		cert:       cert,
	}
}

// Sign signs the given data using the certificate's private key
func (m *MockCertificate) Sign(data []byte, algorithm crypto.Hash) ([]byte, error) {
	// Perform real RSA signature for realistic behavior
	hash := algorithm.New()
	hash.Write(data)
	hashed := hash.Sum(nil)
	
	signature, err := rsa.SignPKCS1v15(rand.Reader, m.privateKey, algorithm, hashed)
	if err != nil {
		return nil, err
	}
	
	return signature, nil
}

// GetPublicKey returns the certificate's public key
func (m *MockCertificate) GetPublicKey() crypto.PublicKey {
	if m.cert != nil {
		return m.cert.PublicKey
	}
	return nil
}

// GetPrivateKey returns the certificate's private key
func (m *MockCertificate) GetPrivateKey() crypto.PrivateKey {
	return m.privateKey
}

// GetCertificate returns the certificate
func (m *MockCertificate) GetCertificate() *x509.Certificate {
	return m.cert
}

// IsValid returns true if the certificate is valid
func (m *MockCertificate) IsValid() bool {
	return true
}

// GetSubject returns the certificate subject as a formatted string
func (m *MockCertificate) GetSubject() string {
	return "Mock Certificate Subject"
}

// GetIssuer returns the certificate issuer as a formatted string
func (m *MockCertificate) GetIssuer() string {
	return "Mock Certificate Issuer"
}

// GetSerialNumber returns the certificate serial number as a string
func (m *MockCertificate) GetSerialNumber() string {
	return "123456789"
}

// GetFingerprint returns the SHA-256 fingerprint of the certificate
func (m *MockCertificate) GetFingerprint() string {
	return "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99"
}

// GetCertificateData returns the raw certificate data for inclusion in KeyInfo
func (m *MockCertificate) GetCertificateData() string {
	if m.cert == nil {
		return ""
	}
	// Return a mock certificate data (base64 encoded)
	return base64.StdEncoding.EncodeToString([]byte("mock-certificate-data-for-testing"))
}

// GetValidityPeriod returns the certificate's not before and not after dates
func (m *MockCertificate) GetValidityPeriod() (notBefore, notAfter time.Time) {
	now := time.Now()
	return now.AddDate(0, -1, 0), now.AddDate(1, 0, 0) // Valid from 1 month ago to 1 year from now
}

// Close releases any resources associated with the certificate
func (m *MockCertificate) Close() error {
	return nil
}
