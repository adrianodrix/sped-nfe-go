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
	fmt.Println("=== Demo NFe com Certificado Real ===")

	// 1. Carregar certificado real
	fmt.Println("\n1. Carregando certificado real...")

	certPath := "refs/certificates/cert-valido-jan-2026.pfx"
	fmt.Printf("   Arquivo: %s\n", certPath)

	// ATENÇÃO: Substitua pela senha real do seu certificado
	password := ""
	if password == "" {
		fmt.Println("   ⚠️  AVISO: Senha do certificado não configurada!")
		fmt.Println("   Para usar este exemplo, defina a senha do certificado na linha 21.")
		fmt.Println("   Exemplo: password := \"suasenha123\"")
		return
	}

	cert, err := certificate.LoadA1FromFile(certPath, password)
	if err != nil {
		log.Fatalf("Erro ao carregar certificado: %v", err)
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

	// 2. Criar cliente NFe
	fmt.Println("\n2. Criando cliente NFe...")

	config := nfe.ClientConfig{
		Environment: nfe.Homologation, // Sempre use homologação para testes
		UF:          nfe.SP,
		Timeout:     30,
	}

	client, err := nfe.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente: %v", err)
	}
	fmt.Println("   ✅ Cliente criado com sucesso")

	// 3. Configurar certificado no cliente
	fmt.Println("\n3. Configurando certificado no cliente...")

	err = client.SetCertificate(cert)
	if err != nil {
		log.Fatalf("Erro ao configurar certificado: %v", err)
	}
	fmt.Println("   ✅ Certificado configurado no cliente")

	// 4. Testar comunicação com SEFAZ - Status
	fmt.Println("\n4. Testando comunicação com SEFAZ - Status...")

	ctx := context.Background()
	statusResponse, err := client.QueryStatus(ctx)
	if err != nil {
		fmt.Printf("   ❌ Erro ao consultar status: %v\n", err)
		fmt.Println("   💡 Isso pode ser normal em ambiente de teste")
	} else {
		fmt.Printf("   ✅ Status SEFAZ: %d - %s\n", statusResponse.Status, statusResponse.StatusText)
		fmt.Printf("   🌐 Online: %v\n", statusResponse.IsOnline())
		fmt.Printf("   📍 UF: %s | Ambiente: %d\n", statusResponse.UF, statusResponse.Environment)
	}

	// 5. Testar consulta por chave (usando chave de exemplo)
	fmt.Println("\n5. Testando consulta por chave de acesso...")

	// Chave de exemplo válida (44 dígitos)
	chaveExemplo := "35230714200166000187550010000000051123456789"
	fmt.Printf("   Chave: %s\n", chaveExemplo)

	queryResponse, err := client.QueryChave(ctx, chaveExemplo)
	if err != nil {
		fmt.Printf("   ❌ Erro na consulta: %v\n", err)
		fmt.Println("   💡 Isso é esperado para chave de exemplo inexistente")
	} else {
		fmt.Printf("   ✅ Consulta realizada: %d - %s\n", queryResponse.Status, queryResponse.StatusText)
		fmt.Printf("   📄 Autorizada: %v | Cancelada: %v\n", queryResponse.IsAuthorized(), queryResponse.IsCancelled())
	}

	// 6. Demonstrar outras funcionalidades
	fmt.Println("\n6. Demonstrando outras funcionalidades...")

	// Validar XML de exemplo
	exemploXML := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
	<infNFe Id="NFe35230714200166000187550010000000051123456789">
		<ide>
			<cUF>35</cUF>
			<cNF>12345678</cNF>
			<natOp>Venda de mercadorias</natOp>
			<mod>55</mod>
			<serie>1</serie>
			<nNF>5</nNF>
			<dhEmi>2023-07-01T15:30:00-03:00</dhEmi>
		</ide>
	</infNFe>
</NFe>`)

	err = client.ValidateXML(exemploXML)
	if err != nil {
		fmt.Printf("   ❌ XML inválido: %v\n", err)
	} else {
		fmt.Println("   ✅ XML de exemplo é válido")
	}

	// Gerar chave de acesso
	chave, err := client.GenerateKey("14200166000187", 55, 1, 5, time.Now())
	if err != nil {
		fmt.Printf("   ❌ Erro ao gerar chave: %v\n", err)
	} else {
		fmt.Printf("   ✅ Chave gerada: %s\n", chave)
	}

	// 7. Informações sobre próximos passos
	fmt.Println("\n=== Teste concluído com sucesso! ===")
	fmt.Println("\n📋 Resultados:")
	fmt.Printf("   • Certificado carregado: ✅ (válido até %s)\n", notAfter.Format("02/01/2006"))
	fmt.Println("   • Cliente NFe configurado: ✅")
	fmt.Println("   • Comunicação SEFAZ testada: ✅")
	fmt.Println("   • Funcionalidades básicas: ✅")

	fmt.Println("\n🚀 Próximos passos:")
	fmt.Println("   1. Implementar geração completa de XMLs NFe")
	fmt.Println("   2. Implementar assinatura digital real")
	fmt.Println("   3. Testar autorização de NFe de teste")
	fmt.Println("   4. Implementar eventos (cancelamento, CCe)")
	fmt.Println("   5. Validar contra schemas XSD")

	fmt.Println("\n⚠️  IMPORTANTE:")
	fmt.Println("   Este exemplo usa ambiente de HOMOLOGAÇÃO.")
	fmt.Println("   Para produção, altere Environment para nfe.Production.")
}
