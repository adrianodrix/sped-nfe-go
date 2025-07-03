package certificate

import (
	"testing"
)

func TestSEFAZ298Validation(t *testing.T) {
	// Create a mock XMLDSig signer for testing
	mockCert := NewMockCertificate()
	signer := NewXMLDSigSigner(mockCert, DefaultXMLDSigConfig())

	// Test NFe XML structure (simplified)
	nfeXML := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
	<infNFe Id="NFe41250710541434000152550010000000011876543215">
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
		<emit>
			<CNPJ>10541434000152</CNPJ>
			<xNome>Empresa Teste LTDA</xNome>
			<enderEmit>
				<xLgr>Rua Teste</xLgr>
				<nro>123</nro>
				<xBairro>Centro</xBairro>
				<cMun>4106902</cMun>
				<xMun>Curitiba</xMun>
				<UF>PR</UF>
				<CEP>80010000</CEP>
			</enderEmit>
			<IE>1234567890</IE>
		</emit>
		<dest>
			<CPF>12345678901</CPF>
			<xNome>Cliente Teste</xNome>
			<enderDest>
				<xLgr>Rua Cliente</xLgr>
				<nro>456</nro>
				<xBairro>Bairro Cliente</xBairro>
				<cMun>4106902</cMun>
				<xMun>Curitiba</xMun>
				<UF>PR</UF>
				<CEP>80020000</CEP>
			</enderDest>
		</dest>
		<det nItem="1">
			<prod>
				<cProd>001</cProd>
				<cEAN/>
				<xProd>Produto Teste</xProd>
				<NCM>12345678</NCM>
				<CFOP>5102</CFOP>
				<uCom>UN</uCom>
				<qCom>1.0000</qCom>
				<vUnCom>100.00</vUnCom>
				<vProd>100.00</vProd>
				<cEANTrib/>
				<uTrib>UN</uTrib>
				<qTrib>1.0000</qTrib>
				<vUnTrib>100.00</vUnTrib>
			</prod>
			<imposto>
				<ICMS>
					<ICMS00>
						<orig>0</orig>
						<CST>00</CST>
						<modBC>0</modBC>
						<vBC>100.00</vBC>
						<pICMS>18.00</pICMS>
						<vICMS>18.00</vICMS>
					</ICMS00>
				</ICMS>
			</imposto>
		</det>
		<total>
			<ICMSTot>
				<vBC>100.00</vBC>
				<vICMS>18.00</vICMS>
				<vICMSDeson>0.00</vICMSDeson>
				<vFCP>0.00</vFCP>
				<vBCST>0.00</vBCST>
				<vST>0.00</vST>
				<vFCPST>0.00</vFCPST>
				<vFCPSTRet>0.00</vFCPSTRet>
				<vProd>100.00</vProd>
				<vFrete>0.00</vFrete>
				<vSeg>0.00</vSeg>
				<vDesc>0.00</vDesc>
				<vII>0.00</vII>
				<vIPI>0.00</vIPI>
				<vIPIDevol>0.00</vIPIDevol>
				<vPIS>0.00</vPIS>
				<vCOFINS>0.00</vCOFINS>
				<vOutro>0.00</vOutro>
				<vNF>100.00</vNF>
			</ICMSTot>
		</total>
		<transp>
			<modFrete>9</modFrete>
		</transp>
		<pag>
			<detPag>
				<tPag>01</tPag>
				<vPag>100.00</vPag>
			</detPag>
		</pag>
	</infNFe>
</NFe>`

	// Sign the NFe XML
	result, err := signer.SignNFeXML(nfeXML)
	if err != nil {
		t.Fatalf("Failed to sign NFe XML: %v", err)
	}

	if result == nil {
		t.Fatal("Signature result is nil")
	}

	// Validate against SEFAZ 298 requirements
	validation, err := ValidateSEFAZ298Compliance(result.SignedXML)
	if err != nil {
		t.Fatalf("Failed to validate SEFAZ 298 compliance: %v", err)
	}

	// Print validation report for debugging
	PrintSEFAZ298ValidationReport(validation)

	// Test specific requirements for error 298
	if !validation.HasReferenceURI {
		t.Error("Missing Reference URI - this will cause SEFAZ error 298")
	}

	if !validation.HasIdAttribute {
		t.Error("Id attribute not properly signed - this will cause SEFAZ error 298")
	}

	if !validation.HasEnvelopedTransform {
		t.Error("Missing enveloped-signature Transform - this will cause SEFAZ error 298")
	}

	if !validation.HasC14NTransform {
		t.Error("Missing C14N Transform - this will cause SEFAZ error 298")
	}

	if !validation.IsValid {
		t.Error("Signature does not meet SEFAZ requirements - will be rejected with error 298")
		for _, issue := range validation.Issues {
			t.Logf("Issue: %s", issue)
		}
	}

	// Log the complete signed XML for manual inspection
	t.Logf("Complete signed NFe XML:\n%s", result.SignedXML)
}

func TestSignatureStructureCompliance(t *testing.T) {
	// Test with a minimal but complete XML structure
	minimalNFeXML := `<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
	<infNFe Id="NFe41250710541434000152550010000000011876543215">
		<ide>
			<cUF>41</cUF>
			<natOp>Venda</natOp>
			<mod>55</mod>
		</ide>
	</infNFe>
</NFe>`

	mockCert := NewMockCertificate()
	signer := NewXMLDSigSigner(mockCert, DefaultXMLDSigConfig())

	result, err := signer.SignNFeXML(minimalNFeXML)
	if err != nil {
		t.Fatalf("Failed to sign minimal NFe XML: %v", err)
	}

	// Validate the signature structure
	validation, err := ValidateSEFAZ298Compliance(result.SignedXML)
	if err != nil {
		t.Fatalf("Failed to validate signature: %v", err)
	}

	// All critical elements should be present
	if !validation.HasCorrectNamespace {
		t.Error("Signature missing correct XML namespace")
	}

	if !validation.HasSignatureMethod {
		t.Error("Signature missing SignatureMethod")
	}

	if !validation.HasDigestMethod {
		t.Error("Signature missing DigestMethod")
	}

	if !validation.HasDigestValue {
		t.Error("Signature missing DigestValue")
	}

	if !validation.HasSignatureValue {
		t.Error("Signature missing SignatureValue")
	}

	// Log for debugging
	t.Logf("Minimal signature validation - Valid: %v", validation.IsValid)
	if len(validation.Issues) > 0 {
		for _, issue := range validation.Issues {
			t.Logf("Issue: %s", issue)
		}
	}
}