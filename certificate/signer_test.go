package certificate

import (
	"crypto"
	"crypto/x509"
	"testing"
	"time"
)

// mockCertificate implements the Certificate interface for testing
type mockCertificate struct {
	cert      *x509.Certificate
	key       crypto.PrivateKey
	valid     bool
	signError error
}

func (m *mockCertificate) Sign(data []byte, algorithm crypto.Hash) ([]byte, error) {
	if m.signError != nil {
		return nil, m.signError
	}
	return []byte("mock signature"), nil
}

func (m *mockCertificate) GetPublicKey() crypto.PublicKey {
	if m.cert != nil {
		return m.cert.PublicKey
	}
	return nil
}

func (m *mockCertificate) GetPrivateKey() crypto.PrivateKey {
	return m.key
}

func (m *mockCertificate) GetCertificate() *x509.Certificate {
	return m.cert
}

func (m *mockCertificate) IsValid() bool {
	return m.valid
}

func (m *mockCertificate) GetSubject() string {
	if m.cert != nil {
		return m.cert.Subject.String()
	}
	return ""
}

func (m *mockCertificate) GetIssuer() string {
	if m.cert != nil {
		return m.cert.Issuer.String()
	}
	return ""
}

func (m *mockCertificate) GetSerialNumber() string {
	if m.cert != nil {
		return m.cert.SerialNumber.String()
	}
	return ""
}

func (m *mockCertificate) GetFingerprint() string {
	return "mock:fingerprint"
}

func (m *mockCertificate) GetValidityPeriod() (notBefore, notAfter time.Time) {
	if m.cert != nil {
		return m.cert.NotBefore, m.cert.NotAfter
	}
	return time.Time{}, time.Time{}
}

func (m *mockCertificate) Close() error {
	return nil
}

func TestNewXMLSigner(t *testing.T) {
	cert, _, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	mockCert := &mockCertificate{cert: cert, valid: true}
	signer := NewXMLSigner(mockCert, nil)

	if signer == nil {
		t.Error("NewXMLSigner should not return nil")
	}

	if signer.certificate == nil {
		t.Error("Signer should use provided certificate")
	}

	if signer.config == nil {
		t.Error("Signer should have default config when nil is provided")
	}
}

func TestDefaultSignerConfig(t *testing.T) {
	config := DefaultSignerConfig()
	if config == nil {
		t.Error("DefaultSignerConfig should not return nil")
	}

	if config.DigestAlgorithm != crypto.SHA1 {
		t.Errorf("Expected SHA1, got %v", config.DigestAlgorithm)
	}

	if config.SignatureAlgorithm != "http://www.w3.org/2000/09/xmldsig#rsa-sha1" {
		t.Errorf("Unexpected signature algorithm: %s", config.SignatureAlgorithm)
	}

	if !config.IncludeCertificate {
		t.Error("Default config should include certificate")
	}
}

func TestSHA256SignerConfig(t *testing.T) {
	config := SHA256SignerConfig()
	if config == nil {
		t.Error("SHA256SignerConfig should not return nil")
	}

	if config.DigestAlgorithm != crypto.SHA256 {
		t.Errorf("Expected SHA256, got %v", config.DigestAlgorithm)
	}

	if config.SignatureAlgorithm != "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256" {
		t.Errorf("Unexpected signature algorithm: %s", config.SignatureAlgorithm)
	}
}

func TestXMLSignerGetCertificate(t *testing.T) {
	cert, _, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	mockCert := &mockCertificate{cert: cert, valid: true}
	signer := NewXMLSigner(mockCert, nil)

	retrievedCert := signer.GetCertificate()
	if retrievedCert != cert {
		t.Error("GetCertificate should return the wrapped certificate")
	}
}

func TestXMLSignerSign(t *testing.T) {
	cert, _, err := createTestCertificate()
	if err != nil {
		t.Fatalf("Failed to create test certificate: %v", err)
	}

	mockCert := &mockCertificate{cert: cert, valid: true}
	signer := NewXMLSigner(mockCert, nil)

	signature, err := signer.Sign([]byte("test data"))
	if err != nil {
		t.Errorf("Sign should not return error: %v", err)
	}

	if string(signature) != "mock signature" {
		t.Errorf("Expected 'mock signature', got '%s'", string(signature))
	}
}

func TestXMLSignerSignatureAlgorithm(t *testing.T) {
	mockCert := &mockCertificate{valid: true}
	signer := NewXMLSigner(mockCert, nil)

	algorithm := signer.SignatureAlgorithm()
	expected := "http://www.w3.org/2000/09/xmldsig#rsa-sha1"
	if algorithm != expected {
		t.Errorf("Expected %s, got %s", expected, algorithm)
	}
}

func TestSignXMLValidation(t *testing.T) {
	mockCert := &mockCertificate{valid: false}
	signer := NewXMLSigner(mockCert, nil)

	// Test empty XML
	_, _, err := signer.SignXML("")
	if err == nil {
		t.Error("SignXML should fail with empty XML content")
	}

	// Test nil certificate
	signer.certificate = nil
	_, _, err = signer.SignXML("<test/>")
	if err == nil {
		t.Error("SignXML should fail with nil certificate")
	}
}

func TestSignXMLWithInvalidCertificate(t *testing.T) {
	mockCert := &mockCertificate{valid: false}
	signer := NewXMLSigner(mockCert, nil)

	_, _, err := signer.SignXML("<test/>")
	if err == nil {
		t.Error("SignXML should fail with invalid certificate")
	}
}

func TestCreateXMLSigner(t *testing.T) {
	mockCert := &mockCertificate{valid: true}
	signer := CreateXMLSigner(mockCert)

	if signer == nil {
		t.Error("CreateXMLSigner should not return nil")
	}

	if signer.config.DigestAlgorithm != crypto.SHA1 {
		t.Error("CreateXMLSigner should use default config")
	}
}

func TestCreateSHA256XMLSigner(t *testing.T) {
	mockCert := &mockCertificate{valid: true}
	signer := CreateSHA256XMLSigner(mockCert)

	if signer == nil {
		t.Error("CreateSHA256XMLSigner should not return nil")
	}

	if signer.config.DigestAlgorithm != crypto.SHA256 {
		t.Error("CreateSHA256XMLSigner should use SHA256 config")
	}
}

func TestCalculateDigest(t *testing.T) {
	mockCert := &mockCertificate{valid: true}
	signer := NewXMLSigner(mockCert, nil)

	testData := []byte("test data")

	// Test SHA1
	digest := signer.calculateDigest(testData)
	if len(digest) != 20 { // SHA1 produces 20 bytes
		t.Errorf("Expected SHA1 digest length 20, got %d", len(digest))
	}

	// Test SHA256
	signer.config.DigestAlgorithm = crypto.SHA256
	digest = signer.calculateDigest(testData)
	if len(digest) != 32 { // SHA256 produces 32 bytes
		t.Errorf("Expected SHA256 digest length 32, got %d", len(digest))
	}
}

func TestGetDigestMethodURI(t *testing.T) {
	mockCert := &mockCertificate{valid: true}
	signer := NewXMLSigner(mockCert, nil)

	// Test SHA1
	uri := signer.getDigestMethodURI()
	expected := "http://www.w3.org/2000/09/xmldsig#sha1"
	if uri != expected {
		t.Errorf("Expected %s, got %s", expected, uri)
	}

	// Test SHA256
	signer.config.DigestAlgorithm = crypto.SHA256
	uri = signer.getDigestMethodURI()
	expected = "http://www.w3.org/2001/04/xmlenc#sha256"
	if uri != expected {
		t.Errorf("Expected %s, got %s", expected, uri)
	}
}