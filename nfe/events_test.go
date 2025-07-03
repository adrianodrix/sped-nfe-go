package nfe

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestGetEventInfo(t *testing.T) {
	tests := []struct {
		name      string
		eventType int
		want      EventInfoNFe
		wantErr   bool
	}{
		{
			name:      "CCe event",
			eventType: EVT_CCE,
			want:      EventInfoNFe{Version: "1.00", Name: "envCCe"},
			wantErr:   false,
		},
		{
			name:      "Cancel event",
			eventType: EVT_CANCELA,
			want:      EventInfoNFe{Version: "1.00", Name: "envEventoCancNFe"},
			wantErr:   false,
		},
		{
			name:      "Invalid event",
			eventType: 999999,
			want:      EventInfoNFe{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEventInfo(tt.eventType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetEventInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateEventXML(t *testing.T) {
	testTime := time.Date(2023, 12, 1, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		params  EventParams
		wantErr bool
	}{
		{
			name: "Valid CCe event",
			params: EventParams{
				UF:         "SP",
				ChNFe:      "35220214200166000187550010000000101234567890",
				TpEvento:   EVT_CCE,
				NSeqEvento: 1,
				TagAdic:    "<xCorrecao>Teste</xCorrecao><xCondUso>Condicoes de uso</xCondUso>",
				DhEvento:   &testTime,
				Lote:       "123",
				CNPJ:       "14200166000187",
				TpAmb:      "2",
				VerEvento:  "1.00",
			},
			wantErr: false,
		},
		{
			name: "Missing ChNFe",
			params: EventParams{
				UF:       "SP",
				TpEvento: EVT_CCE,
				CNPJ:     "14200166000187",
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Missing CNPJ",
			params: EventParams{
				UF:       "SP",
				ChNFe:    "35220214200166000187550010000000101234567890",
				TpEvento: EVT_CCE,
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Invalid event type",
			params: EventParams{
				UF:       "SP",
				ChNFe:    "35220214200166000187550010000000101234567890",
				TpEvento: 999999,
				CNPJ:     "14200166000187",
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Invalid UF",
			params: EventParams{
				UF:       "XX",
				ChNFe:    "35220214200166000187550010000000101234567890",
				TpEvento: EVT_CCE,
				CNPJ:     "14200166000187",
				TpAmb:    "2",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateEventXML(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateEventXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("CreateEventXML() returned nil but no error")
			}
			if !tt.wantErr && got != nil {
				// Validate basic structure
				if got.Xmlns != "http://www.portalfiscal.inf.br/nfe" {
					t.Errorf("CreateEventXML() xmlns = %v, want %v", got.Xmlns, "http://www.portalfiscal.inf.br/nfe")
				}
				if got.Evento.InfEvento.ChNFe != tt.params.ChNFe {
					t.Errorf("CreateEventXML() chNFe = %v, want %v", got.Evento.InfEvento.ChNFe, tt.params.ChNFe)
				}
			}
		})
	}
}

func TestCreateEventXML_XMLGeneration(t *testing.T) {
	testTime := time.Date(2023, 12, 1, 10, 30, 0, 0, time.UTC)

	params := EventParams{
		UF:         "SP",
		ChNFe:      "35220214200166000187550010000000101234567890",
		TpEvento:   EVT_CCE,
		NSeqEvento: 1,
		TagAdic:    "<xCorrecao>Correção de teste</xCorrecao><xCondUso>Condições de uso padrão</xCondUso>",
		DhEvento:   &testTime,
		Lote:       "123456",
		CNPJ:       "14200166000187",
		TpAmb:      "2",
		VerEvento:  "1.00",
	}

	eventXML, err := CreateEventXML(params)
	if err != nil {
		t.Fatalf("CreateEventXML() error = %v", err)
	}

	// Test XML marshaling
	xmlData, err := xml.Marshal(eventXML)
	if err != nil {
		t.Fatalf("xml.Marshal() error = %v", err)
	}

	// Verify XML contains expected elements
	xmlString := string(xmlData)
	expectedElements := []string{
		"<envEvento",
		"xmlns=\"http://www.portalfiscal.inf.br/nfe\"",
		"versao=\"1.00\"",
		"<idLote>123456</idLote>",
		"<evento",
		"<infEvento",
		"<cOrgao>35</cOrgao>",
		"<tpAmb>2</tpAmb>",
		"<CNPJ>14200166000187</CNPJ>",
		"<chNFe>35220214200166000187550010000000101234567890</chNFe>",
		"<tpEvento>110110</tpEvento>",
		"<nSeqEvento>01</nSeqEvento>",
		"<verEvento>1.00</verEvento>",
		"<detEvento",
		"<xCorrecao>Correção de teste</xCorrecao>",
		"<xCondUso>Condições de uso padrão</xCondUso>",
	}

	for _, expected := range expectedElements {
		if !containsStr(xmlString, expected) {
			t.Errorf("XML does not contain expected element: %s", expected)
		}
	}
}

func TestParseAdditionalTags(t *testing.T) {
	tests := []struct {
		name     string
		tagAdic  string
		expected DetEventoNFe
	}{
		{
			name:    "CCe tags",
			tagAdic: "<xCorrecao>Teste de correção</xCorrecao><xCondUso>Condições padrão</xCondUso>",
			expected: DetEventoNFe{
				XCorrecao: "Teste de correção",
				XCondUso:  "Condições padrão",
			},
		},
		{
			name:    "Cancellation tags",
			tagAdic: "<nProt>135220000000123</nProt><xJust>Justificativa do cancelamento</xJust>",
			expected: DetEventoNFe{
				NProt: "135220000000123",
				XJust: "Justificativa do cancelamento",
			},
		},
		{
			name:    "Substitution tags",
			tagAdic: "<nProt>135220000000123</nProt><xJust>Cancelamento por substituição</xJust><chNFeRef>35220214200166000187550010000000101234567891</chNFeRef><verAplic>4.00</verAplic>",
			expected: DetEventoNFe{
				NProt:    "135220000000123",
				XJust:    "Cancelamento por substituição",
				ChNFeRef: "35220214200166000187550010000000101234567891",
				VerAplic: "4.00",
			},
		},
		{
			name:     "Empty tags",
			tagAdic:  "",
			expected: DetEventoNFe{},
		},
		{
			name:    "Malformed tags",
			tagAdic: "<xCorrecao>Test</xCorreca><xCondUso>Condition",
			expected: DetEventoNFe{
				XCorrecao: "",
				XCondUso:  "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var detEvento DetEventoNFe
			err := parseAdditionalTags(&detEvento, tt.tagAdic)
			if err != nil {
				t.Errorf("parseAdditionalTags() error = %v", err)
				return
			}

			if detEvento.XCorrecao != tt.expected.XCorrecao {
				t.Errorf("parseAdditionalTags() XCorrecao = %v, want %v", detEvento.XCorrecao, tt.expected.XCorrecao)
			}
			if detEvento.XCondUso != tt.expected.XCondUso {
				t.Errorf("parseAdditionalTags() XCondUso = %v, want %v", detEvento.XCondUso, tt.expected.XCondUso)
			}
			if detEvento.NProt != tt.expected.NProt {
				t.Errorf("parseAdditionalTags() NProt = %v, want %v", detEvento.NProt, tt.expected.NProt)
			}
			if detEvento.XJust != tt.expected.XJust {
				t.Errorf("parseAdditionalTags() XJust = %v, want %v", detEvento.XJust, tt.expected.XJust)
			}
			if detEvento.ChNFeRef != tt.expected.ChNFeRef {
				t.Errorf("parseAdditionalTags() ChNFeRef = %v, want %v", detEvento.ChNFeRef, tt.expected.ChNFeRef)
			}
			if detEvento.VerAplic != tt.expected.VerAplic {
				t.Errorf("parseAdditionalTags() VerAplic = %v, want %v", detEvento.VerAplic, tt.expected.VerAplic)
			}
		})
	}
}

func TestValidateEventParams(t *testing.T) {
	tests := []struct {
		name    string
		params  EventParams
		wantErr bool
	}{
		{
			name: "Valid params",
			params: EventParams{
				ChNFe:    "35220214200166000187550010000000101234567890",
				CNPJ:     "14200166000187",
				TpEvento: EVT_CCE,
				UF:       "SP",
				TpAmb:    "2",
			},
			wantErr: false,
		},
		{
			name: "Empty ChNFe",
			params: EventParams{
				CNPJ:     "14200166000187",
				TpEvento: EVT_CCE,
				UF:       "SP",
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Invalid ChNFe length",
			params: EventParams{
				ChNFe:    "1234567890",
				CNPJ:     "14200166000187",
				TpEvento: EVT_CCE,
				UF:       "SP",
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Empty CNPJ",
			params: EventParams{
				ChNFe:    "35220214200166000187550010000000101234567890",
				TpEvento: EVT_CCE,
				UF:       "SP",
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Invalid CNPJ length",
			params: EventParams{
				ChNFe:    "35220214200166000187550010000000101234567890",
				CNPJ:     "123456789",
				TpEvento: EVT_CCE,
				UF:       "SP",
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Invalid event type",
			params: EventParams{
				ChNFe:    "35220214200166000187550010000000101234567890",
				CNPJ:     "14200166000187",
				TpEvento: 0,
				UF:       "SP",
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Empty UF",
			params: EventParams{
				ChNFe:    "35220214200166000187550010000000101234567890",
				CNPJ:     "14200166000187",
				TpEvento: EVT_CCE,
				TpAmb:    "2",
			},
			wantErr: true,
		},
		{
			name: "Empty TpAmb",
			params: EventParams{
				ChNFe:    "35220214200166000187550010000000101234567890",
				CNPJ:     "14200166000187",
				TpEvento: EVT_CCE,
				UF:       "SP",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateEventParams(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("ValidateEventParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEventConstants(t *testing.T) {
	// Test that all event constants have the expected values
	expectedValues := map[string]int{
		"EVT_CONFIRMACAO":                      210200,
		"EVT_CIENCIA":                          210210,
		"EVT_DESCONHECIMENTO":                  210220,
		"EVT_NAO_REALIZADA":                    210240,
		"EVT_CCE":                              110110,
		"EVT_CANCELA":                          110111,
		"EVT_CANCELASUBSTITUICAO":              110112,
		"EVT_EPEC":                             110140,
		"EVT_ATORINTERESSADO":                  110150,
		"EVT_COMPROVANTE_ENTREGA":              110130,
		"EVT_CANCELAMENTO_COMPROVANTE_ENTREGA": 110131,
		"EVT_PRORROGACAO_1":                    111500,
		"EVT_PRORROGACAO_2":                    111501,
		"EVT_CANCELA_PRORROGACAO_1":            111502,
		"EVT_CANCELA_PRORROGACAO_2":            111503,
		"EVT_INSUCESSO_ENTREGA":                110192,
		"EVT_CANCELA_INSUCESSO_ENTREGA":        110193,
		"EVT_CONCILIACAO":                      110750,
		"EVT_CANCELA_CONCILIACAO":              110751,
	}

	actualValues := map[string]int{
		"EVT_CONFIRMACAO":                      EVT_CONFIRMACAO,
		"EVT_CIENCIA":                          EVT_CIENCIA,
		"EVT_DESCONHECIMENTO":                  EVT_DESCONHECIMENTO,
		"EVT_NAO_REALIZADA":                    EVT_NAO_REALIZADA,
		"EVT_CCE":                              EVT_CCE,
		"EVT_CANCELA":                          EVT_CANCELA,
		"EVT_CANCELASUBSTITUICAO":              EVT_CANCELASUBSTITUICAO,
		"EVT_EPEC":                             EVT_EPEC,
		"EVT_ATORINTERESSADO":                  EVT_ATORINTERESSADO,
		"EVT_COMPROVANTE_ENTREGA":              EVT_COMPROVANTE_ENTREGA,
		"EVT_CANCELAMENTO_COMPROVANTE_ENTREGA": EVT_CANCELAMENTO_COMPROVANTE_ENTREGA,
		"EVT_PRORROGACAO_1":                    EVT_PRORROGACAO_1,
		"EVT_PRORROGACAO_2":                    EVT_PRORROGACAO_2,
		"EVT_CANCELA_PRORROGACAO_1":            EVT_CANCELA_PRORROGACAO_1,
		"EVT_CANCELA_PRORROGACAO_2":            EVT_CANCELA_PRORROGACAO_2,
		"EVT_INSUCESSO_ENTREGA":                EVT_INSUCESSO_ENTREGA,
		"EVT_CANCELA_INSUCESSO_ENTREGA":        EVT_CANCELA_INSUCESSO_ENTREGA,
		"EVT_CONCILIACAO":                      EVT_CONCILIACAO,
		"EVT_CANCELA_CONCILIACAO":              EVT_CANCELA_CONCILIACAO,
	}

	for name, expected := range expectedValues {
		if actual := actualValues[name]; actual != expected {
			t.Errorf("Constant %s = %v, want %v", name, actual, expected)
		}
	}
}

// Helper function to check if a string contains a substring
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
