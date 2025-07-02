package nfe

import (
	"testing"
)

func TestNew(t *testing.T) {
	config := Config{
		Environment: Homologation,
		UF:          SP,
		Timeout:     30,
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.config.Environment != Homologation {
		t.Errorf("Expected environment %d, got %d", Homologation, client.config.Environment)
	}
}

func TestNewWithDefaultTimeout(t *testing.T) {
	config := Config{
		Environment: Production,
		UF:          RJ,
		Timeout:     0, // Should default to 30
	}

	client, err := New(config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if client.config.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", client.config.Timeout)
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version != Version {
		t.Errorf("Expected version %s, got %s", Version, version)
	}

	if version == "" {
		t.Error("Version should not be empty")
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment Environment
		expected    bool
	}{
		{"Production environment", Production, true},
		{"Homologation environment", Homologation, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Environment: tt.environment, UF: SP}
			client, _ := New(config)

			if client.IsProduction() != tt.expected {
				t.Errorf("Expected IsProduction() = %v, got %v", tt.expected, client.IsProduction())
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid production config",
			config: Config{
				Environment: Production,
				UF:          SP,
				Timeout:     30,
			},
			wantErr: false,
		},
		{
			name: "Valid homologation config",
			config: Config{
				Environment: Homologation,
				UF:          RJ,
				Timeout:     60,
			},
			wantErr: false,
		},
		{
			name: "Invalid environment",
			config: Config{
				Environment: 999,
				UF:          SP,
				Timeout:     30,
			},
			wantErr: true,
		},
		{
			name: "Invalid UF",
			config: Config{
				Environment: Production,
				UF:          0,
				Timeout:     30,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateAccessKey(t *testing.T) {
	config := Config{
		Environment: Homologation,
		UF:          SP,
		Timeout:     30,
	}

	client, _ := New(config)

	cnpj := "12345678000190"
	modelo := 55
	serie := 1
	numero := 123456
	tipoEmissao := 1

	accessKey := client.GenerateAccessKey(cnpj, modelo, serie, numero, tipoEmissao)

	// Access key should be 44 characters
	if len(accessKey) != 44 {
		t.Errorf("Expected access key length 44, got %d", len(accessKey))
	}

	// Should start with UF code (35 for SP)
	if accessKey[:2] != "35" {
		t.Errorf("Expected access key to start with '35', got '%s'", accessKey[:2])
	}
}

func BenchmarkNew(b *testing.B) {
	config := Config{
		Environment: Homologation,
		UF:          SP,
		Timeout:     30,
	}

	for i := 0; i < b.N; i++ {
		_, _ = New(config)
	}
}

func BenchmarkGenerateAccessKey(b *testing.B) {
	config := Config{Environment: Homologation, UF: SP}
	client, _ := New(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.GenerateAccessKey("12345678000190", 55, 1, i+1, 1)
	}
}
