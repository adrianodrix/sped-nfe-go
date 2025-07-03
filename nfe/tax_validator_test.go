package nfe

import (
	"testing"
)

func TestNewTaxValidator(t *testing.T) {
	tests := []struct {
		name   string
		config *ValidationConfig
		want   *ValidationConfig
	}{
		{
			name:   "with nil config",
			config: nil,
			want: &ValidationConfig{
				UF:               "SP",
				Environment:      "HOMOLOGACAO",
				Version:          "4.00",
				StrictValidation: true,
			},
		},
		{
			name: "with custom config",
			config: &ValidationConfig{
				UF:          "RJ",
				Environment: "PRODUCAO",
			},
			want: &ValidationConfig{
				UF:          "RJ",
				Environment: "PRODUCAO",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewTaxValidator(tt.config)

			if validator == nil {
				t.Errorf("NewTaxValidator() returned nil")
				return
			}

			if tt.name == "with nil config" {
				// Check default values were set
				if validator.config.UF != "SP" {
					t.Errorf("Expected UF = SP, got %s", validator.config.UF)
				}
				if validator.config.Environment != "HOMOLOGACAO" {
					t.Errorf("Expected Environment = HOMOLOGACAO, got %s", validator.config.Environment)
				}
			}
		})
	}
}

func TestTaxValidator_ValidateNCM(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name    string
		ncm     string
		wantErr bool
	}{
		{
			name:    "valid NCM",
			ncm:     "12345678",
			wantErr: false,
		},
		{
			name:    "too short",
			ncm:     "1234567",
			wantErr: true,
		},
		{
			name:    "too long",
			ncm:     "123456789",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			ncm:     "1234567a",
			wantErr: true,
		},
		{
			name:    "out of range",
			ncm:     "00000000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateNCM(tt.ncm)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateNCM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxValidator_ValidateCFOP(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name    string
		cfop    string
		wantErr bool
	}{
		{
			name:    "valid entry CFOP",
			cfop:    "1102",
			wantErr: false,
		},
		{
			name:    "valid exit CFOP",
			cfop:    "5102",
			wantErr: false,
		},
		{
			name:    "too short",
			cfop:    "510",
			wantErr: true,
		},
		{
			name:    "too long",
			cfop:    "51022",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			cfop:    "510a",
			wantErr: true,
		},
		{
			name:    "invalid first digit",
			cfop:    "4102",
			wantErr: true,
		},
		{
			name:    "out of range for entries",
			cfop:    "1000",
			wantErr: false, // 1000 is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateCFOP(tt.cfop)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateCFOP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxValidator_ValidateCEST(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name    string
		cest    string
		wantErr bool
	}{
		{
			name:    "valid CEST",
			cest:    "0100100",
			wantErr: false,
		},
		{
			name:    "too short",
			cest:    "010010",
			wantErr: true,
		},
		{
			name:    "too long",
			cest:    "01001000",
			wantErr: true,
		},
		{
			name:    "non-numeric",
			cest:    "010010a",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateCEST(tt.cest)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateCEST() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxValidator_ValidateOrigin(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name    string
		origin  string
		wantErr bool
	}{
		{
			name:    "valid origin 0",
			origin:  "0",
			wantErr: false,
		},
		{
			name:    "valid origin 1",
			origin:  "1",
			wantErr: false,
		},
		{
			name:    "valid origin 8",
			origin:  "8",
			wantErr: false,
		},
		{
			name:    "invalid origin 9",
			origin:  "9",
			wantErr: true,
		},
		{
			name:    "invalid origin a",
			origin:  "a",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateOrigin(tt.origin)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateOrigin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxValidator_IsValidModBC(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name  string
		modBC string
		want  bool
	}{
		{
			name:  "valid modBC 0",
			modBC: "0",
			want:  true,
		},
		{
			name:  "valid modBC 3",
			modBC: "3",
			want:  true,
		},
		{
			name:  "invalid modBC 4",
			modBC: "4",
			want:  false,
		},
		{
			name:  "invalid modBC a",
			modBC: "a",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isValidModBC(tt.modBC)
			if got != tt.want {
				t.Errorf("isValidModBC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaxValidator_ValidateICMS00(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name      string
		icms00    *ICMS00
		wantCount int
	}{
		{
			name: "valid ICMS00",
			icms00: &ICMS00{
				Orig:  "0",
				CST:   "00",
				ModBC: "0",
				VBC:   "100.00",
				PICMS: "18.00",
				VICMS: "18.00",
			},
			wantCount: 0,
		},
		{
			name: "invalid CST",
			icms00: &ICMS00{
				Orig:  "0",
				CST:   "10",
				ModBC: "0",
				VBC:   "100.00",
				PICMS: "18.00",
				VICMS: "18.00",
			},
			wantCount: 1,
		},
		{
			name: "invalid origin",
			icms00: &ICMS00{
				Orig:  "9",
				CST:   "00",
				ModBC: "0",
				VBC:   "100.00",
				PICMS: "18.00",
				VICMS: "18.00",
			},
			wantCount: 1,
		},
		{
			name: "calculation error",
			icms00: &ICMS00{
				Orig:  "0",
				CST:   "00",
				ModBC: "0",
				VBC:   "100.00",
				PICMS: "18.00",
				VICMS: "20.00", // Wrong calculation
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.validateICMS00(tt.icms00, Produto{})

			if len(errors) != tt.wantCount {
				t.Errorf("validateICMS00() returned %d errors, want %d", len(errors), tt.wantCount)
				for _, err := range errors {
					t.Logf("Error: %s - %s", err.Code, err.Message)
				}
			}
		})
	}
}

func TestTaxValidator_ValidateItemTaxes(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name      string
		item      *Item
		wantCount int
	}{
		{
			name: "valid item",
			item: &Item{
				Prod: Produto{
					NCM:  "12345678",
					CFOP: "5102",
				},
				Imposto: Imposto{
					ICMS: &ICMS{
						ICMS00: &ICMS00{
							Orig:  "0",
							CST:   "00",
							ModBC: "0",
							VBC:   "100.00",
							PICMS: "18.00",
							VICMS: "18.00",
						},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "invalid NCM",
			item: &Item{
				Prod: Produto{
					NCM:  "1234567", // Too short
					CFOP: "5102",
				},
				Imposto: Imposto{
					ICMS: &ICMS{
						ICMS00: &ICMS00{
							Orig:  "0",
							CST:   "00",
							ModBC: "0",
							VBC:   "100.00",
							PICMS: "18.00",
							VICMS: "18.00",
						},
					},
				},
			},
			wantCount: 1, // NCM error
		},
		{
			name: "missing ICMS",
			item: &Item{
				Prod: Produto{
					NCM:  "12345678",
					CFOP: "5102",
				},
				Imposto: Imposto{
					// No ICMS
				},
			},
			wantCount: 1, // Missing ICMS error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateItemTaxes(tt.item)

			if len(errors) != tt.wantCount {
				t.Errorf("ValidateItemTaxes() returned %d errors, want %d", len(errors), tt.wantCount)
				for _, err := range errors {
					t.Logf("Error: %s - %s", err.Code, err.Message)
				}
			}
		})
	}
}

func TestTaxValidator_CountActiveICMSModalities(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name string
		icms *ICMS
		want int
	}{
		{
			name: "no modalities",
			icms: &ICMS{},
			want: 0,
		},
		{
			name: "one modality",
			icms: &ICMS{
				ICMS00: &ICMS00{},
			},
			want: 1,
		},
		{
			name: "two modalities",
			icms: &ICMS{
				ICMS00: &ICMS00{},
				ICMS10: &ICMS10{},
			},
			want: 2,
		},
		{
			name: "simples nacional",
			icms: &ICMS{
				ICMSSN102: &ICMSSN102{},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.countActiveICMSModalities(tt.icms)
			if got != tt.want {
				t.Errorf("countActiveICMSModalities() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaxValidator_ValidateCrossTaxRules(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name      string
		item      *Item
		wantCount int
	}{
		{
			name: "consistent PIS/COFINS",
			item: &Item{
				Imposto: Imposto{
					PIS: &PIS{
						PISAliq: &PISAliq{},
					},
					COFINS: &COFINS{
						COFINSAliq: &COFINSAliq{},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "inconsistent PIS/COFINS",
			item: &Item{
				Imposto: Imposto{
					PIS: &PIS{
						PISAliq: &PISAliq{}, // Taxed
					},
					COFINS: &COFINS{
						COFINSNT: &COFINSNT{}, // Non-taxed
					},
				},
			},
			wantCount: 1, // Warning expected
		},
		{
			name: "ISSQN with ICMS conflict",
			item: &Item{
				Imposto: Imposto{
					ICMS: &ICMS{
						ICMS00: &ICMS00{}, // Taxed ICMS
					},
					ISSQN: &ISSQN{}, // ISSQN present
				},
			},
			wantCount: 1, // Warning expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.validateCrossTaxRules(tt.item)

			if len(errors) != tt.wantCount {
				t.Errorf("validateCrossTaxRules() returned %d errors, want %d", len(errors), tt.wantCount)
				for _, err := range errors {
					t.Logf("Error: %s - %s", err.Code, err.Message)
				}
			}
		})
	}
}

func TestTaxValidator_IsPISTaxed(t *testing.T) {
	validator := NewTaxValidator(nil)

	tests := []struct {
		name string
		pis  *PIS
		want bool
	}{
		{
			name: "nil PIS",
			pis:  nil,
			want: false,
		},
		{
			name: "PIS Aliq",
			pis: &PIS{
				PISAliq: &PISAliq{},
			},
			want: true,
		},
		{
			name: "PIS NT",
			pis: &PIS{
				PISNT: &PISNT{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.isPISTaxed(tt.pis)
			if got != tt.want {
				t.Errorf("isPISTaxed() = %v, want %v", got, tt.want)
			}
		})
	}
}
