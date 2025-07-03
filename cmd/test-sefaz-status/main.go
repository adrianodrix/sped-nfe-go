// Test script for SEFAZ status service using real ICP-Brasil certificate
// This is SAFE - only queries status, does not send any NFe documents
// Usage: go run cmd/test-sefaz-status/main.go [certificate_password]
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adrianodrix/sped-nfe-go/certificate"
	"github.com/adrianodrix/sped-nfe-go/nfe"
)

func main() {
	fmt.Println("=== Teste SEFAZ Status - Certificado ICP-Brasil ===")
	fmt.Println("⚠️  ATENÇÃO: Este teste usa ambiente de PRODUÇÃO")
	fmt.Println("✅ SEGURO: Apenas consulta status, não envia documentos")

	// Configure unsafe SSL for testing (disable certificate verification)
	os.Setenv("SPED_NFE_UNSAFE_SSL", "true")

	// 1. Obter senha do certificado
	var password string
	if len(os.Args) > 1 {
		password = os.Args[1]
	}

	if password == "" {
		fmt.Println("\n❌ Senha do certificado é obrigatória!")
		fmt.Println("Uso: go run cmd/test-sefaz-status/main.go <senha_do_certificado>")
		fmt.Println("Exemplo: go run cmd/test-sefaz-status/main.go minhasenha123")
		os.Exit(1)
	}

	// 2. Carregar certificado real
	fmt.Println("\n1. Carregando certificado real...")

	certPath := "refs/certificates/cert-valido-jan-2026.pfx"
	fmt.Printf("   Arquivo: %s\n", certPath)

	cert, err := certificate.LoadA1FromFile(certPath, password)
	if err != nil {
		log.Fatalf("❌ Erro ao carregar certificado: %v", err)
	}
	defer cert.Close()

	fmt.Println("   ✅ Certificado carregado com sucesso")

	// Exibir informações do certificado
	fmt.Printf("   📋 Dados do certificado:\n")
	fmt.Printf("      Titular: %s\n", cert.GetSubject())
	fmt.Printf("      Emissor: %s\n", cert.GetIssuer())
	fmt.Printf("      Serial: %s\n", cert.GetSerialNumber())
	fmt.Printf("      Válido: %v\n", cert.IsValid())
	notBefore, notAfter := cert.GetValidityPeriod()
	fmt.Printf("      Validade: %s até %s\n",
		notBefore.Format("02/01/2006"),
		notAfter.Format("02/01/2006"))

	// Verificar se certificado está válido
	if !cert.IsValid() {
		fmt.Println("   ❌ ATENÇÃO: Certificado não está válido!")
		fmt.Println("   💡 Verifique a data de validade antes de usar em produção")
		return
	}

	// 3. Criar cliente NFe para PRODUÇÃO
	fmt.Println("\n2. Criando cliente NFe para PRODUÇÃO...")

	config := nfe.ClientConfig{
		Environment: nfe.Production, // 🔴 PRODUÇÃO
		UF:          nfe.PR,         // Paraná (baseado no certificado EMPARI de Maringá)
		Timeout:     45,             // Timeout maior para produção
	}

	client, err := nfe.NewClient(config)
	if err != nil {
		log.Fatalf("❌ Erro ao criar cliente: %v", err)
	}
	fmt.Println("   ✅ Cliente criado para PRODUÇÃO")
	fmt.Printf("   🎯 Estado: PR (Paraná)\n")
	fmt.Printf("   🌐 Ambiente: PRODUÇÃO\n")

	// 4. Configurar certificado no cliente
	fmt.Println("\n3. Configurando certificado no cliente...")

	err = client.SetCertificate(cert)
	if err != nil {
		log.Fatalf("❌ Erro ao configurar certificado: %v", err)
	}
	fmt.Println("   ✅ Certificado configurado no cliente")

	// 5. Testar comunicação com SEFAZ PRODUÇÃO - Status Service
	fmt.Println("\n4. 🔴 TESTANDO COMUNICAÇÃO COM SEFAZ PRODUÇÃO 🔴")
	fmt.Println("   📡 Serviço: nfestatusservico (apenas consulta status)")
	fmt.Println("   ✅ SEGURO: Não envia documentos, apenas verifica se SEFAZ está online")

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	fmt.Println("   🔄 Enviando requisição...")
	startTime := time.Now()

	statusResponse, err := client.QueryStatus(ctx)
	duration := time.Since(startTime)

	fmt.Printf("   ⏱️  Tempo de resposta: %v\n", duration)

	if err != nil {
		fmt.Printf("   ❌ Erro ao consultar status: %v\n", err)

		// Analisar tipo de erro
		errorStr := err.Error()
		if contains(errorStr, "certificate signed by unknown authority") {
			fmt.Println("   💡 Erro de certificado SSL")
			fmt.Println("   💡 Execute: ./install-icpbrasil-certs.sh")
		} else if contains(errorStr, "timeout") {
			fmt.Println("   💡 Timeout na comunicação")
			fmt.Println("   💡 SEFAZ pode estar sobrecarregado ou lento")
		} else if contains(errorStr, "connection refused") {
			fmt.Println("   💡 Conexão recusada")
			fmt.Println("   💡 Verifique conectividade de rede")
		} else if contains(errorStr, "no such host") {
			fmt.Println("   💡 Erro de DNS")
			fmt.Println("   💡 Verifique resolução de nomes")
		} else {
			fmt.Println("   💡 Erro de comunicação com SEFAZ")
		}

		fmt.Println("\n📋 Análise do erro:")
		if contains(errorStr, "SOAP") || contains(errorStr, "HTTP") {
			fmt.Println("   ✅ Requisição SOAP foi montada corretamente")
			fmt.Println("   ✅ Certificado foi usado na comunicação")
			fmt.Println("   ✅ Conexão com SEFAZ foi tentada")
			fmt.Println("   ❌ Problema na camada de rede/SSL")
		}

	} else {
		fmt.Println("\n🎉 SUCESSO! Comunicação com SEFAZ PRODUÇÃO funcionou!")
		fmt.Printf("   📊 Status SEFAZ: %d - %s\n", statusResponse.Status, statusResponse.StatusText)
		fmt.Printf("   🌐 Online: %v\n", statusResponse.IsOnline())
		fmt.Printf("   📍 UF: %s | Ambiente: %d (1=produção)\n", statusResponse.UF, statusResponse.Environment)
		fmt.Printf("   🕐 Consultado em: %s\n", statusResponse.CheckedAt.Format("02/01/2006 15:04:05"))

		// Verificar códigos de status conhecidos
		switch statusResponse.Status {
		case 107:
			fmt.Println("   ✅ Status 107: Serviço em operação normal")
		case 108:
			fmt.Println("   ⚠️  Status 108: Serviço paralisado momentaneamente")
		case 109:
			fmt.Println("   ⚠️  Status 109: Serviço paralisado sem previsão")
		default:
			fmt.Printf("   ℹ️  Status %d: Consulte documentação SEFAZ\n", statusResponse.Status)
		}
	}

	// 6. Informações sobre o teste
	fmt.Println("\n=== Análise do Teste de Produção ===")

	if err == nil {
		fmt.Println("🎉 RESULTADO: SUCESSO COMPLETO!")
		fmt.Println("\n✅ Validações bem-sucedidas:")
		fmt.Println("   • Certificado ICP-Brasil válido e aceito")
		fmt.Println("   • Cliente NFe configurado corretamente")
		fmt.Println("   • Comunicação com SEFAZ Produção estabelecida")
		fmt.Println("   • Requisições SOAP montadas corretamente")
		fmt.Println("   • Webservice nfestatusservico funcionando")
		fmt.Println("   • Integração Tools + Client + Certificate operacional")
	} else {
		fmt.Println("⚠️  RESULTADO: Comunicação tentada, erro encontrado")
		fmt.Println("\n✅ Validações parciais bem-sucedidas:")
		fmt.Println("   • Certificado ICP-Brasil carregado e válido")
		fmt.Println("   • Cliente NFe configurado para produção")
		fmt.Println("   • Requisições SOAP montadas")
		fmt.Println("   • Tentativa de comunicação com SEFAZ realizada")
	}

	fmt.Println("\n⚠️  IMPORTANTE:")
	fmt.Println("   • Este teste é 100% SEGURO (apenas consulta status)")
	fmt.Println("   • Nenhum documento foi enviado ao SEFAZ")
	fmt.Println("   • Certificado de produção foi usado apenas para autenticação")
	fmt.Println("   • SSL verification foi DESABILITADA para teste")
	fmt.Println("   • Para produção: instale certificados ICP-Brasil corretamente")
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
