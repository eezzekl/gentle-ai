package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/components/sdd"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// TestRunSyncWithSelection_SharedOrchestratorReferencePrefersAntigravity is the
// wiring guard for design.md D4. gemini-cli and antigravity share one
// ~/.gemini/references/sdd-orchestrator.md, so unless the sync pass reports the
// full agent selection to the SDD injector, whichever agent syncs last decides
// which orchestrator agy obeys — and the gemini body names none of the Mission
// Control delegation tools the antigravity body mandates.
func TestRunSyncWithSelection_SharedOrchestratorReferencePrefersAntigravity(t *testing.T) {
	home := t.TempDir()
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})

	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = func(name string) (string, error) { return "/usr/local/bin/" + name, nil }

	// The split only applies when the injection target is the real home dir.
	sdd.SetUserHomeDirForTest(t, home)

	sel := model.Selection{
		Agents:     []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity},
		Components: []model.ComponentID{model.ComponentSDD},
		SDDMode:    model.SDDModeSingle,
	}

	if _, err := RunSyncWithSelection(home, sel); err != nil {
		t.Fatalf("RunSyncWithSelection() error = %v", err)
	}

	referencePath := filepath.Join(home, ".gemini", "references", "sdd-orchestrator.md")
	body, err := os.ReadFile(referencePath)
	if err != nil {
		t.Fatalf("read %q: %v", referencePath, err)
	}

	got := string(body)
	if !strings.Contains(got, "# Agent Teams Lite — Orchestrator Instructions (Antigravity)") {
		t.Fatalf("shared orchestrator reference must carry the antigravity asset; got first line:\n%s", strings.SplitN(got, "\n", 2)[0])
	}
	if !strings.Contains(got, "define_subagent") {
		t.Fatalf("antigravity asset must keep its Mission Control delegation tools; got:\n%s", got)
	}
}
