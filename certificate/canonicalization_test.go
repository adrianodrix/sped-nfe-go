package certificate

import (
	"strings"
	"testing"

	"github.com/beevik/etree"
)

func TestCanonicalization(t *testing.T) {
	// Create a mock XMLDSig signer for testing
	mockCert := NewMockCertificate()
	signer := NewXMLDSigSigner(mockCert, DefaultXMLDSigConfig())

	// Test XML with the structure similar to NFe infNFe element
	testXML := `<infNFe Id="NFe41250710541434000152550010000000011876543215">
		<ide>
			<cUF>41</cUF>
			<cNF>87654321</cNF>
			<natOp>Venda</natOp>
			<mod>55</mod>
			<serie>1</serie>
			<nNF>1</nNF>
			<dhEmi>2025-07-01T10:00:00-03:00</dhEmi>
			<tpNF>1</tpNF>
			<idDest>1</idDest>
			<cMunFG>4106902</cMunFG>
			<tpImp>1</tpImp>
			<tpEmis>1</tpEmis>
			<cDV>5</cDV>
			<tpAmb>2</tpAmb>
			<finNFe>1</finNFe>
			<indFinal>1</indFinal>
			<indPres>1</indPres>
		</ide>
	</infNFe>`

	// Parse the XML
	doc := etree.NewDocument()
	err := doc.ReadFromString(testXML)
	if err != nil {
		t.Fatalf("Failed to parse test XML: %v", err)
	}

	element := doc.Root()
	if element == nil {
		t.Fatal("No root element found")
	}

	// Test canonicalization
	canonicalized, err := signer.canonicalizeElement(element)
	if err != nil {
		t.Fatalf("Canonicalization failed: %v", err)
	}

	// Verify the canonicalized output
	if canonicalized == "" {
		t.Error("Canonicalized output is empty")
	}

	// Should not contain XML declaration
	if strings.Contains(canonicalized, "<?xml") {
		t.Error("Canonicalized output contains XML declaration")
	}

	// Should be properly formatted
	if !strings.Contains(canonicalized, "infNFe") {
		t.Error("Canonicalized output missing infNFe element")
	}

	// Should maintain the Id attribute
	if !strings.Contains(canonicalized, `Id="NFe41250710541434000152550010000000011876543215"`) {
		t.Error("Canonicalized output missing or incorrect Id attribute")
	}

	t.Logf("Canonicalized output: %s", canonicalized)
}

func TestCanonicalizeSignedInfo(t *testing.T) {
	// Create a mock XMLDSig signer for testing
	mockCert := NewMockCertificate()
	signer := NewXMLDSigSigner(mockCert, DefaultXMLDSigConfig())

	// Create a SignedInfo element similar to what we generate
	signedInfoXML := `<SignedInfo>
		<CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
		<SignatureMethod Algorithm="http://www.w3.org/2000/09/xmldsig#rsa-sha1"/>
		<Reference URI="#NFe41250710541434000152550010000000011876543215">
			<Transforms>
				<Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"/>
				<Transform Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"/>
			</Transforms>
			<DigestMethod Algorithm="http://www.w3.org/2000/09/xmldsig#sha1"/>
			<DigestValue>ueer0M35jDJeuby4dw3oQzy7P5k=</DigestValue>
		</Reference>
	</SignedInfo>`

	// Parse the SignedInfo XML
	doc := etree.NewDocument()
	err := doc.ReadFromString(signedInfoXML)
	if err != nil {
		t.Fatalf("Failed to parse SignedInfo XML: %v", err)
	}

	signedInfo := doc.Root()
	if signedInfo == nil {
		t.Fatal("No SignedInfo element found")
	}

	// Test SignedInfo canonicalization
	canonicalizedBytes, err := signer.canonicalizeSignedInfo(signedInfo)
	if err != nil {
		t.Fatalf("SignedInfo canonicalization failed: %v", err)
	}

	canonicalized := string(canonicalizedBytes)

	// Verify the canonicalized output
	if canonicalized == "" {
		t.Error("Canonicalized SignedInfo output is empty")
	}

	// Should not contain XML declaration
	if strings.Contains(canonicalized, "<?xml") {
		t.Error("Canonicalized SignedInfo output contains XML declaration")
	}

	// Should contain all required elements
	if !strings.Contains(canonicalized, "CanonicalizationMethod") {
		t.Error("Missing CanonicalizationMethod in canonicalized SignedInfo")
	}

	if !strings.Contains(canonicalized, "SignatureMethod") {
		t.Error("Missing SignatureMethod in canonicalized SignedInfo")
	}

	if !strings.Contains(canonicalized, "Reference") {
		t.Error("Missing Reference in canonicalized SignedInfo")
	}

	if !strings.Contains(canonicalized, `URI="#NFe41250710541434000152550010000000011876543215"`) {
		t.Error("Missing or incorrect URI in canonicalized SignedInfo")
	}

	// Should contain Transform algorithms as required by SEFAZ
	if !strings.Contains(canonicalized, "enveloped-signature") {
		t.Error("Missing enveloped-signature transform in canonicalized SignedInfo")
	}

	if !strings.Contains(canonicalized, "xml-c14n-20010315") {
		t.Error("Missing C14N transform in canonicalized SignedInfo")
	}

	t.Logf("Canonicalized SignedInfo: %s", canonicalized)
}

func TestSignatureGeneration(t *testing.T) {
	// Create a mock XMLDSig signer for testing
	mockCert := NewMockCertificate()
	signer := NewXMLDSigSigner(mockCert, DefaultXMLDSigConfig())

	// Test creating signature element with digest
	digestValue := "ueer0M35jDJeuby4dw3oQzy7P5k="
	referenceURI := "#NFe41250710541434000152550010000000011876543215"

	signature := signer.createSignatureElementWithDigest(referenceURI, digestValue)

	// Verify signature structure
	if signature == nil {
		t.Fatal("Generated signature is nil")
	}

	if signature.Tag != "Signature" {
		t.Errorf("Expected Signature element, got %s", signature.Tag)
	}

	// Check namespace
	xmlns := signature.SelectAttr("xmlns")
	if xmlns == nil || xmlns.Value != "http://www.w3.org/2000/09/xmldsig#" {
		t.Error("Missing or incorrect xmlns namespace")
	}

	// Check SignedInfo
	signedInfo := signature.FindElement("SignedInfo")
	if signedInfo == nil {
		t.Error("Missing SignedInfo element")
	}

	// Check Reference URI
	reference := signedInfo.FindElement(".//Reference")
	if reference == nil {
		t.Error("Missing Reference element")
	}

	uriAttr := reference.SelectAttr("URI")
	if uriAttr == nil || uriAttr.Value != referenceURI {
		t.Errorf("Expected URI %s, got %v", referenceURI, uriAttr)
	}

	// Check Transforms
	transforms := reference.FindElement("Transforms")
	if transforms == nil {
		t.Error("Missing Transforms element")
	}

	transformElements := transforms.FindElements("Transform")
	if len(transformElements) != 2 {
		t.Errorf("Expected 2 Transform elements, got %d", len(transformElements))
	}

	// Check specific transform algorithms
	algorithms := make([]string, 0, len(transformElements))
	for _, transform := range transformElements {
		if algAttr := transform.SelectAttr("Algorithm"); algAttr != nil {
			algorithms = append(algorithms, algAttr.Value)
		}
	}

	expectedAlgorithms := []string{
		"http://www.w3.org/2000/09/xmldsig#enveloped-signature",
		"http://www.w3.org/TR/2001/REC-xml-c14n-20010315",
	}

	for _, expected := range expectedAlgorithms {
		found := false
		for _, actual := range algorithms {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing expected transform algorithm: %s", expected)
		}
	}

	// Check DigestValue
	digestValueElement := reference.FindElement("DigestValue")
	if digestValueElement == nil {
		t.Error("Missing DigestValue element")
	} else if digestValueElement.Text() != digestValue {
		t.Errorf("Expected DigestValue %s, got %s", digestValue, digestValueElement.Text())
	}

	// Convert to string and verify structure
	doc := etree.NewDocument()
	doc.SetRoot(signature)
	xmlString, err := doc.WriteToString()
	if err != nil {
		t.Fatalf("Failed to serialize signature: %v", err)
	}

	t.Logf("Generated signature structure:\n%s", xmlString)
}