package classifier //nolint:stylecheck	

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"database-classifier/internal/domain"
)

type Pattern struct {
	InformationType domain.InformationType `json:"information_type"`
	Pattern         string                 `json:"pattern"`
	Description     string                 `json:"description"`
	Priority        int                    `json:"priority"`
	regex           *regexp.Regexp
}

type Classifier struct {
	patterns []Pattern
}

type MatchResult struct {
	InformationType domain.InformationType
	ConfidenceScore float64
	MatchedPatterns []string
}

func NewClassifier(patterns []*domain.ClassificationPattern) (*Classifier, error) {
	c := &Classifier{}
	if err := c.SetPatterns(patterns); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Classifier) SetPatterns(patterns []*domain.ClassificationPattern) error {
	compiled := make([]Pattern, 0, len(patterns))
	for _, p := range patterns {
		regex, err := regexp.Compile(p.Pattern)
		if err != nil {
			return fmt.Errorf("failed to compile regex pattern '%s': %w", p.Pattern, err)
		}
		compiled = append(compiled, Pattern{
			InformationType: p.InformationType,
			Pattern:         p.Pattern,
			Description:     p.Description,
			Priority:        p.Priority,
			regex:           regex,
		})
	}

	sort.Slice(compiled, func(i, j int) bool {
		return compiled[i].Priority > compiled[j].Priority
	})

	c.patterns = compiled
	return nil
}

func (c *Classifier) ClassifyColumn(columnName string) MatchResult {
	if columnName == "" {
		return MatchResult{
			InformationType: domain.InfoTypeNA,
			ConfidenceScore: 0.0,
			MatchedPatterns: []string{},
		}
	}

	var matches []struct {
		pattern Pattern
		score   float64
	}

	cleanName := strings.ToLower(strings.TrimSpace(columnName))

	for _, pattern := range c.patterns {
		if pattern.regex.MatchString(cleanName) {
			score := c.calculateConfidenceScore(cleanName, pattern)
			matches = append(matches, struct {
				pattern Pattern
				score   float64
			}{pattern, score})
		}
	}

	if len(matches) == 0 {
		return MatchResult{
			InformationType: domain.InfoTypeNA,
			ConfidenceScore: 0.0,
			MatchedPatterns: []string{},
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	bestMatch := matches[0]
	matchedPatterns := make([]string, len(matches))
	for i, match := range matches {
		matchedPatterns[i] = match.pattern.Pattern
	}

	return MatchResult{
		InformationType: bestMatch.pattern.InformationType,
		ConfidenceScore: bestMatch.score,
		MatchedPatterns: matchedPatterns,
	}
}

func (c *Classifier) calculateConfidenceScore(columnName string, pattern Pattern) float64 {
	baseScore := float64(pattern.Priority) / 100.0

	exactMatch := 0.0
	if pattern.regex.MatchString(columnName) {
		match := pattern.regex.FindString(columnName)
		if match == columnName {
			exactMatch = 0.2
		} else {
			exactMatch = 0.1
		}
	}

	commonWordsPenalty := 0.0
	commonWords := []string{"id", "name", "number", "date", "time", "status", "type"}
	for _, word := range commonWords {
		if strings.Contains(columnName, word) && len(columnName) < 10 {
			commonWordsPenalty = 0.1
			break
		}
	}

	finalScore := baseScore + exactMatch - commonWordsPenalty

	if finalScore > 1.0 {
		finalScore = 1.0
	}
	if finalScore < 0.0 {
		finalScore = 0.0
	}

	return finalScore
}

func (c *Classifier) AddPattern(pattern Pattern) error {
	regex, err := regexp.Compile(pattern.Pattern)
	if err != nil {
		return fmt.Errorf("failed to compile regex pattern '%s': %w", pattern.Pattern, err)
	}
	pattern.regex = regex

	c.patterns = append(c.patterns, pattern)

	sort.Slice(c.patterns, func(i, j int) bool {
		return c.patterns[i].Priority > c.patterns[j].Priority
	})

	return nil
}

func (c *Classifier) GetPatterns() []Pattern {
	return c.patterns
}

func (c *Classifier) RemovePattern(infoType domain.InformationType, patternStr string) {
	newPatterns := make([]Pattern, 0, len(c.patterns))
	for _, p := range c.patterns {
		if p.InformationType != infoType || p.Pattern != patternStr {
			newPatterns = append(newPatterns, p)
		}
	}
	c.patterns = newPatterns
}
