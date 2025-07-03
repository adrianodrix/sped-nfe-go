package converter

import (
	"testing"
)

func TestNewValidatorSimple(t *testing.T) {
	config := &LayoutConfig{
		Name:    "Test Layout",
		Version: "4.00",
		Structure: map[string]string{
			"A": "A|versao|Id|pk_nItem|",
			"B": "B|cUF|cNF|natOp|mod|serie|nNF|dhEmi|dhSaiEnt|tpNF|idDest|cMunFG|tpImp|tpEmis|cDV|tpAmb|finNFe|indFinal|indPres|indIntermed|procEmi|verProc|dhCont|xJust|",
		},
	}

	validator := NewValidator(config)
	
	if validator == nil {
		t.Error("Expected validator but got nil")
	}
	
	if validator.layoutConfig != config {
		t.Error("Validator config not set correctly")
	}
}

func TestValidationErrorSimple(t *testing.T) {
	err := &ValidationError{
		Line:    5,
		Tag:     "C02",
		Field:   "CNPJ",
		Message: "Invalid CNPJ format",
	}

	expected := "line 5, tag C02, field CNPJ: Invalid CNPJ format"
	result := err.Error()

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}