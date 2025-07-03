// Package main demonstrates the usage of complementary NFe features (simplified version)
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
	fmt.Printf("✓ QR Code configurado (versão %s)\n", qrCode.Version)

	// URL de consulta para diferentes estados
	estados := []string{"SP", "RJ", "PR", "RS"}
	fmt.Println("✓ URLs de consulta por estado:")
	for _, uf := range estados {
		urlProd := factories.GetStateConsultationURL(uf, 1)
		urlHomolog := factories.GetStateConsultationURL(uf, 2)
		fmt.Printf("  - %s: Produção=%s\n", uf, safeSubstring(urlProd, 40)+"...")
		fmt.Printf("     Homologação=%s\n", safeSubstring(urlHomolog, 40)+"...")
	}

	// Demonstrar builder pattern para QR Code
	fmt.Println("\n• Usando QR Code Builder Pattern:")
	builder := factories.NewQRCodeBuilder(qrCode)
	qrURL, err := builder.
		ChaveNFe("41230714200166000187650010000000051123456789").
		URL("https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx").
		Environment("2").
		EmissionDateTime("2023-12-25T15:30:00-03:00").
		TotalValue("100.00").
		ICMSValue("18.00").
		DigestValue("ABC123DEF456").
		Token("TOKEN_SEFAZ", "000001").
		Build()

	if err == nil {
		fmt.Printf("✓ QR Code construído com sucesso (%d caracteres)\n", len(qrURL))
		fmt.Printf("  URL: %s...\n", safeSubstring(qrURL, 80))
	} else {
		fmt.Printf("✗ Erro ao construir QR Code: %v\n", err)
	}

	// Validar QR Code
	if err := qrCode.ValidateQRCode(qrURL); err == nil {
		fmt.Println("✓ QR Code válido")
	}

	fmt.Println()
}

func demonstrateParser() {
	fmt.Println("2. PARSER DE TXT PARA XML")
	fmt.Println("==========================")

	// Dados NFe em formato TXT (simplificado)
	txtData := `NOTAFISCAL|1|
A|4.00|NFe41230714200166000187650010000000051123456789||`

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
		if len(fmt.Sprintf("%v", infNFe["Id"])) > 10 {
			fmt.Printf("  - ID: %v...\n", fmt.Sprintf("%v", infNFe["Id"])[:30])
		}
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
	} else {
		fmt.Printf("✗ Erro na conversão: %v\n", err)
	}

	fmt.Println()
}

func demonstrateContingency() {
	fmt.Println("3. SISTEMA DE CONTINGÊNCIA")
	fmt.Println("===========================")

	// Verificar tipo de contingência padrão para estados
	fmt.Println("• Tipos de contingência por estado:")
	estados := []string{"SP", "RJ", "RS", "BA", "PR", "MG", "GO", "MT"}
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
	fmt.Printf("  - Motivo: %s\n", contingency.Motive[:50]+"...")
	fmt.Printf("  - Status: %t\n", contingency.IsActive())

	// Obter informações para inclusão no XML
	fmt.Println("\n• Informações para XML:")
	xmlInfo := contingency.GetContingencyInfo()
	for key, value := range xmlInfo {
		if str, ok := value.(string); ok && len(str) > 50 {
			fmt.Printf("  - %s: %s...\n", key, str[:50])
		} else {
			fmt.Printf("  - %s: %v\n", key, value)
		}
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
	} else {
		fmt.Printf("✗ Erro ao ativar: %v\n", err)
	}

	// Demonstrar função de conveniência
	fmt.Println("\n• Usando função de conveniência:")
	c3, _, err := factories.CreateContingency("MG", "SEFAZ MG em manutenção programada")
	if err == nil {
		fmt.Printf("✓ Contingência %s criada para MG\n", c3.Type)
	} else {
		fmt.Printf("✗ Erro ao criar: %v\n", err)
	}

	// Desativar contingência
	fmt.Println("\n• Desativando contingência:")
	deactivatedJSON, err := contingency.Deactivate()
	if err == nil {
		fmt.Printf("✓ Contingência desativada\n")
		fmt.Printf("  - Status ativo: %t\n", contingency.IsActive())
		fmt.Printf("  - TpEmis: %d\n", contingency.TpEmis)
	} else {
		fmt.Printf("✗ Erro ao desativar: %v\n", err)
	}

	// Validar dados de contingência
	fmt.Println("\n• Validando dados de contingência:")
	if err := factories.ValidateContingencyData(jsonData); err == nil {
		fmt.Println("✓ Dados de contingência válidos")
	} else {
		fmt.Printf("✗ Dados inválidos: %v\n", err)
	}

	if err := factories.ValidateContingencyData(deactivatedJSON); err == nil {
		fmt.Println("✓ Dados de desativação válidos")
	} else {
		fmt.Printf("✗ Dados inválidos: %v\n", err)
	}

	// Mostrar JSONs de forma segura
	fmt.Printf("\n✓ JSON de contingência (%d chars): %s...\n",
		len(jsonData), safeSubstring(jsonData, 60))
	fmt.Printf("✓ JSON de desativação (%d chars): %s...\n",
		len(deactivatedJSON), safeSubstring(deactivatedJSON, 60))

	fmt.Println()
}

// safeSubstring safely extracts a substring without panicking
func safeSubstring(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func init() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
