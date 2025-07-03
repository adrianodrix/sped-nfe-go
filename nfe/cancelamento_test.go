package nfe

import (
	"strings"
	"testing"
	"time"
)

func TestValidateNFeKey(t *testing.T) {
	tests := []struct {
		name    string
		chave   string
		wantErr bool
	}{
		{
			name:    "Valid NFe key",
			chave:   "35220214200166000187550010000000101234567890",
			wantErr: false,
		},
		{
			name:    "Empty key",
			chave:   "",
			wantErr: true,
		},
		{
			name:    "Too short key",
			chave:   "123456789012345678901234567890123456789",
			wantErr: true,
		},
		{
			name:    "Too long key",
			chave:   "123456789012345678901234567890123456789012345",
			wantErr: true,
		},
		{
			name:    "Non-numeric characters",
			chave:   "3522021420016600018755001000000010123456789A",
			wantErr: true,
		},
		{
			name:    "Invalid UF code",
			chave:   "99220214200166000187550010000000101234567890",
			wantErr: true,
		},
		{
			name:    "Invalid month in date",
			chave:   "35229914200166000187550010000000101234567890",
			wantErr: true,
		},
		{
			name:    "Zero month in date",
			chave:   "35220014200166000187550010000000101234567890",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNFeKey(tt.chave)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNFeKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateJustification(t *testing.T) {
	tests := []struct {
		name          string
		justificativa string
		wantErr       bool
	}{
		{
			name:          "Valid justification",
			justificativa: "Produto entregue com defeito, solicitado pelo cliente",
			wantErr:       false,
		},
		{
			name:          "Empty justification",
			justificativa: "",
			wantErr:       true,
		},
		{
			name:          "Too short justification",
			justificativa: "Muito curto",
			wantErr:       true,
		},
		{
			name:          "Minimum valid length",
			justificativa: "Erro na emisssao",
			wantErr:       false,
		},
		{
			name:          "Maximum valid length",
			justificativa: strings.Repeat("a", MaxJustificationLength),
			wantErr:       false,
		},
		{
			name:          "Too long justification",
			justificativa: strings.Repeat("a", MaxJustificationLength+1),
			wantErr:       true,
		},
		{
			name:          "Only spaces",
			justificativa: "               ",
			wantErr:       true,
		},
		{
			name:          "Only special characters",
			justificativa: "!@#$%^&*()_+{}[]",
			wantErr:       true,
		},
		{
			name:          "Valid with spaces",
			justificativa: "  Produto com defeito solicitado pelo cliente  ",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJustification(tt.justificativa)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJustification() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidarPrazoCancelamento(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		dhAutorizacao time.Time
		wantErr       bool
	}{
		{
			name:          "Within deadline - 1 hour ago",
			dhAutorizacao: now.Add(-1 * time.Hour),
			wantErr:       false,
		},
		{
			name:          "Within deadline - 23 hours ago",
			dhAutorizacao: now.Add(-23 * time.Hour),
			wantErr:       false,
		},
		{
			name:          "Close to deadline - 23h59m ago",
			dhAutorizacao: now.Add(-23*time.Hour - 59*time.Minute),
			wantErr:       false,
		},
		{
			name:          "Outside deadline - 25 hours ago",
			dhAutorizacao: now.Add(-25 * time.Hour),
			wantErr:       true,
		},
		{
			name:          "Outside deadline - 48 hours ago",
			dhAutorizacao: now.Add(-48 * time.Hour),
			wantErr:       true,
		},
		{
			name:          "Zero time",
			dhAutorizacao: time.Time{},
			wantErr:       true,
		},
		{
			name:          "Future time",
			dhAutorizacao: now.Add(1 * time.Hour),
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidarPrazoCancelamento(tt.dhAutorizacao)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidarPrazoCancelamento() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCanBeCancelled(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		authorized    bool
		cancelled     bool
		dhAutorizacao time.Time
		wantResult    bool
		wantErr       bool
	}{
		{
			name:          "Can be cancelled - authorized and within deadline",
			authorized:    true,
			cancelled:     false,
			dhAutorizacao: now.Add(-1 * time.Hour),
			wantResult:    true,
			wantErr:       false,
		},
		{
			name:          "Cannot be cancelled - not authorized",
			authorized:    false,
			cancelled:     false,
			dhAutorizacao: now.Add(-1 * time.Hour),
			wantResult:    false,
			wantErr:       true,
		},
		{
			name:          "Cannot be cancelled - already cancelled",
			authorized:    true,
			cancelled:     true,
			dhAutorizacao: now.Add(-1 * time.Hour),
			wantResult:    false,
			wantErr:       true,
		},
		{
			name:          "Cannot be cancelled - outside deadline",
			authorized:    true,
			cancelled:     false,
			dhAutorizacao: now.Add(-25 * time.Hour),
			wantResult:    false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := CanBeCancelled(tt.authorized, tt.cancelled, tt.dhAutorizacao)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanBeCancelled() error = %v, wantErr %v", err, tt.wantErr)
			}
			if result != tt.wantResult {
				t.Errorf("CanBeCancelled() result = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestSanitizeJustification(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      string
	}{
		{
			name:     "Normal text",
			input:    "Produto com defeito",
			expected: "Produto com defeito",
		},
		{
			name:     "Text with extra spaces",
			input:    "  Produto   com    defeito  ",
			expected: "Produto com defeito",
		},
		{
			name:     "Text with newlines and tabs",
			input:    "Produto\ncom\tdefeito",
			expected: "Produto com defeito",
		},
		{
			name:     "Text with quotes",
			input:    "Produto \"com defeito\"",
			expected: "Produto 'com defeito'",
		},
		{
			name:     "Text with ampersand",
			input:    "Produto com defeito & problema",
			expected: "Produto com defeito e problema",
		},
		{
			name:     "Text with HTML-like tags",
			input:    "Produto <em>com</em> defeito",
			expected: "Produto com defeito",
		},
		{
			name:     "Text too long",
			input:    strings.Repeat("a", 300),
			expected: strings.Repeat("a", MaxJustificationLength),
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
			result := SanitizeJustification(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeJustification() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCreateCancelamentoRequest(t *testing.T) {
	tests := []struct {
		name          string
		chaveNFe      string
		justificativa string
		protocolo     string
		wantErr       bool
	}{
		{
			name:          "Valid request",
			chaveNFe:      "35220214200166000187550010000000101234567890",
			justificativa: "Produto entregue com defeito, solicitado pelo cliente",
			protocolo:     "135220000000123",
			wantErr:       false,
		},
		{
			name:          "Invalid NFe key",
			chaveNFe:      "invalid",
			justificativa: "Produto entregue com defeito, solicitado pelo cliente",
			protocolo:     "135220000000123",
			wantErr:       true,
		},
		{
			name:          "Invalid justification",
			chaveNFe:      "35220214200166000187550010000000101234567890",
			justificativa: "short",
			protocolo:     "135220000000123",
			wantErr:       true,
		},
		{
			name:          "Empty protocol",
			chaveNFe:      "35220214200166000187550010000000101234567890",
			justificativa: "Produto entregue com defeito, solicitado pelo cliente",
			protocolo:     "",
			wantErr:       true,
		},
		{
			name:          "Whitespace in fields",
			chaveNFe:      "  35220214200166000187550010000000101234567890  ",
			justificativa: "  Produto entregue com defeito, solicitado pelo cliente  ",
			protocolo:     "  135220000000123  ",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := CreateCancelamentoRequest(tt.chaveNFe, tt.justificativa, tt.protocolo)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCancelamentoRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && req == nil {
				t.Errorf("CreateCancelamentoRequest() returned nil request but no error")
			}
			if !tt.wantErr && req != nil {
				// Validate that fields are properly trimmed and set
				if strings.TrimSpace(req.ChaveNFe) != req.ChaveNFe {
					t.Errorf("ChaveNFe not properly trimmed: %q", req.ChaveNFe)
				}
				if strings.TrimSpace(req.Protocolo) != req.Protocolo {
					t.Errorf("Protocolo not properly trimmed: %q", req.Protocolo)
				}
			}
		})
	}
}

func TestValidarCancelamento(t *testing.T) {
	validReq := &CancelamentoRequest{
		ChaveNFe:      "35220214200166000187550010000000101234567890",
		Justificativa: "Produto entregue com defeito, solicitado pelo cliente",
		Protocolo:     "135220000000123",
	}

	tests := []struct {
		name    string
		req     *CancelamentoRequest
		wantErr bool
	}{
		{
			name:    "Valid request",
			req:     validReq,
			wantErr: false,
		},
		{
			name:    "Nil request",
			req:     nil,
			wantErr: true,
		},
		{
			name: "Invalid NFe key",
			req: &CancelamentoRequest{
				ChaveNFe:      "invalid",
				Justificativa: "Produto entregue com defeito, solicitado pelo cliente",
				Protocolo:     "135220000000123",
			},
			wantErr: true,
		},
		{
			name: "Invalid justification",
			req: &CancelamentoRequest{
				ChaveNFe:      "35220214200166000187550010000000101234567890",
				Justificativa: "short",
				Protocolo:     "135220000000123",
			},
			wantErr: true,
		},
		{
			name: "Empty protocol",
			req: &CancelamentoRequest{
				ChaveNFe:      "35220214200166000187550010000000101234567890",
				Justificativa: "Produto entregue com defeito, solicitado pelo cliente",
				Protocolo:     "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidarCancelamento(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidarCancelamento() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCancellationStatusText(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   string
	}{
		{
			name:   "Registered status",
			status: int(CancellationStatusRegistered),
			want:   "Evento registrado e vinculado à NFe",
		},
		{
			name:   "Already exists status",
			status: int(CancellationStatusAlreadyExists),
			want:   "Evento de cancelamento já existe para esta NFe",
		},
		{
			name:   "Approved status",
			status: int(CancellationStatusApproved),
			want:   "Cancelamento homologado",
		},
		{
			name:   "Outside deadline status",
			status: int(CancellationStatusOutsideDeadline),
			want:   "Evento rejeitado - fora do prazo de cancelamento",
		},
		{
			name:   "NFe not found status",
			status: int(CancellationStatusNFeNotFound),
			want:   "Evento rejeitado - NFe não encontrada",
		},
		{
			name:   "Unknown status",
			status: 999,
			want:   "Status desconhecido: 999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCancellationStatusText(tt.status)
			if got != tt.want {
				t.Errorf("GetCancellationStatusText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsCancellationSuccessful(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{
			name:   "Registered status - success",
			status: int(CancellationStatusRegistered),
			want:   true,
		},
		{
			name:   "Approved status - success",
			status: int(CancellationStatusApproved),
			want:   true,
		},
		{
			name:   "Already exists status - not success",
			status: int(CancellationStatusAlreadyExists),
			want:   false,
		},
		{
			name:   "Outside deadline status - not success",
			status: int(CancellationStatusOutsideDeadline),
			want:   false,
		},
		{
			name:   "NFe not found status - not success",
			status: int(CancellationStatusNFeNotFound),
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
			got := IsCancellationSuccessful(tt.status)
			if got != tt.want {
				t.Errorf("IsCancellationSuccessful() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateUFCode(t *testing.T) {
	tests := []struct {
		name   string
		ufCode string
		want   bool
	}{
		{"Valid SP", "35", true},
		{"Valid RJ", "33", true},
		{"Valid MG", "31", true},
		{"Valid CE", "85", true},
		{"Invalid 99", "99", false},
		{"Invalid 00", "00", false},
		{"Invalid XX", "XX", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUFCode(tt.ufCode)
			got := err == nil
			if got != tt.want {
				t.Errorf("validateUFCode() = %v, want %v, error: %v", got, tt.want, err)
			}
		})
	}
}

func TestValidateDateInKey(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		want    bool
	}{
		{"Valid date 2212", "2212", true},
		{"Valid date 2201", "2201", true},
		{"Invalid month 2213", "2213", false},
		{"Invalid month 2200", "2200", false},
		{"Too short", "221", false},
		{"Too long", "22121", false},
		{"Non-numeric", "22AB", false},
		{"Empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDateInKey(tt.dateStr)
			got := err == nil
			if got != tt.want {
				t.Errorf("validateDateInKey() = %v, want %v, error: %v", got, tt.want, err)
			}
		})
	}
}

func TestHasValidContent(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "Valid content with enough alphanumeric",
			text: "Produto com defeito solicitado pelo cliente",
			want: true,
		},
		{
			name: "Minimal valid content",
			text: "Erro na emissao NFe",
			want: true,
		},
		{
			name: "Not enough alphanumeric characters",
			text: "!@# $%^ &*()",
			want: false,
		},
		{
			name: "Only spaces",
			text: "          ",
			want: false,
		},
		{
			name: "Empty text",
			text: "",
			want: false,
		},
		{
			name: "Mixed with enough alphanumeric",
			text: "Produto123 com defeito!!!",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasValidContent(tt.text)
			if got != tt.want {
				t.Errorf("hasValidContent() = %v, want %v", got, tt.want)
			}
		})
	}
}