// Package soap provides SOAP envelope construction and parsing functionality
// with support for SOAP 1.1/1.2, namespaces, and SEFAZ-specific requirements.
package soap

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
)

// SOAPVersion represents the SOAP protocol version
type SOAPVersion string

const (
	// SOAP11 represents SOAP version 1.1
	SOAP11 SOAPVersion = "1.1"
	// SOAP12 represents SOAP version 1.2
	SOAP12 SOAPVersion = "1.2"
)

// Namespace constants for SOAP envelopes
const (
	// SOAP 1.1 namespaces
	SOAP11EnvelopeNS = "http://schemas.xmlsoap.org/soap/envelope/"
	SOAP11EncodingNS = "http://schemas.xmlsoap.org/soap/encoding/"

	// SOAP 1.2 namespaces
	SOAP12EnvelopeNS = "http://www.w3.org/2003/05/soap-envelope"
	SOAP12EncodingNS = "http://www.w3.org/2003/05/soap-encoding"

	// Common namespaces for NFe
	XMLSchemaInstanceNS   = "http://www.w3.org/2001/XMLSchema-instance"
	XMLSchemaNS           = "http://www.w3.org/2001/XMLSchema"
	WSSecurityNS          = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd"
	WSUtilityNS           = "http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd"
	XMLDigitalSignatureNS = "http://www.w3.org/2000/09/xmldsig#"
)

// SOAPEnvelope represents a complete SOAP envelope
type SOAPEnvelope struct {
	XMLName   xml.Name    `xml:"Envelope"`
	XmlnsSoap string      `xml:"xmlns:soap,attr"`
	XmlnsXsi  string      `xml:"xmlns:xsi,attr,omitempty"`
	XmlnsXsd  string      `xml:"xmlns:xsd,attr,omitempty"`
	Header    *SOAPHeader `xml:"Header,omitempty"`
	Body      *SOAPBody   `xml:"Body"`
}

// SOAPHeader represents the SOAP header section
type SOAPHeader struct {
	XMLName  xml.Name              `xml:"Header"`
	Security *SecurityHeader       `xml:"Security,omitempty"`
	Custom   []CustomHeaderElement `xml:",omitempty"`
}

// SOAPBody represents the SOAP body section
type SOAPBody struct {
	XMLName xml.Name   `xml:"Body"`
	Content string     `xml:",innerxml"`
	Fault   *SOAPFault `xml:"Fault,omitempty"`
}

// SOAPFault represents a SOAP fault
type SOAPFault struct {
	XMLName xml.Name `xml:"Fault"`
	Code    string   `xml:"faultcode"`
	String  string   `xml:"faultstring"`
	Actor   string   `xml:"faultactor,omitempty"`
	Detail  string   `xml:"detail,omitempty"`
}

// SecurityHeader represents WS-Security header
type SecurityHeader struct {
	XMLName   xml.Name   `xml:"wsse:Security"`
	XmlnsWsse string     `xml:"xmlns:wsse,attr"`
	XmlnsWsu  string     `xml:"xmlns:wsu,attr"`
	Timestamp *Timestamp `xml:"wsu:Timestamp,omitempty"`
}

// Timestamp represents WS-Security timestamp
type Timestamp struct {
	XMLName xml.Name `xml:"wsu:Timestamp"`
	ID      string   `xml:"wsu:Id,attr,omitempty"`
	Created string   `xml:"wsu:Created"`
	Expires string   `xml:"wsu:Expires"`
}

// CustomHeaderElement represents a custom header element
type CustomHeaderElement struct {
	XMLName xml.Name
	Content string `xml:",innerxml"`
}

// EnvelopeBuilder helps construct SOAP envelopes
type EnvelopeBuilder struct {
	version     SOAPVersion
	headers     []CustomHeaderElement
	security    *SecurityHeader
	namespaces  map[string]string
	bodyContent string
}

// NewEnvelopeBuilder creates a new envelope builder
func NewEnvelopeBuilder(version SOAPVersion) *EnvelopeBuilder {
	return &EnvelopeBuilder{
		version:    version,
		headers:    make([]CustomHeaderElement, 0),
		namespaces: make(map[string]string),
	}
}

// NewSOAP11EnvelopeBuilder creates a new SOAP 1.1 envelope builder
func NewSOAP11EnvelopeBuilder() *EnvelopeBuilder {
	builder := NewEnvelopeBuilder(SOAP11)
	builder.AddNamespace("soap", SOAP11EnvelopeNS)
	builder.AddNamespace("xsi", XMLSchemaInstanceNS)
	builder.AddNamespace("xsd", XMLSchemaNS)
	return builder
}

// NewSOAP12EnvelopeBuilder creates a new SOAP 1.2 envelope builder
func NewSOAP12EnvelopeBuilder() *EnvelopeBuilder {
	builder := NewEnvelopeBuilder(SOAP12)
	builder.AddNamespace("soap", SOAP12EnvelopeNS)
	builder.AddNamespace("xsi", XMLSchemaInstanceNS)
	builder.AddNamespace("xsd", XMLSchemaNS)
	return builder
}

// AddNamespace adds a namespace to the envelope
func (b *EnvelopeBuilder) AddNamespace(prefix, uri string) *EnvelopeBuilder {
	b.namespaces[prefix] = uri
	return b
}

// SetBodyContent sets the body content of the envelope
func (b *EnvelopeBuilder) SetBodyContent(content string) *EnvelopeBuilder {
	b.bodyContent = content
	return b
}

// AddCustomHeader adds a custom header element
func (b *EnvelopeBuilder) AddCustomHeader(name xml.Name, content string) *EnvelopeBuilder {
	b.headers = append(b.headers, CustomHeaderElement{
		XMLName: name,
		Content: content,
	})
	return b
}

// AddSecurityHeader adds a WS-Security header with timestamp
func (b *EnvelopeBuilder) AddSecurityHeader(timestampID string, validMinutes int) *EnvelopeBuilder {
	now := time.Now().UTC()
	created := now.Format("2006-01-02T15:04:05.000Z")
	expires := now.Add(time.Duration(validMinutes) * time.Minute).Format("2006-01-02T15:04:05.000Z")

	b.security = &SecurityHeader{
		XmlnsWsse: WSSecurityNS,
		XmlnsWsu:  WSUtilityNS,
		Timestamp: &Timestamp{
			ID:      timestampID,
			Created: created,
			Expires: expires,
		},
	}

	return b
}

// Build constructs the final SOAP envelope
func (b *EnvelopeBuilder) Build() (*SOAPEnvelope, error) {
	if b.bodyContent == "" {
		return nil, errors.NewValidationError("SOAP body content cannot be empty", "bodyContent", "")
	}

	envelope := &SOAPEnvelope{
		Body: &SOAPBody{
			Content: b.bodyContent,
		},
	}

	// Set namespace attributes based on version
	switch b.version {
	case SOAP11:
		envelope.XmlnsSoap = SOAP11EnvelopeNS
	case SOAP12:
		envelope.XmlnsSoap = SOAP12EnvelopeNS
	default:
		return nil, errors.NewValidationError("unsupported SOAP version", "version", string(b.version))
	}

	// Add standard namespaces
	if uri, exists := b.namespaces["xsi"]; exists && uri != "" {
		envelope.XmlnsXsi = uri
	}
	if uri, exists := b.namespaces["xsd"]; exists && uri != "" {
		envelope.XmlnsXsd = uri
	}

	// Add headers if any exist
	if len(b.headers) > 0 || b.security != nil {
		envelope.Header = &SOAPHeader{
			Custom:   b.headers,
			Security: b.security,
		}
	}

	return envelope, nil
}

// ToXML converts the envelope to XML string
func (b *EnvelopeBuilder) ToXML() (string, error) {
	envelope, err := b.Build()
	if err != nil {
		return "", err
	}

	return envelope.ToXML()
}

// ToXML converts a SOAP envelope to XML string
func (e *SOAPEnvelope) ToXML() (string, error) {
	// Manually construct XML to avoid circular reference issues
	var xmlBuilder strings.Builder

	// Start envelope with namespace
	soapNS := e.XmlnsSoap
	if soapNS == "" {
		soapNS = SOAP11EnvelopeNS
	}

	xmlBuilder.WriteString(fmt.Sprintf(`<soap:Envelope xmlns:soap="%s"`, soapNS))

	if e.XmlnsXsi != "" {
		xmlBuilder.WriteString(fmt.Sprintf(` xmlns:xsi="%s"`, e.XmlnsXsi))
	}
	if e.XmlnsXsd != "" {
		xmlBuilder.WriteString(fmt.Sprintf(` xmlns:xsd="%s"`, e.XmlnsXsd))
	}
	xmlBuilder.WriteString(">")

	// Add header if present
	if e.Header != nil && (e.Header.Security != nil || len(e.Header.Custom) > 0) {
		xmlBuilder.WriteString("<soap:Header>")

		if e.Header.Security != nil {
			securityXML, err := xml.Marshal(e.Header.Security)
			if err == nil {
				xmlBuilder.Write(securityXML)
			}
		}

		for _, custom := range e.Header.Custom {
			customXML, err := xml.Marshal(custom)
			if err == nil {
				xmlBuilder.Write(customXML)
			}
		}

		xmlBuilder.WriteString("</soap:Header>")
	}

	// Add body
	xmlBuilder.WriteString("<soap:Body>")

	if e.Body != nil {
		if e.Body.Fault != nil {
			faultXML, err := xml.Marshal(e.Body.Fault)
			if err != nil {
				return "", errors.NewValidationError("failed to marshal SOAP fault", "fault", "")
			}
			xmlBuilder.Write(faultXML)
		} else if e.Body.Content != "" {
			xmlBuilder.WriteString(e.Body.Content)
		}
	}

	xmlBuilder.WriteString("</soap:Body>")
	xmlBuilder.WriteString("</soap:Envelope>")

	// Add XML declaration
	xmlString := xml.Header + xmlBuilder.String()

	return xmlString, nil
}

// ParseSOAPEnvelope parses a SOAP envelope from XML string
func ParseSOAPEnvelope(xmlData string) (*SOAPEnvelope, error) {
	var envelope SOAPEnvelope

	// First try to parse normally
	err := xml.Unmarshal([]byte(xmlData), &envelope)
	if err != nil {
		return nil, errors.NewValidationError("failed to parse SOAP envelope", "xml", xmlData[:min(len(xmlData), 100)])
	}

	// Post-process to handle fault vs content correctly
	if envelope.Body != nil {
		// Check if we have a fault by looking for fault elements in the content
		if strings.Contains(envelope.Body.Content, "<soap:Fault>") ||
			strings.Contains(envelope.Body.Content, "<faultcode>") {
			// Parse fault separately if it wasn't parsed correctly
			if envelope.Body.Fault == nil {
				var faultBody struct {
					XMLName xml.Name   `xml:"soap:Body"`
					Fault   *SOAPFault `xml:"soap:Fault"`
				}
				if faultErr := xml.Unmarshal([]byte(xmlData), &faultBody); faultErr == nil && faultBody.Fault != nil {
					envelope.Body.Fault = faultBody.Fault
				}
			}
		}
	}

	return &envelope, nil
}

// HasFault returns true if the envelope contains a SOAP fault
func (e *SOAPEnvelope) HasFault() bool {
	return e != nil && e.Body != nil && e.Body.Fault != nil
}

// GetFault returns the SOAP fault if present
func (e *SOAPEnvelope) GetFault() *SOAPFault {
	if e.HasFault() {
		return e.Body.Fault
	}
	return nil
}

// GetBodyContent returns the body content as string
func (e *SOAPEnvelope) GetBodyContent() string {
	if e.Body != nil {
		return e.Body.Content
	}
	return ""
}

// SetBodyContent sets the body content
func (e *SOAPEnvelope) SetBodyContent(content string) {
	if e.Body == nil {
		e.Body = &SOAPBody{}
	}
	e.Body.Content = content
}

// GetTimestamp returns the WS-Security timestamp if present
func (e *SOAPEnvelope) GetTimestamp() *Timestamp {
	if e.Header != nil && e.Header.Security != nil {
		return e.Header.Security.Timestamp
	}
	return nil
}

// ValidateTimestamp validates the WS-Security timestamp
func (e *SOAPEnvelope) ValidateTimestamp() error {
	timestamp := e.GetTimestamp()
	if timestamp == nil {
		return nil // No timestamp to validate
	}

	now := time.Now().UTC()

	// Parse created time
	created, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp.Created)
	if err != nil {
		return errors.NewValidationError("invalid timestamp created format", "created", timestamp.Created)
	}

	// Parse expires time
	expires, err := time.Parse("2006-01-02T15:04:05.000Z", timestamp.Expires)
	if err != nil {
		return errors.NewValidationError("invalid timestamp expires format", "expires", timestamp.Expires)
	}

	// Check if timestamp is valid
	if now.Before(created) {
		return errors.NewValidationError("timestamp created is in the future", "created", timestamp.Created)
	}

	if now.After(expires) {
		return errors.NewValidationError("timestamp has expired", "expires", timestamp.Expires)
	}

	return nil
}

// CreateNFeSOAPEnvelope creates a SOAP envelope specifically for NFe webservices
func CreateNFeSOAPEnvelope(bodyContent string) (*SOAPEnvelope, error) {
	builder := NewSOAP11EnvelopeBuilder()

	// Add NFe-specific namespaces if needed
	builder.AddNamespace("wsse", WSSecurityNS)
	builder.AddNamespace("wsu", WSUtilityNS)

	// Add security header with 5-minute validity
	builder.AddSecurityHeader("timestamp-1", 5)

	// Set the NFe body content
	builder.SetBodyContent(bodyContent)

	return builder.Build()
}

// CreateNFeSOAPRequest creates a complete SOAP request for NFe operations
func CreateNFeSOAPRequest(url, action, bodyContent string) (*SOAPRequest, error) {
	if url == "" {
		return nil, errors.NewValidationError("URL cannot be empty", "url", "")
	}

	if bodyContent == "" {
		return nil, errors.NewValidationError("body content cannot be empty", "bodyContent", "")
	}

	envelope, err := CreateNFeSOAPEnvelope(bodyContent)
	if err != nil {
		return nil, err
	}

	envelopeXML, err := envelope.ToXML()
	if err != nil {
		return nil, err
	}

	request := &SOAPRequest{
		URL:     url,
		Action:  action,
		Body:    envelopeXML,
		Headers: make(map[string]string),
	}

	return request, nil
}

// ExtractBodyContent extracts the body content from a SOAP response
func ExtractBodyContent(soapResponse string) (string, error) {
	envelope, err := ParseSOAPEnvelope(soapResponse)
	if err != nil {
		return "", err
	}

	if envelope.HasFault() {
		fault := envelope.GetFault()
		return "", errors.NewSEFAZError(
			fmt.Sprintf("SOAP Fault: %s", fault.String),
			fault.Code,
			fmt.Errorf("%s", fault.Detail),
		)
	}

	return envelope.GetBodyContent(), nil
}

// IsSOAPFaultResponse checks if a response contains a SOAP fault
func IsSOAPFaultResponse(responseBody string) bool {
	return strings.Contains(responseBody, "<soap:Fault>") ||
		strings.Contains(responseBody, "<Fault>") ||
		strings.Contains(responseBody, "faultcode") ||
		strings.Contains(responseBody, "faultstring")
}

// GetSOAPVersion detects the SOAP version from XML content
func GetSOAPVersion(xmlContent string) SOAPVersion {
	if strings.Contains(xmlContent, SOAP12EnvelopeNS) {
		return SOAP12
	}
	return SOAP11 // Default to SOAP 1.1
}

// CleanXMLContent removes XML declaration and normalizes the content
func CleanXMLContent(xmlContent string) string {
	// Remove XML declaration if present
	if strings.HasPrefix(xmlContent, "<?xml") {
		if idx := strings.Index(xmlContent, "?>"); idx != -1 {
			xmlContent = xmlContent[idx+2:]
		}
	}

	return strings.TrimSpace(xmlContent)
}

// AddXMLDeclaration adds XML declaration to content if not present
func AddXMLDeclaration(xmlContent string) string {
	cleaned := CleanXMLContent(xmlContent)
	return xml.Header + cleaned
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CreateSOAPFault creates a SOAP fault envelope
func CreateSOAPFault(faultCode, faultString, faultDetail string) (*SOAPEnvelope, error) {
	envelope := &SOAPEnvelope{
		XmlnsSoap: SOAP11EnvelopeNS,
		Body: &SOAPBody{
			Fault: &SOAPFault{
				Code:   faultCode,
				String: faultString,
				Detail: faultDetail,
			},
		},
	}

	return envelope, nil
}

// ValidateSOAPEnvelope performs basic validation on a SOAP envelope
func ValidateSOAPEnvelope(envelope *SOAPEnvelope) error {
	if envelope == nil {
		return errors.NewValidationError("envelope cannot be nil", "envelope", "")
	}

	if envelope.Body == nil {
		return errors.NewValidationError("SOAP body cannot be nil", "body", "")
	}

	if envelope.XmlnsSoap == "" {
		return errors.NewValidationError("SOAP namespace cannot be empty", "namespace", "")
	}

	// Validate namespace
	if envelope.XmlnsSoap != SOAP11EnvelopeNS && envelope.XmlnsSoap != SOAP12EnvelopeNS {
		return errors.NewValidationError("invalid SOAP namespace", "namespace", envelope.XmlnsSoap)
	}

	// Validate timestamp if present
	if err := envelope.ValidateTimestamp(); err != nil {
		return err
	}

	return nil
}
