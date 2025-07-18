// Package soap provides SOAP client functionality for SEFAZ webservice communication
// with support for timeouts, retries, and WS-Security authentication.
package soap

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/adrianodrix/sped-nfe-go/errors"
)

// SOAPClient provides HTTP/SOAP communication with SEFAZ webservices
type SOAPClient struct {
	httpClient    *http.Client
	timeout       time.Duration
	maxRetries    int
	retryDelay    time.Duration
	userAgent     string
	tlsConfig     *tls.Config
	enableLogging bool
}

// SOAPClientConfig holds configuration for SOAP client
type SOAPClientConfig struct {
	Timeout       time.Duration `json:"timeout"`
	MaxRetries    int           `json:"maxRetries"`
	RetryDelay    time.Duration `json:"retryDelay"`
	UserAgent     string        `json:"userAgent"`
	EnableLogging bool          `json:"enableLogging"`
	TLSConfig     *tls.Config   `json:"-"`
}

// SOAPRequest represents a SOAP request
type SOAPRequest struct {
	URL     string
	Action  string
	Body    string
	Headers map[string]string
}

// SOAPResponse represents a SOAP response
type SOAPResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
	Duration   time.Duration
}

// SOAPError represents a SOAP-specific error
type SOAPError struct {
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
	Detail  string `xml:"Detail"`
}

func (e SOAPError) Error() string {
	return fmt.Sprintf("SOAP Error [%s]: %s - %s", e.Code, e.Message, e.Detail)
}

// DefaultConfig returns a default SOAP client configuration
func DefaultConfig() *SOAPClientConfig {
	return &SOAPClientConfig{
		Timeout:       30 * time.Second,
		MaxRetries:    3,
		RetryDelay:    1 * time.Second,
		UserAgent:     "sped-nfe-go/1.0",
		EnableLogging: false,
		TLSConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		},
	}
}

// NewSOAPClient creates a new SOAP client with the given configuration
func NewSOAPClient(config *SOAPClientConfig) *SOAPClient {
	if config == nil {
		config = DefaultConfig()
	}

	// Create HTTP client with proper configuration
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			TLSClientConfig:     config.TLSConfig,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		},
	}

	return &SOAPClient{
		httpClient:    httpClient,
		timeout:       config.Timeout,
		maxRetries:    config.MaxRetries,
		retryDelay:    config.RetryDelay,
		userAgent:     config.UserAgent,
		tlsConfig:     config.TLSConfig,
		enableLogging: config.EnableLogging,
	}
}

// Call performs a SOAP call with automatic retry and proper error handling
func (c *SOAPClient) Call(ctx context.Context, request *SOAPRequest) (*SOAPResponse, error) {
	if request == nil {
		return nil, errors.NewValidationError("SOAP request cannot be nil", "request", "")
	}

	if request.URL == "" {
		return nil, errors.NewValidationError("SOAP request URL cannot be empty", "url", "")
	}

	if request.Body == "" {
		return nil, errors.NewValidationError("SOAP request body cannot be empty", "body", "")
	}

	var lastError error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			delay := time.Duration(float64(c.retryDelay) * math.Pow(2, float64(attempt-1)))
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
			
			if c.enableLogging {
				logDebug("Retrying SOAP call (attempt %d/%d) after %v", attempt+1, c.maxRetries+1, delay)
			}
			
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				// Continue with retry
			}
		}

		response, err := c.performRequest(ctx, request, attempt+1)
		if err == nil {
			return response, nil
		}

		lastError = err

		// Don't retry on certain types of errors
		if !c.shouldRetry(err, response) {
			break
		}
	}

	return nil, lastError
}

// performRequest executes a single SOAP HTTP request
func (c *SOAPClient) performRequest(ctx context.Context, request *SOAPRequest, attempt int) (*SOAPResponse, error) {
	startTime := time.Now()

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", request.URL, bytes.NewBufferString(request.Body))
	if err != nil {
		return nil, errors.NewNetworkError(fmt.Sprintf("failed to create HTTP request: %v", err), err)
	}

	// Set required headers for SOAP
	c.setDefaultHeaders(httpReq, request)

	// Add custom headers
	for key, value := range request.Headers {
		httpReq.Header.Set(key, value)
	}

	if c.enableLogging {
		logDebug("SOAP Request (attempt %d): %s %s", attempt, httpReq.Method, httpReq.URL.String())
		logDebug("SOAP Headers: %+v", httpReq.Header)
		logDebug("SOAP Body: %s", request.Body)
	}

	// Perform the HTTP request
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errors.NewNetworkError(fmt.Sprintf("HTTP request failed: %v", err), err)
	}
	defer httpResp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, errors.NewNetworkError(fmt.Sprintf("failed to read response body: %v", err), err)
	}

	duration := time.Since(startTime)
	
	response := &SOAPResponse{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       string(bodyBytes),
		Duration:   duration,
	}

	if c.enableLogging {
		logDebug("SOAP Response (attempt %d): Status %d, Duration %v", attempt, httpResp.StatusCode, duration)
		logDebug("SOAP Response Headers: %+v", httpResp.Header)
		logDebug("SOAP Response Body: %s", response.Body)
	}

	// Check for HTTP errors
	if httpResp.StatusCode >= 400 {
		return response, errors.NewNetworkError(
			fmt.Sprintf("HTTP error: %d %s", httpResp.StatusCode, http.StatusText(httpResp.StatusCode)),
			fmt.Errorf("status_%d", httpResp.StatusCode),
		)
	}

	return response, nil
}

// setDefaultHeaders sets the required headers for SOAP requests
func (c *SOAPClient) setDefaultHeaders(req *http.Request, soapReq *SOAPRequest) {
	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/xml")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	
	if soapReq.Action != "" {
		req.Header.Set("SOAPAction", fmt.Sprintf("\"%s\"", soapReq.Action))
	}
	
	// Content-Length is set automatically by Go's HTTP client
}

// shouldRetry determines if a request should be retried based on the error and response
func (c *SOAPClient) shouldRetry(err error, response *SOAPResponse) bool {
	// Don't retry on context cancellation or timeout
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// Don't retry on validation errors
	if netErr, ok := err.(*errors.NFError); ok && netErr.Type == errors.ErrValidation {
		return false
	}

	// Retry on network errors
	if _, ok := err.(*errors.NFError); ok {
		return true
	}

	// Retry on certain HTTP status codes
	if response != nil {
		switch response.StatusCode {
		case 500, 502, 503, 504: // Server errors
			return true
		case 408, 429: // Timeout, Too Many Requests
			return true
		case 401, 403, 404: // Authorization/Not Found errors
			return false
		default:
			return false
		}
	}

	return false
}

// SetTimeout updates the client timeout
func (c *SOAPClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.httpClient.Timeout = timeout
}

// SetMaxRetries updates the maximum number of retries
func (c *SOAPClient) SetMaxRetries(maxRetries int) {
	if maxRetries >= 0 {
		c.maxRetries = maxRetries
	}
}

// SetRetryDelay updates the base retry delay
func (c *SOAPClient) SetRetryDelay(delay time.Duration) {
	if delay > 0 {
		c.retryDelay = delay
	}
}

// EnableLogging enables or disables debug logging
func (c *SOAPClient) EnableLogging(enable bool) {
	c.enableLogging = enable
}

// SetTLSConfig updates the TLS configuration
func (c *SOAPClient) SetTLSConfig(tlsConfig *tls.Config) {
	c.tlsConfig = tlsConfig
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.TLSClientConfig = tlsConfig
	}
}

// GetTimeout returns the current timeout setting
func (c *SOAPClient) GetTimeout() time.Duration {
	return c.timeout
}

// GetMaxRetries returns the current max retries setting
func (c *SOAPClient) GetMaxRetries() int {
	return c.maxRetries
}

// GetRetryDelay returns the current retry delay setting
func (c *SOAPClient) GetRetryDelay() time.Duration {
	return c.retryDelay
}

// IsLoggingEnabled returns true if logging is enabled
func (c *SOAPClient) IsLoggingEnabled() bool {
	return c.enableLogging
}

// Close closes the HTTP client and cleans up resources
func (c *SOAPClient) Close() error {
	// Close idle connections
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}

// logDebug logs debug messages if logging is enabled
func logDebug(format string, args ...interface{}) {
	// For now, this is a simple implementation
	// In a real implementation, you might want to use a proper logger
	fmt.Printf("[SOAP DEBUG] "+format+"\n", args...)
}

// ValidateRequest validates a SOAP request
func ValidateRequest(request *SOAPRequest) error {
	if request == nil {
		return errors.NewValidationError("request cannot be nil", "request", "")
	}

	if request.URL == "" {
		return errors.NewValidationError("URL cannot be empty", "url", "")
	}

	if request.Body == "" {
		return errors.NewValidationError("body cannot be empty", "body", "")
	}

	// Basic URL validation
	if len(request.URL) < 8 || (request.URL[:7] != "http://" && request.URL[:8] != "https://") {
		return errors.NewValidationError("URL must start with http:// or https://", "url", request.URL)
	}

	return nil
}

// CreateSimpleRequest creates a basic SOAP request with the given parameters
func CreateSimpleRequest(url, action, body string) *SOAPRequest {
	return &SOAPRequest{
		URL:     url,
		Action:  action,
		Body:    body,
		Headers: make(map[string]string),
	}
}

// AddHeader adds a custom header to the SOAP request
func (r *SOAPRequest) AddHeader(key, value string) {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
}

// RemoveHeader removes a header from the SOAP request
func (r *SOAPRequest) RemoveHeader(key string) {
	if r.Headers != nil {
		delete(r.Headers, key)
	}
}

// GetHeader returns a header value from the SOAP request
func (r *SOAPRequest) GetHeader(key string) string {
	if r.Headers != nil {
		return r.Headers[key]
	}
	return ""
}

// GetHeader returns a header value from the SOAP response
func (r *SOAPResponse) GetHeader(key string) string {
	if r.Headers != nil {
		values := r.Headers[key]
		if len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// GetHeaders returns all header values for a key from the SOAP response
func (r *SOAPResponse) GetHeaders(key string) []string {
	if r.Headers != nil {
		return r.Headers[key]
	}
	return nil
}

// IsSuccess returns true if the response indicates success
func (r *SOAPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError returns true if the response indicates a client error (4xx)
func (r *SOAPResponse) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError returns true if the response indicates a server error (5xx)
func (r *SOAPResponse) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}