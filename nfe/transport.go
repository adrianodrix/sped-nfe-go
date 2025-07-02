// Package nfe provides transport and logistics structures for NFe documents.
package nfe

import "encoding/xml"

// Transporte represents transport information for the NFe
type Transporte struct {
	XMLName     xml.Name       `xml:"transp"`
	ModFrete    string         `xml:"modFrete" validate:"required,oneof=0 1 2 3 4 9"`  // Freight mode
	Transporta  *Transportador `xml:"transporta,omitempty"`                            // Carrier
	RetTransp   *RetTransp     `xml:"retTransp,omitempty"`                             // Transport withholding
	VeicTransp  *VeicTransp    `xml:"veicTransp,omitempty"`                            // Transport vehicle
	Reboque     []Reboque      `xml:"reboque,omitempty"`                               // Trailers
	Vagao       string         `xml:"vagao,omitempty" validate:"omitempty,max=20"`     // Railway car
	Balsa       string         `xml:"balsa,omitempty" validate:"omitempty,max=20"`     // Ferry
	Vol         []Volume       `xml:"vol,omitempty"`                                   // Volumes
}

// Transportador represents carrier information
type Transportador struct {
	XMLName    xml.Name `xml:"transporta"`
	CNPJ       string   `xml:"CNPJ,omitempty" validate:"omitempty,len=14"`          // CNPJ (juridical person)
	CPF        string   `xml:"CPF,omitempty" validate:"omitempty,len=11"`           // CPF (physical person)
	XNome      string   `xml:"xNome,omitempty" validate:"omitempty,min=1,max=60"`   // Carrier name
	IE         string   `xml:"IE,omitempty"`                                        // State registration
	XEnder     string   `xml:"xEnder,omitempty" validate:"omitempty,min=1,max=60"`  // Address
	XMun       string   `xml:"xMun,omitempty" validate:"omitempty,min=1,max=60"`    // Municipality
	UF         string   `xml:"UF,omitempty" validate:"omitempty,len=2"`             // State
}

// RetTransp represents transport withholding
type RetTransp struct {
	XMLName    xml.Name `xml:"retTransp"`
	VServ      string   `xml:"vServ" validate:"required"`                           // Service value
	VBCR       string   `xml:"vBCR" validate:"required"`                            // Withholding tax base
	PICMSR     string   `xml:"pICMSR" validate:"required"`                          // ICMS withholding rate
	VICMSR     string   `xml:"vICMSR" validate:"required"`                          // ICMS withholding value
	CFOP       string   `xml:"CFOP" validate:"required,len=4"`                      // CFOP code
	CMunFG     string   `xml:"cMunFG" validate:"required,len=7"`                    // Municipality code
}

// VeicTransp represents transport vehicle
type VeicTransp struct {
	XMLName xml.Name `xml:"veicTransp"`
	Placa   string   `xml:"placa" validate:"required,min=1,max=8"`                 // License plate
	UF      string   `xml:"UF" validate:"required,len=2"`                          // State
	RNTRC   string   `xml:"RNTRC,omitempty" validate:"omitempty,max=20"`           // RNTRC number
}

// Reboque represents trailer information
type Reboque struct {
	XMLName xml.Name `xml:"reboque"`
	Placa   string   `xml:"placa" validate:"required,min=1,max=8"`                 // License plate
	UF      string   `xml:"UF" validate:"required,len=2"`                          // State
	RNTRC   string   `xml:"RNTRC,omitempty" validate:"omitempty,max=20"`           // RNTRC number
}

// Volume represents package/volume information
type Volume struct {
	XMLName xml.Name `xml:"vol"`
	QVol    string   `xml:"qVol,omitempty"`                                        // Quantity of volumes
	Esp     string   `xml:"esp,omitempty" validate:"omitempty,min=1,max=60"`       // Species
	Marca   string   `xml:"marca,omitempty" validate:"omitempty,min=1,max=60"`     // Brand
	NVol    string   `xml:"nVol,omitempty" validate:"omitempty,min=1,max=60"`      // Volume numbering
	PesoL   string   `xml:"pesoL,omitempty"`                                       // Net weight (kg)
	PesoB   string   `xml:"pesoB,omitempty"`                                       // Gross weight (kg)
	Lacres  []Lacre  `xml:"lacres,omitempty"`                                      // Seals
}

// Lacre represents seal information
type Lacre struct {
	XMLName xml.Name `xml:"lacres"`
	NLacre  string   `xml:"nLacre" validate:"required,min=1,max=60"`               // Seal number
}

// FreightMode represents freight responsibility codes
type FreightMode int

const (
	FreightSenderResponsibility FreightMode = iota // 0 - Sender responsibility (CIF)
	FreightReceiverResponsibility                  // 1 - Receiver responsibility (FOB)
	FreightThirdPartyResponsibility                // 2 - Third party responsibility
	FreightOwnSenderTransport                      // 3 - Own sender transport
	FreightOwnReceiverTransport                    // 4 - Own receiver transport
	FreightNoFreight                               // 9 - No freight
)

// String returns the string representation of FreightMode
func (fm FreightMode) String() string {
	switch fm {
	case FreightSenderResponsibility:
		return "0"
	case FreightReceiverResponsibility:
		return "1"
	case FreightThirdPartyResponsibility:
		return "2"
	case FreightOwnSenderTransport:
		return "3"
	case FreightOwnReceiverTransport:
		return "4"
	case FreightNoFreight:
		return "9"
	default:
		return "0"
	}
}

// TransportBuilder helps build transport information
type TransportBuilder struct {
	transport *Transporte
}

// NewTransportBuilder creates a new transport builder
func NewTransportBuilder() *TransportBuilder {
	return &TransportBuilder{
		transport: &Transporte{
			ModFrete: FreightSenderResponsibility.String(),
			Vol:      make([]Volume, 0),
			Reboque:  make([]Reboque, 0),
		},
	}
}

// SetFreightMode sets the freight responsibility mode
func (tb *TransportBuilder) SetFreightMode(mode FreightMode) *TransportBuilder {
	tb.transport.ModFrete = mode.String()
	return tb
}

// SetCarrier sets carrier information
func (tb *TransportBuilder) SetCarrier(carrier *Transportador) *TransportBuilder {
	tb.transport.Transporta = carrier
	return tb
}

// SetTransportWithholding sets transport withholding information
func (tb *TransportBuilder) SetTransportWithholding(ret *RetTransp) *TransportBuilder {
	tb.transport.RetTransp = ret
	return tb
}

// SetTransportVehicle sets transport vehicle information
func (tb *TransportBuilder) SetTransportVehicle(vehicle *VeicTransp) *TransportBuilder {
	tb.transport.VeicTransp = vehicle
	return tb
}

// AddTrailer adds a trailer to the transport
func (tb *TransportBuilder) AddTrailer(trailer *Reboque) *TransportBuilder {
	tb.transport.Reboque = append(tb.transport.Reboque, *trailer)
	return tb
}

// SetRailwayInfo sets railway transport information
func (tb *TransportBuilder) SetRailwayInfo(vagao, balsa string) *TransportBuilder {
	tb.transport.Vagao = vagao
	tb.transport.Balsa = balsa
	return tb
}

// AddVolume adds a volume/package to the transport
func (tb *TransportBuilder) AddVolume(volume *Volume) *TransportBuilder {
	tb.transport.Vol = append(tb.transport.Vol, *volume)
	return tb
}

// Build returns the constructed transport information
func (tb *TransportBuilder) Build() *Transporte {
	return tb.transport
}

// VolumeBuilder helps build volume information
type VolumeBuilder struct {
	volume *Volume
}

// NewVolumeBuilder creates a new volume builder
func NewVolumeBuilder() *VolumeBuilder {
	return &VolumeBuilder{
		volume: &Volume{
			Lacres: make([]Lacre, 0),
		},
	}
}

// SetQuantity sets the quantity of volumes
func (vb *VolumeBuilder) SetQuantity(qty string) *VolumeBuilder {
	vb.volume.QVol = qty
	return vb
}

// SetSpecies sets the species/type of package
func (vb *VolumeBuilder) SetSpecies(species string) *VolumeBuilder {
	vb.volume.Esp = species
	return vb
}

// SetBrand sets the brand/marking
func (vb *VolumeBuilder) SetBrand(brand string) *VolumeBuilder {
	vb.volume.Marca = brand
	return vb
}

// SetNumbering sets the volume numbering
func (vb *VolumeBuilder) SetNumbering(numbering string) *VolumeBuilder {
	vb.volume.NVol = numbering
	return vb
}

// SetWeights sets net and gross weights
func (vb *VolumeBuilder) SetWeights(netWeight, grossWeight string) *VolumeBuilder {
	vb.volume.PesoL = netWeight
	vb.volume.PesoB = grossWeight
	return vb
}

// AddSeal adds a seal to the volume
func (vb *VolumeBuilder) AddSeal(sealNumber string) *VolumeBuilder {
	seal := Lacre{NLacre: sealNumber}
	vb.volume.Lacres = append(vb.volume.Lacres, seal)
	return vb
}

// Build returns the constructed volume
func (vb *VolumeBuilder) Build() *Volume {
	return vb.volume
}

// ValidateTransport validates transport information
func ValidateTransport(transport *Transporte) error {
	if transport == nil {
		return nil // Transport is optional
	}

	// TODO: Implement validation logic
	// - Validate freight mode
	// - Check carrier information consistency
	// - Validate vehicle plates format
	// - Check weight values format
	
	return nil
}

// CreateSimpleTransport creates a simple transport with freight mode only
func CreateSimpleTransport(freightMode FreightMode) *Transporte {
	return &Transporte{
		ModFrete: freightMode.String(),
	}
}

// CreateCarrierTransport creates transport with carrier information
func CreateCarrierTransport(freightMode FreightMode, carrier *Transportador) *Transporte {
	return &Transporte{
		ModFrete:   freightMode.String(),
		Transporta: carrier,
	}
}

// CreateVehicleTransport creates transport with vehicle information
func CreateVehicleTransport(freightMode FreightMode, vehicle *VeicTransp) *Transporte {
	return &Transporte{
		ModFrete:   freightMode.String(),
		VeicTransp: vehicle,
	}
}