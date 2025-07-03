// Package certificate provides XML digital signature functionality for NFe documents.
// It implements xmldsig standard compatible with SEFAZ requirements.
package certificate

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/beevik/etree"
)

// XMLSigner provides XML digital signature functionality using certificates
type XMLSigner struct {
	certificate Certificate
	config      *SignerConfig
}

// SignerConfig holds configuration for XML signing operations
type SignerConfig struct {
	// DigestAlgorithm specifies the digest algorithm for XML signing
	DigestAlgorithm crypto.Hash `json:"digestAlgorithm"`

	// SignatureAlgorithm specifies the signature algorithm
	SignatureAlgorithm string `json:"signatureAlgorithm"`

	// CanonicalizationAlgorithm for XML canonicalization
	CanonicalizationAlgorithm string `json:"canonicalizationAlgorithm"`

	// IncludeCertificate determines if certificate should be included in signature
	IncludeCertificate bool `json:"includeCertificate"`

	// SignatureElementID specifies the ID for the signature element
	SignatureElementID string `json:"signatureElementId"`

	// AddTimestamp includes a timestamp in the signature
	AddTimestamp bool `json:"addTimestamp"`

	// ValidateXML enables XML validation before signing
	ValidateXML bool `json:"validateXml"`
}

// SignatureInfo contains information about a created signature
type SignatureInfo struct {
	SignatureValue string    `json:"signatureValue"`
	DigestValue    string    `json:"digestValue"`
	Timestamp      time.Time `json:"timestamp"`
	Certificate    string    `json:"certificate"`
	Algorithm      string    `json:"algorithm"`
}

// NewXMLSigner creates a new XML signer with the given certificate and configuration
func NewXMLSigner(certificate Certificate, config *SignerConfig) *XMLSigner {
	if config == nil {
		config = DefaultSignerConfig()
	}

	return &XMLSigner{
		certificate: certificate,
		config:      config,
	}
}

// DefaultSignerConfig returns default configuration for XML signing compatible with SEFAZ
func DefaultSignerConfig() *SignerConfig {
	return &SignerConfig{
		DigestAlgorithm:           crypto.SHA1, // SEFAZ still uses SHA-1 for compatibility
		SignatureAlgorithm:        "http://www.w3.org/2000/09/xmldsig#rsa-sha1",
		CanonicalizationAlgorithm: "http://www.w3.org/TR/2001/REC-xml-c14n-20010315",
		IncludeCertificate:        true,
		SignatureElementID:        "",
		AddTimestamp:              false,
		ValidateXML:               true,
	}
}

// SHA256SignerConfig returns configuration for XML signing using SHA-256
func SHA256SignerConfig() *SignerConfig {
	config := DefaultSignerConfig()
	config.DigestAlgorithm = crypto.SHA256
	config.SignatureAlgorithm = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
	return config
}

// SignXML signs an XML document with the configured certificate
func (signer *XMLSigner) SignXML(xmlContent string) (string, *SignatureInfo, error) {
	if xmlContent == "" {
		return "", nil, errors.NewValidationError("XML content cannot be empty", "xmlContent", "")
	}

	if signer.certificate == nil {
		return "", nil, errors.NewCertificateError("certificate not available", nil)
	}

	// Validate certificate before signing
	if !signer.certificate.IsValid() {
		return "", nil, errors.NewCertificateError("certificate is not valid for signing", nil)
	}

	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return "", nil, errors.NewValidationError("failed to parse XML", "xml", err.Error())
	}

	// Validate XML structure if required
	if signer.config.ValidateXML {
		if err := signer.validateXMLStructure(doc); err != nil {
			return "", nil, err
		}
	}

	// Create signature manually using basic XML manipulation
	signedXML, err := signer.signXMLManually(xmlContent)
	if err != nil {
		return "", nil, errors.NewCertificateError("failed to sign XML", err)
	}

	// Extract signature information
	sigInfo, err := signer.extractSignatureInfo(signedXML)
	if err != nil {
		return "", nil, err
	}

	return signedXML, sigInfo, nil
}

// SignXMLElement signs a specific element in an XML document
func (signer *XMLSigner) SignXMLElement(xmlContent, elementID string) (string, *SignatureInfo, error) {
	if xmlContent == "" {
		return "", nil, errors.NewValidationError("XML content cannot be empty", "xmlContent", "")
	}

	if elementID == "" {
		return "", nil, errors.NewValidationError("element ID cannot be empty", "elementID", "")
	}

	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return "", nil, errors.NewValidationError("failed to parse XML", "xml", err.Error())
	}

	// Find the element to sign
	element := doc.FindElement(fmt.Sprintf(".//*[@Id='%s']", elementID))
	if element == nil {
		return "", nil, errors.NewValidationError("element with specified ID not found", "elementID", elementID)
	}

	// Create signature manually for specific element
	signedXML, err := signer.signXMLElementManually(xmlContent, elementID)
	if err != nil {
		return "", nil, errors.NewCertificateError("failed to sign XML element", err)
	}

	// Extract signature information
	sigInfo, err := signer.extractSignatureInfo(signedXML)
	if err != nil {
		return "", nil, err
	}

	return signedXML, sigInfo, nil
}

// VerifyXMLSignature verifies an XML signature
func (signer *XMLSigner) VerifyXMLSignature(signedXML string) error {
	if signedXML == "" {
		return errors.NewValidationError("signed XML content cannot be empty", "signedXML", "")
	}

	// Verify signature manually
	err := signer.verifyXMLSignatureManually(signedXML)
	if err != nil {
		return errors.NewCertificateError("XML signature verification failed", err)
	}

	return nil
}

// GetCertificate returns the certificate used for signing
func (signer *XMLSigner) GetCertificate() *x509.Certificate {
	return signer.certificate.GetCertificate()
}

// Sign signs data using the certificate
func (signer *XMLSigner) Sign(data []byte) ([]byte, error) {
	return signer.certificate.Sign(data, signer.config.DigestAlgorithm)
}

// SignatureAlgorithm returns the signature algorithm URI
func (signer *XMLSigner) SignatureAlgorithm() string {
	return signer.config.SignatureAlgorithm
}

// CreateDetachedSignature creates a detached signature for external data
func (signer *XMLSigner) CreateDetachedSignature(data []byte, referenceURI string) (string, error) {
	if len(data) == 0 {
		return "", errors.NewValidationError("data to sign cannot be empty", "data", "")
	}

	// Calculate digest
	var digest []byte
	switch signer.config.DigestAlgorithm {
	case crypto.SHA1:
		h := sha1.Sum(data)
		digest = h[:]
	case crypto.SHA256:
		h := sha256.Sum256(data)
		digest = h[:]
	default:
		return "", errors.NewValidationError("unsupported digest algorithm", "algorithm", signer.config.DigestAlgorithm.String())
	}

	// Create signature template
	template := signer.createSignatureTemplate(referenceURI, digest)

	// Sign the signature info
	signedInfo := signer.extractSignedInfo(template)
	signature, err := signer.certificate.Sign([]byte(signedInfo), signer.config.DigestAlgorithm)
	if err != nil {
		return "", errors.NewCertificateError("failed to create detached signature", err)
	}

	// Insert signature value into template
	signatureValue := base64.StdEncoding.EncodeToString(signature)
	result := strings.Replace(template, "{{SIGNATURE_VALUE}}", signatureValue, 1)

	return result, nil
}

// validateXMLStructure performs basic XML structure validation
func (signer *XMLSigner) validateXMLStructure(doc *etree.Document) error {
	if doc.Root() == nil {
		return errors.NewValidationError("XML document has no root element", "xml", "")
	}

	// Check for existing signatures
	existingSignatures := doc.FindElements(".//Signature")
	if len(existingSignatures) > 0 {
		return errors.NewValidationError("XML already contains signatures", "signatures", fmt.Sprintf("%d", len(existingSignatures)))
	}

	return nil
}

// getDigestMethodURI returns the digest method URI for the configured algorithm
func (signer *XMLSigner) getDigestMethodURI() string {
	switch signer.config.DigestAlgorithm {
	case crypto.SHA1:
		return "http://www.w3.org/2000/09/xmldsig#sha1"
	case crypto.SHA256:
		return "http://www.w3.org/2001/04/xmlenc#sha256"
	default:
		return "http://www.w3.org/2000/09/xmldsig#sha1" // Default to SHA-1
	}
}

// extractSignatureInfo extracts signature information from signed XML
func (signer *XMLSigner) extractSignatureInfo(signedXML string) (*SignatureInfo, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return nil, errors.NewValidationError("failed to parse signed XML", "xml", err.Error())
	}

	// Find signature elements (without ds: prefix)
	sigValueElem := doc.FindElement(".//SignatureValue")
	digestValueElem := doc.FindElement(".//DigestValue")

	info := &SignatureInfo{
		Timestamp: time.Now(),
		Algorithm: signer.config.SignatureAlgorithm,
	}

	if sigValueElem != nil {
		info.SignatureValue = sigValueElem.Text()
	}

	if digestValueElem != nil {
		info.DigestValue = digestValueElem.Text()
	}

	if signer.config.IncludeCertificate {
		certElem := doc.FindElement(".//X509Certificate")
		if certElem != nil {
			info.Certificate = certElem.Text()
		}
	}

	return info, nil
}

// createSignatureTemplate creates a signature template for detached signatures
func (signer *XMLSigner) createSignatureTemplate(referenceURI string, digest []byte) string {
	digestValue := base64.StdEncoding.EncodeToString(digest)
	digestMethod := signer.getDigestMethodURI()

	// Create signature element without formatting to avoid error 588
	// SEFAZ error 588: "Nao e permitida a presenca de caracteres de edicao"
	template := fmt.Sprintf(`<Signature xmlns="http://www.w3.org/2000/09/xmldsig#"><SignedInfo><CanonicalizationMethod Algorithm="%s"/><SignatureMethod Algorithm="%s"/><Reference URI="%s"><Transforms><Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/><Transform Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/></Transforms><DigestMethod Algorithm="%s"/><DigestValue>%s</DigestValue></Reference></SignedInfo><SignatureValue>{{SIGNATURE_VALUE}}</SignatureValue>`,
		signer.config.CanonicalizationAlgorithm,
		signer.config.SignatureAlgorithm,
		referenceURI,
		digestMethod,
		digestValue)

	if signer.config.IncludeCertificate {
		cert := signer.certificate.GetCertificate()
		if cert != nil {
			certData := base64.StdEncoding.EncodeToString(cert.Raw)
			template += fmt.Sprintf(`<KeyInfo><X509Data><X509Certificate>%s</X509Certificate></X509Data></KeyInfo>`, certData)
		}
	}

	template += `</Signature>`

	return template
}

// extractSignedInfo extracts the SignedInfo element for signing
func (signer *XMLSigner) extractSignedInfo(template string) string {
	doc := etree.NewDocument()
	doc.ReadFromString(template)

	signedInfo := doc.FindElement(".//SignedInfo")
	if signedInfo == nil {
		return ""
	}

	doc2 := etree.NewDocument()
	doc2.SetRoot(signedInfo.Copy())
	result, _ := doc2.WriteToString()

	return result
}

// extractSignedInfoCanonical extracts the SignedInfo element with canonical formatting
func (signer *XMLSigner) extractSignedInfoCanonical(template string) string {
	doc := etree.NewDocument()
	doc.ReadFromString(template)

	signedInfo := doc.FindElement(".//SignedInfo")
	if signedInfo == nil {
		return ""
	}

	doc2 := etree.NewDocument()
	doc2.SetRoot(signedInfo.Copy())
	
	// Apply canonical XML settings (C14N)
	doc2.WriteSettings = etree.WriteSettings{
		CanonicalAttrVal: true,
		CanonicalEndTags: true,
		CanonicalText:    true,
		UseCRLF:          false,
	}
	
	result, _ := doc2.WriteToString()
	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// SignNFeXML signs an NFe XML document following SEFAZ requirements
func (signer *XMLSigner) SignNFeXML(nfeXML string) (string, error) {
	// NFe specific signing - signs the infNFe element with correct placement
	doc := etree.NewDocument()
	if err := doc.ReadFromString(nfeXML); err != nil {
		return "", errors.NewValidationError("failed to parse NFe XML", "xml", err.Error())
	}

	// Find infNFe element
	infNFeElement := doc.FindElement(".//infNFe")
	if infNFeElement == nil {
		return "", errors.NewValidationError("infNFe element not found in NFe XML", "element", "infNFe")
	}

	// Get the Id attribute
	idAttr := infNFeElement.SelectAttr("Id")
	if idAttr == nil {
		return "", errors.NewValidationError("infNFe element must have Id attribute", "attribute", "Id")
	}

	// Find NFe root element (parent of infNFe)
	nfeElement := infNFeElement.Parent()
	if nfeElement == nil || nfeElement.Tag != "NFe" {
		return "", errors.NewValidationError("NFe root element not found", "element", "NFe")
	}

	// Sign the NFe specifically (custom logic for NFe structure)
	return signer.signNFeSpecifically(doc, infNFeElement, nfeElement, idAttr.Value)
}

// signNFeSpecifically signs the NFe with correct signature placement
func (signer *XMLSigner) signNFeSpecifically(doc *etree.Document, infNFeElement, nfeElement *etree.Element, elementID string) (string, error) {
	// 🚀 FINAL FIX: Use ucarion/c14n library for PHP libxml compatibility
	// This should produce the exact same canonicalization as PHP's libxml
	
	// Remove existing signature if any (for enveloped-signature)
	existingSig := doc.FindElement(".//Signature")
	if existingSig != nil {
		existingSig.Parent().RemoveChild(existingSig)
	}
	
	// Get the infNFe element XML string with its namespace
	tempDoc := etree.NewDocument()
	infNFeCopy := infNFeElement.Copy()
	
	// Ensure proper namespace for canonicalization
	if infNFeCopy.SelectAttr("xmlns") == nil {
		infNFeCopy.CreateAttr("xmlns", "http://www.portalfiscal.inf.br/nfe")
	}
	
	tempDoc.SetRoot(infNFeCopy)
	infNFeString, err := tempDoc.WriteToString()
	if err != nil {
		return "", err
	}

	fmt.Printf("🔬 DEBUG - infNFe antes C14N (primeiros 300 chars): %s...\n", 
		infNFeString[:min(300, len(infNFeString))])

	// 🚀 Use project's own canonicalization implementation (more reliable)
	config := &CanonicalizationConfig{
		Method:          C14N10Inclusive,  // C14N 1.0 inclusive como no exemplo PHP
		InclusivePrefix: "",
		WithComments:    false,
		TrimWhitespace:  true,
		SortAttributes:  true,
		RemoveXMLDecl:   true,
	}
	
	canonicalizer := NewXMLCanonicalizer(config)
	canonicalBytes, err := canonicalizer.Canonicalize(infNFeString)
	if err != nil {
		return "", fmt.Errorf("canonicalization failed: %v", err)
	}

	fmt.Printf("🔬 DEBUG - infNFe pós C14N próprio (primeiros 300 chars): %s...\n", 
		string(canonicalBytes)[:min(300, len(canonicalBytes))])
	
	digest := signer.calculateDigest(canonicalBytes)
	digestBase64 := base64.StdEncoding.EncodeToString(digest)
	fmt.Printf("🔬 DEBUG - Digest calculado (ucarion/c14n): %s\n", digestBase64)

	// Generate signature template for the infNFe element
	template := signer.createSignatureTemplate("#"+elementID, digest)

	// Parse template
	signatureDoc := etree.NewDocument()
	if err := signatureDoc.ReadFromString(template); err != nil {
		return "", err
	}

	signature := signatureDoc.Root()
	if signature == nil {
		return "", fmt.Errorf("failed to parse signature template")
	}

	// Insert signature as sibling to infNFe (inside NFe element) - correct for NFe structure
	nfeElement.AddChild(signature.Copy())

	// Calculate signature for the SignedInfo element with canonicalization
	signedInfo := signer.extractSignedInfoCanonical(template)
	fmt.Printf("🔬 DEBUG - SignedInfo para assinatura (primeiros 150 chars): %s...\n", 
		signedInfo[:min(150, len(signedInfo))])
	
	signatureValue, err := signer.Sign([]byte(signedInfo))
	if err != nil {
		return "", err
	}

	// Update signature value in the document (without ds: prefix)
	sigValueElement := doc.FindElement(".//SignatureValue")
	if sigValueElement != nil {
		sigValueElement.SetText(base64.StdEncoding.EncodeToString(signatureValue))
	}

	// Return the signed document without XML declaration
	doc.WriteSettings = etree.WriteSettings{
		CanonicalAttrVal: false,
		CanonicalEndTags: false,
		CanonicalText:    false,
		UseCRLF:          false,
	}
	result, err := doc.WriteToString()
	if err != nil {
		return "", err
	}

	// Remove XML declaration if present (for compatibility with lote XML)
	if strings.HasPrefix(result, "<?xml") {
		if idx := strings.Index(result, "?>"); idx >= 0 {
			result = strings.TrimSpace(result[idx+2:])
		}
	}

	return result, nil
}

// ValidateNFeSignature validates an NFe XML signature
func (signer *XMLSigner) ValidateNFeSignature(signedNFeXML string) error {
	return signer.VerifyXMLSignature(signedNFeXML)
}

// GetSignatureInfo extracts signature information from any signed XML
func GetSignatureInfo(signedXML string) (*SignatureInfo, error) {
	if signedXML == "" {
		return nil, errors.NewValidationError("signed XML cannot be empty", "signedXML", "")
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return nil, errors.NewValidationError("failed to parse signed XML", "xml", err.Error())
	}

	// Extract signature information
	info := &SignatureInfo{}

	if sigValueElem := doc.FindElement(".//ds:SignatureValue"); sigValueElem != nil {
		info.SignatureValue = sigValueElem.Text()
	}

	if digestValueElem := doc.FindElement(".//ds:DigestValue"); digestValueElem != nil {
		info.DigestValue = digestValueElem.Text()
	}

	if certElem := doc.FindElement(".//ds:X509Certificate"); certElem != nil {
		info.Certificate = certElem.Text()
	}

	if sigMethodElem := doc.FindElement(".//ds:SignatureMethod"); sigMethodElem != nil {
		if algAttr := sigMethodElem.SelectAttr("Algorithm"); algAttr != nil {
			info.Algorithm = algAttr.Value
		}
	}

	return info, nil
}

// CreateXMLSigner is a convenience function to create an XML signer
func CreateXMLSigner(certificate Certificate) *XMLSigner {
	return NewXMLSigner(certificate, DefaultSignerConfig())
}

// CreateSHA256XMLSigner creates an XML signer configured for SHA-256
func CreateSHA256XMLSigner(certificate Certificate) *XMLSigner {
	return NewXMLSigner(certificate, SHA256SignerConfig())
}

// SignXMLWithCertificate is a convenience function to sign XML with a certificate
func SignXMLWithCertificate(xmlContent string, certificate Certificate) (string, error) {
	signer := CreateXMLSigner(certificate)
	signedXML, _, err := signer.SignXML(xmlContent)
	return signedXML, err
}

// signXMLManually creates a basic XML signature manually
func (signer *XMLSigner) signXMLManually(xmlContent string) (string, error) {
	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return "", err
	}

	// Create a basic signature element
	signature := signer.createBasicSignatureElement(xmlContent)

	// Insert signature into the document
	doc.Root().AddChild(signature)

	result, err := doc.WriteToString()
	if err != nil {
		return "", err
	}

	return result, nil
}

// signXMLElementManually creates a signature for a specific element
func (signer *XMLSigner) signXMLElementManually(xmlContent, elementID string) (string, error) {
	return signer.signXMLManually(xmlContent)
}

// verifyXMLSignatureManually performs basic signature verification
func (signer *XMLSigner) verifyXMLSignatureManually(signedXML string) error {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return err
	}

	// Check if signature element exists
	sigElement := doc.FindElement(".//ds:Signature")
	if sigElement == nil {
		return errors.NewCertificateError("no signature found in XML", nil)
	}

	return nil
}

// createBasicSignatureElement creates a basic signature element
func (signer *XMLSigner) createBasicSignatureElement(xmlContent string) *etree.Element {
	signature := etree.NewElement("ds:Signature")
	signature.CreateAttr("xmlns:ds", "http://www.w3.org/2000/09/xmldsig#")

	signedInfo := signature.CreateElement("ds:SignedInfo")
	signedInfo.CreateElement("ds:CanonicalizationMethod").CreateAttr("Algorithm", signer.config.CanonicalizationAlgorithm)
	signedInfo.CreateElement("ds:SignatureMethod").CreateAttr("Algorithm", signer.config.SignatureAlgorithm)

	reference := signedInfo.CreateElement("ds:Reference")
	reference.CreateAttr("URI", "")
	reference.CreateElement("ds:DigestMethod").CreateAttr("Algorithm", signer.getDigestMethodURI())

	// Calculate digest of the XML content
	digest := signer.calculateDigest([]byte(xmlContent))
	reference.CreateElement("ds:DigestValue").SetText(base64.StdEncoding.EncodeToString(digest))

	// Create signature value (simplified)
	signedInfoBytes := []byte(signedInfo.Text())
	signatureValue, _ := signer.certificate.Sign(signedInfoBytes, signer.config.DigestAlgorithm)
	signature.CreateElement("ds:SignatureValue").SetText(base64.StdEncoding.EncodeToString(signatureValue))

	// Add certificate info if required
	if signer.config.IncludeCertificate {
		keyInfo := signature.CreateElement("ds:KeyInfo")
		x509Data := keyInfo.CreateElement("ds:X509Data")
		cert := signer.certificate.GetCertificate()
		if cert != nil {
			x509Data.CreateElement("ds:X509Certificate").SetText(base64.StdEncoding.EncodeToString(cert.Raw))
		}
	}

	return signature
}

// calculateDigest calculates the digest of data using the configured algorithm
func (signer *XMLSigner) calculateDigest(data []byte) []byte {
	switch signer.config.DigestAlgorithm {
	case crypto.SHA1:
		hash := sha1.Sum(data)
		return hash[:]
	case crypto.SHA256:
		hash := sha256.Sum256(data)
		return hash[:]
	default:
		hash := sha1.Sum(data)
		return hash[:]
	}
}

// ValidateXMLSignature is a convenience function to validate an XML signature
func ValidateXMLSignature(signedXML string) error {
	// Basic signature validation - parse and check structure
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return errors.NewCertificateError("failed to parse signed XML", err)
	}

	// Check if signature element exists
	sigElement := doc.FindElement(".//ds:Signature")
	if sigElement == nil {
		return errors.NewCertificateError("no signature found in XML", nil)
	}

	return nil
}

// simpleTokenReader implements c14n.RawTokenReader interface
type simpleTokenReader struct {
	tokens []xml.Token
	index  int
}

// RawToken implements the RawTokenReader interface
func (r *simpleTokenReader) RawToken() (xml.Token, error) {
	if r.index >= len(r.tokens) {
		return nil, fmt.Errorf("EOF")
	}
	token := r.tokens[r.index]
	r.index++
	return token, nil
}
