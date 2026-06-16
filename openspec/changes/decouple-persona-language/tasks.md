# Tasks: decouple-persona-language — Motor-first slice

Slice scope: model / state / validate / sync+migration / inject / Claude+generic assets / spec update / tests.
Deferred adapters (OpenCode/Kilocode, Kimi, Kiro, Hermes) and the TUI region screen are listed separately as Slice 2+ deferred units.

Strict TDD is active (`go test ./...`). Every implementation task is preceded by its test task (red → green order).

---

## Work Unit 1 — Model layer: types + region map + directive composer

**Satisfies**: R3 (two independent axes), R4 (composed directive), R5 (artifacts checkbox)

### WU-1.1 — Test: `PersonaGentle`, `RegionID` map, `composeLanguageDirective`
- [x] In `internal/model/types_test.go` (create if absent): write table-driven test asserting `PersonaGentle == "gentle"` and that `PersonaGentlemanNeutralArtifacts` is no longer in the selectable constants (compile-level: removed from type block).
- [x] In `internal/model/region_test.go` (new file): write table-driven test for `composeLanguageDirective(region, artifactsInEnglish)` covering:
  - Each curated region (argentina, mexico, colombia, spain, chile) × {true, false} → assert directive string contains the expected language phrase and the correct artifacts clause.
  - `user-language` sentinel → assert "reply in the language the user writes in" variant, no forced region.
  - Free-text input (e.g. "yucateco") → assert free-text injected verbatim, plus artifacts clause.
  - `artifactsInEnglish=true` → assert English-artifacts clause present.
  - `artifactsInEnglish=false` → assert in-language clause present and different from the `=true` case.
- [x] Run `go test ./internal/model/...` — expect compilation failure or test failure (RED). ✓ Confirmed RED.

### WU-1.2 — Implement: types + region + directive
- [x] In `internal/model/types.go`:
  - Add `PersonaGentle PersonaID = "gentle"`.
  - Remove `PersonaGentlemanNeutralArtifacts` constant (keep removal comment: "removed; use PersonaGentle + migration alias in validate.go").
  - Add `RegionID string` type alias.
  - Add curated `RegionMap map[RegionID]string` with the 7 entries (argentina through user-language sentinel). TUI labels live here in English-key, Spanish-label form: key is the `RegionID`, value is the display label.
  - Add sentinel constants: `RegionArgentina RegionID = "argentina"`, `RegionMexico`, `RegionColombia`, `RegionSpain`, `RegionChile`, `RegionUserLanguage RegionID = "user-language"`.
- [x] In `internal/model/region.go` (new file): implement pure function `composeLanguageDirective(region RegionID, artifactsInEnglish bool) string`.
  - Region map lookup → compose the reply-language clause.
  - `user-language` sentinel → follow-user clause (no forced region).
  - Unrecognized region string → treat as free text, inject verbatim.
  - Append the artifacts clause based on `artifactsInEnglish`.
- [x] Run `go test ./internal/model/...` — expect GREEN. ✓ Confirmed GREEN.

**Acceptance**: `composeLanguageDirective("argentina", true)` produces a directive containing Rioplatense/voseo and English-artifacts clause. `composeLanguageDirective("argentina", true)` must be byte-identical to the pre-change gentleman language directive (proving gentle+argentina reproduces today exactly).

---

## Work Unit 2 — Selection + State: new fields, persistence, MergeAgents

**Satisfies**: R6 (persistence), R3 (axes in selection)

### WU-2.1 — Test: `Selection` new fields + `InstallState` round-trip
- [x] In `internal/model/selection_test.go` (create if absent): assert `Selection` has fields `Region string` and `ArtifactsInEnglish bool` (compile-driven).
- [x] In `internal/state/state_test.go`:
  - Table-driven JSON marshal/unmarshal round-trip for `InstallState` with new fields:
    - `region: "chile"`, `artifactsInEnglish: false` → assert both survive after unmarshal (false must NOT disappear — no omitempty on the bool).
    - `region: ""` → assert omitempty omits the key (empty region is "follow user", ok to omit).
    - `artifactsInEnglish: true` → assert key present and value `true`.
  - `MergeAgents` test: assert merged state carries `Region` and `ArtifactsInEnglish` from existing state unchanged.
- [x] Run `go test ./internal/state/... ./internal/model/...` — expect RED. ✓ Confirmed RED.

### WU-2.2 — Implement: `Selection` + `InstallState` + `MergeAgents`
- [x] In `internal/model/selection.go`: add `Region string` and `ArtifactsInEnglish bool` fields to `Selection`.
- [x] In `internal/state/state.go`:
  - Add `Region string \`json:"region,omitempty"\`` and `ArtifactsInEnglish bool \`json:"artifactsInEnglish"\`` (no omitempty on the bool) to `InstallState`.
  - Update `MergeAgents` to carry both new fields from `existing`.
- [x] Run `go test ./internal/state/... ./internal/model/...` — expect GREEN. ✓ Confirmed GREEN.

**Acceptance**: `InstallState{ArtifactsInEnglish: false}` marshals to JSON containing `"artifactsInEnglish":false` (the bool is explicit, never omitted). `MergeAgents` test passes.

---

## Work Unit 3 — Validation: `normalizePersona` accepts new IDs + legacy aliases

**Satisfies**: R7 (validation)

### WU-3.1 — Test: `normalizePersona` table
- [x] In `internal/cli/persona_language_contract_test.go` (replaced existing single-case test): add table-driven cases:
  - `"gentle"` → `PersonaGentle`, no error.
  - `"neutral"` → `PersonaNeutral`, no error.
  - `"custom"` → `PersonaCustom`, no error.
  - `"gentleman"` → `PersonaGentle`, no error (back-compat alias).
  - `"gentleman-neutral-artifacts"` → `PersonaGentle`, no error (back-compat alias, migration convergence).
  - `""` → `PersonaGentle`, no error (default).
  - `"unknown-value"` → error, and the resolved value is NOT a persona that injects regional voice.
- [x] Run `go test ./internal/cli/...` — expect RED. ✓ Confirmed RED.

### WU-3.2 — Implement: update `normalizePersona`
- [x] In `internal/cli/validate.go`:
  - Update `normalizePersona` switch: add `PersonaGentle`; map `PersonaGentleman` → `PersonaGentle`; map `PersonaGentlemanNeutralArtifacts` → `PersonaGentle`.
  - Update empty-string default from `PersonaGentleman` to `PersonaGentle`.
- [x] Run `go test ./internal/cli/...` — expect GREEN. ✓ Confirmed GREEN (full suite passes).

**Acceptance**: Both `"gentleman"` and `"gentleman-neutral-artifacts"` normalize to `PersonaGentle` (confirming migration convergence at the validation layer). The selectable set is `{gentle, neutral, custom}` only.

---

## Work Unit 4 — Sync + migration: `applyResolvedPersona` + `BuildSyncSelection`

**Satisfies**: R1 (back-compat migration, MANDATORY), R2 (sync idempotency), R6 (persistence of new fields)

### WU-4.1 — Test: migration matrix + idempotency
- [x] In `internal/cli/sync_test.go` (extend): write table-driven test for `applyResolvedPersona` covering all 5 migration matrix rows:

  | old persona string | expected style | expected region | expected artifactsInEnglish |
  |---|---|---|---|
  | `"gentleman"` | `gentle` | `argentina` | `true` |
  | `"gentleman-neutral-artifacts"` | `gentle` | `argentina` | `true` |
  | `"neutral"` | `neutral` | `user-language` | `true` |
  | `"custom"` | `custom` | `""` | (N/A — no injection) |
  | `""` | `gentle` | `argentina` | `true` |

  - Explicit assertion: `"gentleman"` and `"gentleman-neutral-artifacts"` produce byte-identical resolved tuples (proving hybrid was a no-op).
  - Explicit assertion: `artifactsInEnglish` is `true` (not Go zero-value `false`) for every non-custom row.
  - Note: `neutral` migrates to `user-language` — acceptance-note confirmed here.

- [x] Write idempotency test for `RunSyncWithSelection`: inject twice, assert resulting content byte-identical on second run.
- [x] Run `go test ./internal/cli/...` — expect RED.

### WU-4.2 — Implement: migrate + sync selection
- [x] In `internal/cli/sync.go`:
  - Update `applyResolvedPersona` signature to populate `Selection.Region` and `Selection.ArtifactsInEnglish` from the persisted `InstallState` (pass `InstallState` instead of bare string, or use a helper that extracts both fields).
  - Implement the full migration matrix: for each legacy `persona` string, set `selection.Persona`, `selection.Region`, and `selection.ArtifactsInEnglish` explicitly. **Never rely on bool zero-value**; set `true` explicitly in each non-custom branch.
  - Update `BuildSyncSelection` to read `state.Region` and `state.ArtifactsInEnglish` from the persisted `InstallState` and write them into the `Selection`.
- [x] Run `go test ./internal/cli/...` — expect GREEN.

**Acceptance notes**:
- `neutral` style → region `user-language` ("Idioma del usuario") — per migration constraint #1118.
- `ArtifactsInEnglish` explicitly `true` in every non-custom migration branch.
- Idempotency test green: second sync produces no file changes.

---

## Work Unit 5 — Asset strip: `stripLanguageSection` + asset edits

**Satisfies**: R4 (region-neutral base), R9 (Claude/generic assets)

### WU-5.1 — Test: `stripLanguageSection` transform + asset guard
- [x] In `internal/components/persona/inject_test.go` (extend) or new `strip_test.go`:
  - Pure-function test for `stripLanguageSection(content string) string`:
    - Input: the full text of `claude/persona-gentleman.md` (embed or inline snippet).
    - Assert: the `## Language` section (lines 37–43) is removed.
    - Assert: all other H2 sections survive byte-for-byte (`## Rules`, `## Personality`, `## Persona Scope`, `## Tone`, `## Philosophy`, `## Expertise`, `## Behavior`, `## Contextual Skill Loading`, `## Engram Persistent Memory`).
    - Assert: the stripped output contains NO hardcoded Rioplatense/voseo language line.
  - Same test for `claude/output-style-gentleman.md`: `## Language Rules` section removed; `## Persona Scope`, `## Tone`, `## Philosophy`, `## Behavior` survive.
  - Same test for `generic/persona-gentleman.md`: `## Language` section removed; all other sections survive.
  - Asset guard: after stripping, the gentle base asset must NOT contain "Rioplatense" or "voseo" as a baked-in instruction.
- [x] Run `go test ./internal/components/persona/...` — expect RED.

### WU-5.2 — Implement: `stripLanguageSection` function
- [x] In `internal/components/persona/strip.go` (new file): implement `stripLanguageSection(content string) string`.
  - Logic: find the `## Language` or `## Language Rules` H2 header (case-sensitive match on both forms); remove from that line through the next `##`-prefixed line (exclusive) or EOF.
  - This is a pure string transform — no file I/O.
- [x] Run `go test ./internal/components/persona/...` — confirm strip tests GREEN.

### WU-5.3 — Strip asset files
- [x] Edit `internal/assets/claude/persona-gentleman.md`: remove the `## Language` section (lines ~37–43).
- [x] Edit `internal/assets/claude/output-style-gentleman.md`: remove the `## Language Rules` section (lines ~40–48). Keep `## Persona Scope`.
- [x] Edit `internal/assets/generic/persona-gentleman.md`: remove its `## Language` section (locate equivalent block).
- [x] Run `go test ./internal/components/persona/...` — asset guard tests GREEN.

**Acceptance**: After stripping, no baked-in Rioplatense/voseo instruction remains in any of the three Claude/generic asset files. The strip is header-bounded: only the `## Language` / `## Language Rules` section is removed.

---

## Work Unit 6 — Inject: wire `composeLanguageDirective` into `personaContent` + rename guard

**Satisfies**: R4 (single composed directive source), R9 (composed directive sole regional source)

### WU-6.1 — Test: `personaContent` with directive + golden idempotency
- [x] In `inject_test.go`:
  - Test `personaContent` (updated signature or via the composition path) with `(AgentClaudeCode, PersonaGentle, RegionArgentina, true)`:
    - Assert result contains exactly ONE language directive line (the composed one).
    - Assert result does NOT contain a second hardcoded Rioplatense line.
    - Assert the composed directive for `(gentle, argentina, true)` is byte-identical to the pre-change gentleman persona's language section output (golden comparison).
  - Test `personaContent` with `(AgentClaudeCode, PersonaNeutral, RegionUserLanguage, true)`: assert neutral contract sections survive, no marked regional voice baked in.
  - Idempotency: call the full inject path twice on a temp file; assert content byte-identical after second call; assert no duplicate directive lines.
  - `isGentlePersona` unit test: `PersonaGentle` → true; `PersonaGentleman` → true (alias); `PersonaGentlemanNeutralArtifacts` → true (alias); `PersonaNeutral` → false; `PersonaCustom` → false.
- [x] Run `go test ./internal/components/persona/...` — expect RED. ✓ Confirmed RED.

### WU-6.2 — Implement: wire directive into inject
- [x] In `internal/components/persona/inject.go`:
  - Updated `personaContent` signature to `personaContent(agent, persona, region, artifactsInEnglish)`: call the existing asset read, run `stripLanguageSection()`, then append `"\n" + model.ComposeLanguageDirective(region, artifactsInEnglish)`.
  - Renamed `isGentlemanConversationPersona` to `isGentlePersona`. Updated function to match all 3 IDs (`PersonaGentle`, `PersonaGentleman`, `PersonaGentlemanNeutralArtifacts`). Updated all 4 call sites.
  - Updated `Inject` and `InjectForSync` signatures to accept `region model.RegionID` and `artifactsInEnglish bool`. Updated callers in `sync.go` (InjectForSync) and `run.go` (Inject).
  - Exported `ComposeLanguageDirective` in `internal/model/region.go`.
  - Updated all test callers in `inject_test.go`, `golden_test.go`, `openclaw_integration_test.go`, `persona_language_contract_test.go`.
  - Regenerated 5 affected golden files with `-update` flag.
- [x] Run `go test ./internal/components/persona/...` — expect GREEN. ✓ Confirmed GREEN (full `go test ./...` passes).

**Acceptance**: Golden round-trip test — legacy state `"persona": "gentleman"` → migrate → inject → output byte-identical to pre-change gentleman persona baseline (committed golden file or inline fixture). `isGentlePersona` correctly matches all three gentleman-family IDs.

---

## Work Unit 7 — TUI persona screen: drop hybrid, update style options

**Satisfies**: R8 (TUI style options), R3 (selectable set)

### WU-7.1 — Test: `PersonaOptions()` and `RenderPersona`
- [x] In `internal/tui/screens/` — new test or extend `persona_preset_test.go`:
  - Assert `PersonaOptions()` returns exactly `[gentle, neutral, custom]` (length 3).
  - Assert `PersonaOptions()` does NOT contain `PersonaGentlemanNeutralArtifacts`.
  - Render test: `RenderPersona(PersonaGentle, 0)` produces output containing "gentle" and NOT containing "gentleman-neutral-artifacts".
- [x] Run `go test ./internal/tui/...` — expect RED.

### WU-7.2 — Implement: update `PersonaOptions` + `RenderPersona`
- [x] In `internal/tui/screens/persona.go`:
  - Update `PersonaOptions()` to return `[]model.PersonaID{model.PersonaGentle, model.PersonaNeutral, model.PersonaCustom}`.
  - Update `personaDescriptions` map: add `PersonaGentle` description; remove `PersonaGentleman` and `PersonaGentlemanNeutralArtifacts` entries.
- [x] In `internal/tui/model.go`: update `ScreenPersona` cursor-count (was 4 options + Back; now 3 options + Back → `len(PersonaOptions()) + 1` already correct since it uses the slice length).
  - Update any `ScreenPersona` routing logic that branched on `PersonaGentlemanNeutralArtifacts`.
- [x] In `internal/tui/router.go`: add new route `ScreenPersona → ScreenPersonaLanguage` for `gentle` and `neutral`; keep `ScreenPersona → ScreenPreset` for `custom` (skips language screen). **Shipped in WU-8** (route requires the screen, which WU-8 introduces).
- [x] Run `go test ./internal/tui/...` — expect GREEN.

**Acceptance**: `PersonaOptions()` returns exactly 3 items. Routing: selecting `custom` skips `ScreenPersonaLanguage` and goes directly to `ScreenPreset`.

---

## Work Unit 8 — TUI language/region screen: new `ScreenPersonaLanguage`

**Satisfies**: R8 (region radios, "Idioma del usuario", free text, checkbox), R5 (checkbox default ON)

### WU-8.1 — Test: `RenderPersonaLanguage` + routing
- [x] In `internal/tui/screens/persona_language_test.go` (new file):
  - `RenderPersonaLanguage(selection, cursor)` renders:
    - 5 curated region radios with gentilicio-first Spanish labels (e.g. "Argentino (rioplatense, voseo)").
    - "Idioma del usuario" option always present.
    - "Otro… (texto libre)" free-text input always present.
    - Artifacts-in-English checkbox, defaulting ON (checked).
  - Routing test via `Model.Update()`:
    - After selecting `gentle` on `ScreenPersona` → next screen is `ScreenPersonaLanguage`.
    - After selecting `neutral` on `ScreenPersona` → next screen is `ScreenPersonaLanguage`.
    - After selecting `custom` on `ScreenPersona` → next screen is `ScreenPreset` (region screen skipped).
  - Assert `Selection.ArtifactsInEnglish` defaults to `true` when screen initializes.
- [x] Run `go test ./internal/tui/...` — expect RED.

### WU-8.2 — Implement: `persona_language.go` screen
- [x] Create `internal/tui/screens/persona_language.go`:
  - `LanguageRegionOptions()` returning the 7 region options (5 curated + user-language + free-text sentinel).
  - `RenderPersonaLanguage(selection model.Selection, cursor int, freeText string) string` rendering the radios, "Idioma del usuario", free-text input, and checkbox.
  - TUI labels for curated regions use the Spanish gentilicio-first form from `model.RegionMap`.
- [x] Add `ScreenPersonaLanguage` to `model.go` screen enum.
- [x] Update `router.go` routes to include `ScreenPersonaLanguage`.
- [x] Update `model.go` `Update()` to handle `ScreenPersonaLanguage` keystrokes: navigate radios, toggle checkbox, accept free text, proceed to `ScreenPreset`.
- [x] Ensure `Selection.ArtifactsInEnglish` is initialized to `true` when building a new default `Selection`.
- [x] Run `go test ./internal/tui/...` — expect GREEN.

**Acceptance**: Selecting `user-language` sets `Selection.Region = "user-language"`. Free-text entry sets `Selection.Region` to the typed string. Checkbox toggle sets `Selection.ArtifactsInEnglish`.

---

## Work Unit 9 — TUI review screen: show region + artifacts flag

**Satisfies**: R8 (review shows region + flag)

### WU-9.1 — Test: `RenderReview` with region + artifacts
- [x] In `internal/tui/screens/review_test.go` (extend):
  - Assert `RenderReview` with a payload containing `Region: "mexico"` and `ArtifactsInEnglish: true` includes a "Region" or "Language" line showing "mexico" (or its label).
  - Assert `RenderReview` with `ArtifactsInEnglish: false` shows the off-state of the checkbox.
- [x] Run `go test ./internal/tui/...` — expect RED.

### WU-9.2 — Implement: extend `ReviewPayload` + `RenderReview`
- [x] In `internal/planner` (or wherever `ReviewPayload` lives): add `Region string` and `ArtifactsInEnglish bool` fields.
- [x] In `internal/tui/screens/review.go`: render the new fields in the review layout.
- [x] Run `go test ./internal/tui/...` — expect GREEN.

---

## Work Unit 10 — Spec update: persona-behavior-contract

**Satisfies**: R9 delta spec update (spec delta requirement in spec.md §"Delta to existing spec")

### WU-10.1 — Edit spec file
- [x] Edit `openspec/specs/persona-behavior-contract/spec.md` applying all 6 deltas from the spec delta table:
  1. "Neutral Mentor Behavior Parity" — re-frame to `gentle` parity, not `gentleman`.
  2. "Gentleman keeps regional mentor behavior" scenario → "Gentle persona with Rioplatense region injects the voseo directive".
  3. "Artifact Language Independence" + "Gentleman voice does not leak" → re-anchor to `artifactsInEnglish` checkbox and `gentle` naming.
  4. "Claude explicit Gentleman output-style" → replace `gentleman` with `gentle`; output-style is region-neutral.
  5. "Safe Persona Fallback Semantics" → update to `gentle` style + region axis naming.
  6. Scenarios referencing `gentleman` → `gentle` + explicit region.
- [x] No test required (doc edit). Verify the file compiles/reads correctly.

---

## Work Unit 11 — Integration: wire state persistence through TUI install path

**Satisfies**: R6 (state write carries new fields), R2 (second sync reads migrated fields directly)

### WU-11.1 — Test: state write + second-sync no-re-migrate
- [x] In `internal/cli/sync_test.go` (extend):
  - After migrate + first sync, assert that `state.json` contains `"region": "argentina"` and `"artifactsInEnglish": true` (for the `gentleman` migration case).
  - Simulate second sync from the already-migrated state: assert `applyResolvedPersona` does NOT re-apply the legacy alias path (reads `PersonaGentle` from persisted state directly, no alias needed).
  - Assert second-sync output byte-identical to first-sync output.
- [x] Run `go test ./internal/cli/...` — expect RED.

### WU-11.2 — Implement: persist new fields on sync state write
- [x] In `internal/cli/sync.go` (or the sync result handler): after resolving the selection, update the state write to include `Region` and `ArtifactsInEnglish` from the resolved selection.
- [x] In TUI install path (`internal/tui/model.go` or `cli` install handler): ensure `state.Write` includes `Region` and `ArtifactsInEnglish` from the TUI selection.
- [x] Run `go test ./internal/cli/...` — expect GREEN.

**Acceptance**: After a single sync with a legacy `gentleman` state, a subsequent `state.Read` returns `Persona: "gentle"`, `Region: "argentina"`, `ArtifactsInEnglish: true`. The second sync hits no migration path and produces byte-identical output.

---

## Deferred work units (Slice 2+, NOT in this PR)

These units reuse the strip-and-compose pattern established in Slice 1. Do not implement until Slice 1 is merged.

- **DWU-A**: OpenCode/Kilocode persona + output-style assets — strip `## Language` section; wire `composeLanguageDirective` in `inject.go` for `StrategyFileReplace`. (~60–80 lines)
- **DWU-B**: Kimi `language.md` Jinja module — write `language.md` module in `StrategyJinjaModules` case; strip Kimi output-style assets; update `KIMI.md` static template. (~80–100 lines)
- **DWU-C**: Kiro steering-file asset — strip `## Language` from `kiro/persona-gentleman.md`; wire directive in `StrategySteeringFile` case. (~40–60 lines)
- **DWU-D**: Hermes asset — strip `## Language` from `hermes/persona-gentleman.md` and `hermes/persona-neutral.md`; wire directive. (~40–60 lines)
- **DWU-E**: Full TUI language-screen integration testing (end-to-end install path with new state fields). (~40–60 lines of test)

---

## Parallel / sequential dependency graph

```
WU-1 (model/types/region/directive)
  └─► WU-2 (selection + state)         [depends on WU-1: RegionID type]
        └─► WU-3 (validate)             [depends on WU-1: PersonaGentle const]
              └─► WU-4 (sync+migrate)   [depends on WU-1,2,3: Selection.Region, PersonaGentle, normalizePersona]
                    └─► WU-6 (inject)   [depends on WU-1 composeLanguageDirective, WU-5 strip]
WU-5 (stripLanguageSection + assets)    [CAN start after WU-1.2 — no direct WU-2/3/4 dependency]
  └─► WU-6 (inject)                     [depends on WU-5 + WU-4]
WU-7 (TUI persona screen)               [depends on WU-1: PersonaGentle]
  └─► WU-8 (TUI language screen)        [depends on WU-1: RegionID/RegionMap, WU-7: ScreenPersonaLanguage route]
        └─► WU-9 (TUI review)           [depends on WU-2: Selection.Region, WU-8]
              └─► WU-11 (state persist) [depends on WU-4 + WU-8 + WU-2]
WU-10 (spec edit)                       [INDEPENDENT — can run at any point]
```

**Parallelizable after WU-1.2 green**:
- WU-2 + WU-5 can start simultaneously.
- WU-3 can start after WU-2 (or after WU-1 alone, since it only needs the const).
- WU-7 can start after WU-1.2 (only needs `PersonaGentle` const).
- WU-10 is fully independent.

---

## Review Workload Forecast

| Category | Estimate |
|---|---|
| WU-1 (model: types, region, directive) | ~80 lines |
| WU-2 (selection + state) | ~40 lines |
| WU-3 (validate) | ~20 lines |
| WU-4 (sync + migration) | ~80 lines |
| WU-5 (strip function + 3 asset edits) | ~60 lines (code) + ~25 lines net removed from assets |
| WU-6 (inject wiring) | ~80 lines |
| WU-7 (TUI persona screen) | ~40 lines |
| WU-8 (TUI language screen, new file) | ~120 lines |
| WU-9 (TUI review) | ~30 lines |
| WU-10 (spec edit) | ~40 lines |
| WU-11 (state persist integration) | ~40 lines |
| **Tests across all WUs** | ~250–280 lines |
| **Total estimated changed lines** | **~850–900 lines** |

**Chained PRs recommended**: YES

**400-line budget risk**: HIGH — the slice exceeds the 400-line budget by roughly 2x.

**Decision needed before apply**: YES

### Recommended split points

**PR 1 — Motor core (model + state + validate + sync/migration + asset strip)**
WU-1 + WU-2 + WU-3 + WU-4 + WU-5
Estimated: ~380 lines of production code + ~150 lines of tests = ~530 lines
Risk: still above 400. Can split further:
  - PR 1a: WU-1 + WU-2 + WU-3 (~140 lines + tests ~100) = ~240 lines ✓
  - PR 1b: WU-4 + WU-5 (~140 lines + tests ~130) = ~270 lines ✓

**PR 2 — Inject wiring**
WU-6 (~80 lines + tests ~100) = ~180 lines ✓

**PR 3 — TUI + spec**
WU-7 + WU-8 + WU-9 + WU-10 + WU-11 (~270 lines + tests ~100) = ~370 lines ✓

Final recommendation: 4 chained PRs (1a → 1b → 2 → 3), each under 400 lines.
Chain strategy: TBD by orchestrator (stacked-to-main or feature-branch-chain).
