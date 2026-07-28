package engram

import (
	"strings"
	"testing"
)

// TestRenderSessionBootstrapStubWordingInvariants asserts the wording
// invariants required by the Session-Bootstrap Block Wording spec
// (gemini-prompt-modularization) and design.md D5: the stub must
// imperatively instruct a file-read of the absolute, `~`-anchored
// engram-protocol reference path, clarify that path is not resolved
// against the workspace, and instruct treating its content as active
// session instructions.
func TestRenderSessionBootstrapStubWordingInvariants(t *testing.T) {
	got := renderSessionBootstrapStub()

	for _, want := range []string{
		"with your file-read tool",
		"~/.gemini/references/engram-protocol.md",
		"NOT the workspace",
		"active session instructions",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderSessionBootstrapStub() missing %q; got:\n%s", want, got)
		}
	}
}

// TestRenderSessionBootstrapStubHasNoTestMarkerText guards against shipping
// placeholder/test-marker content in the rendered stub (Reference File
// Generation spec: "Shipped output MUST NOT contain test-marker blocks").
func TestRenderSessionBootstrapStubHasNoTestMarkerText(t *testing.T) {
	got := renderSessionBootstrapStub()

	for _, marker := range []string{"test-marker", "TEST-MARKER", "TESTMARKER", "TODO"} {
		if strings.Contains(got, marker) {
			t.Fatalf("renderSessionBootstrapStub() must not contain test-marker text %q; got:\n%s", marker, got)
		}
	}
}
