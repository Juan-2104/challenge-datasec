package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"database-classifier/internal/domain"
	"database-classifier/pkg/classifier"
)

type ClassificationService struct {
	repo     domain.ClassificationPatternRepository
	mu       sync.RWMutex
	matcher  *classifier.Classifier
}

func NewClassificationService(ctx context.Context, repo domain.ClassificationPatternRepository, defaultPatternsPath string) (*ClassificationService, error) {
	svc := &ClassificationService{repo: repo}
	if err := svc.ensurePatterns(ctx, defaultPatternsPath); err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *ClassificationService) ensurePatterns(ctx context.Context, defaultPatternsPath string) error {
	patterns, err := s.repo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to load patterns: %w", err)
	}

	if len(patterns) == 0 && defaultPatternsPath != "" {
		seedPatterns, err := loadPatternSeeds(defaultPatternsPath)
		if err != nil {
			return err
		}
		for _, seed := range seedPatterns {
			exists, err := s.repo.ExistsByPattern(ctx, seed.Pattern)
			if err != nil {
				return err
			}
			if exists {
				continue
			}

			model := &domain.ClassificationPattern{
				ID:              uuid.New(),
				InformationType: domain.InformationType(seed.InformationType),
				Pattern:         seed.Pattern,
				Description:     seed.Description,
				Priority:        seed.Priority,
				IsActive:        true,
				CreatedAt:       time.Now().UTC(),
				UpdatedAt:       time.Now().UTC(),
			}

			if err := s.repo.Create(ctx, model); err != nil {
				return fmt.Errorf("failed to seed pattern %s: %w", seed.Pattern, err)
			}
		}

		patterns, err = s.repo.GetActive(ctx)
		if err != nil {
			return fmt.Errorf("failed to reload patterns: %w", err)
		}
	}

	return s.refreshClassifier(patterns)
}

func (s *ClassificationService) refreshClassifier(patterns []*domain.ClassificationPattern) error {
	matcher, err := classifier.NewClassifier(patterns)
	if err != nil {
		return fmt.Errorf("failed to prepare classifier: %w", err)
	}

	s.mu.Lock()
	s.matcher = matcher
	s.mu.Unlock()

	return nil
}

func (s *ClassificationService) CreatePattern(ctx context.Context, req *domain.CreatePatternRequest) (uuid.UUID, error) {
	exists, err := s.repo.ExistsByPattern(ctx, req.Pattern)
	if err != nil {
		return uuid.Nil, err
	}
	if exists {
		return uuid.Nil, fmt.Errorf("pattern already exists")
	}

	id := uuid.New()
	now := time.Now().UTC()
	pattern := &domain.ClassificationPattern{
		ID:              id,
		InformationType: req.InformationType,
		Pattern:         req.Pattern,
		Description:     req.Description,
		Priority:        req.Priority,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, pattern); err != nil {
		return uuid.Nil, fmt.Errorf("failed to create pattern: %w", err)
	}

	if err := s.reloadMatcher(ctx); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

func (s *ClassificationService) GetPattern(ctx context.Context, id uuid.UUID) (*domain.ClassificationPattern, error) {
	pattern, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return pattern, nil
}

func (s *ClassificationService) GetAllPatterns(ctx context.Context) ([]*domain.ClassificationPattern, error) {
	return s.repo.GetAll(ctx)
}

func (s *ClassificationService) UpdatePattern(ctx context.Context, id uuid.UUID, req *domain.CreatePatternRequest) error {
	pattern, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	pattern.InformationType = req.InformationType
	pattern.Pattern = req.Pattern
	pattern.Description = req.Description
	pattern.Priority = req.Priority
	pattern.UpdatedAt = time.Now().UTC()
	pattern.IsActive = true

	if err := s.repo.Update(ctx, pattern); err != nil {
		return fmt.Errorf("failed to update pattern: %w", err)
	}

	return s.reloadMatcher(ctx)
}

func (s *ClassificationService) DeletePattern(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return s.reloadMatcher(ctx)
}

func (s *ClassificationService) ClassifyColumn(columnName string) (domain.InformationType, float64, []string) {
	s.mu.RLock()
	matcher := s.matcher
	s.mu.RUnlock()

	if matcher == nil {
		return domain.InfoTypeNA, 0.0, []string{}
	}

	res := matcher.ClassifyColumn(columnName)
	return res.InformationType, res.ConfidenceScore, res.MatchedPatterns
}

func (s *ClassificationService) reloadMatcher(ctx context.Context) error {
	patterns, err := s.repo.GetActive(ctx)
	if err != nil {
		return err
	}
	return s.refreshClassifier(patterns)
}

type patternSeed struct {
	InformationType string `json:"information_type"`
	Pattern         string `json:"pattern"`
	Description     string `json:"description"`
	Priority        int    `json:"priority"`
}

func loadPatternSeeds(path string) ([]patternSeed, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read patterns file: %w", err)
	}

	var seeds []patternSeed
	if err := json.Unmarshal(data, &seeds); err != nil {
		return nil, fmt.Errorf("failed to parse patterns file: %w", err)
	}

	return seeds, nil
}

