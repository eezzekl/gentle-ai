package cli

import (
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/components/engram"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/sdd"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// These are the install/sync parity guards for design.md D9. The two reference
// files this change introduces are new write targets, so they must appear in
// componentPathsWithWorkspaceScoped — the single function backupTargets,
// syncBackupTargets, post-apply verification, and post-sync verification all
// read. Without that, a sync overwrites them with no snapshot to roll back to
// and no verification catches a broken write.

// referenceParitySelection selects the two agents that share
// ~/.gemini/references and both components that own a reference file.
func referenceParitySelection() model.Selection {
	return model.Selection{
		Agents:     []model.AgentID{model.AgentGeminiCLI, model.AgentAntigravity},
		Components: []model.ComponentID{model.ComponentEngram, model.ComponentSDD},
		Persona:    model.PersonaGentleman,
		SDDMode:    model.SDDModeSingle,
	}
}

// pinReferenceParityHome pins the home both components resolve the split
// against, so the reference layout applies instead of falling back to inline.
func pinReferenceParityHome(t *testing.T, home string) {
	t.Helper()
	engram.SetUserHomeDirForTest(t, home)
	sdd.SetUserHomeDirForTest(t, home)
}

func referenceParityPaths(home string) (engramPath, sddPath string) {
	dir := filepath.Join(home, ".gemini", "references")
	return filepath.Join(dir, "engram-protocol.md"), filepath.Join(dir, "sdd-orchestrator.md")
}

func TestComponentPathsIncludeSharedReferenceFiles(t *testing.T) {
	home := t.TempDir()
	pinReferenceParityHome(t, home)

	selection := referenceParitySelection()
	adapters := resolveAdapters(selection.Agents)
	engramPath, sddPath := referenceParityPaths(home)

	cases := map[string]struct {
		component model.ComponentID
		want      string
	}{
		"engram": {model.ComponentEngram, engramPath},
		"sdd":    {model.ComponentSDD, sddPath},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := componentPathsWithWorkspaceScoped(home, "", ScopeGlobal, selection, adapters, tc.component)
			if !containsPath(got, tc.want) {
				t.Fatalf("componentPathsWithWorkspaceScoped(%s) must include the reference file %q; got:\n%v", name, tc.want, got)
			}
		})
	}
}

// TestComponentPathsOmitReferenceFilesUnderWorkspaceScope is the other half of
// the guard. A workspace-scoped install keeps the body inline and writes no
// reference file, so listing one would make post-apply verification fail on a
// file nothing was ever asked to write.
func TestComponentPathsOmitReferenceFilesUnderWorkspaceScope(t *testing.T) {
	home := t.TempDir()
	workspace := t.TempDir()
	pinReferenceParityHome(t, home)

	selection := referenceParitySelection()
	adapters := resolveAdapters(selection.Agents)
	engramHomePath, sddHomePath := referenceParityPaths(home)
	engramWsPath, sddWsPath := referenceParityPaths(workspace)

	cases := map[string]struct {
		component model.ComponentID
		unwanted  []string
	}{
		"engram": {model.ComponentEngram, []string{engramHomePath, engramWsPath}},
		"sdd":    {model.ComponentSDD, []string{sddHomePath, sddWsPath}},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := componentPathsWithWorkspaceScoped(home, workspace, ScopeWorkspace, selection, adapters, tc.component)
			for _, unwanted := range tc.unwanted {
				if containsPath(got, unwanted) {
					t.Fatalf("workspace-scoped %s paths must not name %q: no reference file is written there; got:\n%v", name, unwanted, got)
				}
			}
		})
	}
}

// TestSyncBackupTargetsIncludeSharedReferenceFiles proves the parity claim
// itself: sync delegates into the same function install uses, so wiring the
// paths in one place has to reach the sync backup set too.
func TestSyncBackupTargetsIncludeSharedReferenceFiles(t *testing.T) {
	home := t.TempDir()
	pinReferenceParityHome(t, home)

	selection := referenceParitySelection()
	adapters := resolveAdapters(selection.Agents)
	engramPath, sddPath := referenceParityPaths(home)

	got, err := syncBackupTargets(home, "", selection, adapters)
	if err != nil {
		t.Fatalf("syncBackupTargets() error = %v", err)
	}

	for _, want := range []string{engramPath, sddPath} {
		if !containsPath(got, want) {
			t.Fatalf("syncBackupTargets must include the reference file %q; got:\n%v", want, got)
		}
	}
}
