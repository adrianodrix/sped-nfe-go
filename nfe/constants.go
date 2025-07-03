// Package nfe provides constants and codes used in Brazilian Electronic Fiscal Documents.
package nfe

// NFEEnvironment represents the NFe environment
type NFEEnvironment int

// Ambiente is an alias for NFEEnvironment for backward compatibility
type Ambiente = NFEEnvironment

const (
	EnvironmentProduction NFEEnvironment = iota + 1 // 1 - Produção
	EnvironmentTesting                              // 2 - Homologação

	// Aliases for webservices compatibility
	AmbienteProducao    = EnvironmentProduction
	AmbienteHomologacao = EnvironmentTesting
)

// String returns the string representation of NFEEnvironment
func (e NFEEnvironment) String() string {
	switch e {
	case EnvironmentProduction:
		return "Produção"
	case EnvironmentTesting:
		return "Homologação"
	default:
		return "Homologação"
	}
}

// DocumentModel represents NFe document model
type DocumentModel int

// ModeloNFe is an alias for DocumentModel for backward compatibility
type ModeloNFe = DocumentModel

const (
	ModelNFe  DocumentModel = 55 // Nota Fiscal Eletrônica
	ModelNFCe DocumentModel = 65 // Nota Fiscal de Consumidor Eletrônica

	// Aliases for webservices compatibility
	ModeloNFe55  = ModelNFe
	ModeloNFCe65 = ModelNFCe
)

// String returns the string representation of DocumentModel
func (dm DocumentModel) String() string {
	switch dm {
	case ModelNFe:
		return "55"
	case ModelNFCe:
		return "65"
	default:
		return "55"
	}
}

// DocumentType represents NFe document type
type DocumentType int

const (
	DocumentEntry DocumentType = iota // 0 - Entrada
	DocumentExit                      // 1 - Saída
)

// String returns the string representation of DocumentType
func (dt DocumentType) String() string {
	switch dt {
	case DocumentEntry:
		return "0"
	case DocumentExit:
		return "1"
	default:
		return "1"
	}
}

// OperationDestination represents operation destination
type OperationDestination int

const (
	DestinationInternal      OperationDestination = iota + 1 // 1 - Operação interna
	DestinationInterstate                                    // 2 - Operação interestadual
	DestinationInternational                                 // 3 - Operação exterior
)

// String returns the string representation of OperationDestination
func (od OperationDestination) String() string {
	switch od {
	case DestinationInternal:
		return "1"
	case DestinationInterstate:
		return "2"
	case DestinationInternational:
		return "3"
	default:
		return "1"
	}
}

// DANFEPrintType represents DANFE print type
type DANFEPrintType int

const (
	PrintTypeNone                DANFEPrintType = iota // 0 - Sem geração de DANFE
	PrintTypePortrait                                  // 1 - DANFE normal, Retrato
	PrintTypeLandscape                                 // 2 - DANFE normal, Paisagem
	PrintTypeSimplifiedPortrait                        // 3 - DANFE Simplificado, Retrato
	PrintTypeSimplifiedLandscape                       // 4 - DANFE Simplificado, Paisagem
	PrintTypeNFCe                                      // 5 - DANFE NFCe
)

// String returns the string representation of DANFEPrintType
func (dpt DANFEPrintType) String() string {
	switch dpt {
	case PrintTypeNone:
		return "0"
	case PrintTypePortrait:
		return "1"
	case PrintTypeLandscape:
		return "2"
	case PrintTypeSimplifiedPortrait:
		return "3"
	case PrintTypeSimplifiedLandscape:
		return "4"
	case PrintTypeNFCe:
		return "5"
	default:
		return "1"
	}
}

// EmissionType represents emission type
type EmissionType int

const (
	EmissionNormal           EmissionType = iota + 1 // 1 - Emissão normal
	EmissionContingencyFS                            // 2 - Contingência FS-IA, com impressão do DANFE em formulário de segurança
	EmissionContingencySCAN                          // 3 - Contingência SCAN
	EmissionContingencyDPEC                          // 4 - Contingência DPEC
	EmissionContingencyFSDA                          // 5 - Contingência FS-DA, com impressão do DANFE em formulário de segurança
	EmissionContingencySVCAN                         // 6 - Contingência SVC-AN
	EmissionContingencySVCRS                         // 7 - Contingência SVC-RS
	EmissionOffline          EmissionType = 9        // 9 - Contingência off-line da NFCe
)

// String returns the string representation of EmissionType
func (et EmissionType) String() string {
	switch et {
	case EmissionNormal:
		return "1"
	case EmissionContingencyFS:
		return "2"
	case EmissionContingencySCAN:
		return "3"
	case EmissionContingencyDPEC:
		return "4"
	case EmissionContingencyFSDA:
		return "5"
	case EmissionContingencySVCAN:
		return "6"
	case EmissionContingencySVCRS:
		return "7"
	case EmissionOffline:
		return "9"
	default:
		return "1"
	}
}

// DocumentPurpose represents NFe purpose
type DocumentPurpose int

const (
	PurposeNormal        DocumentPurpose = iota + 1 // 1 - NFe normal
	PurposeComplementary                            // 2 - NFe complementar
	PurposeAdjustment                               // 3 - NFe de ajuste
	PurposeReturn                                   // 4 - Devolução de mercadoria
)

// String returns the string representation of DocumentPurpose
func (dp DocumentPurpose) String() string {
	switch dp {
	case PurposeNormal:
		return "1"
	case PurposeComplementary:
		return "2"
	case PurposeAdjustment:
		return "3"
	case PurposeReturn:
		return "4"
	default:
		return "1"
	}
}

// FinalConsumerIndicator represents final consumer indicator
type FinalConsumerIndicator int

const (
	ConsumerNo  FinalConsumerIndicator = iota // 0 - Normal
	ConsumerYes                               // 1 - Consumidor final
)

// String returns the string representation of FinalConsumerIndicator
func (fci FinalConsumerIndicator) String() string {
	switch fci {
	case ConsumerNo:
		return "0"
	case ConsumerYes:
		return "1"
	default:
		return "0"
	}
}

// BuyerPresenceIndicator represents buyer presence indicator
type BuyerPresenceIndicator int

const (
	PresenceNotApplicable   BuyerPresenceIndicator = iota // 0 - Não se aplica
	PresencePhysical                                      // 1 - Operação presencial
	PresenceInternet                                      // 2 - Operação não presencial, pela Internet
	PresenceTelemarketing                                 // 3 - Operação não presencial, Telemarketing
	PresenceNFCeDelivery                                  // 4 - NFCe em operação com entrega a domicílio
	PresencePhysicalOutside                               // 5 - Operação presencial, fora do estabelecimento
	PresenceOther           BuyerPresenceIndicator = 9    // 9 - Operação não presencial, outros
)

// String returns the string representation of BuyerPresenceIndicator
func (bpi BuyerPresenceIndicator) String() string {
	switch bpi {
	case PresenceNotApplicable:
		return "0"
	case PresencePhysical:
		return "1"
	case PresenceInternet:
		return "2"
	case PresenceTelemarketing:
		return "3"
	case PresenceNFCeDelivery:
		return "4"
	case PresencePhysicalOutside:
		return "5"
	case PresenceOther:
		return "9"
	default:
		return "0"
	}
}

// ProcessEmission represents emission process
type ProcessEmission int

const (
	ProcessApplicationCorporate ProcessEmission = iota // 0 - Emissão de NFe com aplicativo do contribuinte
	ProcessPortalTax                                   // 1 - Emissão de NFe avulsa pelo Fisco
	ProcessPortalContributor                           // 2 - Emissão de NFe avulsa, pelo contribuinte com seu certificado digital, através do site do Fisco
	ProcessApplicationTax                              // 3 - Emissão NFe pelo contribuinte com aplicativo fornecido pelo Fisco
)

// String returns the string representation of ProcessEmission
func (pe ProcessEmission) String() string {
	switch pe {
	case ProcessApplicationCorporate:
		return "0"
	case ProcessPortalTax:
		return "1"
	case ProcessPortalContributor:
		return "2"
	case ProcessApplicationTax:
		return "3"
	default:
		return "0"
	}
}

// TaxRegime represents tax regime
type TaxRegime int

const (
	RegimeSimples  TaxRegime = iota + 1 // 1 - Simples Nacional
	RegimePSNormal                      // 2 - Simples Nacional, excesso sublimite de receita bruta
	RegimeNormal                        // 3 - Regime Normal
)

// String returns the string representation of TaxRegime
func (tr TaxRegime) String() string {
	switch tr {
	case RegimeSimples:
		return "1"
	case RegimePSNormal:
		return "2"
	case RegimeNormal:
		return "3"
	default:
		return "3"
	}
}

// StateRegistrationIndicator represents state registration indicator
type StateRegistrationIndicator int

const (
	IEContributor    StateRegistrationIndicator = iota + 1 // 1 - Contribuinte ICMS
	IEExempt                                               // 2 - Contribuinte isento de Inscrição no cadastro de Contribuintes do ICMS
	IENonContributor StateRegistrationIndicator = 9        // 9 - Não Contribuinte
)

// String returns the string representation of StateRegistrationIndicator
func (sri StateRegistrationIndicator) String() string {
	switch sri {
	case IEContributor:
		return "1"
	case IEExempt:
		return "2"
	case IENonContributor:
		return "9"
	default:
		return "9"
	}
}

// ICMS modality codes
type ICMSModality int

const (
	ICMSModalityMargin         ICMSModality = iota // 0 - Margem Valor Agregado (%)
	ICMSModalityPauta                              // 1 - Pauta (Valor)
	ICMSModalityPrice                              // 2 - Preço Tabelado Máx. (valor)
	ICMSModalityOperationValue                     // 3 - Valor da operação
)

// String returns the string representation of ICMSModality
func (im ICMSModality) String() string {
	switch im {
	case ICMSModalityMargin:
		return "0"
	case ICMSModalityPauta:
		return "1"
	case ICMSModalityPrice:
		return "2"
	case ICMSModalityOperationValue:
		return "3"
	default:
		return "0"
	}
}

// ST modality codes
type STModality int

const (
	STModalityPriceList      STModality = iota // 0 - Preço tabelado ou máximo sugerido
	STModalityNegativeList                     // 1 - Lista Negativa (valor)
	STModalityPositiveList                     // 2 - Lista Positiva (valor)
	STModalityNeutralList                      // 3 - Lista Neutra (valor)
	STModalityMargin                           // 4 - Margem Valor Agregado (%)
	STModalityPauta                            // 5 - Pauta (valor)
	STModalityOperationValue                   // 6 - Valor da Operação
)

// String returns the string representation of STModality
func (sm STModality) String() string {
	switch sm {
	case STModalityPriceList:
		return "0"
	case STModalityNegativeList:
		return "1"
	case STModalityPositiveList:
		return "2"
	case STModalityNeutralList:
		return "3"
	case STModalityMargin:
		return "4"
	case STModalityPauta:
		return "5"
	case STModalityOperationValue:
		return "6"
	default:
		return "0"
	}
}

// Brazilian states IBGE codes
var StateCodes = map[string]string{
	"AC": "12", "AL": "17", "AP": "16", "AM": "13", "BA": "29",
	"CE": "23", "DF": "53", "ES": "32", "GO": "52", "MA": "21",
	"MT": "51", "MS": "50", "MG": "31", "PA": "15", "PB": "25",
	"PR": "41", "PE": "26", "PI": "22", "RJ": "33", "RN": "24",
	"RS": "43", "RO": "11", "RR": "14", "SC": "42", "SP": "35",
	"SE": "28", "TO": "27",
}

// GetStateCode returns the IBGE code for a given state
func GetStateCode(state string) string {
	if code, ok := StateCodes[state]; ok {
		return code
	}
	return ""
}

// GetStateByCode returns the state for a given IBGE code
func GetStateByCode(code string) string {
	for state, stateCode := range StateCodes {
		if stateCode == code {
			return state
		}
	}
	return ""
}

// CFOP codes for common operations
var CommonCFOPs = map[string]string{
	// Vendas dentro do estado
	"5101": "Venda de produção do estabelecimento",
	"5102": "Venda de mercadoria adquirida ou recebida de terceiros",
	"5103": "Venda de produção do estabelecimento, efetuada fora do estabelecimento",
	"5104": "Venda de mercadoria adquirida ou recebida de terceiros, efetuada fora do estabelecimento",
	"5109": "Venda de produção do estabelecimento, não especificada nos códigos anteriores",
	"5110": "Venda de mercadoria adquirida ou recebida de terceiros, não especificada nos códigos anteriores",

	// Vendas para outros estados
	"6101": "Venda de produção do estabelecimento",
	"6102": "Venda de mercadoria adquirida ou recebida de terceiros",
	"6103": "Venda de produção do estabelecimento, efetuada fora do estabelecimento",
	"6104": "Venda de mercadoria adquirida ou recebida de terceiros, efetuada fora do estabelecimento",
	"6109": "Venda de produção do estabelecimento, não especificada nos códigos anteriores",
	"6110": "Venda de mercadoria adquirida ou recebida de terceiros, não especificada nos códigos anteriores",

	// Compras dentro do estado
	"1102": "Compra para comercialização",
	"1111": "Compra para industrialização",
	"1116": "Compra para industrialização ou produção rural",
	"1117": "Compra para comercialização",
	"1118": "Compra pelo contribuinte do Simples Nacional",

	// Compras de outros estados
	"2102": "Compra para comercialização",
	"2111": "Compra para industrialização",
	"2116": "Compra para industrialização ou produção rural",
	"2117": "Compra para comercialização",
	"2118": "Compra pelo contribuinte do Simples Nacional",

	// Devoluções
	"1201": "Devolução de venda de produção do estabelecimento",
	"1202": "Devolução de venda de mercadoria adquirida ou recebida de terceiros",
	"2201": "Devolução de venda de produção do estabelecimento",
	"2202": "Devolução de venda de mercadoria adquirida ou recebida de terceiros",
	"5201": "Devolução de compra para industrialização",
	"5202": "Devolução de compra para comercialização",
	"6201": "Devolução de compra para industrialização",
	"6202": "Devolução de compra para comercialização",
}

// NCM validation patterns
var NCMPatterns = map[string]string{
	"LENGTH": "8",          // NCM must have exactly 8 digits
	"FORMAT": "^[0-9]{8}$", // NCM format validation
}

// Common NCM codes for validation
var CommonNCMs = map[string]string{
	"00000000": "Produto não especificado",
	"99999999": "Produto não especificado",
}

// CST ICMS codes
var ICMSCSTs = map[string]string{
	"00": "Tributada integralmente",
	"10": "Tributada e com cobrança do ICMS por substituição tributária",
	"20": "Com redução de base de cálculo",
	"30": "Isenta ou não tributada e com cobrança do ICMS por substituição tributária",
	"40": "Isenta",
	"41": "Não tributada",
	"50": "Suspensão",
	"51": "Diferimento",
	"60": "ICMS cobrado anteriormente por substituição tributária",
	"70": "Com redução de base de cálculo e cobrança do ICMS por substituição tributária",
	"90": "Outras",
}

// CSOSN Simples Nacional codes
var SimplicNacionalCSOSNs = map[string]string{
	"101": "Tributada pelo Simples Nacional com permissão de crédito",
	"102": "Tributada pelo Simples Nacional sem permissão de crédito",
	"103": "Isenção do ICMS no Simples Nacional para faixa de receita bruta",
	"201": "Tributada pelo Simples Nacional com permissão de crédito e com cobrança do ICMS por substituição tributária",
	"202": "Tributada pelo Simples Nacional sem permissão de crédito e com cobrança do ICMS por substituição tributária",
	"203": "Isenção do ICMS no Simples Nacional para faixa de receita bruta e com cobrança do ICMS por substituição tributária",
	"300": "Imune",
	"400": "Não tributada pelo Simples Nacional",
	"500": "ICMS cobrado anteriormente por substituição tributária (substituído) ou por antecipação",
	"900": "Outros",
}

// PIS CST codes
var PISCSTs = map[string]string{
	"01": "Operação Tributável - Base de Cálculo = Valor da Operação Alíquota Normal (Cumulativo/Não Cumulativo)",
	"02": "Operação Tributável - Base de Cálculo = Valor da Operação (Alíquota Diferenciada)",
	"03": "Operação Tributável - Base de Cálculo = Quantidade Vendida x Alíquota por Unidade de Produto",
	"04": "Operação Tributável - Tributação Monofásica - (Alíquota Zero)",
	"05": "Operação Tributável - Substituição Tributária",
	"06": "Operação Tributável - Alíquota Zero",
	"07": "Operação Isenta da Contribuição",
	"08": "Operação Sem Incidência da Contribuição",
	"09": "Operação com Suspensão da Contribuição",
	"99": "Outras Operações",
}

// COFINS CST codes (same as PIS)
var COFINSCSTs = map[string]string{
	"01": "Operação Tributável - Base de Cálculo = Valor da Operação Alíquota Normal (Cumulativo/Não Cumulativo)",
	"02": "Operação Tributável - Base de Cálculo = Valor da Operação (Alíquota Diferenciada)",
	"03": "Operação Tributável - Base de Cálculo = Quantidade Vendida x Alíquota por Unidade de Produto",
	"04": "Operação Tributável - Tributação Monofásica - (Alíquota Zero)",
	"05": "Operação Tributável - Substituição Tributária",
	"06": "Operação Tributável - Alíquota Zero",
	"07": "Operação Isenta da Contribuição",
	"08": "Operação Sem Incidência da Contribuição",
	"09": "Operação com Suspensão da Contribuição",
	"99": "Outras Operações",
}

// IPI CST codes
var IPICSTs = map[string]string{
	"00": "Entrada com recuperação de crédito",
	"01": "Entrada tributada com alíquota zero",
	"02": "Entrada isenta",
	"03": "Entrada não-tributada",
	"04": "Entrada imune",
	"05": "Entrada com suspensão",
	"49": "Outras entradas",
	"50": "Saída tributada",
	"51": "Saída tributada com alíquota zero",
	"52": "Saída isenta",
	"53": "Saída não-tributada",
	"54": "Saída imune",
	"55": "Saída com suspensão",
	"99": "Outras saídas",
}

// ISSQN Service list codes (simplified)
var ISSQNServices = map[string]string{
	"01.01": "Análise e desenvolvimento de sistemas",
	"01.02": "Programação",
	"01.03": "Processamento, armazenamento ou hospedagem de dados",
	"01.04": "Elaboração de programas de computadores",
	"01.05": "Licenciamento ou cessão de direito de uso de programas de computação",
	"01.06": "Assessoria e consultoria em informática",
	"01.07": "Suporte técnico em informática",
	"01.08": "Planejamento, confecção, manutenção e atualização de páginas eletrônicas",
	"14.01": "Lubrificação, limpeza, lustração, revisão, carga e recarga",
	"17.01": "Assessoria ou consultoria de qualquer natureza",
}

// XML namespace constants
const (
	NFENamespace = "http://www.portalfiscal.inf.br/nfe"
	DSNamespace  = "http://www.w3.org/2000/09/xmldsig#"
)

// XML version and layout
const (
	XMLVersion    = "1.0"
	XMLEncoding   = "UTF-8"
	LayoutVersion = "4.00"
)

// ValidateCST validates CST codes for different taxes
func ValidateCST(tax string, cst string) bool {
	switch tax {
	case "ICMS":
		_, ok := ICMSCSTs[cst]
		return ok
	case "PIS":
		_, ok := PISCSTs[cst]
		return ok
	case "COFINS":
		_, ok := COFINSCSTs[cst]
		return ok
	case "IPI":
		_, ok := IPICSTs[cst]
		return ok
	default:
		return false
	}
}

// ValidateCSOSN validates CSOSN codes for Simples Nacional
func ValidateCSOSN(csosn string) bool {
	_, ok := SimplicNacionalCSOSNs[csosn]
	return ok
}

// ValidateCFOP validates CFOP codes
func ValidateCFOP(cfop string) bool {
	if len(cfop) != 4 {
		return false
	}

	// Basic validation: CFOP must be 4 digits
	for _, r := range cfop {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

// ValidateNCM validates NCM codes
func ValidateNCM(ncm string) bool {
	if len(ncm) != 8 {
		return false
	}

	// Basic validation: NCM must be 8 digits
	for _, r := range ncm {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
