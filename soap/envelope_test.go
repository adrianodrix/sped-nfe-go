package soap

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestNewEnvelopeBuilder(t *testing.T) {
	builder := NewEnvelopeBuilder(SOAP11)
	
	if builder == nil {
		t.Fatal("NewEnvelopeBuilder should not return nil")
	}
	
	if builder.version != SOAP11 {
		t.Errorf("Expected version %s, got %s", SOAP11, builder.version)
	}
	
	if builder.headers == nil {
		t.Error("Headers should be initialized")
	}
	
	if builder.namespaces == nil {
		t.Error("Namespaces should be initialized")
	}
}

func TestNewSOAP11EnvelopeBuilder(t *testing.T) {
	builder := NewSOAP11EnvelopeBuilder()
	
	if builder == nil {
		t.Fatal("NewSOAP11EnvelopeBuilder should not return nil")
	}
	
	if builder.version != SOAP11 {
		t.Errorf("Expected version %s, got %s", SOAP11, builder.version)
	}
	
	// Check default namespaces
	if builder.namespaces["soap"] != SOAP11EnvelopeNS {
		t.Errorf("Expected soap namespace %s, got %s", SOAP11EnvelopeNS, builder.namespaces["soap"])
	}
	
	if builder.namespaces["xsi"] != XMLSchemaInstanceNS {
		t.Errorf("Expected xsi namespace %s, got %s", XMLSchemaInstanceNS, builder.namespaces["xsi"])
	}
}

func TestNewSOAP12EnvelopeBuilder(t *testing.T) {
	builder := NewSOAP12EnvelopeBuilder()
	
	if builder == nil {
		t.Fatal("NewSOAP12EnvelopeBuilder should not return nil")
	}
	
	if builder.version != SOAP12 {
		t.Errorf("Expected version %s, got %s", SOAP12, builder.version)
	}
	
	// Check default namespaces
	if builder.namespaces["soap"] != SOAP12EnvelopeNS {
		t.Errorf("Expected soap namespace %s, got %s", SOAP12EnvelopeNS, builder.namespaces["soap"])
	}
}

func TestEnvelopeBuilderMethods(t *testing.T) {
	builder := NewSOAP11EnvelopeBuilder()
	
	// Test SetBodyContent
	bodyContent := "<testBody>content</testBody>"
	builder.SetBodyContent(bodyContent)
	if builder.bodyContent != bodyContent {
		t.Errorf("Expected body content %s, got %s", bodyContent, builder.bodyContent)
	}
	
	// Test AddNamespace
	builder.AddNamespace("test", "http://test.example.com")
	if builder.namespaces["test"] != "http://test.example.com" {
		t.Error("AddNamespace should add namespace")
	}
	
	// Test AddCustomHeader
	headerName := xml.Name{Local: "CustomHeader"}
	headerContent := "<custom>value</custom>"
	builder.AddCustomHeader(headerName, headerContent)
	if len(builder.headers) != 1 {
		t.Error("AddCustomHeader should add header")
	}
	if builder.headers[0].XMLName != headerName {
		t.Error("Header name should match")
	}
	if builder.headers[0].Content != headerContent {
		t.Error("Header content should match")
	}
	
	// Test AddSecurityHeader
	builder.AddSecurityHeader("timestamp-1", 5)
	if builder.security == nil {
		t.Error("AddSecurityHeader should add security header")
	}
	if builder.security.Timestamp == nil {
		t.Error("Security header should have timestamp")
	}
	if builder.security.Timestamp.ID != "timestamp-1" {
		t.Error("Timestamp ID should match")
	}
}

func TestEnvelopeBuilderBuild(t *testing.T) {
	builder := NewSOAP11EnvelopeBuilder()
	
	// Test build without body content - should fail
	_, err := builder.Build()
	if err == nil {
		t.Error("Build without body content should return error")
	}
	
	// Test successful build
	builder.SetBodyContent("<testBody>content</testBody>")
	envelope, err := builder.Build()
	if err != nil {
		t.Errorf("Build should not return error, got: %v", err)
	}
	
	if envelope == nil {
		t.Fatal("Build should return envelope")
	}
	
	if envelope.XmlnsSoap != SOAP11EnvelopeNS {
		t.Errorf("Expected SOAP namespace %s, got %s", SOAP11EnvelopeNS, envelope.XmlnsSoap)
	}
	
	if envelope.Body == nil {
		t.Fatal("Envelope should have body")
	}
	
	if envelope.Body.Content != "<testBody>content</testBody>" {
		t.Error("Body content should match")
	}
}

func TestEnvelopeBuilderToXML(t *testing.T) {
	builder := NewSOAP11EnvelopeBuilder()
	builder.SetBodyContent("<testBody>content</testBody>")
	builder.AddSecurityHeader("timestamp-1", 5)
	
	xmlString, err := builder.ToXML()
	if err != nil {
		t.Errorf("ToXML should not return error, got: %v", err)
	}
	
	if xmlString == "" {
		t.Error("ToXML should return non-empty string")
	}
	
	// Check XML structure
	if !strings.Contains(xmlString, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>") {
		t.Error("XML should contain declaration")
	}
	
	if !strings.Contains(xmlString, "soap:Envelope") {
		t.Error("XML should contain soap:Envelope")
	}
	
	if !strings.Contains(xmlString, "soap:Body") {
		t.Error("XML should contain soap:Body")
	}
	
	if !strings.Contains(xmlString, "testBody") {
		t.Error("XML should contain body content")
	}
	
	if !strings.Contains(xmlString, "wsse:Security") {
		t.Error("XML should contain security header")
	}
}

func TestSOAPEnvelopeToXML(t *testing.T) {
	envelope := &SOAPEnvelope{
		XmlnsSoap: SOAP11EnvelopeNS,
		XmlnsXsi:  XMLSchemaInstanceNS,
		Body: &SOAPBody{
			Content: "<testContent>value</testContent>",
		},
	}
	
	xmlString, err := envelope.ToXML()
	if err != nil {
		t.Errorf("ToXML should not return error, got: %v", err)
	}
	
	if !strings.Contains(xmlString, "testContent") {
		t.Error("XML should contain body content")
	}
}

func TestParseSOAPEnvelope(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
		<soap:Body>
			<testResponse>Success</testResponse>
		</soap:Body>
	</soap:Envelope>`
	
	envelope, err := ParseSOAPEnvelope(xmlData)
	if err != nil {
		t.Errorf("ParseSOAPEnvelope should not return error, got: %v", err)
	}
	
	if envelope == nil {
		t.Fatal("ParseSOAPEnvelope should return envelope")
	}
	
	if envelope.Body == nil {
		t.Fatal("Parsed envelope should have body")
	}
	
	if !strings.Contains(envelope.Body.Content, "testResponse") {
		t.Error("Body should contain testResponse")
	}
}

func TestParseSOAPEnvelopeWithFault(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
		<soap:Body>
			<soap:Fault>
				<faultcode>Client</faultcode>
				<faultstring>Invalid request</faultstring>
				<detail>Missing required parameter</detail>
			</soap:Fault>
		</soap:Body>
	</soap:Envelope>`
	
	envelope, err := ParseSOAPEnvelope(xmlData)
	if err != nil {
		t.Errorf("ParseSOAPEnvelope should not return error, got: %v", err)
	}
	
	if !envelope.HasFault() {
		t.Error("Envelope should have fault")
	}
	
	fault := envelope.GetFault()
	if fault == nil {
		t.Fatal("GetFault should return fault")
	}
	
	if fault.Code != "Client" {
		t.Errorf("Expected fault code 'Client', got %s", fault.Code)
	}
	
	if fault.String != "Invalid request" {
		t.Errorf("Expected fault string 'Invalid request', got %s", fault.String)
	}
}

func TestSOAPEnvelopeMethods(t *testing.T) {
	envelope := &SOAPEnvelope{
		Body: &SOAPBody{
			Content: "<testContent>value</testContent>",
		},
	}
	
	// Test GetBodyContent
	content := envelope.GetBodyContent()
	if content != "<testContent>value</testContent>" {
		t.Error("GetBodyContent should return body content")
	}
	
	// Test SetBodyContent
	newContent := "<newContent>newValue</newContent>"
	envelope.SetBodyContent(newContent)
	if envelope.GetBodyContent() != newContent {
		t.Error("SetBodyContent should update body content")
	}
	
	// Test with nil body
	envelope.Body = nil
	envelope.SetBodyContent(newContent)
	if envelope.Body == nil {
		t.Error("SetBodyContent should create body if nil")
	}
}

func TestTimestampMethods(t *testing.T) {
	builder := NewSOAP11EnvelopeBuilder()
	builder.SetBodyContent("<test/>")
	builder.AddSecurityHeader("timestamp-1", 5)
	
	envelope, err := builder.Build()
	if err != nil {
		t.Fatalf("Build should not return error, got: %v", err)
	}
	
	// Test GetTimestamp
	timestamp := envelope.GetTimestamp()
	if timestamp == nil {
		t.Error("GetTimestamp should return timestamp")
	}
	
	if timestamp.ID != "timestamp-1" {
		t.Errorf("Expected timestamp ID 'timestamp-1', got %s", timestamp.ID)
	}
	
	// Test ValidateTimestamp
	err = envelope.ValidateTimestamp()
	if err != nil {
		t.Errorf("ValidateTimestamp should not return error for valid timestamp, got: %v", err)
	}
}

func TestTimestampValidation(t *testing.T) {
	// Test with expired timestamp
	envelope := &SOAPEnvelope{
		Header: &SOAPHeader{
			Security: &SecurityHeader{
				Timestamp: &Timestamp{
					Created: "2020-01-01T00:00:00.000Z",
					Expires: "2020-01-01T00:05:00.000Z",
				},
			},
		},
	}
	
	err := envelope.ValidateTimestamp()
	if err == nil {
		t.Error("ValidateTimestamp should return error for expired timestamp")
	}
	
	// Test with future timestamp
	future := time.Now().Add(10 * time.Minute).UTC()
	envelope.Header.Security.Timestamp = &Timestamp{
		Created: future.Format("2006-01-02T15:04:05.000Z"),
		Expires: future.Add(5 * time.Minute).Format("2006-01-02T15:04:05.000Z"),
	}
	
	err = envelope.ValidateTimestamp()
	if err == nil {
		t.Error("ValidateTimestamp should return error for future timestamp")
	}
}

func TestCreateNFeSOAPEnvelope(t *testing.T) {
	bodyContent := "<nfeRequest>test</nfeRequest>"
	
	envelope, err := CreateNFeSOAPEnvelope(bodyContent)
	if err != nil {
		t.Errorf("CreateNFeSOAPEnvelope should not return error, got: %v", err)
	}
	
	if envelope == nil {
		t.Fatal("CreateNFeSOAPEnvelope should return envelope")
	}
	
	if envelope.Body.Content != bodyContent {
		t.Error("Body content should match")
	}
	
	if envelope.Header == nil {
		t.Error("NFe envelope should have header")
	}
	
	if envelope.Header.Security == nil {
		t.Error("NFe envelope should have security header")
	}
	
	if envelope.Header.Security.Timestamp == nil {
		t.Error("NFe envelope should have timestamp")
	}
}

func TestCreateNFeSOAPRequest(t *testing.T) {
	url := "https://example.com/nfe"
	action := "nfeAutorizacao"
	bodyContent := "<nfeRequest>test</nfeRequest>"
	
	request, err := CreateNFeSOAPRequest(url, action, bodyContent)
	if err != nil {
		t.Errorf("CreateNFeSOAPRequest should not return error, got: %v", err)
	}
	
	if request == nil {
		t.Fatal("CreateNFeSOAPRequest should return request")
	}
	
	if request.URL != url {
		t.Errorf("Expected URL %s, got %s", url, request.URL)
	}
	
	if request.Action != action {
		t.Errorf("Expected action %s, got %s", action, request.Action)
	}
	
	if request.Body == "" {
		t.Error("Request body should not be empty")
	}
	
	// Verify it's valid XML
	if !strings.Contains(request.Body, "soap:Envelope") {
		t.Error("Request body should contain SOAP envelope")
	}
}

func TestExtractBodyContent(t *testing.T) {
	soapResponse := `<?xml version="1.0" encoding="UTF-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
		<soap:Body>
			<nfeResponse>success</nfeResponse>
		</soap:Body>
	</soap:Envelope>`
	
	content, err := ExtractBodyContent(soapResponse)
	if err != nil {
		t.Errorf("ExtractBodyContent should not return error, got: %v", err)
	}
	
	if !strings.Contains(content, "nfeResponse") {
		t.Error("Extracted content should contain nfeResponse")
	}
}

func TestExtractBodyContentWithFault(t *testing.T) {
	soapResponse := `<?xml version="1.0" encoding="UTF-8"?>
	<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
		<soap:Body>
			<soap:Fault>
				<faultcode>Server</faultcode>
				<faultstring>Internal error</faultstring>
			</soap:Fault>
		</soap:Body>
	</soap:Envelope>`
	
	_, err := ExtractBodyContent(soapResponse)
	if err == nil {
		t.Error("ExtractBodyContent should return error for fault response")
	}
}

func TestIsSOAPFaultResponse(t *testing.T) {
	faultResponse := `<soap:Envelope><soap:Body><soap:Fault><faultcode>Client</faultcode></soap:Fault></soap:Body></soap:Envelope>`
	if !IsSOAPFaultResponse(faultResponse) {
		t.Error("Should detect SOAP fault response")
	}
	
	successResponse := `<soap:Envelope><soap:Body><success/></soap:Body></soap:Envelope>`
	if IsSOAPFaultResponse(successResponse) {
		t.Error("Should not detect fault in success response")
	}
}

func TestGetSOAPVersion(t *testing.T) {
	soap11Content := `<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">`
	if GetSOAPVersion(soap11Content) != SOAP11 {
		t.Error("Should detect SOAP 1.1")
	}
	
	soap12Content := `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">`
	if GetSOAPVersion(soap12Content) != SOAP12 {
		t.Error("Should detect SOAP 1.2")
	}
}

func TestCleanXMLContent(t *testing.T) {
	xmlWithDeclaration := `<?xml version="1.0" encoding="UTF-8"?><root>content</root>`
	cleaned := CleanXMLContent(xmlWithDeclaration)
	if strings.Contains(cleaned, "<?xml") {
		t.Error("Should remove XML declaration")
	}
	
	if !strings.Contains(cleaned, "<root>") {
		t.Error("Should preserve XML content")
	}
}

func TestAddXMLDeclaration(t *testing.T) {
	xmlContent := `<root>content</root>`
	withDeclaration := AddXMLDeclaration(xmlContent)
	
	if !strings.HasPrefix(withDeclaration, "<?xml") {
		t.Error("Should add XML declaration")
	}
	
	if !strings.Contains(withDeclaration, "<root>") {
		t.Error("Should preserve XML content")
	}
}

func TestCreateSOAPFault(t *testing.T) {
	faultCode := "Client"
	faultString := "Invalid request"
	faultDetail := "Missing parameter"
	
	envelope, err := CreateSOAPFault(faultCode, faultString, faultDetail)
	if err != nil {
		t.Errorf("CreateSOAPFault should not return error, got: %v", err)
	}
	
	if envelope == nil {
		t.Fatal("CreateSOAPFault should return envelope")
	}
	
	if !envelope.HasFault() {
		t.Error("Envelope should have fault")
	}
	
	fault := envelope.GetFault()
	if fault.Code != faultCode {
		t.Errorf("Expected fault code %s, got %s", faultCode, fault.Code)
	}
	
	if fault.String != faultString {
		t.Errorf("Expected fault string %s, got %s", faultString, fault.String)
	}
	
	if fault.Detail != faultDetail {
		t.Errorf("Expected fault detail %s, got %s", faultDetail, fault.Detail)
	}
}

func TestValidateSOAPEnvelope(t *testing.T) {
	// Test nil envelope
	err := ValidateSOAPEnvelope(nil)
	if err == nil {
		t.Error("ValidateSOAPEnvelope with nil should return error")
	}
	
	// Test envelope without body
	envelope := &SOAPEnvelope{XmlnsSoap: SOAP11EnvelopeNS}
	err = ValidateSOAPEnvelope(envelope)
	if err == nil {
		t.Error("ValidateSOAPEnvelope without body should return error")
	}
	
	// Test envelope without namespace
	envelope = &SOAPEnvelope{Body: &SOAPBody{}}
	err = ValidateSOAPEnvelope(envelope)
	if err == nil {
		t.Error("ValidateSOAPEnvelope without namespace should return error")
	}
	
	// Test valid envelope
	envelope = &SOAPEnvelope{
		XmlnsSoap: SOAP11EnvelopeNS,
		Body:      &SOAPBody{Content: "<test/>"},
	}
	err = ValidateSOAPEnvelope(envelope)
	if err != nil {
		t.Errorf("ValidateSOAPEnvelope with valid envelope should not return error, got: %v", err)
	}
}