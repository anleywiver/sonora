// Package ingestfilter implements per-source ingest filter rules (Sprint
// 14 sisipan, ADR 0008) — genre allow-lists and year ranges that apply
// ONLY to auto-ingest sources (bandcamp/cloud_sync). manual_upload is
// never filtered; see CLAUDE.md's legal-ingest constraint and ADR 0008.
package ingestfilter

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"sonora.dev/go-core/infrastructure/postgres/sqlc"
)

var ErrNotFound = errors.New("ingestfilter: rule not found")

type Rule struct {
	ID         uuid.UUID
	SourceType string
	RuleType   string
	Value      string
}

type Service struct {
	q *sqlc.Queries
}

func NewService(q *sqlc.Queries) *Service {
	return &Service{q: q}
}

func (s *Service) ListRules(ctx context.Context, sourceType string) ([]Rule, error) {
	rows, err := s.q.ListIngestFilterRules(ctx, sourceType)
	if err != nil {
		return nil, fmt.Errorf("ingestfilter: list rules: %w", err)
	}
	out := make([]Rule, 0, len(rows))
	for _, row := range rows {
		out = append(out, Rule{
			ID:         uuid.UUID(row.ID.Bytes),
			SourceType: row.SourceType,
			RuleType:   row.RuleType,
			Value:      row.Value,
		})
	}
	return out, nil
}

func (s *Service) CreateRule(ctx context.Context, sourceType, ruleType, value string) (*Rule, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("ingestfilter: generate id: %w", err)
	}
	row, err := s.q.CreateIngestFilterRule(ctx, sqlc.CreateIngestFilterRuleParams{
		ID:         toPgUUID(id),
		SourceType: sourceType,
		RuleType:   ruleType,
		Value:      value,
	})
	if err != nil {
		return nil, fmt.Errorf("ingestfilter: create rule: %w", err)
	}
	return &Rule{ID: uuid.UUID(row.ID.Bytes), SourceType: row.SourceType, RuleType: row.RuleType, Value: row.Value}, nil
}

func (s *Service) DeleteRule(ctx context.Context, sourceType string, id uuid.UUID) error {
	affected, err := s.q.DeleteIngestFilterRule(ctx, sqlc.DeleteIngestFilterRuleParams{
		ID:         toPgUUID(id),
		SourceType: sourceType,
	})
	if err != nil {
		return fmt.Errorf("ingestfilter: delete rule: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// Check decides whether a downloaded item's genre/year clears the rules
// configured for sourceType. Missing metadata (genre == "" or year ==
// nil) never fails a rule — see ADR 0008's "fail-open" rationale: absent
// ID3 tags are common and shouldn't silently cost the user real music.
func (s *Service) Check(ctx context.Context, sourceType, genre string, year *int) (pass bool, reason string, err error) {
	rules, err := s.ListRules(ctx, sourceType)
	if err != nil {
		return false, "", err
	}

	var allowedGenres []string
	var yearMin, yearMax *int
	for _, rule := range rules {
		switch rule.RuleType {
		case "genre_allow":
			allowedGenres = append(allowedGenres, rule.Value)
		case "year_min":
			if n, err := strconv.Atoi(rule.Value); err == nil {
				yearMin = &n
			}
		case "year_max":
			if n, err := strconv.Atoi(rule.Value); err == nil {
				yearMax = &n
			}
		}
	}

	if len(allowedGenres) > 0 && genre != "" {
		matched := false
		for _, allowed := range allowedGenres {
			if strings.EqualFold(allowed, genre) {
				matched = true
				break
			}
		}
		if !matched {
			return false, fmt.Sprintf("genre %q is not in the allow-list for %s", genre, sourceType), nil
		}
	}

	if year != nil {
		if yearMin != nil && *year < *yearMin {
			return false, fmt.Sprintf("year %d is before the minimum %d for %s", *year, *yearMin, sourceType), nil
		}
		if yearMax != nil && *year > *yearMax {
			return false, fmt.Sprintf("year %d is after the maximum %d for %s", *year, *yearMax, sourceType), nil
		}
	}

	return true, "", nil
}
