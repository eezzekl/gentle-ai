package model

import (
	"testing"
)

// TestSelectionHasRegionAndArtifactsInEnglish verifies that Selection has the
// new Region and ArtifactsInEnglish fields for the language/region axis.
func TestSelectionHasRegionAndArtifactsInEnglish(t *testing.T) {
	s := Selection{}

	// Region defaults to empty string.
	if s.Region != "" {
		t.Fatalf("Selection.Region default = %q, want empty", s.Region)
	}
	s.Region = string(RegionArgentina)
	if s.Region != "argentina" {
		t.Fatalf("Selection.Region set to %q but read back as %q", "argentina", s.Region)
	}

	// ArtifactsInEnglish defaults to false (Go zero value).
	if s.ArtifactsInEnglish {
		t.Fatal("Selection.ArtifactsInEnglish default = true, want false")
	}
	s.ArtifactsInEnglish = true
	if !s.ArtifactsInEnglish {
		t.Fatal("Selection.ArtifactsInEnglish set to true but read back as false")
	}
}

// TestSelectionHasStrictTDDField verifies that the Selection struct has a
// StrictTDD bool field.
func TestSelectionHasStrictTDDField(t *testing.T) {
	s := Selection{}
	// Field must be accessible and default to false.
	if s.StrictTDD {
		t.Fatal("Selection.StrictTDD default = true, want false")
	}

	s.StrictTDD = true
	if !s.StrictTDD {
		t.Fatal("Selection.StrictTDD set to true but read back as false")
	}
}

// TestSyncOverridesHasStrictTDDPointer verifies that SyncOverrides has a
// *bool StrictTDD field (nil = no override semantics).
func TestSyncOverridesHasStrictTDDPointer(t *testing.T) {
	o := SyncOverrides{}
	// Nil means "no override".
	if o.StrictTDD != nil {
		t.Fatal("SyncOverrides.StrictTDD default = non-nil, want nil")
	}

	enabled := true
	o.StrictTDD = &enabled
	if o.StrictTDD == nil || !*o.StrictTDD {
		t.Fatal("SyncOverrides.StrictTDD pointer set to true but read back incorrectly")
	}

	disabled := false
	o.StrictTDD = &disabled
	if o.StrictTDD == nil || *o.StrictTDD {
		t.Fatal("SyncOverrides.StrictTDD pointer set to false but read back incorrectly")
	}
}

// TestSelectionHasCodexModelAssignments verifies that the Selection struct has a
// CodexModelAssignments map field.
func TestSelectionHasCodexModelAssignments(t *testing.T) {
	s := Selection{}
	// Zero value is nil.
	if s.CodexModelAssignments != nil {
		t.Fatal("Selection.CodexModelAssignments zero value should be nil")
	}

	s.CodexModelAssignments = map[string]CodexEffort{"sdd-apply": CodexEffortHigh}
	if s.CodexModelAssignments["sdd-apply"] != CodexEffortHigh {
		t.Fatal("Selection.CodexModelAssignments not accessible after assignment")
	}
}

// TestSyncOverridesCodexModelPreset verifies that SyncOverrides has a
// CodexModelAssignments map field (nil = no override semantics).
func TestSyncOverridesCodexModelPreset(t *testing.T) {
	o := SyncOverrides{}
	if o.CodexModelAssignments != nil {
		t.Fatal("SyncOverrides.CodexModelAssignments zero value should be nil")
	}

	o.CodexModelAssignments = map[string]CodexEffort{"default": CodexEffortMedium}
	if o.CodexModelAssignments == nil {
		t.Fatal("SyncOverrides.CodexModelAssignments should be non-nil after assignment")
	}
}
