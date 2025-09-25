package domain

import (
    "time"

    "github.com/google/uuid"
)

type DatabaseConnection struct {
    ID                uuid.UUID `json:"id"`
    Host              string    `json:"host" binding:"required"`
    Port              int       `json:"port" binding:"required,min=1,max=65535"`
    Username          string    `json:"username" binding:"required"`
    EncryptedPassword string    `json:"-"`
    DatabaseName      string    `json:"database_name"`
    Description       string    `json:"description"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
    LastScannedAt     *time.Time `json:"last_scanned_at,omitempty"`
    IsActive          bool      `json:"is_active"`
}

type CreateDatabaseRequest struct {
	Host         string `json:"host" binding:"required"`
	Port         int    `json:"port" binding:"required,min=1,max=65535"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	DatabaseName string `json:"database_name"`
	Description  string `json:"description"`
}

type ScanResult struct {
    ID           uuid.UUID    `json:"id"`
    DatabaseID   uuid.UUID    `json:"database_id"`
    StartedAt    time.Time    `json:"started_at"`
    CompletedAt  *time.Time   `json:"completed_at,omitempty"`
    Status       ScanStatus   `json:"status"`
    ErrorMessage string       `json:"error_message,omitempty"`
    Schemas      []SchemaResult `json:"schemas"`
    Summary      ScanSummary  `json:"summary"`
}

type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
	ScanStatusCancelled ScanStatus = "cancelled"
)

type SchemaResult struct {
    SchemaName string        `json:"schema_name"`
    Tables     []TableResult `json:"tables"`
}

type TableResult struct {
    TableName string         `json:"table_name"`
    Columns   []ColumnResult `json:"columns"`
}

type ColumnResult struct {
    ColumnName      string          `json:"column_name"`
    DataType        string          `json:"data_type"`
    InformationType InformationType `json:"information_type"`
    ConfidenceScore float64         `json:"confidence_score"`
    MatchedPatterns []string        `json:"matched_patterns"`
    IsNullable      bool            `json:"is_nullable"`
    DefaultValue    *string         `json:"default_value,omitempty"`
}

type InformationType string

const (
	InfoTypeNA               InformationType = "N/A"
	InfoTypeFirstName        InformationType = "FIRST_NAME"
	InfoTypeLastName         InformationType = "LAST_NAME"
	InfoTypeFullName         InformationType = "FULL_NAME"
	InfoTypeUsername         InformationType = "USERNAME"
	InfoTypeEmailAddress     InformationType = "EMAIL_ADDRESS"
	InfoTypePhoneNumber      InformationType = "PHONE_NUMBER"
	InfoTypeCreditCardNumber InformationType = "CREDIT_CARD_NUMBER"
	InfoTypeAccountNumber    InformationType = "ACCOUNT_NUMBER"
	InfoTypeSSN              InformationType = "SSN"
	InfoTypePassportNumber   InformationType = "PASSPORT_NUMBER"
	InfoTypeIPAddress        InformationType = "IP_ADDRESS"
	InfoTypeMACAddress       InformationType = "MAC_ADDRESS"
	InfoTypeAddress          InformationType = "ADDRESS"
	InfoTypePostalCode       InformationType = "POSTAL_CODE"
	InfoTypeDateOfBirth      InformationType = "DATE_OF_BIRTH"
	InfoTypeNationalID       InformationType = "NATIONAL_ID"
	InfoTypeBankAccount      InformationType = "BANK_ACCOUNT"
	InfoTypeDriverLicense    InformationType = "DRIVER_LICENSE"
)

type ScanSummary struct {
    TotalSchemas           int                     `json:"total_schemas"`
    TotalTables            int                     `json:"total_tables"`
    TotalColumns           int                     `json:"total_columns"`
    ClassifiedColumns      int                     `json:"classified_columns"`
    InformationTypesCounts map[InformationType]int `json:"information_types_counts"`
    RiskLevel              RiskLevel               `json:"risk_level"`
    DurationMilliseconds   int64                   `json:"duration_milliseconds"`
}

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type ClassificationPattern struct {
    ID              uuid.UUID        `json:"id"`
    InformationType InformationType  `json:"information_type"`
    Pattern         string           `json:"pattern"`
    Description     string           `json:"description"`
    Priority        int              `json:"priority"`
    IsActive        bool             `json:"is_active"`
    CreatedAt       time.Time        `json:"created_at"`
    UpdatedAt       time.Time        `json:"updated_at"`
}

type CreatePatternRequest struct {
	InformationType InformationType `json:"information_type" binding:"required"`
	Pattern         string          `json:"pattern" binding:"required"`
	Description     string          `json:"description" binding:"required"`
	Priority        int             `json:"priority" binding:"min=1,max=100"`
}

type MySQLTableInfo struct {
    SchemaName string            `json:"schema_name"`
    TableName  string            `json:"table_name"`
    Columns    []MySQLColumnInfo `json:"columns"`
}

type MySQLColumnInfo struct {
	ColumnName   string  `json:"column_name"`
	DataType     string  `json:"data_type"`
	IsNullable   bool    `json:"is_nullable"`
	DefaultValue *string `json:"default_value"`
	ColumnKey    string  `json:"column_key"`
}
