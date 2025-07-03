// Package nfe provides payment structures for NFe documents.
package nfe

import (
	"encoding/xml"
	"fmt"
)

// Pagamento represents the payment group for NFe (container for payment details)
type Pagamento struct {
	XMLName xml.Name `xml:"pag"`
	DetPag  []DetPag `xml:"detPag" validate:"required,min=1,max=100"` // Payment details (1 to 100 allowed)
	VTroco  string   `xml:"vTroco,omitempty" validate:"omitempty"`     // Change value
}

// DetPag represents individual payment details within the payment group
type DetPag struct {
	XMLName   xml.Name `xml:"detPag"`
	IndPag    string   `xml:"indPag,omitempty" validate:"omitempty,oneof=0 1"`                                   // Payment indicator (0=cash, 1=term)
	TPag      string   `xml:"tPag" validate:"required,oneof=01 02 03 04 05 10 11 12 13 14 15 16 17 18 19 90 99"` // Payment type
	XPag      string   `xml:"xPag,omitempty" validate:"omitempty,min=2,max=60"`                                  // Payment description
	VPag      string   `xml:"vPag" validate:"required"`                                                          // Payment value
	DPag      string   `xml:"dPag,omitempty" validate:"omitempty"`                                               // Payment date (YYYY-MM-DD format)
	CNPJPag   string   `xml:"CNPJPag,omitempty" validate:"omitempty,len=14"`                                     // Payment company CNPJ
	UFPag     string   `xml:"UFPag,omitempty" validate:"omitempty,len=2"`                                        // Payment UF
	Card      *Card    `xml:"card,omitempty"`                                                                    // Card payment information
}

// Card represents credit/debit card payment information
type Card struct {
	XMLName   xml.Name `xml:"card"`
	TpIntegra string   `xml:"tpIntegra" validate:"required,oneof=1 2"`                                           // Integration type
	CNPJ      string   `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`                                        // Card operator CNPJ
	TBand     string   `xml:"tBand,omitempty" validate:"omitempty,oneof=01 02 03 04 05 06 07 08 09 10 11 12 99"` // Card brand
	CAut      string   `xml:"cAut,omitempty" validate:"omitempty,min=1,max=20"`                                  // Authorization code
	CNPJReceb string   `xml:"CNPJReceb,omitempty" validate:"omitempty,len=14"`                                   // Receiver CNPJ
	IdTermPag string   `xml:"idTermPag,omitempty" validate:"omitempty,min=1,max=8"`                              // Payment terminal ID
}

// PaymentType represents payment type codes
type PaymentType string

const (
	PaymentTypeMoney          PaymentType = "01" // Dinheiro
	PaymentTypeCheck          PaymentType = "02" // Cheque
	PaymentTypeCreditCard     PaymentType = "03" // Cartão de Crédito
	PaymentTypeDebitCard      PaymentType = "04" // Cartão de Débito
	PaymentTypeStore          PaymentType = "05" // Crédito Loja
	PaymentTypeFood           PaymentType = "10" // Vale Alimentação
	PaymentTypeMeal           PaymentType = "11" // Vale Refeição
	PaymentTypePresent        PaymentType = "12" // Vale Presente
	PaymentTypeFuel           PaymentType = "13" // Vale Combustível
	PaymentTypeDuplicate      PaymentType = "14" // Duplicata Mercantil
	PaymentTypeBoletoBancario PaymentType = "15" // Boleto Bancário
	PaymentTypeDeposito       PaymentType = "16" // Depósito Bancário
	PaymentTypeInstant        PaymentType = "17" // Pagamento Instantâneo (PIX)
	PaymentTypeBankTransfer   PaymentType = "18" // Transferência bancária
	PaymentTypeLoyalty        PaymentType = "19" // Programa de fidelidade
	PaymentTypeWithoutPayment PaymentType = "90" // Sem pagamento
	PaymentTypeOther          PaymentType = "99" // Outros
)

// String returns the string representation of PaymentType
func (pt PaymentType) String() string {
	return string(pt)
}

// Description returns the description of the payment type
func (pt PaymentType) Description() string {
	switch pt {
	case PaymentTypeMoney:
		return "Dinheiro"
	case PaymentTypeCheck:
		return "Cheque"
	case PaymentTypeCreditCard:
		return "Cartão de Crédito"
	case PaymentTypeDebitCard:
		return "Cartão de Débito"
	case PaymentTypeStore:
		return "Crédito Loja"
	case PaymentTypeFood:
		return "Vale Alimentação"
	case PaymentTypeMeal:
		return "Vale Refeição"
	case PaymentTypePresent:
		return "Vale Presente"
	case PaymentTypeFuel:
		return "Vale Combustível"
	case PaymentTypeDuplicate:
		return "Duplicata Mercantil"
	case PaymentTypeBoletoBancario:
		return "Boleto Bancário"
	case PaymentTypeDeposito:
		return "Depósito Bancário"
	case PaymentTypeInstant:
		return "Pagamento Instantâneo (PIX)"
	case PaymentTypeBankTransfer:
		return "Transferência bancária"
	case PaymentTypeLoyalty:
		return "Programa de fidelidade"
	case PaymentTypeWithoutPayment:
		return "Sem pagamento"
	case PaymentTypeOther:
		return "Outros"
	default:
		return "Não identificado"
	}
}

// CardBrand represents credit/debit card brands
type CardBrand string

const (
	CardBrandVisa       CardBrand = "01" // Visa
	CardBrandMastercard CardBrand = "02" // Mastercard
	CardBrandAmex       CardBrand = "03" // American Express
	CardBrandSorocred   CardBrand = "04" // Sorocred
	CardBrandDinersClub CardBrand = "05" // Diners Club
	CardBrandElo        CardBrand = "06" // Elo
	CardBrandHipercard  CardBrand = "07" // Hipercard
	CardBrandAura       CardBrand = "08" // Aura
	CardBrandCabal      CardBrand = "09" // Cabal
	CardBrandAlelo      CardBrand = "10" // Alelo
	CardBrandBanesCard  CardBrand = "11" // BanesCard
	CardBrandCalCard    CardBrand = "12" // CalCard
	CardBrandOther      CardBrand = "99" // Outros
)

// String returns the string representation of CardBrand
func (cb CardBrand) String() string {
	return string(cb)
}

// Description returns the description of the card brand
func (cb CardBrand) Description() string {
	switch cb {
	case CardBrandVisa:
		return "Visa"
	case CardBrandMastercard:
		return "Mastercard"
	case CardBrandAmex:
		return "American Express"
	case CardBrandSorocred:
		return "Sorocred"
	case CardBrandDinersClub:
		return "Diners Club"
	case CardBrandElo:
		return "Elo"
	case CardBrandHipercard:
		return "Hipercard"
	case CardBrandAura:
		return "Aura"
	case CardBrandCabal:
		return "Cabal"
	case CardBrandAlelo:
		return "Alelo"
	case CardBrandBanesCard:
		return "BanesCard"
	case CardBrandCalCard:
		return "CalCard"
	case CardBrandOther:
		return "Outros"
	default:
		return "Não identificado"
	}
}

// IntegrationType represents card integration types
type IntegrationType string

const (
	IntegrationTypeNonIntegrated IntegrationType = "1" // Não integrado
	IntegrationTypeIntegrated    IntegrationType = "2" // Integrado
)

// String returns the string representation of IntegrationType
func (it IntegrationType) String() string {
	return string(it)
}

// Description returns the description of the integration type
func (it IntegrationType) Description() string {
	switch it {
	case IntegrationTypeNonIntegrated:
		return "Não integrado"
	case IntegrationTypeIntegrated:
		return "Integrado"
	default:
		return "Não identificado"
	}
}

// PaymentIndicator represents payment indicators
type PaymentIndicator string

const (
	PaymentIndicatorCash PaymentIndicator = "0" // Pagamento à Vista
	PaymentIndicatorTerm PaymentIndicator = "1" // Pagamento à Prazo
)

// String returns the string representation of PaymentIndicator
func (pi PaymentIndicator) String() string {
	return string(pi)
}

// Description returns the description of the payment indicator
func (pi PaymentIndicator) Description() string {
	switch pi {
	case PaymentIndicatorCash:
		return "Pagamento à Vista"
	case PaymentIndicatorTerm:
		return "Pagamento à Prazo"
	default:
		return "Não identificado"
	}
}

// PaymentBuilder helps build payment information
type PaymentBuilder struct {
	payment *Pagamento
}

// DetPagBuilder helps build individual payment details
type DetPagBuilder struct {
	detPag *DetPag
}

// NewPaymentBuilder creates a new payment builder
func NewPaymentBuilder() *PaymentBuilder {
	return &PaymentBuilder{
		payment: &Pagamento{
			DetPag: make([]DetPag, 0),
		},
	}
}

// NewDetPagBuilder creates a new payment details builder
func NewDetPagBuilder() *DetPagBuilder {
	return &DetPagBuilder{
		detPag: &DetPag{},
	}
}

// AddDetPag adds a payment detail to the payment group
func (pb *PaymentBuilder) AddDetPag(detPag DetPag) *PaymentBuilder {
	pb.payment.DetPag = append(pb.payment.DetPag, detPag)
	return pb
}

// SetVTroco sets the change value
func (pb *PaymentBuilder) SetVTroco(vTroco string) *PaymentBuilder {
	pb.payment.VTroco = vTroco
	return pb
}

// SetIndicator sets the payment indicator (cash/term)
func (dpb *DetPagBuilder) SetIndicator(indicator PaymentIndicator) *DetPagBuilder {
	dpb.detPag.IndPag = indicator.String()
	return dpb
}

// SetType sets the payment type
func (dpb *DetPagBuilder) SetType(paymentType PaymentType) *DetPagBuilder {
	dpb.detPag.TPag = paymentType.String()
	return dpb
}

// SetDescription sets the payment description
func (dpb *DetPagBuilder) SetDescription(description string) *DetPagBuilder {
	dpb.detPag.XPag = description
	return dpb
}

// SetValue sets the payment value
func (dpb *DetPagBuilder) SetValue(value string) *DetPagBuilder {
	dpb.detPag.VPag = value
	return dpb
}

// SetDate sets the payment date (YYYY-MM-DD format)
func (dpb *DetPagBuilder) SetDate(date string) *DetPagBuilder {
	dpb.detPag.DPag = date
	return dpb
}

// SetCNPJ sets the payment company CNPJ
func (dpb *DetPagBuilder) SetCNPJ(cnpj string) *DetPagBuilder {
	dpb.detPag.CNPJPag = cnpj
	return dpb
}

// SetUF sets the payment UF
func (dpb *DetPagBuilder) SetUF(uf string) *DetPagBuilder {
	dpb.detPag.UFPag = uf
	return dpb
}

// SetCard sets card payment information
func (dpb *DetPagBuilder) SetCard(card *Card) *DetPagBuilder {
	dpb.detPag.Card = card
	return dpb
}

// Build returns the constructed payment detail
func (dpb *DetPagBuilder) Build() DetPag {
	return *dpb.detPag
}

// Build returns the constructed payment
func (pb *PaymentBuilder) Build() *Pagamento {
	return pb.payment
}

// CardBuilder helps build card payment information
type CardBuilder struct {
	card *Card
}

// NewCardBuilder creates a new card builder
func NewCardBuilder() *CardBuilder {
	return &CardBuilder{
		card: &Card{},
	}
}

// SetIntegrationType sets the integration type
func (cb *CardBuilder) SetIntegrationType(integrationType IntegrationType) *CardBuilder {
	cb.card.TpIntegra = integrationType.String()
	return cb
}

// SetCNPJ sets the card operator CNPJ
func (cb *CardBuilder) SetCNPJ(cnpj string) *CardBuilder {
	cb.card.CNPJ = cnpj
	return cb
}

// SetBrand sets the card brand
func (cb *CardBuilder) SetBrand(brand CardBrand) *CardBuilder {
	cb.card.TBand = brand.String()
	return cb
}

// SetAuthorizationCode sets the authorization code
func (cb *CardBuilder) SetAuthorizationCode(code string) *CardBuilder {
	cb.card.CAut = code
	return cb
}

// SetReceiverCNPJ sets the receiver CNPJ
func (cb *CardBuilder) SetReceiverCNPJ(cnpj string) *CardBuilder {
	cb.card.CNPJReceb = cnpj
	return cb
}

// SetTerminalID sets the payment terminal ID
func (cb *CardBuilder) SetTerminalID(terminalID string) *CardBuilder {
	cb.card.IdTermPag = terminalID
	return cb
}

// Build returns the constructed card
func (cb *CardBuilder) Build() *Card {
	return cb.card
}

// ValidatePayment validates payment information
func ValidatePayment(payment *Pagamento) error {
	if payment == nil {
		return nil
	}

	// Must have at least one payment detail
	if len(payment.DetPag) == 0 {
		return fmt.Errorf("payment must have at least one detPag")
	}

	// Maximum 100 payment details allowed
	if len(payment.DetPag) > 100 {
		return fmt.Errorf("payment cannot have more than 100 detPag entries")
	}

	// TODO: Implement additional validation logic
	// - Validate payment type codes
	// - Check value format
	// - Validate card information if present
	// - Check CNPJ format

	return nil
}

// CreateCashPayment creates a simple cash payment
func CreateCashPayment(value string) *Pagamento {
	detPag := DetPag{
		TPag: PaymentTypeMoney.String(),
		VPag: value,
	}
	
	return &Pagamento{
		DetPag: []DetPag{detPag},
	}
}

// CreateCardPayment creates a card payment
func CreateCardPayment(value string, card *Card) *Pagamento {
	paymentType := PaymentTypeCreditCard
	if card != nil && card.TpIntegra == IntegrationTypeIntegrated.String() {
		// Logic to determine if it's credit or debit based on integration
		paymentType = PaymentTypeCreditCard
	}

	detPag := DetPag{
		TPag: paymentType.String(),
		VPag: value,
		Card: card,
	}

	return &Pagamento{
		DetPag: []DetPag{detPag},
	}
}

// CreatePIXPayment creates a PIX payment
func CreatePIXPayment(value string) *Pagamento {
	detPag := DetPag{
		TPag: PaymentTypeInstant.String(),
		VPag: value,
		XPag: "PIX",
	}

	return &Pagamento{
		DetPag: []DetPag{detPag},
	}
}

// CreateBoletoPayment creates a boleto payment
func CreateBoletoPayment(value string) *Pagamento {
	detPag := DetPag{
		IndPag: PaymentIndicatorTerm.String(),
		TPag:   PaymentTypeBoletoBancario.String(),
		VPag:   value,
		XPag:   "Boleto Bancário",
	}

	return &Pagamento{
		DetPag: []DetPag{detPag},
	}
}
