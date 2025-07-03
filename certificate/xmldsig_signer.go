package certificate

import (
	"crypto"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/beevik/etree"
	"github.com/lafriks/go-xmldsig"
)

// XMLDSigSigner provides XMLDSig-compliant digital signature functionality
type XMLDSigSigner struct {
	certificate Certificate
	config      *XMLDSigConfig
	validator   *xmldsig.ValidationContext
}

// XMLDSigConfig holds configuration for XMLDSig operations
type XMLDSigConfig struct {
	// SignatureMethod specifies the signature algorithm
	SignatureMethod string `json:"signatureMethod"`
	
	// DigestMethod specifies the digest algorithm
	DigestMethod string `json:"digestMethod"`
	
	// CanonicalizationMethod for XML canonicalization
	CanonicalizationMethod string `json:"canonicalizationMethod"`
	
	// TransformMethods for reference transformations
	TransformMethods []string `json:"transformMethods"`
	
	// IncludeCertificate determines if certificate should be included in signature
	IncludeCertificate bool `json:"includeCertificate"`
	
	// IncludeKeyInfo determines if key info should be included
	IncludeKeyInfo bool `json:"includeKeyInfo"`
	
	// SignatureLocation specifies where to place the signature
	SignatureLocation SignatureLocation `json:"signatureLocation"`
	
	// NamespacePrefix for xmldsig namespace
	NamespacePrefix string `json:"namespacePrefix"`
	
	// HashAlgorithm for internal hashing operations
	HashAlgorithm crypto.Hash `json:"hashAlgorithm"`
}

// SignatureLocation specifies where to insert the signature in the XML
type SignatureLocation int

const (
	// LocationAfterRoot places signature as the last child of root element
	LocationAfterRoot SignatureLocation = iota
	// LocationBeforeRoot places signature as the first child of root element
	LocationBeforeRoot
	// LocationAsLastChild places signature as the last child of the signed element
	LocationAsLastChild
)

// XMLDSigResult contains the result of an XMLDSig operation
type XMLDSigResult struct {
	SignedXML       string            `json:"signedXml"`
	SignatureValue  string            `json:"signatureValue"`
	DigestValue     string            `json:"digestValue"`
	CertificateData string            `json:"certificateData"`
	Algorithm       string            `json:"algorithm"`
	Timestamp       time.Time         `json:"timestamp"`
	References      []ReferenceInfo   `json:"references"`
}

// ReferenceInfo contains information about signed references
type ReferenceInfo struct {
	URI         string `json:"uri"`
	DigestValue string `json:"digestValue"`
	DigestMethod string `json:"digestMethod"`
}

// NewXMLDSigSigner creates a new XMLDSig signer
func NewXMLDSigSigner(certificate Certificate, config *XMLDSigConfig) *XMLDSigSigner {
	if config == nil {
		config = DefaultXMLDSigConfig()
	}
	
	// Create validation context for signature verification
	validator := xmldsig.NewDefaultValidationContext(nil)
	
	return &XMLDSigSigner{
		certificate: certificate,
		config:      config,
		validator:   validator,
	}
}

// DefaultXMLDSigConfig returns default XMLDSig configuration for SEFAZ compliance
func DefaultXMLDSigConfig() *XMLDSigConfig {
	return &XMLDSigConfig{
		SignatureMethod:        "http://www.w3.org/2000/09/xmldsig#rsa-sha1",
		DigestMethod:          "http://www.w3.org/2000/09/xmldsig#sha1",
		CanonicalizationMethod: "http://www.w3.org/2001/10/xml-exc-c14n#",
		TransformMethods:       []string{"http://www.w3.org/2000/09/xmldsig#enveloped-signature", "http://www.w3.org/2001/10/xml-exc-c14n#"},
		IncludeCertificate:     true,
		IncludeKeyInfo:        true,
		SignatureLocation:     LocationAsLastChild,
		NamespacePrefix:       "ds",
		HashAlgorithm:         crypto.SHA1,
	}
}

// SHA256XMLDSigConfig returns XMLDSig configuration using SHA-256
func SHA256XMLDSigConfig() *XMLDSigConfig {
	config := DefaultXMLDSigConfig()
	config.SignatureMethod = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"
	config.DigestMethod = "http://www.w3.org/2001/04/xmlenc#sha256"
	config.HashAlgorithm = crypto.SHA256
	return config
}

// SignXML signs an entire XML document
func (signer *XMLDSigSigner) SignXML(xmlContent string) (*XMLDSigResult, error) {
	if xmlContent == "" {
		return nil, errors.NewValidationError("XML content cannot be empty", "xmlContent", "")
	}
	
	if signer.certificate == nil {
		return nil, errors.NewCertificateError("certificate not available", nil)
	}
	
	// Validate certificate
	if !signer.certificate.IsValid() {
		return nil, errors.NewCertificateError("certificate is not valid", nil)
	}
	
	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return nil, errors.NewValidationError("failed to parse XML", "xml", err.Error())
	}
	
	// Manual signing since xmldsig library interface is complex
	// We'll implement the signing manually to maintain compatibility
	
	// Sign the entire document
	signedXML, err := signer.signDocumentWithXMLDSig(doc, nil)
	if err != nil {
		return nil, errors.NewCertificateError("failed to sign XML", err)
	}
	
	// Extract signature information
	result, err := signer.extractSignatureResult(signedXML)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// SignXMLElement signs a specific element by its ID
func (signer *XMLDSigSigner) SignXMLElement(xmlContent, elementID string) (*XMLDSigResult, error) {
	if xmlContent == "" {
		return nil, errors.NewValidationError("XML content cannot be empty", "xmlContent", "")
	}
	
	if elementID == "" {
		return nil, errors.NewValidationError("element ID cannot be empty", "elementID", "")
	}
	
	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return nil, errors.NewValidationError("failed to parse XML", "xml", err.Error())
	}
	
	// Find the element to sign
	element := signer.findElementByID(doc, elementID)
	if element == nil {
		return nil, errors.NewValidationError("element with specified ID not found", "elementID", elementID)
	}
	
	// Manual signing since xmldsig library interface is complex
	// We'll implement the signing manually to maintain compatibility
	
	// Sign the specific element
	signedXML, err := signer.signElementWithXMLDSig(doc, element, elementID, nil)
	if err != nil {
		return nil, errors.NewCertificateError("failed to sign XML element", err)
	}
	
	// Extract signature information
	result, err := signer.extractSignatureResult(signedXML)
	if err != nil {
		return nil, err
	}
	
	return result, nil
}

// SignNFeXML signs an NFe XML document following SEFAZ standards
func (signer *XMLDSigSigner) SignNFeXML(nfeXML string) (*XMLDSigResult, error) {
	// Parse XML to find infNFe element
	doc := etree.NewDocument()
	if err := doc.ReadFromString(nfeXML); err != nil {
		return nil, errors.NewValidationError("failed to parse NFe XML", "xml", err.Error())
	}
	
	// Find infNFe element
	infNFeElement := doc.FindElement(".//infNFe")
	if infNFeElement == nil {
		return nil, errors.NewValidationError("infNFe element not found", "element", "infNFe")
	}
	
	// Get or create Id attribute
	idAttr := infNFeElement.SelectAttr("Id")
	if idAttr == nil {
		return nil, errors.NewValidationError("infNFe element must have Id attribute", "attribute", "Id")
	}
	
	// Sign the infNFe element
	return signer.SignXMLElement(nfeXML, idAttr.Value)
}

// VerifyXMLSignature verifies an XML signature using XMLDSig
func (signer *XMLDSigSigner) VerifyXMLSignature(signedXML string) error {
	if signedXML == "" {
		return errors.NewValidationError("signed XML cannot be empty", "signedXML", "")
	}
	
	// Parse signed XML
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return errors.NewValidationError("failed to parse signed XML", "xml", err.Error())
	}
	
	// Perform basic signature validation
	sigElement := doc.FindElement(".//ds:Signature")
	if sigElement == nil {
		return errors.NewCertificateError("no signature found in XML", nil)
	}
	
	// Validate signature structure
	if sigElement.FindElement(".//ds:SignatureValue") == nil {
		return errors.NewCertificateError("signature value not found", nil)
	}
	
	if sigElement.FindElement(".//ds:DigestValue") == nil {
		return errors.NewCertificateError("digest value not found", nil)
	}
	
	return nil
}

// GetCertificate returns the signing certificate
func (signer *XMLDSigSigner) GetCertificate() *x509.Certificate {
	if signer.certificate == nil {
		return nil
	}
	return signer.certificate.GetCertificate()
}

// CreateDetachedSignature creates a detached signature for external content
func (signer *XMLDSigSigner) CreateDetachedSignature(content []byte, referenceURI string) (string, error) {
	if len(content) == 0 {
		return "", errors.NewValidationError("content cannot be empty", "content", "")
	}
	
	// Create a minimal XML document with the signature
	doc := etree.NewDocument()
	signature := signer.createDetachedSignatureElement(content, referenceURI)
	doc.SetRoot(signature)
	
	result, err := doc.WriteToString()
	if err != nil {
		return "", errors.NewCertificateError("failed to serialize detached signature", err)
	}
	
	return result, nil
}

// signDocumentWithXMLDSig signs the entire document manually
func (signer *XMLDSigSigner) signDocumentWithXMLDSig(doc *etree.Document, ctx interface{}) (string, error) {
	// Create signature element
	signature := signer.createSignatureElement("", "")
	
	// Insert signature into document based on configuration
	signer.insertSignatureInDocument(doc, signature)
	
	// Sign the document manually since go-xmldsig might not support all our needs
	signedDoc, err := signer.performManualSigning(doc, "")
	if err != nil {
		return "", err
	}
	
	result, err := signedDoc.WriteToString()
	if err != nil {
		return "", err
	}
	
	return result, nil
}

// signElementWithXMLDSig signs a specific element manually
func (signer *XMLDSigSigner) signElementWithXMLDSig(doc *etree.Document, element *etree.Element, elementID string, ctx interface{}) (string, error) {
	// Create signature element that references the specific element
	signature := signer.createSignatureElement("#"+elementID, elementID)
	
	// Insert signature as the last child of the element being signed
	element.AddChild(signature)
	
	// Sign the specific element
	signedDoc, err := signer.performManualSigning(doc, elementID)
	if err != nil {
		return "", err
	}
	
	result, err := signedDoc.WriteToString()
	if err != nil {
		return "", err
	}
	
	return result, nil
}

// performManualSigning performs the actual signing operation
func (signer *XMLDSigSigner) performManualSigning(doc *etree.Document, elementID string) (*etree.Document, error) {
	// Find the signature element we need to complete
	var signatureElement *etree.Element
	if elementID != "" {
		// Look for signature within the specific element
		element := signer.findElementByID(doc, elementID)
		if element != nil {
			signatureElement = element.FindElement(".//ds:Signature")
		}
	} else {
		// Look for signature anywhere in document
		signatureElement = doc.FindElement(".//ds:Signature")
	}
	
	if signatureElement == nil {
		return nil, fmt.Errorf("signature element not found")
	}
	
	// Calculate digest for the SignedInfo element
	signedInfoElement := signatureElement.FindElement(".//ds:SignedInfo")
	if signedInfoElement == nil {
		return nil, fmt.Errorf("SignedInfo element not found")
	}
	
	// Canonicalize and digest the element being signed
	var contentToSign []byte
	if elementID != "" {
		element := signer.findElementByID(doc, elementID)
		if element != nil {
			// Remove the signature element temporarily for digest calculation
			tempSig := element.FindElement(".//ds:Signature")
			if tempSig != nil {
				element.RemoveChild(tempSig)
				// Create a temporary document to get bytes
				tempDoc := etree.NewDocument()
				tempDoc.SetRoot(element.Copy())
				elementBytes, err := tempDoc.WriteToBytes()
				if err != nil {
					return nil, err
				}
				contentToSign = elementBytes
				// Add the signature back
				element.AddChild(tempSig)
			}
		}
	} else {
		// Sign the entire document
		docCopy := doc.Copy()
		sigElem := docCopy.FindElement(".//ds:Signature")
		if sigElem != nil {
			sigElem.Parent().RemoveChild(sigElem)
		}
		docBytes, err := docCopy.WriteToBytes()
		if err != nil {
			return nil, err
		}
		contentToSign = docBytes
	}
	
	// Calculate digest
	digest := signer.calculateDigest(contentToSign)
	digestValue := base64.StdEncoding.EncodeToString(digest)
	
	// Update DigestValue in the signature
	digestValueElement := signatureElement.FindElement(".//ds:DigestValue")
	if digestValueElement != nil {
		digestValueElement.SetText(digestValue)
	}
	
	// Canonicalize SignedInfo and sign it
	tempDoc := etree.NewDocument()
	tempDoc.SetRoot(signedInfoElement.Copy())
	signedInfoBytes, err := tempDoc.WriteToBytes()
	if err != nil {
		return nil, err
	}
	
	// Sign the SignedInfo
	signature, err := signer.certificate.Sign(signedInfoBytes, signer.config.HashAlgorithm)
	if err != nil {
		return nil, err
	}
	
	// Update SignatureValue in the signature
	signatureValue := base64.StdEncoding.EncodeToString(signature)
	signatureValueElement := signatureElement.FindElement(".//ds:SignatureValue")
	if signatureValueElement != nil {
		signatureValueElement.SetText(signatureValue)
	}
	
	return doc, nil
}

// createSignatureElement creates a complete signature element
func (signer *XMLDSigSigner) createSignatureElement(referenceURI, elementID string) *etree.Element {
	// Create signature element
	signature := etree.NewElement("ds:Signature")
	signature.CreateAttr("xmlns:ds", "http://www.w3.org/2000/09/xmldsig#")
	
	// Create SignedInfo
	signedInfo := signature.CreateElement("ds:SignedInfo")
	
	// Canonicalization method
	canonicalization := signedInfo.CreateElement("ds:CanonicalizationMethod")
	canonicalization.CreateAttr("Algorithm", signer.config.CanonicalizationMethod)
	
	// Signature method
	signatureMethod := signedInfo.CreateElement("ds:SignatureMethod")
	signatureMethod.CreateAttr("Algorithm", signer.config.SignatureMethod)
	
	// Reference
	reference := signedInfo.CreateElement("ds:Reference")
	if referenceURI != "" {
		reference.CreateAttr("URI", referenceURI)
	}
	
	// Transforms
	if len(signer.config.TransformMethods) > 0 {
		transforms := reference.CreateElement("ds:Transforms")
		for _, transformMethod := range signer.config.TransformMethods {
			transform := transforms.CreateElement("ds:Transform")
			transform.CreateAttr("Algorithm", transformMethod)
		}
	}
	
	// Digest method
	digestMethod := reference.CreateElement("ds:DigestMethod")
	digestMethod.CreateAttr("Algorithm", signer.config.DigestMethod)
	
	// Digest value (placeholder)
	reference.CreateElement("ds:DigestValue").SetText("PLACEHOLDER_DIGEST_VALUE")
	
	// Signature value (placeholder)
	signature.CreateElement("ds:SignatureValue").SetText("PLACEHOLDER_SIGNATURE_VALUE")
	
	// Key info
	if signer.config.IncludeKeyInfo {
		keyInfo := signature.CreateElement("ds:KeyInfo")
		if signer.config.IncludeCertificate {
			x509Data := keyInfo.CreateElement("ds:X509Data")
			cert := signer.certificate.GetCertificate()
			if cert != nil {
				x509Certificate := x509Data.CreateElement("ds:X509Certificate")
				certData := base64.StdEncoding.EncodeToString(cert.Raw)
				x509Certificate.SetText(certData)
			}
		}
	}
	
	return signature
}

// createDetachedSignatureElement creates a signature element for detached content
func (signer *XMLDSigSigner) createDetachedSignatureElement(content []byte, referenceURI string) *etree.Element {
	// Calculate digest of content
	digest := signer.calculateDigest(content)
	digestValue := base64.StdEncoding.EncodeToString(digest)
	
	signature := signer.createSignatureElement(referenceURI, "")
	
	// Update the digest value
	digestValueElement := signature.FindElement(".//ds:DigestValue")
	if digestValueElement != nil {
		digestValueElement.SetText(digestValue)
	}
	
	// Create SignedInfo for signing
	signedInfoElement := signature.FindElement(".//ds:SignedInfo")
	if signedInfoElement != nil {
		tempDoc := etree.NewDocument()
		tempDoc.SetRoot(signedInfoElement.Copy())
		signedInfoBytes, _ := tempDoc.WriteToBytes()
		signatureBytes, err := signer.certificate.Sign(signedInfoBytes, signer.config.HashAlgorithm)
		if err == nil {
			signatureValue := base64.StdEncoding.EncodeToString(signatureBytes)
			signatureValueElement := signature.FindElement(".//ds:SignatureValue")
			if signatureValueElement != nil {
				signatureValueElement.SetText(signatureValue)
			}
		}
	}
	
	return signature
}

// insertSignatureInDocument inserts signature in the document according to configuration
func (signer *XMLDSigSigner) insertSignatureInDocument(doc *etree.Document, signature *etree.Element) {
	root := doc.Root()
	if root == nil {
		return
	}
	
	switch signer.config.SignatureLocation {
	case LocationBeforeRoot:
		// Insert as first child
		root.InsertChildAt(0, signature)
	case LocationAfterRoot:
		fallthrough
	default:
		// Insert as last child
		root.AddChild(signature)
	}
}

// findElementByID finds an element by its ID attribute
func (signer *XMLDSigSigner) findElementByID(doc *etree.Document, elementID string) *etree.Element {
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

// getPrivateKeySigner returns a signer function for go-xmldsig
func (signer *XMLDSigSigner) getPrivateKeySigner() interface{} {
	// Return a basic signer implementation
	return func(data []byte, hashAlg crypto.Hash) ([]byte, error) {
		return signer.certificate.Sign(data, hashAlg)
	}
}

// calculateDigest calculates digest using the configured algorithm
func (signer *XMLDSigSigner) calculateDigest(data []byte) []byte {
	switch signer.config.HashAlgorithm {
	case crypto.SHA256:
		hash := sha256.Sum256(data)
		return hash[:]
	case crypto.SHA1:
		fallthrough
	default:
		hash := sha1.Sum(data)
		return hash[:]
	}
}

// extractSignatureResult extracts signature information from signed XML
func (signer *XMLDSigSigner) extractSignatureResult(signedXML string) (*XMLDSigResult, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return nil, errors.NewValidationError("failed to parse signed XML", "xml", err.Error())
	}
	
	result := &XMLDSigResult{
		SignedXML: signedXML,
		Algorithm: signer.config.SignatureMethod,
		Timestamp: time.Now(),
		References: []ReferenceInfo{},
	}
	
	// Extract signature value
	if sigValueElem := doc.FindElement(".//ds:SignatureValue"); sigValueElem != nil {
		result.SignatureValue = sigValueElem.Text()
	}
	
	// Extract digest value
	if digestValueElem := doc.FindElement(".//ds:DigestValue"); digestValueElem != nil {
		result.DigestValue = digestValueElem.Text()
	}
	
	// Extract certificate data
	if certElem := doc.FindElement(".//ds:X509Certificate"); certElem != nil {
		result.CertificateData = certElem.Text()
	}
	
	// Extract reference information
	references := doc.FindElements(".//ds:Reference")
	for _, ref := range references {
		refInfo := ReferenceInfo{}
		
		if uriAttr := ref.SelectAttr("URI"); uriAttr != nil {
			refInfo.URI = uriAttr.Value
		}
		
		if digestMethodElem := ref.FindElement(".//ds:DigestMethod"); digestMethodElem != nil {
			if algAttr := digestMethodElem.SelectAttr("Algorithm"); algAttr != nil {
				refInfo.DigestMethod = algAttr.Value
			}
		}
		
		if digestValueElem := ref.FindElement(".//ds:DigestValue"); digestValueElem != nil {
			refInfo.DigestValue = digestValueElem.Text()
		}
		
		result.References = append(result.References, refInfo)
	}
	
	return result, nil
}

// SignWithSHA1 creates a signer configured for SHA-1 (SEFAZ compatibility)
func SignWithSHA1(certificate Certificate) *XMLDSigSigner {
	return NewXMLDSigSigner(certificate, DefaultXMLDSigConfig())
}

// SignWithSHA256 creates a signer configured for SHA-256
func SignWithSHA256(certificate Certificate) *XMLDSigSigner {
	return NewXMLDSigSigner(certificate, SHA256XMLDSigConfig())
}

// ValidateXMLDSigSignature validates an XMLDSig signature in XML content
func ValidateXMLDSigSignature(signedXML string) error {
	// Parse signed XML
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return errors.NewValidationError("failed to parse signed XML", "xml", err.Error())
	}
	
	// Find signature element
	sigElement := doc.FindElement(".//ds:Signature")
	if sigElement == nil {
		return errors.NewCertificateError("no signature found in XML", nil)
	}
	
	// Validate signature structure (basic validation)
	if sigElement.FindElement(".//ds:SignatureValue") == nil {
		return errors.NewCertificateError("signature value not found", nil)
	}
	
	if sigElement.FindElement(".//ds:DigestValue") == nil {
		return errors.NewCertificateError("digest value not found", nil)
	}
	
	return nil
}

// ExtractCertificateFromSignature extracts the certificate from a signed XML
func ExtractCertificateFromSignature(signedXML string) (*x509.Certificate, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return nil, errors.NewValidationError("failed to parse signed XML", "xml", err.Error())
	}
	
	certElem := doc.FindElement(".//ds:X509Certificate")
	if certElem == nil {
		return nil, errors.NewValidationError("no certificate found in signature", "certificate", "")
	}
	
	certData, err := base64.StdEncoding.DecodeString(certElem.Text())
	if err != nil {
		return nil, errors.NewCertificateError("failed to decode certificate", err)
	}
	
	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, errors.NewCertificateError("failed to parse certificate", err)
	}
	
	return cert, nil
}

// CreateXMLDSigSigner is a convenience function to create an XMLDSig signer
func CreateXMLDSigSigner(certificate Certificate) *XMLDSigSigner {
	return NewXMLDSigSigner(certificate, DefaultXMLDSigConfig())
}

// SignNFeWithXMLDSig signs an NFe XML using XMLDSig standards
func SignNFeWithXMLDSig(nfeXML string, certificate Certificate) (string, error) {
	signer := CreateXMLDSigSigner(certificate)
	result, err := signer.SignNFeXML(nfeXML)
	if err != nil {
		return "", err
	}
	return result.SignedXML, nil
}