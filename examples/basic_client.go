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
	// Configurar cliente NFe
	config := nfe.ClientConfig{
		Environment: nfe.Homologation, // Use Production para produção
		UF:          nfe.SP,           // Estado do emitente
		Timeout:     30,               // Timeout em segundos
	}

	// Criar cliente
	client, err := nfe.NewClient(config)
	if err != nil {
		log.Fatalf("Erro ao criar cliente NFe: %v", err)
	}

	// Configurar certificado digital (exemplo com mock)
	cert := certificate.NewMockCertificate()
	err = client.SetCertificate(cert)
	if err != nil {
		log.Fatalf("Erro ao configurar certificado: %v", err)
	}

	// Exemplo 1: Criar uma nova NFe
	fmt.Println("=== Criando nova NFe ===")
	make := client.CreateNFe()
	make.SetVersion("4.00")

	// Adicionar informações básicas da NFe
	identificacao := &nfe.Identificacao{
		CUF:      "35",       // São Paulo
		CNF:      "12345678", // Código numérico fiscal
		NatOp:    "Venda",    // Natureza da operação
		Mod:      "55",       // Modelo NFe
		Serie:    "1",        // Série
		NNF:      "123",      // Número da NFe
		DhEmi:    time.Now().Format("2006-01-02T15:04:05-07:00"),
		TpNF:     "1",       // Saída
		IdDest:   "1",       // Operação interna
		CMunFG:   "3550308", // São Paulo
		TpImp:    "1",       // DANFE normal
		TpEmis:   "1",       // Normal
		CDV:      "1",       // Dígito verificador
		TpAmb:    "2",       // Homologação
		FinNFe:   "1",       // Normal
		IndFinal: "0",       // Não consumidor final
		IndPres:  "1",       // Presencial
		ProcEmi:  "0",       // Aplicativo do contribuinte
		VerProc:  "1.0.0",   // Versão do processo
	}

	err = make.TagIde(identificacao)
	if err != nil {
		log.Fatalf("Erro ao adicionar identificação: %v", err)
	}

	// Adicionar emitente
	emitente := &nfe.Emitente{
		CNPJ:  "12345678000195",
		XNome: "Empresa Exemplo LTDA",
		EnderEmit: nfe.Endereco{
			XLgr:    "Rua das Flores",
			Nro:     "123",
			XBairro: "Centro",
			CMun:    "3550308",
			XMun:    "São Paulo",
			UF:      "SP",
			CEP:     "01234567",
		},
		IE:  "123456789",
		CRT: "1", // Simples Nacional
	}

	err = make.TagEmit(emitente)
	if err != nil {
		log.Fatalf("Erro ao adicionar emitente: %v", err)
	}

	// Gerar XML da NFe
	xml, err := make.GetXML()
	if err != nil {
		log.Fatalf("Erro ao gerar XML: %v", err)
	}

	fmt.Printf("XML gerado com sucesso! Tamanho: %d bytes\n", len(xml))

	// Exemplo 2: Autorizar a NFe
	fmt.Println("\n=== Autorizando NFe ===")
	ctx := context.Background()

	authResponse, err := client.Authorize(ctx, []byte(xml))
	if err != nil {
		log.Fatalf("Erro ao autorizar NFe: %v", err)
	}

	if authResponse.Authorized() {
		fmt.Printf("✅ NFe autorizada com sucesso!\n")
		fmt.Printf("   Status: %d - %s\n", authResponse.Status, authResponse.StatusText)
		if authResponse.HasReceipt() {
			fmt.Printf("   Recibo: %s\n", authResponse.Receipt)
		}
	} else {
		fmt.Printf("❌ NFe não autorizada\n")
		fmt.Printf("   Status: %d - %s\n", authResponse.Status, authResponse.StatusText)
	}

	// Exemplo 3: Consultar status do SEFAZ
	fmt.Println("\n=== Consultando status SEFAZ ===")

	statusResponse, err := client.QueryStatus(ctx)
	if err != nil {
		log.Fatalf("Erro ao consultar status: %v", err)
	}

	if statusResponse.IsOnline() {
		fmt.Printf("✅ SEFAZ online\n")
		fmt.Printf("   Status: %d - %s\n", statusResponse.Status, statusResponse.StatusText)
		fmt.Printf("   UF: %s | Ambiente: %d\n", statusResponse.UF, statusResponse.Environment)
	} else {
		fmt.Printf("❌ SEFAZ offline\n")
	}

	// Exemplo 4: Gerar chave de acesso
	fmt.Println("\n=== Gerando chave de acesso ===")

	chave, err := client.GenerateKey(
		"12345678000195", // CNPJ
		55,               // Modelo NFe
		1,                // Série
		123,              // Número
		time.Now(),       // Data de emissão
	)
	if err != nil {
		log.Printf("Erro ao gerar chave: %v", err)
	} else {
		fmt.Printf("Chave gerada: %s\n", chave)
	}

	// Exemplo 5: Ativar contingência
	fmt.Println("\n=== Testando contingência ===")

	err = client.ActivateContingency("SEFAZ fora do ar para manutenção programada")
	if err != nil {
		log.Printf("Erro ao ativar contingência: %v", err)
	} else {
		fmt.Printf("✅ Contingência ativada\n")

		if client.IsContingencyActive() {
			fmt.Printf("   Status: Ativa\n")

			// Desativar contingência
			err = client.DeactivateContingency()
			if err != nil {
				log.Printf("Erro ao desativar contingência: %v", err)
			} else {
				fmt.Printf("✅ Contingência desativada\n")
			}
		}
	}

	// Exemplo 6: Validar XML
	fmt.Println("\n=== Validando XML ===")

	err = client.ValidateXML([]byte(xml))
	if err != nil {
		fmt.Printf("❌ XML inválido: %v\n", err)
	} else {
		fmt.Printf("✅ XML válido\n")
	}

	fmt.Println("\n=== Exemplo concluído ===")
	fmt.Println("Este exemplo demonstra as funcionalidades básicas da API do sped-nfe-go")
	fmt.Println("Para usar em produção:")
	fmt.Println("1. Configure Environment = nfe.Production")
	fmt.Println("2. Use um certificado digital real (A1 ou A3)")
	fmt.Println("3. Implemente tratamento de erro robusto")
	fmt.Println("4. Configure logs adequados")
}
