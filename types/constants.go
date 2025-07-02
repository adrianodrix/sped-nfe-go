// Package types provides common types and constants for sped-nfe-go library.
// This package defines the core types used throughout the NFe system.
package types

// Environment represents the SEFAZ environment (production or homologation)
type Environment int

const (
	// Production environment for live NFe operations
	Production Environment = 1
	// Homologation environment for testing NFe operations
	Homologation Environment = 2
)

// String returns the string representation of Environment
func (e Environment) String() string {
	switch e {
	case Production:
		return "Production"
	case Homologation:
		return "Homologation"
	default:
		return "Unknown"
	}
}

// IsValid returns true if the environment value is valid
func (e Environment) IsValid() bool {
	return e == Production || e == Homologation
}

// UF represents Brazilian states with their official codes
type UF int

const (
	// Região Norte
	AC UF = 12 // Acre
	AP UF = 16 // Amapá
	AM UF = 13 // Amazonas
	PA UF = 15 // Pará
	RO UF = 11 // Rondônia
	RR UF = 14 // Roraima
	TO UF = 17 // Tocantins

	// Região Nordeste
	AL UF = 27 // Alagoas
	BA UF = 29 // Bahia
	CE UF = 23 // Ceará
	MA UF = 21 // Maranhão
	PB UF = 25 // Paraíba
	PE UF = 26 // Pernambuco
	PI UF = 22 // Piauí
	RN UF = 24 // Rio Grande do Norte
	SE UF = 28 // Sergipe

	// Região Centro-Oeste
	GO UF = 52 // Goiás
	MT UF = 51 // Mato Grosso
	MS UF = 50 // Mato Grosso do Sul
	DF UF = 53 // Distrito Federal

	// Região Sudeste
	ES UF = 32 // Espírito Santo
	MG UF = 31 // Minas Gerais
	RJ UF = 33 // Rio de Janeiro
	SP UF = 35 // São Paulo

	// Região Sul
	PR UF = 41 // Paraná
	RS UF = 43 // Rio Grande do Sul
	SC UF = 42 // Santa Catarina

	// Especiais
	EX UF = 99 // Exterior (para operações de exportação)
)

// String returns the string representation of UF
func (uf UF) String() string {
	switch uf {
	case AC:
		return "AC"
	case AL:
		return "AL"
	case AP:
		return "AP"
	case AM:
		return "AM"
	case BA:
		return "BA"
	case CE:
		return "CE"
	case DF:
		return "DF"
	case ES:
		return "ES"
	case GO:
		return "GO"
	case MA:
		return "MA"
	case MT:
		return "MT"
	case MS:
		return "MS"
	case MG:
		return "MG"
	case PA:
		return "PA"
	case PB:
		return "PB"
	case PR:
		return "PR"
	case PE:
		return "PE"
	case PI:
		return "PI"
	case RJ:
		return "RJ"
	case RN:
		return "RN"
	case RS:
		return "RS"
	case RO:
		return "RO"
	case RR:
		return "RR"
	case SC:
		return "SC"
	case SP:
		return "SP"
	case SE:
		return "SE"
	case TO:
		return "TO"
	case EX:
		return "EX"
	default:
		return "Unknown"
	}
}

// IsValid returns true if the UF value is valid
func (uf UF) IsValid() bool {
	validUFs := []UF{
		AC, AL, AP, AM, BA, CE, DF, ES, GO, MA, MT, MS, MG,
		PA, PB, PR, PE, PI, RJ, RN, RS, RO, RR, SC, SP, SE, TO, EX,
	}
	for _, valid := range validUFs {
		if uf == valid {
			return true
		}
	}
	return false
}

// NFe Model types
type ModeloNFe int

const (
	// ModeloNFe55 represents standard NFe (modelo 55)
	ModeloNFe55 ModeloNFe = 55
	// ModeloNFCe65 represents NFCe (modelo 65)
	ModeloNFCe65 ModeloNFe = 65
)

// String returns the string representation of ModeloNFe
func (m ModeloNFe) String() string {
	switch m {
	case ModeloNFe55:
		return "NFe"
	case ModeloNFCe65:
		return "NFCe"
	default:
		return "Unknown"
	}
}

// NFe Layout versions
type VersaoLayout string

const (
	// Versao310 represents layout version 3.10
	Versao310 VersaoLayout = "3.10"
	// Versao400 represents layout version 4.00 (current)
	Versao400 VersaoLayout = "4.00"
)

// Event types for fiscal events
type TipoEvento int

const (
	// Event types from Tools.php
	EvtConfirmacao                TipoEvento = 210200 // Confirmação da Operação
	EvtCiencia                    TipoEvento = 210210 // Ciência da Operação
	EvtDesconhecimento            TipoEvento = 210220 // Desconhecimento da Operação
	EvtNaoRealizada               TipoEvento = 210240 // Operação não Realizada
	EvtCCe                        TipoEvento = 110110 // Carta de Correção Eletrônica
	EvtCancela                    TipoEvento = 110111 // Cancelamento
	EvtCancelaSubstituicao        TipoEvento = 110112 // Cancelamento por Substituição
	EvtEPEC                       TipoEvento = 110140 // EPEC
	EvtAtorInteressado            TipoEvento = 110150 // Ator Interessado
	EvtComprovanteEntrega         TipoEvento = 110130 // Comprovante de Entrega
	EvtCancelamentoCompEntrega    TipoEvento = 110131 // Cancelamento Comprovante Entrega
	EvtProrrogacao1               TipoEvento = 111500 // Prorrogação ICMS
	EvtProrrogacao2               TipoEvento = 111501 // Prorrogação IPI
	EvtCancelaProrrogacao1        TipoEvento = 111502 // Cancela Prorrogação ICMS
	EvtCancelaProrrogacao2        TipoEvento = 111503 // Cancela Prorrogação IPI
	EvtInsucessoEntrega           TipoEvento = 110192 // Insucesso na Entrega
	EvtCancelaInsucessoEntrega    TipoEvento = 110193 // Cancela Insucesso Entrega
	EvtConciliacao                TipoEvento = 110750 // Conciliação
	EvtCancelaConciliacao         TipoEvento = 110751 // Cancela Conciliação
)

// TipoEmissao represents emission types
type TipoEmissao int

const (
	TeNormal              TipoEmissao = 1 // Emissão normal
	TeContingenciaFS      TipoEmissao = 2 // Contingência FS-IA
	TeContingenciaSCAN    TipoEmissao = 3 // Contingência SCAN (deprecated)
	TeContingenciaDPEC    TipoEmissao = 4 // Contingência DPEC (deprecated)
	TeContingenciaFSDA    TipoEmissao = 5 // Contingência FS-DA
	TeContingenciaSVCAN   TipoEmissao = 6 // Contingência SVC-AN
	TeContingenciaSVCRS   TipoEmissao = 7 // Contingência SVC-RS
	TeOffline             TipoEmissao = 9 // Emissão offline NFCe
)

// TipoAmbiente represents the environment type in XML
type TipoAmbiente int

const (
	TaProducao     TipoAmbiente = 1 // Produção
	TaHomologacao  TipoAmbiente = 2 // Homologação
)

// ProcessoEmissao represents the emission process
type ProcessoEmissao int

const (
	PeAplicativoContribuinte ProcessoEmissao = 0 // Aplicativo do contribuinte
	PeAvulsaFisco           ProcessoEmissao = 1 // Avulsa pelo Fisco
	PeAvulsaContribuinte    ProcessoEmissao = 2 // Avulsa pelo contribuinte
	PeContribuinte          ProcessoEmissao = 3 // Pelo contribuinte
)

// Default values and limits
const (
	// ChaveAcessoLength is the fixed length of NFe access key
	ChaveAcessoLength = 44
	
	// MaxTimeoutSeconds is the maximum allowed timeout
	MaxTimeoutSeconds = 300
	
	// MinTimeoutSeconds is the minimum allowed timeout
	MinTimeoutSeconds = 5
	
	// DefaultTimeoutSeconds is the default timeout
	DefaultTimeoutSeconds = 30
)