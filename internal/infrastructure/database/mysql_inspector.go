package database

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"database-classifier/internal/domain"
)

type MySQLInspector struct {
	db *sql.DB
}

func NewMySQLInspector() *MySQLInspector {
	return &MySQLInspector{}
}

func (m *MySQLInspector) Connect(host string, port int, username, password string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/information_schema?parseTime=true&charset=utf8mb4",
		username, password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0)

	m.db = db
	return nil
}

func (m *MySQLInspector) GetSchemas() ([]string, error) {
	if m.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT SCHEMA_NAME 
		FROM SCHEMATA 
		WHERE SCHEMA_NAME NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
		ORDER BY SCHEMA_NAME
	`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}
		schemas = append(schemas, schemaName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schemas: %w", err)
	}

	return schemas, nil
}

func (m *MySQLInspector) GetTables(schema string) ([]string, error) {
	if m.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT TABLE_NAME 
		FROM TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := m.db.Query(query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables for schema %s: %w", schema, err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	return tables, nil
}

func (m *MySQLInspector) GetTableInfo(schema, table string) (*domain.MySQLTableInfo, error) {
	if m.db == nil {
		return nil, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			COLUMN_KEY
		FROM COLUMNS 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := m.db.Query(query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for table %s.%s: %w", schema, table, err)
	}
	defer rows.Close()

	var columns []domain.MySQLColumnInfo
	for rows.Next() {
		var column domain.MySQLColumnInfo
		var isNullable string
		var defaultValue sql.NullString

		if err := rows.Scan(
			&column.ColumnName,
			&column.DataType,
			&isNullable,
			&defaultValue,
			&column.ColumnKey,
		); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}

		column.IsNullable = isNullable == "YES"
		if defaultValue.Valid {
			column.DefaultValue = &defaultValue.String
		}

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns: %w", err)
	}

	return &domain.MySQLTableInfo{
		SchemaName: schema,
		TableName:  table,
		Columns:    columns,
	}, nil
}

func (m *MySQLInspector) TestConnection(host string, port int, username, password, database string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4",
		username, password, host, port, database)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	return nil
}

func (m *MySQLInspector) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}
func (m *MySQLInspector) GetDatabaseSize() (int64, error) {
	if m.db == nil {
		return 0, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT COALESCE(SUM(DATA_LENGTH + INDEX_LENGTH), 0) as total_size
		FROM TABLES 
		WHERE TABLE_SCHEMA NOT IN ('information_schema', 'performance_schema', 'mysql', 'sys')
	`

	var sizeStr string
	err := m.db.QueryRow(query).Scan(&sizeStr)
	if err != nil {
		return 0, fmt.Errorf("failed to query database size: %w", err)
	}

	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse database size: %w", err)
	}

	return size, nil
}
func (m *MySQLInspector) GetTableRowCount(schema, table string) (int64, error) {
	if m.db == nil {
		return 0, fmt.Errorf("not connected to database")
	}

	query := `
		SELECT TABLE_ROWS 
		FROM TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`

	var count sql.NullInt64
	err := m.db.QueryRow(query, schema, table).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to query table row count: %w", err)
	}

	if count.Valid {
		return count.Int64, nil
	}

	return 0, nil
}
