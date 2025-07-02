package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/pkcs12"
)

func main() {
	fmt.Println("=== Debug Detalhado do Certificado ===")
	
	certPath := "refs/certificates/cert-valido-jan-2026.pfx"
	password := "kzm7rwu!ewv1ymw3YTM@"
	
	fmt.Printf("Arquivo: %s\n", certPath)
	fmt.Printf("Senha: %s\n", password)
	
	// Ler arquivo
	data, err := os.ReadFile(certPath)
	if err != nil {
		fmt.Printf("❌ Erro ao ler arquivo: %v\n", err)
		return
	}
	
	fmt.Printf("✅ Arquivo lido: %d bytes\n", len(data))
	
	// Verificar se é PKCS#12 válido
	fmt.Printf("Primeiros bytes: %x\n", data[:min(16, len(data))])
	
	// Tentar decodificar diretamente
	fmt.Println("\nTentando decodificar...")
	privateKey, cert, err := pkcs12.Decode(data, password)
	if err != nil {
		fmt.Printf("❌ Erro na decodificação: %v\n", err)
		
		// Tentar com senha vazia
		fmt.Println("\nTentando com senha vazia...")
		privateKey2, cert2, err2 := pkcs12.Decode(data, "")
		if err2 != nil {
			fmt.Printf("❌ Erro com senha vazia: %v\n", err2)
		} else {
			fmt.Println("✅ Sucesso com senha vazia!")
			fmt.Printf("Certificado: %v\n", cert2 != nil)
			fmt.Printf("Chave privada: %v\n", privateKey2 != nil)
		}
		return
	}
	
	fmt.Println("✅ Decodificação bem-sucedida!")
	fmt.Printf("Certificado: %v\n", cert != nil)
	fmt.Printf("Chave privada: %v\n", privateKey != nil)
	
	if cert != nil {
		fmt.Printf("Subject: %s\n", cert.Subject.String())
		fmt.Printf("Issuer: %s\n", cert.Issuer.String())
		fmt.Printf("Validade: %s até %s\n", 
			cert.NotBefore.Format("02/01/2006"),
			cert.NotAfter.Format("02/01/2006"))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}