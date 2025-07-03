package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

// createTestCertificate creates a test certificate for testing purposes
func createTestCertificate() (*x509.Certificate, *rsa.PrivateKey, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Organization"},
			Country:       []string{"BR"},
			Province:      []string{"SP"},
			Locality:      []string{"São Paulo"},
			StreetAddress: []string{"Test Address"},
			PostalCode:    []string{"01000-000"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

func TestValidateCertificate(t *testing.T) {
	cert, _, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	config := DefaultConfig()
	err = ValidateCertificate(cert, config)
	if err != nil {
		t.Errorf("Valid certificate should pass validation: %v", err)
	}
}

func TestValidateCertificateNil(t *testing.T) {
	config := DefaultConfig()
	err := ValidateCertificate(nil, config)
	if err == nil {
		t.Error("Nil certificate should fail validation")
	}
}

func TestGetCertificateInfo(t *testing.T) {
	cert, _, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	info := GetCertificateInfo(cert, TypeA1)
	if info == nil {
		t.Error("GetCertificateInfo should not return nil")
	}

	if info.Type != TypeA1 {
		t.Errorf("Expected type A1, got %v", info.Type)
	}

	if info.Subject == "" {
		t.Error("Subject should not be empty")
	}

	if info.SerialNumber == "" {
		t.Error("Serial number should not be empty")
	}
}

func TestIsCertificateValidForSigning(t *testing.T) {
	cert, _, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	if !IsCertificateValidForSigning(cert) {
		t.Error("Valid certificate should be valid for signing")
	}

	if IsCertificateValidForSigning(nil) {
		t.Error("Nil certificate should not be valid for signing")
	}
}

func TestGetCertificateKeySize(t *testing.T) {
	cert, _, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	keySize := GetCertificateKeySize(cert)
	if keySize != 2048 {
		t.Errorf("Expected key size 2048, got %d", keySize)
	}

	if GetCertificateKeySize(nil) != 0 {
		t.Error("Nil certificate should return 0 key size")
	}
}

func TestCertificateTypeString(t *testing.T) {
	if TypeA1.String() != "A1" {
		t.Errorf("Expected A1, got %s", TypeA1.String())
	}

	if TypeA3.String() != "A3" {
		t.Errorf("Expected A3, got %s", TypeA3.String())
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Error("DefaultConfig should not return nil")
	}

	if !config.ValidateChain {
		t.Error("Default config should validate chain")
	}

	if config.AllowExpired {
		t.Error("Default config should not allow expired certificates")
	}

	if config.CacheTimeout != 5*time.Minute {
		t.Errorf("Expected cache timeout 5 minutes, got %v", config.CacheTimeout)
	}
}
