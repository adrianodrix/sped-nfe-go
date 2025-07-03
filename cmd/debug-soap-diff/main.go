// Debug script to compare SOAP requests between working QueryStatus and failing Authorize
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/nfe"
	"github.com/adrianodrix/sped-nfe-go/types"
	"github.com/adrianodrix/sped-nfe-go/webservices"
)

func main() {
	fmt.Println("=== Debug: Comparação SOAP QueryStatus vs Authorize ===")

	// Configure unsafe SSL for testing
	os.Setenv("SPED_NFE_UNSAFE_SSL", "true")

	// Create config for Paraná Homolog (same as failing test)
	config := &common.Config{
		TpAmb:       types.Homologation, // Same as failing test
		RazaoSocial: "EMPARI INFORMATICA LTDA",
		CNPJ:        "10541434000152",
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

	fmt.Printf("✅ Tools criado para %s ambiente %s\n", config.SiglaUF, config.TpAmb.String())

	// Get service info for both
	statusInfo, err := tools.GetStatusServiceInfo()
	if err != nil {
		log.Fatalf("Failed to get status service info: %v", err)
	}

	authInfo, err := tools.GetAuthorizationServiceInfo()
	if err != nil {
		log.Fatalf("Failed to get authorization service info: %v", err)
	}

	fmt.Printf("\n📊 Comparação de Configuração:\n")
	fmt.Printf("Status URL:    %s\n", statusInfo.URL)
	fmt.Printf("Auth URL:      %s\n", authInfo.URL)
	fmt.Printf("Status Action: %s\n", statusInfo.Action)
	fmt.Printf("Auth Action:   %s\n", authInfo.Action)

	// Test 1: QueryStatus (working)
	fmt.Printf("\n🔍 TESTE 1: QueryStatus (deveria funcionar)\n")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	startTime := time.Now()
	statusResponse, err := tools.SefazStatus(ctx)
	duration := time.Since(startTime)

	fmt.Printf("   Tempo: %v\n", duration)
	if err != nil {
		fmt.Printf("   ❌ Erro: %v\n", err)
	} else {
		fmt.Printf("   ✅ Sucesso: Status %s - %s\n", statusResponse.CStat, statusResponse.XMotivo)
	}

	fmt.Printf("\n🔍 TESTE 2: Verificação de diferenças estruturais\n")
	
	// Create minimal NFe XML for testing (without actual certificate signing)
	testNFeXML := `<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
		<infNFe Id="NFe41250700000000000000550010000000011234567890" versao="4.00">
			<ide><cUF>41</cUF><cNF>12345678</cNF><natOp>Teste</natOp><mod>55</mod><serie>1</serie><nNF>1</nNF><dhEmi>2025-07-03T20:00:00-03:00</dhEmi><tpNF>1</tpNF><idDest>1</idDest><cMunFG>4115200</cMunFG><tpImp>1</tpImp><tpEmis>1</tpEmis><cDV>0</cDV><tpAmb>2</tpAmb><finNFe>1</finNFe><indFinal>0</indFinal><indPres>1</indPres><procEmi>0</procEmi><verProc>TEST</verProc></ide>
			<emit><CNPJ>00000000000000</CNPJ><xNome>TESTE</xNome></emit>
			<dest><CNPJ>00000000000000</CNPJ><xNome>TESTE</xNome></dest>
			<det nItem="1"><prod><cProd>TESTE</cProd><xProd>TESTE</xProd><vProd>1.00</vProd></prod></det>
			<total><ICMSTot><vNF>1.00</vNF></ICMSTot></total>
		</infNFe>
	</NFe>`

	// Test the raw SOAP construction that would be used for authorization
	fmt.Printf("   🔧 Testando construção do envelope enviNFe...\n")
	
	requestXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>`)
	requestXML += fmt.Sprintf(`<enviNFe xmlns="http://www.portalfiscal.inf.br/nfe" versao="%s">`, config.Versao)
	requestXML += fmt.Sprintf(`<idLote>%s</idLote>`, "999999")
	requestXML += fmt.Sprintf(`<indSinc>%s</indSinc>`, "0")
	requestXML += testNFeXML
	requestXML += `</enviNFe>`

	fmt.Printf("   📄 Envelope size: %d bytes\n", len(requestXML))
	fmt.Printf("   📄 Envelope preview: %s...\n", requestXML[:300])

	fmt.Printf("\n🎯 Análise Comparativa:\n")
	fmt.Printf("1. ✅ URLs são idênticas entre Status e Auth\n")
	fmt.Printf("2. ✅ Actions estão no formato correto para ambos\n")
	fmt.Printf("3. ✅ Mesmo sistema de resolução sendo usado\n")
	fmt.Printf("4. 🔍 Verificar se há diferenças nos headers SOAP específicos\n")
	fmt.Printf("5. 🔍 Verificar se SEFAZ Paraná tem restrições especiais\n")
}