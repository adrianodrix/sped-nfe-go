package main

import (
	"fmt"
	"log"

	"github.com/adrianodrix/sped-nfe-go/certificate"
)

// Exemplo de assinatura digital XML com certificados A1/A3
func main() {
	fmt.Println("=== Exemplo de Assinatura Digital XML com XMLDSig ===")

	// XML de exemplo da NFe
	nfeXML := `<?xml version="1.0" encoding="UTF-8"?>
<NFe xmlns="http://www.portalfiscal.inf.br/nfe">
  <infNFe Id="NFe35200214200166000187550010000000046550000001">
    <ide>
      <cUF>35</cUF>
      <cNF>55000000</cNF>
      <natOp>Venda</natOp>
      <mod>55</mod>
      <serie>1</serie>
      <nNF>4</nNF>
      <dhEmi>2020-02-01T10:00:00-03:00</dhEmi>
      <tpNF>1</tpNF>
      <idDest>1</idDest>
      <cMunFG>3550308</cMunFG>
      <tpImp>1</tpImp>
      <tpEmis>1</tpEmis>
      <cDV>1</cDV>
      <tpAmb>2</tpAmb>
      <finNFe>1</finNFe>
      <indFinal>1</indFinal>
      <indPres>1</indPres>
    </ide>
    <emit>
      <CNPJ>14200166000187</CNPJ>
      <xNome>Empresa Teste LTDA</xNome>
      <enderEmit>
        <xLgr>Rua Teste</xLgr>
        <nro>123</nro>
        <xBairro>Centro</xBairro>
        <cMun>3550308</cMun>
        <xMun>São Paulo</xMun>
        <UF>SP</UF>
        <CEP>01000000</CEP>
      </enderEmit>
      <IE>123456789012</IE>
      <CRT>1</CRT>
    </emit>
    <dest>
      <CPF>12345678901</CPF>
      <xNome>Cliente Teste</xNome>
      <enderDest>
        <xLgr>Rua Cliente</xLgr>
        <nro>456</nro>
        <xBairro>Vila Teste</xBairro>
        <cMun>3550308</cMun>
        <xMun>São Paulo</xMun>
        <UF>SP</UF>
        <CEP>02000000</CEP>
      </enderDest>
      <indIEDest>9</indIEDest>
    </dest>
    <det nItem="1">
      <prod>
        <cProd>001</cProd>
        <cEAN/>
        <xProd>Produto Teste</xProd>
        <NCM>12345678</NCM>
        <CFOP>5102</CFOP>
        <uCom>UN</uCom>
        <qCom>1.0000</qCom>
        <vUnCom>100.00</vUnCom>
        <vProd>100.00</vProd>
        <cEANTrib/>
        <uTrib>UN</uTrib>
        <qTrib>1.0000</qTrib>
        <vUnTrib>100.00</vUnTrib>
        <indTot>1</indTot>
      </prod>
      <imposto>
        <ICMS>
          <ICMS00>
            <orig>0</orig>
            <CST>00</CST>
            <modBC>0</modBC>
            <vBC>100.00</vBC>
            <pICMS>18.00</pICMS>
            <vICMS>18.00</vICMS>
          </ICMS00>
        </ICMS>
      </imposto>
    </det>
    <total>
      <ICMSTot>
        <vBC>100.00</vBC>
        <vICMS>18.00</vICMS>
        <vICMSDeson>0.00</vICMSDeson>
        <vFCP>0.00</vFCP>
        <vBCST>0.00</vBCST>
        <vST>0.00</vST>
        <vFCPST>0.00</vFCPST>
        <vFCPSTRet>0.00</vFCPSTRet>
        <vProd>100.00</vProd>
        <vFrete>0.00</vFrete>
        <vSeg>0.00</vSeg>
        <vDesc>0.00</vDesc>
        <vII>0.00</vII>
        <vIPI>0.00</vIPI>
        <vIPIDevol>0.00</vIPIDevol>
        <vPIS>0.00</vPIS>
        <vCOFINS>0.00</vCOFINS>
        <vOutro>0.00</vOutro>
        <vNF>100.00</vNF>
      </ICMSTot>
    </total>
    <transp>
      <modFrete>9</modFrete>
    </transp>
    <pag>
      <detPag>
        <tPag>01</tPag>
        <vPag>100.00</vPag>
      </detPag>
    </pag>
  </infNFe>
</NFe>`

	// Exemplo 1: Assinatura com certificado A1 (arquivo .pfx/.p12)
	fmt.Println("\n1. Assinatura com Certificado A1 (.pfx/.p12)")
	exemploA1(nfeXML)

	// Exemplo 2: Assinatura com certificado A3 (token PKCS#11)
	fmt.Println("\n2. Assinatura com Certificado A3 (Token PKCS#11)")
	exemploA3(nfeXML)

	// Exemplo 3: Validação de assinatura existente
	fmt.Println("\n3. Validação de Assinatura Existente")
	exemploValidacao(nfeXML)

	// Exemplo 4: Canonicalização XML
	fmt.Println("\n4. Canonicalização XML")
	exemploCanonicalizacao(nfeXML)
}

func exemploA1(nfeXML string) {
	fmt.Println("--- Carregando certificado A1 ---")

	// Exemplo de carregamento de certificado A1
	// Em produção, substitua pelos caminhos e senhas reais
	certPath := "/caminho/para/certificado.pfx"
	_ = "senha_do_certificado" // Evita warning de variável não usada

	fmt.Printf("Carregando certificado: %s\n", certPath)
	fmt.Println("NOTA: Este é um exemplo. Substitua pelos valores reais em produção.")

	// Simulação de carregamento (descomente para uso real)
	/*
		a1Cert, err := certificate.LoadA1FromFile(certPath, senha)
		if err != nil {
			log.Printf("Erro ao carregar certificado A1: %v", err)
			return
		}
		defer a1Cert.Close()

		fmt.Printf("Certificado carregado com sucesso!\n")
		fmt.Printf("Subject: %s\n", a1Cert.GetSubject())
		fmt.Printf("Serial: %s\n", a1Cert.GetSerialNumber())
		fmt.Printf("Válido até: %v\n", a1Cert.GetValidityPeriod())

		// Criar signer XMLDSig com SHA-1 (padrão SEFAZ)
		signer := certificate.SignWithSHA1(a1Cert)

		// Assinar NFe XML
		result, err := signer.SignNFeXML(nfeXML)
		if err != nil {
			log.Printf("Erro ao assinar NFe: %v", err)
			return
		}

		fmt.Printf("NFe assinada com sucesso!\n")
		fmt.Printf("Algoritmo: %s\n", result.Algorithm)
		fmt.Printf("Timestamp: %v\n", result.Timestamp)
		fmt.Printf("Tamanho do XML assinado: %d bytes\n", len(result.SignedXML))

		// Salvar XML assinado (opcional)
		// ioutil.WriteFile("nfe_assinada.xml", []byte(result.SignedXML), 0644)
	*/

	fmt.Println("Para usar em produção, descomente o código acima e forneça os caminhos corretos.")
}

func exemploA3(nfeXML string) {
	fmt.Println("--- Conectando ao token A3 ---")

	// Configuração do token PKCS#11
	libraryPath := "/usr/lib/x86_64-linux-gnu/opensc-pkcs11.so" // Linux
	// libraryPath := "C:\\Windows\\System32\\eToken.dll"        // Windows
	tokenLabel := "MEU_TOKEN"
	_ = "1234" // Evita warning de variável não usada

	fmt.Printf("Biblioteca PKCS#11: %s\n", libraryPath)
	fmt.Printf("Token: %s\n", tokenLabel)
	fmt.Println("NOTA: Este é um exemplo. Substitua pelos valores reais em produção.")

	// Simulação de carregamento (descomente para uso real)
	/*
		// Listar tokens disponíveis
		fmt.Println("Listando tokens disponíveis...")
		tokens, err := certificate.GetAvailableTokens(libraryPath)
		if err != nil {
			log.Printf("Erro ao listar tokens: %v", err)
			return
		}

		fmt.Printf("Encontrados %d tokens:\n", len(tokens))
		for i, token := range tokens {
			fmt.Printf("  %d. Slot: %d, Label: %s, Presente: %v\n",
				i+1, token.Slot, token.Label, token.IsPresent)
		}

		// Configurar PKCS#11
		pkcs11Config := &certificate.PKCS11Config{
			LibraryPath: libraryPath,
			TokenLabel:  tokenLabel,
			PIN:         pin,
		}

		// Carregar certificado A3
		a3Cert, err := certificate.LoadA3FromToken(pkcs11Config)
		if err != nil {
			log.Printf("Erro ao carregar certificado A3: %v", err)
			return
		}
		defer a3Cert.Close()

		fmt.Printf("Certificado A3 carregado com sucesso!\n")
		fmt.Printf("Subject: %s\n", a3Cert.GetSubject())
		fmt.Printf("Token: %s\n", a3Cert.GetTokenLabel())
		fmt.Printf("Token presente: %v\n", a3Cert.IsTokenPresent())

		// Testar conexão com o token
		err = a3Cert.TestConnection()
		if err != nil {
			log.Printf("Erro na conexão com o token: %v", err)
			return
		}
		fmt.Println("Conexão com token validada!")

		// Criar signer XMLDSig
		signer := certificate.SignWithSHA1(a3Cert)

		// Assinar NFe XML
		result, err := signer.SignNFeXML(nfeXML)
		if err != nil {
			log.Printf("Erro ao assinar NFe com A3: %v", err)
			return
		}

		fmt.Printf("NFe assinada com certificado A3!\n")
		fmt.Printf("Algoritmo: %s\n", result.Algorithm)
		fmt.Printf("Tamanho do XML assinado: %d bytes\n", len(result.SignedXML))
	*/

	fmt.Println("Para usar em produção, descomente o código acima e configure o token corretamente.")
}

func exemploValidacao(nfeXML string) {
	fmt.Println("--- Validando assinatura XML ---")

	// Simulação de XML assinado (substitua por XML real)
	xmlAssinado := nfeXML // Em produção, use XML já assinado

	fmt.Println("Validando estrutura da assinatura...")

	// Validação rápida (apenas estrutura)
	isValid, err := certificate.QuickValidateSignature(xmlAssinado)
	if err != nil {
		log.Printf("Erro na validação rápida: %v", err)
		return
	}

	fmt.Printf("Validação rápida: %v\n", isValid)

	// Validação completa
	fmt.Println("Executando validação completa...")

	// Configuração de validação personalizada
	config := &certificate.ValidationConfig{
		RequireValidCertificate: true,
		RequireICPBrasil:        true,
		CheckRevocation:         false, // Desabilitado para exemplo
		MaxClockSkew:            certificate.DefaultValidationConfig().MaxClockSkew,
		AllowedSignatureAlgorithms: []string{
			"http://www.w3.org/2000/09/xmldsig#rsa-sha1",
			"http://www.w3.org/2001/04/xmldsig-more#rsa-sha256",
		},
		AllowedDigestAlgorithms: []string{
			"http://www.w3.org/2000/09/xmldsig#sha1",
			"http://www.w3.org/2001/04/xmlenc#sha256",
		},
	}

	validator := certificate.NewSignatureValidator(config)
	result, err := validator.ValidateXMLSignature(xmlAssinado)
	if err != nil {
		log.Printf("Erro na validação: %v", err)
		return
	}

	fmt.Printf("Resultado da validação:\n")
	fmt.Printf("  Válida: %v\n", result.IsValid)
	fmt.Printf("  Assinatura válida: %v\n", result.SignatureValid)
	fmt.Printf("  Certificado válido: %v\n", result.CertificateValid)
	fmt.Printf("  Algoritmo de assinatura: %s\n", result.SignatureAlgorithm)

	if len(result.Errors) > 0 {
		fmt.Printf("  Erros encontrados:\n")
		for _, erro := range result.Errors {
			fmt.Printf("    - %s\n", erro)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("  Avisos:\n")
		for _, aviso := range result.Warnings {
			fmt.Printf("    - %s\n", aviso)
		}
	}

	if result.Certificate != nil {
		fmt.Printf("  Certificado extraído:\n")
		fmt.Printf("    Subject: %s\n", result.Certificate.Subject.String())
		fmt.Printf("    Issuer: %s\n", result.Certificate.Issuer.String())
		fmt.Printf("    Válido até: %v\n", result.Certificate.NotAfter)
	}
}

func exemploCanonicalizacao(nfeXML string) {
	fmt.Println("--- Canonicalização XML ---")

	// Criar canonicalizador com configuração padrão
	canonicalizer := certificate.NewXMLCanonicalizer(certificate.DefaultCanonicalizationConfig())

	fmt.Printf("Método de canonicalização: %s\n", certificate.DefaultCanonicalizationConfig().Method)

	// Canonicalizar XML
	resultado, err := canonicalizer.Canonicalize(nfeXML)
	if err != nil {
		log.Printf("Erro na canonicalização: %v", err)
		return
	}

	fmt.Printf("XML canonicalizado com sucesso!\n")
	fmt.Printf("Tamanho original: %d bytes\n", len(nfeXML))
	fmt.Printf("Tamanho canonicalizado: %d bytes\n", len(resultado))

	// Aplicar transformação enveloped signature
	fmt.Println("\nAplicando transformação Enveloped Signature...")

	transformado, err := certificate.EnvelopedSignatureTransform(nfeXML)
	if err != nil {
		log.Printf("Erro na transformação: %v", err)
		return
	}

	fmt.Printf("Transformação aplicada com sucesso!\n")
	fmt.Printf("Tamanho após transformação: %d bytes\n", len(transformado))

	// Canonicalização para assinatura
	fmt.Println("\nPreparando para assinatura...")

	elementID := "NFe35200214200166000187550010000000046550000001"
	preparado, err := certificate.CanonicalizeForSigning(nfeXML, elementID)
	if err != nil {
		log.Printf("Erro na preparação: %v", err)
		return
	}

	fmt.Printf("XML preparado para assinatura!\n")
	fmt.Printf("Elemento ID: %s\n", elementID)
	fmt.Printf("Tamanho preparado: %d bytes\n", len(preparado))

	// Validar forma canônica
	fmt.Println("\nValidando forma canônica...")

	err = certificate.ValidateCanonicalForm(string(resultado))
	if err != nil {
		fmt.Printf("XML não está em forma canônica: %v\n", err)
	} else {
		fmt.Println("XML está em forma canônica!")
	}
}
