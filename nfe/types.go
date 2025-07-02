// Package nfe provides comprehensive data structures for Brazilian Electronic Fiscal Documents (NFe/NFCe).
// This package implements all required structures according to SEFAZ technical specifications.
package nfe

import (
	"encoding/xml"
	"time"
)

// NFe represents the root element of an Electronic Fiscal Document
type NFe struct {
	XMLName xml.Name `xml:"NFe"`
	Xmlns   string   `xml:"xmlns,attr"`
	InfNFe  InfNFe   `xml:"infNFe"`
}

// InfNFe contains all information about the Electronic Fiscal Document
type InfNFe struct {
	XMLName  xml.Name      `xml:"infNFe"`
	ID       string        `xml:"Id,attr" validate:"required,len=47"`      // NFe + 44 digits access key
	Versao   string        `xml:"versao,attr" validate:"required"`         // Layout version (4.00)
	PKNItem  string        `xml:"pk_nItem,attr,omitempty"`                // Partial NFe key item
	
	Ide      Identificacao `xml:"ide"`                                    // Identification
	Emit     Emitente      `xml:"emit"`                                   // Issuer
	Avulsa   *Avulsa       `xml:"avulsa,omitempty"`                      // Isolated operation
	Dest     *Destinatario `xml:"dest,omitempty"`                        // Recipient (optional for NFCe)
	Retirada *Local        `xml:"retirada,omitempty"`                    // Pickup location
	Entrega  *Local        `xml:"entrega,omitempty"`                     // Delivery location
	AutXML   []AutorizXML  `xml:"autXML,omitempty"`                      // XML authorization
	Det      []Item        `xml:"det"`                                    // Items/Products
	Total    Total         `xml:"total"`                                  // Totals
	Transp   Transporte    `xml:"transp"`                                // Transport
	Cobr     *Cobranca     `xml:"cobr,omitempty"`                        // Billing (optional)
	Pag      []Pagamento   `xml:"pag,omitempty"`                         // Payments
	InfAdic  *InfAdicionais `xml:"infAdic,omitempty"`                    // Additional information
	Exporta  *Exportacao   `xml:"exporta,omitempty"`                     // Export information
	Compra   *Compra       `xml:"compra,omitempty"`                      // Purchase information
	Cana     *Cana         `xml:"cana,omitempty"`                        // Sugar cane information
	InfRespTec *InfRespTec `xml:"infRespTec,omitempty"`                  // Technical responsible information
}

// Identificacao contains identification data for the NFe
type Identificacao struct {
	XMLName   xml.Name `xml:"ide"`
	CUF       string   `xml:"cUF" validate:"required,len=2"`              // IBGE UF code
	CNF       string   `xml:"cNF" validate:"required,len=8"`              // Random numeric code
	NatOp     string   `xml:"natOp" validate:"required,min=1,max=60"`     // Operation nature
	Mod       string   `xml:"mod" validate:"required,oneof=55 65"`        // Document model (55=NFe, 65=NFCe)
	Serie     string   `xml:"serie" validate:"required,min=1,max=3"`      // Series
	NNF       string   `xml:"nNF" validate:"required,min=1,max=9"`       // NFe number
	DhEmi     string   `xml:"dhEmi" validate:"required"`                  // Issue date/time
	DhSaiEnt  string   `xml:"dhSaiEnt,omitempty"`                        // Exit/entry date/time
	TpNF      string   `xml:"tpNF" validate:"required,oneof=0 1"`        // NFe type (0=entry, 1=exit)
	IdDest    string   `xml:"idDest" validate:"required,oneof=1 2 3"`    // Operation destination
	CMunFG    string   `xml:"cMunFG" validate:"required,len=7"`          // Municipality code
	TpImp     string   `xml:"tpImp" validate:"required,oneof=0 1 2 3 4 5"` // DANFE print type
	TpEmis    string   `xml:"tpEmis" validate:"required,oneof=1 2 3 4 5 6 7 9"` // Emission type
	CDV       string   `xml:"cDV" validate:"required,len=1"`             // Check digit
	TpAmb     string   `xml:"tpAmb" validate:"required,oneof=1 2"`       // Environment (1=prod, 2=test)
	FinNFe    string   `xml:"finNFe" validate:"required,oneof=1 2 3 4"`  // NFe purpose
	IndFinal  string   `xml:"indFinal" validate:"required,oneof=0 1"`    // Final consumer indicator
	IndPres   string   `xml:"indPres" validate:"required,oneof=0 1 2 3 4 5 9"` // Buyer presence indicator
	IndIntermed string `xml:"indIntermed,omitempty" validate:"omitempty,oneof=0 1"` // Intermediary indicator
	ProcEmi   string   `xml:"procEmi" validate:"required,oneof=0 1 2 3"`  // Emission process
	VerProc   string   `xml:"verProc" validate:"required,min=1,max=20"`   // Process version
	DhCont    string   `xml:"dhCont,omitempty"`                          // Contingency date
	XJust     string   `xml:"xJust,omitempty" validate:"omitempty,min=15,max=256"` // Justification
	
	NFref     []NFRef  `xml:"NFref,omitempty"`                           // Referenced documents
}

// NFRef represents references to other fiscal documents
type NFRef struct {
	XMLName    xml.Name  `xml:"NFref"`
	RefNFe     string    `xml:"refNFe,omitempty" validate:"omitempty,len=44"`      // Referenced NFe
	RefNFeSig  string    `xml:"refNFeSig,omitempty"`                              // NFe signature reference
	RefNF      *RefNF    `xml:"refNF,omitempty"`                                  // Referenced NF model 1/1A
	RefNFP     *RefNFP   `xml:"refNFP,omitempty"`                                 // Referenced NF producer
	RefCTe     string    `xml:"refCTe,omitempty" validate:"omitempty,len=44"`     // Referenced CTe
	RefECF     *RefECF   `xml:"refECF,omitempty"`                                 // Referenced ECF
}

// RefNF represents a reference to NF model 1/1A
type RefNF struct {
	XMLName xml.Name `xml:"refNF"`
	CUF     string   `xml:"cUF" validate:"required,len=2"`
	AAMM    string   `xml:"AAMM" validate:"required,len=4"`
	CNPJ    string   `xml:"CNPJ" validate:"required,len=14"`
	Mod     string   `xml:"mod" validate:"required,len=2"`
	Serie   string   `xml:"serie" validate:"required,min=1,max=3"`
	NNF     string   `xml:"nNF" validate:"required,min=1,max=9"`
}

// RefNFP represents a reference to NF producer
type RefNFP struct {
	XMLName xml.Name `xml:"refNFP"`
	CUF     string   `xml:"cUF" validate:"required,len=2"`
	AAMM    string   `xml:"AAMM" validate:"required,len=4"`
	CNPJ    string   `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`
	CPF     string   `xml:"CPF,omitempty" validate:"omitempty,len=11"`
	IE      string   `xml:"IE" validate:"required"`
	Mod     string   `xml:"mod" validate:"required,len=2"`
	Serie   string   `xml:"serie" validate:"required,min=1,max=3"`
	NNF     string   `xml:"nNF" validate:"required,min=1,max=9"`
}

// RefECF represents a reference to ECF (Electronic Fiscal Coupon)
type RefECF struct {
	XMLName xml.Name `xml:"refECF"`
	Mod     string   `xml:"mod" validate:"required,len=2"`
	NECF    string   `xml:"nECF" validate:"required,min=1,max=3"`
	NCOO    string   `xml:"nCOO" validate:"required,min=1,max=6"`
}

// Emitente represents the document issuer information
type Emitente struct {
	XMLName   xml.Name  `xml:"emit"`
	CNPJ      string    `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`       // CNPJ (juridical person)
	CPF       string    `xml:"CPF,omitempty" validate:"omitempty,len=11"`        // CPF (physical person)
	XNome     string    `xml:"xNome" validate:"required,min=2,max=60"`          // Company/person name
	XFant     string    `xml:"xFant,omitempty" validate:"omitempty,max=60"`     // Trade name
	EnderEmit Endereco  `xml:"enderEmit"`                                       // Address
	IE        string    `xml:"IE" validate:"required"`                          // State registration
	IEST      string    `xml:"IEST,omitempty"`                                  // ST state registration
	IM        string    `xml:"IM,omitempty"`                                    // Municipal registration
	CNAE      string    `xml:"CNAE,omitempty" validate:"omitempty,len=7"`       // CNAE code
	CRT       string    `xml:"CRT" validate:"required,oneof=1 2 3"`             // Tax regime code
}

// Destinatario represents the document recipient information
type Destinatario struct {
	XMLName      xml.Name     `xml:"dest"`
	CNPJ         string       `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`       // CNPJ (juridical person)
	CPF          string       `xml:"CPF,omitempty" validate:"omitempty,len=11"`        // CPF (physical person)
	IdEstrangeiro string      `xml:"idEstrangeiro,omitempty"`                         // Foreign ID
	XNome        string       `xml:"xNome,omitempty" validate:"omitempty,min=2,max=60"` // Company/person name
	EnderDest    *Endereco    `xml:"enderDest,omitempty"`                             // Address
	IndIEDest    string       `xml:"indIEDest" validate:"required,oneof=1 2 9"`      // State registration indicator
	IE           string       `xml:"IE,omitempty"`                                    // State registration
	ISUF         string       `xml:"ISUF,omitempty"`                                  // SUFRAMA registration
	IM           string       `xml:"IM,omitempty"`                                    // Municipal registration
	Email        string       `xml:"email,omitempty" validate:"omitempty,email"`      // Email address
}

// Endereco represents address information
type Endereco struct {
	XMLName xml.Name `xml:"enderEmit"`                                       // Will be overridden by parent
	XLgr    string   `xml:"xLgr" validate:"required,min=2,max=60"`          // Street name
	Nro     string   `xml:"nro" validate:"required,min=1,max=60"`           // Number
	XCpl    string   `xml:"xCpl,omitempty" validate:"omitempty,max=60"`     // Complement
	XBairro string   `xml:"xBairro" validate:"required,min=2,max=60"`       // District
	CMun    string   `xml:"cMun" validate:"required,len=7"`                 // Municipality IBGE code
	XMun    string   `xml:"xMun" validate:"required,min=2,max=60"`          // Municipality name
	UF      string   `xml:"UF" validate:"required,len=2"`                   // State
	CEP     string   `xml:"CEP" validate:"required,len=8"`                  // ZIP code
	CPais   string   `xml:"cPais,omitempty" validate:"omitempty,len=4"`     // Country code
	XPais   string   `xml:"xPais,omitempty" validate:"omitempty,min=1,max=60"` // Country name
	Fone    string   `xml:"fone,omitempty" validate:"omitempty,max=14"`     // Phone number
}

// Local represents pickup or delivery location
type Local struct {
	XMLName xml.Name `xml:"retirada"`                                       // Will be overridden by parent
	CNPJ    string   `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`
	CPF     string   `xml:"CPF,omitempty" validate:"omitempty,len=11"`
	XNome   string   `xml:"xNome,omitempty" validate:"omitempty,min=2,max=60"`
	XLgr    string   `xml:"xLgr" validate:"required,min=2,max=60"`
	Nro     string   `xml:"nro" validate:"required,min=1,max=60"`
	XCpl    string   `xml:"xCpl,omitempty" validate:"omitempty,max=60"`
	XBairro string   `xml:"xBairro" validate:"required,min=2,max=60"`
	CMun    string   `xml:"cMun" validate:"required,len=7"`
	XMun    string   `xml:"xMun" validate:"required,min=2,max=60"`
	UF      string   `xml:"UF" validate:"required,len=2"`
}

// AutorizXML represents XML access authorization
type AutorizXML struct {
	XMLName xml.Name `xml:"autXML"`
	CNPJ    string   `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`
	CPF     string   `xml:"CPF,omitempty" validate:"omitempty,len=11"`
}

// Avulsa represents isolated operation information
type Avulsa struct {
	XMLName xml.Name `xml:"avulsa"`
	CNPJ    string   `xml:"CNPJ" validate:"required,len=14"`
	XOrgao  string   `xml:"xOrgao" validate:"required,min=1,max=60"`
	Matr    string   `xml:"matr" validate:"required,min=1,max=60"`
	XAgente string   `xml:"xAgente" validate:"required,min=1,max=60"`
	Fone    string   `xml:"fone,omitempty" validate:"omitempty,max=14"`
	UF      string   `xml:"UF" validate:"required,len=2"`
	NDAR    string   `xml:"nDAR,omitempty"`
	DEmi    string   `xml:"dEmi,omitempty"`
	VPag    string   `xml:"vPag,omitempty"`
	RepEmi  string   `xml:"repEmi" validate:"required,oneof=1 2"`
	DPag    string   `xml:"dPag,omitempty"`
}

// Item represents a product/service item in the NFe
type Item struct {
	XMLName   xml.Name  `xml:"det"`
	NItem     string    `xml:"nItem,attr" validate:"required"`               // Item sequence number
	Prod      Produto   `xml:"prod"`                                         // Product information
	Imposto   Imposto   `xml:"imposto"`                                      // Tax information
	InfAdProd string    `xml:"infAdProd,omitempty" validate:"omitempty,max=500"` // Additional product info
}

// Produto represents product/service information
type Produto struct {
	XMLName      xml.Name `xml:"prod"`
	CProd        string   `xml:"cProd" validate:"required,min=1,max=60"`      // Product code
	CEAN         string   `xml:"cEAN" validate:"required"`                    // GTIN/EAN code
	CBarra       string   `xml:"cBarra,omitempty"`                            // Barcode (deprecated)
	XProd        string   `xml:"xProd" validate:"required,min=1,max=120"`     // Product description
	NCM          string   `xml:"NCM" validate:"required,len=8"`               // NCM code
	CBenef       string   `xml:"cBenef,omitempty"`                            // Benefit code
	EXTIPI       string   `xml:"EXTIPI,omitempty" validate:"omitempty,len=3"` // IPI exception
	CFOP         string   `xml:"CFOP" validate:"required,len=4"`              // CFOP code
	UCom         string   `xml:"uCom" validate:"required,min=1,max=6"`        // Commercial unit
	QCom         string   `xml:"qCom" validate:"required"`                    // Commercial quantity
	VUnCom       string   `xml:"vUnCom" validate:"required"`                  // Commercial unit value
	VProd        string   `xml:"vProd" validate:"required"`                   // Product total value
	CEANTrib     string   `xml:"cEANTrib" validate:"required"`                // Tributary GTIN/EAN
	CBarraTrib   string   `xml:"cBarraTrib,omitempty"`                        // Tributary barcode (deprecated)
	UTrib        string   `xml:"uTrib" validate:"required,min=1,max=6"`       // Tributary unit
	QTrib        string   `xml:"qTrib" validate:"required"`                   // Tributary quantity
	VUnTrib      string   `xml:"vUnTrib" validate:"required"`                 // Tributary unit value
	VFrete       string   `xml:"vFrete,omitempty"`                            // Freight value
	VSeg         string   `xml:"vSeg,omitempty"`                              // Insurance value
	VDesc        string   `xml:"vDesc,omitempty"`                             // Discount value
	VOutro       string   `xml:"vOutro,omitempty"`                            // Other charges value
	IndTot       string   `xml:"indTot" validate:"required,oneof=0 1"`        // Totals indicator
	XPed         string   `xml:"xPed,omitempty" validate:"omitempty,max=15"`  // Purchase order
	NItemPed     string   `xml:"nItemPed,omitempty" validate:"omitempty,max=6"` // Purchase order item
	NFCI         string   `xml:"nFCI,omitempty" validate:"omitempty,len=36"`  // FCI number
	CEST         string   `xml:"CEST,omitempty" validate:"omitempty,len=7"`   // CEST code
	IndEscala    string   `xml:"indEscala,omitempty" validate:"omitempty,oneof=S N"` // Scale indicator
	CNPJFab      string   `xml:"CNPJFab,omitempty" validate:"omitempty,len=14"` // Manufacturer CNPJ
	
	// Optional sub-groups
	DI           []DI         `xml:"DI,omitempty"`           // Import declaration
	DetExport    []DetExport  `xml:"detExport,omitempty"`    // Export detail
	Rastro       []Rastro     `xml:"rastro,omitempty"`       // Traceability
	VeicProd     *VeicProd    `xml:"veicProd,omitempty"`     // Vehicle
	Med          []Med        `xml:"med,omitempty"`          // Medicine
	Arma         []Arma       `xml:"arma,omitempty"`         // Weapon
	Comb         *Combustivel `xml:"comb,omitempty"`         // Fuel
	CIDE         *CIDE        `xml:"CIDE,omitempty"`         // CIDE tax
}

// DI represents import declaration information
type DI struct {
	XMLName   xml.Name `xml:"DI"`
	NDI       string   `xml:"nDI" validate:"required,max=12"`
	DDI       string   `xml:"dDI" validate:"required"`
	XLocDesemb string  `xml:"xLocDesemb" validate:"required,min=1,max=60"`
	UFDesemb  string   `xml:"UFDesemb" validate:"required,len=2"`
	DDesemb   string   `xml:"dDesemb" validate:"required"`
	TpViaTransp string `xml:"tpViaTransp" validate:"required,oneof=1 2 3 4 5 6 7 8 9 10 11 12"`
	VAfrmm    string   `xml:"vAFRMM,omitempty"`
	TpIntermedio string `xml:"tpIntermedio" validate:"required,oneof=1 2 3"`
	CNPJ      string   `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`
	UFTerceiro string  `xml:"UFTerceiro,omitempty" validate:"omitempty,len=2"`
	CExportador string `xml:"cExportador" validate:"required"`
	Adi       []Adi    `xml:"adi"`
}

// Adi represents addition information in import declaration
type Adi struct {
	XMLName    xml.Name `xml:"adi"`
	NAdicao    string   `xml:"nAdicao" validate:"required,min=1,max=3"`
	NSeqAdic   string   `xml:"nSeqAdic" validate:"required,min=1,max=3"`
	CFabricante string  `xml:"cFabricante" validate:"required"`
	VDescDI    string   `xml:"vDescDI,omitempty"`
	NDrawback  string   `xml:"nDrawback,omitempty" validate:"omitempty,max=11"`
}

// DetExport represents export detail
type DetExport struct {
	XMLName      xml.Name `xml:"detExport"`
	NDrawback    string   `xml:"nDrawback,omitempty" validate:"omitempty,max=11"`
	NRE          string   `xml:"nRE" validate:"required,max=12"`
	ChNFe        string   `xml:"chNFe,omitempty" validate:"omitempty,len=44"`
	QExport      string   `xml:"qExport" validate:"required"`
}

// Rastro represents traceability information
type Rastro struct {
	XMLName   xml.Name `xml:"rastro"`
	NLote     string   `xml:"nLote" validate:"required,min=1,max=20"`
	QLote     string   `xml:"qLote" validate:"required"`
	DFab      string   `xml:"dFab" validate:"required"`
	DVal      string   `xml:"dVal" validate:"required"`
	CAgreg    string   `xml:"cAgreg,omitempty" validate:"omitempty,max=20"`
}

// VeicProd represents vehicle information
type VeicProd struct {
	XMLName     xml.Name `xml:"veicProd"`
	TpOp        string   `xml:"tpOp" validate:"required,oneof=1 2 3"`
	Chassi      string   `xml:"chassi,omitempty" validate:"omitempty,len=17"`
	CCor        string   `xml:"cCor" validate:"required,len=4"`
	XCor        string   `xml:"xCor" validate:"required,min=1,max=40"`
	Pot         string   `xml:"pot" validate:"required"`
	Cilin       string   `xml:"cilin" validate:"required"`
	PesoL       string   `xml:"pesoL" validate:"required"`
	PesoB       string   `xml:"pesoB" validate:"required"`
	NSerie      string   `xml:"nSerie,omitempty" validate:"omitempty,max=9"`
	TpComb      string   `xml:"tpComb" validate:"required"`
	NMotor      string   `xml:"nMotor,omitempty" validate:"omitempty,max=21"`
	CMT         string   `xml:"CMT" validate:"required"`
	Dist        string   `xml:"dist" validate:"required"`
	AnoMod      string   `xml:"anoMod" validate:"required,len=4"`
	AnoFab      string   `xml:"anoFab" validate:"required,len=4"`
	TpPint      string   `xml:"tpPint" validate:"required,oneof=M R"`
	TpVeic      string   `xml:"tpVeic" validate:"required,len=2"`
	EspVeic     string   `xml:"espVeic" validate:"required,oneof=1 2 3 4 5 6"`
	VIN         string   `xml:"VIN,omitempty" validate:"omitempty,oneof=N R S"`
	CondVeic    string   `xml:"condVeic" validate:"required,oneof=1 2 3"`
	CMod        string   `xml:"cMod" validate:"required,len=6"`
	CCorDENATRAN string `xml:"cCorDENATRAN" validate:"required,len=2"`
	LotaMax     string   `xml:"lota" validate:"required"`
	TpRest      string   `xml:"tpRest" validate:"required,oneof=0 1 2 3 4"`
}

// Med represents medicine information
type Med struct {
	XMLName   xml.Name `xml:"med"`
	CloTMed   string   `xml:"cProdANVISA,omitempty" validate:"omitempty,max=13"`
	XMotivoIsencao string `xml:"xMotivoIsencao,omitempty" validate:"omitempty,max=255"`
	NProdANVISA string   `xml:"nProdANVISA" validate:"required"`
	VPMCProdANVISA string `xml:"vPMC" validate:"required"`
}

// Arma represents weapon information
type Arma struct {
	XMLName xml.Name `xml:"arma"`
	TpArma  string   `xml:"tpArma" validate:"required,oneof=0 1"`
	NSerie  string   `xml:"nSerie" validate:"required,min=1,max=15"`
	NCano   string   `xml:"nCano" validate:"required,min=1,max=15"`
	Descr   string   `xml:"descr" validate:"required,min=1,max=256"`
}

// Combustivel represents fuel information
type Combustivel struct {
	XMLName     xml.Name `xml:"comb"`
	CProdANP    string   `xml:"cProdANP" validate:"required,len=9"`
	DescANP     string   `xml:"descANP,omitempty" validate:"omitempty,max=95"`
	PGLP        string   `xml:"pGLP,omitempty"`
	PGNn        string   `xml:"pGNn,omitempty"`
	PGNi        string   `xml:"pGNi,omitempty"`
	VPart       string   `xml:"vPart,omitempty"`
	CODIF       string   `xml:"CODIF,omitempty" validate:"omitempty,len=21"`
	QTemp       string   `xml:"qTemp,omitempty"`
	UFCons      string   `xml:"UFCons" validate:"required,len=2"`
	QBCProd     string   `xml:"qBCProd" validate:"required"`
	VAliqProd   string   `xml:"vAliqProd" validate:"required"`
	VCide       string   `xml:"vCIDE" validate:"required"`
	ENCOPro     *ENCOPro `xml:"encerrProds,omitempty"`
}

// ENCOPro represents closing production information
type ENCOPro struct {
	XMLName   xml.Name `xml:"encerrProds"`
	NbombaPos string   `xml:"nBombaPos" validate:"required,min=1,max=3"`
	NTanque   string   `xml:"nTanque" validate:"required,min=1,max=3"`
	VEncIni   string   `xml:"vEncIni" validate:"required"`
	VEncFin   string   `xml:"vEncFin" validate:"required"`
}

// CIDE represents CIDE tax information
type CIDE struct {
	XMLName     xml.Name `xml:"CIDE"`
	QBCProd     string   `xml:"qBCProd" validate:"required"`
	VAliqProd   string   `xml:"vAliqProd" validate:"required"`
	VCide       string   `xml:"vCIDE" validate:"required"`
}

// Imposto represents tax information for an item
type Imposto struct {
	XMLName   xml.Name   `xml:"imposto"`
	VTotTrib  string     `xml:"vTotTrib,omitempty"`                            // Total tax burden
	ICMS      *ICMS      `xml:"ICMS,omitempty"`                                // ICMS tax
	IPI       *IPI       `xml:"IPI,omitempty"`                                 // IPI tax
	II        *II        `xml:"II,omitempty"`                                  // Import tax
	PIS       *PIS       `xml:"PIS,omitempty"`                                 // PIS tax
	PISST     *PISST     `xml:"PISST,omitempty"`                               // PIS ST tax
	COFINS    *COFINS    `xml:"COFINS,omitempty"`                              // COFINS tax
	COFINSST  *COFINSST  `xml:"COFINSST,omitempty"`                            // COFINS ST tax
	ISSQN     *ISSQN     `xml:"ISSQN,omitempty"`                               // ISSQN tax
}

// Time represents a time with proper timezone handling
type Time struct {
	time.Time
}

// MarshalXML implements xml.Marshaler interface for proper time formatting
func (t Time) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	formatted := t.Format("2006-01-02T15:04:05-07:00")
	return e.EncodeElement(formatted, start)
}

// UnmarshalXML implements xml.Unmarshaler interface for proper time parsing
func (t *Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var timeStr string
	if err := d.DecodeElement(&timeStr, &start); err != nil {
		return err
	}
	
	parsed, err := time.Parse("2006-01-02T15:04:05-07:00", timeStr)
	if err != nil {
		// Try alternative format without timezone
		parsed, err = time.Parse("2006-01-02T15:04:05", timeStr)
		if err != nil {
			return err
		}
	}
	
	t.Time = parsed
	return nil
}