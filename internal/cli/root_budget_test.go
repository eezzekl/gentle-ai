package cli

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/sdd"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/planner"
	"github.com/gentleman-programming/gentle-ai/v2/internal/system"
	"github.com/gentleman-programming/gentle-ai/v2/internal/verify"
)

// writeRootPrompt seeds the shared ~/.gemini/GEMINI.md with exactly size bytes.
func writeRootPrompt(t *testing.T, homeDir string, size int) string {
	t.Helper()
	path := filepath.Join(homeDir, ".gemini", "GEMINI.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(path, []byte(strings.Repeat("x", size)), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}
	return path
}

func sharedRootAdapters(t *testing.T) []agents.Adapter {
	t.Helper()
	return resolveAdapters([]model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity})
}

// realRootBudgetWarning returns the warning production actually emits for an
// oversized root, for the renderer tests to carry.
//
// The renderers append warnings verbatim and add no label of their own, so a
// hand-written fixture only ever proves that the renderer echoes whatever it was
// handed. Producing the real text makes the renderer tests fail if the warning
// ever loses the WARNING prefix or the numbers it is supposed to name.
func realRootBudgetWarning(t *testing.T) string {
	t.Helper()

	home := t.TempDir()
	writeRootPrompt(t, home, 12_500)

	warnings := rootBudgetWarnings(home, "", ScopeGlobal, sharedRootAdapters(t))
	if len(warnings) != 1 {
		t.Fatalf("fixture setup: rootBudgetWarnings() = %d warnings, want exactly 1; got %v", len(warnings), warnings)
	}
	return warnings[0]
}

// TestRootBudgetWarningNamesCountAndLimit asserts the Byte-Budget Warning
// requirement: over budget the warning names the actual character count and the
// 12,000 limit. Without the real number the user cannot tell how much content
// Antigravity is dropping, which is the whole point of the warning.
func TestRootBudgetWarningNamesCountAndLimit(t *testing.T) {
	home := t.TempDir()
	path := writeRootPrompt(t, home, 12_500)

	warnings := rootBudgetWarnings(home, "", ScopeGlobal, sharedRootAdapters(t))
	if len(warnings) != 1 {
		t.Fatalf("rootBudgetWarnings() = %d warnings, want exactly 1 (the shared root is one file); got %v", len(warnings), warnings)
	}

	got := warnings[0]
	// The label is asserted here, at the only place that produces it: the
	// renderers append this text verbatim and add no label of their own.
	// Checked as a prefix rather than containment, because the warning also
	// carries a filesystem path, and a path is free to contain any word.
	if !strings.HasPrefix(got, "WARNING: ") {
		t.Fatalf("warning is not labelled as one; got:\n%s", got)
	}
	for _, want := range []string{path, "12500", "12,000"} {
		if !strings.Contains(got, want) {
			t.Fatalf("warning does not name %q; got:\n%s", want, got)
		}
	}
	if !strings.Contains(strings.ToLower(got), "truncat") {
		t.Fatalf("warning does not explain that content past the limit is truncated; got:\n%s", got)
	}
}

// TestFormatRootBudgetLimitTracksTheConstant keeps the number in the warning tied
// to the number the check enforces. The warning exists to tell the user which
// threshold was crossed, so a limit that drifts from its own message is worse
// than no message: it sends the user looking for the wrong budget.
func TestFormatRootBudgetLimitTracksTheConstant(t *testing.T) {
	got := formatRootBudgetLimit()

	if digits := strings.ReplaceAll(got, ",", ""); digits != strconv.Itoa(rootBudgetLimit) {
		t.Fatalf("formatRootBudgetLimit() = %q, whose digits %q are not rootBudgetLimit (%d)", got, digits, rootBudgetLimit)
	}
	if want := "12,000"; got != want {
		t.Fatalf("formatRootBudgetLimit() = %q, want %q for the current limit", got, want)
	}
}

// TestRootBudgetWarningAbsentUnderLimit is the other half of the requirement: a
// root inside the budget must stay silent, or the warning becomes noise users
// learn to ignore.
func TestRootBudgetWarningAbsentUnderLimit(t *testing.T) {
	home := t.TempDir()
	writeRootPrompt(t, home, 11_999)

	if warnings := rootBudgetWarnings(home, "", ScopeGlobal, sharedRootAdapters(t)); len(warnings) != 0 {
		t.Fatalf("rootBudgetWarnings() = %v, want none under the limit", warnings)
	}
}

// TestRootBudgetWarningIgnoresAgentsWithoutSharedRoot guards the scope: the
// 12,000-character truncation is an Antigravity property of the shared
// ~/.gemini/GEMINI.md, so agents with their own prompt file must never be
// measured against it.
func TestRootBudgetWarningIgnoresAgentsWithoutSharedRoot(t *testing.T) {
	home := t.TempDir()
	writeRootPrompt(t, home, 30_000)

	adapters := resolveAdapters([]model.AgentID{model.AgentClaudeCode})
	if warnings := rootBudgetWarnings(home, "", ScopeGlobal, adapters); len(warnings) != 0 {
		t.Fatalf("rootBudgetWarnings(claude-code) = %v, want none", warnings)
	}
}

// TestRootBudgetWarningAbsentWhenRootMissing covers the fresh machine: no root
// file yet is not an overflow and must not be reported as one.
func TestRootBudgetWarningAbsentWhenRootMissing(t *testing.T) {
	home := t.TempDir()

	if warnings := rootBudgetWarnings(home, "", ScopeGlobal, sharedRootAdapters(t)); len(warnings) != 0 {
		t.Fatalf("rootBudgetWarnings() = %v, want none when the root does not exist", warnings)
	}
}

// TestRenderSyncReportEmitsRootBudgetWarning asserts the warning actually
// reaches the user through the sync report, verbatim.
//
// Only the verbatim check belongs here. Two earlier attempts at a separate
// "the report says WARNING" assertion were both vacuous, and neither failure
// mode is obvious:
//
//   - a zero-valued Verify is not Ready, so RenderSyncReport appends the
//     post-sync verification block, which always prints "…, 0 warnings, …";
//   - t.TempDir() embeds the test's own name in the path, and this test is
//     named …RootBudgetWarning, so the path alone satisfies a case-insensitive
//     search for "warning" once it reaches the report.
//
// The label is a property of the text rootBudgetWarnings produces, not of this
// renderer, which appends whatever it is handed. It is asserted at the producer,
// in TestRootBudgetWarningNamesCountAndLimit. Carrying the real warning verbatim
// is the whole contract here, and Verify stays pinned Ready to keep the report
// down to what this function is responsible for.
func TestRenderSyncReportEmitsRootBudgetWarning(t *testing.T) {
	warning := realRootBudgetWarning(t)
	result := SyncResult{
		Agents:             []model.AgentID{model.AgentAntigravity},
		FilesChanged:       1,
		Verify:             verify.Report{Ready: true},
		RootBudgetWarnings: []string{warning},
	}

	report := RenderSyncReport(result)
	if !strings.Contains(report, warning) {
		t.Fatalf("sync report does not carry the budget warning verbatim;\nwant:\n%s\ngot:\n%s", warning, report)
	}
}

// TestRenderSyncReportNoOpEmitsRootBudgetWarning covers the case that matters
// most in practice: an idempotent re-sync changes nothing, so the no-op branch
// returns early — but an already-oversized root is exactly the state a user
// needs told about, and it would otherwise stay silent forever.
func TestRenderSyncReportNoOpEmitsRootBudgetWarning(t *testing.T) {
	warning := realRootBudgetWarning(t)
	result := SyncResult{
		Agents:             []model.AgentID{model.AgentAntigravity},
		NoOp:               true,
		RootBudgetWarnings: []string{warning},
	}

	if report := RenderSyncReport(result); !strings.Contains(report, warning) {
		t.Fatalf("no-op sync report dropped the budget warning;\nwant:\n%s\ngot:\n%s", warning, report)
	}
}

// TestRenderInstallManualActionsEmitsRootBudgetWarning asserts the install path
// surfaces the same warning, and that it does not depend on the unrelated Pi
// CodeGraph manual actions that previously owned this output.
func TestRenderInstallManualActionsEmitsRootBudgetWarning(t *testing.T) {
	warning := realRootBudgetWarning(t)
	result := InstallResult{RootBudgetWarnings: []string{warning}}

	got := RenderInstallManualActions(result)
	if !strings.Contains(got, warning) {
		t.Fatalf("install output does not carry the budget warning;\nwant:\n%s\ngot:\n%s", warning, got)
	}
}

// TestRenderInstallManualActionsSilentWithoutWarnings pins the pre-existing
// contract: no Pi actions and no overflow means no output at all.
func TestRenderInstallManualActionsSilentWithoutWarnings(t *testing.T) {
	if got := RenderInstallManualActions(InstallResult{}); got != "" {
		t.Fatalf("RenderInstallManualActions(empty) = %q, want empty", got)
	}
}

// TestRunSyncWithSelectionOverBudgetRootDoesNotFail asserts the requirement's
// hard edge: the check warns, it never fails the run.
func TestRunSyncWithSelectionOverBudgetRootDoesNotFail(t *testing.T) {
	home := t.TempDir()
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = func(name string) (string, error) { return "/usr/local/bin/" + name, nil }

	// Seed a root far past the limit, then let a real sync run over it.
	writeRootPrompt(t, home, 60_000)

	sel := model.Selection{
		Agents:     []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity},
		Components: []model.ComponentID{model.ComponentSDD},
		SDDMode:    model.SDDModeSingle,
	}

	result, err := RunSyncWithSelection(home, sel)
	if err != nil {
		t.Fatalf("RunSyncWithSelection() over an oversized root must not fail; error = %v", err)
	}
	if len(result.RootBudgetWarnings) == 0 {
		t.Fatalf("sync over a 60,000-character root reported no budget warning")
	}
}

// TestExecuteTUIInstallSurfacesRootBudgetWarning is the regression guard for the
// gap the R4 review lens corroborated: the byte-budget warning was computed only
// in RunInstall, the non-interactive path. ExecuteTUIInstall is what the
// interactive TUI calls, and it is the default entry point for a plain
// `gentle-ai` invocation — so the warning never reached the users most likely to
// hit an oversized root.
//
// ManualActions is the channel that already carries non-fatal completion notices
// to BOTH the CLI and the TUI completion renderers, and ExecuteTUIInstall already
// forwards Pi CodeGraph actions through it, so the warning rides the same path
// rather than inventing a second one.
func TestExecuteTUIInstallSurfacesRootBudgetWarning(t *testing.T) {
	home := t.TempDir()
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = func(name string) (string, error) { return "/usr/local/bin/" + name, nil }

	writeRootPrompt(t, home, 60_000)

	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity},
		Components: []model.ComponentID{model.ComponentSDD},
		Persona:    model.PersonaGentleman,
		SDDMode:    model.SDDModeSingle,
	}
	resolved := planner.ResolvedPlan{
		Agents:            selection.Agents,
		OrderedComponents: selection.Components,
	}

	result, _ := ExecuteTUIInstallWithBackgroundAndOrchestrator(home, selection, resolved, system.PlatformProfile{PackageManager: "brew"}, model.OpenCodeBackgroundAuto, model.PiBackgroundIntent(""), nil)

	var found string
	for _, action := range result.ManualActions {
		if strings.Contains(action, "12,000") {
			found = action
			break
		}
	}
	if found == "" {
		t.Fatalf("ExecuteTUIInstall did not surface the byte-budget warning through ManualActions; got %v", result.ManualActions)
	}
	if !strings.Contains(found, filepath.Join(home, ".gemini", "GEMINI.md")) {
		t.Fatalf("budget warning does not name the oversized root; got:\n%s", found)
	}
}

// TestExecuteTUIInstallSilentUnderBudget keeps the warning from becoming noise on
// the interactive path the same way the CLI path stays quiet.
func TestExecuteTUIInstallSilentUnderBudget(t *testing.T) {
	home := t.TempDir()
	restoreCommand := runCommand
	restoreLookPath := cmdLookPath
	t.Cleanup(func() {
		runCommand = restoreCommand
		cmdLookPath = restoreLookPath
	})
	runCommand = func(string, ...string) error { return nil }
	cmdLookPath = func(name string) (string, error) { return "/usr/local/bin/" + name, nil }

	// Pin the temp dir as the user home so the reference-file split actually
	// applies. Without this the orchestrator body stays inline — by design, since
	// the stub advertises an absolute home path — and the root is legitimately
	// over budget, which would make an "under budget" assertion untestable.
	sdd.SetUserHomeDirForTest(t, home)

	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentGeminiCLI},
		Components: []model.ComponentID{model.ComponentSDD},
		Persona:    model.PersonaGentleman,
		SDDMode:    model.SDDModeSingle,
	}
	resolved := planner.ResolvedPlan{
		Agents:            selection.Agents,
		OrderedComponents: selection.Components,
	}

	result, _ := ExecuteTUIInstallWithBackgroundAndOrchestrator(home, selection, resolved, system.PlatformProfile{PackageManager: "brew"}, model.OpenCodeBackgroundAuto, model.PiBackgroundIntent(""), nil)

	for _, action := range result.ManualActions {
		if strings.Contains(action, "12,000") {
			t.Fatalf("ExecuteTUIInstall warned about the byte budget on an under-budget root: %s", action)
		}
	}
}
