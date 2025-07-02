package main

import (
	"fmt"

	"github.com/adrianodrix/sped-nfe-go/certificate"
)

func main() {
	fmt.Println("=== Debug Certificado ===")
	
	certPath := "refs/certificates/cert-valido-jan-2026.pfx"
	
	// Tentar diferentes senhas
	passwords := []string{
		"kzm7rwu!ewv1ymw3YTM@",
		"",
		"123456",
		"password",
		"cert",
	}
	
	for i, password := range passwords {
		fmt.Printf("\n%d. Tentando senha: '%s'\n", i+1, password)
		
		cert, err := certificate.LoadA1FromFile(certPath, password)
		if err != nil {
			fmt.Printf("   ❌ Erro: %v\n", err)
			continue
		}
		
		fmt.Println("   ✅ Certificado carregado com sucesso!")
		fmt.Printf("   📋 Titular: %s\n", cert.GetSubject())
		fmt.Printf("   📋 Emissor: %s\n", cert.GetIssuer())
		fmt.Printf("   📋 Válido: %v\n", cert.IsValid())
		
		notBefore, notAfter := cert.GetValidityPeriod()
		fmt.Printf("   📋 Validade: %s até %s\n", 
			notBefore.Format("02/01/2006"), 
			notAfter.Format("02/01/2006"))
		
		cert.Close()
		fmt.Println("   🎉 Senha correta encontrada!")
		return
	}
	
	fmt.Println("\n❌ Nenhuma senha funcionou")
	fmt.Println("💡 Possíveis causas:")
	fmt.Println("   - Arquivo corrompido")
	fmt.Println("   - Formato não é PKCS#12")
	fmt.Println("   - Senha contém caracteres especiais que precisam de escape")
}