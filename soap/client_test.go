package soap

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig should not return nil")
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", config.MaxRetries)
	}

	if config.RetryDelay != 1*time.Second {
		t.Errorf("Expected retry delay 1s, got %v", config.RetryDelay)
	}

	if config.UserAgent != "sped-nfe-go/1.0" {
		t.Errorf("Expected user agent 'sped-nfe-go/1.0', got %s", config.UserAgent)
	}

	if config.TLSConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected TLS 1.2, got %x", config.TLSConfig.MinVersion)
	}
}

func TestNewSOAPClient(t *testing.T) {
	// Test with nil config
	client := NewSOAPClient(nil)
	if client == nil {
		t.Fatal("NewSOAPClient should not return nil")
	}

	// Test with custom config
	config := &SOAPClientConfig{
		Timeout:    10 * time.Second,
		MaxRetries: 5,
		RetryDelay: 2 * time.Second,
		UserAgent:  "test-agent",
	}

	client = NewSOAPClient(config)
	if client == nil {
		t.Fatal("NewSOAPClient should not return nil")
	}

	if client.timeout != config.Timeout {
		t.Errorf("Expected timeout %v, got %v", config.Timeout, client.timeout)
	}

	if client.maxRetries != config.MaxRetries {
		t.Errorf("Expected max retries %d, got %d", config.MaxRetries, client.maxRetries)
	}

	if client.retryDelay != config.RetryDelay {
		t.Errorf("Expected retry delay %v, got %v", config.RetryDelay, client.retryDelay)
	}

	if client.userAgent != config.UserAgent {
		t.Errorf("Expected user agent %s, got %s", config.UserAgent, client.userAgent)
	}
}

func TestSOAPClientCall(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Content-Type") != "text/xml; charset=utf-8" {
			t.Errorf("Expected Content-Type 'text/xml; charset=utf-8', got %s", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("SOAPAction") != "\"testAction\"" {
			t.Errorf("Expected SOAPAction '\"testAction\"', got %s", r.Header.Get("SOAPAction"))
		}

		// Return successful response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
		<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<testResponse>Success</testResponse>
			</soap:Body>
		</soap:Envelope>`))
	}))
	defer server.Close()

	client := NewSOAPClient(DefaultConfig())

	request := &SOAPRequest{
		URL:    server.URL,
		Action: "testAction",
		Body: `<?xml version="1.0" encoding="utf-8"?>
		<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
			<soap:Body>
				<testRequest>Test</testRequest>
			</soap:Body>
		</soap:Envelope>`,
	}

	ctx := context.Background()
	response, err := client.Call(ctx, request)

	if err != nil {
		t.Errorf("Call should not return error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Response should not be nil")
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}

	if !strings.Contains(response.Body, "testResponse") {
		t.Error("Response body should contain testResponse")
	}

	if response.Duration <= 0 {
		t.Error("Response duration should be positive")
	}
}

func TestSOAPClientCallWithRetry(t *testing.T) {
	attempts := 0

	// Create mock server that fails first time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<soap:Envelope><soap:Body><success/></soap:Body></soap:Envelope>`))
	}))
	defer server.Close()

	config := DefaultConfig()
	config.MaxRetries = 2
	config.RetryDelay = 10 * time.Millisecond // Fast retry for testing

	client := NewSOAPClient(config)

	request := &SOAPRequest{
		URL:    server.URL,
		Action: "testAction",
		Body:   "<soap:Envelope><soap:Body><test/></soap:Body></soap:Envelope>",
	}

	ctx := context.Background()
	response, err := client.Call(ctx, request)

	if err != nil {
		t.Errorf("Call should succeed after retry, got: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", response.StatusCode)
	}
}

func TestSOAPClientCallValidation(t *testing.T) {
	client := NewSOAPClient(DefaultConfig())
	ctx := context.Background()

	// Test nil request
	_, err := client.Call(ctx, nil)
	if err == nil {
		t.Error("Call with nil request should return error")
	}

	// Test empty URL
	request := &SOAPRequest{
		URL:    "",
		Action: "test",
		Body:   "<test/>",
	}
	_, err = client.Call(ctx, request)
	if err == nil {
		t.Error("Call with empty URL should return error")
	}

	// Test empty body
	request = &SOAPRequest{
		URL:    "http://example.com",
		Action: "test",
		Body:   "",
	}
	_, err = client.Call(ctx, request)
	if err == nil {
		t.Error("Call with empty body should return error")
	}
}

func TestSOAPClientTimeout(t *testing.T) {
	// Create server that takes too long
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Timeout = 50 * time.Millisecond
	config.MaxRetries = 0 // No retries for timeout test

	client := NewSOAPClient(config)

	request := &SOAPRequest{
		URL:    server.URL,
		Action: "test",
		Body:   "<test/>",
	}

	ctx := context.Background()
	_, err := client.Call(ctx, request)

	if err == nil {
		t.Error("Call should timeout and return error")
	}
}

func TestSOAPClientContextCancellation(t *testing.T) {
	// Create server that takes time to respond
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewSOAPClient(DefaultConfig())

	request := &SOAPRequest{
		URL:    server.URL,
		Action: "test",
		Body:   "<test/>",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Call(ctx, request)

	if err == nil {
		t.Error("Call should be cancelled and return error")
	}

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", err)
	}
}

func TestSOAPClientSettings(t *testing.T) {
	client := NewSOAPClient(DefaultConfig())

	// Test SetTimeout
	client.SetTimeout(15 * time.Second)
	if client.GetTimeout() != 15*time.Second {
		t.Errorf("Expected timeout 15s, got %v", client.GetTimeout())
	}

	// Test SetMaxRetries
	client.SetMaxRetries(5)
	if client.GetMaxRetries() != 5 {
		t.Errorf("Expected max retries 5, got %d", client.GetMaxRetries())
	}

	// Test SetRetryDelay
	client.SetRetryDelay(3 * time.Second)
	if client.GetRetryDelay() != 3*time.Second {
		t.Errorf("Expected retry delay 3s, got %v", client.GetRetryDelay())
	}

	// Test EnableLogging
	client.EnableLogging(true)
	if !client.IsLoggingEnabled() {
		t.Error("Logging should be enabled")
	}

	client.EnableLogging(false)
	if client.IsLoggingEnabled() {
		t.Error("Logging should be disabled")
	}

	// Test SetTLSConfig
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}
	client.SetTLSConfig(tlsConfig)
}

func TestValidateRequest(t *testing.T) {
	// Test nil request
	err := ValidateRequest(nil)
	if err == nil {
		t.Error("ValidateRequest with nil should return error")
	}

	// Test empty URL
	request := &SOAPRequest{URL: "", Body: "<test/>"}
	err = ValidateRequest(request)
	if err == nil {
		t.Error("ValidateRequest with empty URL should return error")
	}

	// Test empty body
	request = &SOAPRequest{URL: "http://example.com", Body: ""}
	err = ValidateRequest(request)
	if err == nil {
		t.Error("ValidateRequest with empty body should return error")
	}

	// Test invalid URL
	request = &SOAPRequest{URL: "invalid-url", Body: "<test/>"}
	err = ValidateRequest(request)
	if err == nil {
		t.Error("ValidateRequest with invalid URL should return error")
	}

	// Test valid request
	request = &SOAPRequest{URL: "https://example.com", Body: "<test/>"}
	err = ValidateRequest(request)
	if err != nil {
		t.Errorf("ValidateRequest with valid request should not return error, got: %v", err)
	}
}

func TestCreateSimpleRequest(t *testing.T) {
	url := "https://example.com"
	action := "testAction"
	body := "<test/>"

	request := CreateSimpleRequest(url, action, body)

	if request == nil {
		t.Fatal("CreateSimpleRequest should not return nil")
	}

	if request.URL != url {
		t.Errorf("Expected URL %s, got %s", url, request.URL)
	}

	if request.Action != action {
		t.Errorf("Expected action %s, got %s", action, request.Action)
	}

	if request.Body != body {
		t.Errorf("Expected body %s, got %s", body, request.Body)
	}

	if request.Headers == nil {
		t.Error("Headers should be initialized")
	}
}

func TestSOAPRequestHeaders(t *testing.T) {
	request := CreateSimpleRequest("https://example.com", "test", "<test/>")

	// Test AddHeader
	request.AddHeader("Custom-Header", "value")
	if request.GetHeader("Custom-Header") != "value" {
		t.Error("AddHeader should set header value")
	}

	// Test GetHeader for non-existent header
	if request.GetHeader("Non-Existent") != "" {
		t.Error("GetHeader for non-existent header should return empty string")
	}

	// Test RemoveHeader
	request.RemoveHeader("Custom-Header")
	if request.GetHeader("Custom-Header") != "" {
		t.Error("RemoveHeader should remove header")
	}
}

func TestSOAPResponseMethods(t *testing.T) {
	// Test success response
	response := &SOAPResponse{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"text/xml"}},
		Body:       "<success/>",
		Duration:   100 * time.Millisecond,
	}

	if !response.IsSuccess() {
		t.Error("Status 200 should be success")
	}

	if response.IsClientError() {
		t.Error("Status 200 should not be client error")
	}

	if response.IsServerError() {
		t.Error("Status 200 should not be server error")
	}

	if response.GetHeader("Content-Type") != "text/xml" {
		t.Error("GetHeader should return header value")
	}

	headers := response.GetHeaders("Content-Type")
	if len(headers) != 1 || headers[0] != "text/xml" {
		t.Error("GetHeaders should return header array")
	}

	// Test client error response
	response.StatusCode = 400
	if response.IsSuccess() {
		t.Error("Status 400 should not be success")
	}

	if !response.IsClientError() {
		t.Error("Status 400 should be client error")
	}

	// Test server error response
	response.StatusCode = 500
	if !response.IsServerError() {
		t.Error("Status 500 should be server error")
	}
}

func TestSOAPClientShouldRetry(t *testing.T) {
	client := NewSOAPClient(DefaultConfig())

	// Test with context cancellation - should not retry
	if client.shouldRetry(context.Canceled, nil) {
		t.Error("Should not retry on context cancellation")
	}

	// Test with timeout - should not retry
	if client.shouldRetry(context.DeadlineExceeded, nil) {
		t.Error("Should not retry on timeout")
	}

	// Test with server error response - should retry
	response := &SOAPResponse{StatusCode: 500}
	if !client.shouldRetry(nil, response) {
		t.Error("Should retry on server error")
	}

	// Test with client error response - should not retry
	response = &SOAPResponse{StatusCode: 404}
	if client.shouldRetry(nil, response) {
		t.Error("Should not retry on client error")
	}

	// Test with success response - should not retry
	response = &SOAPResponse{StatusCode: 200}
	if client.shouldRetry(nil, response) {
		t.Error("Should not retry on success")
	}
}

func TestSOAPClientClose(t *testing.T) {
	client := NewSOAPClient(DefaultConfig())

	err := client.Close()
	if err != nil {
		t.Errorf("Close should not return error, got: %v", err)
	}
}
