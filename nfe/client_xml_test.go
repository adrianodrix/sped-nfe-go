package nfe

import (
	"strings"
	"testing"
)

func TestLoadFromXMLBasic(t *testing.T) {
	client, err := NewClient(ClientConfig{
		Environment: Homologation,
		UF:          35, // SP
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with empty XML
	t.Run("empty XML", func(t *testing.T) {
		_, err := client.LoadFromXML([]byte(""))
		if err == nil {
			t.Error("expected error for empty XML")
		}
	})

	// Test with invalid XML
	t.Run("invalid XML", func(t *testing.T) {
		_, err := client.LoadFromXML([]byte("invalid xml"))
		if err == nil {
			t.Error("expected error for invalid XML")
		}
	})

	// Test with basic valid XML
	t.Run("basic valid XML", func(t *testing.T) {
		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe35200114200166000187550010000000046362100024" versao="4.00">
    <ide>
      <cUF>35</cUF>
      <cNF>62100024</cNF>
      <natOp>Venda de mercadoria</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>4</nNF>
      <dhEmi>2020-01-01T10:00:00-03:00</dhEmi>
      <tpNF>1</tpNF>
      <idDest>1</idDest>
      <cMunFG>3550308</cMunFG>
      <tpImp>1</tpImp>
      <tpEmis>1</tpEmis>
      <cDV>4</cDV>
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
      <enderEmit>
        <xLgr>Rua teste</xLgr>
        <nro>123</nro>
        <xBairro>Centro</xBairro>
        <cMun>3550308</cMun>
        <xMun>São Paulo</xMun>
        <UF>SP</UF>
        <CEP>01000000</CEP>
      </enderEmit>
      <IE>123456789</IE>
      <CRT>3</CRT>
    </emit>
    <det nItem="1">
      <prod>
        <cProd>001</cProd>
        <cEAN>SEM GTIN</cEAN>
        <xProd>Produto Teste</xProd>
        <NCM>12345678</NCM>
        <CFOP>5102</CFOP>
        <uCom>UN</uCom>
        <qCom>1.0000</qCom>
        <vUnCom>100.00</vUnCom>
        <vProd>100.00</vProd>
        <cEANTrib>SEM GTIN</cEANTrib>
        <uTrib>UN</uTrib>
        <qTrib>1.0000</qTrib>
        <vUnTrib>100.00</vUnTrib>
        <indTot>1</indTot>
      </prod>
      <imposto>
        <ICMS>
          <ICMS00>
            <orig>0</orig>
            <CST>00</CST>
            <modBC>3</modBC>
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
  </infNFe>
</NFe>`

		nfe, err := client.LoadFromXML([]byte(xmlData))
		if err != nil {
			t.Fatalf("unexpected error parsing valid XML: %v", err)
		}

		if nfe == nil {
			t.Fatal("NFe should not be nil")
		}

		// Validate parsed data
		if nfe.InfNFe.ID != "NFe35200114200166000187550010000000046362100024" {
			t.Errorf("expected ID 'NFe35200114200166000187550010000000046362100024', got '%s'", nfe.InfNFe.ID)
		}

		if nfe.InfNFe.Versao != "4.00" {
			t.Errorf("expected version '4.00', got '%s'", nfe.InfNFe.Versao)
		}

		if nfe.InfNFe.Emit.XNome != "Empresa Teste LTDA" {
			t.Errorf("expected issuer name 'Empresa Teste LTDA', got '%s'", nfe.InfNFe.Emit.XNome)
		}

		if len(nfe.InfNFe.Det) != 1 {
			t.Errorf("expected 1 item, got %d", len(nfe.InfNFe.Det))
		}
	})

	// Test XML without required fields
	t.Run("missing required fields", func(t *testing.T) {
		xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe versao="4.00">
    <ide>
      <natOp>Venda</natOp>
    </ide>
  </infNFe>
</NFe>`

		_, err := client.LoadFromXML([]byte(xmlData))
		if err == nil {
			t.Error("expected error for missing required fields")
		}
	})
}

func TestAddProtocolBasic(t *testing.T) {
	client, err := NewClient(ClientConfig{
		Environment: Homologation,
		UF:          35, // SP
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test with empty inputs
	t.Run("empty NFe XML", func(t *testing.T) {
		_, err := client.AddProtocol([]byte(""), []byte("protocol"))
		if err == nil {
			t.Error("expected error for empty NFe XML")
		}
	})

	t.Run("empty protocol XML", func(t *testing.T) {
		_, err := client.AddProtocol([]byte("nfe"), []byte(""))
		if err == nil {
			t.Error("expected error for empty protocol XML")
		}
	})

	// Test with valid inputs
	t.Run("valid inputs", func(t *testing.T) {
		nfeXML := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe35200114200166000187550010000000046362100024" versao="4.00">
    <ide>
      <cUF>35</cUF>
      <cNF>62100024</cNF>
      <natOp>Venda de mercadoria</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>4</nNF>
      <dhEmi>2020-01-01T10:00:00-03:00</dhEmi>
      <tpNF>1</tpNF>
      <idDest>1</idDest>
      <cMunFG>3550308</cMunFG>
      <tpImp>1</tpImp>
      <tpEmis>1</tpEmis>
      <cDV>4</cDV>
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
      <enderEmit>
        <xLgr>Rua teste</xLgr>
        <nro>123</nro>
        <xBairro>Centro</xBairro>
        <cMun>3550308</cMun>
        <xMun>São Paulo</xMun>
        <UF>SP</UF>
        <CEP>01000000</CEP>
      </enderEmit>
      <IE>123456789</IE>
      <CRT>3</CRT>
    </emit>
    <det nItem="1">
      <prod>
        <cProd>001</cProd>
        <cEAN>SEM GTIN</cEAN>
        <xProd>Produto Teste</xProd>
        <NCM>12345678</NCM>
        <CFOP>5102</CFOP>
        <uCom>UN</uCom>
        <qCom>1.0000</qCom>
        <vUnCom>100.00</vUnCom>
        <vProd>100.00</vProd>
        <cEANTrib>SEM GTIN</cEANTrib>
        <uTrib>UN</uTrib>
        <qTrib>1.0000</qTrib>
        <vUnTrib>100.00</vUnTrib>
        <indTot>1</indTot>
      </prod>
      <imposto>
        <ICMS>
          <ICMS00>
            <orig>0</orig>
            <CST>00</CST>
            <modBC>3</modBC>
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
  </infNFe>
</NFe>`

		protocolXML := `<?xml version="1.0" encoding="UTF-8"?>
<protNFe versao="4.00">
  <infProt>
    <tpAmb>2</tpAmb>
    <verAplic>SP_NFE_PL_008i2</verAplic>
    <chNFe>35200114200166000187550010000000046362100024</chNFe>
    <dhRecbto>2020-01-01T10:01:00-03:00</dhRecbto>
    <nProt>135200000000001</nProt>
    <digVal>abc123</digVal>
    <cStat>100</cStat>
    <xMotivo>Autorizado o uso da NF-e</xMotivo>
  </infProt>
</protNFe>`

		procXML, err := client.AddProtocol([]byte(nfeXML), []byte(protocolXML))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(procXML) == 0 {
			t.Fatal("procNFe XML should not be empty")
		}

		// Check if result contains both NFe and protocol
		procString := string(procXML)
		if !strings.Contains(procString, "nfeProc") {
			t.Error("result should contain nfeProc element")
		}

		if !strings.Contains(procString, "Empresa Teste LTDA") {
			t.Error("result should contain NFe data")
		}

		if !strings.Contains(procString, "135200000000001") {
			t.Error("result should contain protocol data")
		}

		if !strings.Contains(procString, "Autorizado o uso da NF-e") {
			t.Error("result should contain protocol message")
		}
	})
}
