package sdd

import (
	"strings"
	"testing"
)

// TestRenderSDDGateStubWordingInvariants asserts the wording invariants
// required by the SDD-Stub Block Wording spec (gemini-prompt-modularization)
// and design.md D5: the stub must list SDD commands, imperatively instruct
// reading the sdd-orchestrator reference file before any SDD command or
// natural-language equivalent, and contain no orchestration detail.
func TestRenderSDDGateStubWordingInvariants(t *testing.T) {
	got := renderSDDGateStub()

	for _, want := range []string{
		"/sdd-apply",
		"/sdd-verify",
		"~/.gemini/references/sdd-orchestrator.md",
		"before handling any SDD command or request",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderSDDGateStub() missing %q; got:\n%s", want, got)
		}
	}
}

// TestRenderSDDGateStubHasNoOrchestrationDetail guards against the stub
// growing back into a full orchestrator body (defeats the reference-file
// split) by asserting phase/delegation vocabulary specific to the full
// orchestrator asset is absent from the short gate stub.
func TestRenderSDDGateStubHasNoOrchestrationDetail(t *testing.T) {
	got := renderSDDGateStub()

	for _, detail := range []string{
		"Phase 1", "sub-agent", "delegate", "tasks.md", "test-marker",
	} {
		if strings.Contains(got, detail) {
			t.Fatalf("renderSDDGateStub() must not contain orchestration detail %q; got:\n%s", detail, got)
		}
	}
}
