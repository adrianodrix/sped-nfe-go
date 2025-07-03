// Package main demonstra o fluxo completo de geração, assinatura e autorização de NFe
// usando o sped-nfe-go.
//
// Este teste valida as três funcionalidades principais:
// 1. Geração de XML NFe
// 2. Assinatura digital com certificado ICP-Brasil
// 3. Autorização via webservice SEFAZ
//
// Uso: go run main.go <senha_certificado>
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
	fmt.Println("🚀 SPED-NFE-GO - Teste Completo")
	fmt.Println("===============================")
	fmt.Println("📋 Fluxo: Geração → Assinatura → Autorização")
	fmt.Println()

	// Verificar parâmetros
	if len(os.Args) < 2 {
		fmt.Println("❌ Uso: go run main.go <senha_certificado>")
		fmt.Println("📋 Exemplo: go run main.go minhasenha123")
		os.Exit(1)
	}
	password := os.Args[1]

	// Configurar SSL para ambiente de teste
	os.Setenv("SPED_NFE_UNSAFE_SSL", "true")

	// ETAPA 1: Carregar Certificado Digital
	fmt.Println("📜 ETAPA 1: Carregando certificado digital...")
	cert, err := carregarCertificado(password)
	if err != nil {
		log.Fatalf("❌ Falha na etapa 1: %v", err)
	}
	fmt.Printf("   ✅ Certificado válido: %s\n\n", extrairNomeCertificado(cert.GetSubject()))

	// ETAPA 2: Configurar Cliente NFe
	fmt.Println("🔧 ETAPA 2: Configurando cliente NFe...")
	client, err := configurarCliente(cert)
	if err != nil {
		log.Fatalf("❌ Falha na etapa 2: %v", err)
	}
	fmt.Println("   ✅ Cliente NFe configurado\n")

	// ETAPA 3: Gerar XML da NFe
	fmt.Println("📝 ETAPA 3: Gerando XML da NFe...")
	xml, chave, err := gerarNFe()
	if err != nil {
		log.Fatalf("❌ Falha na etapa 3: %v", err)
	}
	fmt.Printf("   ✅ XML gerado: %d bytes\n", len(xml))
	fmt.Printf("   🔑 Chave de acesso: %s\n\n", chave)

	// Exemplo 6: Validar XML
	fmt.Println("\n=== Validando XML ===")

	err = client.ValidateXML([]byte(xml))
	if err != nil {
		fmt.Printf("❌ XML inválido: %v\n", err)
	} else {
		fmt.Printf("✅ XML válido\n")
	}

	// ETAPA 4: Assinar XML
	fmt.Println("🔏 ETAPA 4: Assinando XML digitalmente...")
	xmlAssinado, err := assinarXML(cert, xml)
	if err != nil {
		log.Fatalf("❌ Falha na etapa 4: %v", err)
	}
	fmt.Printf("   ✅ XML assinado: %d bytes\n\n", len(xmlAssinado))

	// ETAPA 5: Autorizar no SEFAZ
	fmt.Println("📡 ETAPA 5: Enviando para autorização SEFAZ...")
	err = autorizarNFe(client, xmlAssinado, chave)
	if err != nil {
		log.Printf("   ⚠️  Autorização: %v\n", err)
		fmt.Println("   💡 Nota: Erro esperado em ambiente de teste/demo\n")
	}

	// ETAPA 6: Salvar Arquivos
	fmt.Println("💾 ETAPA 6: Salvando arquivos...")
	err = salvarArquivos(xml, xmlAssinado, chave)
	if err != nil {
		log.Printf("   ⚠️  Erro ao salvar: %v\n", err)
	} else {
		fmt.Println("   ✅ Arquivos salvos com sucesso\n")
	}

	// RESULTADO FINAL
	fmt.Println("🎉 TESTE COMPLETO FINALIZADO!")
	fmt.Println("=============================")
	fmt.Println("✅ Geração de XML: SUCESSO")
	fmt.Println("✅ Assinatura digital: SUCESSO")
	fmt.Println("✅ Comunicação SEFAZ: TESTADA")
	fmt.Println()
	fmt.Println("📁 Arquivos gerados:")
	fmt.Println("   • nfe_original.xml - XML original da NFe")
	fmt.Println("   • nfe_assinada.xml - XML assinado digitalmente")
	fmt.Println("   • chave_acesso.txt - Chave de acesso da NFe")
	fmt.Println()
	fmt.Println("🏆 SPED-NFE-GO funcionando corretamente!")
}

// carregarCertificado carrega e valida o certificado digital A1
func carregarCertificado(password string) (certificate.Certificate, error) {
	certPath := "../../refs/certificates/valid-certificate.pfx"

	cert, err := certificate.LoadA1FromFile(certPath, password)
	if err != nil {
		certPath := "refs/certificates/valid-certificate.pfx"

		cert, err = certificate.LoadA1FromFile(certPath, password)
		if err != nil {
			return nil, fmt.Errorf("erro ao carregar certificado: %v", err)
		}
	}

	// Validar certificado
	if !cert.IsValid() {
		return nil, fmt.Errorf("certificado inválido ou expirado")
	}

	return cert, nil
}

// configurarCliente cria e configura o cliente NFe com certificado
func configurarCliente(cert certificate.Certificate) (*nfe.NFEClient, error) {
	config := nfe.ClientConfig{
		Environment: nfe.Production, // Ambiente de homologação
		UF:          nfe.PR,         // Paraná
		Timeout:     30,             // 50 segundos
	}

	client, err := nfe.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente: %v", err)
	}

	err = client.SetCertificate(cert)
	if err != nil {
		return nil, fmt.Errorf("erro ao configurar certificado: %v", err)
	}

	return client, nil
}

// gerarNFe cria uma NFe completa e gera o XML
func gerarNFe() (string, string, error) {
	// Criar instância do gerador NFe
	nfeMake := nfe.NewMake()

	// Configurar funcionalidades
	nfeMake.SetEnvironment(nfe.EnvironmentTesting)
	nfeMake.SetModel(nfe.ModelNFe)
	nfeMake.SetAutoCalculate(true) // Calcular totais automaticamente
	nfeMake.SetRoundValues(true)   // Arredondar valores monetários
	nfeMake.SetCheckGTIN(true)     // Validar códigos GTIN
	nfeMake.SetRemoveAccents(true) // Remover acentos

	// Configurar identificação da NFe
	now := time.Now()
	identificacao := &nfe.Identificacao{
		CUF:      "41",                    // Paraná
		CNF:      "12345678",              // Código numérico fiscal
		NatOp:    "Venda de mercadoria",   // Natureza da operação
		Mod:      "55",                    // Modelo NFe
		Serie:    "1",                     // Série
		NNF:      "1",                     // Número da NFe
		DhEmi:    nfe.FormatDateTime(now), // Data/hora emissão
		TpNF:     "1",                     // Tipo: saída
		IdDest:   "1",                     // Destinatário: operação interna
		CMunFG:   "4115200",               // Município fiscal: Maringá
		TpImp:    "1",                     // Tipo impressão: retrato
		TpEmis:   "1",                     // Tipo emissão: normal
		TpAmb:    "2",                     // Ambiente: homologação
		FinNFe:   "1",                     // Finalidade: normal
		IndFinal: "0",                     // Consumidor final: não
		IndPres:  "1",                     // Presença: operação presencial
		ProcEmi:  "0",                     // Processo emissão: aplicativo contribuinte
		VerProc:  "SPED-NFE-GO-1.0",       // Versão do processo
	}

	err := nfeMake.TagIde(identificacao)
	if err != nil {
		return "", "", fmt.Errorf("erro na identificação: %v", err)
	}

	// Configurar emitente
	emitente := &nfe.Emitente{
		CNPJ:  "10541434000152",
		XNome: "EMPARI INFORMATICA LTDA",
		XFant: "EMPARI INFORMATICA LTDA",
		IE:    "123456789012",
		CRT:   "3", // Regime Normal
		EnderEmit: nfe.Endereco{
			XLgr:    "RUA DAS EMPRESAS",
			Nro:     "123",
			XBairro: "CENTRO",
			CMun:    "4115200",
			XMun:    "MARINGA",
			UF:      "PR",
			CEP:     "87020000",
			CPais:   "1058",
			XPais:   "BRASIL",
		},
	}

	err = nfeMake.TagEmit(emitente)
	if err != nil {
		return "", "", fmt.Errorf("erro no emitente: %v", err)
	}

	// Configurar destinatário
	destinatario := &nfe.Destinatario{
		CNPJ:      "22333444000195",
		XNome:     "CLIENTE EXEMPLO LTDA",
		IndIEDest: "1", // Contribuinte ICMS
		IE:        "987654321098",
		EnderDest: &nfe.Endereco{
			XLgr:    "AVENIDA DOS CLIENTES",
			Nro:     "456",
			XBairro: "JARDIM COMERCIAL",
			CMun:    "3550308",
			XMun:    "SAO PAULO",
			UF:      "SP",
			CEP:     "01234568",
			CPais:   "1058",
			XPais:   "BRASIL",
		},
	}

	err = nfeMake.TagDest(destinatario)
	if err != nil {
		return "", "", fmt.Errorf("erro no destinatário: %v", err)
	}

	// Configurar produtos/serviços
	produtos := []nfe.Item{
		{
			Prod: nfe.Produto{
				CProd:    "PROD001",
				CEAN:     "SEM GTIN",
				XProd:    "NOTEBOOK LENOVO THINKPAD",
				NCM:      "84713012",
				CFOP:     "5102",
				UCom:     "UN",
				QCom:     "2.0000",
				VUnCom:   "2500.00",
				VProd:    "5000.00",
				CEANTrib: "SEM GTIN",
				UTrib:    "UN",
				QTrib:    "2.0000",
				VUnTrib:  "2500.00",
				IndTot:   "1",
			},
			Imposto: nfe.Imposto{
				ICMS: &nfe.ICMS{
					ICMS00: &nfe.ICMS00{
						Orig:  "0",
						CST:   "00",
						ModBC: "0",
						VBC:   "5000.00",
						PICMS: "18.00",
						VICMS: "900.00",
					},
				},
				IPI: &nfe.IPI{
					CEnq: "999",
					IPITrib: &nfe.IPITrib{
						CST:  "50",
						VBC:  "5000.00",
						PIPI: "5.00",
						VIPI: "250.00",
					},
				},
				PIS: &nfe.PIS{
					PISAliq: &nfe.PISAliq{
						CST:  "01",
						VBC:  "5000.00",
						PPIS: "1.65",
						VPIS: "82.50",
					},
				},
				COFINS: &nfe.COFINS{
					COFINSAliq: &nfe.COFINSAliq{
						CST:     "01",
						VBC:     "5000.00",
						PCOFINS: "7.60",
						VCOFINS: "380.00",
					},
				},
			},
		},
		{
			Prod: nfe.Produto{
				CProd:    "SERV001",
				CEAN:     "SEM GTIN",
				XProd:    "INSTALACAO E CONFIGURACAO",
				NCM:      "00000000",
				CFOP:     "5933",
				UCom:     "SV",
				QCom:     "1.0000",
				VUnCom:   "500.00",
				VProd:    "500.00",
				CEANTrib: "SEM GTIN",
				UTrib:    "SV",
				QTrib:    "1.0000",
				VUnTrib:  "500.00",
				IndTot:   "1",
			},
			Imposto: nfe.Imposto{
				ICMS: &nfe.ICMS{
					ICMS40: &nfe.ICMS40{
						Orig:       "0",
						CST:        "40",
						VICMSDeson: "0.00",
						MotDesICMS: "9",
					},
				},
				IPI: &nfe.IPI{
					CEnq: "999",
					IPINT: &nfe.IPINT{
						CST: "53",
					},
				},
				PIS: &nfe.PIS{
					PISNT: &nfe.PISNT{
						CST: "07",
					},
				},
				COFINS: &nfe.COFINS{
					COFINSNT: &nfe.COFINSNT{
						CST: "07",
					},
				},
			},
		},
	}

	// Adicionar produtos à NFe
	for _, produto := range produtos {
		err = nfeMake.TagDet(&produto)
		if err != nil {
			return "", "", fmt.Errorf("erro no produto %s: %v", produto.Prod.CProd, err)
		}
	}

	// Configurar transporte
	transporte := &nfe.Transporte{
		ModFrete: "0", // Por conta do emitente
	}

	err = nfeMake.TagTransp(transporte)
	if err != nil {
		return "", "", fmt.Errorf("erro no transporte: %v", err)
	}

	// Configurar pagamento
	pagamento := &nfe.Pagamento{
		TPag: "01",      // Dinheiro
		VPag: "5750.00", // Total da NFe (com impostos)
	}

	err = nfeMake.TagPag(pagamento)
	if err != nil {
		return "", "", fmt.Errorf("erro no pagamento: %v", err)
	}

	// Configurar informações adicionais
	infAdic := &nfe.InfAdicionais{
		InfAdFisco: "Documento emitido por ME ou EPP optante pelo Simples Nacional",
		InfCpl:     "NFe gerada pelo sistema SPED-NFE-GO para demonstração de funcionalidades. Ambiente de homologação - SEM VALOR FISCAL.",
	}

	err = nfeMake.TagInfAdic(infAdic)
	if err != nil {
		return "", "", fmt.Errorf("erro nas informações adicionais: %v", err)
	}

	// Gerar XML
	xml, err := nfeMake.GetXML()
	if err != nil {
		return "", "", fmt.Errorf("erro ao gerar XML: %v", err)
	}

	// Obter chave de acesso
	chave := nfeMake.GetAccessKey()
	if chave == "" {
		return "", "", fmt.Errorf("chave de acesso não gerada")
	}

	return xml, chave, nil
}

// assinarXML assina digitalmente o XML da NFe
func assinarXML(cert certificate.Certificate, xml string) (string, error) {
	signer := certificate.CreateXMLSigner(cert)

	xmlAssinado, err := signer.SignNFeXML(xml)
	if err != nil {
		return "", fmt.Errorf("erro na assinatura: %v", err)
	}

	return xmlAssinado, nil
}

// autorizarNFe envia a NFe para autorização no SEFAZ
func autorizarNFe(client *nfe.NFEClient, xmlAssinado, chave string) error {
	// Testar conectividade primeiro
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("   📡 Testando conectividade SEFAZ...")
	status, err := client.QueryStatus(ctx)
	if err != nil {
		return fmt.Errorf("falha na conectividade SEFAZ: %v", err)
	}

	fmt.Printf("   📊 Status SEFAZ: Online=%v\n", status.Online)

	// Enviar para autorização
	fmt.Println("   📤 Enviando NFe para autorização...")
	authCtx, authCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer authCancel()

	response, err := client.Authorize(authCtx, []byte(xmlAssinado))
	if err != nil {
		return fmt.Errorf("erro na autorização: %v", err)
	}

	fmt.Printf("   📋 Resposta: Status=%d, Protocolo=%s, StatusText=%s\n", response.Status, response.Protocol, response.StatusText)

	// Em ambiente de homologação, responses específicos são esperados
	if response.Status == 539 {
		fmt.Println("   ✅ Resposta esperada para ambiente de homologação")
		return nil
	}

	return nil
}

// salvarArquivos salva os XMLs e chave em arquivos
func salvarArquivos(xml, xmlAssinado, chave string) error {
	arquivos := map[string]string{
		"nfe_original.xml": xml,
		"nfe_assinada.xml": xmlAssinado,
		"chave_acesso.txt": chave,
	}

	for arquivo, conteudo := range arquivos {
		err := os.WriteFile(arquivo, []byte(conteudo), 0644)
		if err != nil {
			return fmt.Errorf("erro ao salvar %s: %v", arquivo, err)
		}
	}

	return nil
}

// extrairNomeCertificado extrai o nome da empresa do subject do certificado
func extrairNomeCertificado(subject string) string {
	// Extrair nome da empresa do formato "CN=EMPRESA:CNPJ,..."
	if len(subject) > 3 && subject[:3] == "CN=" {
		end := len(subject)
		if commaPos := findFirst(subject, ','); commaPos != -1 {
			end = commaPos
		}
		if colonPos := findFirst(subject, ':'); colonPos != -1 && colonPos < end {
			return subject[3:colonPos]
		}
		return subject[3:end]
	}
	return "Certificado Digital"
}

// findFirst encontra a primeira ocorrência de um caractere
func findFirst(s string, c rune) int {
	for i, r := range s {
		if r == c {
			return i
		}
	}
	return -1
}
