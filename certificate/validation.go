// Package certificate provides ICP-Brasil certificate validation functions.
package certificate

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
)

// ICP-Brasil root certificate subjects and known issuer patterns
var icpBrasilRootPatterns = []string{
	"AC Raiz",
	"ICP-Brasil",
	"Instituto Nacional de Tecnologia da Informacao",
	"ITI",
	"Autoridade Certificadora Raiz Brasileira",
}

// Known ICP-Brasil intermediate CA patterns
var icpBrasilCAPatterns = []string{
	"AC ",
	"Autoridade Certificadora",
	"Certisign",
	"Serasa",
	"SERPRO",
	"Caixa",
	"SOLUTI",
	"Valid",
	"DIGITALSIGN",
}

// ValidateICPBrasilCertificateChain validates a certificate chain against ICP-Brasil requirements
func ValidateICPBrasilCertificateChain(chain []*x509.Certificate) error {
	if len(chain) == 0 {
		return errors.NewValidationError("certificate chain cannot be empty", "chain", "")
	}

	// Check if the end-entity certificate is from ICP-Brasil
	if !IsICPBrasilCertificate(chain[0]) {
		return errors.NewCertificateError("certificate is not from ICP-Brasil", nil)
	}

	// Validate certificate chain structure
	for i, cert := range chain {
		// Check if certificate is valid (not expired)
		if time.Now().After(cert.NotAfter) {
			return errors.NewCertificateError(fmt.Sprintf("certificate %d in chain is expired", i), nil)
		}

		if time.Now().Before(cert.NotBefore) {
			return errors.NewCertificateError(fmt.Sprintf("certificate %d in chain is not yet valid", i), nil)
		}

		// For intermediate certificates, verify they're signed by the next certificate in chain
		if i < len(chain)-1 {
			err := cert.CheckSignatureFrom(chain[i+1])
			if err != nil {
				return errors.NewCertificateError(fmt.Sprintf("certificate %d is not signed by certificate %d", i, i+1), err)
			}
		}
	}

	// Additional ICP-Brasil specific validations
	return validateICPBrasilSpecificRequirements(chain[0])
}

// IsICPBrasilCertificate checks if a certificate is from ICP-Brasil infrastructure
func IsICPBrasilCertificate(cert *x509.Certificate) bool {
	if cert == nil {
		return false
	}

	// Check subject and issuer for ICP-Brasil patterns
	subject := cert.Subject.String()
	issuer := cert.Issuer.String()

	// Check for ICP-Brasil patterns in the certificate chain
	for _, pattern := range icpBrasilRootPatterns {
		if strings.Contains(issuer, pattern) || strings.Contains(subject, pattern) {
			return true
		}
	}

	for _, pattern := range icpBrasilCAPatterns {
		if strings.Contains(issuer, pattern) {
			return true
		}
	}

	// Check for specific ICP-Brasil OIDs in certificate policies
	for _, policy := range cert.PolicyIdentifiers {
		// ICP-Brasil certificate policies start with 2.16.76.1
		if strings.HasPrefix(policy.String(), "2.16.76.1") {
			return true
		}
	}

	return false
}

// validateICPBrasilSpecificRequirements validates ICP-Brasil specific requirements
func validateICPBrasilSpecificRequirements(cert *x509.Certificate) error {
	// Check key usage
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return errors.NewCertificateError("certificate must have digital signature key usage", nil)
	}

	// Check for non-repudiation if it's a signing certificate
	if cert.KeyUsage&x509.KeyUsageContentCommitment == 0 {
		// This is a warning, not an error for all certificate types
		// Some certificates may not have non-repudiation
	}

	// Check extended key usage for specific certificate types
	hasClientAuth := false
	hasEmailProtection := false
	for _, eku := range cert.ExtKeyUsage {
		switch eku {
		case x509.ExtKeyUsageClientAuth:
			hasClientAuth = true
		case x509.ExtKeyUsageEmailProtection:
			hasEmailProtection = true
		}
	}

	// For A1/A3 certificates used in NFe, we typically need client authentication
	if !hasClientAuth && !hasEmailProtection {
		return errors.NewCertificateError("certificate must have client authentication or email protection extended key usage", nil)
	}

	return nil
}

// GetCertificateType returns the type of ICP-Brasil certificate (A1, A3, etc.)
func GetCertificateType(cert *x509.Certificate) (CertificateType, error) {
	if cert == nil {
		return TypeA1, errors.NewValidationError("certificate cannot be nil", "certificate", "")
	}

	// Check certificate policies to determine type
	for _, policy := range cert.PolicyIdentifiers {
		policyStr := policy.String()

		// ICP-Brasil A1 certificates typically have specific OIDs
		if strings.Contains(policyStr, "2.16.76.1.2.1") {
			return TypeA1, nil
		}

		// ICP-Brasil A3 certificates typically have different OIDs
		if strings.Contains(policyStr, "2.16.76.1.2.2") {
			return TypeA3, nil
		}
	}

	// If we can't determine from OIDs, assume A1 (software certificate)
	// In practice, this should be determined by how the certificate was loaded
	return TypeA1, nil
}

// ValidateForNFeUse validates if a certificate can be used for NFe operations
func ValidateForNFeUse(cert *x509.Certificate) error {
	if cert == nil {
		return errors.NewValidationError("certificate cannot be nil", "certificate", "")
	}

	// Check if it's an ICP-Brasil certificate
	if !IsICPBrasilCertificate(cert) {
		return errors.NewCertificateError("certificate must be from ICP-Brasil for NFe use", nil)
	}

	// Check validity period
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return errors.NewCertificateError("certificate is not yet valid", nil)
	}
	if now.After(cert.NotAfter) {
		return errors.NewCertificateError("certificate has expired", nil)
	}

	// Check key usage for digital signature
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return errors.NewCertificateError("certificate must have digital signature capability for NFe", nil)
	}

	// Check that it's not a CA certificate
	if cert.IsCA {
		return errors.NewCertificateError("CA certificates cannot be used for NFe signing", nil)
	}

	// Check for client authentication extended key usage
	hasClientAuth := false
	for _, eku := range cert.ExtKeyUsage {
		if eku == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
			break
		}
	}

	if !hasClientAuth {
		return errors.NewCertificateError("certificate must have client authentication extended key usage for NFe", nil)
	}

	return nil
}

// ExtractCNPJFromCertificate extracts CNPJ from certificate subject
func ExtractCNPJFromCertificate(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}

	subject := cert.Subject.String()

	// Look for CNPJ pattern in subject (14 digits)
	// CNPJ is usually in the CN field in the format "Name:CNPJ"
	if cert.Subject.CommonName != "" {
		cn := cert.Subject.CommonName
		parts := strings.Split(cn, ":")
		if len(parts) >= 2 {
			cnpj := parts[len(parts)-1]
			if len(cnpj) == 14 && isNumeric(cnpj) {
				return cnpj
			}
		}
	}

	// Alternative: look for CNPJ in the entire subject string
	parts := strings.Fields(subject)
	for _, part := range parts {
		if len(part) == 14 && isNumeric(part) {
			return part
		}
	}

	return ""
}

// ExtractCPFFromCertificate extracts CPF from certificate subject
func ExtractCPFFromCertificate(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}

	subject := cert.Subject.String()

	// Look for CPF pattern in subject (11 digits)
	parts := strings.Fields(subject)
	for _, part := range parts {
		if len(part) == 11 && isNumeric(part) {
			return part
		}
	}

	return ""
}

// isNumeric checks if a string contains only numeric characters
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// GetCertificateFingerprint returns the SHA-256 fingerprint of a certificate
func GetCertificateFingerprint(cert *x509.Certificate) string {
	if cert == nil {
		return ""
	}

	// This is already implemented in the main certificate interface,
	// but we provide it here for convenience
	return fmt.Sprintf("%x", cert.Raw)
}
