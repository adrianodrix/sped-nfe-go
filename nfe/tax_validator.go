package nfe

import (
	"fmt"
	"regexp"
	"strconv"
)

// TaxValidator provides comprehensive tax validation according to SEFAZ rules
type TaxValidator struct {
	config *ValidationConfig
}

// ValidationConfig holds configuration for tax validation
type ValidationConfig struct {
	UF               string // State code for state-specific rules
	Environment      string // "PRODUCAO" or "HOMOLOGACAO"
	Version          string // NFe layout version
	StrictValidation bool   // Enable strict validation mode
}

// NewTaxValidator creates a new tax validator with configuration
func NewTaxValidator(config *ValidationConfig) *TaxValidator {
	if config == nil {
		config = &ValidationConfig{
			UF:               "SP",
			Environment:      "HOMOLOGACAO",
			Version:          "4.00",
			StrictValidation: true,
		}
	}

	return &TaxValidator{config: config}
}

// ValidateItemTaxes performs comprehensive validation of item taxes
func (tv *TaxValidator) ValidateItemTaxes(item *Item) []ValidationError {
	var errors []ValidationError

	// Validate product information first
	errors = append(errors, tv.validateProductForTax(item.Prod)...)

	// Validate tax structure consistency
	errors = append(errors, tv.validateTaxStructure(item.Imposto)...)

	// Validate ICMS
	if item.Imposto.ICMS != nil {
		errors = append(errors, tv.validateICMSAdvanced(item.Imposto.ICMS, item.Prod)...)
	}

	// Validate IPI
	if item.Imposto.IPI != nil {
		errors = append(errors, tv.validateIPIAdvanced(item.Imposto.IPI, item.Prod)...)
	}

	// Validate PIS
	if item.Imposto.PIS != nil {
		errors = append(errors, tv.validatePISAdvanced(item.Imposto.PIS, item.Prod)...)
	}

	// Validate COFINS
	if item.Imposto.COFINS != nil {
		errors = append(errors, tv.validateCOFINSAdvanced(item.Imposto.COFINS, item.Prod)...)
	}

	// Validate ISSQN
	if item.Imposto.ISSQN != nil {
		errors = append(errors, tv.validateISSQNAdvanced(item.Imposto.ISSQN, item.Prod)...)
	}

	// Cross-validation between taxes
	errors = append(errors, tv.validateCrossTaxRules(item)...)

	return errors
}

// ValidationError represents a tax validation error
type ValidationError struct {
	Field    string `json:"field"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "ERROR", "WARNING", "INFO"
	Rule     string `json:"rule"`     // SEFAZ rule reference
}

// validateProductForTax validates product information for tax calculation
func (tv *TaxValidator) validateProductForTax(prod Produto) []ValidationError {
	var errors []ValidationError

	// Validate NCM
	if err := tv.validateNCM(prod.NCM); err != nil {
		errors = append(errors, ValidationError{
			Field:    "prod.NCM",
			Code:     "NCM_INVALID",
			Message:  err.Error(),
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate CFOP
	if err := tv.validateCFOP(prod.CFOP); err != nil {
		errors = append(errors, ValidationError{
			Field:    "prod.CFOP",
			Code:     "CFOP_INVALID",
			Message:  err.Error(),
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate CEST (if applicable)
	if prod.CEST != "" {
		if err := tv.validateCEST(prod.CEST); err != nil {
			errors = append(errors, ValidationError{
				Field:    "prod.CEST",
				Code:     "CEST_INVALID",
				Message:  err.Error(),
				Severity: "ERROR",
				Rule:     "NT2019.001",
			})
		}
	}

	return errors
}

// validateNCM validates NCM (Nomenclatura Comum do Mercosul)
func (tv *TaxValidator) validateNCM(ncm string) error {
	// NCM must have 8 digits
	if len(ncm) != 8 {
		return fmt.Errorf("NCM deve conter 8 dígitos")
	}

	// Check if all characters are numeric
	if !regexp.MustCompile(`^\d{8}$`).MatchString(ncm) {
		return fmt.Errorf("NCM deve conter apenas dígitos")
	}

	// Validate against known NCM ranges (simplified validation)
	// Complete validation would require a full NCM table
	ncmInt, err := strconv.Atoi(ncm)
	if err != nil {
		return fmt.Errorf("NCM inválido: %v", err)
	}

	if ncmInt < 1000000 || ncmInt > 99999999 {
		return fmt.Errorf("NCM fora da faixa válida")
	}

	return nil
}

// validateCFOP validates CFOP (Código Fiscal de Operações e Prestações)
func (tv *TaxValidator) validateCFOP(cfop string) error {
	// CFOP must have 4 digits
	if len(cfop) != 4 {
		return fmt.Errorf("CFOP deve conter 4 dígitos")
	}

	// Check if all characters are numeric
	if !regexp.MustCompile(`^\d{4}$`).MatchString(cfop) {
		return fmt.Errorf("CFOP deve conter apenas dígitos")
	}

	cfopInt, err := strconv.Atoi(cfop)
	if err != nil {
		return fmt.Errorf("CFOP inválido: %v", err)
	}

	// Validate CFOP ranges
	firstDigit := cfopInt / 1000

	switch firstDigit {
	case 1: // Entries
		// Valid range: 1000-1999
		if cfopInt < 1000 || cfopInt > 1999 {
			return fmt.Errorf("CFOP de entrada fora da faixa válida")
		}
	case 2: // Entries from other states
		if cfopInt < 2000 || cfopInt > 2999 {
			return fmt.Errorf("CFOP de entrada de outros estados fora da faixa válida")
		}
	case 3: // Entries from abroad
		if cfopInt < 3000 || cfopInt > 3999 {
			return fmt.Errorf("CFOP de entrada do exterior fora da faixa válida")
		}
	case 5: // Exits within the state
		if cfopInt < 5000 || cfopInt > 5999 {
			return fmt.Errorf("CFOP de saída dentro do estado fora da faixa válida")
		}
	case 6: // Exits to other states
		if cfopInt < 6000 || cfopInt > 6999 {
			return fmt.Errorf("CFOP de saída para outros estados fora da faixa válida")
		}
	case 7: // Exits abroad
		if cfopInt < 7000 || cfopInt > 7999 {
			return fmt.Errorf("CFOP de saída para o exterior fora da faixa válida")
		}
	default:
		return fmt.Errorf("CFOP com primeiro dígito inválido")
	}

	return nil
}

// validateCEST validates CEST (Código Especificador da Substituição Tributária)
func (tv *TaxValidator) validateCEST(cest string) error {
	// CEST must have 7 digits
	if len(cest) != 7 {
		return fmt.Errorf("CEST deve conter 7 dígitos")
	}

	// Check if all characters are numeric
	if !regexp.MustCompile(`^\d{7}$`).MatchString(cest) {
		return fmt.Errorf("CEST deve conter apenas dígitos")
	}

	return nil
}

// validateTaxStructure validates the overall tax structure
func (tv *TaxValidator) validateTaxStructure(imposto Imposto) []ValidationError {
	var errors []ValidationError

	// ICMS is mandatory for products
	if imposto.ICMS == nil {
		errors = append(errors, ValidationError{
			Field:    "imposto.ICMS",
			Code:     "ICMS_MISSING",
			Message:  "ICMS é obrigatório",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate tax value format if present
	if imposto.VTotTrib != "" {
		if err := tv.validateDecimalField(imposto.VTotTrib, "vTotTrib"); err != nil {
			errors = append(errors, ValidationError{
				Field:    "imposto.vTotTrib",
				Code:     "VALUE_FORMAT_INVALID",
				Message:  err.Error(),
				Severity: "ERROR",
				Rule:     "NT2019.001",
			})
		}
	}

	return errors
}

// validateICMSAdvanced performs advanced ICMS validation
func (tv *TaxValidator) validateICMSAdvanced(icms *ICMS, prod Produto) []ValidationError {
	var errors []ValidationError

	// Count active modalities
	activeModalities := tv.countActiveICMSModalities(icms)

	if activeModalities == 0 {
		errors = append(errors, ValidationError{
			Field:    "imposto.ICMS",
			Code:     "ICMS_NO_MODALITY",
			Message:  "Nenhuma modalidade de ICMS informada",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	} else if activeModalities > 1 {
		errors = append(errors, ValidationError{
			Field:    "imposto.ICMS",
			Code:     "ICMS_MULTIPLE_MODALITIES",
			Message:  "Apenas uma modalidade de ICMS deve ser informada",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate specific modalities
	if icms.ICMS00 != nil {
		errors = append(errors, tv.validateICMS00(icms.ICMS00, prod)...)
	}

	if icms.ICMS10 != nil {
		errors = append(errors, tv.validateICMS10(icms.ICMS10, prod)...)
	}

	if icms.ICMS20 != nil {
		errors = append(errors, tv.validateICMS20(icms.ICMS20, prod)...)
	}

	// Add validation for other modalities...

	return errors
}

// validateICMS00 validates ICMS00 (normal taxation)
func (tv *TaxValidator) validateICMS00(icms00 *ICMS00, prod Produto) []ValidationError {
	var errors []ValidationError

	// Validate origin
	if err := tv.validateOrigin(icms00.Orig); err != nil {
		errors = append(errors, ValidationError{
			Field:    "ICMS00.orig",
			Code:     "ORIGIN_INVALID",
			Message:  err.Error(),
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate CST
	if icms00.CST != "00" {
		errors = append(errors, ValidationError{
			Field:    "ICMS00.CST",
			Code:     "CST_INVALID",
			Message:  "CST deve ser '00' para ICMS00",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate modBC
	if !tv.isValidModBC(icms00.ModBC) {
		errors = append(errors, ValidationError{
			Field:    "ICMS00.modBC",
			Code:     "MODBC_INVALID",
			Message:  "Modalidade de base de cálculo inválida",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate values
	vbc, err := parseDecimal(icms00.VBC)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:    "ICMS00.vBC",
			Code:     "VALUE_FORMAT_INVALID",
			Message:  "Base de cálculo inválida",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	picms, err := parseDecimal(icms00.PICMS)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:    "ICMS00.pICMS",
			Code:     "VALUE_FORMAT_INVALID",
			Message:  "Alíquota de ICMS inválida",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	vicms, err := parseDecimal(icms00.VICMS)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:    "ICMS00.vICMS",
			Code:     "VALUE_FORMAT_INVALID",
			Message:  "Valor de ICMS inválido",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Cross-validation: vICMS should equal vBC * pICMS / 100
	if err == nil {
		expectedVICMS := vbc * picms / 100
		tolerance := 0.01 // Allow 1 cent tolerance

		if abs(vicms-expectedVICMS) > tolerance {
			errors = append(errors, ValidationError{
				Field:    "ICMS00.vICMS",
				Code:     "VALUE_CALCULATION_ERROR",
				Message:  fmt.Sprintf("Valor de ICMS inconsistente. Esperado: %.2f, Informado: %.2f", expectedVICMS, vicms),
				Severity: "ERROR",
				Rule:     "NT2019.001",
			})
		}
	}

	return errors
}

// validateICMS10 validates ICMS10 (taxed with ST)
func (tv *TaxValidator) validateICMS10(icms10 *ICMS10, prod Produto) []ValidationError {
	var errors []ValidationError

	// Validate origin
	if err := tv.validateOrigin(icms10.Orig); err != nil {
		errors = append(errors, ValidationError{
			Field:    "ICMS10.orig",
			Code:     "ORIGIN_INVALID",
			Message:  err.Error(),
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate CST
	if icms10.CST != "10" {
		errors = append(errors, ValidationError{
			Field:    "ICMS10.CST",
			Code:     "CST_INVALID",
			Message:  "CST deve ser '10' para ICMS10",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate ST modality
	if !tv.isValidModBCST(icms10.ModBCST) {
		errors = append(errors, ValidationError{
			Field:    "ICMS10.modBCST",
			Code:     "MODBCST_INVALID",
			Message:  "Modalidade de base de cálculo ST inválida",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Additional validations for ST calculations would go here...

	return errors
}

// validateICMS20 validates ICMS20 (reduced tax base)
func (tv *TaxValidator) validateICMS20(icms20 *ICMS20, prod Produto) []ValidationError {
	var errors []ValidationError

	// Validate reduction percentage
	predbc, err := parseDecimal(icms20.PRedBC)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:    "ICMS20.pRedBC",
			Code:     "VALUE_FORMAT_INVALID",
			Message:  "Percentual de redução inválido",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	} else if predbc <= 0 || predbc >= 100 {
		errors = append(errors, ValidationError{
			Field:    "ICMS20.pRedBC",
			Code:     "VALUE_RANGE_INVALID",
			Message:  "Percentual de redução deve estar entre 0 e 100",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	return errors
}

// validateIPIAdvanced performs advanced IPI validation
func (tv *TaxValidator) validateIPIAdvanced(ipi *IPI, prod Produto) []ValidationError {
	var errors []ValidationError

	// Validate framework code
	if err := tv.validateIPIFramework(ipi.CEnq); err != nil {
		errors = append(errors, ValidationError{
			Field:    "IPI.cEnq",
			Code:     "FRAMEWORK_INVALID",
			Message:  err.Error(),
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate modality exclusivity
	if ipi.IPITrib != nil && ipi.IPINT != nil {
		errors = append(errors, ValidationError{
			Field:    "IPI",
			Code:     "IPI_MULTIPLE_MODALITIES",
			Message:  "IPITrib e IPINT são mutuamente exclusivos",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	if ipi.IPITrib == nil && ipi.IPINT == nil {
		errors = append(errors, ValidationError{
			Field:    "IPI",
			Code:     "IPI_NO_MODALITY",
			Message:  "IPITrib ou IPINT deve ser informado",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	return errors
}

// validatePISAdvanced performs advanced PIS validation
func (tv *TaxValidator) validatePISAdvanced(pis *PIS, prod Produto) []ValidationError {
	var errors []ValidationError

	// Count active modalities
	activeModalities := tv.countActivePISModalities(pis)

	if activeModalities == 0 {
		errors = append(errors, ValidationError{
			Field:    "PIS",
			Code:     "PIS_NO_MODALITY",
			Message:  "Nenhuma modalidade de PIS informada",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	} else if activeModalities > 1 {
		errors = append(errors, ValidationError{
			Field:    "PIS",
			Code:     "PIS_MULTIPLE_MODALITIES",
			Message:  "Apenas uma modalidade de PIS deve ser informada",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	return errors
}

// validateCOFINSAdvanced performs advanced COFINS validation
func (tv *TaxValidator) validateCOFINSAdvanced(cofins *COFINS, prod Produto) []ValidationError {
	var errors []ValidationError

	// Count active modalities
	activeModalities := tv.countActiveCOFINSModalities(cofins)

	if activeModalities == 0 {
		errors = append(errors, ValidationError{
			Field:    "COFINS",
			Code:     "COFINS_NO_MODALITY",
			Message:  "Nenhuma modalidade de COFINS informada",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	} else if activeModalities > 1 {
		errors = append(errors, ValidationError{
			Field:    "COFINS",
			Code:     "COFINS_MULTIPLE_MODALITIES",
			Message:  "Apenas uma modalidade de COFINS deve ser informada",
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	return errors
}

// validateISSQNAdvanced performs advanced ISSQN validation
func (tv *TaxValidator) validateISSQNAdvanced(issqn *ISSQN, prod Produto) []ValidationError {
	var errors []ValidationError

	// Validate municipality code
	if err := tv.validateMunicipalityCode(issqn.CMunFG); err != nil {
		errors = append(errors, ValidationError{
			Field:    "ISSQN.cMunFG",
			Code:     "MUNICIPALITY_INVALID",
			Message:  err.Error(),
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	// Validate service list code
	if err := tv.validateServiceListCode(issqn.CListServ); err != nil {
		errors = append(errors, ValidationError{
			Field:    "ISSQN.cListServ",
			Code:     "SERVICE_LIST_INVALID",
			Message:  err.Error(),
			Severity: "ERROR",
			Rule:     "NT2019.001",
		})
	}

	return errors
}

// validateCrossTaxRules validates rules that apply across different taxes
func (tv *TaxValidator) validateCrossTaxRules(item *Item) []ValidationError {
	var errors []ValidationError

	// Rule: If PIS is taxed, COFINS should also be taxed (and vice versa)
	pisIsTaxed := tv.isPISTaxed(item.Imposto.PIS)
	cofinsIsTaxed := tv.isCOFINSTaxed(item.Imposto.COFINS)

	if pisIsTaxed != cofinsIsTaxed {
		errors = append(errors, ValidationError{
			Field:    "imposto",
			Code:     "PIS_COFINS_INCONSISTENT",
			Message:  "PIS e COFINS devem ter a mesma situação tributária",
			Severity: "WARNING",
			Rule:     "NT2019.001",
		})
	}

	// Rule: ISSQN and ICMS are generally mutually exclusive for services
	if item.Imposto.ISSQN != nil && tv.isICMSTaxed(item.Imposto.ICMS) {
		errors = append(errors, ValidationError{
			Field:    "imposto",
			Code:     "ISSQN_ICMS_CONFLICT",
			Message:  "ISSQN e ICMS tributado são geralmente mutuamente exclusivos",
			Severity: "WARNING",
			Rule:     "NT2019.001",
		})
	}

	return errors
}

// Helper methods

func (tv *TaxValidator) validateOrigin(orig string) error {
	validOrigins := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
	for _, validOrig := range validOrigins {
		if orig == validOrig {
			return nil
		}
	}
	return fmt.Errorf("origem inválida: %s", orig)
}

func (tv *TaxValidator) isValidModBC(modBC string) bool {
	validModes := []string{"0", "1", "2", "3"}
	for _, validMode := range validModes {
		if modBC == validMode {
			return true
		}
	}
	return false
}

func (tv *TaxValidator) isValidModBCST(modBCST string) bool {
	validModes := []string{"0", "1", "2", "3", "4", "5", "6"}
	for _, validMode := range validModes {
		if modBCST == validMode {
			return true
		}
	}
	return false
}

func (tv *TaxValidator) validateIPIFramework(cenq string) error {
	if len(cenq) != 3 {
		return fmt.Errorf("código de enquadramento deve ter 3 dígitos")
	}

	if !regexp.MustCompile(`^\d{3}$`).MatchString(cenq) {
		return fmt.Errorf("código de enquadramento deve conter apenas dígitos")
	}

	return nil
}

func (tv *TaxValidator) validateDecimalField(value, fieldName string) error {
	if _, err := parseDecimal(value); err != nil {
		return fmt.Errorf("%s tem formato inválido: %v", fieldName, err)
	}
	return nil
}

func (tv *TaxValidator) validateMunicipalityCode(code string) error {
	if len(code) != 7 {
		return fmt.Errorf("código do município deve ter 7 dígitos")
	}

	if !regexp.MustCompile(`^\d{7}$`).MatchString(code) {
		return fmt.Errorf("código do município deve conter apenas dígitos")
	}

	return nil
}

func (tv *TaxValidator) validateServiceListCode(code string) error {
	if len(code) < 2 || len(code) > 5 {
		return fmt.Errorf("código da lista de serviços deve ter entre 2 e 5 caracteres")
	}

	return nil
}

func (tv *TaxValidator) countActiveICMSModalities(icms *ICMS) int {
	count := 0
	if icms.ICMS00 != nil {
		count++
	}
	if icms.ICMS10 != nil {
		count++
	}
	if icms.ICMS20 != nil {
		count++
	}
	if icms.ICMS30 != nil {
		count++
	}
	if icms.ICMS40 != nil {
		count++
	}
	if icms.ICMS41 != nil {
		count++
	}
	if icms.ICMS50 != nil {
		count++
	}
	if icms.ICMS51 != nil {
		count++
	}
	if icms.ICMS60 != nil {
		count++
	}
	if icms.ICMS70 != nil {
		count++
	}
	if icms.ICMS90 != nil {
		count++
	}
	if icms.ICMSPart != nil {
		count++
	}
	if icms.ICMSST != nil {
		count++
	}
	// Simples Nacional modalities
	if icms.ICMSSN101 != nil {
		count++
	}
	if icms.ICMSSN102 != nil {
		count++
	}
	if icms.ICMSSN103 != nil {
		count++
	}
	if icms.ICMSSN201 != nil {
		count++
	}
	if icms.ICMSSN202 != nil {
		count++
	}
	if icms.ICMSSN203 != nil {
		count++
	}
	if icms.ICMSSN300 != nil {
		count++
	}
	if icms.ICMSSN400 != nil {
		count++
	}
	if icms.ICMSSN500 != nil {
		count++
	}
	if icms.ICMSSN900 != nil {
		count++
	}
	return count
}

func (tv *TaxValidator) countActivePISModalities(pis *PIS) int {
	count := 0
	if pis.PISAliq != nil {
		count++
	}
	if pis.PISQtde != nil {
		count++
	}
	if pis.PISNT != nil {
		count++
	}
	if pis.PISOutr != nil {
		count++
	}
	return count
}

func (tv *TaxValidator) countActiveCOFINSModalities(cofins *COFINS) int {
	count := 0
	if cofins.COFINSAliq != nil {
		count++
	}
	if cofins.COFINSQtde != nil {
		count++
	}
	if cofins.COFINSNT != nil {
		count++
	}
	if cofins.COFINSOutr != nil {
		count++
	}
	return count
}

func (tv *TaxValidator) isPISTaxed(pis *PIS) bool {
	if pis == nil {
		return false
	}
	return pis.PISAliq != nil || pis.PISQtde != nil || pis.PISOutr != nil
}

func (tv *TaxValidator) isCOFINSTaxed(cofins *COFINS) bool {
	if cofins == nil {
		return false
	}
	return cofins.COFINSAliq != nil || cofins.COFINSQtde != nil || cofins.COFINSOutr != nil
}

func (tv *TaxValidator) isICMSTaxed(icms *ICMS) bool {
	if icms == nil {
		return false
	}
	// ICMS is considered taxed if it's not in exempt/non-taxed modalities
	return icms.ICMS00 != nil || icms.ICMS10 != nil || icms.ICMS20 != nil ||
		icms.ICMS51 != nil || icms.ICMS70 != nil || icms.ICMS90 != nil ||
		icms.ICMSPart != nil || icms.ICMSSN101 != nil || icms.ICMSSN201 != nil || icms.ICMSSN900 != nil
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
