// Package main demonstrates the usage of complementary NFe features
package main

import (
	"fmt"
	"log"

	"github.com/adrianodrix/sped-nfe-go/factories"
)

func main() {
	fmt.Println("=== Demonstração das Funcionalidades Complementares NFe ===\n")

	// 1. Demonstrar geração de QR Code para NFCe
	demonstrateQRCode()

	// 2. Demonstrar parser de TXT para XML
	demonstrateParser()

	// 3. Demonstrar sistema de contingência
	demonstrateContingency()
}

func demonstrateQRCode() {
	fmt.Println("1. GERAÇÃO DE QR CODE PARA NFCe")
	fmt.Println("=====================================")

	// Criar QR Code para NFCe versão 2.00 (mais comum)
	qrConfig := factories.QRCodeConfig{
		Version: "2.00",
		CSC:     "CODIGO_SEGURANCA_CONTRIBUINTE_123456789",
		CSCId:   "000001",
	}
	
	qrCode := factories.NewQRCode(qrConfig)

	// XML de NFCe de exemplo (simplificado)
	xmlNFCe := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
	<infNFe Id="NFe41230714200166000187650010000000051123456789" versao="4.00">
		<ide>
			<cUF>41</cUF>
			<cNF>12345678</cNF>
			<natOp>Venda</natOp>
			<mod>65</mod>
			<serie>001</serie>
			<nNF>5</nNF>
			<dhEmi>2023-12-25T15:30:00-03:00</dhEmi>
			<tpNF>1</tpNF>
			<idDest>1</idDest>
			<cMunFG>4106902</cMunFG>
			<tpImp>4</tpImp>
			<tpEmis>1</tpEmis>
			<cDV>9</cDV>
			<tpAmb>2</tpAmb>
			<finNFe>1</finNFe>
			<indFinal>1</indFinal>
			<indPres>1</indPres>
		</ide>
		<dest>
			<CNPJ>12345678000195</CNPJ>
			<xNome>Cliente Teste NFCe</xNome>
		</dest>
		<det nItem="1">
			<prod>
				<cProd>PROD001</cProd>
				<xProd>Produto de Teste</xProd>
				<uCom>UN</uCom>
				<qCom>1.0000</qCom>
				<vUnCom>100.00</vUnCom>
				<vProd>100.00</vProd>
			</prod>
		</det>
		<total>
			<ICMSTot>
				<vBC>100.00</vBC>
				<vICMS>18.00</vICMS>
				<vNF>100.00</vNF>
			</ICMSTot>
		</total>
		<Signature xmlns="http://www.w3.org/2000/09/xmldsig#">
			<SignedInfo>
				<Reference>
					<DigestValue>ABC123DEF456</DigestValue>
				</Reference>
			</SignedInfo>
		</Signature>
	</infNFe>
</NFe>`

	// Gerar QR Code e inserir no XML
	urlConsulta := factories.GetStateConsultationURL("PR", 2) // Paraná, homologação
	xmlWithQR, err := qrCode.PutQRTag([]byte(xmlNFCe), "TOKEN_SEFAZ", "000001", "4.00", urlConsulta, "https://consulta.sefaz.pr.gov.br")
	
	if err != nil {
		log.Printf("Erro ao gerar QR Code: %v", err)
	} else {
		fmt.Printf("✓ QR Code versão %s gerado com sucesso!\n", qrCode.Version)
		fmt.Printf("✓ URL de consulta: %s\n", urlConsulta)
		
		// Extrair QR Code do XML para visualização
		qrCodeURL, err := qrCode.GetQRCodeFromXML(string(xmlWithQR))
		if err == nil {
			fmt.Printf("✓ QR Code URL: %s\n", qrCodeURL[:100]+"...")
		}
	}

	// Demonstrar builder pattern para QR Code
	fmt.Println("\n• Usando QR Code Builder Pattern:")
	builder := factories.NewQRCodeBuilder(qrCode)
	qrURL, err := builder.
		ChaveNFe("41230714200166000187650010000000051123456789").
		URL(urlConsulta).
		Environment("2").
		EmissionDateTime("2023-12-25T15:30:00-03:00").
		TotalValue("100.00").
		ICMSValue("18.00").
		DigestValue("ABC123DEF456").
		Token("TOKEN_SEFAZ", "000001").
		Build()

	if err == nil {
		fmt.Printf("✓ QR Code construído: %s\n", qrURL[:80]+"...")
	}

	fmt.Println()
}

func demonstrateParser() {
	fmt.Println("2. PARSER DE TXT PARA XML")
	fmt.Println("==========================")

	// Dados NFe em formato TXT
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||
B|41|00000005|Venda de produtos|55|001|5|||1|1|4106902|1|1|5|2|1|1|0|Sistema Teste|1.0|||
C|Empresa Teste LTDA|||||||
C02|14200166000187|
E|Cliente Teste|||||||
E02|12345678000195|
I|1||
I02|PROD001||Produto Teste|||||UN|1|100.00|100.00||UN|1|100.00|||||1||||
M|100.00|18.00||||||||100.00||||||||||100.00|||`

	// Criar parser
	parser, err := factories.NewParser(factories.ParserConfig{
		Version: "4.00",
		Layout:  factories.LayoutLocal,
	})
	
	if err != nil {
		log.Printf("Erro ao criar parser: %v", err)
		return
	}

	fmt.Println("✓ Parser NFe 4.00 criado com sucesso")

	// Validar TXT antes do parsing
	errors := parser.ValidateTXT(txtData)
	if len(errors) > 0 {
		fmt.Printf("⚠ Erros de validação encontrados: %d\n", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
	} else {
		fmt.Println("✓ TXT validado com sucesso")
	}

	// Fazer parsing
	data, err := parser.ParseTXT(txtData)
	if err != nil {
		log.Printf("Erro no parsing: %v", err)
		return
	}

	fmt.Printf("✓ TXT convertido para estrutura de dados (%d seções)\n", len(data))
	
	// Mostrar algumas informações extraídas
	if infNFe, ok := data["infNFe"].(map[string]interface{}); ok {
		fmt.Printf("  - Versão NFe: %v\n", infNFe["versao"])
		fmt.Printf("  - ID: %v\n", infNFe["Id"])
	}
	
	if ide, ok := data["ide"].(map[string]interface{}); ok {
		fmt.Printf("  - UF: %v\n", ide["cUF"])
		fmt.Printf("  - Modelo: %v\n", ide["mod"])
		fmt.Printf("  - Série: %v\n", ide["serie"])
	}

	// Gerar XML
	xml, err := parser.GetXML()
	if err != nil {
		log.Printf("Erro ao gerar XML: %v", err)
	} else {
		fmt.Printf("✓ XML gerado com sucesso (%d caracteres)\n", len(xml))
	}

	// Demonstrar função de conveniência
	fmt.Println("\n• Usando função de conveniência:")
	xmlDirect, err := factories.ConvertTXTToXML(txtData, "4.00", factories.LayoutLocal)
	if err == nil {
		fmt.Printf("✓ Conversão direta TXT→XML realizada (%d caracteres)\n", len(xmlDirect))
	}

	fmt.Println()
}

func demonstrateContingency() {
	fmt.Println("3. SISTEMA DE CONTINGÊNCIA")
	fmt.Println("===========================")

	// Verificar tipo de contingência padrão para estados
	fmt.Println("• Tipos de contingência por estado:")
	estados := []string{"SP", "RJ", "RS", "BA", "PR", "MG"}
	for _, uf := range estados {
		if contingencyType, err := factories.GetStateContingencyType(uf); err == nil {
			fmt.Printf("  - %s: %s\n", uf, contingencyType)
		}
	}

	// Criar contingência para SP
	fmt.Println("\n• Ativando contingência para São Paulo:")
	contingency, err := factories.NewContingency()
	if err != nil {
		log.Printf("Erro ao criar contingência: %v", err)
		return
	}

	config := factories.ContingencyConfig{
		UF:     "SP",
		Motive: "SEFAZ SP fora do ar devido a problemas técnicos de infraestrutura",
	}

	jsonData, err := contingency.Activate(config)
	if err != nil {
		log.Printf("Erro ao ativar contingência: %v", err)
		return
	}

	fmt.Printf("✓ Contingência ativada para %s\n", config.UF)
	fmt.Printf("  - Tipo: %s\n", contingency.Type)
	fmt.Printf("  - TpEmis: %d\n", contingency.TpEmis)
	fmt.Printf("  - Data/Hora: %s\n", contingency.GetFormattedDateTime())
	fmt.Printf("  - Motivo: %s\n", contingency.Motive)
	fmt.Printf("  - Status: %t\n", contingency.IsActive())

	// Obter informações para inclusão no XML
	fmt.Println("\n• Informações para XML:")
	xmlInfo := contingency.GetContingencyInfo()
	for key, value := range xmlInfo {
		fmt.Printf("  - %s: %v\n", key, value)
	}

	// Demonstrar builder pattern
	fmt.Println("\n• Usando Contingency Builder Pattern:")
	c2, _, err := factories.NewContingencyBuilder().
		ForState("RS").
		WithMotive("SEFAZ RS com problemas de conectividade").
		WithType(factories.ContingencySVCRS).
		Activate()

	if err == nil {
		fmt.Printf("✓ Contingência %s ativada para RS\n", c2.Type)
	}

	// Demonstrar função de conveniência
	fmt.Println("\n• Usando função de conveniência:")
	c3, _, err := factories.CreateContingency("MG", "SEFAZ MG em manutenção programada")
	if err == nil {
		fmt.Printf("✓ Contingência %s criada para MG\n", c3.Type)
	}

	// Desativar contingência
	fmt.Println("\n• Desativando contingência:")
	deactivatedJSON, err := contingency.Deactivate()
	if err == nil {
		fmt.Printf("✓ Contingência desativada\n")
		fmt.Printf("  - Status ativo: %t\n", contingency.IsActive())
		fmt.Printf("  - TpEmis: %d\n", contingency.TpEmis)
	}

	// Validar dados de contingência
	fmt.Println("\n• Validando dados de contingência:")
	if err := factories.ValidateContingencyData(jsonData); err == nil {
		fmt.Println("✓ Dados de contingência válidos")
	}

	if err := factories.ValidateContingencyData(deactivatedJSON); err == nil {
		fmt.Println("✓ Dados de desativação válidos")
	}

	fmt.Printf("\n✓ JSON de contingência: %s\n", jsonData[:100]+"...")
	fmt.Printf("✓ JSON de desativação: %s\n", deactivatedJSON[:80]+"...")

	fmt.Println()
}

func init() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}