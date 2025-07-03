// Package nfe provides comprehensive tax structures for NFe documents.
package nfe

import "encoding/xml"

// ICMS represents ICMS tax with all possible modalities
type ICMS struct {
	XMLName xml.Name `xml:"ICMS"`

	// Normal ICMS modalities
	ICMS00   *ICMS00   `xml:"ICMS00,omitempty"`   // Taxed normally
	ICMS10   *ICMS10   `xml:"ICMS10,omitempty"`   // Taxed with ST
	ICMS20   *ICMS20   `xml:"ICMS20,omitempty"`   // Reduced tax base
	ICMS30   *ICMS30   `xml:"ICMS30,omitempty"`   // Exempt/non-taxed with ST charge
	ICMS40   *ICMS40   `xml:"ICMS40,omitempty"`   // Exempt
	ICMS41   *ICMS41   `xml:"ICMS41,omitempty"`   // Non-taxed
	ICMS50   *ICMS50   `xml:"ICMS50,omitempty"`   // Suspended
	ICMS51   *ICMS51   `xml:"ICMS51,omitempty"`   // Deferred
	ICMS60   *ICMS60   `xml:"ICMS60,omitempty"`   // Charged previously by ST
	ICMS70   *ICMS70   `xml:"ICMS70,omitempty"`   // Reduced tax base with ST charge
	ICMS90   *ICMS90   `xml:"ICMS90,omitempty"`   // Other
	ICMSPart *ICMSPart `xml:"ICMSPart,omitempty"` // Partial partition
	ICMSST   *ICMSST   `xml:"ICMSST,omitempty"`   // Substitution group

	// Simples Nacional modalities
	ICMSSN101 *ICMSSN101 `xml:"ICMSSN101,omitempty"` // Simples Nacional - with permitted credit
	ICMSSN102 *ICMSSN102 `xml:"ICMSSN102,omitempty"` // Simples Nacional - without credit
	ICMSSN103 *ICMSSN103 `xml:"ICMSSN103,omitempty"` // Simples Nacional - exempt
	ICMSSN201 *ICMSSN201 `xml:"ICMSSN201,omitempty"` // Simples Nacional - with ST permitted credit
	ICMSSN202 *ICMSSN202 `xml:"ICMSSN202,omitempty"` // Simples Nacional - with ST without credit
	ICMSSN203 *ICMSSN203 `xml:"ICMSSN203,omitempty"` // Simples Nacional - with ST exempt
	ICMSSN300 *ICMSSN300 `xml:"ICMSSN300,omitempty"` // Simples Nacional - immunity
	ICMSSN400 *ICMSSN400 `xml:"ICMSSN400,omitempty"` // Simples Nacional - not taxed
	ICMSSN500 *ICMSSN500 `xml:"ICMSSN500,omitempty"` // Simples Nacional - retido por ST
	ICMSSN900 *ICMSSN900 `xml:"ICMSSN900,omitempty"` // Simples Nacional - other
}

// ICMS00 - Taxed normally (CST 00)
type ICMS00 struct {
	XMLName xml.Name `xml:"ICMS00"`
	Orig    string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CST     string   `xml:"CST" validate:"required,eq=00"`                    // Tax situation
	ModBC   string   `xml:"modBC" validate:"required,oneof=0 1 2 3"`          // Tax base modality
	VBC     string   `xml:"vBC" validate:"required"`                          // Tax base value
	PICMS   string   `xml:"pICMS" validate:"required"`                        // ICMS rate
	VICMS   string   `xml:"vICMS" validate:"required"`                        // ICMS value
	PFCP    string   `xml:"pFCP,omitempty"`                                   // FCP rate
	VFCP    string   `xml:"vFCP,omitempty"`                                   // FCP value
}

// ICMS10 - Taxed with ST (CST 10)
type ICMS10 struct {
	XMLName  xml.Name `xml:"ICMS10"`
	Orig     string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CST      string   `xml:"CST" validate:"required,eq=10"`                    // Tax situation
	ModBC    string   `xml:"modBC" validate:"required,oneof=0 1 2 3"`          // Tax base modality
	VBC      string   `xml:"vBC" validate:"required"`                          // Tax base value
	PICMS    string   `xml:"pICMS" validate:"required"`                        // ICMS rate
	VICMS    string   `xml:"vICMS" validate:"required"`                        // ICMS value
	VBCFCP   string   `xml:"vBCFCP,omitempty"`                                 // FCP tax base
	PFCP     string   `xml:"pFCP,omitempty"`                                   // FCP rate
	VFCP     string   `xml:"vFCP,omitempty"`                                   // FCP value
	ModBCST  string   `xml:"modBCST" validate:"required,oneof=0 1 2 3 4 5 6"`  // ST tax base modality
	PMVAST   string   `xml:"pMVAST,omitempty"`                                 // ST value added margin
	PRedBCST string   `xml:"pRedBCST,omitempty"`                               // ST tax base reduction
	VBCST    string   `xml:"vBCST" validate:"required"`                        // ST tax base value
	PICMSST  string   `xml:"pICMSST" validate:"required"`                      // ST ICMS rate
	VICMSST  string   `xml:"vICMSST" validate:"required"`                      // ST ICMS value
	VBCFCPST string   `xml:"vBCFCPST,omitempty"`                               // ST FCP tax base
	PFCPST   string   `xml:"pFCPST,omitempty"`                                 // ST FCP rate
	VFCPST   string   `xml:"vFCPST,omitempty"`                                 // ST FCP value
}

// ICMS20 - Reduced tax base (CST 20)
type ICMS20 struct {
	XMLName    xml.Name `xml:"ICMS20"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`       // Origin
	CST        string   `xml:"CST" validate:"required,eq=20"`                          // Tax situation
	ModBC      string   `xml:"modBC" validate:"required,oneof=0 1 2 3"`                // Tax base modality
	PRedBC     string   `xml:"pRedBC" validate:"required"`                             // Tax base reduction
	VBC        string   `xml:"vBC" validate:"required"`                                // Tax base value
	PICMS      string   `xml:"pICMS" validate:"required"`                              // ICMS rate
	VICMS      string   `xml:"vICMS" validate:"required"`                              // ICMS value
	VBCFCP     string   `xml:"vBCFCP,omitempty"`                                       // FCP tax base
	PFCP       string   `xml:"pFCP,omitempty"`                                         // FCP rate
	VFCP       string   `xml:"vFCP,omitempty"`                                         // FCP value
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                   // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=3 9 12"` // ICMS relief reason
}

// ICMS30 - Exempt/non-taxed with ST charge (CST 30)
type ICMS30 struct {
	XMLName    xml.Name `xml:"ICMS30"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`      // Origin
	CST        string   `xml:"CST" validate:"required,eq=30"`                         // Tax situation
	ModBCST    string   `xml:"modBCST" validate:"required,oneof=0 1 2 3 4 5 6"`       // ST tax base modality
	PMVAST     string   `xml:"pMVAST,omitempty"`                                      // ST value added margin
	PRedBCST   string   `xml:"pRedBCST,omitempty"`                                    // ST tax base reduction
	VBCST      string   `xml:"vBCST" validate:"required"`                             // ST tax base value
	PICMSST    string   `xml:"pICMSST" validate:"required"`                           // ST ICMS rate
	VICMSST    string   `xml:"vICMSST" validate:"required"`                           // ST ICMS value
	VBCFCPST   string   `xml:"vBCFCPST,omitempty"`                                    // ST FCP tax base
	PFCPST     string   `xml:"pFCPST,omitempty"`                                      // ST FCP rate
	VFCPST     string   `xml:"vFCPST,omitempty"`                                      // ST FCP value
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                  // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=6 7 9"` // ICMS relief reason
}

// ICMS40 - Exempt (CST 40)
type ICMS40 struct {
	XMLName    xml.Name `xml:"ICMS40"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`                            // Origin
	CST        string   `xml:"CST" validate:"required,oneof=40 41 50"`                                      // Tax situation
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                                        // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=1 3 4 5 6 7 8 9 10 11 16 90"` // ICMS relief reason
}

// ICMS41 - Non-taxed (CST 41)
type ICMS41 struct {
	XMLName    xml.Name `xml:"ICMS41"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`                            // Origin
	CST        string   `xml:"CST" validate:"required,eq=41"`                                               // Tax situation
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                                        // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=1 3 4 5 6 7 8 9 10 11 16 90"` // ICMS relief reason
}

// ICMS50 - Suspended (CST 50)
type ICMS50 struct {
	XMLName    xml.Name `xml:"ICMS50"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`                            // Origin
	CST        string   `xml:"CST" validate:"required,eq=50"`                                               // Tax situation
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                                        // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=1 3 4 5 6 7 8 9 10 11 16 90"` // ICMS relief reason
}

// ICMS51 - Deferred (CST 51)
type ICMS51 struct {
	XMLName  xml.Name `xml:"ICMS51"`
	Orig     string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`   // Origin
	CST      string   `xml:"CST" validate:"required,eq=51"`                      // Tax situation
	ModBC    string   `xml:"modBC,omitempty" validate:"omitempty,oneof=0 1 2 3"` // Tax base modality
	PRedBC   string   `xml:"pRedBC,omitempty"`                                   // Tax base reduction
	VBC      string   `xml:"vBC,omitempty"`                                      // Tax base value
	PICMS    string   `xml:"pICMS,omitempty"`                                    // ICMS rate
	VICMSOp  string   `xml:"vICMSOp,omitempty"`                                  // ICMS operation value
	PDif     string   `xml:"pDif,omitempty"`                                     // Deferral percentage
	VICMSDif string   `xml:"vICMSDif,omitempty"`                                 // ICMS deferred value
	VICMS    string   `xml:"vICMS,omitempty"`                                    // ICMS value
	VBCFCP   string   `xml:"vBCFCP,omitempty"`                                   // FCP tax base
	PFCP     string   `xml:"pFCP,omitempty"`                                     // FCP rate
	VFCP     string   `xml:"vFCP,omitempty"`                                     // FCP value
}

// ICMS60 - Charged previously by ST (CST 60)
type ICMS60 struct {
	XMLName    xml.Name `xml:"ICMS60"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`       // Origin
	CST        string   `xml:"CST" validate:"required,eq=60"`                          // Tax situation
	VBCST      string   `xml:"vBCSTRet,omitempty"`                                     // ST withholding tax base
	PST        string   `xml:"pST,omitempty"`                                          // ST withholding rate
	VICMSSubst string   `xml:"vICMSSubst,omitempty"`                                   // ST ICMS substitution value
	VBCFCPST   string   `xml:"vBCFCPSTRet,omitempty"`                                  // ST FCP withholding tax base
	PFCPST     string   `xml:"pFCPSTRet,omitempty"`                                    // ST FCP withholding rate
	VFCPSTRet  string   `xml:"vFCPSTRet,omitempty"`                                    // ST FCP withholding value
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                   // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=3 9 12"` // ICMS relief reason
}

// ICMS70 - Reduced tax base with ST charge (CST 70)
type ICMS70 struct {
	XMLName    xml.Name `xml:"ICMS70"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`       // Origin
	CST        string   `xml:"CST" validate:"required,eq=70"`                          // Tax situation
	ModBC      string   `xml:"modBC" validate:"required,oneof=0 1 2 3"`                // Tax base modality
	PRedBC     string   `xml:"pRedBC" validate:"required"`                             // Tax base reduction
	VBC        string   `xml:"vBC" validate:"required"`                                // Tax base value
	PICMS      string   `xml:"pICMS" validate:"required"`                              // ICMS rate
	VICMS      string   `xml:"vICMS" validate:"required"`                              // ICMS value
	VBCFCP     string   `xml:"vBCFCP,omitempty"`                                       // FCP tax base
	PFCP       string   `xml:"pFCP,omitempty"`                                         // FCP rate
	VFCP       string   `xml:"vFCP,omitempty"`                                         // FCP value
	ModBCST    string   `xml:"modBCST" validate:"required,oneof=0 1 2 3 4 5 6"`        // ST tax base modality
	PMVAST     string   `xml:"pMVAST,omitempty"`                                       // ST value added margin
	PRedBCST   string   `xml:"pRedBCST,omitempty"`                                     // ST tax base reduction
	VBCST      string   `xml:"vBCST" validate:"required"`                              // ST tax base value
	PICMSST    string   `xml:"pICMSST" validate:"required"`                            // ST ICMS rate
	VICMSST    string   `xml:"vICMSST" validate:"required"`                            // ST ICMS value
	VBCFCPST   string   `xml:"vBCFCPST,omitempty"`                                     // ST FCP tax base
	PFCPST     string   `xml:"pFCPST,omitempty"`                                       // ST FCP rate
	VFCPST     string   `xml:"vFCPST,omitempty"`                                       // ST FCP value
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                   // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=3 9 12"` // ICMS relief reason
}

// ICMS90 - Other (CST 90)
type ICMS90 struct {
	XMLName    xml.Name `xml:"ICMS90"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`           // Origin
	CST        string   `xml:"CST" validate:"required,eq=90"`                              // Tax situation
	ModBC      string   `xml:"modBC,omitempty" validate:"omitempty,oneof=0 1 2 3"`         // Tax base modality
	VBC        string   `xml:"vBC,omitempty"`                                              // Tax base value
	PRedBC     string   `xml:"pRedBC,omitempty"`                                           // Tax base reduction
	PICMS      string   `xml:"pICMS,omitempty"`                                            // ICMS rate
	VICMS      string   `xml:"vICMS,omitempty"`                                            // ICMS value
	VBCFCP     string   `xml:"vBCFCP,omitempty"`                                           // FCP tax base
	PFCP       string   `xml:"pFCP,omitempty"`                                             // FCP rate
	VFCP       string   `xml:"vFCP,omitempty"`                                             // FCP value
	ModBCST    string   `xml:"modBCST,omitempty" validate:"omitempty,oneof=0 1 2 3 4 5 6"` // ST tax base modality
	PMVAST     string   `xml:"pMVAST,omitempty"`                                           // ST value added margin
	PRedBCST   string   `xml:"pRedBCST,omitempty"`                                         // ST tax base reduction
	VBCST      string   `xml:"vBCST,omitempty"`                                            // ST tax base value
	PICMSST    string   `xml:"pICMSST,omitempty"`                                          // ST ICMS rate
	VICMSST    string   `xml:"vICMSST,omitempty"`                                          // ST ICMS value
	VBCFCPST   string   `xml:"vBCFCPST,omitempty"`                                         // ST FCP tax base
	PFCPST     string   `xml:"pFCPST,omitempty"`                                           // ST FCP rate
	VFCPST     string   `xml:"vFCPST,omitempty"`                                           // ST FCP value
	VICMSDeson string   `xml:"vICMSDeson,omitempty"`                                       // ICMS relief value
	MotDesICMS string   `xml:"motDesICMS,omitempty" validate:"omitempty,oneof=3 9 12"`     // ICMS relief reason
}

// ICMSPart - Partial partition
type ICMSPart struct {
	XMLName  xml.Name `xml:"ICMSPart"`
	Orig     string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CST      string   `xml:"CST" validate:"required,eq=10 90"`                 // Tax situation
	ModBC    string   `xml:"modBC" validate:"required,oneof=0 1 2 3"`          // Tax base modality
	VBC      string   `xml:"vBC" validate:"required"`                          // Tax base value
	PRedBC   string   `xml:"pRedBC,omitempty"`                                 // Tax base reduction
	PICMS    string   `xml:"pICMS" validate:"required"`                        // ICMS rate
	VICMS    string   `xml:"vICMS" validate:"required"`                        // ICMS value
	ModBCST  string   `xml:"modBCST" validate:"required,oneof=0 1 2 3 4 5 6"`  // ST tax base modality
	PMVAST   string   `xml:"pMVAST,omitempty"`                                 // ST value added margin
	PRedBCST string   `xml:"pRedBCST,omitempty"`                               // ST tax base reduction
	VBCST    string   `xml:"vBCST" validate:"required"`                        // ST tax base value
	PICMSST  string   `xml:"pICMSST" validate:"required"`                      // ST ICMS rate
	VICMSST  string   `xml:"vICMSST" validate:"required"`                      // ST ICMS value
	PBCOp    string   `xml:"pBCOp" validate:"required"`                        // Operation base percentage
	UFST     string   `xml:"UFST" validate:"required,len=2"`                   // ST UF
}

// ICMSST - Substitution group
type ICMSST struct {
	XMLName    xml.Name `xml:"ICMSST"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CST        string   `xml:"CST" validate:"required,eq=41 60"`                 // Tax situation
	VBCST      string   `xml:"vBCSTRet" validate:"required"`                     // ST withholding tax base
	PST        string   `xml:"pST,omitempty"`                                    // ST withholding rate
	VICMSSubst string   `xml:"vICMSSubst" validate:"required"`                   // ST ICMS substitution value
	VBCFCPST   string   `xml:"vBCFCPSTRet,omitempty"`                            // ST FCP withholding tax base
	PFCPST     string   `xml:"pFCPSTRet,omitempty"`                              // ST FCP withholding rate
	VFCPSTRet  string   `xml:"vFCPSTRet,omitempty"`                              // ST FCP withholding value
}

// Simples Nacional ICMS structures

// ICMSSN101 - Simples Nacional with permitted credit
type ICMSSN101 struct {
	XMLName     xml.Name `xml:"ICMSSN101"`
	Orig        string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN       string   `xml:"CSOSN" validate:"required,eq=101"`                 // Simples Nacional tax situation
	PCredSN     string   `xml:"pCredSN" validate:"required"`                      // SN credit rate
	VCredICMSSN string   `xml:"vCredICMSSN" validate:"required"`                  // SN ICMS credit value
}

// ICMSSN102 - Simples Nacional without credit
type ICMSSN102 struct {
	XMLName xml.Name `xml:"ICMSSN102"`
	Orig    string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN   string   `xml:"CSOSN" validate:"required,oneof=102 103 300 400"`  // Simples Nacional tax situation
}

// ICMSSN103 - Simples Nacional exempt
type ICMSSN103 struct {
	XMLName xml.Name `xml:"ICMSSN103"`
	Orig    string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN   string   `xml:"CSOSN" validate:"required,eq=103"`                 // Simples Nacional tax situation
}

// ICMSSN201 - Simples Nacional with ST permitted credit
type ICMSSN201 struct {
	XMLName     xml.Name `xml:"ICMSSN201"`
	Orig        string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN       string   `xml:"CSOSN" validate:"required,eq=201"`                 // Simples Nacional tax situation
	ModBCST     string   `xml:"modBCST" validate:"required,oneof=0 1 2 3 4 5 6"`  // ST tax base modality
	PMVAST      string   `xml:"pMVAST,omitempty"`                                 // ST value added margin
	PRedBCST    string   `xml:"pRedBCST,omitempty"`                               // ST tax base reduction
	VBCST       string   `xml:"vBCST" validate:"required"`                        // ST tax base value
	PICMSST     string   `xml:"pICMSST" validate:"required"`                      // ST ICMS rate
	VICMSST     string   `xml:"vICMSST" validate:"required"`                      // ST ICMS value
	VBCFCPST    string   `xml:"vBCFCPST,omitempty"`                               // ST FCP tax base
	PFCPST      string   `xml:"pFCPST,omitempty"`                                 // ST FCP rate
	VFCPST      string   `xml:"vFCPST,omitempty"`                                 // ST FCP value
	PCredSN     string   `xml:"pCredSN" validate:"required"`                      // SN credit rate
	VCredICMSSN string   `xml:"vCredICMSSN" validate:"required"`                  // SN ICMS credit value
}

// ICMSSN202 - Simples Nacional with ST without credit
type ICMSSN202 struct {
	XMLName  xml.Name `xml:"ICMSSN202"`
	Orig     string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN    string   `xml:"CSOSN" validate:"required,oneof=202 203"`          // Simples Nacional tax situation
	ModBCST  string   `xml:"modBCST" validate:"required,oneof=0 1 2 3 4 5 6"`  // ST tax base modality
	PMVAST   string   `xml:"pMVAST,omitempty"`                                 // ST value added margin
	PRedBCST string   `xml:"pRedBCST,omitempty"`                               // ST tax base reduction
	VBCST    string   `xml:"vBCST" validate:"required"`                        // ST tax base value
	PICMSST  string   `xml:"pICMSST" validate:"required"`                      // ST ICMS rate
	VICMSST  string   `xml:"vICMSST" validate:"required"`                      // ST ICMS value
	VBCFCPST string   `xml:"vBCFCPST,omitempty"`                               // ST FCP tax base
	PFCPST   string   `xml:"pFCPST,omitempty"`                                 // ST FCP rate
	VFCPST   string   `xml:"vFCPST,omitempty"`                                 // ST FCP value
}

// ICMSSN203 - Simples Nacional with ST exempt
type ICMSSN203 struct {
	XMLName  xml.Name `xml:"ICMSSN203"`
	Orig     string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN    string   `xml:"CSOSN" validate:"required,eq=203"`                 // Simples Nacional tax situation
	ModBCST  string   `xml:"modBCST" validate:"required,oneof=0 1 2 3 4 5 6"`  // ST tax base modality
	PMVAST   string   `xml:"pMVAST,omitempty"`                                 // ST value added margin
	PRedBCST string   `xml:"pRedBCST,omitempty"`                               // ST tax base reduction
	VBCST    string   `xml:"vBCST" validate:"required"`                        // ST tax base value
	PICMSST  string   `xml:"pICMSST" validate:"required"`                      // ST ICMS rate
	VICMSST  string   `xml:"vICMSST" validate:"required"`                      // ST ICMS value
	VBCFCPST string   `xml:"vBCFCPST,omitempty"`                               // ST FCP tax base
	PFCPST   string   `xml:"pFCPST,omitempty"`                                 // ST FCP rate
	VFCPST   string   `xml:"vFCPST,omitempty"`                                 // ST FCP value
}

// ICMSSN300 - Simples Nacional immunity
type ICMSSN300 struct {
	XMLName xml.Name `xml:"ICMSSN300"`
	Orig    string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN   string   `xml:"CSOSN" validate:"required,eq=300"`                 // Simples Nacional tax situation
}

// ICMSSN400 - Simples Nacional not taxed
type ICMSSN400 struct {
	XMLName xml.Name `xml:"ICMSSN400"`
	Orig    string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN   string   `xml:"CSOSN" validate:"required,eq=400"`                 // Simples Nacional tax situation
}

// ICMSSN500 - Simples Nacional retido por ST
type ICMSSN500 struct {
	XMLName    xml.Name `xml:"ICMSSN500"`
	Orig       string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"` // Origin
	CSOSN      string   `xml:"CSOSN" validate:"required,eq=500"`                 // Simples Nacional tax situation
	VBCST      string   `xml:"vBCSTRet,omitempty"`                               // ST withholding tax base
	PST        string   `xml:"pST,omitempty"`                                    // ST withholding rate
	VICMSSubst string   `xml:"vICMSSubst,omitempty"`                             // ST ICMS substitution value
	VBCFCPST   string   `xml:"vBCFCPSTRet,omitempty"`                            // ST FCP withholding tax base
	PFCPST     string   `xml:"pFCPSTRet,omitempty"`                              // ST FCP withholding rate
	VFCPSTRet  string   `xml:"vFCPSTRet,omitempty"`                              // ST FCP withholding value
}

// ICMSSN900 - Simples Nacional other
type ICMSSN900 struct {
	XMLName     xml.Name `xml:"ICMSSN900"`
	Orig        string   `xml:"orig" validate:"required,oneof=0 1 2 3 4 5 6 7 8"`           // Origin
	CSOSN       string   `xml:"CSOSN" validate:"required,eq=900"`                           // Simples Nacional tax situation
	ModBC       string   `xml:"modBC,omitempty" validate:"omitempty,oneof=0 1 2 3"`         // Tax base modality
	VBC         string   `xml:"vBC,omitempty"`                                              // Tax base value
	PRedBC      string   `xml:"pRedBC,omitempty"`                                           // Tax base reduction
	PICMS       string   `xml:"pICMS,omitempty"`                                            // ICMS rate
	VICMS       string   `xml:"vICMS,omitempty"`                                            // ICMS value
	ModBCST     string   `xml:"modBCST,omitempty" validate:"omitempty,oneof=0 1 2 3 4 5 6"` // ST tax base modality
	PMVAST      string   `xml:"pMVAST,omitempty"`                                           // ST value added margin
	PRedBCST    string   `xml:"pRedBCST,omitempty"`                                         // ST tax base reduction
	VBCST       string   `xml:"vBCST,omitempty"`                                            // ST tax base value
	PICMSST     string   `xml:"pICMSST,omitempty"`                                          // ST ICMS rate
	VICMSST     string   `xml:"vICMSST,omitempty"`                                          // ST ICMS value
	VBCFCPST    string   `xml:"vBCFCPST,omitempty"`                                         // ST FCP tax base
	PFCPST      string   `xml:"pFCPST,omitempty"`                                           // ST FCP rate
	VFCPST      string   `xml:"vFCPST,omitempty"`                                           // ST FCP value
	PCredSN     string   `xml:"pCredSN,omitempty"`                                          // SN credit rate
	VCredICMSSN string   `xml:"vCredICMSSN,omitempty"`                                      // SN ICMS credit value
}

// IPI represents IPI tax information
type IPI struct {
	XMLName  xml.Name `xml:"IPI"`
	CNPJProd string   `xml:"CNPJProd,omitempty" validate:"omitempty,len=14"` // Producer CNPJ
	CSelo    string   `xml:"cSelo,omitempty" validate:"omitempty,max=60"`    // Seal code
	QSelo    string   `xml:"qSelo,omitempty"`                                // Seal quantity
	CEnq     string   `xml:"cEnq" validate:"required,len=3"`                 // Framework code

	// IPI modalities (mutually exclusive)
	IPITrib *IPITrib `xml:"IPITrib,omitempty"` // Taxed IPI
	IPINT   *IPINT   `xml:"IPINT,omitempty"`   // Non-taxed IPI
}

// IPITrib represents taxed IPI
type IPITrib struct {
	XMLName xml.Name `xml:"IPITrib"`
	CST     string   `xml:"CST" validate:"required,oneof=00 49 50 99"` // Tax situation
	VBC     string   `xml:"vBC,omitempty"`                             // Tax base value
	PIPI    string   `xml:"pIPI,omitempty"`                            // IPI rate
	QUnid   string   `xml:"qUnid,omitempty"`                           // Unit quantity
	VUnid   string   `xml:"vUnid,omitempty"`                           // Unit value
	VIPI    string   `xml:"vIPI" validate:"required"`                  // IPI value
}

// IPINT represents non-taxed IPI
type IPINT struct {
	XMLName xml.Name `xml:"IPINT"`
	CST     string   `xml:"CST" validate:"required,oneof=01 02 03 04 05 51 52 53 54 55"` // Tax situation
}

// II represents Import Tax
type II struct {
	XMLName  xml.Name `xml:"II"`
	VBC      string   `xml:"vBC" validate:"required"`      // Tax base value
	VDespAdu string   `xml:"vDespAdu" validate:"required"` // Customs expenses
	VII      string   `xml:"vII" validate:"required"`      // Import tax value
	VIOF     string   `xml:"vIOF" validate:"required"`     // IOF value
}

// PIS represents PIS tax information
type PIS struct {
	XMLName xml.Name `xml:"PIS"`

	// PIS modalities (mutually exclusive)
	PISAliq *PISAliq `xml:"PISAliq,omitempty"` // PIS with rate
	PISQtde *PISQtde `xml:"PISQtde,omitempty"` // PIS with quantity
	PISNT   *PISNT   `xml:"PISNT,omitempty"`   // PIS non-taxed
	PISOutr *PISOutr `xml:"PISOutr,omitempty"` // PIS other
}

// PISAliq represents PIS with rate
type PISAliq struct {
	XMLName xml.Name `xml:"PISAliq"`
	CST     string   `xml:"CST" validate:"required,oneof=01 02"` // Tax situation
	VBC     string   `xml:"vBC" validate:"required"`             // Tax base value
	PPIS    string   `xml:"pPIS" validate:"required"`            // PIS rate
	VPIS    string   `xml:"vPIS" validate:"required"`            // PIS value
}

// PISQtde represents PIS with quantity
type PISQtde struct {
	XMLName   xml.Name `xml:"PISQtde"`
	CST       string   `xml:"CST" validate:"required,eq=03"` // Tax situation
	QBCProd   string   `xml:"qBCProd" validate:"required"`   // Base quantity
	VAliqProd string   `xml:"vAliqProd" validate:"required"` // Rate per unit
	VPIS      string   `xml:"vPIS" validate:"required"`      // PIS value
}

// PISNT represents PIS non-taxed
type PISNT struct {
	XMLName xml.Name `xml:"PISNT"`
	CST     string   `xml:"CST" validate:"required,oneof=04 05 06 07 08 09"` // Tax situation
}

// PISOutr represents PIS other
type PISOutr struct {
	XMLName   xml.Name `xml:"PISOutr"`
	CST       string   `xml:"CST" validate:"required,eq=99"` // Tax situation
	VBC       string   `xml:"vBC,omitempty"`                 // Tax base value
	PPIS      string   `xml:"pPIS,omitempty"`                // PIS rate
	QBCProd   string   `xml:"qBCProd,omitempty"`             // Base quantity
	VAliqProd string   `xml:"vAliqProd,omitempty"`           // Rate per unit
	VPIS      string   `xml:"vPIS" validate:"required"`      // PIS value
}

// PISST represents PIS ST tax information
type PISST struct {
	XMLName   xml.Name `xml:"PISST"`
	VBC       string   `xml:"vBC,omitempty"`            // Tax base value
	PPIS      string   `xml:"pPIS,omitempty"`           // PIS rate
	QBCProd   string   `xml:"qBCProd,omitempty"`        // Base quantity
	VAliqProd string   `xml:"vAliqProd,omitempty"`      // Rate per unit
	VPIS      string   `xml:"vPIS" validate:"required"` // PIS value
}

// COFINS represents COFINS tax information
type COFINS struct {
	XMLName xml.Name `xml:"COFINS"`

	// COFINS modalities (mutually exclusive)
	COFINSAliq *COFINSAliq `xml:"COFINSAliq,omitempty"` // COFINS with rate
	COFINSQtde *COFINSQtde `xml:"COFINSQtde,omitempty"` // COFINS with quantity
	COFINSNT   *COFINSNT   `xml:"COFINSNT,omitempty"`   // COFINS non-taxed
	COFINSOutr *COFINSOutr `xml:"COFINSOutr,omitempty"` // COFINS other
}

// COFINSAliq represents COFINS with rate
type COFINSAliq struct {
	XMLName xml.Name `xml:"COFINSAliq"`
	CST     string   `xml:"CST" validate:"required,oneof=01 02"` // Tax situation
	VBC     string   `xml:"vBC" validate:"required"`             // Tax base value
	PCOFINS string   `xml:"pCOFINS" validate:"required"`         // COFINS rate
	VCOFINS string   `xml:"vCOFINS" validate:"required"`         // COFINS value
}

// COFINSQtde represents COFINS with quantity
type COFINSQtde struct {
	XMLName   xml.Name `xml:"COFINSQtde"`
	CST       string   `xml:"CST" validate:"required,eq=03"` // Tax situation
	QBCProd   string   `xml:"qBCProd" validate:"required"`   // Base quantity
	VAliqProd string   `xml:"vAliqProd" validate:"required"` // Rate per unit
	VCOFINS   string   `xml:"vCOFINS" validate:"required"`   // COFINS value
}

// COFINSNT represents COFINS non-taxed
type COFINSNT struct {
	XMLName xml.Name `xml:"COFINSNT"`
	CST     string   `xml:"CST" validate:"required,oneof=04 05 06 07 08 09"` // Tax situation
}

// COFINSOutr represents COFINS other
type COFINSOutr struct {
	XMLName   xml.Name `xml:"COFINSOutr"`
	CST       string   `xml:"CST" validate:"required,eq=99"` // Tax situation
	VBC       string   `xml:"vBC,omitempty"`                 // Tax base value
	PCOFINS   string   `xml:"pCOFINS,omitempty"`             // COFINS rate
	QBCProd   string   `xml:"qBCProd,omitempty"`             // Base quantity
	VAliqProd string   `xml:"vAliqProd,omitempty"`           // Rate per unit
	VCOFINS   string   `xml:"vCOFINS" validate:"required"`   // COFINS value
}

// COFINSST represents COFINS ST tax information
type COFINSST struct {
	XMLName   xml.Name `xml:"COFINSST"`
	VBC       string   `xml:"vBC,omitempty"`               // Tax base value
	PCOFINS   string   `xml:"pCOFINS,omitempty"`           // COFINS rate
	QBCProd   string   `xml:"qBCProd,omitempty"`           // Base quantity
	VAliqProd string   `xml:"vAliqProd,omitempty"`         // Rate per unit
	VCOFINS   string   `xml:"vCOFINS" validate:"required"` // COFINS value
}

// ISSQN represents ISSQN tax information
type ISSQN struct {
	XMLName      xml.Name `xml:"ISSQN"`
	VBC          string   `xml:"vBC" validate:"required"`                         // Tax base value
	VAliq        string   `xml:"vAliq" validate:"required"`                       // Rate value
	VISSQN       string   `xml:"vISSQN" validate:"required"`                      // ISSQN value
	CMunFG       string   `xml:"cMunFG" validate:"required,len=7"`                // Service municipality
	CListServ    string   `xml:"cListServ" validate:"required,min=2,max=5"`       // Service list code
	VDeducao     string   `xml:"vDeducao,omitempty"`                              // Deduction value
	VOutro       string   `xml:"vOutro,omitempty"`                                // Other reductions
	VDescIncond  string   `xml:"vDescIncond,omitempty"`                           // Unconditional discount
	VDescCond    string   `xml:"vDescCond,omitempty"`                             // Conditional discount
	VISSRet      string   `xml:"vISSRet,omitempty"`                               // ISSQN withholding
	IndISS       string   `xml:"indISS" validate:"required,oneof=1 2 3 4 5 6 7"`  // ISSQN indicator
	CServico     string   `xml:"cServico,omitempty" validate:"omitempty,max=20"`  // Service code
	CMun         string   `xml:"cMun,omitempty" validate:"omitempty,len=7"`       // Municipality code
	CPais        string   `xml:"cPais,omitempty" validate:"omitempty,len=4"`      // Country code
	NProcesso    string   `xml:"nProcesso,omitempty" validate:"omitempty,max=30"` // Process number
	IndIncentivo string   `xml:"indIncentivo" validate:"required,oneof=1 2"`      // Incentive indicator
}

// TaxOrigin represents product origin codes
type TaxOrigin string

const (
	OriginNational             TaxOrigin = "0" // Nacional, exceto as indicadas nos códigos 3, 4, 5 e 8
	OriginForeignImported      TaxOrigin = "1" // Estrangeira - Importação direta, exceto a indicada no código 6
	OriginForeignMarket        TaxOrigin = "2" // Estrangeira - Adquirida no mercado interno, exceto a indicada no código 7
	OriginNationalContent      TaxOrigin = "3" // Nacional, mercadoria ou bem com Conteúdo de Importação superior a 40% e inferior ou igual a 70%
	OriginNationalBasicProcess TaxOrigin = "4" // Nacional, cuja produção tenha sido feita em conformidade com os processos produtivos básicos
	OriginNationalContentLess  TaxOrigin = "5" // Nacional, mercadoria ou bem com Conteúdo de Importação inferior ou igual a 40%
	OriginForeignDirectWithout TaxOrigin = "6" // Estrangeira - Importação direta, sem similar nacional, constante em lista da CAMEX e gás natural
	OriginForeignMarketWithout TaxOrigin = "7" // Estrangeira - Adquirida no mercado interno, sem similar nacional, constante em lista da CAMEX e gás natural
	OriginNationalContentMore  TaxOrigin = "8" // Nacional, mercadoria ou bem com Conteúdo de Importação superior a 70%
)

// String returns the string representation of TaxOrigin
func (to TaxOrigin) String() string {
	return string(to)
}

// ValidateTaxes validates tax information
func ValidateTaxes(imposto *Imposto) error {
	if imposto == nil {
		return nil
	}

	// TODO: Implement comprehensive tax validation
	// - Validate mutually exclusive tax modalities
	// - Check required fields for each tax type
	// - Validate CST/CSOSN codes
	// - Check value formats and calculations

	return nil
}
