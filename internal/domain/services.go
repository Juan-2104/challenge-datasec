package domain

import (
    "context"

    "github.com/google/uuid"
)

type DatabaseService interface {
    CreateConnection(ctx context.Context, req *CreateDatabaseRequest) (uuid.UUID, error)
    GetConnection(ctx context.Context, id uuid.UUID) (*DatabaseConnection, error)
    GetAllConnections(ctx context.Context) ([]*DatabaseConnection, error)
    UpdateConnection(ctx context.Context, id uuid.UUID, req *CreateDatabaseRequest) error
    DeleteConnection(ctx context.Context, id uuid.UUID) error
    TestConnection(ctx context.Context, id uuid.UUID) error
}

type ScanService interface {
    StartScan(ctx context.Context, databaseID uuid.UUID) (uuid.UUID, error)
    GetScanResult(ctx context.Context, scanID uuid.UUID) (*ScanResult, error)
    GetScanHistory(ctx context.Context, databaseID uuid.UUID, limit int) ([]*ScanResult, error)
    GetLatestClassification(ctx context.Context, databaseID uuid.UUID) (*ScanResult, error)
    CancelScan(ctx context.Context, scanID uuid.UUID) error
}

type ClassificationService interface {
    CreatePattern(ctx context.Context, req *CreatePatternRequest) (uuid.UUID, error)
    GetPattern(ctx context.Context, id uuid.UUID) (*ClassificationPattern, error)
    GetAllPatterns(ctx context.Context) ([]*ClassificationPattern, error)
    UpdatePattern(ctx context.Context, id uuid.UUID, req *CreatePatternRequest) error
    DeletePattern(ctx context.Context, id uuid.UUID) error
    ClassifyColumn(columnName string) (InformationType, float64, []string)
}

type MySQLInspector interface {
	Connect(host string, port int, username, password string) error
	GetSchemas() ([]string, error)
	GetTables(schema string) ([]string, error)
	GetTableInfo(schema, table string) (*MySQLTableInfo, error)
	Close() error
}
