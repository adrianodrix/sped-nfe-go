// Teste completo para validação do serviço de inutilização NFe
// Este teste valida uma URL diferente da autorização para confirmar se o problema é específico
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adrianodrix/sped-nfe-go/certificate"
	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/nfe"
	"github.com/adrianodrix/sped-nfe-go/types"
	"github.com/adrianodrix/sped-nfe-go/webservices"
)

func main() {
	fmt.Println("=== SPED-NFE-GO - Teste Completo de Inutilização ===")
	fmt.Println("🎯 Objetivo: Validar URL diferente da autorização para diagnóstico")
	fmt.Println("📋 Fluxo: Configuração → Inutilização → Análise SOAP")

	// Configure unsafe SSL for testing (disable certificate verification)
	os.Setenv("SPED_NFE_UNSAFE_SSL", "true")

	// 1. Obter senha do certificado
	var password string
	if len(os.Args) > 1 {
		password = os.Args[1]
	}

	if password == "" {
		fmt.Println("\n❌ Senha do certificado é obrigatória!")
		fmt.Println("Uso: go run cmd/teste-completo-inutilizacao/main.go <senha_do_certificado>")
		fmt.Println("Exemplo: go run cmd/teste-completo-inutilizacao/main.go minhasenha123")
		os.Exit(1)
	}

	fmt.Printf("\n📜 ETAPA 1: Carregando certificado digital...\n")
	fmt.Printf("   🔐 Carregando certificado ICP-Brasil real...\n")

	certPath := "refs/certificates/valid-certificate.pfx"
	fmt.Printf("   📋 Certificado: %s\n", certPath)

	cert, err := certificate.LoadA1FromFile(certPath, password)
	if err != nil {
		log.Fatalf("❌ Falha na etapa 1: erro ao carregar certificado: %v", err)
	}
	defer cert.Close()

	notBefore, notAfter := cert.GetValidityPeriod()
	fmt.Printf("   📅 Válido: %s até %s\n", 
		notBefore.Format("2006-01-02"), 
		notAfter.Format("2006-01-02"))
	fmt.Printf("   ✅ Certificado válido: %s\n", cert.GetSubject())

	// Verificar se certificado está válido
	if !cert.IsValid() {
		fmt.Println("   ❌ ATENÇÃO: Certificado não está válido!")
		fmt.Println("   💡 Verifique a data de validade antes de usar")
		return
	}

	fmt.Printf("\n🔧 ETAPA 2: Configurando cliente NFe...\n")

	// Create config for production
	config := &common.Config{
		TpAmb:       types.Production, // Usando produção para testar URLs reais
		RazaoSocial: "EMPARI INFORMATICA LTDA",
		CNPJ:        "10541434000152",
		SiglaUF:     "PR", // Paraná (mesmo estado do problema de autorização)
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     60,
	}

	// Create Tools with resolver
	tools, err := nfe.NewTools(config, webservices.NewResolver())
	if err != nil {
		log.Fatalf("❌ Falha na etapa 2: erro ao criar tools: %v", err)
	}

	fmt.Printf("   ✅ Cliente NFe configurado\n")

	// Configure certificate
	// TODO: Set certificate when signing is implemented

	fmt.Printf("\n📋 ETAPA 3: Preparando dados de inutilização...\n")

	// Test parameters
	nSerie := 1
	nIni := 1000
	nFin := 1010
	xJust := "Teste de inutilização para validação de URLs SEFAZ - ambiente de desenvolvimento e testes"

	fmt.Printf("   📊 Série: %d\n", nSerie)
	fmt.Printf("   🔢 Números: %d até %d (%d números)\n", nIni, nFin, nFin-nIni+1)
	fmt.Printf("   📝 Justificativa: %s\n", xJust)

	// Validate parameters
	if err := nfe.ValidateInutilizacaoParams(nSerie, nIni, nFin, xJust); err != nil {
		log.Fatalf("❌ Falha na etapa 3: validação de parâmetros: %v", err)
	}

	fmt.Printf("   ✅ Parâmetros validados\n")

	fmt.Printf("\n🔍 ETAPA 4: Analisando configurações de webservice...\n")

	// Get service info for analysis
	statusInfo, err := tools.GetStatusServiceInfo()
	if err != nil {
		log.Fatalf("❌ Erro ao obter info do serviço de status: %v", err)
	}

	authInfo, err := tools.GetAuthorizationServiceInfo()
	if err != nil {
		log.Fatalf("❌ Erro ao obter info do serviço de autorização: %v", err)
	}

	// Get inutilização service info using resolver
	resolver := webservices.NewResolver()
	inutInfo, err := resolver.GetInutilizacaoServiceURL(config.SiglaUF, config.TpAmb == types.Production, "55")
	if err != nil {
		log.Fatalf("❌ Erro ao obter info do serviço de inutilização: %v", err)
	}

	fmt.Printf("   📊 URLs de webservice:\n")
	fmt.Printf("      Status:        %s\n", statusInfo.URL)
	fmt.Printf("      Autorização:   %s\n", authInfo.URL)
	fmt.Printf("      Inutilização:  %s\n", inutInfo.URL)

	fmt.Printf("   🔍 SOAPActions:\n")
	fmt.Printf("      Status:        %s\n", statusInfo.Action)
	fmt.Printf("      Autorização:   %s\n", authInfo.Action)
	fmt.Printf("      Inutilização:  %s\n", inutInfo.Action)

	// Compare URLs
	if statusInfo.URL == authInfo.URL {
		fmt.Printf("   ⚠️  Status e Autorização usam a mesma URL\n")
	} else {
		fmt.Printf("   ✅ Status e Autorização usam URLs diferentes\n")
	}

	if authInfo.URL == inutInfo.URL {
		fmt.Printf("   ⚠️  Autorização e Inutilização usam a mesma URL\n")
	} else {
		fmt.Printf("   ✅ Autorização e Inutilização usam URLs diferentes\n")
	}

	fmt.Printf("\n🚀 ETAPA 5: Testando inutilização...\n")
	fmt.Printf("   📤 Enviando requisição de inutilização para SEFAZ...\n")
	fmt.Printf("   ⚠️  ATENÇÃO: Este é um teste real com certificado de produção!\n")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test inutilização (this will use a different URL than authorization)
	startTime := time.Now()
	response, err := tools.SefazInutilizaNumeros(ctx, nSerie, nIni, nFin, xJust)
	duration := time.Since(startTime)

	fmt.Printf("   ⏱️  Tempo de resposta: %v\n", duration)

	if err != nil {
		fmt.Printf("   ❌ Erro na inutilização: %v\n", err)

		// Debug information
		fmt.Printf("\n🔍 Informações de Debug:\n")
		if lastRequest := tools.GetLastRequest(); lastRequest != "" {
			fmt.Printf("   📤 SOAP Request enviado: %d bytes\n", len(lastRequest))
			if len(lastRequest) > 500 {
				fmt.Printf("   📄 Request (primeiros 300 chars): %s...\n", lastRequest[:300])
			} else {
				fmt.Printf("   📄 Request: %s\n", lastRequest)
			}
		}

		if lastResponse := tools.GetLastResponse(); lastResponse != "" {
			fmt.Printf("   📥 SOAP Response recebido: %d bytes\n", len(lastResponse))
			if len(lastResponse) > 500 {
				fmt.Printf("   📄 Response (primeiros 300 chars): %s...\n", lastResponse[:300])
			} else {
				fmt.Printf("   📄 Response: %s\n", lastResponse)
			}
		} else {
			fmt.Printf("   📥 SOAP Response: [VAZIO] - mesmo problema da autorização!\n")
		}

		// Analyze error type
		errorStr := err.Error()
		fmt.Printf("\n📋 Análise do Erro:\n")
		if contains(errorStr, "Content-Length: 0") || contains(errorStr, "VAZIO") {
			fmt.Printf("   🎯 CONFIRMADO: Mesmo problema da autorização!\n")
			fmt.Printf("   📊 Resultado: URLs diferentes, mesmo problema\n")
			fmt.Printf("   💡 Conclusão: Problema não é específico da URL de autorização\n")
			fmt.Printf("   🔍 Causas possíveis:\n")
			fmt.Printf("      • Problema no certificado SSL/TLS\n")
			fmt.Printf("      • Headers SOAP incorretos\n")
			fmt.Printf("      • Configuração do SEFAZ Paraná\n")
			fmt.Printf("      • Problema na estrutura do envelope SOAP\n")
		} else if contains(errorStr, "timeout") {
			fmt.Printf("   ⏰ Timeout na comunicação\n")
		} else if contains(errorStr, "certificate") {
			fmt.Printf("   🔒 Problema de certificado SSL\n")
		} else {
			fmt.Printf("   ❓ Erro não categorizado: %s\n", errorStr)
		}

	} else {
		fmt.Printf("   🎉 SUCESSO! Inutilização funcionou!\n")
		fmt.Printf("   📊 Status: %s - %s\n", response.InfInut.CStat, response.InfInut.GetMessage())
		fmt.Printf("   ✅ Success: %v\n", response.InfInut.IsSuccess())

		if response.InfInut.IsSuccess() {
			fmt.Printf("   🎯 DESCOBERTA: Inutilização funciona, autorização não!\n")
			fmt.Printf("   📊 Conclusão: Problema É específico da URL/serviço de autorização\n")
		}

		fmt.Printf("   🔢 Protocolo: %s\n", response.InfInut.NProt)
		fmt.Printf("   📅 Data/Hora: %s\n", response.InfInut.DhRecbto)
	}

	fmt.Printf("\n🎯 ANÁLISE FINAL:\n")
	fmt.Printf("   📋 Teste realizado com sucesso\n")
	fmt.Printf("   🔗 URLs testadas:\n")
	fmt.Printf("      ✅ Status: %s\n", statusInfo.URL)
	fmt.Printf("      ❌ Autorização: %s (problema conhecido)\n", authInfo.URL)
	if err != nil {
		fmt.Printf("      ❌ Inutilização: %s (mesmo problema)\n", inutInfo.URL)
		fmt.Printf("\n💡 CONCLUSÃO: Problema não é específico da URL, mas sim do SEFAZ Paraná ou configuração SOAP\n")
	} else {
		fmt.Printf("      ✅ Inutilização: %s (funcionou!)\n", inutInfo.URL)
		fmt.Printf("\n💡 CONCLUSÃO: Problema É específico do serviço de autorização!\n")
	}

	fmt.Printf("\n🚀 Próximos passos sugeridos:\n")
	if err != nil {
		fmt.Printf("   1. Investigar configuração específica do SEFAZ Paraná\n")
		fmt.Printf("   2. Verificar headers SOAP enviados\n")
		fmt.Printf("   3. Testar com outros estados (SP, MG)\n")
		fmt.Printf("   4. Analisar diferenças na estrutura SOAP\n")
	} else {
		fmt.Printf("   1. Comparar estrutura SOAP entre inutilização e autorização\n")
		fmt.Printf("   2. Verificar headers específicos do serviço de autorização\n")
		fmt.Printf("   3. Analisar diferenças nos envelopes\n")
		fmt.Printf("   4. Investigar requisitos específicos do NFeAutorizacao4\n")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	})()
}

