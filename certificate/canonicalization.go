package certificate

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"sort"
	"strings"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/beevik/etree"
)

// CanonicalizationMethod represents different XML canonicalization methods
type CanonicalizationMethod string

const (
	// C14N10Inclusive represents Canonical XML 1.0 (inclusive)
	C14N10Inclusive CanonicalizationMethod = "http://www.w3.org/TR/2001/REC-xml-c14n-20010315"

	// C14N10Exclusive represents Canonical XML 1.0 (exclusive)
	C14N10Exclusive CanonicalizationMethod = "http://www.w3.org/2001/10/xml-exc-c14n#"

	// C14N11Inclusive represents Canonical XML 1.1 (inclusive)
	C14N11Inclusive CanonicalizationMethod = "http://www.w3.org/2006/12/xml-c14n11"

	// C14N11Exclusive represents Canonical XML 1.1 (exclusive)
	C14N11Exclusive CanonicalizationMethod = "http://www.w3.org/2006/12/xml-c14n11#WithComments"
)

// XMLCanonicalizer provides XML canonicalization functionality
type XMLCanonicalizer struct {
	method          CanonicalizationMethod
	inclusivePrefix string
	withComments    bool
}

// CanonicalizationConfig holds configuration for XML canonicalization
type CanonicalizationConfig struct {
	Method          CanonicalizationMethod `json:"method"`
	InclusivePrefix string                 `json:"inclusivePrefix"`
	WithComments    bool                   `json:"withComments"`
	TrimWhitespace  bool                   `json:"trimWhitespace"`
	SortAttributes  bool                   `json:"sortAttributes"`
	RemoveXMLDecl   bool                   `json:"removeXmlDecl"`
}

// NewXMLCanonicalizer creates a new XML canonicalizer
func NewXMLCanonicalizer(config *CanonicalizationConfig) *XMLCanonicalizer {
	if config == nil {
		config = DefaultCanonicalizationConfig()
	}

	return &XMLCanonicalizer{
		method:          config.Method,
		inclusivePrefix: config.InclusivePrefix,
		withComments:    config.WithComments,
	}
}

// DefaultCanonicalizationConfig returns default canonicalization configuration for SEFAZ
func DefaultCanonicalizationConfig() *CanonicalizationConfig {
	return &CanonicalizationConfig{
		Method:          C14N10Exclusive,
		InclusivePrefix: "",
		WithComments:    false,
		TrimWhitespace:  true,
		SortAttributes:  true,
		RemoveXMLDecl:   true,
	}
}

// Canonicalize canonicalizes XML content according to the specified method
func (canonicalizer *XMLCanonicalizer) Canonicalize(xmlContent string) ([]byte, error) {
	if xmlContent == "" {
		return nil, errors.NewValidationError("XML content cannot be empty", "xmlContent", "")
	}

	// Parse XML document
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return nil, errors.NewValidationError("failed to parse XML", "xml", err.Error())
	}

	return canonicalizer.CanonicalizeDocument(doc)
}

// CanonicalizeDocument canonicalizes an etree document
func (canonicalizer *XMLCanonicalizer) CanonicalizeDocument(doc *etree.Document) ([]byte, error) {
	if doc == nil {
		return nil, errors.NewValidationError("document cannot be nil", "document", "")
	}

	root := doc.Root()
	if root == nil {
		return nil, errors.NewValidationError("document has no root element", "root", "")
	}

	// Apply canonicalization
	canonicalizedRoot := canonicalizer.canonicalizeElement(root)

	// Create new document with canonicalized root
	canonDoc := etree.NewDocument()
	canonDoc.SetRoot(canonicalizedRoot)

	// Configure output settings for canonicalization
	doc.WriteSettings.CanonicalEndTags = true
	doc.WriteSettings.CanonicalText = true
	doc.WriteSettings.CanonicalAttrVal = true

	return canonDoc.WriteToBytes()
}

// CanonicalizeElement canonicalizes a specific XML element
func (canonicalizer *XMLCanonicalizer) CanonicalizeElement(element *etree.Element) ([]byte, error) {
	if element == nil {
		return nil, errors.NewValidationError("element cannot be nil", "element", "")
	}

	canonicalizedElement := canonicalizer.canonicalizeElement(element)

	// Create temporary document
	tempDoc := etree.NewDocument()
	tempDoc.SetRoot(canonicalizedElement)

	// Configure output settings for canonicalization
	tempDoc.WriteSettings.CanonicalEndTags = true
	tempDoc.WriteSettings.CanonicalText = true
	tempDoc.WriteSettings.CanonicalAttrVal = true

	return tempDoc.WriteToBytes()
}

// canonicalizeElement performs the actual canonicalization of an element
func (canonicalizer *XMLCanonicalizer) canonicalizeElement(element *etree.Element) *etree.Element {
	// Create a copy of the element
	canonical := element.Copy()

	// Sort attributes according to C14N rules
	canonicalizer.sortAttributes(canonical)

	// Normalize namespace declarations
	canonicalizer.normalizeNamespaces(canonical)

	// Process child elements recursively
	for _, child := range canonical.Child {
		if childElement, ok := child.(*etree.Element); ok {
			canonicalizedChild := canonicalizer.canonicalizeElement(childElement)
			// Replace the child with the canonicalized version
			canonical.RemoveChild(childElement)
			canonical.AddChild(canonicalizedChild)
		} else if charData, ok := child.(*etree.CharData); ok {
			// Normalize text content
			canonicalizer.normalizeText(charData)
		}
	}

	// Remove comments if not preserving them
	if !canonicalizer.withComments {
		canonicalizer.removeComments(canonical)
	}

	return canonical
}

// sortAttributes sorts attributes according to C14N specification
func (canonicalizer *XMLCanonicalizer) sortAttributes(element *etree.Element) {
	if len(element.Attr) <= 1 {
		return
	}

	// Separate namespace declarations from regular attributes
	var nsAttrs []etree.Attr
	var regularAttrs []etree.Attr

	for _, attr := range element.Attr {
		if attr.Space == "xmlns" || attr.Key == "xmlns" {
			nsAttrs = append(nsAttrs, attr)
		} else {
			regularAttrs = append(regularAttrs, attr)
		}
	}

	// Sort namespace declarations
	sort.Slice(nsAttrs, func(i, j int) bool {
		// xmlns comes before xmlns:prefix
		if nsAttrs[i].Key == "xmlns" && nsAttrs[j].Key != "xmlns" {
			return true
		}
		if nsAttrs[i].Key != "xmlns" && nsAttrs[j].Key == "xmlns" {
			return false
		}
		return nsAttrs[i].Key < nsAttrs[j].Key
	})

	// Sort regular attributes
	sort.Slice(regularAttrs, func(i, j int) bool {
		// Compare by namespace URI first, then by local name
		if regularAttrs[i].Space != regularAttrs[j].Space {
			return regularAttrs[i].Space < regularAttrs[j].Space
		}
		return regularAttrs[i].Key < regularAttrs[j].Key
	})

	// Rebuild attribute list: namespace declarations first, then regular attributes
	element.Attr = nil
	element.Attr = append(element.Attr, nsAttrs...)
	element.Attr = append(element.Attr, regularAttrs...)
}

// normalizeNamespaces normalizes namespace declarations according to C14N rules
func (canonicalizer *XMLCanonicalizer) normalizeNamespaces(element *etree.Element) {
	// For exclusive canonicalization, remove unused namespace declarations
	if canonicalizer.method == C14N10Exclusive || canonicalizer.method == C14N11Exclusive {
		canonicalizer.removeUnusedNamespaces(element)
	}
}

// removeUnusedNamespaces removes namespace declarations that are not used
func (canonicalizer *XMLCanonicalizer) removeUnusedNamespaces(element *etree.Element) {
	usedNamespaces := make(map[string]bool)

	// Collect used namespaces from element and its descendants
	canonicalizer.collectUsedNamespaces(element, usedNamespaces)

	// Remove unused namespace declarations
	var filteredAttrs []etree.Attr
	for _, attr := range element.Attr {
		if attr.Space == "xmlns" {
			if usedNamespaces[attr.Key] || canonicalizer.isInclusivePrefix(attr.Key) {
				filteredAttrs = append(filteredAttrs, attr)
			}
		} else if attr.Key == "xmlns" {
			if usedNamespaces[""] || canonicalizer.isInclusivePrefix("") {
				filteredAttrs = append(filteredAttrs, attr)
			}
		} else {
			filteredAttrs = append(filteredAttrs, attr)
		}
	}

	element.Attr = filteredAttrs
}

// collectUsedNamespaces collects all namespace prefixes used in an element and its descendants
func (canonicalizer *XMLCanonicalizer) collectUsedNamespaces(element *etree.Element, used map[string]bool) {
	// Mark namespace of the element itself
	if element.Space != "" {
		used[element.Space] = true
	}

	// Mark namespaces used in attributes
	for _, attr := range element.Attr {
		if attr.Space != "" && attr.Space != "xmlns" {
			used[attr.Space] = true
		}
	}

	// Recursively process child elements
	for _, child := range element.Child {
		if childElement, ok := child.(*etree.Element); ok {
			canonicalizer.collectUsedNamespaces(childElement, used)
		}
	}
}

// isInclusivePrefix checks if a prefix should be included (for inclusive prefix list)
func (canonicalizer *XMLCanonicalizer) isInclusivePrefix(prefix string) bool {
	if canonicalizer.inclusivePrefix == "" {
		return false
	}

	prefixes := strings.Split(canonicalizer.inclusivePrefix, " ")
	for _, p := range prefixes {
		if strings.TrimSpace(p) == prefix {
			return true
		}
	}
	return false
}

// normalizeText normalizes text content according to C14N rules
func (canonicalizer *XMLCanonicalizer) normalizeText(charData *etree.CharData) {
	// Normalize line endings and character references
	normalized := strings.ReplaceAll(charData.Data, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	charData.Data = normalized
}

// removeComments removes comment nodes from the element tree
func (canonicalizer *XMLCanonicalizer) removeComments(element *etree.Element) {
	var filteredChildren []etree.Token

	for _, child := range element.Child {
		if _, isComment := child.(*etree.Comment); !isComment {
			filteredChildren = append(filteredChildren, child)

			// Recursively remove comments from child elements
			if childElement, ok := child.(*etree.Element); ok {
				canonicalizer.removeComments(childElement)
			}
		}
	}

	element.Child = filteredChildren
}

// C14NTransform applies the enveloped signature transform
func C14NTransform(xmlContent string) ([]byte, error) {
	canonicalizer := NewXMLCanonicalizer(DefaultCanonicalizationConfig())
	return canonicalizer.Canonicalize(xmlContent)
}

// EnvelopedSignatureTransform removes the signature element and canonicalizes
func EnvelopedSignatureTransform(xmlContent string) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return nil, errors.NewValidationError("failed to parse XML", "xml", err.Error())
	}

	// Remove signature elements
	for {
		sigElement := doc.FindElement(".//ds:Signature")
		if sigElement == nil {
			break
		}
		sigElement.Parent().RemoveChild(sigElement)
	}

	// Apply canonicalization
	canonicalizer := NewXMLCanonicalizer(DefaultCanonicalizationConfig())
	return canonicalizer.CanonicalizeDocument(doc)
}

// CanonicalizeForSigning prepares XML for signing by applying appropriate transforms
func CanonicalizeForSigning(xmlContent string, elementID string) ([]byte, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return nil, errors.NewValidationError("failed to parse XML", "xml", err.Error())
	}

	var elementToCanonnicalize *etree.Element

	if elementID != "" {
		// Find specific element by ID
		selectors := []string{
			fmt.Sprintf(".//*[@Id='%s']", elementID),
			fmt.Sprintf(".//*[@id='%s']", elementID),
			fmt.Sprintf(".//*[@ID='%s']", elementID),
		}

		for _, selector := range selectors {
			if element := doc.FindElement(selector); element != nil {
				elementToCanonnicalize = element
				break
			}
		}

		if elementToCanonnicalize == nil {
			return nil, errors.NewValidationError("element with specified ID not found", "elementID", elementID)
		}
	} else {
		// Use root element
		elementToCanonnicalize = doc.Root()
	}

	// Remove any existing signature elements from the element being signed
	for {
		sigElement := elementToCanonnicalize.FindElement(".//ds:Signature")
		if sigElement == nil {
			break
		}
		elementToCanonnicalize.RemoveChild(sigElement)
	}

	// Apply canonicalization
	canonicalizer := NewXMLCanonicalizer(DefaultCanonicalizationConfig())
	return canonicalizer.CanonicalizeElement(elementToCanonnicalize)
}

// ValidateCanonicalForm validates if XML is in canonical form
func ValidateCanonicalForm(xmlContent string) error {
	canonicalizer := NewXMLCanonicalizer(DefaultCanonicalizationConfig())

	// Canonicalize the content
	canonicalized, err := canonicalizer.Canonicalize(xmlContent)
	if err != nil {
		return err
	}

	// Compare with original (simplified check)
	doc := etree.NewDocument()
	if err := doc.ReadFromString(xmlContent); err != nil {
		return errors.NewValidationError("failed to parse original XML", "xml", err.Error())
	}

	original, err := doc.WriteToBytes()
	if err != nil {
		return errors.NewValidationError("failed to serialize original XML", "xml", err.Error())
	}

	// Simple byte comparison (in practice, you might want more sophisticated comparison)
	if !bytes.Equal(canonicalized, original) {
		return errors.NewValidationError("XML is not in canonical form", "form", "not-canonical")
	}

	return nil
}

// GetICPBrasilRootCertificates returns the ICP-Brasil root certificates
func GetICPBrasilRootCertificates() []*x509.Certificate {
	// This would normally load real ICP-Brasil root certificates
	// For now, return empty slice - in production, load from embedded data or files
	return []*x509.Certificate{}
}

// CanonicalizationMethodFromURI returns the canonicalization method from URI
func CanonicalizationMethodFromURI(uri string) CanonicalizationMethod {
	switch uri {
	case "http://www.w3.org/TR/2001/REC-xml-c14n-20010315":
		return C14N10Inclusive
	case "http://www.w3.org/2001/10/xml-exc-c14n#":
		return C14N10Exclusive
	case "http://www.w3.org/2006/12/xml-c14n11":
		return C14N11Inclusive
	case "http://www.w3.org/2006/12/xml-c14n11#WithComments":
		return C14N11Exclusive
	default:
		return C14N10Exclusive // Default to most common
	}
}

// CanonicalizationURIFromMethod returns the URI for a canonicalization method
func CanonicalizationURIFromMethod(method CanonicalizationMethod) string {
	return string(method)
}
