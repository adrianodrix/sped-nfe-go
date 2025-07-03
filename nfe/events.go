// Package nfe provides event structures and functions for NFe events like cancellation and correction letters
package nfe

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Event types constants matching the PHP implementation
const (
	EVT_CONFIRMACAO                      = 210200 // Confirmação da Operação
	EVT_CIENCIA                          = 210210 // Ciência da Operação
	EVT_DESCONHECIMENTO                  = 210220 // Desconhecimento da Operação
	EVT_NAO_REALIZADA                    = 210240 // Operação não Realizada
	EVT_CCE                              = 110110 // Carta de Correção
	EVT_CANCELA                          = 110111 // Cancelamento
	EVT_CANCELASUBSTITUICAO              = 110112 // Cancelamento por Substituição
	EVT_EPEC                             = 110140 // Emissão em Contingência EPEC
	EVT_ATORINTERESSADO                  = 110150 // Ator Interessado
	EVT_COMPROVANTE_ENTREGA              = 110130 // Comprovante de Entrega
	EVT_CANCELAMENTO_COMPROVANTE_ENTREGA = 110131 // Cancelamento Comprovante de Entrega
	EVT_PRORROGACAO_1                    = 111500 // Prorrogação 1
	EVT_PRORROGACAO_2                    = 111501 // Prorrogação 2
	EVT_CANCELA_PRORROGACAO_1            = 111502 // Cancelamento Prorrogação 1
	EVT_CANCELA_PRORROGACAO_2            = 111503 // Cancelamento Prorrogação 2
	EVT_INSUCESSO_ENTREGA                = 110192 // Insucesso na Entrega
	EVT_CANCELA_INSUCESSO_ENTREGA        = 110193 // Cancelamento Insucesso na Entrega
	EVT_CONCILIACAO                      = 110750 // Conciliação Financeira
	EVT_CANCELA_CONCILIACAO              = 110751 // Cancelamento Conciliação Financeira
)

// EventInfoNFe contains information about each event type
type EventInfoNFe struct {
	Version string
	Name    string
}

// GetEventInfo returns the event information for a given event type
func GetEventInfo(eventType int) (EventInfoNFe, error) {
	events := map[int]EventInfoNFe{
		EVT_CCE:                              {Version: "1.00", Name: "envCCe"},
		EVT_CANCELA:                          {Version: "1.00", Name: "envEventoCancNFe"},
		EVT_CANCELASUBSTITUICAO:              {Version: "1.00", Name: "envEventoCancSubst"},
		EVT_ATORINTERESSADO:                  {Version: "1.00", Name: "envEventoAtorInteressado"},
		EVT_COMPROVANTE_ENTREGA:              {Version: "1.00", Name: "envEventoEntregaNFe"},
		EVT_CANCELAMENTO_COMPROVANTE_ENTREGA: {Version: "1.00", Name: "envEventoCancEntregaNFe"},
		EVT_CIENCIA:                          {Version: "1.00", Name: "envConfRecebto"},
		EVT_CONFIRMACAO:                      {Version: "1.00", Name: "envConfRecebto"},
		EVT_DESCONHECIMENTO:                  {Version: "1.00", Name: "envConfRecebto"},
		EVT_NAO_REALIZADA:                    {Version: "1.00", Name: "envConfRecebto"},
		EVT_PRORROGACAO_1:                    {Version: "1.00", Name: "envRemIndus"},
		EVT_PRORROGACAO_2:                    {Version: "1.00", Name: "envRemIndus"},
		EVT_CANCELA_PRORROGACAO_1:            {Version: "1.00", Name: "envRemIndus"},
		EVT_CANCELA_PRORROGACAO_2:            {Version: "1.00", Name: "envRemIndus"},
		EVT_EPEC:                             {Version: "1.00", Name: "envEPEC"},
		EVT_INSUCESSO_ENTREGA:                {Version: "1.00", Name: "envEventoInsucessoNFe"},
		EVT_CANCELA_INSUCESSO_ENTREGA:        {Version: "1.00", Name: "envEventoCancInsucessoNFe"},
		EVT_CONCILIACAO:                      {Version: "1.00", Name: "envEventoEConf"},
		EVT_CANCELA_CONCILIACAO:              {Version: "1.00", Name: "envEventoCancEConf"},
	}

	if info, exists := events[eventType]; exists {
		return info, nil
	}
	return EventInfoNFe{}, fmt.Errorf("event type %d not found", eventType)
}

// EventRequestNFe represents the structure for event requests
type EventRequestNFe struct {
	XMLName xml.Name  `xml:"envEvento"`
	Xmlns   string    `xml:"xmlns,attr"`
	Versao  string    `xml:"versao,attr"`
	IdLote  string    `xml:"idLote"`
	Evento  EventoNFe `xml:"evento"`
}

// EventoNFe represents the event structure
type EventoNFe struct {
	XMLName   xml.Name      `xml:"evento"`
	Xmlns     string        `xml:"xmlns,attr"`
	Versao    string        `xml:"versao,attr"`
	InfEvento InfEventoNFe  `xml:"infEvento"`
	Signature *SignatureNFe `xml:"Signature,omitempty"`
}

// InfEventoNFe represents the event information structure
type InfEventoNFe struct {
	XMLName    xml.Name     `xml:"infEvento"`
	ID         string       `xml:"Id,attr"`
	COrgao     string       `xml:"cOrgao"`
	TpAmb      string       `xml:"tpAmb"`
	CNPJ       string       `xml:"CNPJ"`
	ChNFe      string       `xml:"chNFe"`
	DhEvento   string       `xml:"dhEvento"`
	TpEvento   string       `xml:"tpEvento"`
	NSeqEvento string       `xml:"nSeqEvento"`
	VerEvento  string       `xml:"verEvento"`
	DetEvento  DetEventoNFe `xml:"detEvento"`
}

// DetEventoNFe represents the event details structure
type DetEventoNFe struct {
	XMLName xml.Name `xml:"detEvento"`
	Versao  string   `xml:"versao,attr"`
	// CCe fields
	XCorrecao string `xml:"xCorrecao,omitempty"`
	XCondUso  string `xml:"xCondUso,omitempty"`
	// Cancellation fields
	NProt string `xml:"nProt,omitempty"`
	XJust string `xml:"xJust,omitempty"`
	// Substitution fields
	ChNFeRef string `xml:"chNFeRef,omitempty"`
	VerAplic string `xml:"verAplic,omitempty"`
}

// SignatureNFe represents the XML signature structure
type SignatureNFe struct {
	XMLName xml.Name `xml:"Signature"`
	Xmlns   string   `xml:"xmlns,attr"`
	// Signature content will be added by the signing process
}

// EventResponseNFe represents the response structure from SEFAZ
type EventResponseNFe struct {
	XMLName   xml.Name       `xml:"retEnvEvento"`
	Xmlns     string         `xml:"xmlns,attr"`
	Versao    string         `xml:"versao,attr"`
	IdLote    string         `xml:"idLote"`
	TpAmb     string         `xml:"tpAmb"`
	COrgao    string         `xml:"cOrgao"`
	CStat     string         `xml:"cStat"`
	XMotivo   string         `xml:"xMotivo"`
	RetEvento []RetEventoNFe `xml:"retEvento"`
}

// RetEventoNFe represents individual event response
type RetEventoNFe struct {
	XMLName   xml.Name        `xml:"retEvento"`
	Versao    string          `xml:"versao,attr"`
	InfEvento InfEventoRetNFe `xml:"infEvento"`
}

// InfEventoRetNFe represents the event response information
type InfEventoRetNFe struct {
	XMLName     xml.Name `xml:"infEvento"`
	ID          string   `xml:"Id,attr"`
	TpAmb       string   `xml:"tpAmb"`
	VerAplic    string   `xml:"verAplic"`
	COrgao      string   `xml:"cOrgao"`
	CStat       string   `xml:"cStat"`
	XMotivo     string   `xml:"xMotivo"`
	ChNFe       string   `xml:"chNFe"`
	TpEvento    string   `xml:"tpEvento"`
	XEvento     string   `xml:"xEvento"`
	NSeqEvento  string   `xml:"nSeqEvento"`
	CNPJDest    string   `xml:"CNPJDest,omitempty"`
	EmailDest   string   `xml:"emailDest,omitempty"`
	DhRegEvento string   `xml:"dhRegEvento"`
	NProt       string   `xml:"nProt"`
}

// EventParams represents the parameters for creating an event
type EventParams struct {
	UF         string
	ChNFe      string
	TpEvento   int
	NSeqEvento int
	TagAdic    string
	DhEvento   *time.Time
	Lote       string
	CNPJ       string
	TpAmb      string
	VerEvento  string
}

// CreateEventXML creates the XML structure for an event
func CreateEventXML(params EventParams) (*EventRequestNFe, error) {
	// Validate required fields
	if params.ChNFe == "" {
		return nil, fmt.Errorf("chNFe is required")
	}
	if params.CNPJ == "" {
		return nil, fmt.Errorf("CNPJ is required")
	}
	if params.TpEvento == 0 {
		return nil, fmt.Errorf("tpEvento is required")
	}

	// Get event info
	eventInfo, err := GetEventInfo(params.TpEvento)
	if err != nil {
		return nil, err
	}

	// Set default values
	if params.Lote == "" {
		params.Lote = strconv.FormatInt(time.Now().Unix(), 10)
	}
	if params.NSeqEvento == 0 {
		params.NSeqEvento = 1
	}
	if params.VerEvento == "" {
		params.VerEvento = eventInfo.Version
	}

	// Set event date
	dhEvento := time.Now()
	if params.DhEvento != nil {
		dhEvento = *params.DhEvento
	}

	// Format sequence number with leading zeros
	sSeqEvento := fmt.Sprintf("%02d", params.NSeqEvento)
	eventID := fmt.Sprintf("ID%d%s%s", params.TpEvento, params.ChNFe, sSeqEvento)

	// Get UF code - temporarily using a simple map until utils package is ready
	ufCodes := map[string]int{
		"AC": 12, "AL": 17, "AP": 16, "AM": 23, "BA": 29, "CE": 23, "DF": 53,
		"ES": 32, "GO": 52, "MA": 21, "MT": 51, "MS": 50, "MG": 31, "PA": 15,
		"PB": 25, "PR": 41, "PE": 26, "PI": 22, "RJ": 33, "RN": 20, "RS": 43,
		"RO": 11, "RR": 14, "SC": 42, "SP": 35, "SE": 28, "TO": 17,
	}
	ufCode, exists := ufCodes[params.UF]
	if !exists {
		return nil, fmt.Errorf("invalid UF: %s", params.UF)
	}

	// Create event structure
	event := &EventRequestNFe{
		Xmlns:  "http://www.portalfiscal.inf.br/nfe",
		Versao: eventInfo.Version,
		IdLote: params.Lote,
		Evento: EventoNFe{
			Xmlns:  "http://www.portalfiscal.inf.br/nfe",
			Versao: eventInfo.Version,
			InfEvento: InfEventoNFe{
				ID:         eventID,
				COrgao:     strconv.Itoa(ufCode),
				TpAmb:      params.TpAmb,
				CNPJ:       params.CNPJ,
				ChNFe:      params.ChNFe,
				DhEvento:   dhEvento.Format("2006-01-02T15:04:05-07:00"),
				TpEvento:   strconv.Itoa(params.TpEvento),
				NSeqEvento: sSeqEvento,
				VerEvento:  params.VerEvento,
				DetEvento: DetEventoNFe{
					Versao: eventInfo.Version,
				},
			},
		},
	}

	// Parse additional tags if provided
	if params.TagAdic != "" {
		if err := parseAdditionalTags(&event.Evento.InfEvento.DetEvento, params.TagAdic); err != nil {
			return nil, fmt.Errorf("failed to parse additional tags: %v", err)
		}
	}

	return event, nil
}

// parseAdditionalTags parses the additional XML tags and populates the DetEventoNFe structure
func parseAdditionalTags(detEvento *DetEventoNFe, tagAdic string) error {
	// Simple XML parsing for specific tags
	if strings.Contains(tagAdic, "<xCorrecao>") {
		start := strings.Index(tagAdic, "<xCorrecao>") + len("<xCorrecao>")
		end := strings.Index(tagAdic, "</xCorrecao>")
		if start > 0 && end > start {
			detEvento.XCorrecao = tagAdic[start:end]
		}
	}

	if strings.Contains(tagAdic, "<xCondUso>") {
		start := strings.Index(tagAdic, "<xCondUso>") + len("<xCondUso>")
		end := strings.Index(tagAdic, "</xCondUso>")
		if start > 0 && end > start {
			detEvento.XCondUso = tagAdic[start:end]
		}
	}

	if strings.Contains(tagAdic, "<nProt>") {
		start := strings.Index(tagAdic, "<nProt>") + len("<nProt>")
		end := strings.Index(tagAdic, "</nProt>")
		if start > 0 && end > start {
			detEvento.NProt = tagAdic[start:end]
		}
	}

	if strings.Contains(tagAdic, "<xJust>") {
		start := strings.Index(tagAdic, "<xJust>") + len("<xJust>")
		end := strings.Index(tagAdic, "</xJust>")
		if start > 0 && end > start {
			detEvento.XJust = tagAdic[start:end]
		}
	}

	if strings.Contains(tagAdic, "<chNFeRef>") {
		start := strings.Index(tagAdic, "<chNFeRef>") + len("<chNFeRef>")
		end := strings.Index(tagAdic, "</chNFeRef>")
		if start > 0 && end > start {
			detEvento.ChNFeRef = tagAdic[start:end]
		}
	}

	if strings.Contains(tagAdic, "<verAplic>") {
		start := strings.Index(tagAdic, "<verAplic>") + len("<verAplic>")
		end := strings.Index(tagAdic, "</verAplic>")
		if start > 0 && end > start {
			detEvento.VerAplic = tagAdic[start:end]
		}
	}

	return nil
}

// ValidateEventParams validates the event parameters
func ValidateEventParams(params EventParams) error {
	if params.ChNFe == "" {
		return fmt.Errorf("chNFe is required")
	}
	if len(params.ChNFe) != 44 {
		return fmt.Errorf("chNFe must be 44 characters long")
	}
	if params.CNPJ == "" {
		return fmt.Errorf("CNPJ is required")
	}
	if len(params.CNPJ) != 14 {
		return fmt.Errorf("CNPJ must be 14 characters long")
	}
	if params.TpEvento == 0 {
		return fmt.Errorf("tpEvento is required")
	}
	if params.UF == "" {
		return fmt.Errorf("UF is required")
	}
	if params.TpAmb == "" {
		return fmt.Errorf("tpAmb is required")
	}

	// Validate event type
	if _, err := GetEventInfo(params.TpEvento); err != nil {
		return err
	}

	return nil
}
