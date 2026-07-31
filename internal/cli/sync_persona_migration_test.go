package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// This file IS the persona migration compatibility matrix requested on issue
// #912. If the reply posted on that issue and this table ever disagree, the
// table is authoritative — it is executable and the comment is not.
//
// Two properties are covered:
//
//  1. Every row of the R1 migration matrix is idempotent at N=2. Not just the
//     resolved tuple, but the bytes actually written to disk: a migration that
//     settles in memory while rewriting files on every run is still a bug for
//     anyone diffing their config.
//
//  2. This change and PR #1712 converge regardless of merge order, because both
//     send `gentleman-neutral-artifacts` to plain `neutral`.

// syncPersonaTestArgs is the sync invocation shared by every case here. Kept in
// one place so a row cannot silently diverge from the others.
var syncPersonaTestArgs = []string{"--agents", "claude-code", "--sdd-mode", "single"}

// personaSyncOutcome is everything observable about one sync run that the
// migration contract cares about.
type personaSyncOutcome struct {
	persona    model.PersonaID
	region     string
	artifacts  bool
	stateBytes string // normalized: see normalizeHomePaths
	surface    string // normalized persona surface: see readPersonaSurface
	persistedP string // the `persona` value left in state.json afterwards
}

// normalizeHomePaths replaces the temp home with a stable token. Each case runs
// under its own t.TempDir(), whose path embeds the test name, so two cases that
// produce identical content would still differ byte-for-byte on any line that
// echoes a path. Comparing raw bytes across cases without this would fail for a
// reason that has nothing to do with persona migration.
func normalizeHomePaths(s, home string) string {
	return strings.ReplaceAll(s, home, "$HOME")
}

// writeRawLegacyState writes state.json verbatim instead of going through
// state.Write. The matrix needs shapes state.InstallState cannot express — most
// importantly a file with no `persona` key at all, which is what a pre-persona
// install actually has on disk and which marshalling would always emit.
func writeRawLegacyState(t *testing.T, home, rawJSON string) {
	t.Helper()
	stateDir := filepath.Join(home, ".gentle-ai")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", stateDir, err)
	}
	if err := os.WriteFile(filepath.Join(stateDir, "state.json"), []byte(rawJSON), 0o644); err != nil {
		t.Fatalf("WriteFile(state.json): %v", err)
	}
}

// runPersonaSync performs one sync against home and captures everything the
// migration contract asserts on.
func runPersonaSync(t *testing.T, home string, run int) personaSyncOutcome {
	t.Helper()

	result, err := RunSync(syncPersonaTestArgs)
	if err != nil {
		t.Fatalf("RunSync() run %d: %v", run, err)
	}

	stateBytes, err := os.ReadFile(filepath.Join(home, ".gentle-ai", "state.json"))
	if err != nil {
		t.Fatalf("ReadFile(state.json) run %d: %v", run, err)
	}

	out := personaSyncOutcome{
		persona:    result.Selection.Persona,
		region:     result.Selection.Region,
		artifacts:  result.Selection.ArtifactsInEnglish,
		stateBytes: normalizeHomePaths(string(stateBytes), home),
	}

	persisted, err := readPersistedPersonaValue(string(stateBytes))
	if err != nil {
		t.Fatalf("read persisted persona run %d: %v", run, err)
	}
	out.persistedP = persisted

	out.surface = readPersonaSurface(t, home)

	return out
}

// readPersonaSurface renders every file the persona component writes for Claude
// Code into one deterministic string: the CLAUDE.md persona section plus the
// whole output-styles directory, sorted by path, with a presence marker for
// files that are absent.
//
// Reading only CLAUDE.md is not enough, and the first version of this file made
// exactly that mistake. Claude Code delivers its reply voice through the output
// style — voiceLivesInOutputStyle in the persona component — so the composed
// regional directive never appears in CLAUDE.md at all. A test comparing only
// CLAUDE.md would have reported the gentleman and neutral paths as carrying no
// regional voice, and every "no Rioplatense here" assertion would have passed
// while verifying nothing.
func readPersonaSurface(t *testing.T, home string) string {
	t.Helper()

	paths := []string{filepath.Join(home, ".claude", "CLAUDE.md")}

	styleDir := filepath.Join(home, ".claude", "output-styles")
	entries, err := os.ReadDir(styleDir)
	switch {
	case err == nil:
		for _, e := range entries {
			if !e.IsDir() {
				paths = append(paths, filepath.Join(styleDir, e.Name()))
			}
		}
	case os.IsNotExist(err):
		// No output styles written (custom persona). Recorded by absence below.
	default:
		t.Fatalf("ReadDir(%s): %v", styleDir, err)
	}
	sort.Strings(paths)

	var b strings.Builder
	for _, p := range paths {
		rel, err := filepath.Rel(home, p)
		if err != nil {
			t.Fatalf("Rel(%s): %v", p, err)
		}
		data, err := os.ReadFile(p)
		switch {
		case err == nil:
			fmt.Fprintf(&b, "=== %s ===\n", rel)
			b.WriteString(normalizeHomePaths(string(data), home))
			b.WriteString("\n")
		case os.IsNotExist(err):
			// Absence is part of the contract: custom must inject nothing, and a
			// run that silently stopped writing a file must not look idempotent.
			fmt.Fprintf(&b, "=== %s (absent) ===\n", rel)
		default:
			t.Fatalf("ReadFile(%s): %v", p, err)
		}
	}
	return b.String()
}

// readPersistedPersonaValue pulls the raw `persona` value back out of state.json.
// Asserting on this is what proves a second sync no longer needs an alias: if the
// stored value is still a legacy alias, every future run re-enters the migration
// path, and "idempotent output" would be hiding a state that never converged.
func readPersistedPersonaValue(stateJSON string) (string, error) {
	var parsed struct {
		Persona string `json:"persona"`
	}
	if err := json.Unmarshal([]byte(stateJSON), &parsed); err != nil {
		return "", err
	}
	return parsed.Persona, nil
}

// TestPersonaMigrationMatrixIsIdempotentAcrossEveryRow runs each R1 row through
// two consecutive syncs and asserts nothing moves on the second one.
//
// The existing WU-4 idempotency test covers a single path. This covers all five
// rows, and compares written bytes rather than only the in-memory tuple.
func TestPersonaMigrationMatrixIsIdempotentAcrossEveryRow(t *testing.T) {
	rows := []struct {
		name          string
		stateJSON     string
		wantPersona   model.PersonaID
		wantRegion    string
		wantArtifacts bool
		// wantPersistedPersona is the value state.json must hold afterwards, i.e.
		// the migrated identity rather than the legacy alias it arrived as.
		wantPersistedPersona string
	}{
		{
			// A readable file predating the persona field. This install was already
			// getting the gentleman default, so migration reproduces it.
			name:                 "absent persona field",
			stateJSON:            `{"installed_agents":["claude-code"]}`,
			wantPersona:          model.PersonaGentle,
			wantRegion:           string(model.RegionArgentina),
			wantArtifacts:        true,
			wantPersistedPersona: string(model.PersonaGentle),
		},
		{
			name:                 "gentleman",
			stateJSON:            `{"installed_agents":["claude-code"],"persona":"gentleman"}`,
			wantPersona:          model.PersonaGentle,
			wantRegion:           string(model.RegionArgentina),
			wantArtifacts:        true,
			wantPersistedPersona: string(model.PersonaGentle),
		},
		{
			// The hybrid alias. Resolves to neutral, not gentle — see #1702 defect 1.
			name:                 "gentleman-neutral-artifacts",
			stateJSON:            `{"installed_agents":["claude-code"],"persona":"gentleman-neutral-artifacts"}`,
			wantPersona:          model.PersonaNeutral,
			wantRegion:           "",
			wantArtifacts:        true,
			wantPersistedPersona: string(model.PersonaNeutral),
		},
		{
			name:                 "neutral",
			stateJSON:            `{"installed_agents":["claude-code"],"persona":"neutral"}`,
			wantPersona:          model.PersonaNeutral,
			wantRegion:           "",
			wantArtifacts:        true,
			wantPersistedPersona: string(model.PersonaNeutral),
		},
		{
			// Custom injects no persona content, so ArtifactsInEnglish stays as the
			// caller left it rather than being forced true.
			name:                 "custom",
			stateJSON:            `{"installed_agents":["claude-code"],"persona":"custom"}`,
			wantPersona:          model.PersonaCustom,
			wantRegion:           "",
			wantArtifacts:        false,
			wantPersistedPersona: string(model.PersonaCustom),
		},
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			home := t.TempDir()
			setSyncTestHome(t, home)
			writeRawLegacyState(t, home, row.stateJSON)

			first := runPersonaSync(t, home, 1)

			if first.persona != row.wantPersona {
				t.Errorf("run 1 Persona = %q, want %q", first.persona, row.wantPersona)
			}
			if first.region != row.wantRegion {
				t.Errorf("run 1 Region = %q, want %q", first.region, row.wantRegion)
			}
			if first.artifacts != row.wantArtifacts {
				t.Errorf("run 1 ArtifactsInEnglish = %v, want %v", first.artifacts, row.wantArtifacts)
			}
			if first.persistedP != row.wantPersistedPersona {
				t.Errorf("run 1 persisted persona = %q, want %q — a later sync would re-enter the alias path",
					first.persistedP, row.wantPersistedPersona)
			}

			second := runPersonaSync(t, home, 2)

			if second.persona != first.persona || second.region != first.region || second.artifacts != first.artifacts {
				t.Errorf("resolved tuple moved between runs: run1 {%q, %q, %v}, run2 {%q, %q, %v}",
					first.persona, first.region, first.artifacts,
					second.persona, second.region, second.artifacts)
			}
			if second.stateBytes != first.stateBytes {
				t.Errorf("state.json changed on the second sync:\n--- run1 ---\n%s\n--- run2 ---\n%s",
					first.stateBytes, second.stateBytes)
			}
			if second.surface != first.surface {
				t.Errorf("the injected persona surface changed on the second sync (len %d then %d)",
					len(first.surface), len(second.surface))
			}
		})
	}
}

// TestPersonaMigrationConvergesAcrossMergeOrderWith1712 is the sequencing answer
// owed to issue #912.
//
// PR #1712 remaps the stored value `gentleman-neutral-artifacts` to `neutral`.
// This change migrates that same alias to regionless `neutral`. Whichever lands
// first, a user must end up in the same place — otherwise the merge order of two
// unrelated PRs would decide someone's persona.
//
// Sequence A models #1712 landing first: the stored value has already been
// rewritten to `neutral` by the time this change's migration sees it.
// Sequence B models this change landing first: the migration meets the original
// legacy value.
func TestPersonaMigrationConvergesAcrossMergeOrderWith1712(t *testing.T) {
	run := func(t *testing.T, label, stateJSON string) personaSyncOutcome {
		t.Helper()
		home := t.TempDir()
		setSyncTestHome(t, home)
		writeRawLegacyState(t, home, stateJSON)
		out := runPersonaSync(t, home, 1)
		if out.persona != model.PersonaNeutral || out.region != "" || !out.artifacts {
			t.Fatalf("%s resolved {%q, %q, %v}, want {neutral, \"\", true}",
				label, out.persona, out.region, out.artifacts)
		}
		if out.persistedP != string(model.PersonaNeutral) {
			t.Errorf("%s persisted persona = %q, want %q — the alias must not survive the first sync",
				label, out.persistedP, model.PersonaNeutral)
		}
		return out
	}

	seqA := run(t, "sequence A (#1712 remap applied first)",
		`{"installed_agents":["claude-code"],"persona":"neutral"}`)
	seqB := run(t, "sequence B (this change migrates the original alias)",
		`{"installed_agents":["claude-code"],"persona":"gentleman-neutral-artifacts"}`)

	if seqA.stateBytes != seqB.stateBytes {
		t.Errorf("merge order changed the written state.json:\n--- sequence A ---\n%s\n--- sequence B ---\n%s",
			seqA.stateBytes, seqB.stateBytes)
	}
	if seqA.surface != seqB.surface {
		t.Errorf("merge order changed the injected persona surface (len %d vs %d)",
			len(seqA.surface), len(seqB.surface))
	}
}

// TestPersonaMigrationHybridDoesNotConvergeWithGentlemanEndToEnd is the negative
// half of the matrix, at the sync level rather than the resolver level.
//
// The convergence test above would still pass if every alias resolved to the
// gentleman tuple, so on its own it cannot detect the defect this change exists
// to remove. This one pins the separation: the hybrid must produce different
// injected content than `gentleman`, with no Rioplatense directive in it.
func TestPersonaMigrationHybridDoesNotConvergeWithGentlemanEndToEnd(t *testing.T) {
	run := func(t *testing.T, stateJSON string) personaSyncOutcome {
		t.Helper()
		home := t.TempDir()
		setSyncTestHome(t, home)
		writeRawLegacyState(t, home, stateJSON)
		return runPersonaSync(t, home, 1)
	}

	hybrid := run(t, `{"installed_agents":["claude-code"],"persona":"gentleman-neutral-artifacts"}`)
	gentleman := run(t, `{"installed_agents":["claude-code"],"persona":"gentleman"}`)

	if hybrid.persona == gentleman.persona && hybrid.region == gentleman.region {
		t.Fatalf("the hybrid alias still resolves to the gentleman tuple {%q, %q} — #1702 defect 1 regression",
			hybrid.persona, hybrid.region)
	}
	if hybrid.surface == gentleman.surface {
		t.Error("hybrid and gentleman produced an identical persona surface; the alias is still delivering the regional voice it never promised")
	}
	// What these markers actually guard, verified by mutation: they catch regional
	// voice being baked back into the neutral ASSET. They do NOT discriminate the
	// Decision 5 routing, because the neutral path never calls
	// ComposeLanguageDirective on any agent — output-style-neutral.md is written
	// verbatim. Injecting a regional clause into the composer's empty-region branch
	// leaves these assertions green, so do not read them as covering that.
	for _, marker := range []string{"Rioplatense", "voseo"} {
		if strings.Contains(hybrid.surface, marker) {
			t.Errorf("hybrid persona surface contains %q — neutral must carry no regional directive", marker)
		}
	}
	// Vacuity guard. Without it the two assertions above would also pass if the
	// surface were read from the wrong place and contained no voice at all — which
	// is precisely how the first draft of this test failed.
	if !strings.Contains(gentleman.surface, "Rioplatense") {
		t.Error("gentleman persona surface carries no Rioplatense directive; the marker assertions above would pass vacuously")
	}
}
