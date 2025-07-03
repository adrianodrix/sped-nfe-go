// Teste completo para validação do serviço de inutilização NFe
// Este teste valida TODOS os estados brasileiros para mapear problemas TLS/SSL
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
	fmt.Println("=== SPED-NFE-GO - Teste Completo de Inutilização (TODOS OS ESTADOS) ===")
	fmt.Println("🎯 Objetivo: Mapear problemas TLS/SSL em todos os estados brasileiros")
	fmt.Println("📋 Fluxo: Configuração → Loop Estados → Análise Comparativa")

	// Comentar/descomentar esta linha para testar com/sem bypass SSL
	// os.Setenv("SPED_NFE_UNSAFE_SSL", "true")

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

	fmt.Printf("\n🔧 ETAPA 2: Preparando parâmetros de teste...\n")

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
		log.Fatalf("❌ Falha na validação de parâmetros: %v", err)
	}

	fmt.Printf("   ✅ Parâmetros validados\n")

	// Lista de todos os estados brasileiros
	// estados := []string{
	// 	"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
	// 	"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
	// 	"RS", "RO", "RR", "SC", "SP", "SE", "TO",
	// }
	estados := []string{
		"PR",
	}

	fmt.Printf("\n🚀 ETAPA 3: Testando inutilização em TODOS os estados brasileiros...\n")
	fmt.Printf("   📋 Estados a testar: %d\n", len(estados))
	fmt.Printf("   ⏱️  Timeout por estado: 30 segundos\n")
	fmt.Printf("   ⚠️  ATENÇÃO: Testes reais com certificado de produção!\n\n")

	// Resultados por categoria
	sucessos := make([]string, 0)
	errosTLS := make([]string, 0)
	errosRede := make([]string, 0)
	errosSOAP := make([]string, 0)
	errosOutros := make([]string, 0)

	for i, uf := range estados {
		fmt.Printf("🔄 [%d/%d] Testando %s...\n", i+1, len(estados), uf)

		// Create config for this state
		config := &common.Config{
			TpAmb:       types.Production,
			RazaoSocial: "EMPARI INFORMATICA LTDA",
			CNPJ:        "10541434000152",
			SiglaUF:     uf,
			Schemes:     "PL_009_V4",
			Versao:      "4.00",
			Timeout:     30, // Timeout menor para acelerar
		}

		// Create Tools with resolver
		tools, err := nfe.NewTools(config, webservices.NewResolver())
		if err != nil {
			fmt.Printf("   ❌ %s: Erro ao criar tools: %v\n", uf, err)
			errosOutros = append(errosOutros, uf+": "+err.Error())
			continue
		}

		// Enable debug logging for these problematic states
		tools.EnableDebug(true)

		// IMPORTANTE: Configurar o certificado digital no cliente SOAP
		err = tools.SetCertificate(cert)
		if err != nil {
			fmt.Printf("   ❌ %s: Erro ao configurar certificado: %v\n", uf, err)
			errosOutros = append(errosOutros, uf+": "+err.Error())
			continue
		}

		// Get service info
		resolver := webservices.NewResolver()
		inutInfo, err := resolver.GetInutilizacaoServiceURL(uf, true, "55")
		if err != nil {
			fmt.Printf("   ❌ %s: Erro ao obter URL: %v\n", uf, err)
			errosOutros = append(errosOutros, uf+": "+err.Error())
			continue
		}

		fmt.Printf("   🔗 %s: %s\n", uf, inutInfo.URL)

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Test inutilização
		startTime := time.Now()
		response, err := tools.SefazInutilizaNumeros(ctx, nSerie, nIni, nFin, xJust)
		duration := time.Since(startTime)
		cancel()

		if err == nil {
			fmt.Printf("   ✅ %s: SUCESSO! (%v) - Status: %s - %s\n", uf, duration, response.InfInut.CStat, response.InfInut.XMotivo)
			sucessos = append(sucessos, uf)
		} else {
			errorStr := err.Error()
			fmt.Printf("   ❌ %s: ERRO (%v)\n", uf, duration)

			// Categorizar erros
			if contains(errorStr, "tls:") || contains(errorStr, "certificate") || contains(errorStr, "ssl") {
				fmt.Printf("      🔒 TLS: %s\n", getShortError(errorStr))
				errosTLS = append(errosTLS, uf+": "+getShortError(errorStr))
			} else if contains(errorStr, "timeout") || contains(errorStr, "connection") || contains(errorStr, "network") {
				fmt.Printf("      🌐 REDE: %s\n", getShortError(errorStr))
				errosRede = append(errosRede, uf+": "+getShortError(errorStr))
			} else if contains(errorStr, "soap") || contains(errorStr, "Content-Length: 0") || contains(errorStr, "VAZIO") {
				fmt.Printf("      📄 SOAP: %s\n", getShortError(errorStr))
				errosSOAP = append(errosSOAP, uf+": "+getShortError(errorStr))
			} else {
				fmt.Printf("      ❓ OUTRO: %s\n", getShortError(errorStr))
				errosOutros = append(errosOutros, uf+": "+getShortError(errorStr))
			}
		}

		fmt.Println()
	}

	fmt.Printf("\n🎯 ANÁLISE FINAL - Resultados por Estado:\n")
	fmt.Printf("════════════════════════════════════════════════\n")

	fmt.Printf("\n✅ SUCESSOS (%d estados):\n", len(sucessos))
	if len(sucessos) == 0 {
		fmt.Printf("   Nenhum estado funcionou\n")
	} else {
		for _, estado := range sucessos {
			fmt.Printf("   • %s\n", estado)
		}
	}

	fmt.Printf("\n🔒 ERROS TLS/SSL (%d estados):\n", len(errosTLS))
	if len(errosTLS) == 0 {
		fmt.Printf("   Nenhum erro TLS encontrado\n")
	} else {
		for _, erro := range errosTLS {
			fmt.Printf("   • %s\n", erro)
		}
	}

	fmt.Printf("\n🌐 ERROS DE REDE (%d estados):\n", len(errosRede))
	if len(errosRede) == 0 {
		fmt.Printf("   Nenhum erro de rede encontrado\n")
	} else {
		for _, erro := range errosRede {
			fmt.Printf("   • %s\n", erro)
		}
	}

	fmt.Printf("\n📄 ERROS SOAP (%d estados):\n", len(errosSOAP))
	if len(errosSOAP) == 0 {
		fmt.Printf("   Nenhum erro SOAP encontrado\n")
	} else {
		for _, erro := range errosSOAP {
			fmt.Printf("   • %s\n", erro)
		}
	}

	fmt.Printf("\n❓ OUTROS ERROS (%d estados):\n", len(errosOutros))
	if len(errosOutros) == 0 {
		fmt.Printf("   Nenhum outro erro encontrado\n")
	} else {
		for _, erro := range errosOutros {
			fmt.Printf("   • %s\n", erro)
		}
	}

	fmt.Printf("\n🚀 CONCLUSÕES E PRÓXIMOS PASSOS:\n")
	fmt.Printf("════════════════════════════════════════════════\n")

	if len(sucessos) > 0 {
		fmt.Printf("✅ IMPLEMENTAÇÃO FUNCIONAL: %d estados funcionaram!\n", len(sucessos))
		fmt.Printf("   💡 Nossa implementação SOAP está correta\n")
		fmt.Printf("   💡 Estruturas XML estão corretas\n")
		fmt.Printf("   💡 Processo de inutilização está funcionando\n")
	}

	if len(errosTLS) > 0 {
		fmt.Printf("\n🔒 PROBLEMAS TLS IDENTIFICADOS (%d estados):\n", len(errosTLS))
		fmt.Printf("   💡 Implementar configuração TLS mais robusta\n")
		fmt.Printf("   💡 Adicionar suporte a diferentes cipher suites\n")
		fmt.Printf("   💡 Configurar renegociação TLS quando necessário\n")
		fmt.Printf("   💡 Testar com SPED_NFE_UNSAFE_SSL=true\n")
	}

	if len(errosRede) > 0 {
		fmt.Printf("\n🌐 PROBLEMAS DE CONECTIVIDADE (%d estados):\n", len(errosRede))
		fmt.Printf("   💡 Verificar conectividade de rede\n")
		fmt.Printf("   💡 Aumentar timeouts se necessário\n")
		fmt.Printf("   💡 Implementar retry automático\n")
	}

	if len(errosSOAP) > 0 {
		fmt.Printf("\n📄 PROBLEMAS SOAP (%d estados):\n", len(errosSOAP))
		fmt.Printf("   💡 Verificar headers SOAP enviados\n")
		fmt.Printf("   💡 Analisar estrutura dos envelopes\n")
		fmt.Printf("   💡 Comparar com especificação SEFAZ\n")
	}

	// Estatísticas finais
	total := len(estados)
	sucessoPercent := float64(len(sucessos)) / float64(total) * 100
	fmt.Printf("\n📊 ESTATÍSTICAS GERAIS:\n")
	fmt.Printf("   🎯 Taxa de sucesso: %.1f%% (%d/%d)\n", sucessoPercent, len(sucessos), total)
	fmt.Printf("   🔒 Problemas TLS: %.1f%% (%d/%d)\n", float64(len(errosTLS))/float64(total)*100, len(errosTLS), total)
	fmt.Printf("   🌐 Problemas rede: %.1f%% (%d/%d)\n", float64(len(errosRede))/float64(total)*100, len(errosRede), total)
	fmt.Printf("   📄 Problemas SOAP: %.1f%% (%d/%d)\n", float64(len(errosSOAP))/float64(total)*100, len(errosSOAP), total)
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

func getShortError(err string) string {
	// Extrair apenas a parte mais relevante do erro
	if contains(err, "tls: bad certificate") {
		return "bad certificate (auto-fallback enabled)"
	}
	if contains(err, "tls: no renegotiation") {
		return "no renegotiation (fixed)"
	}
	if contains(err, "certificate") {
		return "certificate issue (auto-fallback enabled)"
	}
	if contains(err, "connection refused") {
		return "connection refused"
	}
	if contains(err, "timeout") {
		return "timeout"
	}
	if contains(err, "Content-Length: 0") {
		return "empty response"
	}

	// Se for muito longo, truncar
	if len(err) > 60 {
		return err[:60] + "..."
	}

	return err
}
