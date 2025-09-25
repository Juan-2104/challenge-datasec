package domain

import (
    "context"
    "time"

    "github.com/google/uuid"
)

type DatabaseConnectionRepository interface {
    Create(ctx context.Context, conn *DatabaseConnection) error
    GetByID(ctx context.Context, id uuid.UUID) (*DatabaseConnection, error)
    GetAll(ctx context.Context) ([]*DatabaseConnection, error)
    Update(ctx context.Context, conn *DatabaseConnection) error
    Delete(ctx context.Context, id uuid.UUID) error
    GetActive(ctx context.Context) ([]*DatabaseConnection, error)
    UpdateLastScannedAt(ctx context.Context, id uuid.UUID, scannedAt time.Time) error
}

type ScanResultRepository interface {
    Create(ctx context.Context, result *ScanResult) error
    GetByID(ctx context.Context, id uuid.UUID) (*ScanResult, error)
    GetByDatabaseID(ctx context.Context, databaseID uuid.UUID, limit int) ([]*ScanResult, error)
    GetLatestByDatabaseID(ctx context.Context, databaseID uuid.UUID) (*ScanResult, error)
    Update(ctx context.Context, result *ScanResult) error
    Delete(ctx context.Context, id uuid.UUID) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status ScanStatus, errorMessage string) error
    GetRunningScans(ctx context.Context) ([]*ScanResult, error)
}

type ClassificationPatternRepository interface {
    Create(ctx context.Context, pattern *ClassificationPattern) error
    GetByID(ctx context.Context, id uuid.UUID) (*ClassificationPattern, error)
    GetAll(ctx context.Context) ([]*ClassificationPattern, error)
    GetActive(ctx context.Context) ([]*ClassificationPattern, error)
    GetByInformationType(ctx context.Context, infoType InformationType) ([]*ClassificationPattern, error)
    Update(ctx context.Context, pattern *ClassificationPattern) error
    Delete(ctx context.Context, id uuid.UUID) error
    ExistsByPattern(ctx context.Context, pattern string) (bool, error)
}
