package nfe

import (
	"fmt"
	"strconv"
	"strings"
)

// TaxCalculator provides automatic tax calculation capabilities
type TaxCalculator struct {
	config *TaxConfig
}

// TaxConfig holds configuration for tax calculations
type TaxConfig struct {
	// ICMS Configuration
	ICMSRate         float64 // Default ICMS rate
	ICMSSTRate       float64 // ICMS ST rate
	ICMSSTMargin     float64 // ICMS ST margin (MVA)
	ICMSReduction    float64 // ICMS reduction percentage
	
	// IPI Configuration
	IPIRate          float64 // IPI rate
	
	// PIS/COFINS Configuration
	PISRate          float64 // PIS rate
	COFINSRate       float64 // COFINS rate
	
	// ISSQN Configuration
	ISSQNRate        float64 // ISSQN rate
	ServiceMunCode   string  // Service municipality code
	ServiceListCode  string  // Service list code
	
	// General Configuration
	FederalTaxRegime string  // "NORMAL" or "SIMPLES"
	UF               string  // State code for tax rules
}

// NewTaxCalculator creates a new tax calculator with default configuration
func NewTaxCalculator(config *TaxConfig) *TaxCalculator {
	if config == nil {
		config = &TaxConfig{
			ICMSRate:         18.0, // Default ICMS rate
			ICMSSTRate:       18.0,
			ICMSSTMargin:     30.0,
			IPIRate:          0.0,
			PISRate:          1.65,
			COFINSRate:       7.6,
			ISSQNRate:        5.0,
			FederalTaxRegime: "NORMAL",
		}
	}
	
	return &TaxCalculator{config: config}
}

// CalculateItemTaxes calculates all taxes for a given item
func (tc *TaxCalculator) CalculateItemTaxes(item *Item) error {
	if item == nil {
		return fmt.Errorf("item cannot be nil")
	}
	
	// Parse values
	unitValue, err := parseDecimal(item.Prod.VUnCom)
	if err != nil {
		return fmt.Errorf("invalid unit value: %w", err)
	}
	
	quantity, err := parseDecimal(item.Prod.QCom)
	if err != nil {
		return fmt.Errorf("invalid quantity: %w", err)
	}
	
	totalValue := unitValue * quantity
	
	// Initialize tax structure if needed
	// Since Imposto is not a pointer in Item struct, we don't need to check for nil
	
	// Calculate taxes based on regime
	if tc.config.FederalTaxRegime == "SIMPLES" {
		return tc.calculateSimplesNacionalTaxes(item, totalValue)
	}
	
	return tc.calculateNormalTaxes(item, totalValue)
}

// calculateNormalTaxes calculates taxes for normal tax regime
func (tc *TaxCalculator) calculateNormalTaxes(item *Item, totalValue float64) error {
	// Calculate ICMS
	if err := tc.calculateICMS(item, totalValue); err != nil {
		return fmt.Errorf("ICMS calculation error: %w", err)
	}
	
	// Calculate IPI
	if err := tc.calculateIPI(item, totalValue); err != nil {
		return fmt.Errorf("IPI calculation error: %w", err)
	}
	
	// Calculate PIS
	if err := tc.calculatePIS(item, totalValue); err != nil {
		return fmt.Errorf("PIS calculation error: %w", err)
	}
	
	// Calculate COFINS
	if err := tc.calculateCOFINS(item, totalValue); err != nil {
		return fmt.Errorf("COFINS calculation error: %w", err)
	}
	
	// Calculate ISSQN if applicable (services)
	if tc.isService(item) {
		if err := tc.calculateISSQN(item, totalValue); err != nil {
			return fmt.Errorf("ISSQN calculation error: %w", err)
		}
	}
	
	return nil
}

// calculateSimplesNacionalTaxes calculates taxes for Simples Nacional regime
func (tc *TaxCalculator) calculateSimplesNacionalTaxes(item *Item, totalValue float64) error {
	// For Simples Nacional, use simplified ICMS calculation
	return tc.calculateICMSSimplesNacional(item, totalValue)
}

// calculateICMS calculates ICMS tax
func (tc *TaxCalculator) calculateICMS(item *Item, totalValue float64) error {
	if item.Imposto.ICMS == nil {
		item.Imposto.ICMS = &ICMS{}
	}
	
	// Default to ICMS00 (normal taxation)
	item.Imposto.ICMS.ICMS00 = &ICMS00{
		Orig:  "0", // National origin
		CST:   "00",
		ModBC: "0", // Value of operation
		VBC:   formatDecimal(totalValue),
		PICMS: formatDecimal(tc.config.ICMSRate),
		VICMS: formatDecimal(totalValue * tc.config.ICMSRate / 100),
	}
	
	return nil
}

// calculateICMSSimplesNacional calculates ICMS for Simples Nacional
func (tc *TaxCalculator) calculateICMSSimplesNacional(item *Item, _ float64) error {
	if item.Imposto.ICMS == nil {
		item.Imposto.ICMS = &ICMS{}
	}
	
	// Use ICMSSN102 (without credit)
	item.Imposto.ICMS.ICMSSN102 = &ICMSSN102{
		Orig:  "0",   // National origin
		CSOSN: "102", // Simples Nacional without credit
	}
	
	return nil
}

// calculateIPI calculates IPI tax
func (tc *TaxCalculator) calculateIPI(item *Item, totalValue float64) error {
	if tc.config.IPIRate == 0 {
		// Non-taxed IPI
		item.Imposto.IPI = &IPI{
			CEnq: "999", // General framework
			IPINT: &IPINT{
				CST: "01", // Exempt
			},
		}
		return nil
	}
	
	// Taxed IPI
	item.Imposto.IPI = &IPI{
		CEnq: "999", // General framework
		IPITrib: &IPITrib{
			CST:  "00",
			VBC:  formatDecimal(totalValue),
			PIPI: formatDecimal(tc.config.IPIRate),
			VIPI: formatDecimal(totalValue * tc.config.IPIRate / 100),
		},
	}
	
	return nil
}

// calculatePIS calculates PIS tax
func (tc *TaxCalculator) calculatePIS(item *Item, totalValue float64) error {
	if tc.config.PISRate == 0 {
		// Non-taxed PIS
		item.Imposto.PIS = &PIS{
			PISNT: &PISNT{
				CST: "04", // Exempt
			},
		}
		return nil
	}
	
	// Taxed PIS
	item.Imposto.PIS = &PIS{
		PISAliq: &PISAliq{
			CST:  "01",
			VBC:  formatDecimal(totalValue),
			PPIS: formatDecimal(tc.config.PISRate),
			VPIS: formatDecimal(totalValue * tc.config.PISRate / 100),
		},
	}
	
	return nil
}

// calculateCOFINS calculates COFINS tax
func (tc *TaxCalculator) calculateCOFINS(item *Item, totalValue float64) error {
	if tc.config.COFINSRate == 0 {
		// Non-taxed COFINS
		item.Imposto.COFINS = &COFINS{
			COFINSNT: &COFINSNT{
				CST: "04", // Exempt
			},
		}
		return nil
	}
	
	// Taxed COFINS
	item.Imposto.COFINS = &COFINS{
		COFINSAliq: &COFINSAliq{
			CST:     "01",
			VBC:     formatDecimal(totalValue),
			PCOFINS: formatDecimal(tc.config.COFINSRate),
			VCOFINS: formatDecimal(totalValue * tc.config.COFINSRate / 100),
		},
	}
	
	return nil
}

// calculateISSQN calculates ISSQN tax for services
func (tc *TaxCalculator) calculateISSQN(item *Item, totalValue float64) error {
	if tc.config.ISSQNRate == 0 {
		return nil // No ISSQN for this item
	}
	
	item.Imposto.ISSQN = &ISSQN{
		VBC:          formatDecimal(totalValue),
		VAliq:        formatDecimal(tc.config.ISSQNRate),
		VISSQN:       formatDecimal(totalValue * tc.config.ISSQNRate / 100),
		CMunFG:       tc.config.ServiceMunCode,
		CListServ:    tc.config.ServiceListCode,
		IndISS:       "1", // Exigible in issuer municipality
		IndIncentivo: "2", // No fiscal incentive
	}
	
	return nil
}

// isService determines if an item is a service based on its characteristics
func (tc *TaxCalculator) isService(item *Item) bool {
	// Check if NCM starts with service codes (generally 00 for services)
	ncm := item.Prod.NCM
	if len(ncm) >= 2 && ncm[:2] == "00" {
		return true
	}
	
	// Check CFOP for service operations (5900-5999, 6900-6999, etc.)
	cfop := item.Prod.CFOP
	if len(cfop) == 4 {
		cfopInt, err := strconv.Atoi(cfop)
		if err == nil {
			// Service CFOPs typically end in 9xx
			lastTwoDigits := cfopInt % 1000
			if lastTwoDigits >= 900 && lastTwoDigits <= 999 {
				return true
			}
		}
	}
	
	return false
}

// CalculateTotalTaxes calculates total tax values for all items
func (tc *TaxCalculator) CalculateTotalTaxes(items []Item) (*TaxTotals, error) {
	totals := &TaxTotals{}
	
	for _, item := range items {
		// Since Imposto is not a pointer, we don't need to check for nil
		
		// Sum ICMS values
		if err := tc.sumICMSValues(item.Imposto.ICMS, totals); err != nil {
			return nil, fmt.Errorf("error summing ICMS: %w", err)
		}
		
		// Sum IPI values
		if err := tc.sumIPIValues(item.Imposto.IPI, totals); err != nil {
			return nil, fmt.Errorf("error summing IPI: %w", err)
		}
		
		// Sum PIS values
		if err := tc.sumPISValues(item.Imposto.PIS, totals); err != nil {
			return nil, fmt.Errorf("error summing PIS: %w", err)
		}
		
		// Sum COFINS values
		if err := tc.sumCOFINSValues(item.Imposto.COFINS, totals); err != nil {
			return nil, fmt.Errorf("error summing COFINS: %w", err)
		}
		
		// Sum ISSQN values
		if err := tc.sumISSQNValues(item.Imposto.ISSQN, totals); err != nil {
			return nil, fmt.Errorf("error summing ISSQN: %w", err)
		}
	}
	
	return totals, nil
}

// TaxTotals holds total tax values
type TaxTotals struct {
	TotalICMS   float64
	TotalICMSST float64
	TotalIPI    float64
	TotalPIS    float64
	TotalCOFINS float64
	TotalISSQN  float64
}

// sumICMSValues sums ICMS values from different modalities
func (tc *TaxCalculator) sumICMSValues(icms *ICMS, totals *TaxTotals) error {
	if icms == nil {
		return nil
	}
	
	// Sum values from different ICMS modalities
	if icms.ICMS00 != nil {
		value, err := parseDecimal(icms.ICMS00.VICMS)
		if err == nil {
			totals.TotalICMS += value
		}
	}
	
	if icms.ICMS10 != nil {
		value, err := parseDecimal(icms.ICMS10.VICMS)
		if err == nil {
			totals.TotalICMS += value
		}
		stValue, err := parseDecimal(icms.ICMS10.VICMSST)
		if err == nil {
			totals.TotalICMSST += stValue
		}
	}
	
	if icms.ICMS20 != nil {
		value, err := parseDecimal(icms.ICMS20.VICMS)
		if err == nil {
			totals.TotalICMS += value
		}
	}
	
	// Add other ICMS modalities as needed...
	
	return nil
}

// sumIPIValues sums IPI values
func (tc *TaxCalculator) sumIPIValues(ipi *IPI, totals *TaxTotals) error {
	if ipi == nil || ipi.IPITrib == nil {
		return nil
	}
	
	value, err := parseDecimal(ipi.IPITrib.VIPI)
	if err == nil {
		totals.TotalIPI += value
	}
	
	return nil
}

// sumPISValues sums PIS values
func (tc *TaxCalculator) sumPISValues(pis *PIS, totals *TaxTotals) error {
	if pis == nil {
		return nil
	}
	
	if pis.PISAliq != nil {
		value, err := parseDecimal(pis.PISAliq.VPIS)
		if err == nil {
			totals.TotalPIS += value
		}
	}
	
	if pis.PISQtde != nil {
		value, err := parseDecimal(pis.PISQtde.VPIS)
		if err == nil {
			totals.TotalPIS += value
		}
	}
	
	if pis.PISOutr != nil {
		value, err := parseDecimal(pis.PISOutr.VPIS)
		if err == nil {
			totals.TotalPIS += value
		}
	}
	
	return nil
}

// sumCOFINSValues sums COFINS values
func (tc *TaxCalculator) sumCOFINSValues(cofins *COFINS, totals *TaxTotals) error {
	if cofins == nil {
		return nil
	}
	
	if cofins.COFINSAliq != nil {
		value, err := parseDecimal(cofins.COFINSAliq.VCOFINS)
		if err == nil {
			totals.TotalCOFINS += value
		}
	}
	
	if cofins.COFINSQtde != nil {
		value, err := parseDecimal(cofins.COFINSQtde.VCOFINS)
		if err == nil {
			totals.TotalCOFINS += value
		}
	}
	
	if cofins.COFINSOutr != nil {
		value, err := parseDecimal(cofins.COFINSOutr.VCOFINS)
		if err == nil {
			totals.TotalCOFINS += value
		}
	}
	
	return nil
}

// sumISSQNValues sums ISSQN values
func (tc *TaxCalculator) sumISSQNValues(issqn *ISSQN, totals *TaxTotals) error {
	if issqn == nil {
		return nil
	}
	
	value, err := parseDecimal(issqn.VISSQN)
	if err == nil {
		totals.TotalISSQN += value
	}
	
	return nil
}

// Helper functions

// parseDecimal parses a decimal string to float64
func parseDecimal(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	
	// Remove any thousand separators and replace comma with dot if needed
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, " ", "")
	
	return strconv.ParseFloat(s, 64)
}

// formatDecimal formats a float64 to string with 2 decimal places
func formatDecimal(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

// ValidateCalculatedTaxes validates calculated tax values against business rules
func (tc *TaxCalculator) ValidateCalculatedTaxes(item *Item) []string {
	var errors []string
	
	// Since Imposto is not a pointer, we don't need to check for nil
	
	// Validate ICMS
	if item.Imposto.ICMS != nil {
		errors = append(errors, tc.validateICMS(item.Imposto.ICMS)...)
	}
	
	// Validate IPI
	if item.Imposto.IPI != nil {
		errors = append(errors, tc.validateIPI(item.Imposto.IPI)...)
	}
	
	// Validate PIS
	if item.Imposto.PIS != nil {
		errors = append(errors, tc.validatePIS(item.Imposto.PIS)...)
	}
	
	// Validate COFINS
	if item.Imposto.COFINS != nil {
		errors = append(errors, tc.validateCOFINS(item.Imposto.COFINS)...)
	}
	
	return errors
}

// validateICMS validates ICMS tax values
func (tc *TaxCalculator) validateICMS(icms *ICMS) []string {
	var errors []string
	
	// Count active modalities (should be exactly one)
	activeModalities := 0
	
	if icms.ICMS00 != nil {
		activeModalities++
		if icms.ICMS00.CST != "00" {
			errors = append(errors, "ICMS00: CST deve ser '00'")
		}
	}
	
	if icms.ICMS10 != nil {
		activeModalities++
		if icms.ICMS10.CST != "10" {
			errors = append(errors, "ICMS10: CST deve ser '10'")
		}
	}
	
	// Add validation for other modalities...
	
	if activeModalities == 0 {
		errors = append(errors, "ICMS: nenhuma modalidade de ICMS foi informada")
	} else if activeModalities > 1 {
		errors = append(errors, "ICMS: apenas uma modalidade de ICMS deve ser informada")
	}
	
	return errors
}

// validateIPI validates IPI tax values
func (tc *TaxCalculator) validateIPI(ipi *IPI) []string {
	var errors []string
	
	// Framework code is required
	if ipi.CEnq == "" {
		errors = append(errors, "IPI: código de enquadramento (cEnq) é obrigatório")
	}
	
	// Either IPITrib or IPINT must be present
	if ipi.IPITrib == nil && ipi.IPINT == nil {
		errors = append(errors, "IPI: deve ser informado IPITrib ou IPINT")
	}
	
	if ipi.IPITrib != nil && ipi.IPINT != nil {
		errors = append(errors, "IPI: não é possível informar IPITrib e IPINT simultaneamente")
	}
	
	return errors
}

// validatePIS validates PIS tax values
func (tc *TaxCalculator) validatePIS(pis *PIS) []string {
	var errors []string
	
	// Count active modalities (should be exactly one)
	activeModalities := 0
	
	if pis.PISAliq != nil {
		activeModalities++
	}
	if pis.PISQtde != nil {
		activeModalities++
	}
	if pis.PISNT != nil {
		activeModalities++
	}
	if pis.PISOutr != nil {
		activeModalities++
	}
	
	if activeModalities == 0 {
		errors = append(errors, "PIS: nenhuma modalidade de PIS foi informada")
	} else if activeModalities > 1 {
		errors = append(errors, "PIS: apenas uma modalidade de PIS deve ser informada")
	}
	
	return errors
}

// validateCOFINS validates COFINS tax values
func (tc *TaxCalculator) validateCOFINS(cofins *COFINS) []string {
	var errors []string
	
	// Count active modalities (should be exactly one)
	activeModalities := 0
	
	if cofins.COFINSAliq != nil {
		activeModalities++
	}
	if cofins.COFINSQtde != nil {
		activeModalities++
	}
	if cofins.COFINSNT != nil {
		activeModalities++
	}
	if cofins.COFINSOutr != nil {
		activeModalities++
	}
	
	if activeModalities == 0 {
		errors = append(errors, "COFINS: nenhuma modalidade de COFINS foi informada")
	} else if activeModalities > 1 {
		errors = append(errors, "COFINS: apenas uma modalidade de COFINS deve ser informada")
	}
	
	return errors
}