// Package common provides common interfaces and types for sped-nfe-go library.
package common

// WebserviceResolver defines the interface for resolving webservice URLs
type WebserviceResolver interface {
	// GetStatusServiceURL returns the status service URL for the given parameters
	GetStatusServiceURL(uf string, isProduction bool, model string) (WebServiceInfo, error)
}

// WebServiceLookup represents a webservice lookup result
type WebServiceLookup struct {
	URL       string
	Method    string
	Operation string
	Version   string
}
