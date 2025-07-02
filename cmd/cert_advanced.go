package main

import (
	"crypto/x509"
	"fmt"
	"os"

	"golang.org/x/crypto/pkcs12"
)

func main() {
	fmt.Println("=== Teste Avançado do Certificado ===")
	
	certPath := "refs/certificates/cert-valido-jan-2026.pfx"
	password := "kzm7rwu!ewv1ymw3YTM@"
	
	// Ler arquivo
	data, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Printf("❌ Erro ao ler arquivo: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Arquivo lido: %d bytes\n", len(data))
	
	// Tentar decodificar com ToPEM para extrair certificados e chaves
	fmt.Println("\nTentando extrair via ToPEM...")
	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		fmt.Printf("❌ Erro ToPEM: %v\n", err)
		return
	}
	
	fmt.Printf("✅ ToPEM bem-sucedido! Encontrados %d blocos PEM\n", len(blocks))
	
	var cert *x509.Certificate
	var privateKey interface{}
	
	for i, block := range blocks {
		fmt.Printf("\nBloco %d: Tipo = %s\n", i+1, block.Type)
		
		switch block.Type {
		case "CERTIFICATE":
			parsedCert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				fmt.Printf("   ❌ Erro ao parsear certificado: %v\n", err)
				continue
			}
			cert = parsedCert
			fmt.Printf("   ✅ Certificado parseado\n")
			fmt.Printf("   Subject: %s\n", cert.Subject.String())
			fmt.Printf("   Issuer: %s\n", cert.Issuer.String())
			fmt.Printf("   Validade: %s até %s\n", 
				cert.NotBefore.Format("02/01/2006"),
				cert.NotAfter.Format("02/01/2006"))
			fmt.Printf("   Válido agora: %v\n", cert.NotBefore.Before(cert.NotAfter))
			
		case "PRIVATE KEY":
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				// Tentar PKCS#1
				key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
				if err != nil {
					fmt.Printf("   ❌ Erro ao parsear chave privada: %v\n", err)
					continue
				}
			}
			privateKey = key
			fmt.Printf("   ✅ Chave privada parseada\n")
		}
	}
	
	if cert != nil && privateKey != nil {
		fmt.Println("\n🎉 SUCESSO! Certificado e chave extraídos com sucesso")
		fmt.Println("\n📋 Informações do certificado:")
		fmt.Printf("   Titular: %s\n", cert.Subject.String())
		fmt.Printf("   Emissor: %s\n", cert.Issuer.String())
		fmt.Printf("   Serial: %s\n", cert.SerialNumber.String())
		fmt.Printf("   Validade: %s até %s\n", 
			cert.NotBefore.Format("02/01/2006 15:04:05"),
			cert.NotAfter.Format("02/01/2006 15:04:05"))
		
		// Verificar se é um certificado ICP-Brasil
		if len(cert.Subject.Organization) > 0 {
			fmt.Printf("   Organização: %s\n", cert.Subject.Organization[0])
		}
		if len(cert.Subject.Country) > 0 {
			fmt.Printf("   País: %s\n", cert.Subject.Country[0])
		}
		
		fmt.Println("\n✅ O certificado pode ser usado com uma implementação customizada!")
		
	} else {
		fmt.Println("\n❌ Não foi possível extrair certificado e chave")
	}
}