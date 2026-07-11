package domain

import (
	"errors"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		wrap  string
		match error
	}{
		{
			name:  "ErrNotFound is comparable",
			err:   ErrNotFound,
			match: ErrNotFound,
		},
		{
			name:  "ErrDuplicateRate is comparable",
			err:   ErrDuplicateRate,
			match: ErrDuplicateRate,
		},
		{
			name:  "ErrInvalidInput is comparable",
			err:   ErrInvalidInput,
			match: ErrInvalidInput,
		},
		{
			name:  "ErrDatabase is comparable",
			err:   ErrDatabase,
			match: ErrDatabase,
		},
		{
			name:  "ErrScrapeFailed is comparable",
			err:   ErrScrapeFailed,
			match: ErrScrapeFailed,
		},
		{
			name:  "ErrParseFailed is comparable",
			err:   ErrParseFailed,
			match: ErrParseFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.match) {
				t.Errorf("expected errors.Is to return true for %v", tt.err)
			}
		})
	}
}

func TestSentinelErrorsAreDistinct(t *testing.T) {
	errs := []error{ErrNotFound, ErrDuplicateRate, ErrInvalidInput, ErrDatabase, ErrScrapeFailed, ErrParseFailed}

	for i := 0; i < len(errs); i++ {
		for j := i + 1; j < len(errs); j++ {
			if errors.Is(errs[i], errs[j]) {
				t.Errorf("errors should be distinct: %v is same as %v", errs[i], errs[j])
			}
		}
	}
}

func TestSentinelErrorsCanBeWrapped(t *testing.T) {
	wrapped := errors.New("db query failed: " + ErrDatabase.Error())
	// sentinel errors should match when wrapped with errors.Is
	err := errors.Join(ErrDatabase, wrapped)
	if !errors.Is(err, ErrDatabase) {
		t.Errorf("wrapped error should match ErrDatabase via errors.Is")
	}
}
