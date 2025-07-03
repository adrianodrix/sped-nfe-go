package nfe

import (
	"strings"
	"testing"
)

func TestValidateCorrection(t *testing.T) {
	tests := []struct {
		name      string
		correcao  string
		wantErr   bool
		expectErr string
	}{
		{
			name:     "Valid correction text",
			correcao: "Corrigir nome do produto de ABC para XYZ",
			wantErr:  false,
		},
		{
			name:      "Empty correction text",
			correcao:  "",
			wantErr:   true,
			expectErr: "correction text cannot be empty",
		},
		{
			name:      "Too short correction text",
			correcao:  "Muito curto",
			wantErr:   true,
			expectErr: "correction text must be at least 15 characters",
		},
		{
			name:     "Minimum valid length",
			correcao: "Corrigir produto",
			wantErr:  false,
		},
		{
			name:     "Maximum valid length",
			correcao: strings.Repeat("a", CCeMaxCorrectionLength),
			wantErr:  false,
		},
		{
			name:      "Too long correction text",
			correcao:  strings.Repeat("a", CCeMaxCorrectionLength+1),
			wantErr:   true,
			expectErr: "correction text cannot exceed 1000 characters",
		},
		{
			name:     "Valid with spaces",
			correcao: "  Corrigir nome do produto de ABC para XYZ  ",
			wantErr:  false,
		},
		{
			name:     "Valid with special characters",
			correcao: "Corrigir: produto ABC -> XYZ (quantidade 10)",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCorrection(tt.correcao)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCorrection() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("ValidateCorrection() error = %v, expected to contain %q", err, tt.expectErr)
			}
		})
	}
}

func TestValidateSequence(t *testing.T) {
	tests := []struct {
		name      string
		sequencia int
		wantErr   bool
		expectErr string
	}{
		{
			name:      "Valid sequence 1",
			sequencia: 1,
			wantErr:   false,
		},
		{
			name:      "Valid sequence 10",
			sequencia: 10,
			wantErr:   false,
		},
		{
			name:      "Valid sequence 20",
			sequencia: 20,
			wantErr:   false,
		},
		{
			name:      "Too low sequence",
			sequencia: 0,
			wantErr:   true,
			expectErr: "sequence must be at least 1",
		},
		{
			name:      "Too high sequence",
			sequencia: 21,
			wantErr:   true,
			expectErr: "sequence cannot exceed 20",
		},
		{
			name:      "Negative sequence",
			sequencia: -1,
			wantErr:   true,
			expectErr: "sequence must be at least 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSequence(tt.sequencia)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSequence() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("ValidateSequence() error = %v, expected to contain %q", err, tt.expectErr)
			}
		})
	}
}

func TestSanitizeCorrection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal text",
			input:    "Corrigir nome do produto",
			expected: "Corrigir nome do produto",
		},
		{
			name:     "Text with extra spaces",
			input:    "  Corrigir   nome    do   produto  ",
			expected: "Corrigir nome do produto",
		},
		{
			name:     "Text with newlines and tabs",
			input:    "Corrigir\nnome\tdo\rproduto",
			expected: "Corrigir nome do produto",
		},
		{
			name:     "Text with ampersand",
			input:    "Corrigir nome do produto & quantidade",
			expected: "Corrigir nome do produto e quantidade",
		},
		{
			name:     "Text with HTML-like tags",
			input:    "Corrigir <nome> do produto",
			expected: "Corrigir nome do produto",
		},
		{
			name:     "Text too long",
			input:    strings.Repeat("a", 1200),
			expected: strings.Repeat("a", CCeMaxCorrectionLength),
		},
		{
			name:     "Empty text",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "   \n\t   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeCorrection(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeCorrection() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateCCeRequest(t *testing.T) {
	tests := []struct {
		name      string
		chaveNFe  string
		correcao  string
		sequencia int
		wantErr   bool
		expectErr string
	}{
		{
			name:      "Valid request",
			chaveNFe:  "35220214200166000187550010000000101234567890",
			correcao:  "Corrigir nome do produto de ABC para XYZ",
			sequencia: 1,
			wantErr:   false,
		},
		{
			name:      "Invalid NFe key",
			chaveNFe:  "invalid",
			correcao:  "Corrigir nome do produto de ABC para XYZ",
			sequencia: 1,
			wantErr:   true,
			expectErr: "invalid NFe key",
		},
		{
			name:      "Invalid correction text",
			chaveNFe:  "35220214200166000187550010000000101234567890",
			correcao:  "short",
			sequencia: 1,
			wantErr:   true,
			expectErr: "invalid correction text",
		},
		{
			name:      "Invalid sequence",
			chaveNFe:  "35220214200166000187550010000000101234567890",
			correcao:  "Corrigir nome do produto de ABC para XYZ",
			sequencia: 0,
			wantErr:   true,
			expectErr: "invalid sequence",
		},
		{
			name:      "Whitespace in fields",
			chaveNFe:  "  35220214200166000187550010000000101234567890  ",
			correcao:  "  Corrigir nome do produto de ABC para XYZ  ",
			sequencia: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := CreateCCeRequest(tt.chaveNFe, tt.correcao, tt.sequencia)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCCeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("CreateCCeRequest() error = %v, expected to contain %q", err, tt.expectErr)
			}
			if !tt.wantErr && req == nil {
				t.Errorf("CreateCCeRequest() returned nil request but no error")
			}
			if !tt.wantErr && req != nil {
				// Validate that fields are properly trimmed and set
				if strings.TrimSpace(req.ChaveNFe) != req.ChaveNFe {
					t.Errorf("ChaveNFe not properly trimmed: %q", req.ChaveNFe)
				}
				if req.XCondUso != CCeUsageConditions {
					t.Errorf("XCondUso not set to default: %q", req.XCondUso)
				}
				if req.Sequencia != tt.sequencia {
					t.Errorf("Sequencia not set correctly: got %d, want %d", req.Sequencia, tt.sequencia)
				}
			}
		})
	}
}

func TestValidarCCe(t *testing.T) {
	validReq := &CCeRequest{
		ChaveNFe:  "35220214200166000187550010000000101234567890",
		Correcao:  "Corrigir nome do produto de ABC para XYZ",
		Sequencia: 1,
	}

	tests := []struct {
		name      string
		req       *CCeRequest
		wantErr   bool
		expectErr string
	}{
		{
			name:    "Valid request",
			req:     validReq,
			wantErr: false,
		},
		{
			name:      "Nil request",
			req:       nil,
			wantErr:   true,
			expectErr: "CCe request cannot be nil",
		},
		{
			name: "Invalid NFe key",
			req: &CCeRequest{
				ChaveNFe:  "invalid",
				Correcao:  "Corrigir nome do produto de ABC para XYZ",
				Sequencia: 1,
			},
			wantErr:   true,
			expectErr: "invalid NFe key",
		},
		{
			name: "Invalid correction text",
			req: &CCeRequest{
				ChaveNFe:  "35220214200166000187550010000000101234567890",
				Correcao:  "short",
				Sequencia: 1,
			},
			wantErr:   true,
			expectErr: "invalid correction text",
		},
		{
			name: "Invalid sequence",
			req: &CCeRequest{
				ChaveNFe:  "35220214200166000187550010000000101234567890",
				Correcao:  "Corrigir nome do produto de ABC para XYZ",
				Sequencia: 0,
			},
			wantErr:   true,
			expectErr: "invalid sequence",
		},
		{
			name: "Request with empty usage conditions",
			req: &CCeRequest{
				ChaveNFe:  "35220214200166000187550010000000101234567890",
				Correcao:  "Corrigir nome do produto de ABC para XYZ",
				Sequencia: 1,
				XCondUso:  "",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalCondUso := ""
			if tt.req != nil {
				originalCondUso = tt.req.XCondUso
			}

			err := ValidarCCe(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidarCCe() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("ValidarCCe() error = %v, expected to contain %q", err, tt.expectErr)
			}

			// Check if usage conditions were set when empty
			if !tt.wantErr && tt.req != nil && originalCondUso == "" {
				if tt.req.XCondUso != CCeUsageConditions {
					t.Errorf("XCondUso not set to default when empty")
				}
			}
		})
	}
}

func TestGetCCeStatusText(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{
			name:   "Registered status",
			status: int(CCeStatusRegistered),
			want:   "Evento registrado e vinculado à NFe",
		},
		{
			name:   "Registered not linked status",
			status: int(CCeStatusRegisteredNotLinked),
			want:   "Evento registrado, mas não vinculado à NFe",
		},
		{
			name:   "Sequence exceeded status",
			status: int(CCeStatusSequenceExceeded),
			want:   "Número de sequência maior que permitido",
		},
		{
			name:   "Already exists status",
			status: int(CCeStatusAlreadyExists),
			want:   "Evento já existe para esta NFe com a mesma sequência",
		},
		{
			name:   "NFe not found status",
			status: int(CCeStatusNFeNotFound),
			want:   "Evento rejeitado - NFe não encontrada",
		},
		{
			name:   "NFe not authorized status",
			status: int(CCeStatusNFeNotAuthorized),
			want:   "Evento rejeitado - NFe não autorizada",
		},
		{
			name:   "Invalid correction status",
			status: int(CCeStatusInvalidCorrection),
			want:   "Evento rejeitado - texto de correção inválido",
		},
		{
			name:   "Unknown status",
			status: 999,
			want:   "Status desconhecido: 999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCCeStatusText(tt.status)
			if got != tt.want {
				t.Errorf("GetCCeStatusText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsCCeSuccessful(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{
			name:   "Registered status - success",
			status: int(CCeStatusRegistered),
			want:   true,
		},
		{
			name:   "Registered not linked status - success",
			status: int(CCeStatusRegisteredNotLinked),
			want:   true,
		},
		{
			name:   "Sequence exceeded status - not success",
			status: int(CCeStatusSequenceExceeded),
			want:   false,
		},
		{
			name:   "Already exists status - not success",
			status: int(CCeStatusAlreadyExists),
			want:   false,
		},
		{
			name:   "NFe not found status - not success",
			status: int(CCeStatusNFeNotFound),
			want:   false,
		},
		{
			name:   "Unknown status - not success",
			status: 999,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCCeSuccessful(tt.status)
			if got != tt.want {
				t.Errorf("IsCCeSuccessful() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanSendCCe(t *testing.T) {
	tests := []struct {
		name       string
		authorized bool
		sequencia  int
		wantResult bool
		wantErr    bool
		expectErr  string
	}{
		{
			name:       "Can send CCe - authorized and valid sequence",
			authorized: true,
			sequencia:  1,
			wantResult: true,
			wantErr:    false,
		},
		{
			name:       "Cannot send CCe - not authorized",
			authorized: false,
			sequencia:  1,
			wantResult: false,
			wantErr:    true,
			expectErr:  "NFe must be authorized",
		},
		{
			name:       "Cannot send CCe - invalid sequence too low",
			authorized: true,
			sequencia:  0,
			wantResult: false,
			wantErr:    true,
			expectErr:  "invalid sequence number",
		},
		{
			name:       "Cannot send CCe - invalid sequence too high",
			authorized: true,
			sequencia:  21,
			wantResult: false,
			wantErr:    true,
			expectErr:  "maximum CCe sequence exceeded",
		},
		{
			name:       "Can send CCe - maximum valid sequence",
			authorized: true,
			sequencia:  20,
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CanSendCCe(tt.authorized, tt.sequencia)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanSendCCe() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result != tt.wantResult {
				t.Errorf("CanSendCCe() result = %v, want %v", result, tt.wantResult)
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("CanSendCCe() error = %v, expected to contain %q", err, tt.expectErr)
			}
		})
	}
}

func TestGetNextSequence(t *testing.T) {
	tests := []struct {
		name         string
		lastSequence int
		wantResult   int
		wantErr      bool
		expectErr    string
	}{
		{
			name:         "Next sequence from 1",
			lastSequence: 1,
			wantResult:   2,
			wantErr:      false,
		},
		{
			name:         "Next sequence from 19",
			lastSequence: 19,
			wantResult:   20,
			wantErr:      false,
		},
		{
			name:         "Next sequence from 20 - exceeds maximum",
			lastSequence: 20,
			wantResult:   0,
			wantErr:      true,
			expectErr:    "maximum CCe sequence exceeded",
		},
		{
			name:         "Next sequence from 0",
			lastSequence: 0,
			wantResult:   1,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetNextSequence(tt.lastSequence)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNextSequence() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result != tt.wantResult {
				t.Errorf("GetNextSequence() result = %v, want %v", result, tt.wantResult)
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("GetNextSequence() error = %v, expected to contain %q", err, tt.expectErr)
			}
		})
	}
}

func TestValidateSequenceIncrement(t *testing.T) {
	tests := []struct {
		name            string
		currentSequence int
		newSequence     int
		wantErr         bool
		expectErr       string
	}{
		{
			name:            "Valid increment from 1 to 2",
			currentSequence: 1,
			newSequence:     2,
			wantErr:         false,
		},
		{
			name:            "Valid increment from 19 to 20",
			currentSequence: 19,
			newSequence:     20,
			wantErr:         false,
		},
		{
			name:            "Invalid increment - skipping sequence",
			currentSequence: 1,
			newSequence:     3,
			wantErr:         true,
			expectErr:       "sequence must be incremental",
		},
		{
			name:            "Invalid increment - going backwards",
			currentSequence: 5,
			newSequence:     4,
			wantErr:         true,
			expectErr:       "sequence must be incremental",
		},
		{
			name:            "Invalid increment - exceeds maximum",
			currentSequence: 20,
			newSequence:     21,
			wantErr:         true,
			expectErr:       "sequence 21 exceeds maximum allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSequenceIncrement(tt.currentSequence, tt.newSequence)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSequenceIncrement() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("ValidateSequenceIncrement() error = %v, expected to contain %q", err, tt.expectErr)
			}
		})
	}
}

func TestCreateCCeTagAdic(t *testing.T) {
	tests := []struct {
		name     string
		correcao string
		condUso  string
		expected string
	}{
		{
			name:     "Normal correction with default conditions",
			correcao: "Corrigir nome do produto",
			condUso:  "",
			expected: "<xCorrecao>Corrigir nome do produto</xCorrecao><xCondUso>" + CCeUsageConditions + "</xCondUso>",
		},
		{
			name:     "Normal correction with custom conditions",
			correcao: "Corrigir nome do produto",
			condUso:  "Custom conditions",
			expected: "<xCorrecao>Corrigir nome do produto</xCorrecao><xCondUso>Custom conditions</xCondUso>",
		},
		{
			name:     "Correction with extra spaces",
			correcao: "  Corrigir   nome   do   produto  ",
			condUso:  "",
			expected: "<xCorrecao>Corrigir nome do produto</xCorrecao><xCondUso>" + CCeUsageConditions + "</xCondUso>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateCCeTagAdic(tt.correcao, tt.condUso)
			if result != tt.expected {
				t.Errorf("CreateCCeTagAdic() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseCCeSequenceFromKey(t *testing.T) {
	tests := []struct {
		name      string
		eventID   string
		wantSeq   int
		wantErr   bool
		expectErr string
	}{
		{
			name:    "Valid CCe event ID with sequence 01",
			eventID: "ID11011035220214200166000187550010000000101234567890001",
			wantSeq: 1,
			wantErr: false,
		},
		{
			name:    "Valid CCe event ID with sequence 10",
			eventID: "ID11011035220214200166000187550010000000101234567890010",
			wantSeq: 10,
			wantErr: false,
		},
		{
			name:    "Valid CCe event ID with sequence 20",
			eventID: "ID11011035220214200166000187550010000000101234567890020",
			wantSeq: 20,
			wantErr: false,
		},
		{
			name:      "Invalid event ID - too short",
			eventID:   "ID110110352202142001660001875500100000001012345678901",
			wantSeq:   0,
			wantErr:   true,
			expectErr: "invalid event ID format",
		},
		{
			name:      "Invalid event ID - non-numeric sequence",
			eventID:   "ID110110352202142001660001875500100000001012345678900AB",
			wantSeq:   0,
			wantErr:   true,
			expectErr: "invalid sequence in event ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq, err := ParseCCeSequenceFromKey(tt.eventID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCCeSequenceFromKey() error = %v, wantErr %v", err, tt.wantErr)
			}
			if seq != tt.wantSeq {
				t.Errorf("ParseCCeSequenceFromKey() sequence = %v, want %v", seq, tt.wantSeq)
			}
			if tt.wantErr && tt.expectErr != "" && !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("ParseCCeSequenceFromKey() error = %v, expected to contain %q", err, tt.expectErr)
			}
		})
	}
}

func TestFormatCCeSequence(t *testing.T) {
	tests := []struct {
		name      string
		sequencia int
		expected  string
	}{
		{
			name:      "Format sequence 1",
			sequencia: 1,
			expected:  "01",
		},
		{
			name:      "Format sequence 10",
			sequencia: 10,
			expected:  "10",
		},
		{
			name:      "Format sequence 20",
			sequencia: 20,
			expected:  "20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCCeSequence(tt.sequencia)
			if result != tt.expected {
				t.Errorf("FormatCCeSequence() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIsValidCorrectionContent(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "Valid correction text",
			text: "Corrigir nome do produto de ABC para XYZ",
			want: true,
		},
		{
			name: "Valid single character",
			text: "A",
			want: true,
		},
		{
			name: "Valid text with special characters",
			text: "Corrigir: produto ABC -> XYZ (quantidade 10)",
			want: true,
		},
		{
			name: "Valid text with accents",
			text: "Correção do produto açúcar",
			want: true,
		},
		{
			name: "Empty text",
			text: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidCorrectionContent(tt.text)
			if got != tt.want {
				t.Errorf("isValidCorrectionContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetCCeEventName(t *testing.T) {
	expected := "Carta de Correção Eletrônica"
	got := GetCCeEventName()
	if got != expected {
		t.Errorf("GetCCeEventName() = %q, want %q", got, expected)
	}
}

func TestGetCCeEventDescription(t *testing.T) {
	expected := "Carta de Correção Eletrônica"
	got := GetCCeEventDescription()
	if got != expected {
		t.Errorf("GetCCeEventDescription() = %q, want %q", got, expected)
	}
}

func TestCCeUsageConditions(t *testing.T) {
	// Test that the usage conditions constant is properly defined
	if CCeUsageConditions == "" {
		t.Errorf("CCeUsageConditions constant is empty")
	}

	// Test that it contains expected text
	expectedParts := []string{
		"Carta de Correção",
		"disciplinada",
		"art. 7º",
		"Convênio S/N",
		"15 de dezembro de 1970",
		"regularização de erro",
		"valor do imposto",
		"base de cálculo",
		"alíquota",
		"data de emissão",
	}

	for _, part := range expectedParts {
		if !strings.Contains(CCeUsageConditions, part) {
			t.Errorf("CCeUsageConditions does not contain expected text: %q", part)
		}
	}
}

func TestCCeConstants(t *testing.T) {
	// Test that constants are properly defined
	if CCeTypeEvent != 110110 {
		t.Errorf("CCeTypeEvent = %d, want 110110", CCeTypeEvent)
	}

	if CCeMinSequence != 1 {
		t.Errorf("CCeMinSequence = %d, want 1", CCeMinSequence)
	}

	if CCeMaxSequence != 20 {
		t.Errorf("CCeMaxSequence = %d, want 20", CCeMaxSequence)
	}

	if CCeMinCorrectionLength != 15 {
		t.Errorf("CCeMinCorrectionLength = %d, want 15", CCeMinCorrectionLength)
	}

	if CCeMaxCorrectionLength != 1000 {
		t.Errorf("CCeMaxCorrectionLength = %d, want 1000", CCeMaxCorrectionLength)
	}
}