// Package nfe provides access key generation and validation for NFe documents.
package nfe

import (
	"fmt"
	"strconv"
	"time"
)

// AccessKey represents an NFe access key with its components
type AccessKey struct {
	State          string    // UF code (2 digits)
	YearMonth      string    // AAMM (4 digits)
	Document       string    // CNPJ (14 digits)
	Model          string    // Document model (2 digits)
	Series         string    // Series (3 digits)
	Number         string    // Number (9 digits)
	EmissionType   string    // Emission type (1 digit)
	RandomCode     string    // Random code (8 digits)
	CheckDigit     string    // Check digit (1 digit)
	FullKey        string    // Complete 44-digit key
	IssueDateTime  time.Time // Issue date/time for AAMM calculation
}

// NewAccessKey creates a new access key
func NewAccessKey() *AccessKey {
	return &AccessKey{
		IssueDateTime: time.Now(),
	}
}

// SetState sets the state code
func (ak *AccessKey) SetState(state string) *AccessKey {
	ak.State = state
	return ak
}

// SetStateByCode sets the state using IBGE code
func (ak *AccessKey) SetStateByCode(code string) *AccessKey {
	ak.State = code
	return ak
}

// SetStateByName sets the state using state name
func (ak *AccessKey) SetStateByName(stateName string) *AccessKey {
	if code := GetStateCode(stateName); code != "" {
		ak.State = code
	}
	return ak
}

// SetDocument sets the document (CNPJ)
func (ak *AccessKey) SetDocument(cnpj string) *AccessKey {
	ak.Document = OnlyNumbers(cnpj)
	return ak
}

// SetModel sets the document model
func (ak *AccessKey) SetModel(model DocumentModel) *AccessKey {
	ak.Model = model.String()
	return ak
}

// SetSeries sets the series
func (ak *AccessKey) SetSeries(series int) *AccessKey {
	ak.Series = ZeroPad(series, 3)
	return ak
}

// SetNumber sets the document number
func (ak *AccessKey) SetNumber(number int) *AccessKey {
	ak.Number = ZeroPad(number, 9)
	return ak
}

// SetEmissionType sets the emission type
func (ak *AccessKey) SetEmissionType(emissionType EmissionType) *AccessKey {
	ak.EmissionType = emissionType.String()
	return ak
}

// SetRandomCode sets the random code
func (ak *AccessKey) SetRandomCode(code string) *AccessKey {
	ak.RandomCode = ZeroPad(0, 8) // Will be replaced with actual random code
	if code != "" {
		ak.RandomCode = PadLeft(OnlyNumbers(code), 8, '0')
	}
	return ak
}

// SetIssueDateTime sets the issue date/time for AAMM calculation
func (ak *AccessKey) SetIssueDateTime(dateTime time.Time) *AccessKey {
	ak.IssueDateTime = dateTime
	return ak
}

// GenerateRandomCode generates a random 8-digit code
func (ak *AccessKey) GenerateRandomCode() *AccessKey {
	// Generate random code, ensuring it's different from document number
	for {
		code := GenerateRandomCode(8)
		if code != ak.Number[len(ak.Number)-8:] { // Ensure cNF != nNF (NT2019.001)
			ak.RandomCode = code
			break
		}
	}
	return ak
}

// CalculateCheckDigit calculates and sets the check digit
func (ak *AccessKey) CalculateCheckDigit() *AccessKey {
	// Build key without check digit
	keyWithoutDV := ak.buildKeyWithoutCheckDigit()
	
	// Calculate check digit using modulo 11
	checkDigit := CalculateModulo11(keyWithoutDV)
	ak.CheckDigit = strconv.Itoa(checkDigit)
	
	return ak
}

// Build builds the complete access key
func (ak *AccessKey) Build() error {
	// Validate required fields
	if err := ak.validate(); err != nil {
		return err
	}
	
	// Calculate year-month if not set
	if ak.YearMonth == "" {
		ak.YearMonth = FormatYearMonth(ak.IssueDateTime)
	}
	
	// Generate random code if not set
	if ak.RandomCode == "" {
		ak.GenerateRandomCode()
	}
	
	// Calculate check digit
	ak.CalculateCheckDigit()
	
	// Build full key
	ak.FullKey = ak.buildFullKey()
	
	return nil
}

// GetKey returns the complete 44-digit access key
func (ak *AccessKey) GetKey() string {
	if ak.FullKey == "" {
		ak.Build()
	}
	return ak.FullKey
}

// GetFormattedKey returns the access key with formatting
func (ak *AccessKey) GetFormattedKey() string {
	key := ak.GetKey()
	if len(key) != 44 {
		return key
	}
	
	// Format: NNNN NNNN NNNN NNNN NNNN NNNN NNNN NNNN NNNN NNNN NNNN
	formatted := ""
	for i := 0; i < len(key); i += 4 {
		if i > 0 {
			formatted += " "
		}
		end := i + 4
		if end > len(key) {
			end = len(key)
		}
		formatted += key[i:end]
	}
	
	return formatted
}

// IsValid validates the access key
func (ak *AccessKey) IsValid() bool {
	if ak.FullKey == "" {
		return false
	}
	
	if len(ak.FullKey) != 44 {
		return false
	}
	
	// Validate check digit
	keyWithoutDV := ak.FullKey[:43]
	expectedCheckDigit := CalculateModulo11(keyWithoutDV)
	actualCheckDigit, err := strconv.Atoi(string(ak.FullKey[43]))
	if err != nil {
		return false
	}
	
	return expectedCheckDigit == actualCheckDigit
}

// ParseAccessKey parses a 44-digit access key into its components
func ParseAccessKey(key string) (*AccessKey, error) {
	key = OnlyNumbers(key)
	
	if len(key) != 44 {
		return nil, fmt.Errorf("access key must have exactly 44 digits, got %d", len(key))
	}
	
	ak := &AccessKey{
		State:        key[0:2],
		YearMonth:    key[2:6],
		Document:     key[6:20],
		Model:        key[20:22],
		Series:       key[22:25],
		Number:       key[25:34],
		EmissionType: key[34:35],
		RandomCode:   key[35:43],
		CheckDigit:   key[43:44],
		FullKey:      key,
	}
	
	// Parse year-month to set issue date
	if len(ak.YearMonth) == 4 {
		year, _ := strconv.Atoi("20" + ak.YearMonth[:2]) // Assume 20XX
		month, _ := strconv.Atoi(ak.YearMonth[2:4])
		ak.IssueDateTime = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}
	
	// Validate check digit
	if !ak.IsValid() {
		return nil, fmt.Errorf("invalid access key check digit")
	}
	
	return ak, nil
}

// validate validates the access key components
func (ak *AccessKey) validate() error {
	if ak.State == "" {
		return fmt.Errorf("state code is required")
	}
	
	if len(ak.State) != 2 {
		return fmt.Errorf("state code must have exactly 2 digits")
	}
	
	if ak.Document == "" {
		return fmt.Errorf("document (CNPJ) is required")
	}
	
	if len(ak.Document) != 14 {
		return fmt.Errorf("document (CNPJ) must have exactly 14 digits")
	}
	
	if ak.Model == "" {
		return fmt.Errorf("document model is required")
	}
	
	if len(ak.Model) != 2 {
		return fmt.Errorf("document model must have exactly 2 digits")
	}
	
	if ak.Series == "" {
		return fmt.Errorf("series is required")
	}
	
	if len(ak.Series) != 3 {
		return fmt.Errorf("series must have exactly 3 digits")
	}
	
	if ak.Number == "" {
		return fmt.Errorf("document number is required")
	}
	
	if len(ak.Number) != 9 {
		return fmt.Errorf("document number must have exactly 9 digits")
	}
	
	if ak.EmissionType == "" {
		return fmt.Errorf("emission type is required")
	}
	
	if len(ak.EmissionType) != 1 {
		return fmt.Errorf("emission type must have exactly 1 digit")
	}
	
	return nil
}

// buildKeyWithoutCheckDigit builds the key without the check digit (43 digits)
func (ak *AccessKey) buildKeyWithoutCheckDigit() string {
	yearMonth := ak.YearMonth
	if yearMonth == "" {
		yearMonth = FormatYearMonth(ak.IssueDateTime)
	}
	
	return ak.State + yearMonth + ak.Document + ak.Model + ak.Series + ak.Number + ak.EmissionType + ak.RandomCode
}

// buildFullKey builds the complete 44-digit key
func (ak *AccessKey) buildFullKey() string {
	return ak.buildKeyWithoutCheckDigit() + ak.CheckDigit
}

// AccessKeyBuilder provides a fluent interface for building access keys
type AccessKeyBuilder struct {
	accessKey *AccessKey
}

// NewAccessKeyBuilder creates a new access key builder
func NewAccessKeyBuilder() *AccessKeyBuilder {
	return &AccessKeyBuilder{
		accessKey: NewAccessKey(),
	}
}

// State sets the state
func (akb *AccessKeyBuilder) State(state string) *AccessKeyBuilder {
	akb.accessKey.SetState(state)
	return akb
}

// StateByName sets the state by name
func (akb *AccessKeyBuilder) StateByName(stateName string) *AccessKeyBuilder {
	akb.accessKey.SetStateByName(stateName)
	return akb
}

// Document sets the document (CNPJ)
func (akb *AccessKeyBuilder) Document(cnpj string) *AccessKeyBuilder {
	akb.accessKey.SetDocument(cnpj)
	return akb
}

// Model sets the document model
func (akb *AccessKeyBuilder) Model(model DocumentModel) *AccessKeyBuilder {
	akb.accessKey.SetModel(model)
	return akb
}

// Series sets the series
func (akb *AccessKeyBuilder) Series(series int) *AccessKeyBuilder {
	akb.accessKey.SetSeries(series)
	return akb
}

// Number sets the document number
func (akb *AccessKeyBuilder) Number(number int) *AccessKeyBuilder {
	akb.accessKey.SetNumber(number)
	return akb
}

// EmissionType sets the emission type
func (akb *AccessKeyBuilder) EmissionType(emissionType EmissionType) *AccessKeyBuilder {
	akb.accessKey.SetEmissionType(emissionType)
	return akb
}

// RandomCode sets the random code
func (akb *AccessKeyBuilder) RandomCode(code string) *AccessKeyBuilder {
	akb.accessKey.SetRandomCode(code)
	return akb
}

// IssueDateTime sets the issue date/time
func (akb *AccessKeyBuilder) IssueDateTime(dateTime time.Time) *AccessKeyBuilder {
	akb.accessKey.SetIssueDateTime(dateTime)
	return akb
}

// Build builds and returns the access key
func (akb *AccessKeyBuilder) Build() (*AccessKey, error) {
	err := akb.accessKey.Build()
	if err != nil {
		return nil, err
	}
	return akb.accessKey, nil
}

// MustBuild builds the access key and panics on error
func (akb *AccessKeyBuilder) MustBuild() *AccessKey {
	ak, err := akb.Build()
	if err != nil {
		panic(err)
	}
	return ak
}

// Convenience functions for quick access key generation

// GenerateAccessKey generates an access key with the given parameters
func GenerateAccessKey(state, cnpj string, model DocumentModel, series, number int, emissionType EmissionType) (*AccessKey, error) {
	return NewAccessKeyBuilder().
		State(state).
		Document(cnpj).
		Model(model).
		Series(series).
		Number(number).
		EmissionType(emissionType).
		Build()
}

// GenerateAccessKeyFromIdentification generates access key from Identificacao struct
func GenerateAccessKeyFromIdentification(ide *Identificacao, cnpj string) (*AccessKey, error) {
	series, _ := strconv.Atoi(ide.Serie)
	number, _ := strconv.Atoi(ide.NNF)
	emissionType := EmissionNormal
	
	// Parse emission type
	switch ide.TpEmis {
	case "1":
		emissionType = EmissionNormal
	case "2":
		emissionType = EmissionContingencyFS
	case "3":
		emissionType = EmissionContingencySCAN
	case "4":
		emissionType = EmissionContingencyDPEC
	case "5":
		emissionType = EmissionContingencyFSDA
	case "6":
		emissionType = EmissionContingencySVCAN
	case "7":
		emissionType = EmissionContingencySVCRS
	case "9":
		emissionType = EmissionOffline
	}
	
	// Parse model
	model := ModelNFe
	if ide.Mod == "65" {
		model = ModelNFCe
	}
	
	builder := NewAccessKeyBuilder().
		State(ide.CUF).
		Document(cnpj).
		Model(model).
		Series(series).
		Number(number).
		EmissionType(emissionType)
	
	// Parse issue date if available
	if ide.DhEmi != "" {
		if issueDate, err := ParseDateTime(ide.DhEmi); err == nil {
			builder.IssueDateTime(issueDate)
		}
	}
	
	// Use existing cNF if available
	if ide.CNF != "" {
		builder.RandomCode(ide.CNF)
	}
	
	return builder.Build()
}

// ValidateAccessKey validates a 44-digit access key
func ValidateAccessKey(key string) bool {
	ak, err := ParseAccessKey(key)
	if err != nil {
		return false
	}
	return ak.IsValid()
}

// FormatAccessKey formats an access key for display
func FormatAccessKey(key string) string {
	if ak, err := ParseAccessKey(key); err == nil {
		return ak.GetFormattedKey()
	}
	return key
}