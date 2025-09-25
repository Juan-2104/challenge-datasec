package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"database-classifier/internal/domain"
	"database-classifier/internal/infrastructure/database"
	"database-classifier/pkg/security"
)

type ScanService struct {
	scanRepo            domain.ScanResultRepository
	dbConnRepo          domain.DatabaseConnectionRepository
	encryptor           *security.Encryptor
	classificationSvc   domain.ClassificationService
}

func NewScanService(
	scanRepo domain.ScanResultRepository,
	dbConnRepo domain.DatabaseConnectionRepository,
	encryptor *security.Encryptor,
	classificationSvc domain.ClassificationService,
) *ScanService {
	return &ScanService{
		scanRepo:          scanRepo,
		dbConnRepo:        dbConnRepo,
		encryptor:         encryptor,
		classificationSvc: classificationSvc,
	}
}

func (s *ScanService) StartScan(ctx context.Context, databaseID uuid.UUID) (uuid.UUID, error) {
	conn, err := s.dbConnRepo.GetByID(ctx, databaseID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	scanID := uuid.New()
	scanResult := &domain.ScanResult{
		ID:         scanID,
		DatabaseID: databaseID,
		Status:     domain.ScanStatusPending,
		Summary: domain.ScanSummary{
			InformationTypesCounts: make(map[domain.InformationType]int),
		},
		StartedAt: time.Now().UTC(),
	}

	if err := s.scanRepo.Create(ctx, scanResult); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create scan result: %w", err)
	}

	go func() {
		scanCtx := context.Background()
		if err := s.performScan(scanCtx, scanResult, conn); err != nil {
			s.scanRepo.UpdateStatus(scanCtx, scanResult.ID, domain.ScanStatusFailed, err.Error())
		}
	}()

	return scanID, nil
}


func (s *ScanService) performScan(ctx context.Context, scanResult *domain.ScanResult, conn *domain.DatabaseConnection) error {
	startTime := time.Now()

	if err := s.scanRepo.UpdateStatus(ctx, scanResult.ID, domain.ScanStatusRunning, ""); err != nil {
		return fmt.Errorf("failed to update scan status to running: %w", err)
	}

	password, err := s.encryptor.Decrypt(conn.EncryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	inspector := database.NewMySQLInspector()
	defer inspector.Close()

	if err := inspector.Connect(conn.Host, conn.Port, conn.Username, password); err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	schemas, err := inspector.GetSchemas()
	if err != nil {
		return fmt.Errorf("failed to get schemas: %w", err)
	}

	var schemaResults []domain.SchemaResult
	totalTables := 0
	totalColumns := 0
	classifiedColumns := 0
	infoTypeCounts := make(map[domain.InformationType]int)

	for _, schemaName := range schemas {
		tables, err := inspector.GetTables(schemaName)
		if err != nil {
			return fmt.Errorf("failed to get tables for schema %s: %w", schemaName, err)
		}

		var tableResults []domain.TableResult
		totalTables += len(tables)

		for _, tableName := range tables {
			tableInfo, err := inspector.GetTableInfo(schemaName, tableName)
			if err != nil {
				return fmt.Errorf("failed to get table info for %s.%s: %w", schemaName, tableName, err)
			}

			var columnResults []domain.ColumnResult
			totalColumns += len(tableInfo.Columns)

			for _, colInfo := range tableInfo.Columns {
				infoType, score, matched := s.classificationSvc.ClassifyColumn(colInfo.ColumnName)

				columnResult := domain.ColumnResult{
					ColumnName:      colInfo.ColumnName,
					DataType:        colInfo.DataType,
					InformationType: infoType,
					ConfidenceScore: score,
					MatchedPatterns: matched,
					IsNullable:      colInfo.IsNullable,
					DefaultValue:    colInfo.DefaultValue,
				}

				columnResults = append(columnResults, columnResult)

				if infoType != domain.InfoTypeNA {
					classifiedColumns++
					infoTypeCounts[infoType]++
				}
			}

			tableResults = append(tableResults, domain.TableResult{
				TableName: tableName,
				Columns:   columnResults,
			})
		}

		schemaResults = append(schemaResults, domain.SchemaResult{
			SchemaName: schemaName,
			Tables:     tableResults,
		})
	}

	riskLevel := s.calculateRiskLevel(infoTypeCounts, totalColumns)

	endTime := time.Now()
	scanResult.CompletedAt = &endTime
	scanResult.Status = domain.ScanStatusCompleted
	scanResult.Schemas = schemaResults
	scanResult.Summary = domain.ScanSummary{
		TotalSchemas:           len(schemas),
		TotalTables:            totalTables,
		TotalColumns:           totalColumns,
		ClassifiedColumns:      classifiedColumns,
		InformationTypesCounts: infoTypeCounts,
		RiskLevel:              riskLevel,
		DurationMilliseconds:   endTime.Sub(startTime).Milliseconds(),
	}

	if err := s.scanRepo.Update(ctx, scanResult); err != nil {
		return fmt.Errorf("failed to update scan result: %w", err)
	}

	if err := s.dbConnRepo.UpdateLastScannedAt(ctx, conn.ID, endTime); err != nil {
		fmt.Printf("Warning: failed to update last scanned time: %v\n", err)
	}

	return nil
}

func (s *ScanService) calculateRiskLevel(infoTypeCounts map[domain.InformationType]int, totalColumns int) domain.RiskLevel {
	if totalColumns == 0 {
		return domain.RiskLevelLow
	}

	highRiskTypes := []domain.InformationType{
		domain.InfoTypeCreditCardNumber,
		domain.InfoTypeSSN,
		domain.InfoTypePassportNumber,
		domain.InfoTypeNationalID,
		domain.InfoTypeBankAccount,
	}

	mediumRiskTypes := []domain.InformationType{
		domain.InfoTypeEmailAddress,
		domain.InfoTypePhoneNumber,
		domain.InfoTypeDateOfBirth,
		domain.InfoTypeDriverLicense,
		domain.InfoTypeAccountNumber,
	}

	highRiskCount := 0
	mediumRiskCount := 0

	for infoType, count := range infoTypeCounts {
		for _, hrType := range highRiskTypes {
			if infoType == hrType {
				highRiskCount += count
				break
			}
		}
		for _, mrType := range mediumRiskTypes {
			if infoType == mrType {
				mediumRiskCount += count
				break
			}
		}
	}

	totalSensitiveColumns := highRiskCount + mediumRiskCount
	riskPercentage := float64(totalSensitiveColumns) / float64(totalColumns) * 100

	if highRiskCount > 0 && riskPercentage > 20 {
		return domain.RiskLevelCritical
	} else if highRiskCount > 0 || riskPercentage > 15 {
		return domain.RiskLevelHigh
	} else if mediumRiskCount > 0 || riskPercentage > 5 {
		return domain.RiskLevelMedium
	}

	return domain.RiskLevelLow
}

func (s *ScanService) GetScanResult(ctx context.Context, scanID uuid.UUID) (*domain.ScanResult, error) {
	result, err := s.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan result: %w", err)
	}

	return result, nil
}

func (s *ScanService) GetScanHistory(ctx context.Context, databaseID uuid.UUID, limit int) ([]*domain.ScanResult, error) {
	if limit <= 0 {
		limit = 10
	}

	results, err := s.scanRepo.GetByDatabaseID(ctx, databaseID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan history: %w", err)
	}

	return results, nil
}

func (s *ScanService) GetLatestClassification(ctx context.Context, databaseID uuid.UUID) (*domain.ScanResult, error) {
	result, err := s.scanRepo.GetLatestByDatabaseID(ctx, databaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest classification: %w", err)
	}

	return result, nil
}

func (s *ScanService) CancelScan(ctx context.Context, scanID uuid.UUID) error {
	scanResult, err := s.scanRepo.GetByID(ctx, scanID)
	if err != nil {
		return fmt.Errorf("failed to get scan result: %w", err)
	}

	if scanResult.Status != domain.ScanStatusPending && scanResult.Status != domain.ScanStatusRunning {
		return fmt.Errorf("scan cannot be cancelled, current status: %s", scanResult.Status)
	}

	if err := s.scanRepo.UpdateStatus(ctx, scanID, domain.ScanStatusCancelled, "Cancelled by user"); err != nil {
		return fmt.Errorf("failed to cancel scan: %w", err)
	}

	return nil
}
