// Package certificate_test tests certificate functionality with real certificates.
package certificate

import (
	"crypto"
	"os"
	"testing"
)

func TestA1CertificateLoadingIntegration(t *testing.T) {
	// Test with our real certificate
	certPath := "../refs/certificates/cert-valido-jan-2026.pfx"
	password := "kzm7rwu!ewv1ymw3YTM@"

	// Check if certificate file exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Skip("Certificate file not found, skipping test")
		return
	}

	// Load certificate
	cert, err := LoadA1FromFile(certPath, password)
	if err != nil {
		t.Fatalf("Failed to load certificate: %v", err)
	}
	defer cert.Close()

	// Test basic certificate methods
	if !cert.IsValid() {
		t.Error("Certificate should be valid")
	}

	subject := cert.GetSubject()
	if subject == "" {
		t.Error("Certificate subject should not be empty")
	}

	issuer := cert.GetIssuer()
	if issuer == "" {
		t.Error("Certificate issuer should not be empty")
	}

	fingerprint := cert.GetFingerprint()
	if fingerprint == "" {
		t.Error("Certificate fingerprint should not be empty")
	}

	notBefore, notAfter := cert.GetValidityPeriod()
	if notBefore.IsZero() || notAfter.IsZero() {
		t.Error("Certificate validity period should not be zero")
	}

	// Test signing capability
	testData := []byte("test data for signing")
	signature, err := cert.Sign(testData, crypto.SHA256) // Use SHA-256 algorithm
	if err != nil {
		t.Errorf("Failed to sign data: %v", err)
	}

	if len(signature) == 0 {
		t.Error("Signature should not be empty")
	}

	// Test PKCS#12 export
	exportedP12, err := cert.ExportToPKCS12("testpassword")
	if err != nil {
		t.Errorf("Failed to export to PKCS#12: %v", err)
	}
	if len(exportedP12) == 0 {
		t.Error("Exported PKCS#12 should not be empty")
	}

	t.Logf("Certificate loaded successfully:")
	t.Logf("  Subject: %s", subject)
	t.Logf("  Issuer: %s", issuer)
	t.Logf("  Valid from %s to %s", notBefore.Format("2006-01-02"), notAfter.Format("2006-01-02"))
	t.Logf("  Fingerprint: %s", fingerprint)
}

func TestCertificateValidationIntegration(t *testing.T) {
	certPath := "../refs/certificates/cert-valido-jan-2026.pfx"
	password := "kzm7rwu!ewv1ymw3YTM@"

	// Check if certificate file exists
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		t.Skip("Certificate file not found, skipping test")
		return
	}

	// Load certificate
	cert, err := LoadA1FromFile(certPath, password)
	if err != nil {
		t.Fatalf("Failed to load certificate: %v", err)
	}
	defer cert.Close()

	// Test ICP-Brasil validation
	if !cert.IsICPBrasilCertificate() {
		t.Error("Certificate should be recognized as ICP-Brasil")
	}

	// Test NFe validation - note: since LoadA1FromFile returns *A1Certificate, 
	// we need to cast to interface to use ValidateForNFe
	var certInterface Certificate = cert
	err = ValidateForNFe(certInterface)
	if err != nil {
		t.Errorf("Certificate should be valid for NFe use: %v", err)
	}
}