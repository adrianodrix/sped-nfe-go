// Package webservices provides URL mapping and configuration for SEFAZ webservices
// across all Brazilian states, supporting both production and testing environments.
package webservices

import (
	"encoding/json"
	"fmt"

	"github.com/adrianodrix/sped-nfe-go/errors"
	"github.com/adrianodrix/sped-nfe-go/types"
)

// Service represents a specific webservice operation
type Service struct {
	Method    string `json:"method"`
	Operation string `json:"operation"`
	Version   string `json:"version"`
	URL       string `json:"url"`
}

// Environment represents either production or testing environment
type Environment struct {
	NfeStatusServico      *Service `json:"NfeStatusServico,omitempty"`
	NfeAutorizacao        *Service `json:"NfeAutorizacao,omitempty"`
	NfeConsultaProtocolo  *Service `json:"NfeConsultaProtocolo,omitempty"`
	NfeInutilizacao       *Service `json:"NfeInutilizacao,omitempty"`
	NfeRetAutorizacao     *Service `json:"NfeRetAutorizacao,omitempty"`
	RecepcaoEvento        *Service `json:"RecepcaoEvento,omitempty"`
	NfeConsultaCadastro   *Service `json:"NfeConsultaCadastro,omitempty"`
	NfeDistribuicaoDFe    *Service `json:"NfeDistribuicaoDFe,omitempty"`
	NfeConsultaDest       *Service `json:"NfeConsultaDest,omitempty"`
	NfeDownloadNF         *Service `json:"NfeDownloadNF,omitempty"`
	RecepcaoEPEC          *Service `json:"RecepcaoEPEC,omitempty"`
}

// StateWebservices represents all webservices for a specific state
type StateWebservices struct {
	Producao     *Environment `json:"producao,omitempty"`
	Homologacao  *Environment `json:"homologacao,omitempty"`
}

// WebserviceConfig holds the complete webservice configuration
type WebserviceConfig map[string]*StateWebservices

// AuthorizeMapping maps state codes to their authorizing entities for different models
var AuthorizeMapping = map[types.ModeloNFe]map[types.UF]string{
	types.ModeloNFe55: {
		types.AC: "SVRS", types.AL: "SVRS", types.AM: "AM", types.AN: "AN",
		types.AP: "SVRS", types.BA: "BA", types.CE: "SVRS", types.DF: "SVRS",
		types.ES: "SVRS", types.GO: "GO", types.MA: "SVAN", types.MG: "MG",
		types.MS: "MS", types.MT: "MT", types.PA: "SVRS", types.PB: "SVRS",
		types.PE: "PE", types.PI: "SVRS", types.PR: "PR", types.RJ: "SVRS",
		types.RN: "SVRS", types.RO: "SVRS", types.RR: "SVRS", types.RS: "RS",
		types.SC: "SVRS", types.SE: "SVRS", types.SP: "SP", types.TO: "SVRS",
		types.SVAN: "SVAN", types.SVRS: "SVRS", types.SVCAN: "SVCAN", types.SVCRS: "SVCRS",
	},
	types.ModeloNFCe65: {
		types.AC: "SVRS", types.AL: "SVRS", types.AM: "AM", types.AP: "SVRS",
		types.BA: "SVRS", types.CE: "SVRS", types.DF: "SVRS", types.ES: "SVRS",
		types.GO: "GO", types.MA: "SVRS", types.MG: "MG", types.MS: "MS",
		types.MT: "MT", types.PA: "SVRS", types.PB: "SVRS", types.PE: "SVRS",
		types.PI: "SVRS", types.PR: "PR", types.RJ: "SVRS", types.RN: "SVRS",
		types.RO: "SVRS", types.RR: "SVRS", types.RS: "RS", types.SC: "SVRS",
		types.SE: "SVRS", types.SP: "SP", types.TO: "SVRS", types.SVRS: "SVRS",
	},
}

// ServiceType represents the different types of webservice operations
type ServiceType string

const (
	ServiceStatusServico      ServiceType = "NfeStatusServico"
	ServiceAutorizacao        ServiceType = "NfeAutorizacao"
	ServiceConsultaProtocolo  ServiceType = "NfeConsultaProtocolo"
	ServiceInutilizacao       ServiceType = "NfeInutilizacao"
	ServiceRetAutorizacao     ServiceType = "NfeRetAutorizacao"
	ServiceRecepcaoEvento     ServiceType = "RecepcaoEvento"
	ServiceConsultaCadastro   ServiceType = "NfeConsultaCadastro"
	ServiceDistribuicaoDFe    ServiceType = "NfeDistribuicaoDFe"
	ServiceConsultaDest       ServiceType = "NfeConsultaDest"
	ServiceDownloadNF         ServiceType = "NfeDownloadNF"
	ServiceRecepcaoEPEC       ServiceType = "RecepcaoEPEC"
)

// NFe 4.0 Model 55 Webservices Configuration
var NFe55Config = WebserviceConfig{
	"AM": &StateWebservices{
		Homologacao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://homnfe.sefaz.am.gov.br/services2/services/NfeStatusServico4",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://homnfe.sefaz.am.gov.br/services2/services/NfeAutorizacao4",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://homnfe.sefaz.am.gov.br/services2/services/NfeConsulta4",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://homnfe.sefaz.am.gov.br/services2/services/NfeInutilizacao4",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://homnfe.sefaz.am.gov.br/services2/services/NfeRetAutorizacao4",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://homnfe.sefaz.am.gov.br/services2/services/RecepcaoEvento4",
			},
		},
		Producao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe.sefaz.am.gov.br/services2/services/NfeStatusServico4",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefaz.am.gov.br/services2/services/NfeAutorizacao4",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe.sefaz.am.gov.br/services2/services/NfeConsulta4",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe.sefaz.am.gov.br/services2/services/NfeInutilizacao4",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefaz.am.gov.br/services2/services/NfeRetAutorizacao4",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://nfe.sefaz.am.gov.br/services2/services/RecepcaoEvento4",
			},
		},
	},
	"AN": &StateWebservices{
		Homologacao: &Environment{
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://hom1.types.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx",
			},
			NfeDistribuicaoDFe: &Service{
				Method: "nfeDistDFeInteresse", Operation: "NFeDistribuicaoDFe", Version: "1.01",
				URL: "https://hom1.types.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx",
			},
			NfeConsultaDest: &Service{
				Method: "nfeConsultaNFDest", Operation: "NfeConsultaDest", Version: "1.01",
				URL: "https://hom.types.fazenda.gov.br/NFeConsultaDest/NFeConsultaDest.asmx",
			},
			NfeDownloadNF: &Service{
				Method: "nfeDownloadNF", Operation: "NfeDownloadNF", Version: "4.00",
				URL: "https://hom.types.fazenda.gov.br/NfeDownloadNF/NfeDownloadNF.asmx",
			},
			RecepcaoEPEC: &Service{
				Method: "nfeRecepcaoEvento", Operation: "RecepcaoEvento", Version: "4.00",
				URL: "https://hom.types.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx",
			},
		},
		Producao: &Environment{
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://www.types.fazenda.gov.br/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx",
			},
			NfeDistribuicaoDFe: &Service{
				Method: "nfeDistDFeInteresse", Operation: "NFeDistribuicaoDFe", Version: "1.01",
				URL: "https://www1.types.fazenda.gov.br/NFeDistribuicaoDFe/NFeDistribuicaoDFe.asmx",
			},
			NfeConsultaDest: &Service{
				Method: "nfeConsultaNFDest", Operation: "NfeConsultaDest", Version: "1.01",
				URL: "https://www.types.fazenda.gov.br/NFeConsultaDest/NFeConsultaDest.asmx",
			},
			NfeDownloadNF: &Service{
				Method: "nfeDownloadNF", Operation: "NfeDownloadNF", Version: "4.00",
				URL: "https://www.types.fazenda.gov.br/NfeDownloadNF/NfeDownloadNF.asmx",
			},
			RecepcaoEPEC: &Service{
				Method: "nfeRecepcaoEvento", Operation: "RecepcaoEvento", Version: "4.00",
				URL: "https://www.types.fazenda.gov.br/RecepcaoEvento/RecepcaoEvento.asmx",
			},
		},
	},
	"BA": &StateWebservices{
		Homologacao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://htypes.sefaz.ba.gov.br/webservices/NFeStatusServico4/NFeStatusServico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://htypes.sefaz.ba.gov.br/webservices/NFeAutorizacao4/NFeAutorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://htypes.sefaz.ba.gov.br/webservices/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://htypes.sefaz.ba.gov.br/webservices/NFeInutilizacao4/NFeInutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://htypes.sefaz.ba.gov.br/webservices/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://htypes.sefaz.ba.gov.br/webservices/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx",
			},
			NfeConsultaCadastro: &Service{
				Method: "consultaCadastro", Operation: "CadConsultaCadastro4", Version: "2.00",
				URL: "https://htypes.sefaz.ba.gov.br/webservices/CadConsultaCadastro4/CadConsultaCadastro4.asmx",
			},
		},
		Producao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe.sefaz.ba.gov.br/webservices/NFeStatusServico4/NFeStatusServico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefaz.ba.gov.br/webservices/NFeAutorizacao4/NFeAutorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe.sefaz.ba.gov.br/webservices/NFeConsultaProtocolo4/NFeConsultaProtocolo4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe.sefaz.ba.gov.br/webservices/NFeInutilizacao4/NFeInutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefaz.ba.gov.br/webservices/NFeRetAutorizacao4/NFeRetAutorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://nfe.sefaz.ba.gov.br/webservices/NFeRecepcaoEvento4/NFeRecepcaoEvento4.asmx",
			},
			NfeConsultaCadastro: &Service{
				Method: "consultaCadastro", Operation: "CadConsultaCadastro4", Version: "2.00",
				URL: "https://nfe.sefaz.ba.gov.br/webservices/CadConsultaCadastro4/CadConsultaCadastro4.asmx",
			},
		},
	},
	// PR (Paraná) - SEFAZ PR
	"PR": &StateWebservices{
		Homologacao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeConsultaProtocolo4?wsdl",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeInutilizacao4?wsdl",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeRetAutorizacao4?wsdl",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://homologacao.nfe.sefa.pr.gov.br/nfe/NFeRecepcaoEvento4?wsdl",
			},
		},
		Producao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe.sefa.pr.gov.br/nfe/NFeStatusServico4?wsdl",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefa.pr.gov.br/nfe/NFeAutorizacao4?wsdl",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe.sefa.pr.gov.br/nfe/NFeConsultaProtocolo4?wsdl",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe.sefa.pr.gov.br/nfe/NFeInutilizacao4?wsdl",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefa.pr.gov.br/nfe/NFeRetAutorizacao4?wsdl",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://nfe.sefa.pr.gov.br/nfe/NFeRecepcaoEvento4?wsdl",
			},
		},
	},
	// SVRS (Sefaz Virtual do Rio Grande do Sul) - Default for many states
	"SVRS": &StateWebservices{
		Homologacao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe-homologacao.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe-homologacao.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://nfe-homologacao.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx",
			},
		},
		Producao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe.svrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe.svrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe.svrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe.svrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe.svrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "1.00",
				URL: "https://nfe.svrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx",
			},
		},
	},
	// São Paulo (SP)
	"SP": &StateWebservices{
		Homologacao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeautorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeconsultaprotocolo4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://homologacao.nfe.fazenda.sp.gov.br/ws/nfeinutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://homologacao.nfe.fazenda.sp.gov.br/ws/nferetautorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "4.00",
				URL: "https://homologacao.nfe.fazenda.sp.gov.br/ws/nferecepcaoevento4.asmx",
			},
			NfeConsultaCadastro: &Service{
				Method: "consultaCadastro", Operation: "CadConsultaCadastro4", Version: "4.00",
				URL: "https://homologacao.nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx",
			},
		},
		Producao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe.fazenda.sp.gov.br/ws/nfestatusservico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe.fazenda.sp.gov.br/ws/nfeautorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe.fazenda.sp.gov.br/ws/nfeconsultaprotocolo4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe.fazenda.sp.gov.br/ws/nfeinutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe.fazenda.sp.gov.br/ws/nferetautorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "4.00",
				URL: "https://nfe.fazenda.sp.gov.br/ws/nferecepcaoevento4.asmx",
			},
			NfeConsultaCadastro: &Service{
				Method: "consultaCadastro", Operation: "CadConsultaCadastro4", Version: "4.00",
				URL: "https://nfe.fazenda.sp.gov.br/ws/cadconsultacadastro4.asmx",
			},
		},
	},
	// Rio Grande do Sul (RS)
	"RS": &StateWebservices{
		Homologacao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe-homologacao.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe-homologacao.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "4.00",
				URL: "https://nfe-homologacao.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx",
			},
			NfeConsultaCadastro: &Service{
				Method: "consultaCadastro", Operation: "CadConsultaCadastro4", Version: "4.00",
				URL: "https://cad-homologacao.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro4.asmx",
			},
		},
		Producao: &Environment{
			NfeStatusServico: &Service{
				Method: "nfeStatusServicoNF", Operation: "NFeStatusServico4", Version: "4.00",
				URL: "https://nfe.sefazrs.rs.gov.br/ws/NfeStatusServico/NfeStatusServico4.asmx",
			},
			NfeAutorizacao: &Service{
				Method: "nfeAutorizacaoLote", Operation: "NFeAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefazrs.rs.gov.br/ws/NfeAutorizacao/NFeAutorizacao4.asmx",
			},
			NfeConsultaProtocolo: &Service{
				Method: "nfeConsultaNF", Operation: "NFeConsultaProtocolo4", Version: "4.00",
				URL: "https://nfe.sefazrs.rs.gov.br/ws/NfeConsulta/NfeConsulta4.asmx",
			},
			NfeInutilizacao: &Service{
				Method: "nfeInutilizacaoNF", Operation: "NFeInutilizacao4", Version: "4.00",
				URL: "https://nfe.sefazrs.rs.gov.br/ws/nfeinutilizacao/nfeinutilizacao4.asmx",
			},
			NfeRetAutorizacao: &Service{
				Method: "nfeRetAutorizacaoLote", Operation: "NFeRetAutorizacao4", Version: "4.00",
				URL: "https://nfe.sefazrs.rs.gov.br/ws/NfeRetAutorizacao/NFeRetAutorizacao4.asmx",
			},
			RecepcaoEvento: &Service{
				Method: "nfeRecepcaoEvento", Operation: "NFeRecepcaoEvento4", Version: "4.00",
				URL: "https://nfe.sefazrs.rs.gov.br/ws/recepcaoevento/recepcaoevento4.asmx",
			},
			NfeConsultaCadastro: &Service{
				Method: "consultaCadastro", Operation: "CadConsultaCadastro4", Version: "4.00",
				URL: "https://cad.svrs.rs.gov.br/ws/cadconsultacadastro/cadconsultacadastro4.asmx",
			},
		},
	},
}

// GetWebserviceURL retrieves the webservice URL for a specific state, environment and service type
func GetWebserviceURL(uf types.UF, ambiente types.Ambiente, modelo types.ModeloNFe, serviceType ServiceType) (*Service, error) {
	// Get the authorizing entity for this state and model
	authorizer, err := GetAuthorizer(uf, modelo)
	if err != nil {
		return nil, err
	}

	// Get the webservice configuration for the authorizer
	config := getWebserviceConfig(modelo)
	stateConfig, exists := config[authorizer]
	if !exists {
		return nil, errors.NewValidationError(
			fmt.Sprintf("webservice configuration not found for authorizer: %s", authorizer),
			"authorizer", authorizer,
		)
	}

	// Select the appropriate environment
	var env *Environment
	if ambiente == types.AmbienteProducao {
		env = stateConfig.Producao
	} else {
		env = stateConfig.Homologacao
	}

	if env == nil {
		return nil, errors.NewValidationError(
			fmt.Sprintf("environment configuration not found for %s in %s", ambiente.String(), authorizer),
			"ambiente", ambiente.String(),
		)
	}

	// Get the specific service
	service := getServiceFromEnvironment(env, serviceType)
	if service == nil {
		return nil, errors.NewValidationError(
			fmt.Sprintf("service %s not available for %s in %s environment", serviceType, authorizer, ambiente.String()),
			"service", string(serviceType),
		)
	}

	return service, nil
}

// GetAuthorizer returns the authorizing entity for a given state and model
func GetAuthorizer(uf types.UF, modelo types.ModeloNFe) (string, error) {
	modelMapping, exists := AuthorizeMapping[modelo]
	if !exists {
		return "", errors.NewValidationError(
			fmt.Sprintf("model %d not supported", int(modelo)),
			"modelo", fmt.Sprintf("%d", int(modelo)),
		)
	}

	authorizer, exists := modelMapping[uf]
	if !exists {
		return "", errors.NewValidationError(
			fmt.Sprintf("UF %s not supported for model %d", uf.String(), int(modelo)),
			"uf", uf.String(),
		)
	}

	return authorizer, nil
}

// getWebserviceConfig returns the appropriate configuration based on model
func getWebserviceConfig(modelo types.ModeloNFe) WebserviceConfig {
	switch modelo {
	case types.ModeloNFe55:
		return NFe55Config
	case types.ModeloNFCe65:
		// For now, NFCe uses the same endpoints as NFe
		return NFe55Config
	default:
		return NFe55Config
	}
}

// getServiceFromEnvironment extracts a specific service from an environment configuration
func getServiceFromEnvironment(env *Environment, serviceType ServiceType) *Service {
	if env == nil {
		return nil
	}
	
	switch serviceType {
	case ServiceStatusServico:
		return env.NfeStatusServico
	case ServiceAutorizacao:
		return env.NfeAutorizacao
	case ServiceConsultaProtocolo:
		return env.NfeConsultaProtocolo
	case ServiceInutilizacao:
		return env.NfeInutilizacao
	case ServiceRetAutorizacao:
		return env.NfeRetAutorizacao
	case ServiceRecepcaoEvento:
		return env.RecepcaoEvento
	case ServiceConsultaCadastro:
		return env.NfeConsultaCadastro
	case ServiceDistribuicaoDFe:
		return env.NfeDistribuicaoDFe
	case ServiceConsultaDest:
		return env.NfeConsultaDest
	case ServiceDownloadNF:
		return env.NfeDownloadNF
	case ServiceRecepcaoEPEC:
		return env.RecepcaoEPEC
	default:
		return nil
	}
}

// GetAllServices returns all available services for a state and environment
func GetAllServices(uf types.UF, ambiente types.Ambiente, modelo types.ModeloNFe) (*Environment, error) {
	authorizer, err := GetAuthorizer(uf, modelo)
	if err != nil {
		return nil, err
	}

	config := getWebserviceConfig(modelo)
	stateConfig, exists := config[authorizer]
	if !exists {
		return nil, errors.NewValidationError(
			fmt.Sprintf("webservice configuration not found for authorizer: %s", authorizer),
			"authorizer", authorizer,
		)
	}

	if ambiente == types.AmbienteProducao {
		return stateConfig.Producao, nil
	}
	return stateConfig.Homologacao, nil
}

// ToJSON converts webservice configuration to JSON format
func (config WebserviceConfig) ToJSON() (string, error) {
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", errors.NewValidationError("failed to marshal webservice config to JSON", "config", "")
	}
	return string(jsonData), nil
}

// IsServiceAvailable checks if a specific service is available for a state and environment
func IsServiceAvailable(uf types.UF, ambiente types.Ambiente, modelo types.ModeloNFe, serviceType ServiceType) bool {
	service, err := GetWebserviceURL(uf, ambiente, modelo, serviceType)
	return err == nil && service != nil && service.URL != ""
}