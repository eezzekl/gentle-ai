package engram

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
)

// sharedReferenceAdapters are the adapters that implement SharedReferenceLayout
// and therefore route the full protocol body to ~/.gemini/references/ instead of
// inlining it into GEMINI.md (design.md D1).
func sharedReferenceAdapters() map[string]agents.Adapter {
	return map[string]agents.Adapter{
		"gemini":      geminiAdapter(),
		"antigravity": antigravityAdapter(),
	}
}

// newSharedReferenceHome returns a temp dir pinned as the user home, so the
// reference path the injector computes matches the home-anchored path the
// session-bootstrap stub advertises.
func newSharedReferenceHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	SetUserHomeDirForTest(t, home)
	SetLookPathForTest(t, "/opt/homebrew/bin/engram", "")
	return home
}

func engramReferencePath(t *testing.T, adapter agents.Adapter, dir string) string {
	t.Helper()
	layout, ok := adapter.(SharedReferenceLayout)
	if !ok {
		t.Fatalf("adapter %T does not implement SharedReferenceLayout", adapter)
	}
	return filepath.Join(layout.ReferencesDir(dir), engramReferenceFileName)
}

// TestInjectWritesEngramReferenceFileVerbatim asserts the Reference-File
// Emission requirement (engram-protocol-injection spec): for adapters that
// implement SharedReferenceLayout, the full protocol content is written to
// ~/.gemini/references/engram-protocol.md verbatim, with no section markers.
func TestInjectWritesEngramReferenceFileVerbatim(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			result, err := Inject(home, adapter)
			if err != nil {
				t.Fatalf("Inject(%s) error = %v", name, err)
			}

			refPath := engramReferencePath(t, adapter, home)
			got := readFileForTest(t, refPath)

			if got != protocolFull() {
				t.Fatalf("reference file content != protocolFull()\n--- got (%d bytes) ---\n%s", len(got), got)
			}
			if strings.Contains(got, "gentle-ai:engram-protocol") {
				t.Fatalf("reference file must not carry section markers; got:\n%s", got)
			}
			if !containsPath(result.Files, refPath) {
				t.Fatalf("Inject(%s) result.Files missing %q; got %v", name, refPath, result.Files)
			}
		})
	}
}

// TestInjectReplacesRootProtocolWithBootstrapStub asserts that the root system
// prompt carries only the pointer stub under the existing engram-protocol
// marker — the full body must no longer be inlined (Root Assembly Byte Budget).
func TestInjectReplacesRootProtocolWithBootstrapStub(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			if _, err := Inject(home, adapter); err != nil {
				t.Fatalf("Inject(%s) error = %v", name, err)
			}

			root := readFileForTest(t, adapter.SystemPromptFile(home))

			if !strings.Contains(root, "<!-- gentle-ai:engram-protocol -->") {
				t.Fatalf("root prompt missing engram-protocol open marker; got:\n%s", root)
			}
			if !strings.Contains(root, renderSessionBootstrapStub()) {
				t.Fatalf("root prompt missing session-bootstrap stub; got:\n%s", root)
			}
			if strings.Contains(root, "### PROACTIVE SAVE TRIGGERS") {
				t.Fatalf("root prompt still inlines the full protocol body; got:\n%s", root)
			}
		})
	}
}

// TestInjectKeepsProtocolInlineWhenTargetIsNotHome is the regression guard for
// the workspace-scoped install. `agy install --scope=workspace` passes the
// workspace, not the home dir, as the injection target. The stub names an
// absolute home-anchored path and tells the agent NOT to resolve it against the
// project, so splitting there would write the body somewhere the pointer never
// sends the agent and the whole protocol would be silently lost. The inline
// body — which always reaches the agent — must survive instead.
func TestInjectKeepsProtocolInlineWhenTargetIsNotHome(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			workspace := t.TempDir()
			SetUserHomeDirForTest(t, home)
			SetLookPathForTest(t, "/opt/homebrew/bin/engram", "")

			if _, err := Inject(workspace, adapter); err != nil {
				t.Fatalf("Inject(%s, workspace) error = %v", name, err)
			}

			root := readFileForTest(t, adapter.SystemPromptFile(workspace))
			if !strings.Contains(root, "### PROACTIVE SAVE TRIGGERS") {
				t.Fatalf("workspace-scoped root must keep the full inline protocol; got:\n%s", root)
			}
			if strings.Contains(root, renderSessionBootstrapStub()) {
				t.Fatalf("workspace-scoped root must not carry the home-anchored stub; got:\n%s", root)
			}

			// No reference file may be written under either directory: the one
			// under the workspace would be unreachable, and the one under home
			// would be a file this install was never asked to touch.
			for label, dir := range map[string]string{"workspace": workspace, "home": home} {
				if _, err := os.Stat(engramReferencePath(t, adapter, dir)); !os.IsNotExist(err) {
					t.Fatalf("unexpected reference file under %s (stat err = %v)", label, err)
				}
			}
		})
	}
}

// TestInjectReferenceLayoutIsIdempotent asserts design.md D3: a second sync over
// an already-converted layout produces byte-identical root and reference files.
func TestInjectReferenceLayoutIsIdempotent(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name, func(t *testing.T) {
			home := newSharedReferenceHome(t)

			refPath := engramReferencePath(t, adapter, home)
			rootPath := adapter.SystemPromptFile(home)

			if _, err := Inject(home, adapter); err != nil {
				t.Fatalf("Inject(%s) first pass error = %v", name, err)
			}
			firstRef := readFileForTest(t, refPath)
			firstRoot := readFileForTest(t, rootPath)

			if _, err := Inject(home, adapter); err != nil {
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

func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	return string(data)
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

// TestManagedReferencePathMatchesWhatInjectWrites is the single-source-of-truth
// guard for design.md D9. The install/sync path list must name the reference
// file exactly when Inject writes one, and never otherwise. A path that is
// listed but never written has no backup snapshot to restore and turns
// post-apply and post-sync verification into a false failure on every
// workspace-scoped install.
func TestManagedReferencePathMatchesWhatInjectWrites(t *testing.T) {
	for name, adapter := range sharedReferenceAdapters() {
		t.Run(name+"/home_anchored", func(t *testing.T) {
			home := newSharedReferenceHome(t)

			got := ManagedReferencePath(adapter, home)
			if want := engramReferencePath(t, adapter, home); got != want {
				t.Fatalf("ManagedReferencePath(%s, home) = %q, want %q", name, got, want)
			}

			if _, err := Inject(home, adapter); err != nil {
				t.Fatalf("Inject(%s, home) error = %v", name, err)
			}
			if _, err := os.Stat(got); err != nil {
				t.Fatalf("ManagedReferencePath named %q but Inject wrote no such file: %v", got, err)
			}
		})

		t.Run(name+"/workspace_scoped", func(t *testing.T) {
			home := t.TempDir()
			workspace := t.TempDir()
			SetUserHomeDirForTest(t, home)
			SetLookPathForTest(t, "/opt/homebrew/bin/engram", "")

			if got := ManagedReferencePath(adapter, workspace); got != "" {
				t.Fatalf("ManagedReferencePath(%s, workspace) = %q, want empty: Inject keeps the body inline there", name, got)
			}
		})
	}

	t.Run("adapter_without_shared_reference_layout", func(t *testing.T) {
		if got := ManagedReferencePath(struct{}{}, t.TempDir()); got != "" {
			t.Fatalf("ManagedReferencePath(non-layout adapter) = %q, want empty", got)
		}
	})
}
