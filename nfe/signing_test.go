package nfe

import (
	"testing"

	"github.com/adrianodrix/sped-nfe-go/certificate"
)

func TestSignIfNeeded(t *testing.T) {
	client, err := NewClient(ClientConfig{
		Environment: Homologation,
		UF:          35, // SP
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with no certificate
	t.Run("no certificate", func(t *testing.T) {
		xmlData := `<?xml version="1.0" encoding="UTF-8"?><NFe><infNFe Id="test">content</infNFe></NFe>`
		_, err := client.signIfNeeded([]byte(xmlData))
		if err == nil {
			t.Error("expected error when no certificate is set")
		}
	})

	// Test with mock certificate
	t.Run("with mock certificate", func(t *testing.T) {
		mockCert := certificate.NewMockCertificate()
		client.SetCertificate(mockCert)

		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe35200114200166000187550010000000046362100024" versao="4.00">
    <ide>
      <cUF>35</cUF>
      <natOp>Venda</natOp>
    </ide>
    <emit>
      <xNome>Test</xNome>
    </emit>
    <det nItem="1">
      <prod>
        <xProd>Test Product</xProd>
      </prod>
    </det>
  </infNFe>
</NFe>`

		signedXML, err := client.signIfNeeded([]byte(xmlData))
		if err != nil {
			t.Fatalf("unexpected error signing XML: %v", err)
		}

		if len(signedXML) == 0 {
			t.Error("signed XML should not be empty")
		}

		// Check if signature was added (basic check)
		signedStr := string(signedXML)
		if !stringContains(signedStr, "ds:Signature") && !stringContains(signedStr, "Signature") {
			t.Error("signed XML should contain signature elements")
		}
	})

	// Test with already signed XML
	t.Run("already signed XML", func(t *testing.T) {
		mockCert := certificate.NewMockCertificate()
		client.SetCertificate(mockCert)

		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="test">
    <content>test</content>
  </infNFe>
  <ds:Signature xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
    <ds:SignatureValue>already-signed</ds:SignatureValue>
  </ds:Signature>
</NFe>`

		result, err := client.signIfNeeded([]byte(xmlData))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return the same XML since it's already signed
		if string(result) != xmlData {
			t.Error("already signed XML should be returned unchanged")
		}
	})
}

func TestXMLSigningIntegration(t *testing.T) {
	client, err := NewClient(ClientConfig{
		Environment: Homologation,
		UF:          35, // SP
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Create mock certificate
	mockCert := certificate.NewMockCertificate()
	err = client.SetCertificate(mockCert)
	if err != nil {
		t.Fatalf("Failed to set certificate: %v", err)
	}

	// Test integration with authorization flow
	t.Run("authorization with signing", func(t *testing.T) {
		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe35200114200166000187550010000000046362100024" versao="4.00">
    <ide>
      <cUF>35</cUF>
      <cNF>12345678</cNF>
      <natOp>Venda de mercadoria</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>1</nNF>
      <dhEmi>2020-01-01T10:00:00-03:00</dhEmi>
      <tpNF>1</tpNF>
      <idDest>1</idDest>
      <cMunFG>3550308</cMunFG>
      <tpImp>1</tpImp>
      <tpEmis>1</tpEmis>
      <cDV>1</cDV>
      <tpAmb>2</tpAmb>
      <finNFe>1</finNFe>
      <indFinal>1</indFinal>
      <indPres>1</indPres>
      <procEmi>0</procEmi>
      <verProc>1.0</verProc>
    </ide>
    <emit>
      <CNPJ>14200166000187</CNPJ>
      <xNome>Empresa Teste LTDA</xNome>
      <IE>123456789</IE>
      <CRT>3</CRT>
    </emit>
    <det nItem="1">
      <prod>
        <cProd>001</cProd>
        <xProd>Produto Teste</xProd>
        <NCM>12345678</NCM>
        <CFOP>5102</CFOP>
        <uCom>UN</uCom>
        <qCom>1.0000</qCom>
        <vUnCom>100.00</vUnCom>
        <vProd>100.00</vProd>
        <uTrib>UN</uTrib>
        <qTrib>1.0000</qTrib>
        <vUnTrib>100.00</vUnTrib>
        <indTot>1</indTot>
      </prod>
    </det>
    <total>
      <ICMSTot>
        <vBC>100.00</vBC>
        <vICMS>18.00</vICMS>
        <vProd>100.00</vProd>
        <vNF>100.00</vNF>
      </ICMSTot>
    </total>
    <transp>
      <modFrete>9</modFrete>
    </transp>
  </infNFe>
</NFe>`

		// Test that signing works in the authorization flow
		signedXML, err := client.signIfNeeded([]byte(xmlData))
		if err != nil {
			t.Fatalf("Failed to sign XML: %v", err)
		}

		if len(signedXML) == 0 {
			t.Error("Signed XML should not be empty")
		}

		// Verify basic structure is preserved
		signedStr := string(signedXML)
		if !stringContains(signedStr, "Empresa Teste LTDA") {
			t.Error("Signed XML should preserve original content")
		}

		if !stringContains(signedStr, "NFe35200114200166000187550010000000046362100024") {
			t.Error("Signed XML should preserve NFe ID")
		}
	})
}

// Helper function to check if a string contains a substring
func stringContains(haystack, needle string) bool {
	return len(haystack) >= len(needle) &&
		(haystack == needle ||
			haystack[:len(needle)] == needle ||
			haystack[len(haystack)-len(needle):] == needle ||
			stringContainsAt(haystack, needle))
}

func stringContainsAt(haystack, needle string) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
