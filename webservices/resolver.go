// Package webservices provides webservice URL resolution that implements common interfaces.
package webservices

import (
	"fmt"
	"strings"

	"github.com/adrianodrix/sped-nfe-go/common"
	"github.com/adrianodrix/sped-nfe-go/types"
)

// Resolver implements the WebserviceResolver interface
type Resolver struct{}

// NewResolver creates a new webservice resolver
func NewResolver() *Resolver {
	return &Resolver{}
}

// GetStatusServiceURL implements the WebserviceResolver interface
func (r *Resolver) GetStatusServiceURL(uf string, isProduction bool, model string) (common.WebServiceInfo, error) {
	return r.getServiceURL(uf, isProduction, model, ServiceStatusServico)
}

// GetAuthorizationServiceURL gets the authorization service URL
func (r *Resolver) GetAuthorizationServiceURL(uf string, isProduction bool, model string) (common.WebServiceInfo, error) {
	return r.getServiceURL(uf, isProduction, model, ServiceAutorizacao)
}

// GetInutilizacaoServiceURL gets the inutilization service URL
func (r *Resolver) GetInutilizacaoServiceURL(uf string, isProduction bool, model string) (common.WebServiceInfo, error) {
	return r.getServiceURL(uf, isProduction, model, ServiceInutilizacao)
}

// getServiceURL is a common function to get service URLs with proper SOAPAction
func (r *Resolver) getServiceURL(uf string, isProduction bool, model string, serviceType ServiceType) (common.WebServiceInfo, error) {
	// Convert parameters to internal types
	typesUF := convertStringToUF(uf)
	ambiente := convertBoolToAmbiente(isProduction)
	modelo := convertStringToModelo(model)

	// Use the existing webservices system
	service, err := GetWebserviceURL(typesUF, ambiente, modelo, serviceType)
	if err != nil {
		return common.WebServiceInfo{}, err
	}

	// Convert back to common.WebServiceInfo
	// Create proper SOAPAction from operation and method (matching common/webservices.go format)
	action := fmt.Sprintf("http://www.portalfiscal.inf.br/nfe/wsdl/%s/%s", service.Operation, service.Method)
	
	return common.WebServiceInfo{
		URL:       service.URL,
		Method:    service.Method,
		Operation: service.Operation,
		Version:   service.Version,
		Action:    action,
	}, nil
}

// Helper conversion functions
func convertStringToUF(uf string) types.UF {
	switch strings.ToUpper(uf) {
	case "AC":
		return types.AC
	case "AL":
		return types.AL
	case "AM":
		return types.AM
	case "AP":
		return types.AP
	case "BA":
		return types.BA
	case "CE":
		return types.CE
	case "DF":
		return types.DF
	case "ES":
		return types.ES
	case "GO":
		return types.GO
	case "MA":
		return types.MA
	case "MG":
		return types.MG
	case "MS":
		return types.MS
	case "MT":
		return types.MT
	case "PA":
		return types.PA
	case "PB":
		return types.PB
	case "PE":
		return types.PE
	case "PI":
		return types.PI
	case "PR":
		return types.PR
	case "RJ":
		return types.RJ
	case "RN":
		return types.RN
	case "RO":
		return types.RO
	case "RR":
		return types.RR
	case "RS":
		return types.RS
	case "SC":
		return types.SC
	case "SE":
		return types.SE
	case "SP":
		return types.SP
	case "TO":
		return types.TO
	default:
		return types.SVRS
	}
}

func convertBoolToAmbiente(isProduction bool) types.Ambiente {
	if isProduction {
		return types.AmbienteProducao
	}
	return types.AmbienteHomologacao
}

func convertStringToModelo(model string) types.ModeloNFe {
	switch model {
	case "55":
		return types.ModeloNFe55
	case "65":
		return types.ModeloNFCe65
	default:
		return types.ModeloNFe55
	}
}
