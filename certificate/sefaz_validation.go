package certificate

import (
	"fmt"
	"strings"

	"github.com/beevik/etree"
)

// SEFAZ298ValidationResult contains validation results for SEFAZ error 298
type SEFAZ298ValidationResult struct {
	IsValid                bool     `json:"isValid"`
	HasReferenceURI        bool     `json:"hasReferenceURI"`
	HasIdAttribute         bool     `json:"hasIdAttribute"`
	HasEnvelopedTransform  bool     `json:"hasEnvelopedTransform"`
	HasC14NTransform       bool     `json:"hasC14NTransform"`
	HasCorrectNamespace    bool     `json:"hasCorrectNamespace"`
	HasSignatureMethod     bool     `json:"hasSignatureMethod"`
	HasDigestMethod        bool     `json:"hasDigestMethod"`
	HasDigestValue         bool     `json:"hasDigestValue"`
	HasSignatureValue      bool     `json:"hasSignatureValue"`
	Issues                 []string `json:"issues"`
}

// ValidateSEFAZ298Compliance validates XML signature against SEFAZ requirements for error 298
func ValidateSEFAZ298Compliance(signedXML string) (*SEFAZ298ValidationResult, error) {
	result := &SEFAZ298ValidationResult{
		Issues: make([]string, 0),
	}

	// Parse the signed XML
	doc := etree.NewDocument()
	if err := doc.ReadFromString(signedXML); err != nil {
		return nil, fmt.Errorf("failed to parse signed XML: %v", err)
	}

	// Find the Signature element
	signature := doc.FindElement(".//Signature")
	if signature == nil {
		result.Issues = append(result.Issues, "No Signature element found")
		return result, nil
	}

	// 1. Check for correct XML namespace
	xmlns := signature.SelectAttr("xmlns")
	if xmlns != nil && xmlns.Value == "http://www.w3.org/2000/09/xmldsig#" {
		result.HasCorrectNamespace = true
	} else {
		result.Issues = append(result.Issues, "Missing or incorrect xmlns namespace on Signature")
	}

	// Find SignedInfo
	signedInfo := signature.FindElement("SignedInfo")
	if signedInfo == nil {
		result.Issues = append(result.Issues, "No SignedInfo element found")
		return result, nil
	}

	// 2. Check for SignatureMethod
	signatureMethod := signedInfo.FindElement("SignatureMethod")
	if signatureMethod != nil {
		algAttr := signatureMethod.SelectAttr("Algorithm")
		if algAttr != nil && algAttr.Value == "http://www.w3.org/2000/09/xmldsig#rsa-sha1" {
			result.HasSignatureMethod = true
		} else {
			result.Issues = append(result.Issues, "Missing or incorrect SignatureMethod algorithm")
		}
	} else {
		result.Issues = append(result.Issues, "No SignatureMethod element found")
	}

	// Find Reference
	reference := signedInfo.FindElement("Reference")
	if reference == nil {
		result.Issues = append(result.Issues, "No Reference element found")
		return result, nil
	}

	// 3. Check for Reference URI (should sign the Id attribute)
	uriAttr := reference.SelectAttr("URI")
	if uriAttr != nil && uriAttr.Value != "" && strings.HasPrefix(uriAttr.Value, "#") {
		result.HasReferenceURI = true
		
		// Extract the ID from URI and check if corresponding element exists
		idValue := strings.TrimPrefix(uriAttr.Value, "#")
		if idValue != "" {
			// Check if element with this ID exists
			elementWithId := doc.FindElement(fmt.Sprintf(".//*[@Id='%s']", idValue))
			if elementWithId != nil {
				result.HasIdAttribute = true
			} else {
				result.Issues = append(result.Issues, fmt.Sprintf("No element found with Id='%s'", idValue))
			}
		}
	} else {
		result.Issues = append(result.Issues, "Missing Reference URI or URI does not start with #")
	}

	// 4. Check for Transforms
	transforms := reference.FindElement("Transforms")
	if transforms == nil {
		result.Issues = append(result.Issues, "No Transforms element found")
	} else {
		transformElements := transforms.FindElements("Transform")
		
		envelopedFound := false
		c14nFound := false
		
		for _, transform := range transformElements {
			algAttr := transform.SelectAttr("Algorithm")
			if algAttr != nil {
				switch algAttr.Value {
				case "http://www.w3.org/2000/09/xmldsig#enveloped-signature":
					envelopedFound = true
					result.HasEnvelopedTransform = true
				case "http://www.w3.org/TR/2001/REC-xml-c14n-20010315":
					c14nFound = true
					result.HasC14NTransform = true
				}
			}
		}
		
		if !envelopedFound {
			result.Issues = append(result.Issues, "Missing enveloped-signature Transform")
		}
		
		if !c14nFound {
			result.Issues = append(result.Issues, "Missing C14N Transform")
		}
	}

	// 5. Check for DigestMethod
	digestMethod := reference.FindElement("DigestMethod")
	if digestMethod != nil {
		algAttr := digestMethod.SelectAttr("Algorithm")
		if algAttr != nil && algAttr.Value == "http://www.w3.org/2000/09/xmldsig#sha1" {
			result.HasDigestMethod = true
		} else {
			result.Issues = append(result.Issues, "Missing or incorrect DigestMethod algorithm")
		}
	} else {
		result.Issues = append(result.Issues, "No DigestMethod element found")
	}

	// 6. Check for DigestValue
	digestValue := reference.FindElement("DigestValue")
	if digestValue != nil && digestValue.Text() != "" {
		result.HasDigestValue = true
	} else {
		result.Issues = append(result.Issues, "Missing or empty DigestValue")
	}

	// 7. Check for SignatureValue
	signatureValue := signature.FindElement("SignatureValue")
	if signatureValue != nil && signatureValue.Text() != "" {
		result.HasSignatureValue = true
	} else {
		result.Issues = append(result.Issues, "Missing or empty SignatureValue")
	}

	// Determine overall validity
	result.IsValid = result.HasReferenceURI &&
		result.HasIdAttribute &&
		result.HasEnvelopedTransform &&
		result.HasC14NTransform &&
		result.HasCorrectNamespace &&
		result.HasSignatureMethod &&
		result.HasDigestMethod &&
		result.HasDigestValue &&
		result.HasSignatureValue

	return result, nil
}

// PrintSEFAZ298ValidationReport prints a detailed validation report
func PrintSEFAZ298ValidationReport(result *SEFAZ298ValidationResult) {
	fmt.Println("=== SEFAZ Error 298 Validation Report ===")
	fmt.Printf("Overall Valid: %v\n", result.IsValid)
	fmt.Println()
	
	fmt.Println("Requirements Check:")
	fmt.Printf("✓ Reference URI present: %v\n", result.HasReferenceURI)
	fmt.Printf("✓ Id attribute signed: %v\n", result.HasIdAttribute)
	fmt.Printf("✓ Enveloped Transform: %v\n", result.HasEnvelopedTransform)
	fmt.Printf("✓ C14N Transform: %v\n", result.HasC14NTransform)
	fmt.Printf("✓ Correct namespace: %v\n", result.HasCorrectNamespace)
	fmt.Printf("✓ Signature method: %v\n", result.HasSignatureMethod)
	fmt.Printf("✓ Digest method: %v\n", result.HasDigestMethod)
	fmt.Printf("✓ Digest value: %v\n", result.HasDigestValue)
	fmt.Printf("✓ Signature value: %v\n", result.HasSignatureValue)
	
	if len(result.Issues) > 0 {
		fmt.Println()
		fmt.Println("Issues Found:")
		for i, issue := range result.Issues {
			fmt.Printf("%d. %s\n", i+1, issue)
		}
	}
	
	fmt.Println()
	if result.IsValid {
		fmt.Println("🟢 SIGNATURE SHOULD BE ACCEPTED BY SEFAZ")
	} else {
		fmt.Println("🔴 SIGNATURE WILL BE REJECTED BY SEFAZ (Error 298)")
	}
	fmt.Println("==========================================")
}

// ValidateAndPrintSEFAZ298 is a convenience function to validate and print results
func ValidateAndPrintSEFAZ298(signedXML string) error {
	result, err := ValidateSEFAZ298Compliance(signedXML)
	if err != nil {
		return err
	}
	
	PrintSEFAZ298ValidationReport(result)
	return nil
}