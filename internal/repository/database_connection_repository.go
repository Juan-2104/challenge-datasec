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

type DatabaseConnectionRepository struct {
	db *sql.DB
}

func NewDatabaseConnectionRepository(db *sql.DB) *DatabaseConnectionRepository {
	return &DatabaseConnectionRepository{db: db}
}

func (r *DatabaseConnectionRepository) Create(ctx context.Context, conn *domain.DatabaseConnection) error {
	query := `
		INSERT INTO database_connections (
			id, host, port, username, encrypted_password, database_name, description,
			created_at, updated_at, last_scanned_at, is_active
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		conn.ID.String(),
		conn.Host,
		conn.Port,
		conn.Username,
		conn.EncryptedPassword,
		conn.DatabaseName,
		conn.Description,
		conn.CreatedAt.UTC(),
		conn.UpdatedAt.UTC(),
		nullTime(conn.LastScannedAt),
		boolToInt(conn.IsActive),
	)
	if err != nil {
		return fmt.Errorf("failed to insert database connection: %w", err)
	}

	return nil
}

func (r *DatabaseConnectionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.DatabaseConnection, error) {
	query := `
		SELECT id, host, port, username, encrypted_password, database_name, description,
			created_at, updated_at, last_scanned_at, is_active
		FROM database_connections
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id.String())
	return scanDatabaseConnection(row)
}

func (r *DatabaseConnectionRepository) GetAll(ctx context.Context) ([]*domain.DatabaseConnection, error) {
	query := `
		SELECT id, host, port, username, encrypted_password, database_name, description,
			created_at, updated_at, last_scanned_at, is_active
		FROM database_connections
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query database connections: %w", err)
	}
	defer rows.Close()

	var result []*domain.DatabaseConnection
	for rows.Next() {
		conn, err := scanDatabaseConnection(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	return result, nil
}

func (r *DatabaseConnectionRepository) GetActive(ctx context.Context) ([]*domain.DatabaseConnection, error) {
	query := `
		SELECT id, host, port, username, encrypted_password, database_name, description,
			created_at, updated_at, last_scanned_at, is_active
		FROM database_connections
		WHERE is_active = 1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active database connections: %w", err)
	}
	defer rows.Close()

	var result []*domain.DatabaseConnection
	for rows.Next() {
		conn, err := scanDatabaseConnection(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating active connections: %w", err)
	}

	return result, nil
}

func (r *DatabaseConnectionRepository) Update(ctx context.Context, conn *domain.DatabaseConnection) error {
	query := `
		UPDATE database_connections
		SET host = ?, port = ?, username = ?, encrypted_password = ?, database_name = ?,
			description = ?, updated_at = ?, last_scanned_at = ?, is_active = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		conn.Host,
		conn.Port,
		conn.Username,
		conn.EncryptedPassword,
		conn.DatabaseName,
		conn.Description,
		conn.UpdatedAt.UTC(),
		nullTime(conn.LastScannedAt),
		boolToInt(conn.IsActive),
		conn.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("failed to update database connection: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("database connection not found")
	}

	return nil
}

func (r *DatabaseConnectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM database_connections WHERE id = ?", id.String())
	if err != nil {
		return fmt.Errorf("failed to delete database connection: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("database connection not found")
	}

	return nil
}

func (r *DatabaseConnectionRepository) UpdateLastScannedAt(ctx context.Context, id uuid.UUID, scannedAt time.Time) error {
	query := `
		UPDATE database_connections
		SET last_scanned_at = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, scannedAt.UTC(), time.Now().UTC(), id.String())
	if err != nil {
		return fmt.Errorf("failed to update last scanned timestamp: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("database connection not found")
	}

	return nil
}

func scanDatabaseConnection(scanner interface {
	Scan(dest ...any) error
}) (*domain.DatabaseConnection, error) {
	var (
		idStr          string
		host           string
		port           int
		username       string
		encrypted      string
		databaseName   sql.NullString
		description    sql.NullString
		createdAt      time.Time
		updatedAt      time.Time
		lastScannedRaw sql.NullTime
		isActive       int
	)

	if err := scanner.Scan(
		&idStr,
		&host,
		&port,
		&username,
		&encrypted,
		&databaseName,
		&description,
		&createdAt,
		&updatedAt,
		&lastScannedRaw,
		&isActive,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("database connection not found")
		}
		return nil, fmt.Errorf("failed to scan database connection: %w", err)
	}

	connectionID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid database connection id: %w", err)
	}

	var lastScanned *time.Time
	if lastScannedRaw.Valid {
		v := lastScannedRaw.Time
		lastScanned = &v
	}

	return &domain.DatabaseConnection{
		ID:                connectionID,
		Host:              host,
		Port:              port,
		Username:          username,
		EncryptedPassword: encrypted,
		DatabaseName:      stringOrEmpty(databaseName),
		Description:       stringOrEmpty(description),
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		LastScannedAt:     lastScanned,
		IsActive:          isActive == 1,
	}, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func stringOrEmpty(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func nullTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC()
}

