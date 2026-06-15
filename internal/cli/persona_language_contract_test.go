package cli

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// TestNormalizePersonaTable is a table-driven test covering all selectable
// personas and back-compat aliases. This supersedes the older single-case test.
func TestNormalizePersonaTable(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantID    model.PersonaID
		wantErr   bool
	}{
		{
			name:    "gentle is accepted",
			input:   "gentle",
			wantID:  model.PersonaGentle,
			wantErr: false,
		},
		{
			name:    "neutral is accepted",
			input:   "neutral",
			wantID:  model.PersonaNeutral,
			wantErr: false,
		},
		{
			name:    "custom is accepted",
			input:   "custom",
			wantID:  model.PersonaCustom,
			wantErr: false,
		},
		{
			name:    "gentleman back-compat alias maps to gentle",
			input:   "gentleman",
			wantID:  model.PersonaGentle,
			wantErr: false,
		},
		{
			name:    "gentleman-neutral-artifacts back-compat alias maps to gentle (migration convergence)",
			input:   "gentleman-neutral-artifacts",
			wantID:  model.PersonaGentle,
			wantErr: false,
		},
		{
			name:    "empty string defaults to gentle",
			input:   "",
			wantID:  model.PersonaGentle,
			wantErr: false,
		},
		{
			name:    "unknown value returns error",
			input:   "unknown-value",
			wantID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizePersona(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizePersona(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.wantID {
				t.Errorf("normalizePersona(%q) = %q, want %q", tt.input, got, tt.wantID)
			}
			// Unknown values must NOT resolve to a regional-voice persona.
			if tt.wantErr && got == model.PersonaGentle {
				t.Errorf("normalizePersona(%q) resolved to %q for unknown value — must not inject regional voice", tt.input, got)
			}
		})
	}
}

// TestNormalizePersonaBothGentlemanAliasesConverge verifies that "gentleman" and
// "gentleman-neutral-artifacts" normalize to byte-identical values, proving that
// the hybrid was always a no-op and migration convergence is correct.
func TestNormalizePersonaBothGentlemanAliasesConverge(t *testing.T) {
	a, err := normalizePersona("gentleman")
	if err != nil {
		t.Fatalf("normalizePersona(gentleman) error = %v", err)
	}
	b, err := normalizePersona("gentleman-neutral-artifacts")
	if err != nil {
		t.Fatalf("normalizePersona(gentleman-neutral-artifacts) error = %v", err)
	}
	if a != b {
		t.Errorf("migration convergence failed: gentleman → %q, gentleman-neutral-artifacts → %q; must be identical", a, b)
	}
}
