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

type DatabaseService struct {
    dbConnRepo domain.DatabaseConnectionRepository
    encryptor  *security.Encryptor
    inspector  *database.MySQLInspector
}

func NewDatabaseService(
	dbConnRepo domain.DatabaseConnectionRepository,
	encryptor *security.Encryptor,
) *DatabaseService {
	return &DatabaseService{
		dbConnRepo: dbConnRepo,
		encryptor:  encryptor,
		inspector:  database.NewMySQLInspector(),
	}
}

func (s *DatabaseService) CreateConnection(ctx context.Context, req *domain.CreateDatabaseRequest) (uuid.UUID, error) {
    err := s.inspector.TestConnection(req.Host, req.Port, req.Username, req.Password, req.DatabaseName)
    if err != nil {
        return uuid.Nil, fmt.Errorf("failed to connect to MySQL database: %w", err)
    }

    // Encrypt the password
    encryptedPassword, err := s.encryptor.Encrypt(req.Password)
    if err != nil {
        return uuid.Nil, fmt.Errorf("failed to encrypt password: %w", err)
    }

    id := uuid.New()
    now := time.Now().UTC()
    conn := &domain.DatabaseConnection{
        ID:                id,
        Host:              req.Host,
        Port:              req.Port,
        Username:          req.Username,
        EncryptedPassword: encryptedPassword,
        DatabaseName:      req.DatabaseName,
        Description:       req.Description,
        IsActive:          true,
        CreatedAt:         now,
        UpdatedAt:         now,
    }

    if err := s.dbConnRepo.Create(ctx, conn); err != nil {
        return uuid.Nil, fmt.Errorf("failed to save database connection: %w", err)
    }

    return id, nil
}

func (s *DatabaseService) GetConnection(ctx context.Context, id uuid.UUID) (*domain.DatabaseConnection, error) {
    conn, err := s.dbConnRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	return conn, nil
}

func (s *DatabaseService) GetAllConnections(ctx context.Context) ([]*domain.DatabaseConnection, error) {
    connections, err := s.dbConnRepo.GetAll(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get all database connections: %w", err)
    }

    return connections, nil
}

func (s *DatabaseService) UpdateConnection(ctx context.Context, id uuid.UUID, req *domain.CreateDatabaseRequest) error {
    conn, err := s.dbConnRepo.GetByID(ctx, id)
    if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	needsTest := conn.Host != req.Host ||
		conn.Port != req.Port ||
		conn.Username != req.Username ||
		conn.DatabaseName != req.DatabaseName

	if req.Password != "" || needsTest {
		password := req.Password
		if password == "" {
			password, err = s.encryptor.Decrypt(conn.EncryptedPassword)
			if err != nil {
				return fmt.Errorf("failed to decrypt existing password: %w", err)
			}
		}

		err = s.inspector.TestConnection(req.Host, req.Port, req.Username, password, req.DatabaseName)
		if err != nil {
			return fmt.Errorf("failed to connect to MySQL database: %w", err)
		}
	}

    conn.Host = req.Host
    conn.Port = req.Port
    conn.Username = req.Username
    conn.DatabaseName = req.DatabaseName
    conn.Description = req.Description
    conn.UpdatedAt = time.Now().UTC()

    if req.Password != "" {
		encryptedPassword, err := s.encryptor.Encrypt(req.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		conn.EncryptedPassword = encryptedPassword
	}

	if err := s.dbConnRepo.Update(ctx, conn); err != nil {
		return fmt.Errorf("failed to update database connection: %w", err)
	}

	return nil
}

func (s *DatabaseService) DeleteConnection(ctx context.Context, id uuid.UUID) error {
    if err := s.dbConnRepo.Delete(ctx, id); err != nil {
        return fmt.Errorf("failed to delete database connection: %w", err)
    }

    return nil
}

func (s *DatabaseService) TestConnection(ctx context.Context, id uuid.UUID) error {
	conn, err := s.dbConnRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	password, err := s.encryptor.Decrypt(conn.EncryptedPassword)
	if err != nil {
		return fmt.Errorf("failed to decrypt password: %w", err)
	}

	err = s.inspector.TestConnection(conn.Host, conn.Port, conn.Username, password, conn.DatabaseName)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

    return nil
}
