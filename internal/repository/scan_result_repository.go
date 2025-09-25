package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"database-classifier/internal/domain"
)

type ScanResultRepository struct {
	db *sql.DB
}

func NewScanResultRepository(db *sql.DB) *ScanResultRepository {
	return &ScanResultRepository{db: db}
}

func (r *ScanResultRepository) Create(ctx context.Context, result *domain.ScanResult) error {
	if result.ID == uuid.Nil {
		result.ID = uuid.New()
	}
	if result.StartedAt.IsZero() {
		result.StartedAt = time.Now().UTC()
	}

	schemasJSON, err := json.Marshal(result.Schemas)
	if err != nil {
		return fmt.Errorf("failed to marshal schemas: %w", err)
	}

	summaryJSON, err := json.Marshal(result.Summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	query := `
		INSERT INTO scan_results (
			id, database_id, started_at, completed_at, status, error_message, schemas_json, summary_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecContext(
		ctx,
		query,
		result.ID.String(),
		result.DatabaseID.String(),
		result.StartedAt.UTC(),
		nullTime(result.CompletedAt),
		result.Status,
		result.ErrorMessage,
		schemasJSON,
		summaryJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create scan result: %w", err)
	}

	return nil
}

func (r *ScanResultRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ScanResult, error) {
	query := `
		SELECT id, database_id, started_at, completed_at, status, error_message, schemas_json, summary_json
		FROM scan_results
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())
	return scanScanResult(row)
}

func (r *ScanResultRepository) GetByDatabaseID(ctx context.Context, databaseID uuid.UUID, limit int) ([]*domain.ScanResult, error) {
	query := `
		SELECT id, database_id, started_at, completed_at, status, error_message, schemas_json, summary_json
		FROM scan_results
		WHERE database_id = ?
		ORDER BY started_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, databaseID.String(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query scan results: %w", err)
	}
	defer rows.Close()

	var results []*domain.ScanResult
	for rows.Next() {
		scan, err := scanScanResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, scan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating scan results: %w", err)
	}

	return results, nil
}

func (r *ScanResultRepository) GetLatestByDatabaseID(ctx context.Context, databaseID uuid.UUID) (*domain.ScanResult, error) {
	query := `
		SELECT id, database_id, started_at, completed_at, status, error_message, schemas_json, summary_json
		FROM scan_results
		WHERE database_id = ? AND status = ?
		ORDER BY started_at DESC
		LIMIT 1
	`

	row := r.db.QueryRowContext(ctx, query, databaseID.String(), domain.ScanStatusCompleted)
	return scanScanResult(row)
}

func (r *ScanResultRepository) Update(ctx context.Context, result *domain.ScanResult) error {
	schemasJSON, err := json.Marshal(result.Schemas)
	if err != nil {
		return fmt.Errorf("failed to marshal schemas: %w", err)
	}

	summaryJSON, err := json.Marshal(result.Summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	query := `
		UPDATE scan_results
		SET database_id = ?, started_at = ?, completed_at = ?, status = ?, error_message = ?,
			schemas_json = ?, summary_json = ?
		WHERE id = ?
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		result.DatabaseID.String(),
		result.StartedAt.UTC(),
		nullTime(result.CompletedAt),
		result.Status,
		result.ErrorMessage,
		schemasJSON,
		summaryJSON,
		result.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update scan result: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("scan result not found")
	}

	return nil
}

func (r *ScanResultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM scan_results WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete scan result: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("scan result not found")
	}

	return nil
}

func (r *ScanResultRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ScanStatus, errorMessage string) error {
	var completedAt any
	if status == domain.ScanStatusCompleted || status == domain.ScanStatusFailed || status == domain.ScanStatusCancelled {
		completedAt = time.Now().UTC()
	}

	query := `
		UPDATE scan_results
		SET status = ?, completed_at = ?, error_message = ?
		WHERE id = ?
	`

	res, err := r.db.ExecContext(ctx, query, status, completedAt, errorMessage, id.String())
	if err != nil {
		return fmt.Errorf("failed to update scan status: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("scan result not found")
	}

	return nil
}

func (r *ScanResultRepository) GetRunningScans(ctx context.Context) ([]*domain.ScanResult, error) {
	query := `
		SELECT id, database_id, started_at, completed_at, status, error_message, schemas_json, summary_json
		FROM scan_results
		WHERE status IN (?, ?)
		ORDER BY started_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, domain.ScanStatusPending, domain.ScanStatusRunning)
	if err != nil {
		return nil, fmt.Errorf("failed to query running scans: %w", err)
	}
	defer rows.Close()

	var results []*domain.ScanResult
	for rows.Next() {
		scan, err := scanScanResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, scan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating running scans: %w", err)
	}

	return results, nil
}

func scanScanResult(scanner interface {
	Scan(dest ...any) error
}) (*domain.ScanResult, error) {
	var (
		idStr        string
		dbIDStr      string
		startedAt    time.Time
		completedRaw sql.NullTime
		status       string
		errorMessage sql.NullString
		schemasJSON  []byte
		summaryJSON  []byte
	)

	if err := scanner.Scan(&idStr, &dbIDStr, &startedAt, &completedRaw, &status, &errorMessage, &schemasJSON, &summaryJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("scan result not found")
		}
		return nil, fmt.Errorf("failed to scan result: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid scan result id: %w", err)
	}

	dbID, err := uuid.Parse(dbIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid database id: %w", err)
	}

	var schemas []domain.SchemaResult
	if len(schemasJSON) > 0 {
		if err := json.Unmarshal(schemasJSON, &schemas); err != nil {
			return nil, fmt.Errorf("failed to unmarshal schemas: %w", err)
		}
	}

	var summary domain.ScanSummary
	if len(summaryJSON) > 0 {
		if err := json.Unmarshal(summaryJSON, &summary); err != nil {
			return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
		}
	}

	var completedAt *time.Time
	if completedRaw.Valid {
		v := completedRaw.Time
		completedAt = &v
	}

	result := &domain.ScanResult{
		ID:           id,
		DatabaseID:   dbID,
		StartedAt:    startedAt,
		CompletedAt:  completedAt,
		Status:       domain.ScanStatus(status),
		ErrorMessage: stringOrEmpty(errorMessage),
		Schemas:      schemas,
		Summary:      summary,
	}

	return result, nil
}
