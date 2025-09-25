package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"database-classifier/internal/domain"
)

type ClassificationPatternRepository struct {
	db *sql.DB
}

func NewClassificationPatternRepository(db *sql.DB) *ClassificationPatternRepository {
	return &ClassificationPatternRepository{db: db}
}

func (r *ClassificationPatternRepository) Create(ctx context.Context, pattern *domain.ClassificationPattern) error {
	query := `
		INSERT INTO classification_patterns (
			id, information_type, pattern, description, priority, is_active, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		pattern.ID.String(),
		pattern.InformationType,
		pattern.Pattern,
		pattern.Description,
		pattern.Priority,
		boolToInt(pattern.IsActive),
		pattern.CreatedAt.UTC(),
		pattern.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to create classification pattern: %w", err)
	}

	return nil
}

func (r *ClassificationPatternRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ClassificationPattern, error) {
	query := `
		SELECT id, information_type, pattern, description, priority, is_active, created_at, updated_at
		FROM classification_patterns
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())
	return scanClassificationPattern(row)
}

func (r *ClassificationPatternRepository) GetAll(ctx context.Context) ([]*domain.ClassificationPattern, error) {
	query := `
		SELECT id, information_type, pattern, description, priority, is_active, created_at, updated_at
		FROM classification_patterns
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query classification patterns: %w", err)
	}
	defer rows.Close()

	var result []*domain.ClassificationPattern
	for rows.Next() {
		pattern, err := scanClassificationPattern(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, pattern)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating classification patterns: %w", err)
	}

	return result, nil
}

func (r *ClassificationPatternRepository) GetActive(ctx context.Context) ([]*domain.ClassificationPattern, error) {
	query := `
		SELECT id, information_type, pattern, description, priority, is_active, created_at, updated_at
		FROM classification_patterns
		WHERE is_active = 1
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active patterns: %w", err)
	}
	defer rows.Close()

	var result []*domain.ClassificationPattern
	for rows.Next() {
		pattern, err := scanClassificationPattern(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, pattern)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active patterns: %w", err)
	}

	return result, nil
}

func (r *ClassificationPatternRepository) GetByInformationType(ctx context.Context, infoType domain.InformationType) ([]*domain.ClassificationPattern, error) {
	query := `
		SELECT id, information_type, pattern, description, priority, is_active, created_at, updated_at
		FROM classification_patterns
		WHERE information_type = ? AND is_active = 1
		ORDER BY priority DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, infoType)
	if err != nil {
		return nil, fmt.Errorf("failed to query patterns by information type: %w", err)
	}
	defer rows.Close()

	var result []*domain.ClassificationPattern
	for rows.Next() {
		pattern, err := scanClassificationPattern(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, pattern)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating patterns: %w", err)
	}

	return result, nil
}

func (r *ClassificationPatternRepository) Update(ctx context.Context, pattern *domain.ClassificationPattern) error {
	query := `
		UPDATE classification_patterns
		SET information_type = ?, pattern = ?, description = ?, priority = ?, is_active = ?, updated_at = ?
		WHERE id = ?
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		pattern.InformationType,
		pattern.Pattern,
		pattern.Description,
		pattern.Priority,
		boolToInt(pattern.IsActive),
		pattern.UpdatedAt.UTC(),
		pattern.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update classification pattern: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("classification pattern not found")
	}

	return nil
}

func (r *ClassificationPatternRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM classification_patterns WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete classification pattern: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("classification pattern not found")
	}

	return nil
}

func (r *ClassificationPatternRepository) ExistsByPattern(ctx context.Context, pattern string) (bool, error) {
	row := r.db.QueryRowContext(ctx, "SELECT COUNT(1) FROM classification_patterns WHERE pattern = ?", pattern)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check pattern existence: %w", err)
	}
	return count > 0, nil
}

func scanClassificationPattern(scanner interface {
	Scan(dest ...any) error
}) (*domain.ClassificationPattern, error) {
	var (
		idStr      string
		infoType   string
		patternStr string
		description sql.NullString
		priority   int
		isActive   int
		createdAt  time.Time
		updatedAt  time.Time
	)

	if err := scanner.Scan(&idStr, &infoType, &patternStr, &description, &priority, &isActive, &createdAt, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("classification pattern not found")
		}
		return nil, fmt.Errorf("failed to scan classification pattern: %w", err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern id: %w", err)
	}

	return &domain.ClassificationPattern{
		ID:              id,
		InformationType: domain.InformationType(infoType),
		Pattern:         patternStr,
		Description:     stringOrEmpty(description),
		Priority:        priority,
		IsActive:        isActive == 1,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}, nil
}

