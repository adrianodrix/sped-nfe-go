package main

import (
	"fmt"
	"log"

	"github.com/adrianodrix/sped-nfe-go/nfe"
)

func main() {
	fmt.Println("=== Exemplos de Cálculos de Impostos ===\n")

	// Example 1: Basic tax calculation for normal regime
	example1()

	// Example 2: Tax calculation for Simples Nacional
	example2()

	// Example 3: Service tax calculation (ISSQN)
	example3()

	// Example 4: Complex product with multiple taxes
	example4()

	// Example 5: Tax validation
	example5()

	// Example 6: Batch tax calculation
	example6()
}

// example1 demonstrates basic tax calculation for normal tax regime
func example1() {
	fmt.Println("1. Cálculo Básico de Impostos - Regime Normal")
	fmt.Println("===========================================")

	// Configure tax calculator for normal regime
	config := &nfe.TaxConfig{
		ICMSRate:         18.0, // 18% ICMS rate
		IPIRate:          5.0,  // 5% IPI rate
		PISRate:          1.65, // 1.65% PIS rate
		COFINSRate:       7.6,  // 7.6% COFINS rate
		FederalTaxRegime: "NORMAL",
		UF:               "SP",
	}

	calculator := nfe.NewTaxCalculator(config)

	// Create a sample item
	item := &nfe.Item{
		NItem: "1",
		Prod: nfe.Produto{
			CProd:    "001",
			CEAN:     "SEM GTIN",
			XProd:    "Smartphone Android",
			NCM:      "85171211",
			CFOP:     "5102",
			UCom:     "UN",
			QCom:     "1.00",
			VUnCom:   "1000.00",
			VProd:    "1000.00",
			CEANTrib: "SEM GTIN",
			UTrib:    "UN",
			QTrib:    "1.00",
			VUnTrib:  "1000.00",
		},
		Imposto: nfe.Imposto{},
	}

	// Calculate taxes
	err := calculator.CalculateItemTaxes(item)
	if err != nil {
		log.Printf("Erro ao calcular impostos: %v", err)
		return
	}

	// Display results
	fmt.Printf("Produto: %s\n", item.Prod.XProd)
	fmt.Printf("Valor: R$ %s\n", item.Prod.VProd)
	fmt.Println("\nImpostos Calculados:")

	if item.Imposto.ICMS != nil && item.Imposto.ICMS.ICMS00 != nil {
		fmt.Printf("  ICMS (18%%): R$ %s\n", item.Imposto.ICMS.ICMS00.VICMS)
	}

	if item.Imposto.IPI != nil && item.Imposto.IPI.IPITrib != nil {
		fmt.Printf("  IPI (5%%): R$ %s\n", item.Imposto.IPI.IPITrib.VIPI)
	}

	if item.Imposto.PIS != nil && item.Imposto.PIS.PISAliq != nil {
		fmt.Printf("  PIS (1.65%%): R$ %s\n", item.Imposto.PIS.PISAliq.VPIS)
	}

	if item.Imposto.COFINS != nil && item.Imposto.COFINS.COFINSAliq != nil {
		fmt.Printf("  COFINS (7.6%%): R$ %s\n", item.Imposto.COFINS.COFINSAliq.VCOFINS)
	}

	fmt.Println()
}

// example2 demonstrates tax calculation for Simples Nacional regime
func example2() {
	fmt.Println("2. Cálculo de Impostos - Simples Nacional")
	fmt.Println("========================================")

	// Configure tax calculator for Simples Nacional
	config := &nfe.TaxConfig{
		FederalTaxRegime: "SIMPLES",
		UF:               "SP",
	}

	calculator := nfe.NewTaxCalculator(config)

	// Create a sample item
	item := &nfe.Item{
		NItem: "1",
		Prod: nfe.Produto{
			CProd:  "002",
			XProd:  "Camiseta Algodão",
			NCM:    "61091000",
			CFOP:   "5102",
			QCom:   "5.00",
			VUnCom: "25.00",
			VProd:  "125.00",
		},
		Imposto: nfe.Imposto{},
	}

	// Calculate taxes
	err := calculator.CalculateItemTaxes(item)
	if err != nil {
		log.Printf("Erro ao calcular impostos: %v", err)
		return
	}

	// Display results
	fmt.Printf("Produto: %s\n", item.Prod.XProd)
	fmt.Printf("Quantidade: %s\n", item.Prod.QCom)
	fmt.Printf("Valor Total: R$ %s\n", item.Prod.VProd)
	fmt.Println("\nTributação Simples Nacional:")

	if item.Imposto.ICMS != nil && item.Imposto.ICMS.ICMSSN102 != nil {
		fmt.Printf("  ICMS SN: CSOSN %s (sem direito a crédito)\n", item.Imposto.ICMS.ICMSSN102.CSOSN)
	}

	fmt.Println()
}

// example3 demonstrates service tax calculation with ISSQN
func example3() {
	fmt.Println("3. Cálculo de Impostos para Serviços (ISSQN)")
	fmt.Println("============================================")

	// Configure tax calculator for services
	config := &nfe.TaxConfig{
		ISSQNRate:        5.0,       // 5% ISSQN rate
		ServiceMunCode:   "3550308", // São Paulo municipality code
		ServiceListCode:  "14.01",   // Service list code
		FederalTaxRegime: "NORMAL",
	}

	calculator := nfe.NewTaxCalculator(config)

	// Create a service item
	item := &nfe.Item{
		NItem: "1",
		Prod: nfe.Produto{
			CProd:  "SRV001",
			XProd:  "Desenvolvimento de Software",
			NCM:    "00000000", // Services typically use 00000000
			CFOP:   "5933",     // Service CFOP
			QCom:   "1.00",
			VUnCom: "5000.00",
			VProd:  "5000.00",
		},
		Imposto: nfe.Imposto{},
	}

	// Calculate taxes
	err := calculator.CalculateItemTaxes(item)
	if err != nil {
		log.Printf("Erro ao calcular impostos: %v", err)
		return
	}

	// Display results
	fmt.Printf("Serviço: %s\n", item.Prod.XProd)
	fmt.Printf("Valor: R$ %s\n", item.Prod.VProd)
	fmt.Println("\nImpostos Calculados:")

	if item.Imposto.ISSQN != nil {
		fmt.Printf("  ISSQN (5%%): R$ %s\n", item.Imposto.ISSQN.VISSQN)
		fmt.Printf("  Município: %s\n", item.Imposto.ISSQN.CMunFG)
		fmt.Printf("  Lista de Serviço: %s\n", item.Imposto.ISSQN.CListServ)
	}

	fmt.Println()
}

// example4 demonstrates complex product with multiple taxes
func example4() {
	fmt.Println("4. Produto Complexo com Múltiplos Impostos")
	fmt.Println("==========================================")

	// Configure tax calculator with higher rates
	config := &nfe.TaxConfig{
		ICMSRate:         25.0, // Higher ICMS rate for luxury items
		IPIRate:          15.0, // Higher IPI rate
		PISRate:          1.65,
		COFINSRate:       7.6,
		FederalTaxRegime: "NORMAL",
		UF:               "RJ",
	}

	calculator := nfe.NewTaxCalculator(config)

	// Create a luxury item
	item := &nfe.Item{
		NItem: "1",
		Prod: nfe.Produto{
			CProd:  "LUX001",
			XProd:  "Relógio de Luxo Importado",
			NCM:    "91011100",
			CFOP:   "6102", // Interstate sale
			QCom:   "1.00",
			VUnCom: "15000.00",
			VProd:  "15000.00",
		},
		Imposto: nfe.Imposto{},
	}

	// Calculate taxes
	err := calculator.CalculateItemTaxes(item)
	if err != nil {
		log.Printf("Erro ao calcular impostos: %v", err)
		return
	}

	// Calculate total taxes
	items := []nfe.Item{*item}
	totals, err := calculator.CalculateTotalTaxes(items)
	if err != nil {
		log.Printf("Erro ao calcular totais: %v", err)
		return
	}

	// Display results
	fmt.Printf("Produto: %s\n", item.Prod.XProd)
	fmt.Printf("Valor: R$ %s\n", item.Prod.VProd)
	fmt.Println("\nImpostos Calculados:")

	if item.Imposto.ICMS != nil && item.Imposto.ICMS.ICMS00 != nil {
		fmt.Printf("  ICMS (25%%): R$ %s\n", item.Imposto.ICMS.ICMS00.VICMS)
	}

	if item.Imposto.IPI != nil && item.Imposto.IPI.IPITrib != nil {
		fmt.Printf("  IPI (15%%): R$ %s\n", item.Imposto.IPI.IPITrib.VIPI)
	}

	if item.Imposto.PIS != nil && item.Imposto.PIS.PISAliq != nil {
		fmt.Printf("  PIS (1.65%%): R$ %s\n", item.Imposto.PIS.PISAliq.VPIS)
	}

	if item.Imposto.COFINS != nil && item.Imposto.COFINS.COFINSAliq != nil {
		fmt.Printf("  COFINS (7.6%%): R$ %s\n", item.Imposto.COFINS.COFINSAliq.VCOFINS)
	}

	fmt.Println("\nTotais:")
	fmt.Printf("  Total ICMS: R$ %.2f\n", totals.TotalICMS)
	fmt.Printf("  Total IPI: R$ %.2f\n", totals.TotalIPI)
	fmt.Printf("  Total PIS: R$ %.2f\n", totals.TotalPIS)
	fmt.Printf("  Total COFINS: R$ %.2f\n", totals.TotalCOFINS)
	fmt.Printf("  Total Impostos: R$ %.2f\n",
		totals.TotalICMS+totals.TotalIPI+totals.TotalPIS+totals.TotalCOFINS)

	fmt.Println()
}

// example5 demonstrates tax validation
func example5() {
	fmt.Println("5. Validação de Impostos")
	fmt.Println("========================")

	// Create validator
	validatorConfig := &nfe.ValidationConfig{
		UF:               "SP",
		Environment:      "HOMOLOGACAO",
		StrictValidation: true,
	}

	validator := nfe.NewTaxValidator(validatorConfig)

	// Create an item with validation issues
	item := &nfe.Item{
		NItem: "1",
		Prod: nfe.Produto{
			CProd: "VAL001",
			XProd: "Produto para Validação",
			NCM:   "1234567", // Invalid NCM (too short)
			CFOP:  "5102",
		},
		Imposto: nfe.Imposto{
			ICMS: &nfe.ICMS{
				ICMS00: &nfe.ICMS00{
					Orig:  "0",
					CST:   "10", // Wrong CST for ICMS00
					ModBC: "0",
					VBC:   "100.00",
					PICMS: "18.00",
					VICMS: "20.00", // Wrong calculation
				},
			},
		},
	}

	// Validate taxes
	errors := validator.ValidateItemTaxes(item)

	fmt.Printf("Item: %s\n", item.Prod.XProd)
	fmt.Printf("Erros encontrados: %d\n\n", len(errors))

	for i, err := range errors {
		fmt.Printf("Erro %d:\n", i+1)
		fmt.Printf("  Campo: %s\n", err.Field)
		fmt.Printf("  Código: %s\n", err.Code)
		fmt.Printf("  Mensagem: %s\n", err.Message)
		fmt.Printf("  Severidade: %s\n", err.Severity)
		fmt.Printf("  Regra: %s\n\n", err.Rule)
	}
}

// example6 demonstrates batch tax calculation for multiple items
func example6() {
	fmt.Println("6. Cálculo em Lote para Múltiplos Itens")
	fmt.Println("=======================================")

	// Configure tax calculator
	config := &nfe.TaxConfig{
		ICMSRate:         18.0,
		IPIRate:          0.0, // No IPI for this example
		PISRate:          1.65,
		COFINSRate:       7.6,
		FederalTaxRegime: "NORMAL",
	}

	calculator := nfe.NewTaxCalculator(config)

	// Create multiple items
	items := []nfe.Item{
		{
			NItem: "1",
			Prod: nfe.Produto{
				CProd:  "PROD001",
				XProd:  "Notebook",
				QCom:   "2.00",
				VUnCom: "2500.00",
				VProd:  "5000.00",
			},
			Imposto: nfe.Imposto{},
		},
		{
			NItem: "2",
			Prod: nfe.Produto{
				CProd:  "PROD002",
				XProd:  "Mouse Wireless",
				QCom:   "5.00",
				VUnCom: "80.00",
				VProd:  "400.00",
			},
			Imposto: nfe.Imposto{},
		},
		{
			NItem: "3",
			Prod: nfe.Produto{
				CProd:  "PROD003",
				XProd:  "Teclado Mecânico",
				QCom:   "3.00",
				VUnCom: "350.00",
				VProd:  "1050.00",
			},
			Imposto: nfe.Imposto{},
		},
	}

	// Calculate taxes for all items
	fmt.Println("Calculando impostos para todos os itens...")
	for i := range items {
		err := calculator.CalculateItemTaxes(&items[i])
		if err != nil {
			log.Printf("Erro ao calcular impostos para item %d: %v", i+1, err)
			continue
		}
		fmt.Printf("✓ Item %d: %s\n", i+1, items[i].Prod.XProd)
	}

	// Calculate totals
	totals, err := calculator.CalculateTotalTaxes(items)
	if err != nil {
		log.Printf("Erro ao calcular totais: %v", err)
		return
	}

	// Display summary
	fmt.Println("\nResumo da Nota Fiscal:")
	fmt.Println("======================")

	var totalProducts float64
	for _, item := range items {
		value, _ := parseDecimal(item.Prod.VProd)
		totalProducts += value
		fmt.Printf("  %s: R$ %s\n", item.Prod.XProd, item.Prod.VProd)
	}

	fmt.Printf("\nSubtotal Produtos: R$ %.2f\n", totalProducts)
	fmt.Printf("Total ICMS: R$ %.2f\n", totals.TotalICMS)
	fmt.Printf("Total PIS: R$ %.2f\n", totals.TotalPIS)
	fmt.Printf("Total COFINS: R$ %.2f\n", totals.TotalCOFINS)
	fmt.Printf("Total Impostos: R$ %.2f\n",
		totals.TotalICMS+totals.TotalPIS+totals.TotalCOFINS)
	fmt.Printf("Total da Nota: R$ %.2f\n",
		totalProducts+(totals.TotalICMS+totals.TotalPIS+totals.TotalCOFINS))

	fmt.Println()
}

// Helper function to parse decimal values
func parseDecimal(s string) (float64, error) {
	// Implementation would be similar to the one in tax_calculator.go
	// This is a simplified version for the example
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}
