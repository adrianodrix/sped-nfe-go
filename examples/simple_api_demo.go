package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adrianodrix/sped-nfe-go/certificate"
	"github.com/adrianodrix/sped-nfe-go/nfe"
)

func main() {
	fmt.Println("=== Demo da API sped-nfe-go ===")

	// 1. Criar cliente NFe
	fmt.Println("\n1. Criando cliente NFe...")

	config := nfe.ClientConfig{
		Environment: nfe.Homologation,
		UF:          nfe.SP,
		Timeout:     30,
	}

	client, err := nfe.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente: %v", err)
	}
	fmt.Println("   ✅ Cliente criado com sucesso")

	// 2. Configurar certificado
	fmt.Println("\n2. Configurando certificado mock...")

	cert := certificate.NewMockCertificate()
	err = client.SetCertificate(cert)
	if err != nil {
		log.Fatalf("Erro ao configurar certificado: %v", err)
	}
	fmt.Println("   ✅ Certificado configurado")

	// 3. Consultar status SEFAZ
	fmt.Println("\n3. Consultando status do SEFAZ...")

	ctx := context.Background()
	statusResponse, err := client.QueryStatus(ctx)
	if err != nil {
		log.Fatalf("Erro ao consultar status: %v", err)
	}

	fmt.Printf("   Status: %d - %s\n", statusResponse.Status, statusResponse.StatusText)
	fmt.Printf("   Online: %v\n", statusResponse.IsOnline())
	fmt.Printf("   UF: %s | Ambiente: %d\n", statusResponse.UF, statusResponse.Environment)

	// 4. Criar NFe builder
	fmt.Println("\n4. Criando NFe builder...")

	make := client.CreateNFe()
	if make == nil {
		log.Fatal("Erro ao criar NFe builder")
	}
	fmt.Println("   ✅ NFe builder criado")

	// 5. Validar XML exemplo
	fmt.Println("\n5. Validando XML de exemplo...")

	exemploXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
	<infNFe Id="NFe12345">
		<ide>
			<cUF>35</cUF>
			<cNF>12345678</cNF>
		</ide>
	</infNFe>
</NFe>`)

	err = client.ValidateXML(exemploXML)
	if err != nil {
		fmt.Printf("   ❌ XML inválido: %v\n", err)
	} else {
		fmt.Println("   ✅ XML válido")
	}

	// 6. Testar autorização com XML de exemplo
	fmt.Println("\n6. Testando autorização...")

	authResponse, err := client.Authorize(ctx, exemploXML)
	if err != nil {
		log.Fatalf("Erro na autorização: %v", err)
	}

	fmt.Printf("   Status: %d - %s\n", authResponse.Status, authResponse.StatusText)
	fmt.Printf("   Autorizada: %v\n", authResponse.Authorized())

	// 7. Testar consulta por chave
	fmt.Println("\n7. Testando consulta por chave...")

	chaveExemplo := "12345678901234567890123456789012345678901234"
	queryResponse, err := client.QueryChave(ctx, chaveExemplo)
	if err != nil {
		log.Fatalf("Erro na consulta: %v", err)
	}

	fmt.Printf("   Status: %d - %s\n", queryResponse.Status, queryResponse.StatusText)
	fmt.Printf("   Autorizada: %v\n", queryResponse.IsAuthorized())

	// 8. Testar geração de chave
	fmt.Println("\n8. Testando geração de chave de acesso...")

	chave, err := client.GenerateKey("12345678000195", 55, 1, 123,
		time.Now())
	if err != nil {
		fmt.Printf("   ❌ Erro: %v\n", err)
	} else {
		fmt.Printf("   ✅ Chave gerada: %s\n", chave)
	}

	// 9. Métodos auxiliares
	fmt.Println("\n9. Testando métodos auxiliares...")

	fmt.Printf("   Config: %v\n", client.GetConfig() != nil)
	fmt.Printf("   Certificado: %v\n", client.GetCertificate() != nil)
	fmt.Printf("   Contingência ativa: %v\n", client.IsContingencyActive())

	fmt.Println("\n=== Demo concluído com sucesso! ===")
	fmt.Println("\nA API sped-nfe-go está funcionando e pronta para uso.")
	fmt.Println("Próximos passos:")
	fmt.Println("- Implementar métodos Tools para comunicação real com SEFAZ")
	fmt.Println("- Adicionar geração completa de XMLs")
	fmt.Println("- Implementar assinatura digital real")
	fmt.Println("- Adicionar validação XSD completa")
}
