package cli

import (
	"context"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestAntigravityCollisionCheckIncludesGeminiCLI(t *testing.T) {
	checks := antigravityCollisionCheck([]model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity})
	if len(checks) != 1 {
		t.Fatalf("antigravityCollisionCheck() len = %d, want 1", len(checks))
	}

	check := checks[0]
	if !check.Soft {
		t.Fatal("antigravityCollisionCheck() should return a soft warning")
	}
	if check.ID != "verify:antigravity:rules-collision" {
		t.Fatalf("check ID = %q, want verify:antigravity:rules-collision", check.ID)
	}

	err := check.Run(context.Background())
	if err == nil {
		t.Fatal("check.Run() error = nil, want warning message")
	}
	message := err.Error()
	for _, want := range []string{
		"Antigravity intentionally uses the Gemini-compatible global prompt surface",
		// The winner is fixed by referenceOrchestratorAgent's antigravity priority,
		// not by sync order. Pinning the phrase keeps the operator-facing message
		// from drifting back to the last-writer-wins wording it used to carry,
		// which told operators to control the outcome with something that does
		// not control it.
		"whenever Antigravity is selected it owns the shared gentle-ai:sdd-orchestrator section, regardless of which agent synced last",
		"Prefer Antigravity for new installs",
		"~/.gemini/references/",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("warning message missing %q; got:\n%s", want, message)
		}
	}
}

func TestAntigravityCollisionCheckNoWarningWithoutGemini(t *testing.T) {
	checks := antigravityCollisionCheck([]model.AgentID{model.AgentAntigravity})
	if len(checks) != 0 {
		t.Fatalf("antigravityCollisionCheck() len = %d, want 0", len(checks))
	}
}
