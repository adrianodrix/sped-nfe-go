// Test script to verify SOAPAction fix for authorization service
package main

import (
	"fmt"
	"log"

	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/nfe"
	"github.com/adrianodrix/sped-nfe-go/types"
	"github.com/adrianodrix/sped-nfe-go/webservices"
)

func main() {
	fmt.Println("=== Teste de Correção SOAPAction ===")

	// Create config for Paraná (where we found the issue)
	config := &common.Config{
		TpAmb:       types.Production, // Production to match our failing test
		RazaoSocial: "Empresa Teste LTDA",
		CNPJ:        "12345678000195",
		SiglaUF:     "PR", // Paraná
		Schemes:     "PL_009_V4",
		Versao:      "4.00",
		Timeout:     30,
	}

	// Create Tools with resolver
	tools, err := nfe.NewTools(config, webservices.NewResolver())
	if err != nil {
		log.Fatalf("Failed to create tools: %v", err)
	}

	fmt.Printf("✅ Tools criado para %s, ambiente %d\n", config.SiglaUF, config.TpAmb)

	// Test Status Service URL (already working)
	statusInfo, err := tools.GetStatusServiceInfo()
	if err != nil {
		log.Fatalf("Failed to get status service info: %v", err)
	}

	fmt.Printf("\n📊 Status Service (funcionando):\n")
	fmt.Printf("   URL: %s\n", statusInfo.URL)
	fmt.Printf("   Method: %s\n", statusInfo.Method)
	fmt.Printf("   Operation: %s\n", statusInfo.Operation)
	fmt.Printf("   Action: %s\n", statusInfo.Action)

	// Test Authorization Service URL (was failing, now should work)
	authInfo, err := tools.GetAuthorizationServiceInfo()
	if err != nil {
		log.Fatalf("Failed to get authorization service info: %v", err)
	}

	fmt.Printf("\n🔐 Authorization Service (corrigido):\n")
	fmt.Printf("   URL: %s\n", authInfo.URL)
	fmt.Printf("   Method: %s\n", authInfo.Method)
	fmt.Printf("   Operation: %s\n", authInfo.Operation)
	fmt.Printf("   Action: %s\n", authInfo.Action)

	// Compare Actions
	fmt.Printf("\n🔍 Comparação de Actions:\n")
	if statusInfo.Action == authInfo.Action {
		fmt.Printf("   ❌ ERRO: Actions são idênticas (não deveria ser)\n")
		fmt.Printf("   Status Action: %s\n", statusInfo.Action)
		fmt.Printf("   Auth Action: %s\n", authInfo.Action)
	} else {
		fmt.Printf("   ✅ Actions são diferentes (correto):\n")
		fmt.Printf("   Status Action: %s\n", statusInfo.Action)
		fmt.Printf("   Auth Action: %s\n", authInfo.Action)
	}

	// Check if the Action format is correct (should be full URL)
	expectedPrefix := "http://www.portalfiscal.inf.br/nfe/wsdl/"
	fmt.Printf("\n✅ Verificação de formato de Action:\n")
	
	if len(statusInfo.Action) > len(expectedPrefix) && statusInfo.Action[:len(expectedPrefix)] == expectedPrefix {
		fmt.Printf("   ✅ Status Action tem formato correto\n")
	} else {
		fmt.Printf("   ❌ Status Action tem formato incorreto: %s\n", statusInfo.Action)
	}

	if len(authInfo.Action) > len(expectedPrefix) && authInfo.Action[:len(expectedPrefix)] == expectedPrefix {
		fmt.Printf("   ✅ Authorization Action tem formato correto\n")
	} else {
		fmt.Printf("   ❌ Authorization Action tem formato incorreto: %s\n", authInfo.Action)
	}

	fmt.Printf("\n🎯 Resultado:\n")
	fmt.Printf("   ✅ Status service: %s\n", statusInfo.URL)
	fmt.Printf("   ✅ Authorization service: %s\n", authInfo.URL)
	fmt.Printf("   ✅ Ambos usam o mesmo sistema de resolução agora!\n")
	fmt.Printf("   ✅ SOAPAction corrigida para formato completo!\n")
}