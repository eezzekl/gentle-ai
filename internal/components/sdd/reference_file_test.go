package sdd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/antigravity"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/gemini"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// Distinguishing first lines of the two per-agent orchestrator assets. The
// antigravity body mandates Mission Control define_subagent/invoke_subagent
// delegation the gemini body has no tools for, which is exactly why the two
// assets are not reconcilable into one shared body (design.md D4).
const (
	geminiOrchestratorHeading      = "# Agent Teams Lite — Orchestrator Rule for Gemini"
	antigravityOrchestratorHeading = "# Agent Teams Lite — Orchestrator Instructions (Antigravity)"
)

// sharedReferenceAdapters are the adapters that implement SharedReferenceLayout
// and therefore route the orchestrator body to ~/.gemini/references/ instead of
// inlining it into GEMINI.md (design.md D1).
func sharedReferenceAdapters() map[string]agents.Adapter {
	return map[string]agents.Adapter{
		"gemini":      gemini.NewAdapter(),
		"antigravity": antigravity.NewAdapter(),
	}
}

// newSharedReferenceHome returns a temp dir pinned as the user home, so the
// reference path the injector computes matches the home-anchored path the
// gate stub advertises.
func newSharedReferenceHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	SetUserHomeDirForTest(t, home)
	return home
}

func sddReferencePath(t *testing.T, adapter agents.Adapter, dir string) string {
	t.Helper()
	layout, ok := adapter.(SharedReferenceLayout)
	if !ok {
		t.Fatalf("adapter %T does not implement SharedReferenceLayout", adapter)
	}
	return filepath.Join(layout.ReferencesDir(dir), sddReferenceFileName)
}

// TestReferenceOrchestratorAgentPrefersAntigravity asserts the antigravity
// priority selection (design.md D4): both agents share one reference file, so
// the content must be deterministic across sync order. Antigravity wins because
// it is the agent that actually truncates, and its body names tools the generic
// gemini body never mentions.
func TestReferenceOrchestratorAgentPrefersAntigravity(t *testing.T) {
	tests := map[string]struct {
		adapter  model.AgentID
		selected []model.AgentID
		want     model.AgentID
	}{
		"gemini alone selected keeps the gemini asset": {
			adapter:  model.AgentGeminiCLI,
			selected: []model.AgentID{model.AgentGeminiCLI},
			want:     model.AgentGeminiCLI,
		},
		"both selected promotes antigravity for the gemini pass": {
			adapter:  model.AgentGeminiCLI,
			selected: []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity},
			want:     model.AgentAntigravity,
		},
		"both selected keeps antigravity for the antigravity pass": {
			adapter:  model.AgentAntigravity,
			selected: []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity},
			want:     model.AgentAntigravity,
		},
		"antigravity alone selected keeps antigravity": {
			adapter:  model.AgentAntigravity,
			selected: []model.AgentID{model.AgentAntigravity},
			want:     model.AgentAntigravity,
		},
		"antigravity selected elsewhere in the list still wins": {
			adapter:  model.AgentGeminiCLI,
			selected: []model.AgentID{model.AgentClaudeCode, model.AgentAntigravity, model.AgentGeminiCLI},
			want:     model.AgentAntigravity,
		},
		"unpopulated selection falls back to the injected adapter": {
			adapter:  model.AgentGeminiCLI,
			selected: nil,
			want:     model.AgentGeminiCLI,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := referenceOrchestratorAgent(tc.adapter, tc.selected)
			if got != tc.want {
				t.Fatalf("referenceOrchestratorAgent(%q, %v) = %q, want %q", tc.adapter, tc.selected, got, tc.want)
			}
		})
	}
}

// TestInjectWritesOrchestratorReferenceFile asserts the Orchestrator Reference
// File Emission requirement (sdd-orchestrator-assets spec): the per-agent body
// lands in ~/.gemini/references/sdd-orchestrator.md verbatim, with no section
// markers, and is reported in the injection result.
func TestInjectWritesOrchestratorReferenceFile(t *testing.T) {
	tests := map[string]struct {
		adapter agents.Adapter
		want    model.AgentID
	}{
		"gemini": {adapter: gemini.NewAdapter(), want: model.AgentGeminiCLI},
		"antigravity": {
			adapter: antigravity.NewAdapter(),
			want:    model.AgentAntigravity,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			result, err := Inject(home, tc.adapter, "")
			if err != nil {
				t.Fatalf("Inject(%s) error = %v", name, err)
			}

			refPath := sddReferencePath(t, tc.adapter, home)
			got := readFileForTest(t, refPath)

			if want := renderSDDOrchestratorAsset(tc.want); got != want {
				t.Fatalf("reference file content != renderSDDOrchestratorAsset(%q)\n--- got (%d bytes) ---\n%s", tc.want, len(got), got)
			}
			if strings.Contains(got, "gentle-ai:sdd-orchestrator") {
				t.Fatalf("reference file must not carry section markers; got:\n%s", got)
			}
			if !containsPath(result.Files, refPath) {
				t.Fatalf("Inject(%s) result.Files missing %q; got %v", name, refPath, result.Files)
			}
		})
	}
}

// TestInjectSelectsAntigravityAssetForBothAgents asserts design.md D4 end to
// end: with both agents selected, the gemini pass must not overwrite the shared
// reference file with the generic gemini body. Without this, sync order decides
// which orchestrator agy ends up obeying.
func TestInjectSelectsAntigravityAssetForBothAgents(t *testing.T) {
	selected := []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity}

	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			if _, err := Inject(home, adapter, "", InjectOptions{SelectedAgents: selected}); err != nil {
				t.Fatalf("Inject(%s) error = %v", name, err)
			}

			got := readFileForTest(t, sddReferencePath(t, adapter, home))
			if !strings.Contains(got, antigravityOrchestratorHeading) {
				t.Fatalf("reference file must carry the antigravity asset; got first line:\n%s", firstLine(got))
			}
			if strings.Contains(got, geminiOrchestratorHeading) {
				t.Fatalf("reference file must not carry the gemini asset; got first line:\n%s", firstLine(got))
			}
		})
	}
}

// TestInjectDoesNotDowngradeSharedAntigravityReference is the regression guard
// for the clobber the R1 review lens corroborated: SelectedAgents only reports
// the agents of the CURRENT pass, so a later `agy sync --agents gemini` on a
// machine where antigravity is also installed used to overwrite the shared
// reference file with the generic gemini body — stripping the Mission Control
// delegation antigravity depends on, for an agent that run never touched.
func TestInjectDoesNotDowngradeSharedAntigravityReference(t *testing.T) {
	adapter := gemini.NewAdapter()
	home := newSharedReferenceHome(t)

	// First pass: both agents selected, so antigravity's asset lands in the
	// shared file (design.md D4).
	both := []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity}
	if _, err := Inject(home, adapter, "", InjectOptions{SelectedAgents: both}); err != nil {
		t.Fatalf("Inject(first pass) error = %v", err)
	}

	refPath := sddReferencePath(t, adapter, home)
	if got := readFileForTest(t, refPath); !strings.Contains(got, antigravityOrchestratorHeading) {
		t.Fatalf("setup failed: shared reference does not hold the antigravity asset; got first line:\n%s", firstLine(got))
	}

	// Second pass: gemini alone. The shared file must survive intact.
	geminiOnly := []model.AgentID{model.AgentGeminiCLI}
	if _, err := Inject(home, adapter, "", InjectOptions{SelectedAgents: geminiOnly}); err != nil {
		t.Fatalf("Inject(gemini-only pass) error = %v", err)
	}

	got := readFileForTest(t, refPath)
	if !strings.Contains(got, antigravityOrchestratorHeading) {
		t.Fatalf("gemini-only sync downgraded the shared reference; got first line:\n%s", firstLine(got))
	}
	if strings.Contains(got, geminiOrchestratorHeading) {
		t.Fatalf("gemini-only sync overwrote the antigravity asset; got first line:\n%s", firstLine(got))
	}
	if want := renderSDDOrchestratorAsset(model.AgentAntigravity); got != want {
		t.Fatalf("shared reference is no longer the verbatim antigravity asset (%d bytes)", len(got))
	}
}

// TestPreservedOrchestratorAgentKeepsAntigravityOnDisk pins the decision rule
// itself: the priority is resolved from what the shared file already holds, not
// only from the current selection.
func TestPreservedOrchestratorAgentKeepsAntigravityOnDisk(t *testing.T) {
	tests := map[string]struct {
		selected model.AgentID
		existing string
		want     model.AgentID
	}{
		"gemini pass over an antigravity file keeps antigravity": {
			selected: model.AgentGeminiCLI,
			existing: renderSDDOrchestratorAsset(model.AgentAntigravity),
			want:     model.AgentAntigravity,
		},
		"gemini pass over a gemini file stays gemini": {
			selected: model.AgentGeminiCLI,
			existing: renderSDDOrchestratorAsset(model.AgentGeminiCLI),
			want:     model.AgentGeminiCLI,
		},
		"gemini pass with no file yet stays gemini": {
			selected: model.AgentGeminiCLI,
			existing: "",
			want:     model.AgentGeminiCLI,
		},
		"gemini pass over unrecognized content stays gemini": {
			selected: model.AgentGeminiCLI,
			existing: "# hand-edited by the user\n",
			want:     model.AgentGeminiCLI,
		},
		"antigravity pass is unaffected": {
			selected: model.AgentAntigravity,
			existing: renderSDDOrchestratorAsset(model.AgentGeminiCLI),
			want:     model.AgentAntigravity,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := preservedOrchestratorAgent(tc.selected, tc.existing)
			if got != tc.want {
				t.Fatalf("preservedOrchestratorAgent(%q, <%d bytes>) = %q, want %q", tc.selected, len(tc.existing), got, tc.want)
			}
		})
	}
}

// TestInjectReplacesRootOrchestratorWithGateStub asserts that the root system
// prompt carries only the pointer stub under the existing sdd-orchestrator
// marker — the full body must no longer be inlined (Root Assembly Byte Budget).
func TestInjectReplacesRootOrchestratorWithGateStub(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			if _, err := Inject(home, adapter, ""); err != nil {
				t.Fatalf("Inject(%s) error = %v", name, err)
			}

			root := readFileForTest(t, adapter.SystemPromptFile(home))

			if !strings.Contains(root, "<!-- gentle-ai:sdd-orchestrator -->") {
				t.Fatalf("root prompt missing sdd-orchestrator open marker; got:\n%s", root)
			}
			if !strings.Contains(root, renderSDDGateStub()) {
				t.Fatalf("root prompt missing SDD gate stub; got:\n%s", root)
			}
			if strings.Contains(root, "## Agent Teams Orchestrator") {
				t.Fatalf("root prompt still inlines the full orchestrator body; got:\n%s", root)
			}
		})
	}
}

// TestInjectKeepsOrchestratorInlineWhenTargetIsNotHome is the regression guard
// for the workspace-scoped install. `agy install --scope=workspace` passes the
// workspace, not the home dir, as the injection target. The stub names an
// absolute home-anchored path and tells the agent NOT to resolve it against the
// project, so splitting there would write the body somewhere the pointer never
// sends the agent and the whole orchestrator contract would be silently lost.
// The inline body — which always reaches the agent — must survive instead.
func TestInjectKeepsOrchestratorInlineWhenTargetIsNotHome(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			workspace := t.TempDir()
			SetUserHomeDirForTest(t, home)

			if _, err := Inject(workspace, adapter, ""); err != nil {
				t.Fatalf("Inject(%s, workspace) error = %v", name, err)
			}

			root := readFileForTest(t, adapter.SystemPromptFile(workspace))
			if !strings.Contains(root, "## Agent Teams Orchestrator") {
				t.Fatalf("workspace-scoped root must keep the full inline orchestrator; got:\n%s", root)
			}
			if strings.Contains(root, renderSDDGateStub()) {
				t.Fatalf("workspace-scoped root must not carry the home-anchored stub; got:\n%s", root)
			}

			// No reference file may be written under either directory: the one
			// under the workspace would be unreachable, and the one under home
			// would be a file this install was never asked to touch.
			for label, dir := range map[string]string{"workspace": workspace, "home": home} {
				if _, err := os.Stat(sddReferencePath(t, adapter, dir)); !os.IsNotExist(err) {
					t.Fatalf("unexpected reference file under %s (stat err = %v)", label, err)
				}
			}
		})
	}
}

// TestInjectOrchestratorReferenceIsIdempotent asserts design.md D3: a second
// sync over an already-converted layout produces byte-identical output.
func TestInjectOrchestratorReferenceIsIdempotent(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			refPath := sddReferencePath(t, adapter, home)
			rootPath := adapter.SystemPromptFile(home)

			if _, err := Inject(home, adapter, ""); err != nil {
				t.Fatalf("Inject(%s) first pass error = %v", name, err)
			}
			firstRef := readFileForTest(t, refPath)
			firstRoot := readFileForTest(t, rootPath)

			if _, err := Inject(home, adapter, ""); err != nil {
				t.Fatalf("Inject(%s) second pass error = %v", name, err)
			}

			if got := readFileForTest(t, refPath); got != firstRef {
				t.Fatalf("second Inject(%s) changed the reference file", name)
			}
			if got := readFileForTest(t, rootPath); got != firstRoot {
				t.Fatalf("second Inject(%s) changed the root prompt:\n%s", name, got)
			}
		})
	}
}

// TestInjectMigratesLegacyInlineOrchestrator covers the install that already
// carries the full body under the marker: the converted layout must replace it,
// not append a second copy or leave the oversized body behind.
func TestInjectMigratesLegacyInlineOrchestrator(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			rootPath := adapter.SystemPromptFile(home)
			legacy := "<!-- gentle-ai:sdd-orchestrator -->\n" +
				renderSDDOrchestratorAsset(adapter.Agent()) +
				"\n<!-- /gentle-ai:sdd-orchestrator -->\n"
			if err := os.MkdirAll(filepath.Dir(rootPath), 0o755); err != nil {
				t.Fatalf("MkdirAll error = %v", err)
			}
			if err := os.WriteFile(rootPath, []byte(legacy), 0o644); err != nil {
				t.Fatalf("WriteFile error = %v", err)
			}

			if _, err := Inject(home, adapter, ""); err != nil {
				t.Fatalf("Inject(%s) error = %v", name, err)
			}

			root := readFileForTest(t, rootPath)
			if strings.Contains(root, "## Agent Teams Orchestrator") {
				t.Fatalf("legacy inline body survived migration; got:\n%s", root)
			}
			if !strings.Contains(root, renderSDDGateStub()) {
				t.Fatalf("migrated root missing SDD gate stub; got:\n%s", root)
			}
			if n := strings.Count(root, "<!-- gentle-ai:sdd-orchestrator -->"); n != 1 {
				t.Fatalf("migrated root has %d sdd-orchestrator markers, want 1; got:\n%s", n, root)
			}
		})
	}
}

func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(data)
}

func firstLine(s string) string {
	if idx := strings.Index(s, "\n"); idx >= 0 {
		return s[:idx]
	}
	return s
}
