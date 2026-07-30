package cli

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// TestNormalizePersonaTable is a table-driven test covering all selectable
// personas and back-compat aliases. This supersedes the older single-case test.
func TestNormalizePersonaTable(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  model.PersonaID
		wantErr bool
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
			name:    "gentleman-neutral-artifacts back-compat alias maps to neutral (Decision 5)",
			input:   "gentleman-neutral-artifacts",
			wantID:  model.PersonaNeutral,
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
// TestNormalizePersonaAliasesDoNotShareATarget inverts what this test asserted
// before design Decision 5. The two legacy aliases MUST resolve differently:
// `gentleman` keeps the teaching voice, while `gentleman-neutral-artifacts`
// resolves to `neutral`, because issue #1702 defect 1 documents that the alias
// emitted full gentleman behavior despite its name. Convergence here would carry
// that defect forward and would also desynchronize validation from the migration
// matrix in sync.go.
func TestNormalizePersonaAliasesDoNotShareATarget(t *testing.T) {
	gentleman, err := normalizePersona("gentleman")
	if err != nil {
		t.Fatalf("normalizePersona(gentleman) error = %v", err)
	}
	hybrid, err := normalizePersona("gentleman-neutral-artifacts")
	if err != nil {
		t.Fatalf("normalizePersona(gentleman-neutral-artifacts) error = %v", err)
	}

	if gentleman != model.PersonaGentle {
		t.Errorf("normalizePersona(gentleman) = %q, want %q", gentleman, model.PersonaGentle)
	}
	if hybrid != model.PersonaNeutral {
		t.Errorf("normalizePersona(gentleman-neutral-artifacts) = %q, want %q", hybrid, model.PersonaNeutral)
	}
	if gentleman == hybrid {
		t.Errorf("the two legacy aliases must NOT share a target; both resolved to %q", gentleman)
	}
}
