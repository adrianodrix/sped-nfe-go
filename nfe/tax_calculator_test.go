package nfe

import (
	"testing"
)

func TestNewTaxCalculator(t *testing.T) {
	tests := []struct {
		name   string
		config *TaxConfig
		want   *TaxCalculator
	}{
		{
			name:   "with nil config",
			config: nil,
			want: &TaxCalculator{
				config: &TaxConfig{
					ICMSRate:         18.0,
					ICMSSTRate:       18.0,
					ICMSSTMargin:     30.0,
					IPIRate:          0.0,
					PISRate:          1.65,
					COFINSRate:       7.6,
					ISSQNRate:        5.0,
					FederalTaxRegime: "NORMAL",
				},
			},
		},
		{
			name: "with custom config",
			config: &TaxConfig{
				ICMSRate:         12.0,
				IPIRate:          5.0,
				FederalTaxRegime: "SIMPLES",
			},
			want: &TaxCalculator{
				config: &TaxConfig{
					ICMSRate:         12.0,
					IPIRate:          5.0,
					FederalTaxRegime: "SIMPLES",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTaxCalculator(tt.config)

			if got == nil {
				t.Errorf("NewTaxCalculator() returned nil")
				return
			}

			if tt.name == "with nil config" {
				// Check default values were set
				if got.config.ICMSRate != 18.0 {
					t.Errorf("Expected ICMSRate = 18.0, got %f", got.config.ICMSRate)
				}
				if got.config.FederalTaxRegime != "NORMAL" {
					t.Errorf("Expected FederalTaxRegime = NORMAL, got %s", got.config.FederalTaxRegime)
				}
			}
		})
	}
}

func TestTaxCalculator_CalculateItemTaxes(t *testing.T) {
	calculator := NewTaxCalculator(&TaxConfig{
		ICMSRate:         18.0,
		IPIRate:          0.0,
		PISRate:          1.65,
		COFINSRate:       7.6,
		FederalTaxRegime: "NORMAL",
	})

	tests := []struct {
		name    string
		item    *Item
		wantErr bool
	}{
		{
			name:    "nil item",
			item:    nil,
			wantErr: true,
		},
		{
			name: "valid item",
			item: &Item{
				NItem: "1",
				Prod: Produto{
					CProd:    "001",
					CEAN:     "SEM GTIN",
					XProd:    "Produto Teste",
					NCM:      "12345678",
					CFOP:     "5102",
					UCom:     "UN",
					QCom:     "1.00",
					VUnCom:   "100.00",
					VProd:    "100.00",
					CEANTrib: "SEM GTIN",
					UTrib:    "UN",
					QTrib:    "1.00",
					VUnTrib:  "100.00",
				},
				Imposto: Imposto{},
			},
			wantErr: false,
		},
		{
			name: "invalid unit value",
			item: &Item{
				NItem: "1",
				Prod: Produto{
					VUnCom: "invalid",
					QCom:   "1.00",
				},
				Imposto: Imposto{},
			},
			wantErr: true,
		},
		{
			name: "invalid quantity",
			item: &Item{
				NItem: "1",
				Prod: Produto{
					VUnCom: "100.00",
					QCom:   "invalid",
				},
				Imposto: Imposto{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := calculator.CalculateItemTaxes(tt.item)

			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateItemTaxes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.item != nil {
				// Verify taxes were calculated
				if tt.item.Imposto.ICMS == nil {
					t.Error("ICMS not calculated")
				}

				if tt.item.Imposto.IPI == nil {
					t.Error("IPI not calculated")
				}

				if tt.item.Imposto.PIS == nil {
					t.Error("PIS not calculated")
				}

				if tt.item.Imposto.COFINS == nil {
					t.Error("COFINS not calculated")
				}
			}
		})
	}
}

func TestTaxCalculator_CalculateICMS(t *testing.T) {
	calculator := NewTaxCalculator(&TaxConfig{
		ICMSRate: 18.0,
	})

	item := &Item{
		Imposto: Imposto{},
	}

	err := calculator.calculateICMS(item, 100.0)
	if err != nil {
		t.Errorf("calculateICMS() error = %v", err)
		return
	}

	if item.Imposto.ICMS == nil {
		t.Error("ICMS not set")
		return
	}

	if item.Imposto.ICMS.ICMS00 == nil {
		t.Error("ICMS00 not set")
		return
	}

	icms00 := item.Imposto.ICMS.ICMS00
	if icms00.CST != "00" {
		t.Errorf("Expected CST = 00, got %s", icms00.CST)
	}

	if icms00.VBC != "100.00" {
		t.Errorf("Expected VBC = 100.00, got %s", icms00.VBC)
	}

	if icms00.PICMS != "18.00" {
		t.Errorf("Expected PICMS = 18.00, got %s", icms00.PICMS)
	}

	if icms00.VICMS != "18.00" {
		t.Errorf("Expected VICMS = 18.00, got %s", icms00.VICMS)
	}
}

func TestTaxCalculator_CalculateIPI(t *testing.T) {
	tests := []struct {
		name     string
		ipiRate  float64
		wantTrib bool
	}{
		{
			name:     "with IPI rate",
			ipiRate:  5.0,
			wantTrib: true,
		},
		{
			name:     "without IPI rate",
			ipiRate:  0.0,
			wantTrib: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculator := NewTaxCalculator(&TaxConfig{
				IPIRate: tt.ipiRate,
			})

			item := &Item{
				Imposto: Imposto{},
			}

			err := calculator.calculateIPI(item, 100.0)
			if err != nil {
				t.Errorf("calculateIPI() error = %v", err)
				return
			}

			if item.Imposto.IPI == nil {
				t.Error("IPI not set")
				return
			}

			if tt.wantTrib {
				if item.Imposto.IPI.IPITrib == nil {
					t.Error("IPITrib not set when expected")
				}
				if item.Imposto.IPI.IPINT != nil {
					t.Error("IPINT set when not expected")
				}
			} else {
				if item.Imposto.IPI.IPINT == nil {
					t.Error("IPINT not set when expected")
				}
				if item.Imposto.IPI.IPITrib != nil {
					t.Error("IPITrib set when not expected")
				}
			}
		})
	}
}

func TestTaxCalculator_IsService(t *testing.T) {
	calculator := NewTaxCalculator(nil)

	tests := []struct {
		name string
		item *Item
		want bool
	}{
		{
			name: "service NCM",
			item: &Item{
				Prod: Produto{
					NCM:  "00000000",
					CFOP: "5933",
				},
			},
			want: true,
		},
		{
			name: "service CFOP",
			item: &Item{
				Prod: Produto{
					NCM:  "12345678",
					CFOP: "5933", // Service CFOP
				},
			},
			want: true,
		},
		{
			name: "product",
			item: &Item{
				Prod: Produto{
					NCM:  "12345678",
					CFOP: "5102", // Product CFOP
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculator.isService(tt.item)
			if got != tt.want {
				t.Errorf("isService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaxCalculator_CalculateTotalTaxes(t *testing.T) {
	calculator := NewTaxCalculator(&TaxConfig{
		ICMSRate:   18.0,
		IPIRate:    5.0,
		PISRate:    1.65,
		COFINSRate: 7.6,
	})

	items := []Item{
		{
			NItem: "1",
			Prod: Produto{
				QCom:   "1.00",
				VUnCom: "100.00",
			},
			Imposto: Imposto{
				ICMS: &ICMS{
					ICMS00: &ICMS00{
						VICMS: "18.00",
					},
				},
				IPI: &IPI{
					IPITrib: &IPITrib{
						VIPI: "5.00",
					},
				},
				PIS: &PIS{
					PISAliq: &PISAliq{
						VPIS: "1.65",
					},
				},
				COFINS: &COFINS{
					COFINSAliq: &COFINSAliq{
						VCOFINS: "7.60",
					},
				},
			},
		},
	}

	totals, err := calculator.CalculateTotalTaxes(items)
	if err != nil {
		t.Errorf("CalculateTotalTaxes() error = %v", err)
		return
	}

	if totals.TotalICMS != 18.0 {
		t.Errorf("Expected TotalICMS = 18.0, got %f", totals.TotalICMS)
	}

	if totals.TotalIPI != 5.0 {
		t.Errorf("Expected TotalIPI = 5.0, got %f", totals.TotalIPI)
	}

	if totals.TotalPIS != 1.65 {
		t.Errorf("Expected TotalPIS = 1.65, got %f", totals.TotalPIS)
	}

	if totals.TotalCOFINS != 7.6 {
		t.Errorf("Expected TotalCOFINS = 7.6, got %f", totals.TotalCOFINS)
	}
}

func TestParseDecimal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:    "valid decimal",
			input:   "123.45",
			want:    123.45,
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: false,
		},
		{
			name:    "integer",
			input:   "100",
			want:    100.0,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			want:    0,
			wantErr: true,
		},
		{
			name:    "comma as decimal separator",
			input:   "123,45",
			want:    123.45,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDecimal(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseDecimal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("parseDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDecimal(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{
			name:  "two decimal places",
			input: 123.45,
			want:  "123.45",
		},
		{
			name:  "round to two decimal places",
			input: 123.456,
			want:  "123.46",
		},
		{
			name:  "integer",
			input: 100,
			want:  "100.00",
		},
		{
			name:  "zero",
			input: 0,
			want:  "0.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDecimal(tt.input)
			if got != tt.want {
				t.Errorf("formatDecimal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTaxCalculator_ValidateCalculatedTaxes(t *testing.T) {
	calculator := NewTaxCalculator(nil)

	tests := []struct {
		name      string
		item      *Item
		wantCount int
	}{
		{
			name: "valid taxes",
			item: &Item{
				Imposto: Imposto{
					ICMS: &ICMS{
						ICMS00: &ICMS00{
							CST: "00",
						},
					},
					IPI: &IPI{
						CEnq: "999",
						IPITrib: &IPITrib{
							CST: "00",
						},
					},
					PIS: &PIS{
						PISAliq: &PISAliq{
							CST: "01",
						},
					},
					COFINS: &COFINS{
						COFINSAliq: &COFINSAliq{
							CST: "01",
						},
					},
				},
			},
			wantCount: 0, // No errors expected
		},
		{
			name: "invalid ICMS CST",
			item: &Item{
				Imposto: Imposto{
					ICMS: &ICMS{
						ICMS00: &ICMS00{
							CST: "10", // Wrong CST for ICMS00
						},
					},
				},
			},
			wantCount: 1, // One error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := calculator.ValidateCalculatedTaxes(tt.item)

			if len(errors) != tt.wantCount {
				t.Errorf("ValidateCalculatedTaxes() returned %d errors, want %d", len(errors), tt.wantCount)
			}
		})
	}
}
