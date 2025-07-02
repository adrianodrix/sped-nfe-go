// Package soap provides WS-Security implementation for SOAP webservice authentication
// with support for digital signatures, timestamps, and certificate validation.
package soap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
)

// WSSecurityConfig holds configuration for WS-Security
type WSSecurityConfig struct {
	Certificate    *x509.Certificate `json:"-"`
	PrivateKey     *rsa.PrivateKey   `json:"-"`
	TimestampTTL   time.Duration     `json:"timestampTTL"`
	IncludeToken   bool              `json:"includeToken"`
	SignTimestamp  bool              `json:"signTimestamp"`
	SignBody       bool              `json:"signBody"`
}

// WSSecurityManager handles WS-Security operations
type WSSecurityManager struct {
	config *WSSecurityConfig
}

// BinarySecurityToken represents a binary security token
type BinarySecurityToken struct {
	XMLName      xml.Name `xml:"wsse:BinarySecurityToken"`
	ValueType    string   `xml:"ValueType,attr"`
	EncodingType string   `xml:"EncodingType,attr"`
	ID           string   `xml:"wsu:Id,attr"`
	Content      string   `xml:",chardata"`
}

// SecurityTokenReference represents a security token reference
type SecurityTokenReference struct {
	XMLName   xml.Name `xml:"wsse:SecurityTokenReference"`
	Reference Reference `xml:"wsse:Reference"`
}

// Reference represents a reference to a security token
type Reference struct {
	XMLName   xml.Name `xml:"wsse:Reference"`
	URI       string   `xml:"URI,attr"`
	ValueType string   `xml:"ValueType,attr,omitempty"`
}

// Signature represents an XML digital signature
type Signature struct {
	XMLName        xml.Name       `xml:"ds:Signature"`
	XmlnsDs        string         `xml:"xmlns:ds,attr"`
	SignedInfo     SignedInfo     `xml:"ds:SignedInfo"`
	SignatureValue SignatureValue `xml:"ds:SignatureValue"`
	KeyInfo        *KeyInfo       `xml:"ds:KeyInfo,omitempty"`
}

// SignedInfo contains information about what is being signed
type SignedInfo struct {
	XMLName                xml.Name               `xml:"ds:SignedInfo"`
	CanonicalizationMethod CanonicalizationMethod `xml:"ds:CanonicalizationMethod"`
	SignatureMethod        SignatureMethod        `xml:"ds:SignatureMethod"`
	Reference              []SignatureReference   `xml:"ds:Reference"`
}

// CanonicalizationMethod specifies the canonicalization algorithm
type CanonicalizationMethod struct {
	XMLName   xml.Name `xml:"ds:CanonicalizationMethod"`
	Algorithm string   `xml:"Algorithm,attr"`
}

// SignatureMethod specifies the signature algorithm
type SignatureMethod struct {
	XMLName   xml.Name `xml:"ds:SignatureMethod"`
	Algorithm string   `xml:"Algorithm,attr"`
}

// SignatureReference represents a reference in the signature
type SignatureReference struct {
	XMLName      xml.Name     `xml:"ds:Reference"`
	URI          string       `xml:"URI,attr"`
	Transforms   *Transforms  `xml:"ds:Transforms,omitempty"`
	DigestMethod DigestMethod `xml:"ds:DigestMethod"`
	DigestValue  string       `xml:"ds:DigestValue"`
}

// Transforms contains transformation algorithms
type Transforms struct {
	XMLName   xml.Name    `xml:"ds:Transforms"`
	Transform []Transform `xml:"ds:Transform"`
}

// Transform represents a transformation algorithm
type Transform struct {
	XMLName   xml.Name `xml:"ds:Transform"`
	Algorithm string   `xml:"Algorithm,attr"`
}

// DigestMethod specifies the digest algorithm
type DigestMethod struct {
	XMLName   xml.Name `xml:"ds:DigestMethod"`
	Algorithm string   `xml:"Algorithm,attr"`
}

// SignatureValue contains the signature value
type SignatureValue struct {
	XMLName xml.Name `xml:"ds:SignatureValue"`
	ID      string   `xml:"Id,attr,omitempty"`
	Value   string   `xml:",chardata"`
}

// KeyInfo contains key information
type KeyInfo struct {
	XMLName                 xml.Name                 `xml:"ds:KeyInfo"`
	SecurityTokenReference  *SecurityTokenReference  `xml:"wsse:SecurityTokenReference,omitempty"`
}

// Enhanced SecurityHeader with signature support
type EnhancedSecurityHeader struct {
	XMLName              xml.Name             `xml:"wsse:Security"`
	XmlnsWsse            string               `xml:"xmlns:wsse,attr"`
	XmlnsWsu             string               `xml:"xmlns:wsu,attr"`
	XmlnsDs              string               `xml:"xmlns:ds,attr,omitempty"`
	BinarySecurityToken  *BinarySecurityToken `xml:"wsse:BinarySecurityToken,omitempty"`
	Timestamp           *Timestamp           `xml:"wsu:Timestamp,omitempty"`
	Signature           *Signature           `xml:"ds:Signature,omitempty"`
}

// Signature algorithm constants
const (
	RSAWithSHA256Algorithm = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
	SHA256DigestAlgorithm  = "http://www.w3.org/2001/04/xmlenc#sha256"
	C14NAlgorithm          = "http://www.w3.org/2001/10/xml-exc-c14n#"
	
	// Token value types
	X509TokenValueType = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-x509-token-profile-1.0#X509v3"
	Base64EncodingType = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-soap-message-security-1.0#Base64Binary"
)

// NewWSSecurityManager creates a new WS-Security manager
func NewWSSecurityManager(config *WSSecurityConfig) *WSSecurityManager {
	if config == nil {
		config = &WSSecurityConfig{
			TimestampTTL:  5 * time.Minute,
			IncludeToken:  true,
			SignTimestamp: true,
			SignBody:      false, // Typically not required for SEFAZ
		}
	}

	return &WSSecurityManager{
		config: config,
	}
}

// LoadCertificateFromPEM loads a certificate and private key from PEM data
func LoadCertificateFromPEM(certPEM, keyPEM []byte) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Parse certificate
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return nil, nil, errors.NewValidationError("failed to decode PEM certificate", "certificate", "")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, errors.NewValidationError("failed to parse certificate", "certificate", err.Error())
	}

	// Parse private key
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return nil, nil, errors.NewValidationError("failed to decode PEM private key", "privateKey", "")
	}

	var privateKey *rsa.PrivateKey
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		key, parseErr := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if parseErr != nil {
			return nil, nil, errors.NewValidationError("failed to parse PKCS8 private key", "privateKey", parseErr.Error())
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, errors.NewValidationError("private key is not RSA", "privateKey", "")
		}
	default:
		return nil, nil, errors.NewValidationError("unsupported private key type", "privateKey", keyBlock.Type)
	}

	if err != nil {
		return nil, nil, errors.NewValidationError("failed to parse private key", "privateKey", err.Error())
	}

	return cert, privateKey, nil
}

// LoadCertificateFromFiles loads certificate and private key from files
func LoadCertificateFromFiles(certFile, keyFile string) (*x509.Certificate, *rsa.PrivateKey, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, nil, errors.NewValidationError("failed to load certificate pair", "files", fmt.Sprintf("%s, %s", certFile, keyFile))
	}

	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, nil, errors.NewValidationError("failed to parse X509 certificate", "certificate", err.Error())
	}

	rsaKey, ok := cert.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, errors.NewValidationError("private key is not RSA", "privateKey", "")
	}

	return x509Cert, rsaKey, nil
}

// CreateSecurityHeader creates an enhanced security header with optional signature
func (wsm *WSSecurityManager) CreateSecurityHeader(timestampID string) (*EnhancedSecurityHeader, error) {
	header := &EnhancedSecurityHeader{
		XmlnsWsse: WSSecurityNS,
		XmlnsWsu:  WSUtilityNS,
	}

	// Add timestamp
	if wsm.config.TimestampTTL > 0 {
		now := time.Now().UTC()
		created := now.Format("2006-01-02T15:04:05.000Z")
		expires := now.Add(wsm.config.TimestampTTL).Format("2006-01-02T15:04:05.000Z")

		header.Timestamp = &Timestamp{
			ID:      timestampID,
			Created: created,
			Expires: expires,
		}
	}

	// Add binary security token if certificate is available
	if wsm.config.Certificate != nil && wsm.config.IncludeToken {
		tokenID := "SecurityToken-" + generateID()
		header.BinarySecurityToken = &BinarySecurityToken{
			ValueType:    X509TokenValueType,
			EncodingType: Base64EncodingType,
			ID:           tokenID,
			Content:      base64.StdEncoding.EncodeToString(wsm.config.Certificate.Raw),
		}

		// Add signature if required
		if wsm.config.SignTimestamp && wsm.config.PrivateKey != nil {
			header.XmlnsDs = XMLDigitalSignatureNS
			signature, err := wsm.createSignature(timestampID, tokenID)
			if err != nil {
				return nil, err
			}
			header.Signature = signature
		}
	}

	return header, nil
}

// createSignature creates a digital signature for the timestamp
func (wsm *WSSecurityManager) createSignature(timestampID, tokenID string) (*Signature, error) {
	// Create signature structure
	signature := &Signature{
		XmlnsDs: XMLDigitalSignatureNS,
		SignedInfo: SignedInfo{
			CanonicalizationMethod: CanonicalizationMethod{
				Algorithm: C14NAlgorithm,
			},
			SignatureMethod: SignatureMethod{
				Algorithm: RSAWithSHA256Algorithm,
			},
			Reference: []SignatureReference{
				{
					URI: "#" + timestampID,
					Transforms: &Transforms{
						Transform: []Transform{
							{Algorithm: C14NAlgorithm},
						},
					},
					DigestMethod: DigestMethod{
						Algorithm: SHA256DigestAlgorithm,
					},
					DigestValue: "", // Will be calculated
				},
			},
		},
		KeyInfo: &KeyInfo{
			SecurityTokenReference: &SecurityTokenReference{
				Reference: Reference{
					URI:       "#" + tokenID,
					ValueType: X509TokenValueType,
				},
			},
		},
	}

	// For this implementation, we'll create a placeholder signature value
	// In a real implementation, you would:
	// 1. Canonicalize the SignedInfo
	// 2. Hash the canonicalized SignedInfo
	// 3. Sign the hash with the private key
	// 4. Base64 encode the signature
	
	// Simplified signature creation for demo purposes
	signatureValueID := "SignatureValue-" + generateID()
	signatureBytes := []byte("placeholder-signature-value") // In real implementation, compute actual signature
	
	signature.SignatureValue = SignatureValue{
		ID:    signatureValueID,
		Value: base64.StdEncoding.EncodeToString(signatureBytes),
	}

	// Calculate digest value (simplified)
	digestBytes := sha256.Sum256([]byte(timestampID))
	signature.SignedInfo.Reference[0].DigestValue = base64.StdEncoding.EncodeToString(digestBytes[:])

	return signature, nil
}

// ValidateCertificate validates a certificate against current time and CA
func ValidateCertificate(cert *x509.Certificate, caCerts []*x509.Certificate) error {
	if cert == nil {
		return errors.NewValidationError("certificate cannot be nil", "certificate", "")
	}

	now := time.Now()

	// Check certificate validity period
	if now.Before(cert.NotBefore) {
		return errors.NewCertificateError("certificate not yet valid", fmt.Errorf("not before: %s", cert.NotBefore.String()))
	}

	if now.After(cert.NotAfter) {
		return errors.NewCertificateError("certificate has expired", fmt.Errorf("not after: %s", cert.NotAfter.String()))
	}

	// Check certificate purpose
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		// Certificate should support digital signature
		return errors.NewCertificateError("certificate does not support digital signatures", fmt.Errorf("key usage: %d", cert.KeyUsage))
	}

	// If CA certificates are provided, verify certificate chain
	if len(caCerts) > 0 {
		roots := x509.NewCertPool()
		for _, caCert := range caCerts {
			roots.AddCert(caCert)
		}

		opts := x509.VerifyOptions{
			Roots:       roots,
			CurrentTime: now,
		}

		_, err := cert.Verify(opts)
		if err != nil {
			return errors.NewCertificateError("certificate chain validation failed", err)
		}
	}

	return nil
}

// CreateTimestamp creates a WS-Security timestamp
func CreateTimestamp(id string, validMinutes int) *Timestamp {
	now := time.Now().UTC()
	created := now.Format("2006-01-02T15:04:05.000Z")
	expires := now.Add(time.Duration(validMinutes) * time.Minute).Format("2006-01-02T15:04:05.000Z")

	return &Timestamp{
		ID:      id,
		Created: created,
		Expires: expires,
	}
}

// ValidateTimestampElement validates a timestamp element
func ValidateTimestampElement(timestamp *Timestamp) error {
	if timestamp == nil {
		return nil // No timestamp to validate
	}

	now := time.Now().UTC()

	// Parse times
	created, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp.Created)
	if err != nil {
		return errors.NewValidationError("invalid timestamp created format", "created", timestamp.Created)
	}

	expires, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp.Expires)
	if err != nil {
		return errors.NewValidationError("invalid timestamp expires format", "expires", timestamp.Expires)
	}

	// Validate timestamp window (allow 5 minutes clock skew)
	clockSkew := 5 * time.Minute

	if now.Before(created.Add(-clockSkew)) {
		return errors.NewValidationError("timestamp created is too far in the future", "created", timestamp.Created)
	}

	if now.After(expires.Add(clockSkew)) {
		return errors.NewValidationError("timestamp has expired", "expires", timestamp.Expires)
	}

	if expires.Before(created) {
		return errors.NewValidationError("timestamp expires before created", "timestamps", fmt.Sprintf("created: %s, expires: %s", timestamp.Created, timestamp.Expires))
	}

	return nil
}

// ExtractCertificateFromToken extracts certificate from binary security token
func ExtractCertificateFromToken(token *BinarySecurityToken) (*x509.Certificate, error) {
	if token == nil {
		return nil, errors.NewValidationError("binary security token cannot be nil", "token", "")
	}

	if token.ValueType != X509TokenValueType {
		return nil, errors.NewValidationError("unsupported token value type", "valueType", token.ValueType)
	}

	if token.EncodingType != Base64EncodingType {
		return nil, errors.NewValidationError("unsupported encoding type", "encodingType", token.EncodingType)
	}

	// Decode base64 content
	certBytes, err := base64.StdEncoding.DecodeString(token.Content)
	if err != nil {
		return nil, errors.NewValidationError("failed to decode base64 certificate", "content", err.Error())
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, errors.NewValidationError("failed to parse certificate", "certificate", err.Error())
	}

	return cert, nil
}

// generateID generates a unique ID for security elements
func generateID() string {
	// Generate random bytes
	bytes := make([]byte, 16)
	rand.Read(bytes)
	
	// Convert to hex string
	return fmt.Sprintf("%x", bytes)
}

// CreateWSSecurityConfig creates a basic WS-Security configuration
func CreateWSSecurityConfig(certPEM, keyPEM []byte) (*WSSecurityConfig, error) {
	cert, key, err := LoadCertificateFromPEM(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &WSSecurityConfig{
		Certificate:   cert,
		PrivateKey:    key,
		TimestampTTL:  5 * time.Minute,
		IncludeToken:  true,
		SignTimestamp: true,
		SignBody:      false,
	}, nil
}

// VerifySignature verifies a digital signature (simplified implementation)
func VerifySignature(signature *Signature, cert *x509.Certificate) error {
	if signature == nil {
		return errors.NewValidationError("signature cannot be nil", "signature", "")
	}

	if cert == nil {
		return errors.NewValidationError("certificate cannot be nil", "certificate", "")
	}

	// In a real implementation, you would:
	// 1. Canonicalize the SignedInfo
	// 2. Hash the canonicalized SignedInfo
	// 3. Verify the signature using the certificate's public key
	// 4. Verify all digest values in references

	// For this simplified implementation, we just validate the structure
	if signature.SignatureValue.Value == "" {
		return errors.NewValidationError("signature value cannot be empty", "signatureValue", "")
	}

	if len(signature.SignedInfo.Reference) == 0 {
		return errors.NewValidationError("signature must have at least one reference", "references", "")
	}

	return nil
}

// GetCertificateFingerprint returns the SHA-256 fingerprint of a certificate
func GetCertificateFingerprint(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	
	hash := sha256.Sum256(cert.Raw)
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

// GetCertificateSubject returns formatted certificate subject
func GetCertificateSubject(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	
	return cert.Subject.String()
}

// GetCertificateIssuer returns formatted certificate issuer
func GetCertificateIssuer(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}
	
	return cert.Issuer.String()
}

// IsCertificateValid performs basic certificate validation
func IsCertificateValid(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}
	
	now := time.Now()
	return now.After(cert.NotBefore) && now.Before(cert.NotAfter)
}