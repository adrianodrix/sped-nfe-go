package factories

import (
	"strings"
	"testing"
)

func TestNewQRCode(t *testing.T) {
	tests := []struct {
		name     string
		config   QRCodeConfig
		expected QRCode
	}{
		{
			name: "default version",
			config: QRCodeConfig{
				CSC:   "123456",
				CSCId: "000001",
			},
			expected: QRCode{
				Version: "2.00",
				CSC:     "123456",
				CSCId:   "000001",
			},
		},
		{
			name: "explicit version 1.00",
			config: QRCodeConfig{
				Version: "1.00",
				CSC:     "123456",
				CSCId:   "000001",
			},
			expected: QRCode{
				Version: "1.00",
				CSC:     "123456",
				CSCId:   "000001",
			},
		},
		{
			name: "version 3.00",
			config: QRCodeConfig{
				Version: "3.00",
				CSC:     "123456",
				CSCId:   "000001",
			},
			expected: QRCode{
				Version: "3.00",
				CSC:     "123456",
				CSCId:   "000001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qr := NewQRCode(tt.config)
			if qr.Version != tt.expected.Version {
				t.Errorf("Expected version %s, got %s", tt.expected.Version, qr.Version)
			}
			if qr.CSC != tt.expected.CSC {
				t.Errorf("Expected CSC %s, got %s", tt.expected.CSC, qr.CSC)
			}
			if qr.CSCId != tt.expected.CSCId {
				t.Errorf("Expected CSCId %s, got %s", tt.expected.CSCId, qr.CSCId)
			}
		})
	}
}

func TestQRCode_str2Hex(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple text",
			input:    "test",
			expected: "74657374",
		},
		{
			name:     "date time",
			input:    "2023-12-25T15:30:00-03:00",
			expected: "323032332d31322d32355431353a33303a30302d30333a3030",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qr.str2Hex(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestQRCode_generateSHA1Hash(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "DA39A3EE5E6B4B0D3255BFEF95601890AFD80709",
		},
		{
			name:     "test string",
			input:    "test",
			expected: "A94A8FE5CCB19BA61C4C0873D391E987982FBBD3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := qr.generateSHA1Hash(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestQRCode_generate100(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{Version: "1.00"})

	chNFe := "41230714200166000187650010000000051123456789"
	url := "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx"
	tpAmb := "2"
	dhEmi := "2023-12-25T15:30:00-03:00"
	vNF := "150.00"
	digVal := "testDigest"

	result := qr.generate100(chNFe, url, tpAmb, dhEmi, vNF, digVal)

	expected := url + "?chNFe=" + chNFe + "&nVersao=100&tpAmb=" + tpAmb +
		"&dhEmi=" + qr.str2Hex(dhEmi) + "&vNF=" + vNF + "&digVal=" + qr.str2Hex(digVal)

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Verify URL contains required parameters
	if !strings.Contains(result, "nVersao=100") {
		t.Error("QR code should contain nVersao=100")
	}
	if !strings.Contains(result, "chNFe="+chNFe) {
		t.Error("QR code should contain chNFe")
	}
	if !strings.Contains(result, "tpAmb="+tpAmb) {
		t.Error("QR code should contain tpAmb")
	}
}

func TestQRCode_generate200(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{Version: "2.00"})

	chNFe := "41230714200166000187650010000000051123456789"
	url := "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx"
	tpAmb := "2"
	dhEmi := "2023-12-25T15:30:00-03:00"
	vNF := "150.00"
	vICMS := "18.00"
	digVal := "testDigest"
	token := "testToken"
	idToken := "000001"
	versao := "4.00"
	tpEmis := 1
	cDest := "12345678901"

	result := qr.generate200(chNFe, url, tpAmb, dhEmi, vNF, vICMS, digVal, token, idToken, versao, tpEmis, cDest)

	// Verify URL contains required parameters
	if !strings.Contains(result, "nVersao=200") {
		t.Error("QR code should contain nVersao=200")
	}
	if !strings.Contains(result, "chNFe="+chNFe) {
		t.Error("QR code should contain chNFe")
	}
	if !strings.Contains(result, "cHashQRCode=") {
		t.Error("QR code should contain cHashQRCode")
	}
	if !strings.Contains(result, "cIdToken="+idToken) {
		t.Error("QR code should contain cIdToken")
	}
}

func TestQRCode_generate300(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{Version: "3.00"})

	chNFe := "41230714200166000187650010000000051123456789"
	url := "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx"
	tpAmb := "2"
	dhEmi := "2023-12-25T15:30:00-03:00"
	vNF := "150.00"
	tpEmis := 1
	idDest := 1
	cDest := "12345678901"
	assinatura := "testSignature"

	result := qr.generate300(chNFe, url, tpAmb, dhEmi, vNF, tpEmis, idDest, cDest, assinatura)

	expected := url + "?chNFe=" + chNFe + "&nVersao=300&tpAmb=" + tpAmb +
		"&idDest=1&cDest=" + cDest + "&dhEmi=" + qr.str2Hex(dhEmi) +
		"&vNF=" + vNF + "&tpEmis=1&assinatura=" + assinatura

	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// Verify URL contains required parameters
	if !strings.Contains(result, "nVersao=300") {
		t.Error("QR code should contain nVersao=300")
	}
	if !strings.Contains(result, "assinatura="+assinatura) {
		t.Error("QR code should contain assinatura")
	}
}

func TestQRCode_extractChaveNFe(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name        string
		xml         string
		expected    string
		expectError bool
	}{
		{
			name:        "valid chave",
			xml:         `<infNFe Id="NFe41230714200166000187650010000000051123456789">`,
			expected:    "41230714200166000187650010000000051123456789",
			expectError: false,
		},
		{
			name:        "invalid xml",
			xml:         `<infNFe>`,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qr.extractChaveNFe(tt.xml)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestQRCode_extractTpAmb(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name        string
		xml         string
		expected    string
		expectError bool
	}{
		{
			name:        "production environment",
			xml:         `<tpAmb>1</tpAmb>`,
			expected:    "1",
			expectError: false,
		},
		{
			name:        "homologation environment",
			xml:         `<tpAmb>2</tpAmb>`,
			expected:    "2",
			expectError: false,
		},
		{
			name:        "missing tpAmb",
			xml:         `<other>test</other>`,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qr.extractTpAmb(tt.xml)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestQRCode_extractCDest(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name     string
		xml      string
		expected string
	}{
		{
			name:     "CNPJ destination",
			xml:      `<dest><CNPJ>12345678000195</CNPJ></dest>`,
			expected: "12345678000195",
		},
		{
			name:     "CPF destination",
			xml:      `<dest><CPF>12345678901</CPF></dest>`,
			expected: "12345678901",
		},
		{
			name:     "no destination",
			xml:      `<other>test</other>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := qr.extractCDest(tt.xml)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestQRCode_ValidateQRCode(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name        string
		qrCode      string
		expectError bool
	}{
		{
			name:        "empty QR code",
			qrCode:      "",
			expectError: true,
		},
		{
			name:        "invalid URL",
			qrCode:      "invalid-url",
			expectError: true,
		},
		{
			name:        "valid v1.00 QR code",
			qrCode:      "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx?chNFe=41230714200166000187650010000000051123456789&nVersao=100&tpAmb=2&dhEmi=test&vNF=150.00&digVal=test",
			expectError: false,
		},
		{
			name:        "invalid v1.00 QR code - missing parameter",
			qrCode:      "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx?chNFe=41230714200166000187650010000000051123456789&nVersao=100&tpAmb=2",
			expectError: true,
		},
		{
			name:        "valid v2.00 QR code",
			qrCode:      "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx?chNFe=41230714200166000187650010000000051123456789&nVersao=200&tpAmb=2&cDest=&dhEmi=test&vNF=150.00&vICMS=18.00&digVal=test&cIdToken=000001&cHashQRCode=ABCD1234",
			expectError: false,
		},
		{
			name:        "valid v3.00 QR code",
			qrCode:      "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx?chNFe=41230714200166000187650010000000051123456789&nVersao=300&tpAmb=2&dhEmi=test&vNF=150.00&tpEmis=1&assinatura=testSig",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := qr.ValidateQRCode(tt.qrCode)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestGetStateConsultationURL(t *testing.T) {
	tests := []struct {
		name        string
		uf          string
		environment int
		expected    string
	}{
		{
			name:        "SP production",
			uf:          "SP",
			environment: 1,
			expected:    "https://www.fazenda.sp.gov.br/nfce/qrcode",
		},
		{
			name:        "SP homologation",
			uf:          "SP",
			environment: 2,
			expected:    "https://www.homologacao.fazenda.sp.gov.br/nfce/qrcode",
		},
		{
			name:        "unknown state production",
			uf:          "XX",
			environment: 1,
			expected:    "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx",
		},
		{
			name:        "unknown state homologation",
			uf:          "XX",
			environment: 2,
			expected:    "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStateConsultationURL(tt.uf, tt.environment)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestQRCode_insertQRCodeIntoXML(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name        string
		xml         string
		qrCodeData  string
		uriChave    string
		expectError bool
	}{
		{
			name: "insert new QR code",
			xml: `<NFe>
				<infNFe>
					<ide>test</ide>
				</infNFe>
			</NFe>`,
			qrCodeData:  "https://test.com/qr",
			uriChave:    "https://test.com/consulta",
			expectError: false,
		},
		{
			name: "replace existing QR code",
			xml: `<NFe>
				<infNFe>
					<ide>test</ide>
					<infNFeSupl>
						<qrCode><![CDATA[old-qr-code]]></qrCode>
						<urlChave>old-url</urlChave>
					</infNFeSupl>
				</infNFe>
			</NFe>`,
			qrCodeData:  "https://test.com/qr",
			uriChave:    "https://test.com/consulta",
			expectError: false,
		},
		{
			name: "invalid XML - no infNFe",
			xml: `<NFe>
				<other>test</other>
			</NFe>`,
			qrCodeData:  "https://test.com/qr",
			uriChave:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qr.insertQRCodeIntoXML(tt.xml, tt.qrCodeData, tt.uriChave)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError {
				if !strings.Contains(result, tt.qrCodeData) {
					t.Error("Result should contain QR code data")
				}
				if tt.uriChave != "" && !strings.Contains(result, tt.uriChave) {
					t.Error("Result should contain URI chave")
				}
			}
		})
	}
}

func TestQRCodeBuilder(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{Version: "2.00"})
	builder := NewQRCodeBuilder(qr)

	// Test fluent interface
	result, err := builder.
		ChaveNFe("41230714200166000187650010000000051123456789").
		URL("https://test.com").
		Environment("2").
		EmissionDateTime("2023-12-25T15:30:00-03:00").
		TotalValue("150.00").
		ICMSValue("18.00").
		DigestValue("testDigest").
		Token("testToken", "000001").
		Build()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "chNFe=41230714200166000187650010000000051123456789") {
		t.Error("Result should contain chNFe")
	}
	if !strings.Contains(result, "nVersao=200") {
		t.Error("Result should contain nVersao=200")
	}

	// Test error cases
	emptyBuilder := NewQRCodeBuilder(qr)
	_, err = emptyBuilder.Build()
	if err == nil {
		t.Error("Expected error for empty builder")
	}
}

func TestQRCode_GetQRCodeFromXML(t *testing.T) {
	qr := NewQRCode(QRCodeConfig{})

	tests := []struct {
		name        string
		xml         string
		expected    string
		expectError bool
	}{
		{
			name: "valid QR code in XML",
			xml: `<infNFeSupl>
				<qrCode><![CDATA[https://test.com/qr?param=value]]></qrCode>
			</infNFeSupl>`,
			expected:    "https://test.com/qr?param=value",
			expectError: false,
		},
		{
			name:        "no QR code in XML",
			xml:         `<other>test</other>`,
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := qr.GetQRCodeFromXML(tt.xml)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Benchmark tests
func BenchmarkQRCode_generate200(b *testing.B) {
	qr := NewQRCode(QRCodeConfig{Version: "2.00"})
	chNFe := "41230714200166000187650010000000051123456789"
	url := "https://www.sefaz.rs.gov.br/NFCE/NFCE-COM.aspx"
	tpAmb := "2"
	dhEmi := "2023-12-25T15:30:00-03:00"
	vNF := "150.00"
	vICMS := "18.00"
	digVal := "testDigest"
	token := "testToken"
	idToken := "000001"
	versao := "4.00"
	tpEmis := 1
	cDest := "12345678901"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qr.generate200(chNFe, url, tpAmb, dhEmi, vNF, vICMS, digVal, token, idToken, versao, tpEmis, cDest)
	}
}

func BenchmarkQRCode_extractChaveNFe(b *testing.B) {
	qr := NewQRCode(QRCodeConfig{})
	xml := `<infNFe Id="NFe41230714200166000187650010000000051123456789">`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qr.extractChaveNFe(xml)
	}
}
