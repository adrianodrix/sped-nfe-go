// Package converter provides parsing functionality for TXT to NFe conversion.
package converter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/adrianodrix/sped-nfe-go/nfe"
)

// Parser handles parsing of TXT lines into structured NFe data
type Parser struct {
	layoutConfig *LayoutConfig
	currentItem  int
	currentNFe   *NFEData
}

// NFEData represents the complete structure of an NFe
type NFEData struct {
	InfNFe        *nfe.InfNFe        `json:"infNFe,omitempty"`
	Identificacao *nfe.Identificacao `json:"identificacao,omitempty"`
	Emitente      *nfe.Emitente      `json:"emitente,omitempty"`
	Destinatario  *nfe.Destinatario  `json:"destinatario,omitempty"`
	Itens         []*nfe.Item        `json:"itens,omitempty"`
	Total         *nfe.Total         `json:"total,omitempty"`
	Transporte    *nfe.Transporte    `json:"transporte,omitempty"`
	InfAdic       *nfe.InfAdicionais `json:"infAdic,omitempty"`
	Referencias   []Referencia       `json:"referencias,omitempty"`
}

// FieldMap represents parsed fields from a TXT line
type FieldMap map[string]string

// NewParser creates a new parser instance
func NewParser(config *LayoutConfig) *Parser {
	return &Parser{
		layoutConfig: config,
		currentItem:  0,
	}
}

// ParseNFe parses TXT lines into NFe data structure
func (p *Parser) ParseNFe(lines []string) (*NFEData, error) {
	p.currentNFe = &NFEData{
		Itens:       []*nfe.Item{},
		Referencias: []Referencia{},
	}
	p.currentItem = 0

	// Process each line
	for lineNum, line := range lines {
		if err := p.parseLine(line, lineNum+1); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum+1, err)
		}
	}

	// Validate required fields
	if err := p.validateRequiredFields(); err != nil {
		return nil, err
	}

	return p.currentNFe, nil
}

// parseLine parses a single TXT line
func (p *Parser) parseLine(line string, lineNum int) error {
	if strings.TrimSpace(line) == "" {
		return nil // Skip empty lines
	}

	// Extract tag and fields
	parts := strings.Split(line, "|")
	if len(parts) < 2 {
		return fmt.Errorf("invalid line format: %s", line)
	}

	tag := parts[0]

	// Get structure for this tag
	structure, exists := p.layoutConfig.Structure[tag]
	if !exists {
		return fmt.Errorf("unknown tag: %s", tag)
	}

	// Parse fields according to structure
	fieldMap, err := p.parseFields(parts, structure)
	if err != nil {
		return fmt.Errorf("failed to parse tag %s: %w", tag, err)
	}

	// Process the tag
	return p.processTag(tag, fieldMap)
}

// parseFields parses TXT fields according to structure definition
func (p *Parser) parseFields(parts []string, structure string) (FieldMap, error) {
	structParts := strings.Split(structure, "|")
	fieldMap := make(FieldMap)

	// Validate field count (subtract 1 for tag, 1 for trailing |)
	expectedFields := len(structParts) - 2
	actualFields := len(parts) - 1

	if actualFields != expectedFields {
		return nil, fmt.Errorf("field count mismatch: expected %d, got %d", expectedFields, actualFields)
	}

	// Map fields to names
	for i := 1; i < len(structParts)-1; i++ {
		fieldName := structParts[i]
		if fieldName != "" && i < len(parts) {
			fieldValue := strings.TrimSpace(parts[i])
			if fieldValue != "" {
				fieldMap[fieldName] = fieldValue
			}
		}
	}

	return fieldMap, nil
}

// processTag processes a parsed tag and updates NFe data
func (p *Parser) processTag(tag string, fields FieldMap) error {
	switch tag {
	case "A":
		return p.processTagA(fields)
	case "B":
		return p.processTagB(fields)
	case "C":
		return p.processTagC(fields)
	case "C02":
		return p.processTagC02(fields)
	case "C02A", "C02a":
		return p.processTagC02A(fields)
	case "C05":
		return p.processTagC05(fields)
	case "E":
		return p.processTagE(fields)
	case "E02":
		return p.processTagE02(fields)
	case "E03":
		return p.processTagE03(fields)
	case "E05":
		return p.processTagE05(fields)
	case "H":
		return p.processTagH(fields)
	case "I":
		return p.processTagI(fields)
	case "M":
		return p.processTagM(fields)
	case "N":
		return p.processTagN(fields)
	case "W":
		return p.processTagW(fields)
	case "X":
		return p.processTagX(fields)
	case "Z":
		return p.processTagZ(fields)
	case "BA02", "BA03", "BA10", "BA19", "BA20":
		return p.processReferenceTag(tag, fields)
	default:
		// For now, ignore unknown tags (they might be optional)
		return nil
	}
}

// processTagA processes the main NFe information (tag A)
func (p *Parser) processTagA(fields FieldMap) error {
	p.currentNFe.InfNFe = &nfe.InfNFe{
		Versao: fields["versao"],
		ID:     fields["Id"],
	}
	return nil
}

// processTagB processes identification information (tag B)
func (p *Parser) processTagB(fields FieldMap) error {
	ide := &nfe.Identificacao{}

	if cuf := fields["cUF"]; cuf != "" {
		ide.CUF = strings.TrimSpace(cuf)
	}

	if cnf := fields["cNF"]; cnf != "" {
		ide.CNF = cnf
	}

	if natop := fields["natOp"]; natop != "" {
		ide.NatOp = natop
	}

	if mod := fields["mod"]; mod != "" {
		ide.Mod = mod
	}

	if serie := fields["serie"]; serie != "" {
		ide.Serie = serie
	}

	if nnf := fields["nNF"]; nnf != "" {
		ide.NNF = nnf
	}

	if dhemi := fields["dhEmi"]; dhemi != "" {
		ide.DhEmi = dhemi
	}

	if tpnf := fields["tpNF"]; tpnf != "" {
		ide.TpNF = tpnf
	}

	if idDest := fields["idDest"]; idDest != "" {
		ide.IdDest = idDest
	}

	if cMunFG := fields["cMunFG"]; cMunFG != "" {
		ide.CMunFG = cMunFG
	}

	if tpImp := fields["tpImp"]; tpImp != "" {
		ide.TpImp = tpImp
	}

	if tpEmis := fields["tpEmis"]; tpEmis != "" {
		ide.TpEmis = tpEmis
	}

	if cdv := fields["cDV"]; cdv != "" {
		ide.CDV = cdv
	}

	if tpAmb := fields["tpAmb"]; tpAmb != "" {
		ide.TpAmb = tpAmb
	}

	if finNFe := fields["finNFe"]; finNFe != "" {
		ide.FinNFe = finNFe
	}

	if indFinal := fields["indFinal"]; indFinal != "" {
		ide.IndFinal = indFinal
	}

	if indPres := fields["indPres"]; indPres != "" {
		ide.IndPres = indPres
	}

	if procEmi := fields["procEmi"]; procEmi != "" {
		ide.ProcEmi = procEmi
	}

	if verProc := fields["verProc"]; verProc != "" {
		ide.VerProc = verProc
	}

	p.currentNFe.Identificacao = ide
	return nil
}

// processTagC processes issuer information (tag C)
func (p *Parser) processTagC(fields FieldMap) error {
	if p.currentNFe.Emitente == nil {
		p.currentNFe.Emitente = &nfe.Emitente{}
	}

	if xnome := fields["xNome"]; xnome != "" {
		p.currentNFe.Emitente.XNome = xnome
	}

	if xfant := fields["xFant"]; xfant != "" {
		p.currentNFe.Emitente.XFant = xfant
	}

	if ie := fields["IE"]; ie != "" {
		p.currentNFe.Emitente.IE = ie
	}

	if iest := fields["IEST"]; iest != "" {
		p.currentNFe.Emitente.IEST = iest
	}

	if im := fields["IM"]; im != "" {
		p.currentNFe.Emitente.IM = im
	}

	if cnae := fields["CNAE"]; cnae != "" {
		p.currentNFe.Emitente.CNAE = cnae
	}

	if crt := fields["CRT"]; crt != "" {
		p.currentNFe.Emitente.CRT = crt
	}

	return nil
}

// processTagC02 processes issuer CNPJ (tag C02)
func (p *Parser) processTagC02(fields FieldMap) error {
	if p.currentNFe.Emitente == nil {
		p.currentNFe.Emitente = &nfe.Emitente{}
	}

	if cnpj := fields["CNPJ"]; cnpj != "" {
		p.currentNFe.Emitente.CNPJ = cnpj
	}

	return nil
}

// processTagC02A processes issuer CPF (tag C02A)
func (p *Parser) processTagC02A(fields FieldMap) error {
	if p.currentNFe.Emitente == nil {
		p.currentNFe.Emitente = &nfe.Emitente{}
	}

	if cpf := fields["CPF"]; cpf != "" {
		p.currentNFe.Emitente.CPF = cpf
	}

	return nil
}

// processTagC05 processes issuer address (tag C05)
func (p *Parser) processTagC05(fields FieldMap) error {
	if p.currentNFe.Emitente == nil {
		p.currentNFe.Emitente = &nfe.Emitente{}
	}

	endereco := &nfe.Endereco{}

	if xlgr := fields["xLgr"]; xlgr != "" {
		endereco.XLgr = xlgr
	}

	if nro := fields["nro"]; nro != "" {
		endereco.Nro = nro
	}

	if xcpl := fields["xCpl"]; xcpl != "" {
		endereco.XCpl = xcpl
	}

	if xbairro := fields["xBairro"]; xbairro != "" {
		endereco.XBairro = xbairro
	}

	if cmun := fields["cMun"]; cmun != "" {
		endereco.CMun = cmun
	}

	if xmun := fields["xMun"]; xmun != "" {
		endereco.XMun = xmun
	}

	if uf := fields["UF"]; uf != "" {
		endereco.UF = uf
	}

	if cep := fields["CEP"]; cep != "" {
		endereco.CEP = cep
	}

	if cpais := fields["cPais"]; cpais != "" {
		endereco.CPais = cpais
	}

	if xpais := fields["xPais"]; xpais != "" {
		endereco.XPais = xpais
	}

	if fone := fields["fone"]; fone != "" {
		endereco.Fone = fone
	}

	p.currentNFe.Emitente.EnderEmit = *endereco
	return nil
}

// processTagE processes recipient information (tag E)
func (p *Parser) processTagE(fields FieldMap) error {
	if p.currentNFe.Destinatario == nil {
		p.currentNFe.Destinatario = &nfe.Destinatario{}
	}

	if xnome := fields["xNome"]; xnome != "" {
		p.currentNFe.Destinatario.XNome = xnome
	}

	if indIEDest := fields["indIEDest"]; indIEDest != "" {
		p.currentNFe.Destinatario.IndIEDest = indIEDest
	}

	if ie := fields["IE"]; ie != "" {
		p.currentNFe.Destinatario.IE = ie
	}

	if isuf := fields["ISUF"]; isuf != "" {
		p.currentNFe.Destinatario.ISUF = isuf
	}

	if im := fields["IM"]; im != "" {
		p.currentNFe.Destinatario.IM = im
	}

	if email := fields["email"]; email != "" {
		p.currentNFe.Destinatario.Email = email
	}

	return nil
}

// processTagE02 processes recipient CNPJ (tag E02)
func (p *Parser) processTagE02(fields FieldMap) error {
	if p.currentNFe.Destinatario == nil {
		p.currentNFe.Destinatario = &nfe.Destinatario{}
	}

	if cnpj := fields["CNPJ"]; cnpj != "" {
		p.currentNFe.Destinatario.CNPJ = cnpj
	}

	return nil
}

// processTagE03 processes recipient CPF (tag E03)
func (p *Parser) processTagE03(fields FieldMap) error {
	if p.currentNFe.Destinatario == nil {
		p.currentNFe.Destinatario = &nfe.Destinatario{}
	}

	if cpf := fields["CPF"]; cpf != "" {
		p.currentNFe.Destinatario.CPF = cpf
	}

	return nil
}

// processTagE05 processes recipient address (tag E05)
func (p *Parser) processTagE05(fields FieldMap) error {
	if p.currentNFe.Destinatario == nil {
		p.currentNFe.Destinatario = &nfe.Destinatario{}
	}

	endereco := &nfe.Endereco{}

	if xlgr := fields["xLgr"]; xlgr != "" {
		endereco.XLgr = xlgr
	}

	if nro := fields["nro"]; nro != "" {
		endereco.Nro = nro
	}

	if xcpl := fields["xCpl"]; xcpl != "" {
		endereco.XCpl = xcpl
	}

	if xbairro := fields["xBairro"]; xbairro != "" {
		endereco.XBairro = xbairro
	}

	if cmun := fields["cMun"]; cmun != "" {
		endereco.CMun = cmun
	}

	if xmun := fields["xMun"]; xmun != "" {
		endereco.XMun = xmun
	}

	if uf := fields["UF"]; uf != "" {
		endereco.UF = uf
	}

	if cep := fields["CEP"]; cep != "" {
		endereco.CEP = cep
	}

	if cpais := fields["cPais"]; cpais != "" {
		endereco.CPais = cpais
	}

	if xpais := fields["xPais"]; xpais != "" {
		endereco.XPais = xpais
	}

	if fone := fields["fone"]; fone != "" {
		endereco.Fone = fone
	}

	p.currentNFe.Destinatario.EnderDest = endereco
	return nil
}

// processTagH processes item header (tag H) - increments item counter
func (p *Parser) processTagH(fields FieldMap) error {
	if nitem := fields["nItem"]; nitem != "" {
		if num, err := strconv.Atoi(nitem); err == nil {
			p.currentItem = num
		}
	}
	return nil
}

// processTagI processes item/product information (tag I)
func (p *Parser) processTagI(fields FieldMap) error {
	item := &nfe.Item{
		NItem: strconv.Itoa(p.currentItem),
		Prod:  nfe.Produto{},
	}

	if cprod := fields["cProd"]; cprod != "" {
		item.Prod.CProd = cprod
	}

	if cean := fields["cEAN"]; cean != "" {
		item.Prod.CEAN = cean
	}

	if xprod := fields["xProd"]; xprod != "" {
		item.Prod.XProd = xprod
	}

	if ncm := fields["NCM"]; ncm != "" {
		item.Prod.NCM = ncm
	}

	if cfop := fields["CFOP"]; cfop != "" {
		item.Prod.CFOP = cfop
	}

	if ucom := fields["uCom"]; ucom != "" {
		item.Prod.UCom = ucom
	}

	if qcom := fields["qCom"]; qcom != "" {
		item.Prod.QCom = qcom
	}

	if vuncom := fields["vUnCom"]; vuncom != "" {
		item.Prod.VUnCom = vuncom
	}

	if vprod := fields["vProd"]; vprod != "" {
		item.Prod.VProd = vprod
	}

	if ceantrib := fields["cEANTrib"]; ceantrib != "" {
		item.Prod.CEANTrib = ceantrib
	}

	if utrib := fields["uTrib"]; utrib != "" {
		item.Prod.UTrib = utrib
	}

	if qtrib := fields["qTrib"]; qtrib != "" {
		item.Prod.QTrib = qtrib
	}

	if vuntrib := fields["vUnTrib"]; vuntrib != "" {
		item.Prod.VUnTrib = vuntrib
	}

	p.currentNFe.Itens = append(p.currentNFe.Itens, item)
	p.currentItem++

	return nil
}

// processTagM processes tax information (tag M)
func (p *Parser) processTagM(fields FieldMap) error {
	// Tax information processing would go here
	// For now, just return nil to indicate successful parsing
	return nil
}

// processTagN processes ICMS information (tag N)
func (p *Parser) processTagN(fields FieldMap) error {
	// ICMS information processing would go here
	// This would update the current item's tax information
	return nil
}

// processTagW processes totals (tag W)
func (p *Parser) processTagW(fields FieldMap) error {
	total := &nfe.Total{
		ICMSTot: nfe.ICMSTotal{},
	}

	// Process total fields - these would need to be extracted from the W structure
	// For now, create an empty total structure
	p.currentNFe.Total = total
	return nil
}

// processTagX processes transport information (tag X)
func (p *Parser) processTagX(fields FieldMap) error {
	if p.currentNFe.Transporte == nil {
		p.currentNFe.Transporte = &nfe.Transporte{}
	}

	if modFrete := fields["modFrete"]; modFrete != "" {
		p.currentNFe.Transporte.ModFrete = modFrete
	}

	return nil
}

// processTagZ processes additional information (tag Z)
func (p *Parser) processTagZ(fields FieldMap) error {
	if p.currentNFe.InfAdic == nil {
		p.currentNFe.InfAdic = &nfe.InfAdicionais{}
	}

	if infCpl := fields["infCpl"]; infCpl != "" {
		p.currentNFe.InfAdic.InfCpl = infCpl
	}

	return nil
}

// processReferenceTag processes reference tags (BA series)
func (p *Parser) processReferenceTag(tag string, fields FieldMap) error {
	ref := Referencia{}

	switch tag {
	case "BA02":
		if refnfe := fields["refNFe"]; refnfe != "" {
			ref.RefNFe = refnfe
		}
	case "BA03":
		// Process NFe model 1/1A reference
		refNF := &RefNF{}
		if cuf := fields["cUF"]; cuf != "" {
			refNF.CUF = cuf
		}
		if aamm := fields["AAMM"]; aamm != "" {
			refNF.AAMM = aamm
		}
		if cnpj := fields["CNPJ"]; cnpj != "" {
			refNF.CNPJ = cnpj
		}
	}

	p.currentNFe.Referencias = append(p.currentNFe.Referencias, ref)
	return nil
}

// validateRequiredFields validates that all required fields are present
func (p *Parser) validateRequiredFields() error {
	if p.currentNFe.InfNFe == nil {
		return fmt.Errorf("missing required tag A (infNFe)")
	}

	if p.currentNFe.Identificacao == nil {
		return fmt.Errorf("missing required tag B (identification)")
	}

	if p.currentNFe.Emitente == nil {
		return fmt.Errorf("missing required tag C (issuer)")
	}

	if len(p.currentNFe.Itens) == 0 {
		return fmt.Errorf("missing required tag I (items)")
	}

	return nil
}

// GetFieldValue safely gets a field value with default
func GetFieldValue(fields FieldMap, key, defaultValue string) string {
	if value, exists := fields[key]; exists && value != "" {
		return value
	}
	return defaultValue
}

// ParseFloat safely parses a float value from field
func ParseFloat(fields FieldMap, key string) (float64, error) {
	value := GetFieldValue(fields, key, "0")
	return strconv.ParseFloat(value, 64)
}

// ParseInt safely parses an int value from field
func ParseInt(fields FieldMap, key string) (int, error) {
	value := GetFieldValue(fields, key, "0")
	return strconv.Atoi(value)
}

// ParseDateTime safely parses a datetime value from field
func ParseDateTime(fields FieldMap, key string) (time.Time, error) {
	value := GetFieldValue(fields, key, "")
	if value == "" {
		return time.Time{}, fmt.Errorf("empty datetime field")
	}

	// Try different datetime formats
	formats := []string{
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime format: %s", value)
}
