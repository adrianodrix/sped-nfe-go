// Package factories provides utility factories for NFe processing including QR Code generation.
package factories

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// QRCode handles QR code generation for NFCe documents
type QRCode struct {
	Version string // QR Code version (1.00, 2.00, 3.00)
	CSC     string // Security Code (Código de Segurança do Contribuinte)
	CSCId   string // CSC identifier
}

// QRCodeConfig holds configuration for QR code generation
type QRCodeConfig struct {
	Version string
	CSC     string
	CSCId   string
}

// NewQRCode creates a new QR code generator
func NewQRCode(config QRCodeConfig) *QRCode {
	if config.Version == "" {
		config.Version = "2.00" // Default to version 2.00
	}
	
	return &QRCode{
		Version: config.Version,
		CSC:     config.CSC,
		CSCId:   config.CSCId,
	}
}

// PutQRTag inserts QR code tags into NFCe XML
func (q *QRCode) PutQRTag(xml []byte, token, idToken, versao, urlqr, urichave string) ([]byte, error) {
	xmlStr := string(xml)
	
	// Extract NFCe information from XML
	chNFe, err := q.extractChaveNFe(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract chave NFe: %v", err)
	}
	
	tpAmb, err := q.extractTpAmb(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tpAmb: %v", err)
	}
	
	dhEmi, err := q.extractDhEmi(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract dhEmi: %v", err)
	}
	
	vNF, err := q.extractVNF(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract vNF: %v", err)
	}
	
	vICMS, err := q.extractVICMS(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract vICMS: %v", err)
	}
	
	digVal, err := q.extractDigVal(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract digVal: %v", err)
	}
	
	cDest, _ := q.extractCDest(xmlStr) // Optional field
	
	tpEmis, err := q.extractTpEmis(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tpEmis: %v", err)
	}
	
	// Generate QR code based on version
	var qrCodeData string
	switch q.Version {
	case "1.00":
		qrCodeData = q.generate100(chNFe, urlqr, tpAmb, dhEmi, vNF, digVal)
	case "2.00":
		qrCodeData = q.generate200(chNFe, urlqr, tpAmb, dhEmi, vNF, vICMS, digVal, token, idToken, versao, tpEmis, cDest)
	case "3.00":
		assinatura, err := q.extractAssinatura(xmlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to extract assinatura for v3.00: %v", err)
		}
		
		idDest, err := q.extractIdDest(xmlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to extract idDest: %v", err)
		}
		
		qrCodeData = q.generate300(chNFe, urlqr, tpAmb, dhEmi, vNF, tpEmis, idDest, cDest, assinatura)
	default:
		return nil, fmt.Errorf("unsupported QR code version: %s", q.Version)
	}
	
	// Insert QR code into XML
	xmlWithQR, err := q.insertQRCodeIntoXML(xmlStr, qrCodeData, urichave)
	if err != nil {
		return nil, fmt.Errorf("failed to insert QR code into XML: %v", err)
	}
	
	return []byte(xmlWithQR), nil
}

// generate100 generates QR code for version 1.00
func (q *QRCode) generate100(chNFe, url, tpAmb, dhEmi, vNF, digVal string) string {
	return fmt.Sprintf("%s?chNFe=%s&nVersao=100&tpAmb=%s&dhEmi=%s&vNF=%s&digVal=%s",
		url, chNFe, tpAmb, q.str2Hex(dhEmi), vNF, q.str2Hex(digVal))
}

// generate200 generates QR code for version 2.00 (with CSC)
func (q *QRCode) generate200(chNFe, url, tpAmb, dhEmi, vNF, vICMS, digVal, token, idToken, versao string, tpEmis int, cDest string) string {
	// Build base parameters
	params := fmt.Sprintf("chNFe=%s&nVersao=200&tpAmb=%s&cDest=%s&dhEmi=%s&vNF=%s&vICMS=%s&digVal=%s&cIdToken=%s",
		chNFe, tpAmb, cDest, q.str2Hex(dhEmi), vNF, vICMS, q.str2Hex(digVal), idToken)
	
	// Add CSC and generate hash
	dataToHash := params + token
	hash := q.generateSHA1Hash(dataToHash)
	
	// Build final URL
	return fmt.Sprintf("%s?%s&cHashQRCode=%s", url, params, hash)
}

// generate300 generates QR code for version 3.00 (with signature)
func (q *QRCode) generate300(chNFe, url, tpAmb, dhEmi, vNF string, tpEmis, idDest int, cDest, assinatura string) string {
	return fmt.Sprintf("%s?chNFe=%s&nVersao=300&tpAmb=%s&idDest=%d&cDest=%s&dhEmi=%s&vNF=%s&tpEmis=%d&assinatura=%s",
		url, chNFe, tpAmb, idDest, cDest, q.str2Hex(dhEmi), vNF, tpEmis, assinatura)
}

// str2Hex converts string to hexadecimal representation
func (q *QRCode) str2Hex(str string) string {
	if str == "" {
		return ""
	}
	return hex.EncodeToString([]byte(str))
}

// generateSHA1Hash generates SHA1 hash for the given data
func (q *QRCode) generateSHA1Hash(data string) string {
	hash := sha1.Sum([]byte(data))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// insertQRCodeIntoXML inserts the QR code data into NFCe XML
func (q *QRCode) insertQRCodeIntoXML(xml, qrCodeData, urichave string) (string, error) {
	// Check if infNFeSupl already exists
	if strings.Contains(xml, "<infNFeSupl>") {
		// Replace existing QR code
		re := regexp.MustCompile(`<qrCode>.*?</qrCode>`)
		xml = re.ReplaceAllString(xml, fmt.Sprintf("<qrCode><![CDATA[%s]]></qrCode>", qrCodeData))
		
		if urichave != "" {
			re = regexp.MustCompile(`<urlChave>.*?</urlChave>`)
			xml = re.ReplaceAllString(xml, fmt.Sprintf("<urlChave>%s</urlChave>", urichave))
		}
	} else {
		// Insert new infNFeSupl section
		infNFeSupl := q.buildInfNFeSupl(qrCodeData, urichave)
		
		// Find insertion point (before </infNFe>)
		insertPoint := strings.LastIndex(xml, "</infNFe>")
		if insertPoint == -1 {
			return "", fmt.Errorf("could not find </infNFe> tag in XML")
		}
		
		// Insert infNFeSupl before </infNFe>
		xml = xml[:insertPoint] + infNFeSupl + xml[insertPoint:]
	}
	
	return xml, nil
}

// buildInfNFeSupl builds the infNFeSupl XML section
func (q *QRCode) buildInfNFeSupl(qrCodeData, urichave string) string {
	infNFeSupl := "<infNFeSupl>"
	infNFeSupl += fmt.Sprintf("<qrCode><![CDATA[%s]]></qrCode>", qrCodeData)
	
	if urichave != "" {
		infNFeSupl += fmt.Sprintf("<urlChave>%s</urlChave>", urichave)
	}
	
	infNFeSupl += "</infNFeSupl>"
	return infNFeSupl
}

// Extraction methods for XML parsing

func (q *QRCode) extractChaveNFe(xml string) (string, error) {
	re := regexp.MustCompile(`Id="NFe([0-9]{44})"`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "", fmt.Errorf("chave NFe not found in XML")
	}
	return matches[1], nil
}

func (q *QRCode) extractTpAmb(xml string) (string, error) {
	re := regexp.MustCompile(`<tpAmb>([12])</tpAmb>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "", fmt.Errorf("tpAmb not found in XML")
	}
	return matches[1], nil
}

func (q *QRCode) extractDhEmi(xml string) (string, error) {
	re := regexp.MustCompile(`<dhEmi>([^<]+)</dhEmi>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "", fmt.Errorf("dhEmi not found in XML")
	}
	return matches[1], nil
}

func (q *QRCode) extractVNF(xml string) (string, error) {
	re := regexp.MustCompile(`<vNF>([0-9.]+)</vNF>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "", fmt.Errorf("vNF not found in XML")
	}
	return matches[1], nil
}

func (q *QRCode) extractVICMS(xml string) (string, error) {
	re := regexp.MustCompile(`<vICMS>([0-9.]+)</vICMS>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "0.00", nil // Default value if not found
	}
	return matches[1], nil
}

func (q *QRCode) extractDigVal(xml string) (string, error) {
	re := regexp.MustCompile(`<digVal>([^<]+)</digVal>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "", fmt.Errorf("digVal not found in XML")
	}
	return matches[1], nil
}

func (q *QRCode) extractCDest(xml string) (string, error) {
	// Try CNPJ first
	re := regexp.MustCompile(`<dest>.*?<CNPJ>([0-9]{14})</CNPJ>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) >= 2 {
		return matches[1], nil
	}
	
	// Try CPF
	re = regexp.MustCompile(`<dest>.*?<CPF>([0-9]{11})</CPF>`)
	matches = re.FindStringSubmatch(xml)
	if len(matches) >= 2 {
		return matches[1], nil
	}
	
	// Return empty if no destination found (valid for NFCe)
	return "", nil
}

func (q *QRCode) extractTpEmis(xml string) (int, error) {
	re := regexp.MustCompile(`<tpEmis>([0-9])</tpEmis>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return 1, nil // Default to normal emission
	}
	
	tpEmis, err := strconv.Atoi(matches[1])
	if err != nil {
		return 1, nil
	}
	
	return tpEmis, nil
}

func (q *QRCode) extractIdDest(xml string) (int, error) {
	re := regexp.MustCompile(`<idDest>([0-9])</idDest>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return 1, nil // Default to internal operation
	}
	
	idDest, err := strconv.Atoi(matches[1])
	if err != nil {
		return 1, nil
	}
	
	return idDest, nil
}

func (q *QRCode) extractAssinatura(xml string) (string, error) {
	// Extract signature from XML digital signature
	re := regexp.MustCompile(`<SignatureValue>([^<]+)</SignatureValue>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "", fmt.Errorf("signature not found in XML")
	}
	return matches[1], nil
}

// GetQRCodeFromXML extracts existing QR code from XML
func (q *QRCode) GetQRCodeFromXML(xml string) (string, error) {
	re := regexp.MustCompile(`<qrCode><!\[CDATA\[([^\]]+)\]\]></qrCode>`)
	matches := re.FindStringSubmatch(xml)
	if len(matches) < 2 {
		return "", fmt.Errorf("QR code not found in XML")
	}
	return matches[1], nil
}

// ValidateQRCode validates a QR code URL format
func (q *QRCode) ValidateQRCode(qrCode string) error {
	if qrCode == "" {
		return fmt.Errorf("QR code cannot be empty")
	}
	
	// Check if it's a valid URL
	if !strings.HasPrefix(qrCode, "http://") && !strings.HasPrefix(qrCode, "https://") {
		return fmt.Errorf("QR code must be a valid URL")
	}
	
	// Check for required parameters based on version
	if strings.Contains(qrCode, "nVersao=100") {
		requiredParams := []string{"chNFe=", "tpAmb=", "dhEmi=", "vNF=", "digVal="}
		for _, param := range requiredParams {
			if !strings.Contains(qrCode, param) {
				return fmt.Errorf("QR code v1.00 missing required parameter: %s", param)
			}
		}
	} else if strings.Contains(qrCode, "nVersao=200") {
		requiredParams := []string{"chNFe=", "tpAmb=", "dhEmi=", "vNF=", "vICMS=", "digVal=", "cIdToken=", "cHashQRCode="}
		for _, param := range requiredParams {
			if !strings.Contains(qrCode, param) {
				return fmt.Errorf("QR code v2.00 missing required parameter: %s", param)
			}
		}
	} else if strings.Contains(qrCode, "nVersao=300") {
		requiredParams := []string{"chNFe=", "tpAmb=", "dhEmi=", "vNF=", "tpEmis=", "assinatura="}
		for _, param := range requiredParams {
			if !strings.Contains(qrCode, param) {
				return fmt.Errorf("QR code v3.00 missing required parameter: %s", param)
			}
		}
	}
	
	return nil
}

// GetStateConsultationURL returns the consultation URL for a given state
func GetStateConsultationURL(uf string, environment int) string {
	// URLs from storage/uri_consulta_nfce.json
	consultationURLs := map[string]map[int]string{
		"SP": {
			1: "https://www.fazenda.sp.gov.br/nfce/qrcode",
			2: "https://www.homologacao.fazenda.sp.gov.br/nfce/qrcode",
		},
		"RJ": {
			1: "http://www4.fazenda.rj.gov.br/consultaDFe/paginas/consultaQRCode.faces",
			2: "http://www4.fazenda.rj.gov.br/consultaDFe/paginas/consultaQRCode.faces",
		},
		"MG": {
			1: "https://portalsped.fazenda.mg.gov.br/portalnfce/sistema/qrcode.xhtml",
			2: "https://portalsped.fazenda.mg.gov.br/portalnfce/sistema/qrcode.xhtml",
		},
		"RS": {
			1: "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx",
			2: "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx",
		},
		// Add more states as needed
	}
	
	if stateURLs, exists := consultationURLs[strings.ToUpper(uf)]; exists {
		if url, exists := stateURLs[environment]; exists {
			return url
		}
	}
	
	// Default to SVRS if state not found
	if environment == 1 {
		return "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx"
	}
	return "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx"
}

// QRCodeBuilder provides a fluent interface for QR code generation
type QRCodeBuilder struct {
	qrcode *QRCode
	chNFe  string
	url    string
	tpAmb  string
	dhEmi  string
	vNF    string
	vICMS  string
	digVal string
	cDest  string
	token  string
	idToken string
	tpEmis int
	idDest int
	assinatura string
}

// NewQRCodeBuilder creates a new QR code builder
func NewQRCodeBuilder(qrcode *QRCode) *QRCodeBuilder {
	return &QRCodeBuilder{
		qrcode: qrcode,
		tpEmis: 1,
		idDest: 1,
	}
}

// ChaveNFe sets the NFe access key
func (b *QRCodeBuilder) ChaveNFe(chave string) *QRCodeBuilder {
	b.chNFe = chave
	return b
}

// URL sets the base consultation URL
func (b *QRCodeBuilder) URL(url string) *QRCodeBuilder {
	b.url = url
	return b
}

// Environment sets the environment
func (b *QRCodeBuilder) Environment(tpAmb string) *QRCodeBuilder {
	b.tpAmb = tpAmb
	return b
}

// EmissionDateTime sets the emission date/time
func (b *QRCodeBuilder) EmissionDateTime(dhEmi string) *QRCodeBuilder {
	b.dhEmi = dhEmi
	return b
}

// TotalValue sets the total NFe value
func (b *QRCodeBuilder) TotalValue(vNF string) *QRCodeBuilder {
	b.vNF = vNF
	return b
}

// ICMSValue sets the ICMS value
func (b *QRCodeBuilder) ICMSValue(vICMS string) *QRCodeBuilder {
	b.vICMS = vICMS
	return b
}

// DigestValue sets the digest value
func (b *QRCodeBuilder) DigestValue(digVal string) *QRCodeBuilder {
	b.digVal = digVal
	return b
}

// DestinationDocument sets the destination document
func (b *QRCodeBuilder) DestinationDocument(cDest string) *QRCodeBuilder {
	b.cDest = cDest
	return b
}

// Token sets the CSC token
func (b *QRCodeBuilder) Token(token, idToken string) *QRCodeBuilder {
	b.token = token
	b.idToken = idToken
	return b
}

// EmissionType sets the emission type
func (b *QRCodeBuilder) EmissionType(tpEmis int) *QRCodeBuilder {
	b.tpEmis = tpEmis
	return b
}

// DestinationType sets the destination type
func (b *QRCodeBuilder) DestinationType(idDest int) *QRCodeBuilder {
	b.idDest = idDest
	return b
}

// Signature sets the XML signature
func (b *QRCodeBuilder) Signature(assinatura string) *QRCodeBuilder {
	b.assinatura = assinatura
	return b
}

// Build generates the QR code string
func (b *QRCodeBuilder) Build() (string, error) {
	if b.chNFe == "" || b.url == "" {
		return "", fmt.Errorf("chNFe and url are required")
	}
	
	switch b.qrcode.Version {
	case "1.00":
		return b.qrcode.generate100(b.chNFe, b.url, b.tpAmb, b.dhEmi, b.vNF, b.digVal), nil
	case "2.00":
		return b.qrcode.generate200(b.chNFe, b.url, b.tpAmb, b.dhEmi, b.vNF, b.vICMS, b.digVal, b.token, b.idToken, "4.00", b.tpEmis, b.cDest), nil
	case "3.00":
		return b.qrcode.generate300(b.chNFe, b.url, b.tpAmb, b.dhEmi, b.vNF, b.tpEmis, b.idDest, b.cDest, b.assinatura), nil
	default:
		return "", fmt.Errorf("unsupported QR code version: %s", b.qrcode.Version)
	}
}