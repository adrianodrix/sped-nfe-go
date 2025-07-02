// Package certificate provides XML digital signature functionality for NFe documents.
// It implements xmldsig standard compatible with SEFAZ requirements.
package certificate

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/beevik/etree"
	"github.com/lafriks/go-xmldsig"
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
		SignatureElementID:        "xmldsig",
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

	// Create xmldsig context
	ctx := xmldsig.NewDefaultSigningContext(signer)
	ctx.SetSignatureMethod(signer.config.SignatureAlgorithm)
	ctx.SetDigestMethod(signer.getDigestMethodURI())
	ctx.SetCanonicalizationMethod(signer.config.CanonicalizationAlgorithm)

	// Sign the document
	signedXML, err := ctx.SignEnveloped(xmlContent)
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

	// Create xmldsig context
	ctx := xmldsig.NewDefaultSigningContext(signer)
	ctx.SetSignatureMethod(signer.config.SignatureAlgorithm)
	ctx.SetDigestMethod(signer.getDigestMethodURI())
	ctx.SetCanonicalizationMethod(signer.config.CanonicalizationAlgorithm)

	// Sign the specific element
	signedXML, err := ctx.SignEnveloped(xmlContent)
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

	// Create validation context
	ctx := xmldsig.NewDefaultValidationContext(nil)
	
	// Verify the signature
	_, err := ctx.Validate(signedXML)
	if err != nil {
		return errors.NewCertificateError("XML signature verification failed", err)
	}

	return nil
}

// GetCertificate returns the certificate used for signing (implements xmldsig.Signer interface)
func (signer *XMLSigner) GetCertificate() *xmldsig.Certificate {
	cert := signer.certificate.GetCertificate()
	if cert == nil {
		return nil
	}

	return &xmldsig.Certificate{
		Certificate: cert,
	}
}

// Sign signs data using the certificate (implements xmldsig.Signer interface)
func (signer *XMLSigner) Sign(data []byte) ([]byte, error) {
	return signer.certificate.Sign(data, signer.config.DigestAlgorithm)
}

// SignatureAlgorithm returns the signature algorithm URI (implements xmldsig.Signer interface)
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
	existingSignatures := doc.FindElements(".//ds:Signature")
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

	// Find signature elements
	sigValueElem := doc.FindElement(".//ds:SignatureValue")
	digestValueElem := doc.FindElement(".//ds:DigestValue")
	
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
		certElem := doc.FindElement(".//ds:X509Certificate")
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

	template := fmt.Sprintf(`<ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#" Id="%s">
  <ds:SignedInfo>
    <ds:CanonicalizationMethod Algorithm="%s"/>
    <ds:SignatureMethod Algorithm="%s"/>
    <ds:Reference URI="%s">
      <ds:DigestMethod Algorithm="%s"/>
      <ds:DigestValue>%s</ds:DigestValue>
    </ds:Reference>
  </ds:SignedInfo>
  <ds:SignatureValue>{{SIGNATURE_VALUE}}</ds:SignatureValue>`,
		signer.config.SignatureElementID,
		signer.config.CanonicalizationAlgorithm,
		signer.config.SignatureAlgorithm,
		referenceURI,
		digestMethod,
		digestValue)

	if signer.config.IncludeCertificate {
		cert := signer.certificate.GetCertificate()
		if cert != nil {
			certData := base64.StdEncoding.EncodeToString(cert.Raw)
			template += fmt.Sprintf(`
  <ds:KeyInfo>
    <ds:X509Data>
      <ds:X509Certificate>%s</ds:X509Certificate>
    </ds:X509Data>
  </ds:KeyInfo>`, certData)
		}
	}

	template += `
</ds:Signature>`

	return template
}

// extractSignedInfo extracts the SignedInfo element for signing
func (signer *XMLSigner) extractSignedInfo(template string) string {
	doc := etree.NewDocument()
	doc.ReadFromString(template)
	
	signedInfo := doc.FindElement(".//ds:SignedInfo")
	if signedInfo == nil {
		return ""
	}

	doc2 := etree.NewDocument()
	doc2.SetRoot(signedInfo.Copy())
	result, _ := doc2.WriteToString()
	
	return result
}

// SignNFeXML signs an NFe XML document following SEFAZ requirements
func (signer *XMLSigner) SignNFeXML(nfeXML string) (string, error) {
	// NFe specific signing - signs the infNFe element
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

	// Sign the XML
	signedXML, _, err := signer.SignXMLElement(nfeXML, idAttr.Value)
	if err != nil {
		return "", err
	}

	return signedXML, nil
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

// ValidateXMLSignature is a convenience function to validate an XML signature
func ValidateXMLSignature(signedXML string) error {
	ctx := xmldsig.NewDefaultValidationContext(nil)
	_, err := ctx.Validate(signedXML)
	if err != nil {
		return errors.NewCertificateError("XML signature validation failed", err)
	}
	return nil
}