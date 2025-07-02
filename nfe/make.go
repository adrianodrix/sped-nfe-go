// Package nfe provides the main NFe generation and assembly functionality.
package nfe

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"sync"
	"time"
)

// Make is the main NFe generator following the PHP Make.php pattern
type Make struct {
	mu                  sync.RWMutex
	version             string                // Layout version (4.00)
	model               DocumentModel         // NFe model (55 or 65)
	environment         NFEEnvironment        // Environment (1=prod, 2=test)
	accessKey           *AccessKey            // NFe access key
	xml                 string                // Generated XML
	errors              []string              // Validation errors
	
	// NFe structure
	nfe                 *NFe                  // Root element
	infNFe              *InfNFe               // NFe information
	identification      *Identificacao        // Identification
	issuer              *Emitente             // Issuer
	recipient           *Destinatario         // Recipient
	items               []Item                // Items/Products
	transport           *Transporte           // Transport
	payment             []Pagamento           // Payment information
	additionalInfo      *InfAdicionais        // Additional information
	
	// Automatic totalizers
	totals              *Totalizer            // Automatic totals calculator
	
	// Configuration
	checkGTIN           bool                  // Validate GTIN codes
	removeAccents       bool                  // Remove accents from text
	roundValues         bool                  // Round monetary values
	autoCalculate       bool                  // Auto-calculate totals
}

// Totalizer handles automatic calculation of NFe totals
type Totalizer struct {
	mu                  sync.RWMutex
	productValue        float64               // vProd
	freightValue        float64               // vFrete
	insuranceValue      float64               // vSeg
	discountValue       float64               // vDesc
	otherValue          float64               // vOutro
	icmsBaseValue       float64               // vBC
	icmsValue           float64               // vICMS
	icmsSTBaseValue     float64               // vBCST
	icmsSTValue         float64               // vST
	ipiValue            float64               // vIPI
	pisValue            float64               // vPIS
	cofinsValue         float64               // vCOFINS
	importTaxValue      float64               // vII
	icmsReliefValue     float64               // vICMSDeson
	fcpValue            float64               // vFCP
	fcpSTValue          float64               // vFCPST
	fcpSTRetValue       float64               // vFCPSTRet
	totalValue          float64               // vNF
}

// NewMake creates a new NFe generator
func NewMake() *Make {
	return &Make{
		version:       LayoutVersion,
		model:         ModelNFe,
		environment:   EnvironmentTesting,
		errors:        make([]string, 0),
		items:         make([]Item, 0),
		payment:       make([]Pagamento, 0),
		totals:        NewTotalizer(),
		checkGTIN:     true,
		removeAccents: false,
		roundValues:   true,
		autoCalculate: true,
	}
}

// NewTotalizer creates a new totalizer
func NewTotalizer() *Totalizer {
	return &Totalizer{}
}

// Configuration methods

// SetVersion sets the layout version
func (m *Make) SetVersion(version string) *Make {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.version = version
	return m
}

// SetModel sets the document model
func (m *Make) SetModel(model DocumentModel) *Make {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.model = model
	return m
}

// SetEnvironment sets the environment
func (m *Make) SetEnvironment(env NFEEnvironment) *Make {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.environment = env
	return m
}

// SetCheckGTIN enables/disables GTIN validation
func (m *Make) SetCheckGTIN(check bool) *Make {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkGTIN = check
	return m
}

// SetRemoveAccents enables/disables accent removal
func (m *Make) SetRemoveAccents(remove bool) *Make {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.removeAccents = remove
	return m
}

// SetRoundValues enables/disables value rounding
func (m *Make) SetRoundValues(round bool) *Make {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.roundValues = round
	return m
}

// SetAutoCalculate enables/disables automatic totals calculation
func (m *Make) SetAutoCalculate(auto bool) *Make {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.autoCalculate = auto
	return m
}

// Tag methods following PHP Make.php pattern

// TagInfNFe sets the NFe information root element
func (m *Make) TagInfNFe(id string, version string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if id == "" {
		return fmt.Errorf("NFe ID is required")
	}
	
	if version == "" {
		version = m.version
	}
	
	m.infNFe = &InfNFe{
		ID:     id,
		Versao: version,
	}
	
	return nil
}

// TagIde sets the identification information
func (m *Make) TagIde(ide *Identificacao) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if ide == nil {
		return fmt.Errorf("identification data is required")
	}
	
	// Validate required fields
	if err := m.validateIdentification(ide); err != nil {
		return err
	}
	
	// Auto-generate cNF if empty
	if ide.CNF == "" {
		ide.CNF = GenerateRandomCode(8)
	}
	
	// Ensure cNF ≠ nNF (NT2019.001)
	if len(ide.NNF) >= 8 && ide.CNF == ide.NNF[len(ide.NNF)-8:] {
		ide.CNF = GenerateRandomCode(8)
	}
	
	// Set default values
	if ide.TpAmb == "" {
		ide.TpAmb = m.environment.String()
	}
	
	if ide.Mod == "" {
		ide.Mod = m.model.String()
	}
	
	// Set issue date/time if empty
	if ide.DhEmi == "" {
		ide.DhEmi = FormatDateTime(time.Now())
	}
	
	m.identification = ide
	
	return nil
}

// TagEmit sets the issuer information
func (m *Make) TagEmit(emit *Emitente) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if emit == nil {
		return fmt.Errorf("issuer data is required")
	}
	
	// Validate required fields
	if err := m.validateIssuer(emit); err != nil {
		return err
	}
	
	// Normalize text fields
	emit.XNome = m.normalizeText(emit.XNome, 60)
	emit.XFant = m.normalizeText(emit.XFant, 60)
	
	m.issuer = emit
	
	return nil
}

// TagDest sets the recipient information
func (m *Make) TagDest(dest *Destinatario) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if dest == nil && m.model == ModelNFe {
		return fmt.Errorf("recipient data is required for NFe")
	}
	
	if dest != nil {
		// Validate recipient data
		if err := m.validateRecipient(dest); err != nil {
			return err
		}
		
		// Normalize text fields
		dest.XNome = m.normalizeText(dest.XNome, 60)
	}
	
	m.recipient = dest
	
	return nil
}

// TagDet adds an item to the NFe
func (m *Make) TagDet(item *Item) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if item == nil {
		return fmt.Errorf("item data is required")
	}
	
	// Validate item
	if err := m.validateItem(item); err != nil {
		return err
	}
	
	// Set item sequence number
	item.NItem = strconv.Itoa(len(m.items) + 1)
	
	// Normalize text fields
	item.Prod.XProd = m.normalizeText(item.Prod.XProd, 120)
	item.InfAdProd = m.normalizeText(item.InfAdProd, 500)
	
	// Validate GTIN if enabled
	if m.checkGTIN {
		if !IsValidGTIN(item.Prod.CEAN) {
			return fmt.Errorf("invalid GTIN code: %s", item.Prod.CEAN)
		}
		if !IsValidGTIN(item.Prod.CEANTrib) {
			return fmt.Errorf("invalid tributary GTIN code: %s", item.Prod.CEANTrib)
		}
	}
	
	// Round values if enabled
	if m.roundValues {
		m.roundItemValues(&item.Prod)
	}
	
	// Add to items list
	m.items = append(m.items, *item)
	
	// Update totals if auto-calculate is enabled
	if m.autoCalculate {
		m.updateTotalsWithItem(item)
	}
	
	return nil
}

// TagTransp sets transport information
func (m *Make) TagTransp(transp *Transporte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if transp == nil {
		return fmt.Errorf("transport data is required")
	}
	
	// Validate transport
	if err := ValidateTransport(transp); err != nil {
		return err
	}
	
	m.transport = transp
	
	return nil
}

// TagPag adds payment information
func (m *Make) TagPag(pag *Pagamento) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if pag == nil {
		return fmt.Errorf("payment data is required")
	}
	
	// Validate payment
	if err := ValidatePayment(pag); err != nil {
		return err
	}
	
	m.payment = append(m.payment, *pag)
	
	return nil
}

// TagInfAdic sets additional information
func (m *Make) TagInfAdic(infAdic *InfAdicionais) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if infAdic != nil {
		// Normalize text fields
		infAdic.InfAdFisco = m.normalizeText(infAdic.InfAdFisco, 2000)
		infAdic.InfCpl = m.normalizeText(infAdic.InfCpl, 5000)
	}
	
	m.additionalInfo = infAdic
	
	return nil
}

// Build methods

// BuildNFe assembles the complete NFe structure
func (m *Make) BuildNFe() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Validate required components
	if err := m.validateRequired(); err != nil {
		return err
	}
	
	// Generate access key if not set
	if m.accessKey == nil {
		if err := m.generateAccessKey(); err != nil {
			return err
		}
	}
	
	// Calculate totals if auto-calculate is enabled
	if m.autoCalculate {
		if err := m.calculateTotals(); err != nil {
			return err
		}
	}
	
	// Build InfNFe structure
	if err := m.buildInfNFe(); err != nil {
		return err
	}
	
	// Create NFe root element
	m.nfe = &NFe{
		Xmlns:  NFENamespace,
		InfNFe: *m.infNFe,
	}
	
	return nil
}

// GetXML generates and returns the complete NFe XML
func (m *Make) GetXML() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.xml != "" {
		return m.xml, nil
	}
	
	// Build NFe if not built yet
	if m.nfe == nil {
		if err := m.BuildNFe(); err != nil {
			return "", err
		}
	}
	
	// Generate XML
	xmlData, err := xml.MarshalIndent(m.nfe, "", "")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NFe XML: %v", err)
	}
	
	// Add XML declaration
	xmlHeader := `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
	m.xml = xmlHeader + string(xmlData)
	
	return m.xml, nil
}

// GetAccessKey returns the NFe access key
func (m *Make) GetAccessKey() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.accessKey == nil {
		return ""
	}
	
	return m.accessKey.GetKey()
}

// GetErrors returns validation errors
func (m *Make) GetErrors() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return m.errors
}

// HasErrors checks if there are validation errors
func (m *Make) HasErrors() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return len(m.errors) > 0
}

// AddError adds a validation error
func (m *Make) AddError(err string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors = append(m.errors, err)
}

// ClearErrors clears all validation errors
func (m *Make) ClearErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.errors = make([]string, 0)
}

// Helper methods

func (m *Make) normalizeText(text string, maxLength int) string {
	return NormalizeString(text, m.removeAccents, maxLength)
}

func (m *Make) validateRequired() error {
	if m.identification == nil {
		return fmt.Errorf("identification is required")
	}
	
	if m.issuer == nil {
		return fmt.Errorf("issuer is required")
	}
	
	if m.model == ModelNFe && m.recipient == nil {
		return fmt.Errorf("recipient is required for NFe")
	}
	
	if len(m.items) == 0 {
		return fmt.Errorf("at least one item is required")
	}
	
	if m.transport == nil {
		return fmt.Errorf("transport information is required")
	}
	
	return nil
}

func (m *Make) generateAccessKey() error {
	if m.identification == nil || m.issuer == nil {
		return fmt.Errorf("identification and issuer are required to generate access key")
	}
	
	// Parse issue date
	issueDate := time.Now()
	if m.identification.DhEmi != "" {
		if parsed, err := ParseDateTime(m.identification.DhEmi); err == nil {
			issueDate = parsed
		}
	}
	
	// Generate access key
	accessKey, err := NewAccessKeyBuilder().
		State(m.identification.CUF).
		Document(m.issuer.CNPJ).
		Model(m.model).
		Series(m.parseIntField(m.identification.Serie)).
		Number(m.parseIntField(m.identification.NNF)).
		EmissionType(EmissionNormal).
		RandomCode(m.identification.CNF).
		IssueDateTime(issueDate).
		Build()
	
	if err != nil {
		return fmt.Errorf("failed to generate access key: %v", err)
	}
	
	m.accessKey = accessKey
	
	// Update identification with generated values
	m.identification.CNF = accessKey.RandomCode
	m.identification.CDV = accessKey.CheckDigit
	
	// Update InfNFe ID if set
	if m.infNFe != nil {
		m.infNFe.ID = "NFe" + accessKey.GetKey()
	}
	
	return nil
}

func (m *Make) parseIntField(field string) int {
	if value, err := strconv.Atoi(field); err == nil {
		return value
	}
	return 0
}

func (m *Make) buildInfNFe() error {
	if m.infNFe == nil {
		// Create InfNFe with access key ID
		keyID := ""
		if m.accessKey != nil {
			keyID = "NFe" + m.accessKey.GetKey()
		}
		
		m.infNFe = &InfNFe{
			ID:     keyID,
			Versao: m.version,
		}
	}
	
	// Set required components
	m.infNFe.Ide = *m.identification
	m.infNFe.Emit = *m.issuer
	
	if m.recipient != nil {
		m.infNFe.Dest = m.recipient
	}
	
	m.infNFe.Det = m.items
	
	if m.transport != nil {
		m.infNFe.Transp = *m.transport
	}
	
	if len(m.payment) > 0 {
		m.infNFe.Pag = m.payment
	}
	
	if m.additionalInfo != nil {
		m.infNFe.InfAdic = m.additionalInfo
	}
	
	// Build totals
	total, err := m.buildTotal()
	if err != nil {
		return err
	}
	
	m.infNFe.Total = *total
	
	return nil
}

func (m *Make) buildTotal() (*Total, error) {
	// Calculate ICMS totals
	icmsTot := &ICMSTotal{
		VBC:            FormatCurrency(m.totals.icmsBaseValue),
		VICMS:          FormatCurrency(m.totals.icmsValue),
		VICMSDeson:     FormatCurrency(m.totals.icmsReliefValue),
		VFCP:           FormatCurrency(m.totals.fcpValue),
		VBCST:          FormatCurrency(m.totals.icmsSTBaseValue),
		VST:            FormatCurrency(m.totals.icmsSTValue),
		VFCPST:         FormatCurrency(m.totals.fcpSTValue),
		VFCPSTRet:      FormatCurrency(m.totals.fcpSTRetValue),
		VProd:          FormatCurrency(m.totals.productValue),
		VFrete:         FormatCurrency(m.totals.freightValue),
		VSeg:           FormatCurrency(m.totals.insuranceValue),
		VDesc:          FormatCurrency(m.totals.discountValue),
		VII:            FormatCurrency(m.totals.importTaxValue),
		VIPI:           FormatCurrency(m.totals.ipiValue),
		VIPIDevol:      "0.00",
		VPIS:           FormatCurrency(m.totals.pisValue),
		VCOFINS:        FormatCurrency(m.totals.cofinsValue),
		VOutro:         FormatCurrency(m.totals.otherValue),
		VNF:            FormatCurrency(m.totals.totalValue),
	}
	
	return &Total{
		ICMSTot: *icmsTot,
	}, nil
}

// Validation methods - will be implemented in validation.go

func (m *Make) validateIdentification(ide *Identificacao) error {
	// Basic validation - detailed validation in validation.go
	if ide.CUF == "" {
		return fmt.Errorf("state code (cUF) is required")
	}
	if ide.NatOp == "" {
		return fmt.Errorf("operation nature is required")
	}
	if ide.Serie == "" {
		return fmt.Errorf("series is required")
	}
	if ide.NNF == "" {
		return fmt.Errorf("document number is required")
	}
	return nil
}

func (m *Make) validateIssuer(emit *Emitente) error {
	if emit.CNPJ == "" && emit.CPF == "" {
		return fmt.Errorf("CNPJ or CPF is required")
	}
	if emit.XNome == "" {
		return fmt.Errorf("company name is required")
	}
	if emit.IE == "" {
		return fmt.Errorf("state registration (IE) is required")
	}
	return nil
}

func (m *Make) validateRecipient(dest *Destinatario) error {
	if m.model == ModelNFe {
		if dest.CNPJ == "" && dest.CPF == "" && dest.IdEstrangeiro == "" {
			return fmt.Errorf("CNPJ, CPF or foreign ID is required for NFe")
		}
	}
	return nil
}

func (m *Make) validateItem(item *Item) error {
	if item.Prod.CProd == "" {
		return fmt.Errorf("product code is required")
	}
	if item.Prod.XProd == "" {
		return fmt.Errorf("product description is required")
	}
	if item.Prod.NCM == "" {
		return fmt.Errorf("NCM code is required")
	}
	if item.Prod.CFOP == "" {
		return fmt.Errorf("CFOP code is required")
	}
	return nil
}

func (m *Make) roundItemValues(prod *Produto) {
	if vProd, err := ParseValue(prod.VProd); err == nil {
		prod.VProd = FormatCurrency(RoundCurrency(vProd))
	}
	if vUnCom, err := ParseValue(prod.VUnCom); err == nil {
		prod.VUnCom = FormatDecimal(vUnCom, 4)
	}
	if vUnTrib, err := ParseValue(prod.VUnTrib); err == nil {
		prod.VUnTrib = FormatDecimal(vUnTrib, 4)
	}
}

func (m *Make) updateTotalsWithItem(item *Item) {
	// Add item values to totals
	if vProd, err := ParseValue(item.Prod.VProd); err == nil && item.Prod.IndTot == "1" {
		m.totals.productValue += vProd
	}
	
	if vFrete, err := ParseValue(item.Prod.VFrete); err == nil {
		m.totals.freightValue += vFrete
	}
	
	if vSeg, err := ParseValue(item.Prod.VSeg); err == nil {
		m.totals.insuranceValue += vSeg
	}
	
	if vDesc, err := ParseValue(item.Prod.VDesc); err == nil {
		m.totals.discountValue += vDesc
	}
	
	if vOutro, err := ParseValue(item.Prod.VOutro); err == nil {
		m.totals.otherValue += vOutro
	}
	
	// Add tax values
	m.updateTotalsWithTaxes(&item.Imposto)
}

func (m *Make) updateTotalsWithTaxes(imposto *Imposto) {
	// Update ICMS totals
	if imposto.ICMS != nil {
		m.updateTotalsWithICMS(imposto.ICMS)
	}
	
	// Update IPI totals
	if imposto.IPI != nil && imposto.IPI.IPITrib != nil {
		if vIPI, err := ParseValue(imposto.IPI.IPITrib.VIPI); err == nil {
			m.totals.ipiValue += vIPI
		}
	}
	
	// Update PIS totals
	if imposto.PIS != nil {
		if imposto.PIS.PISAliq != nil {
			if vPIS, err := ParseValue(imposto.PIS.PISAliq.VPIS); err == nil {
				m.totals.pisValue += vPIS
			}
		}
		if imposto.PIS.PISQtde != nil {
			if vPIS, err := ParseValue(imposto.PIS.PISQtde.VPIS); err == nil {
				m.totals.pisValue += vPIS
			}
		}
	}
	
	// Update COFINS totals
	if imposto.COFINS != nil {
		if imposto.COFINS.COFINSAliq != nil {
			if vCOFINS, err := ParseValue(imposto.COFINS.COFINSAliq.VCOFINS); err == nil {
				m.totals.cofinsValue += vCOFINS
			}
		}
		if imposto.COFINS.COFINSQtde != nil {
			if vCOFINS, err := ParseValue(imposto.COFINS.COFINSQtde.VCOFINS); err == nil {
				m.totals.cofinsValue += vCOFINS
			}
		}
	}
	
	// Update Import Tax totals
	if imposto.II != nil {
		if vII, err := ParseValue(imposto.II.VII); err == nil {
			m.totals.importTaxValue += vII
		}
	}
}

func (m *Make) updateTotalsWithICMS(icms *ICMS) {
	// Handle different ICMS types
	if icms.ICMS00 != nil {
		if vBC, err := ParseValue(icms.ICMS00.VBC); err == nil {
			m.totals.icmsBaseValue += vBC
		}
		if vICMS, err := ParseValue(icms.ICMS00.VICMS); err == nil {
			m.totals.icmsValue += vICMS
		}
	}
	// Add other ICMS types as needed...
}

func (m *Make) calculateTotals() error {
	m.totals.mu.Lock()
	defer m.totals.mu.Unlock()
	
	// Calculate total NFe value
	m.totals.totalValue = m.totals.productValue - 
		m.totals.discountValue - 
		m.totals.icmsReliefValue + 
		m.totals.icmsSTValue + 
		m.totals.freightValue + 
		m.totals.insuranceValue + 
		m.totals.otherValue + 
		m.totals.importTaxValue + 
		m.totals.ipiValue
	
	// Round all values
	if m.roundValues {
		m.totals.productValue = RoundCurrency(m.totals.productValue)
		m.totals.freightValue = RoundCurrency(m.totals.freightValue)
		m.totals.insuranceValue = RoundCurrency(m.totals.insuranceValue)
		m.totals.discountValue = RoundCurrency(m.totals.discountValue)
		m.totals.otherValue = RoundCurrency(m.totals.otherValue)
		m.totals.icmsBaseValue = RoundCurrency(m.totals.icmsBaseValue)
		m.totals.icmsValue = RoundCurrency(m.totals.icmsValue)
		m.totals.icmsSTBaseValue = RoundCurrency(m.totals.icmsSTBaseValue)
		m.totals.icmsSTValue = RoundCurrency(m.totals.icmsSTValue)
		m.totals.ipiValue = RoundCurrency(m.totals.ipiValue)
		m.totals.pisValue = RoundCurrency(m.totals.pisValue)
		m.totals.cofinsValue = RoundCurrency(m.totals.cofinsValue)
		m.totals.importTaxValue = RoundCurrency(m.totals.importTaxValue)
		m.totals.totalValue = RoundCurrency(m.totals.totalValue)
	}
	
	return nil
}