package converter

// Referencia represents a reference to another fiscal document
type Referencia struct {
	RefNFe        string `json:"refNFe,omitempty"`        // NFe reference
	RefNF         *RefNF `json:"refNF,omitempty"`         // NF model 1/1A reference
	RefNFP        *RefNFP `json:"refNFP,omitempty"`       // Producer NFe reference
	RefCTe        string `json:"refCTe,omitempty"`        // CTe reference
	RefECF        *RefECF `json:"refECF,omitempty"`       // ECF reference
}

// RefNF represents reference to model 1/1A fiscal document
type RefNF struct {
	CUF   string `json:"cUF"`   // UF code
	AAMM  string `json:"AAMM"`  // Year and month
	CNPJ  string `json:"CNPJ"`  // CNPJ
	Mod   string `json:"mod"`   // Model
	Serie string `json:"serie"` // Series
	NNF   string `json:"nNF"`   // Number
}

// RefNFP represents reference to producer NFe
type RefNFP struct {
	CUF   string `json:"cUF"`   // UF code
	AAMM  string `json:"AAMM"`  // Year and month
	CNPJ  string `json:"CNPJ,omitempty"` // CNPJ
	CPF   string `json:"CPF,omitempty"`  // CPF
	IE    string `json:"IE"`    // State registration
	Mod   string `json:"mod"`   // Model
	Serie string `json:"serie"` // Series
	NNF   string `json:"nNF"`   // Number
}

// RefECF represents reference to ECF
type RefECF struct {
	Mod  string `json:"mod"`  // ECF model
	NECF string `json:"nECF"` // ECF number
	NCOO string `json:"nCOO"` // COO number
}