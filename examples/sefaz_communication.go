// Package main demonstrates basic SEFAZ communication using sped-nfe-go library.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/nfe"
	"github.com/adrianodrix/sped-nfe-go/types"
)

func main() {
	// Example of SEFAZ communication setup and status check
	fmt.Println("=== SPED NFe Go - SEFAZ Communication Example ===")

	// 1. Create configuration
	config := &common.Config{
		TpAmb:       types.Homologation, // Use homologation environment
		RazaoSocial: "Empresa Exemplo LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "SP",
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	fmt.Printf("Configuration created for: %s\n", config.RazaoSocial)
	fmt.Printf("Environment: %s\n", config.TpAmb.String())
	fmt.Printf("State: %s\n", config.SiglaUF)

	// 2. Create Tools instance
	tools, err := nfe.NewTools(config)
	if err != nil {
		log.Fatalf("Failed to create tools: %v", err)
	}

	fmt.Println("✓ Tools instance created successfully")

	// 3. Enable debug logging
	tools.EnableDebug(true)
	fmt.Println("✓ Debug logging enabled")

	// 4. Set timeout
	tools.SetTimeout(60 * time.Second)
	fmt.Println("✓ Timeout set to 60 seconds")

	// 5. Check SEFAZ service status
	fmt.Println("\n--- Checking SEFAZ Service Status ---")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Note: This would normally make a real HTTP request to SEFAZ
	// In this example, it will fail because we're not in a real environment
	// but it demonstrates the usage pattern
	statusResp, err := tools.SefazStatus(ctx)
	if err != nil {
		fmt.Printf("⚠️  SEFAZ Status check failed (expected in example): %v\n", err)
		fmt.Println("   This is normal - no real SEFAZ connection in this example")
	} else {
		fmt.Printf("✓ SEFAZ Status: %s - %s\n", statusResp.CStat, statusResp.XMotivo)
		fmt.Printf("  Application Version: %s\n", statusResp.VerAplic)
		fmt.Printf("  Response Time: %s\n", statusResp.TMedResposta)
	}

	// 6. Demonstrate access key consultation structure
	fmt.Println("\n--- Access Key Consultation Example ---")
	
	// This is a sample access key format (44 digits)
	sampleAccessKey := "35230712345678000195550010000000123123456789"
	
	fmt.Printf("Sample access key: %s\n", sampleAccessKey)
	fmt.Printf("Length: %d digits\n", len(sampleAccessKey))
	
	// This would fail with network error in example environment
	_, err = tools.SefazConsultaChave(ctx, sampleAccessKey)
	if err != nil {
		fmt.Printf("⚠️  Access key consultation failed (expected): %v\n", err)
		fmt.Println("   This demonstrates parameter validation")
	}

	// 7. Demonstrate registry consultation structure
	fmt.Println("\n--- Registry Consultation Example ---")
	
	// This would also fail with network error in example environment
	_, err = tools.SefazConsultaCadastro(ctx, config.CNPJ, config.SiglaUF)
	if err != nil {
		fmt.Printf("⚠️  Registry consultation failed (expected): %v\n", err)
		fmt.Println("   This demonstrates the API structure")
	}

	// 8. Demonstrate batch processing structure  
	fmt.Println("\n--- Batch Processing Example ---")
	
	// Create a sample batch structure
	lote := &nfe.LoteNFe{
		IdLote: fmt.Sprintf("%d", time.Now().Unix()),
		NFes:   make([]nfe.NFe, 0), // Empty for demo
	}
	
	fmt.Printf("Sample batch ID: %s\n", lote.IdLote)
	fmt.Printf("NFe count: %d\n", len(lote.NFes))
	
	// This would fail because the batch is empty
	_, err = tools.SefazEnviaLote(ctx, lote, false)
	if err != nil {
		fmt.Printf("⚠️  Batch submission failed (expected): %v\n", err)
		fmt.Println("   Empty batch for demonstration purposes")
	}

	// 9. Demonstrate invalidation structure
	fmt.Println("\n--- Number Invalidation Example ---")
	
	inutilizacao := &nfe.InutilizacaoRequest{
		InfInut: nfe.InfInut{
			XServ:  "INUTILIZAR",
			CUF:    "35", // São Paulo
			Ano:    "23", // Year 2023
			CNPJ:   config.CNPJ,
			Mod:    "55", // NFe model
			Serie:  "1",
			NNFIni: "1",
			NNFFin: "10",
			XJust:  "Teste de inutilização para demonstração do sistema",
		},
	}
	
	fmt.Printf("Invalidation range: %s to %s\n", inutilizacao.InfInut.NNFIni, inutilizacao.InfInut.NNFFin)
	fmt.Printf("Justification: %s\n", inutilizacao.InfInut.XJust)
	
	_, err = tools.SefazInutiliza(ctx, inutilizacao)
	if err != nil {
		fmt.Printf("⚠️  Number invalidation failed (expected): %v\n", err)
	}

	// 10. Demonstrate cancellation structure
	fmt.Println("\n--- Cancellation Example ---")
	
	// Sample cancellation parameters
	chaveParaCancelar := "35230712345678000195550010000000123123456789"
	protocoloCancelamento := "135230000000123"
	justificativaCancelamento := "Cancelamento para demonstração do sistema de comunicação SEFAZ"
	
	fmt.Printf("Access key to cancel: %s\n", chaveParaCancelar)
	fmt.Printf("Protocol: %s\n", protocoloCancelamento)
	fmt.Printf("Justification: %s\n", justificativaCancelamento)
	
	_, err = tools.SefazCancela(ctx, chaveParaCancelar, protocoloCancelamento, justificativaCancelamento)
	if err != nil {
		fmt.Printf("⚠️  Cancellation failed (expected): %v\n", err)
	}

	// 11. Demonstrate correction letter structure
	fmt.Println("\n--- Correction Letter Example ---")
	
	chaveParaCorrigir := "35230712345678000195550010000000123123456789"
	textoCorrecao := "Correção do endereço do destinatário: Rua das Flores, 123, Bairro Centro"
	sequenciaCorrecao := 1
	
	fmt.Printf("Access key to correct: %s\n", chaveParaCorrigir)
	fmt.Printf("Correction text: %s\n", textoCorrecao)
	fmt.Printf("Sequence: %d\n", sequenciaCorrecao)
	
	_, err = tools.SefazCCe(ctx, chaveParaCorrigir, textoCorrecao, sequenciaCorrecao)
	if err != nil {
		fmt.Printf("⚠️  Correction letter failed (expected): %v\n", err)
	}

	// 12. Show request/response debugging
	fmt.Println("\n--- Debug Information ---")
	
	lastRequest := tools.GetLastRequest()
	lastResponse := tools.GetLastResponse()
	
	if lastRequest != "" {
		fmt.Printf("Last request size: %d bytes\n", len(lastRequest))
		fmt.Printf("Request preview: %s...\n", lastRequest[:min(100, len(lastRequest))])
	} else {
		fmt.Println("No requests made yet")
	}
	
	if lastResponse != "" {
		fmt.Printf("Last response size: %d bytes\n", len(lastResponse))
		fmt.Printf("Response preview: %s...\n", lastResponse[:min(100, len(lastResponse))])
	} else {
		fmt.Println("No responses received yet")
	}

	// 13. Configuration validation
	fmt.Println("\n--- Configuration Validation ---")
	
	err = tools.ValidateConfig()
	if err != nil {
		fmt.Printf("⚠️  Configuration validation failed: %v\n", err)
	} else {
		fmt.Println("✓ Configuration is valid")
	}

	// 14. Model switching
	fmt.Println("\n--- Model Switching Example ---")
	
	currentModel := tools.GetModel()
	fmt.Printf("Current model: %s (NFe)\n", currentModel)
	
	err = tools.SetModel("65")
	if err != nil {
		fmt.Printf("⚠️  Failed to switch to NFCe: %v\n", err)
	} else {
		fmt.Printf("✓ Switched to model: %s (NFCe)\n", tools.GetModel())
	}
	
	// Switch back to NFe
	err = tools.SetModel("55")
	if err != nil {
		fmt.Printf("⚠️  Failed to switch back to NFe: %v\n", err)
	} else {
		fmt.Printf("✓ Switched back to model: %s (NFe)\n", tools.GetModel())
	}

	fmt.Println("\n=== Example completed ===")
	fmt.Println("Note: Network errors are expected in this demo environment.")
	fmt.Println("In a real implementation, you would:")
	fmt.Println("1. Have proper network connectivity to SEFAZ")
	fmt.Println("2. Load a valid digital certificate")
	fmt.Println("3. Create properly structured NFe documents")
	fmt.Println("4. Handle responses according to business logic")
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}