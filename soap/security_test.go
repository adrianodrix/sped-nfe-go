package soap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"strings"
	"testing"
	"time"
)

// Helper function to create a test certificate and private key
func createTestCertificate(t *testing.T) (*x509.Certificate, *rsa.PrivateKey) {
	t.Helper()

	// Generate RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:      []string{"BR"},
			Organization: []string{"Test Org"},
			CommonName:   "Test Certificate",
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		IPAddresses: nil,
	}

	// Create certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert, privateKey
}

// Helper function to create test certificate PEM data
func createTestCertificatePEM(t *testing.T) ([]byte, []byte) {
	t.Helper()

	cert, privateKey := createTestCertificate(t)

	// Convert certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	// Convert private key to PEM
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	return certPEM, keyPEM
}

// Helper function to create expired certificate
func createExpiredCertificate(t *testing.T) *x509.Certificate {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Country:      []string{"BR"},
			Organization: []string{"Expired Test Org"},
			CommonName:   "Expired Test Certificate",
		},
		NotBefore: time.Now().Add(-48 * time.Hour), // Started 2 days ago
		NotAfter:  time.Now().Add(-24 * time.Hour), // Expired 1 day ago
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("Failed to create expired certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse expired certificate: %v", err)
	}

	return cert
}

func TestNewWSSecurityManager(t *testing.T) {
	tests := []struct {
		name   string
		config *WSSecurityConfig
		want   *WSSecurityManager
	}{
		{
			name:   "nil config creates default",
			config: nil,
			want: &WSSecurityManager{
				config: &WSSecurityConfig{
					TimestampTTL:  5 * time.Minute,
					IncludeToken:  true,
					SignTimestamp: true,
					SignBody:      false,
				},
			},
		},
		{
			name: "custom config",
			config: &WSSecurityConfig{
				TimestampTTL:  10 * time.Minute,
				IncludeToken:  false,
				SignTimestamp: false,
				SignBody:      true,
			},
			want: &WSSecurityManager{
				config: &WSSecurityConfig{
					TimestampTTL:  10 * time.Minute,
					IncludeToken:  false,
					SignTimestamp: false,
					SignBody:      true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWSSecurityManager(tt.config)
			if got == nil {
				t.Error("NewWSSecurityManager() returned nil")
				return
			}

			if got.config.TimestampTTL != tt.want.config.TimestampTTL {
				t.Errorf("TimestampTTL = %v, want %v", got.config.TimestampTTL, tt.want.config.TimestampTTL)
			}
			if got.config.IncludeToken != tt.want.config.IncludeToken {
				t.Errorf("IncludeToken = %v, want %v", got.config.IncludeToken, tt.want.config.IncludeToken)
			}
			if got.config.SignTimestamp != tt.want.config.SignTimestamp {
				t.Errorf("SignTimestamp = %v, want %v", got.config.SignTimestamp, tt.want.config.SignTimestamp)
			}
			if got.config.SignBody != tt.want.config.SignBody {
				t.Errorf("SignBody = %v, want %v", got.config.SignBody, tt.want.config.SignBody)
			}
		})
	}
}

func TestLoadCertificateFromPEM(t *testing.T) {
	t.Run("valid PEM data", func(t *testing.T) {
		certPEM, keyPEM := createTestCertificatePEM(t)

		cert, key, err := LoadCertificateFromPEM(certPEM, keyPEM)
		if err != nil {
			t.Fatalf("LoadCertificateFromPEM() error = %v", err)
		}

		if cert == nil {
			t.Error("LoadCertificateFromPEM() returned nil certificate")
		}
		if key == nil {
			t.Error("LoadCertificateFromPEM() returned nil private key")
		}

		// Verify the public key matches
		if cert.PublicKey.(*rsa.PublicKey).N.Cmp(key.N) != 0 {
			t.Error("Certificate public key doesn't match private key")
		}
	})

	t.Run("invalid certificate PEM", func(t *testing.T) {
		invalidCert := []byte("invalid cert data")
		_, keyPEM := createTestCertificatePEM(t)

		_, _, err := LoadCertificateFromPEM(invalidCert, keyPEM)
		if err == nil {
			t.Error("LoadCertificateFromPEM() should fail with invalid certificate")
		}
	})

	t.Run("invalid private key PEM", func(t *testing.T) {
		certPEM, _ := createTestCertificatePEM(t)
		invalidKey := []byte("invalid key data")

		_, _, err := LoadCertificateFromPEM(certPEM, invalidKey)
		if err == nil {
			t.Error("LoadCertificateFromPEM() should fail with invalid private key")
		}
	})

	t.Run("PKCS8 private key", func(t *testing.T) {
		cert, privateKey := createTestCertificate(t)

		// Convert certificate to PEM
		certPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})

		// Convert private key to PKCS8 PEM
		privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			t.Fatalf("Failed to marshal PKCS8 private key: %v", err)
		}

		keyPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privateKeyDER,
		})

		_, _, err = LoadCertificateFromPEM(certPEM, keyPEM)
		if err != nil {
			t.Fatalf("LoadCertificateFromPEM() should handle PKCS8 keys: %v", err)
		}
	})
}

func TestCreateSecurityHeader(t *testing.T) {
	cert, privateKey := createTestCertificate(t)

	t.Run("with certificate and timestamp", func(t *testing.T) {
		config := &WSSecurityConfig{
			Certificate:   cert,
			PrivateKey:    privateKey,
			TimestampTTL:  5 * time.Minute,
			IncludeToken:  true,
			SignTimestamp: true,
		}

		wsm := NewWSSecurityManager(config)
		header, err := wsm.CreateSecurityHeader("test-timestamp-id")

		if err != nil {
			t.Fatalf("CreateSecurityHeader() error = %v", err)
		}

		if header.Timestamp == nil {
			t.Error("CreateSecurityHeader() should include timestamp")
		}

		if header.BinarySecurityToken == nil {
			t.Error("CreateSecurityHeader() should include binary security token")
		}

		if header.Signature == nil {
			t.Error("CreateSecurityHeader() should include signature when SignTimestamp is true")
		}

		if header.XmlnsWsse == "" {
			t.Error("CreateSecurityHeader() should set WSSE namespace")
		}

		if header.XmlnsWsu == "" {
			t.Error("CreateSecurityHeader() should set WSU namespace")
		}
	})

	t.Run("without certificate", func(t *testing.T) {
		config := &WSSecurityConfig{
			TimestampTTL: 5 * time.Minute,
		}

		wsm := NewWSSecurityManager(config)
		header, err := wsm.CreateSecurityHeader("test-timestamp-id")

		if err != nil {
			t.Fatalf("CreateSecurityHeader() error = %v", err)
		}

		if header.Timestamp == nil {
			t.Error("CreateSecurityHeader() should include timestamp")
		}

		if header.BinarySecurityToken != nil {
			t.Error("CreateSecurityHeader() should not include binary security token without certificate")
		}

		if header.Signature != nil {
			t.Error("CreateSecurityHeader() should not include signature without certificate")
		}
	})

	t.Run("without timestamp", func(t *testing.T) {
		config := &WSSecurityConfig{
			Certificate:  cert,
			TimestampTTL: 0, // No timestamp
			IncludeToken: true,
		}

		wsm := NewWSSecurityManager(config)
		header, err := wsm.CreateSecurityHeader("test-timestamp-id")

		if err != nil {
			t.Fatalf("CreateSecurityHeader() error = %v", err)
		}

		if header.Timestamp != nil {
			t.Error("CreateSecurityHeader() should not include timestamp when TTL is 0")
		}

		if header.BinarySecurityToken == nil {
			t.Error("CreateSecurityHeader() should include binary security token")
		}
	})
}

func TestValidateCertificate(t *testing.T) {
	t.Run("valid certificate", func(t *testing.T) {
		cert, _ := createTestCertificate(t)

		err := ValidateCertificate(cert, nil)
		if err != nil {
			t.Errorf("ValidateCertificate() error = %v, want nil", err)
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		err := ValidateCertificate(nil, nil)
		if err == nil {
			t.Error("ValidateCertificate() should fail with nil certificate")
		}
	})

	t.Run("expired certificate", func(t *testing.T) {
		expiredCert := createExpiredCertificate(t)

		err := ValidateCertificate(expiredCert, nil)
		if err == nil {
			t.Error("ValidateCertificate() should fail with expired certificate")
		}
	})

	t.Run("certificate without digital signature usage", func(t *testing.T) {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("Failed to generate private key: %v", err)
		}

		template := x509.Certificate{
			SerialNumber: big.NewInt(3),
			Subject: pkix.Name{
				CommonName: "No Digital Signature",
			},
			NotBefore: time.Now().Add(-1 * time.Hour),
			NotAfter:  time.Now().Add(24 * time.Hour),
			KeyUsage:  x509.KeyUsageKeyEncipherment, // No digital signature
		}

		certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			t.Fatalf("Failed to create certificate: %v", err)
		}

		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			t.Fatalf("Failed to parse certificate: %v", err)
		}

		err = ValidateCertificate(cert, nil)
		if err == nil {
			t.Error("ValidateCertificate() should fail with certificate without digital signature usage")
		}
	})
}

func TestCreateTimestamp(t *testing.T) {
	id := "test-timestamp"
	validMinutes := 5

	timestamp := CreateTimestamp(id, validMinutes)

	if timestamp == nil {
		t.Fatal("CreateTimestamp() returned nil")
	}

	if timestamp.ID != id {
		t.Errorf("CreateTimestamp() ID = %v, want %v", timestamp.ID, id)
	}

	if timestamp.Created == "" {
		t.Error("CreateTimestamp() Created should not be empty")
	}

	if timestamp.Expires == "" {
		t.Error("CreateTimestamp() Expires should not be empty")
	}

	// Parse and validate times
	created, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp.Created)
	if err != nil {
		t.Errorf("CreateTimestamp() Created format invalid: %v", err)
	}

	expires, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp.Expires)
	if err != nil {
		t.Errorf("CreateTimestamp() Expires format invalid: %v", err)
	}

	expectedDuration := time.Duration(validMinutes) * time.Minute
	actualDuration := expires.Sub(created)

	// Allow some tolerance for execution time
	if actualDuration < expectedDuration-time.Second || actualDuration > expectedDuration+time.Second {
		t.Errorf("CreateTimestamp() duration = %v, want ~%v", actualDuration, expectedDuration)
	}
}

func TestValidateTimestampElement(t *testing.T) {
	t.Run("nil timestamp", func(t *testing.T) {
		err := ValidateTimestampElement(nil)
		if err != nil {
			t.Errorf("ValidateTimestampElement() should accept nil timestamp: %v", err)
		}
	})

	t.Run("valid timestamp", func(t *testing.T) {
		timestamp := CreateTimestamp("test-id", 5)
		err := ValidateTimestampElement(timestamp)
		if err != nil {
			t.Errorf("ValidateTimestampElement() error = %v", err)
		}
	})

	t.Run("expired timestamp", func(t *testing.T) {
		now := time.Now().UTC()
		created := now.Add(-10 * time.Minute).Format("2006-01-02T15:04:05.000Z")
		expires := now.Add(-5 * time.Minute).Format("2006-01-02T15:04:05.000Z")

		timestamp := &Timestamp{
			ID:      "expired",
			Created: created,
			Expires: expires,
		}

		err := ValidateTimestampElement(timestamp)
		if err == nil {
			t.Error("ValidateTimestampElement() should fail with expired timestamp")
		}
	})

	t.Run("future timestamp", func(t *testing.T) {
		now := time.Now().UTC()
		created := now.Add(10 * time.Minute).Format("2006-01-02T15:04:05.000Z")
		expires := now.Add(15 * time.Minute).Format("2006-01-02T15:04:05.000Z")

		timestamp := &Timestamp{
			ID:      "future",
			Created: created,
			Expires: expires,
		}

		err := ValidateTimestampElement(timestamp)
		if err == nil {
			t.Error("ValidateTimestampElement() should fail with future timestamp")
		}
	})

	t.Run("invalid timestamp format", func(t *testing.T) {
		timestamp := &Timestamp{
			ID:      "invalid",
			Created: "invalid-format",
			Expires: "2024-01-01T12:00:00.000Z",
		}

		err := ValidateTimestampElement(timestamp)
		if err == nil {
			t.Error("ValidateTimestampElement() should fail with invalid Created format")
		}
	})

	t.Run("expires before created", func(t *testing.T) {
		now := time.Now().UTC()
		created := now.Format("2006-01-02T15:04:05.000Z")
		expires := now.Add(-1 * time.Minute).Format("2006-01-02T15:04:05.000Z")

		timestamp := &Timestamp{
			ID:      "backwards",
			Created: created,
			Expires: expires,
		}

		err := ValidateTimestampElement(timestamp)
		if err == nil {
			t.Error("ValidateTimestampElement() should fail when expires is before created")
		}
	})
}

func TestExtractCertificateFromToken(t *testing.T) {
	cert, _ := createTestCertificate(t)

	t.Run("valid token", func(t *testing.T) {
		token := &BinarySecurityToken{
			ValueType:    X509TokenValueType,
			EncodingType: Base64EncodingType,
			ID:           "test-token",
			Content:      base64.StdEncoding.EncodeToString(cert.Raw),
		}

		extractedCert, err := ExtractCertificateFromToken(token)
		if err != nil {
			t.Fatalf("ExtractCertificateFromToken() error = %v", err)
		}

		if extractedCert == nil {
			t.Fatal("ExtractCertificateFromToken() returned nil certificate")
		}

		if !extractedCert.Equal(cert) {
			t.Error("ExtractCertificateFromToken() returned different certificate")
		}
	})

	t.Run("nil token", func(t *testing.T) {
		_, err := ExtractCertificateFromToken(nil)
		if err == nil {
			t.Error("ExtractCertificateFromToken() should fail with nil token")
		}
	})

	t.Run("unsupported value type", func(t *testing.T) {
		token := &BinarySecurityToken{
			ValueType:    "unsupported",
			EncodingType: Base64EncodingType,
			ID:           "test-token",
			Content:      base64.StdEncoding.EncodeToString(cert.Raw),
		}

		_, err := ExtractCertificateFromToken(token)
		if err == nil {
			t.Error("ExtractCertificateFromToken() should fail with unsupported value type")
		}
	})

	t.Run("unsupported encoding type", func(t *testing.T) {
		token := &BinarySecurityToken{
			ValueType:    X509TokenValueType,
			EncodingType: "unsupported",
			ID:           "test-token",
			Content:      base64.StdEncoding.EncodeToString(cert.Raw),
		}

		_, err := ExtractCertificateFromToken(token)
		if err == nil {
			t.Error("ExtractCertificateFromToken() should fail with unsupported encoding type")
		}
	})

	t.Run("invalid base64 content", func(t *testing.T) {
		token := &BinarySecurityToken{
			ValueType:    X509TokenValueType,
			EncodingType: Base64EncodingType,
			ID:           "test-token",
			Content:      "invalid-base64-content!@#$",
		}

		_, err := ExtractCertificateFromToken(token)
		if err == nil {
			t.Error("ExtractCertificateFromToken() should fail with invalid base64 content")
		}
	})

	t.Run("invalid certificate data", func(t *testing.T) {
		token := &BinarySecurityToken{
			ValueType:    X509TokenValueType,
			EncodingType: Base64EncodingType,
			ID:           "test-token",
			Content:      base64.StdEncoding.EncodeToString([]byte("invalid cert data")),
		}

		_, err := ExtractCertificateFromToken(token)
		if err == nil {
			t.Error("ExtractCertificateFromToken() should fail with invalid certificate data")
		}
	})
}

func TestCreateWSSecurityConfig(t *testing.T) {
	t.Run("valid PEM data", func(t *testing.T) {
		certPEM, keyPEM := createTestCertificatePEM(t)

		config, err := CreateWSSecurityConfig(certPEM, keyPEM)
		if err != nil {
			t.Fatalf("CreateWSSecurityConfig() error = %v", err)
		}

		if config == nil {
			t.Fatal("CreateWSSecurityConfig() returned nil config")
		}

		if config.Certificate == nil {
			t.Error("CreateWSSecurityConfig() should set certificate")
		}

		if config.PrivateKey == nil {
			t.Error("CreateWSSecurityConfig() should set private key")
		}

		if config.TimestampTTL != 5*time.Minute {
			t.Errorf("CreateWSSecurityConfig() TimestampTTL = %v, want %v", config.TimestampTTL, 5*time.Minute)
		}

		if !config.IncludeToken {
			t.Error("CreateWSSecurityConfig() should set IncludeToken to true")
		}

		if !config.SignTimestamp {
			t.Error("CreateWSSecurityConfig() should set SignTimestamp to true")
		}

		if config.SignBody {
			t.Error("CreateWSSecurityConfig() should set SignBody to false")
		}
	})

	t.Run("invalid PEM data", func(t *testing.T) {
		_, err := CreateWSSecurityConfig([]byte("invalid"), []byte("invalid"))
		if err == nil {
			t.Error("CreateWSSecurityConfig() should fail with invalid PEM data")
		}
	})
}

func TestVerifySignature(t *testing.T) {
	cert, _ := createTestCertificate(t)

	t.Run("valid signature structure", func(t *testing.T) {
		signature := &Signature{
			SignatureValue: SignatureValue{
				Value: "test-signature-value",
			},
			SignedInfo: SignedInfo{
				Reference: []SignatureReference{
					{
						URI:         "#test-uri",
						DigestValue: "test-digest-value",
					},
				},
			},
		}

		err := VerifySignature(signature, cert)
		if err != nil {
			t.Errorf("VerifySignature() error = %v", err)
		}
	})

	t.Run("nil signature", func(t *testing.T) {
		err := VerifySignature(nil, cert)
		if err == nil {
			t.Error("VerifySignature() should fail with nil signature")
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		signature := &Signature{
			SignatureValue: SignatureValue{Value: "test"},
			SignedInfo:     SignedInfo{Reference: []SignatureReference{{}}},
		}

		err := VerifySignature(signature, nil)
		if err == nil {
			t.Error("VerifySignature() should fail with nil certificate")
		}
	})

	t.Run("empty signature value", func(t *testing.T) {
		signature := &Signature{
			SignatureValue: SignatureValue{Value: ""},
			SignedInfo:     SignedInfo{Reference: []SignatureReference{{}}},
		}

		err := VerifySignature(signature, cert)
		if err == nil {
			t.Error("VerifySignature() should fail with empty signature value")
		}
	})

	t.Run("no references", func(t *testing.T) {
		signature := &Signature{
			SignatureValue: SignatureValue{Value: "test"},
			SignedInfo:     SignedInfo{Reference: []SignatureReference{}},
		}

		err := VerifySignature(signature, cert)
		if err == nil {
			t.Error("VerifySignature() should fail with no references")
		}
	})
}

func TestGetCertificateFingerprint(t *testing.T) {
	t.Run("valid certificate", func(t *testing.T) {
		cert, _ := createTestCertificate(t)

		fingerprint := GetCertificateFingerprint(cert)
		if fingerprint == "" {
			t.Error("GetCertificateFingerprint() should not return empty string")
		}

		if len(fingerprint) != 64 { // SHA-256 in hex is 64 characters
			t.Errorf("GetCertificateFingerprint() length = %v, want 64", len(fingerprint))
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		fingerprint := GetCertificateFingerprint(nil)
		if fingerprint != "" {
			t.Errorf("GetCertificateFingerprint() = %v, want empty string", fingerprint)
		}
	})
}

func TestGetCertificateSubject(t *testing.T) {
	t.Run("valid certificate", func(t *testing.T) {
		cert, _ := createTestCertificate(t)

		subject := GetCertificateSubject(cert)
		if subject == "" {
			t.Error("GetCertificateSubject() should not return empty string")
		}

		if !strings.Contains(subject, "Test Certificate") {
			t.Errorf("GetCertificateSubject() = %v, should contain 'Test Certificate'", subject)
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		subject := GetCertificateSubject(nil)
		if subject != "" {
			t.Errorf("GetCertificateSubject() = %v, want empty string", subject)
		}
	})
}

func TestGetCertificateIssuer(t *testing.T) {
	t.Run("valid certificate", func(t *testing.T) {
		cert, _ := createTestCertificate(t)

		issuer := GetCertificateIssuer(cert)
		if issuer == "" {
			t.Error("GetCertificateIssuer() should not return empty string")
		}

		if !strings.Contains(issuer, "Test Certificate") {
			t.Errorf("GetCertificateIssuer() = %v, should contain 'Test Certificate'", issuer)
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		issuer := GetCertificateIssuer(nil)
		if issuer != "" {
			t.Errorf("GetCertificateIssuer() = %v, want empty string", issuer)
		}
	})
}

func TestIsCertificateValid(t *testing.T) {
	t.Run("valid certificate", func(t *testing.T) {
		cert, _ := createTestCertificate(t)

		if !IsCertificateValid(cert) {
			t.Error("IsCertificateValid() should return true for valid certificate")
		}
	})

	t.Run("expired certificate", func(t *testing.T) {
		expiredCert := createExpiredCertificate(t)

		if IsCertificateValid(expiredCert) {
			t.Error("IsCertificateValid() should return false for expired certificate")
		}
	})

	t.Run("nil certificate", func(t *testing.T) {
		if IsCertificateValid(nil) {
			t.Error("IsCertificateValid() should return false for nil certificate")
		}
	})
}

func TestGenerateID(t *testing.T) {
	// Test that generateID returns unique IDs
	id1 := generateID()
	id2 := generateID()

	if id1 == id2 {
		t.Error("generateID() should return unique IDs")
	}

	if len(id1) != 32 { // 16 bytes in hex = 32 characters
		t.Errorf("generateID() length = %v, want 32", len(id1))
	}

	if len(id2) != 32 {
		t.Errorf("generateID() length = %v, want 32", len(id2))
	}
}
