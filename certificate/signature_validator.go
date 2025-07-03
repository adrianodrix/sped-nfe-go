package certificate

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/beevik/etree"
)

// SignatureValidator provides comprehensive signature validation capabilities
type SignatureValidator struct {
	config           *ValidationConfig
	trustedCerts     []*x509.Certificate
	crlCache         map[string]*x509.RevocationList
	ocspCache        map[string]*OCSPResponse
	timestampService string
}

// ValidationConfig holds configuration for signature validation
type ValidationConfig struct {
	// RequireValidCertificate determines if certificate validity is required
	RequireValidCertificate bool `json:"requireValidCertificate"`

	// RequireICPBrasil determines if certificate must be from ICP-Brasil
	RequireICPBrasil bool `json:"requireIcpBrasil"`

	// CheckRevocation determines if certificate revocation should be checked
	CheckRevocation bool `json:"checkRevocation"`

	// MaxClockSkew allows for minor time differences
	MaxClockSkew time.Duration `json:"maxClockSkew"`

	// AllowedSignatureAlgorithms specifies which signature algorithms are acceptable
	AllowedSignatureAlgorithms []string `json:"allowedSignatureAlgorithms"`

	// AllowedDigestAlgorithms specifies which digest algorithms are acceptable
	AllowedDigestAlgorithms []string `json:"allowedDigestAlgorithms"`

	// RequireTimestamp determines if timestamp is required
	RequireTimestamp bool `json:"requireTimestamp"`

	// TrustedRoots specifies trusted root certificates
	TrustedRoots []*x509.Certificate `json:"trustedRoots"`
}

// ValidationResult contains the result of signature validation
type ValidationResult struct {
	IsValid            bool                  `json:"isValid"`
	SignatureValid     bool                  `json:"signatureValid"`
	CertificateValid   bool                  `json:"certificateValid"`
	TimestampValid     bool                  `json:"timestampValid"`
	TrustedChain       bool                  `json:"trustedChain"`
	Certificate        *x509.Certificate     `json:"certificate"`
	SignatureAlgorithm string                `json:"signatureAlgorithm"`
	DigestAlgorithm    string                `json:"digestAlgorithm"`
	SigningTime        *time.Time            `json:"signingTime"`
	Errors             []string              `json:"errors"`
	Warnings           []string              `json:"warnings"`
	SignatureInfo      *XMLDSigSignatureInfo `json:"signatureInfo"`
}

// XMLDSigSignatureInfo contains detailed information about an XMLDSig signature
type XMLDSigSignatureInfo struct {
	SignatureMethod        string          `json:"signatureMethod"`
	CanonicalizationMethod string          `json:"canonicalizationMethod"`
	DigestMethod           string          `json:"digestMethod"`
	Transforms             []string        `json:"transforms"`
	References             []ReferenceInfo `json:"references"`
	KeyInfo                *KeyInfo        `json:"keyInfo"`
	SignatureValue         string          `json:"signatureValue"`
	DigestValue            string          `json:"digestValue"`
}

// KeyInfo contains information about the signing key
type KeyInfo struct {
	Certificate *CertificateInfo `json:"certificate"`
	KeyValue    *KeyValue        `json:"keyValue"`
	HasX509Data bool             `json:"hasX509Data"`
	HasKeyValue bool             `json:"hasKeyValue"`
}

// KeyValue contains raw key information
type KeyValue struct {
	RSAKeyValue *RSAKeyValue `json:"rsaKeyValue"`
}

// RSAKeyValue contains RSA key components
type RSAKeyValue struct {
	Modulus  string `json:"modulus"`
	Exponent string `json:"exponent"`
}

// OCSPResponse represents an OCSP response for certificate validation
type OCSPResponse struct {
	Status      int        `json:"status"`
	RevokedAt   *time.Time `json:"revokedAt"`
	NextUpdate  time.Time  `json:"nextUpdate"`
	LastChecked time.Time  `json:"lastChecked"`
}

// NewSignatureValidator creates a new signature validator
func NewSignatureValidator(config *ValidationConfig) *SignatureValidator {
	if config == nil {
		config = DefaultValidationConfig()
	}

	return &SignatureValidator{
		config:           config,
		trustedCerts:     config.TrustedRoots,
		crlCache:         make(map[string]*x509.RevocationList),
		ocspCache:        make(map[string]*OCSPResponse),
		timestampService: "http://timestamp.sectigo.com", // Default timestamp service
	}
}

// DefaultValidationConfig returns a default validation configuration for SEFAZ compliance
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		RequireValidCertificate: true,
		RequireICPBrasil:        true,
		CheckRevocation:         false, // Disabled by default for performance
		MaxClockSkew:            5 * time.Minute,
		AllowedSignatureAlgorithms: []string{
			"http://www.w3.org/2000/09/xmldsig#rsa-sha1",
			"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
		},
		AllowedDigestAlgorithms: []string{
			"http://www.w3.org/2000/09/xmldsig#sha1",
			"http://www.w3.org/2001/04/xmlenc#sha256",
		},
		RequireTimestamp: false,
		TrustedRoots:     GetICPBrasilRootCertificates(),
	}
}

// ValidateXMLSignature validates an XML signature comprehensively
func (validator *SignatureValidator) ValidateXMLSignature(signedXML string) (*ValidationResult, error) {
	if signedXML == "" {
		return nil, errors.NewValidationError("signed XML cannot be empty", "signedXML", "")
	}

	result := &ValidationResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse XML: %v", err))
		return result, nil
	}

	// Find signature element
	sigElement := doc.FindElement(".//ds:Signature")
	if sigElement == nil {
		result.Errors = append(result.Errors, "No XMLDSig signature found in document")
		return result, nil
	}

	// Extract signature information
	sigInfo, err := validator.extractSignatureInfo(sigElement)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to extract signature info: %v", err))
		return result, nil
	}
	result.SignatureInfo = sigInfo

	// Validate signature structure
	if !validator.validateSignatureStructure(sigElement, result) {
		return result, nil
	}

	// Extract and validate certificate
	cert, err := validator.extractCertificate(sigElement)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to extract certificate: %v", err))
		return result, nil
	}
	result.Certificate = cert

	// Validate certificate
	validator.validateCertificate(cert, result)

	// Validate signature value
	validator.validateSignatureValue(doc, sigElement, cert, result)

	// Validate references and digest values
	validator.validateReferences(doc, sigElement, result)

	// Set overall validity
	result.IsValid = result.SignatureValid && result.CertificateValid && len(result.Errors) == 0

	return result, nil
}

// ValidateDetachedSignature validates a detached signature against external content
func (validator *SignatureValidator) ValidateDetachedSignature(signatureXML string, content []byte, referenceURI string) (*ValidationResult, error) {
	result, err := validator.ValidateXMLSignature(signatureXML)
	if err != nil {
		return result, err
	}

	if result.SignatureInfo == nil {
		result.Errors = append(result.Errors, "No signature information available")
		return result, nil
	}

	// Find the reference that matches the URI
	var targetRef *ReferenceInfo
	for _, ref := range result.SignatureInfo.References {
		if ref.URI == referenceURI {
			targetRef = &ref
			break
		}
	}

	if targetRef == nil {
		result.Errors = append(result.Errors, fmt.Sprintf("No reference found for URI: %s", referenceURI))
		return result, nil
	}

	// Validate digest of external content
	calculatedDigest := validator.calculateDigest(content, targetRef.DigestMethod)
	if calculatedDigest != targetRef.DigestValue {
		result.Errors = append(result.Errors, "Digest validation failed for external content")
		result.SignatureValid = false
	}

	result.IsValid = result.SignatureValid && result.CertificateValid && len(result.Errors) == 0
	return result, nil
}

// extractSignatureInfo extracts detailed signature information from signature element
func (validator *SignatureValidator) extractSignatureInfo(sigElement *etree.Element) (*XMLDSigSignatureInfo, error) {
	info := &XMLDSigSignatureInfo{
		References: []ReferenceInfo{},
		Transforms: []string{},
	}

	// Extract signature method
	if sigMethodElem := sigElement.FindElement(".//ds:SignatureMethod"); sigMethodElem != nil {
		if algAttr := sigMethodElem.SelectAttr("Algorithm"); algAttr != nil {
			info.SignatureMethod = algAttr.Value
		}
	}

	// Extract canonicalization method
	if canonElem := sigElement.FindElement(".//ds:CanonicalizationMethod"); canonElem != nil {
		if algAttr := canonElem.SelectAttr("Algorithm"); algAttr != nil {
			info.CanonicalizationMethod = algAttr.Value
		}
	}

	// Extract signature value
	if sigValueElem := sigElement.FindElement(".//ds:SignatureValue"); sigValueElem != nil {
		info.SignatureValue = strings.TrimSpace(sigValueElem.Text())
	}

	// Extract references
	refElements := sigElement.FindElements(".//ds:Reference")
	for _, refElem := range refElements {
		refInfo := ReferenceInfo{}

		if uriAttr := refElem.SelectAttr("URI"); uriAttr != nil {
			refInfo.URI = uriAttr.Value
		}

		if digestMethodElem := refElem.FindElement(".//ds:DigestMethod"); digestMethodElem != nil {
			if algAttr := digestMethodElem.SelectAttr("Algorithm"); algAttr != nil {
				refInfo.DigestMethod = algAttr.Value
				info.DigestMethod = algAttr.Value // Set overall digest method
			}
		}

		if digestValueElem := refElem.FindElement(".//ds:DigestValue"); digestValueElem != nil {
			refInfo.DigestValue = strings.TrimSpace(digestValueElem.Text())
			info.DigestValue = refInfo.DigestValue // Set overall digest value
		}

		info.References = append(info.References, refInfo)
	}

	// Extract transforms
	transformElements := sigElement.FindElements(".//ds:Transform")
	for _, transformElem := range transformElements {
		if algAttr := transformElem.SelectAttr("Algorithm"); algAttr != nil {
			info.Transforms = append(info.Transforms, algAttr.Value)
		}
	}

	// Extract key info
	if keyInfoElem := sigElement.FindElement(".//ds:KeyInfo"); keyInfoElem != nil {
		keyInfo := &KeyInfo{
			HasX509Data: keyInfoElem.FindElement(".//ds:X509Data") != nil,
			HasKeyValue: keyInfoElem.FindElement(".//ds:KeyValue") != nil,
		}
		info.KeyInfo = keyInfo
	}

	return info, nil
}

// validateSignatureStructure validates the basic structure of the signature
func (validator *SignatureValidator) validateSignatureStructure(sigElement *etree.Element, result *ValidationResult) bool {
	valid := true

	// Check for required elements
	if sigElement.FindElement(".//ds:SignedInfo") == nil {
		result.Errors = append(result.Errors, "Missing SignedInfo element")
		valid = false
	}

	if sigElement.FindElement(".//ds:SignatureValue") == nil {
		result.Errors = append(result.Errors, "Missing SignatureValue element")
		valid = false
	}

	if sigElement.FindElement(".//ds:Reference") == nil {
		result.Errors = append(result.Errors, "Missing Reference element")
		valid = false
	}

	// Validate signature algorithm
	if sigMethodElem := sigElement.FindElement(".//ds:SignatureMethod"); sigMethodElem != nil {
		if algAttr := sigMethodElem.SelectAttr("Algorithm"); algAttr != nil {
			result.SignatureAlgorithm = algAttr.Value
			if !validator.isAllowedSignatureAlgorithm(algAttr.Value) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Signature algorithm not in allowed list: %s", algAttr.Value))
			}
		}
	}

	// Validate digest algorithm
	if digestMethodElem := sigElement.FindElement(".//ds:DigestMethod"); digestMethodElem != nil {
		if algAttr := digestMethodElem.SelectAttr("Algorithm"); algAttr != nil {
			result.DigestAlgorithm = algAttr.Value
			if !validator.isAllowedDigestAlgorithm(algAttr.Value) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Digest algorithm not in allowed list: %s", algAttr.Value))
			}
		}
	}

	return valid
}

// extractCertificate extracts the certificate from the signature
func (validator *SignatureValidator) extractCertificate(sigElement *etree.Element) (*x509.Certificate, error) {
	certElem := sigElement.FindElement(".//ds:X509Certificate")
	if certElem == nil {
		return nil, fmt.Errorf("no certificate found in signature")
	}

	certData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(certElem.Text()))
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return cert, nil
}

// validateCertificate validates the signing certificate
func (validator *SignatureValidator) validateCertificate(cert *x509.Certificate, result *ValidationResult) {
	if cert == nil {
		result.Errors = append(result.Errors, "Certificate is nil")
		return
	}

	now := time.Now()

	// Check certificate validity period
	if now.Before(cert.NotBefore.Add(-validator.config.MaxClockSkew)) {
		result.Errors = append(result.Errors, "Certificate is not yet valid")
		result.CertificateValid = false
		return
	}

	if now.After(cert.NotAfter.Add(validator.config.MaxClockSkew)) {
		result.Errors = append(result.Errors, "Certificate has expired")
		result.CertificateValid = false
		return
	}

	// Check if certificate is suitable for digital signing
	if !IsCertificateValidForSigning(cert) {
		result.Errors = append(result.Errors, "Certificate is not valid for digital signing")
		result.CertificateValid = false
		return
	}

	// Check ICP-Brasil requirement
	if validator.config.RequireICPBrasil && !IsICPBrasilCertificate(cert) {
		result.Errors = append(result.Errors, "Certificate is not from ICP-Brasil")
		result.CertificateValid = false
		return
	}

	// Additional validations can be added here (CRL, OCSP, etc.)
	result.CertificateValid = true
}

// validateSignatureValue validates the signature value against the SignedInfo
func (validator *SignatureValidator) validateSignatureValue(doc *etree.Document, sigElement *etree.Element, cert *x509.Certificate, result *ValidationResult) {
	if cert == nil {
		result.Errors = append(result.Errors, "Certificate not available for signature validation")
		return
	}

	// Extract SignedInfo element
	signedInfoElem := sigElement.FindElement(".//ds:SignedInfo")
	if signedInfoElem == nil {
		result.Errors = append(result.Errors, "SignedInfo element not found")
		return
	}

	// Extract signature value
	sigValueElem := sigElement.FindElement(".//ds:SignatureValue")
	if sigValueElem == nil {
		result.Errors = append(result.Errors, "SignatureValue element not found")
		return
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(sigValueElem.Text()))
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to decode signature value: %v", err))
		return
	}

	// Canonicalize SignedInfo (simplified canonicalization)
	tempDoc := etree.NewDocument()
	tempDoc.SetRoot(signedInfoElem.Copy())
	signedInfoBytes, err := tempDoc.WriteToBytes()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to serialize SignedInfo: %v", err))
		return
	}

	// Get the hash algorithm from signature method
	hashAlg := validator.getHashAlgorithmFromSignatureMethod(result.SignatureAlgorithm)
	if hashAlg == 0 {
		result.Errors = append(result.Errors, "Unsupported signature algorithm")
		return
	}

	// Verify signature
	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		result.Errors = append(result.Errors, "Certificate public key is not RSA")
		return
	}

	// Hash the SignedInfo
	hasher := hashAlg.New()
	hasher.Write(signedInfoBytes)
	hashed := hasher.Sum(nil)

	// Verify the signature
	err = rsa.VerifyPKCS1v15(pubKey, hashAlg, hashed, signatureBytes)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Signature verification failed: %v", err))
		result.SignatureValid = false
		return
	}

	result.SignatureValid = true
}

// validateReferences validates all reference elements and their digest values
func (validator *SignatureValidator) validateReferences(doc *etree.Document, sigElement *etree.Element, result *ValidationResult) {
	refElements := sigElement.FindElements(".//ds:Reference")

	for _, refElem := range refElements {
		if !validator.validateSingleReference(doc, sigElement, refElem, result) {
			result.SignatureValid = false
		}
	}
}

// validateSingleReference validates a single reference element
func (validator *SignatureValidator) validateSingleReference(doc *etree.Document, sigElement *etree.Element, refElem *etree.Element, result *ValidationResult) bool {
	// Get reference URI
	var refURI string
	if uriAttr := refElem.SelectAttr("URI"); uriAttr != nil {
		refURI = uriAttr.Value
	}

	// Get expected digest value
	digestValueElem := refElem.FindElement(".//ds:DigestValue")
	if digestValueElem == nil {
		result.Errors = append(result.Errors, fmt.Sprintf("DigestValue not found for reference %s", refURI))
		return false
	}
	expectedDigest := strings.TrimSpace(digestValueElem.Text())

	// Get digest method
	digestMethodElem := refElem.FindElement(".//ds:DigestMethod")
	if digestMethodElem == nil {
		result.Errors = append(result.Errors, fmt.Sprintf("DigestMethod not found for reference %s", refURI))
		return false
	}
	digestMethod := digestMethodElem.SelectAttr("Algorithm").Value

	// Find and canonicalize the referenced element
	var elementToDigest []byte
	var err error

	if refURI == "" {
		// Reference to entire document
		docCopy := doc.Copy()
		// Remove signature element for digest calculation
		if sigElem := docCopy.FindElement(".//ds:Signature"); sigElem != nil {
			sigElem.Parent().RemoveChild(sigElem)
		}
		elementToDigest, err = docCopy.WriteToBytes()
	} else if strings.HasPrefix(refURI, "#") {
		// Reference to element by ID
		elementID := strings.TrimPrefix(refURI, "#")
		element := validator.findElementByID(doc, elementID)
		if element == nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Referenced element not found: %s", elementID))
			return false
		}

		// Create temporary document for the element (without signature)
		tempDoc := etree.NewDocument()
		elementCopy := element.Copy()
		// Remove signature if it's inside the element
		if sigElem := elementCopy.FindElement(".//ds:Signature"); sigElem != nil {
			elementCopy.RemoveChild(sigElem)
		}
		tempDoc.SetRoot(elementCopy)
		elementToDigest, err = tempDoc.WriteToBytes()
	} else {
		result.Warnings = append(result.Warnings, fmt.Sprintf("External reference not supported: %s", refURI))
		return true // Skip validation for external references
	}

	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to serialize referenced element: %v", err))
		return false
	}

	// Calculate digest
	calculatedDigest := validator.calculateDigest(elementToDigest, digestMethod)

	// Compare digests
	if calculatedDigest != expectedDigest {
		result.Errors = append(result.Errors, fmt.Sprintf("Digest mismatch for reference %s", refURI))
		return false
	}

	return true
}

// Helper methods

func (validator *SignatureValidator) isAllowedSignatureAlgorithm(algorithm string) bool {
	for _, allowed := range validator.config.AllowedSignatureAlgorithms {
		if allowed == algorithm {
			return true
		}
	}
	return false
}

func (validator *SignatureValidator) isAllowedDigestAlgorithm(algorithm string) bool {
	for _, allowed := range validator.config.AllowedDigestAlgorithms {
		if allowed == algorithm {
			return true
		}
	}
	return false
}

func (validator *SignatureValidator) getHashAlgorithmFromSignatureMethod(signatureMethod string) crypto.Hash {
	switch signatureMethod {
	case "http://www.w3.org/2000/09/xmldsig#rsa-sha1":
		return crypto.SHA1
	case "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256":
		return crypto.SHA256
	default:
		return 0
	}
}

func (validator *SignatureValidator) calculateDigest(data []byte, digestMethod string) string {
	var hash []byte

	switch digestMethod {
	case "http://www.w3.org/2000/09/xmldsig#sha1":
		h := sha1.Sum(data)
		hash = h[:]
	case "http://www.w3.org/2001/04/xmlenc#sha256":
		h := sha256.Sum256(data)
		hash = h[:]
	default:
		// Default to SHA1 for compatibility
		h := sha1.Sum(data)
		hash = h[:]
	}

	return base64.StdEncoding.EncodeToString(hash)
}

func (validator *SignatureValidator) findElementByID(doc *etree.Document, elementID string) *etree.Element {
	// Try different common ID attribute names
	selectors := []string{
		fmt.Sprintf(".//*[@Id='%s']", elementID),
		fmt.Sprintf(".//*[@id='%s']", elementID),
		fmt.Sprintf(".//*[@ID='%s']", elementID),
	}

	for _, selector := range selectors {
		if element := doc.FindElement(selector); element != nil {
			return element
		}
	}

	return nil
}

// ValidateNFeSignature validates an NFe signature specifically
func ValidateNFeSignature(nfeXML string) (*ValidationResult, error) {
	validator := NewSignatureValidator(DefaultValidationConfig())
	return validator.ValidateXMLSignature(nfeXML)
}

// QuickValidateSignature performs a quick signature validation
func QuickValidateSignature(signedXML string) (bool, error) {
	validator := NewSignatureValidator(&ValidationConfig{
		RequireValidCertificate: false,
		RequireICPBrasil:        false,
		CheckRevocation:         false,
	})

	result, err := validator.ValidateXMLSignature(signedXML)
	if err != nil {
		return false, err
	}

	return result.SignatureValid, nil
}
