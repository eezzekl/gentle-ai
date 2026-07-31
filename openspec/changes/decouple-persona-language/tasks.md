# Tasks: decouple-persona-language — Motor-first slice

Slice scope: model / state / validate / sync+migration / inject / Claude+generic assets / spec update / tests.
Deferred adapters (OpenCode/Kilocode, Kimi, Kiro, Hermes) and the TUI region screen are listed separately as Slice 2+ deferred units.

Strict TDD is active (`go test ./...`). Every implementation task is preceded by its test task (red → green order).

---

## STATUS — every work unit in this slice is done (2026-07-31)

WU-1 through WU-12 are implemented, tested, and committed. The Decision 5 rework that this file
previously flagged as outstanding has landed in full. Read this block first; everything below is
the unit-by-unit record.

**Test state**: `go test ./...` → **62 packages ok, 1 failure**. The single failure is
`TestEveryProductionRefusalNamesResolutionOrDeclaresByDesign` in `internal/cli`, which is
**pre-existing on `main`** and unrelated to this change — verified by reproducing it on a clean
`origin/main` worktree. `354c5b50` added a refusal in `review_facade.go` without updating the
ratchet baseline. Its fix belongs to `main`:
`GENTLE_AI_REFUSAL_RATCHET_UPDATE=1 go test ./internal/cli -run TestEveryProductionRefusalNamesResolutionOrDeclaresByDesign -count=1`

**Branch**: `feat/decouple-persona-language`, rebased onto `origin/main` `d260cdbc`.
The fork remote still holds the eight PRE-rebase commits, so publishing requires
`--force-with-lease`, never a plain push.

### Divergences from the plan, all deliberate and recorded in place

| Where | What changed and why |
|---|---|
| WU-5 | Planned as a header-bounded `## Language` SECTION strip; shipped as a LINE-level removal of the regional reply bullet, `stripRegionalVoiceDirective`. Those sections also carry region-agnostic language contract that a section strip would silently delete. |
| Decision 1 (design) | "Persona content block only" holds for every agent EXCEPT Claude Code and Kimi, which deliver reply voice through their output-style asset (`voiceLivesInOutputStyle`). |
| Safe fallback | Unknown, invalid, and unreadable persona state now resolve to regionless `neutral`, not to `gentle + argentina`. An empty field in a READABLE state still migrates to `gentle + argentina`. Both arrive as an empty string; only readability separates them. |
| WU-12 | No genuine RED phase existed — WU-3/WU-4 had already shipped the behavior. Non-vacuity was established by mutation testing instead, recorded under WU-12.3. |

### Still open (NOT blockers for this slice)

1. Reply on issue #912 quoting the WU-12 compatibility matrix.
2. Choose the chained-PR split. The forecast below recommends five links, each under 400 lines.
3. Deferred units DWU-A through DWU-E (OpenCode/Kilocode, Kimi, Kiro, Hermes, TUI end-to-end) — Slice 2+ by design, explicitly not in this PR.

---

## ✅ Rework COMPLETE — design Decision 5 (`neutral` is regionless)

WU-1 through WU-11 were originally implemented under the **previous** posture, in which
`gentleman-neutral-artifacts` migrated to `gentle + argentina` and `neutral` carried a
`user-language` region. Design Decision 5 reversed that: `neutral` is regionless, and the alias
migrates to `neutral`.

Every unit below has since been reworked and the code now matches this document.

| Unit | What changed | Done |
|---|---|---|
| WU-1 | `ComposeLanguageDirective` handles an empty region → artifacts clause only, no regional-voice clause. | ✅ |
| WU-3 | `normalizePersona`: `gentleman-neutral-artifacts` → `PersonaNeutral`, no longer `PersonaGentle`. | ✅ |
| WU-4 | Migration matrix rows 2 and 3 both target `neutral` + no region. | ✅ |
| WU-6 | `isGentlePersona` returns **false** for `gentleman-neutral-artifacts`. | ✅ |
| WU-7 / WU-8 | Router sends `gentle` only to `ScreenPersonaLanguage`; `neutral` skips it like `custom`. | ✅ |
| WU-9 | Review renders the regionless state for `neutral` explicitly. | ✅ |
| WU-10 | `persona-behavior-contract` re-anchored — see WU-10.2. | ✅ |
| WU-12 | Compatibility-matrix and order-independence tests requested on issue #912. | ✅ |

Goldens touched by these units were regenerated with `-update`, never hand-merged.

---

## Work Unit 1 — Model layer: types + region map + directive composer

**Satisfies**: R3 (two independent axes), R4 (composed directive), R5 (artifacts checkbox)

### WU-1.1 — Test: `PersonaGentle`, `RegionID` map, `composeLanguageDirective`
- [x] In `internal/model/types_test.go` (create if absent): write table-driven test asserting `PersonaGentle == "gentle"` and that `PersonaGentlemanNeutralArtifacts` is no longer in the selectable constants (compile-level: removed from type block).
- [x] In `internal/model/region_test.go` (new file): write table-driven test for `composeLanguageDirective(region, artifactsInEnglish)` covering:
  - Each curated region (argentina, mexico, colombia, spain, chile) × {true, false} → assert directive string contains the expected language phrase and the correct artifacts clause.
  - `user-language` sentinel → assert "reply in the language the user writes in" variant, no forced region.
  - **REWORK (Decision 5)** — empty region `""` (the `neutral` case) → assert the artifacts clause is present and NO reply-language or regional-voice clause is emitted; assert the result differs from the `user-language` case.
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
  - **REWORK (Decision 5)** — empty region → emit NO reply-language clause at all. Distinct from `user-language`, which does emit one.
  - Unrecognized NON-EMPTY region string → treat as free text, inject verbatim.
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
  - **REWORK (Decision 5)** — `"gentleman-neutral-artifacts"` → `PersonaNeutral`, no error. Add an explicit assertion that it does NOT resolve to `PersonaGentle`; the two aliases have different targets.
  - `""` → `PersonaGentle`, no error (default).
  - `"unknown-value"` → error, and the resolved value is NOT a persona that injects regional voice.
- [x] Run `go test ./internal/cli/...` — expect RED. ✓ Confirmed RED.

### WU-3.2 — Implement: update `normalizePersona`
- [x] In `internal/cli/validate.go`:
  - Update `normalizePersona` switch: add `PersonaGentle`; map `PersonaGentleman` → `PersonaGentle`; **REWORK (Decision 5)** map `PersonaGentlemanNeutralArtifacts` → `PersonaNeutral`.
  - Update empty-string default from `PersonaGentleman` to `PersonaGentle`.
- [x] Run `go test ./internal/cli/...` — expect GREEN. ✓ Confirmed GREEN (full suite passes).

**Acceptance**: `"gentleman"` normalizes to `PersonaGentle` and `"gentleman-neutral-artifacts"` normalizes to `PersonaNeutral`, matching the R1 migration matrix at the validation layer. The selectable set is `{gentle, neutral, custom}` only.

---

## Work Unit 4 — Sync + migration: `applyResolvedPersona` + `BuildSyncSelection`

**Satisfies**: R1 (back-compat migration, MANDATORY), R2 (sync idempotency), R6 (persistence of new fields)

### WU-4.1 — Test: migration matrix + idempotency
- [x] In `internal/cli/sync_test.go` (extend): write table-driven test for `applyResolvedPersona` covering all 5 migration matrix rows:

  **REWORK (Decision 5)** — rows 2 and 3 changed:

  | old persona string | expected style | expected region | expected artifactsInEnglish |
  |---|---|---|---|
  | `"gentleman"` | `gentle` | `argentina` | `true` |
  | `"gentleman-neutral-artifacts"` | `neutral` | `""` (regionless) | `true` |
  | `"neutral"` | `neutral` | `""` (regionless) | `true` |
  | `"custom"` | `custom` | `""` | (N/A — no injection) |
  | `""` | `gentle` | `argentina` | `true` |

  - Explicit assertion: `"gentleman-neutral-artifacts"` and `"neutral"` produce byte-identical resolved tuples.
  - Explicit assertion: `"gentleman-neutral-artifacts"` does NOT equal the `"gentleman"` tuple, and its injected content carries no Rioplatense/voseo directive (#1702 defect 1).
  - Explicit assertion: `artifactsInEnglish` is `true` (not Go zero-value `false`) for every non-custom row.

- [x] Write idempotency test for `RunSyncWithSelection`: inject twice, assert resulting content byte-identical on second run.
- [x] Run `go test ./internal/cli/...` — expect RED.

### WU-4.2 — Implement: migrate + sync selection
- [x] In `internal/cli/sync.go`:
  - Update `applyResolvedPersona` signature to populate `Selection.Region` and `Selection.ArtifactsInEnglish` from the persisted `InstallState` (pass `InstallState` instead of bare string, or use a helper that extracts both fields).
  - Implement the full migration matrix: for each legacy `persona` string, set `selection.Persona`, `selection.Region`, and `selection.ArtifactsInEnglish` explicitly. **Never rely on bool zero-value**; set `true` explicitly in each non-custom branch.
  - Update `BuildSyncSelection` to read `state.Region` and `state.ArtifactsInEnglish` from the persisted `InstallState` and write them into the `Selection`.
- [x] Run `go test ./internal/cli/...` — expect GREEN.

**Acceptance notes**:
- `neutral` style → NO region (empty `RegionID`), per design Decision 5. It is regionless, not `user-language`.
- `ArtifactsInEnglish` explicitly `true` in every non-custom migration branch, and forced `true` for `neutral`.
- Idempotency test green: second sync produces no file changes.

---

## Work Unit 5 — Asset strip: `stripRegionalVoiceDirective` + asset edits

**Satisfies**: R4 (region-neutral base), R9 (Claude/generic assets)

> ### ⚠ Contract narrowed during implementation — this text was rewritten to match the code
>
> This unit was planned as a HEADER-BOUNDED SECTION strip: find `## Language` /
> `## Language Rules`, delete through the next H2. It shipped as a LINE-level strip of the
> single regional bullet, and the function is named `stripRegionalVoiceDirective`, not
> `stripLanguageSection`.
>
> The narrowing was correct and must not be undone. Those sections hold region-AGNOSTIC
> contract — anchoring the reply language to the latest user request, no drift from persona
> wording, no code-switching — which is compatible with ANY region. Deleting the section to
> remove one line would silently drop all of it. `strip.go` carries the same warning at its
> `regionalVoiceDirectives` declaration.

### WU-5.1 — Test: `stripRegionalVoiceDirective` transform + asset guard
- [x] Pure-function tests in `internal/components/persona/strip_test.go`:
  - The regional reply bullet is removed.
  - Content carrying no such directive is returned UNCHANGED — this is what makes the transform a no-op for assets whose voice already lives elsewhere, and idempotent across repeated syncs.
  - Every other line survives, including the region-agnostic language contract.
  - The `## Persona Scope` artifacts guard survives — it names voseo but governs generated artifacts, so stripping it would delete a protection this change relies on.
- [x] Asset guard: no shipped gentle asset carries the hardcoded regional reply line.
- [x] Run `go test ./internal/components/persona/...` — expect RED.

### WU-5.2 — Implement: `stripRegionalVoiceDirective` function
- [x] In `internal/components/persona/strip.go` (new file): implement `stripRegionalVoiceDirective(content string) string`.
  - Logic: remove the exact regional reply bullet listed in `regionalVoiceDirectives`. NOT header-bounded — see the narrowing note above.
  - Deliberately EXCLUDED from the list: the `## Persona Scope` guard ("Never inject Rioplatense slang, voseo, ... into generated code"). It names voseo but governs generated ARTIFACTS, not reply voice, so removing it would delete a protection this change depends on.
  - This is a pure string transform — no file I/O.
- [x] Run `go test ./internal/components/persona/...` — confirm strip tests GREEN.

### WU-5.3 — Strip asset files
- [x] `internal/assets/claude/persona-gentleman.md`: regional reply line gone. This asset ends up with no `## Language` section at all, but that is the residual-channel reduction (Canonical Tone Channel), not this transform.
- [x] `internal/assets/claude/output-style-gentleman.md`: regional reply line gone; **`## Language Rules` KEPT** (still present, verified).
- [x] `internal/assets/generic/persona-gentleman.md`: regional reply line gone; **`## Language` KEPT** (still present, verified).
- [x] Run `go test ./internal/components/persona/...` — asset guard tests GREEN.

**Acceptance**: No hardcoded regional reply directive remains in any of the three Claude/generic
gentle assets, and the region-agnostic language contract in those assets is untouched. The
remaining "Rioplatense"/"voseo" mentions in the shipped assets are the `## Persona Scope`
artifact guard and are supposed to stay.

---

## Work Unit 6 — Inject: wire `composeLanguageDirective` into `personaContent` + rename guard

**Satisfies**: R4 (single composed directive source), R9 (composed directive sole regional source)

### WU-6.1 — Test: `personaContent` with directive + golden idempotency
- [x] In `inject_test.go`:
  - Test `personaContent` (updated signature or via the composition path) with `(AgentClaudeCode, PersonaGentle, RegionArgentina, true)`:
    - Assert result contains exactly ONE language directive line (the composed one).
    - Assert result does NOT contain a second hardcoded Rioplatense line.
    - Assert the composed directive for `(gentle, argentina, true)` is byte-identical to the pre-change gentleman persona's language section output (golden comparison).
  - **REWORK (Decision 5)** — test `personaContent` with `(AgentClaudeCode, PersonaNeutral, "" /* regionless */, true)`: assert neutral contract sections survive, no marked regional voice baked in, no regional directive appended, and the output is byte-identical to the pre-change neutral baseline.
  - Idempotency: call the full inject path twice on a temp file; assert content byte-identical after second call; assert no duplicate directive lines.
  - **REWORK (Decision 5)** — `isGentlePersona` unit test: `PersonaGentle` → true; `PersonaGentleman` → true (alias); `PersonaGentlemanNeutralArtifacts` → **false** (it migrates to neutral; returning true would keep routing it to the Gentleman output-style, which is #1702 defect 1); `PersonaNeutral` → false; `PersonaCustom` → false.
- [x] Run `go test ./internal/components/persona/...` — expect RED. ✓ Confirmed RED.

### WU-6.2 — Implement: wire directive into inject
- [x] In `internal/components/persona/inject.go`:
  - Updated `personaContent` signature to `personaContent(agent, persona, region, artifactsInEnglish)`: call the existing asset read, run `stripRegionalVoiceDirective()`, then append `"\n" + model.ComposeLanguageDirective(region, artifactsInEnglish)`. Agents whose voice lives in the output style (Claude Code, Kimi) return before that append — the composed directive goes onto their output-style asset instead.
  - Renamed `isGentlemanConversationPersona` to `isGentlePersona`. **REWORK (Decision 5)**: it must match `PersonaGentle` and `PersonaGentleman` ONLY — drop `PersonaGentlemanNeutralArtifacts` from the match set. Updated all 4 call sites.
  - Updated `Inject` and `InjectForSync` signatures to accept `region model.RegionID` and `artifactsInEnglish bool`. Updated callers in `sync.go` (InjectForSync) and `run.go` (Inject).
  - Exported `ComposeLanguageDirective` in `internal/model/region.go`.
  - Updated all test callers in `inject_test.go`, `golden_test.go`, `openclaw_integration_test.go`, `persona_language_contract_test.go`.
  - Regenerated 5 affected golden files with `-update` flag.
- [x] Run `go test ./internal/components/persona/...` — expect GREEN. ✓ Confirmed GREEN (full `go test ./...` passes).

**Acceptance**: Golden round-trip test — legacy state `"persona": "gentleman"` → migrate → inject → output byte-identical to pre-change gentleman persona baseline (committed golden file or inline fixture). Legacy `"persona": "gentleman-neutral-artifacts"` → output byte-identical to the pre-change **neutral** baseline, not the gentleman one. `isGentlePersona` matches `PersonaGentle` and `PersonaGentleman` only.

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
- [x] **REWORK (Decision 5)** — In `internal/tui/router.go`: add new route `ScreenPersona → ScreenPersonaLanguage` for `gentle` ONLY; keep `ScreenPersona → ScreenPreset` for `neutral` AND `custom` (both skip the language screen). **Shipped in WU-8** (route requires the screen, which WU-8 introduces).
- [x] Run `go test ./internal/tui/...` — expect GREEN.

**Acceptance**: `PersonaOptions()` returns exactly 3 items. Routing: selecting `neutral` or `custom` skips `ScreenPersonaLanguage` and goes directly to `ScreenPreset`.

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
  - Routing test via `Model.Update()` — **REWORK (Decision 5)**:
    - After selecting `gentle` on `ScreenPersona` → next screen is `ScreenPersonaLanguage`.
    - After selecting `neutral` on `ScreenPersona` → next screen is `ScreenPreset` (region screen skipped), `Selection.Region` stays empty, and `Selection.ArtifactsInEnglish` is set to `true`.
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
  - **REWORK (Decision 5)** — assert `RenderReview` for a `neutral` selection (empty `Region`) renders an explicit "no region" line rather than a blank value, and shows artifacts-in-English as on.
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

### WU-10.2 — Re-edit spec file for Decision 5

WU-10.1 landed under the previous posture. Verified scope of the drift (2026-07-30): exactly ONE
stale sentence, at `openspec/specs/persona-behavior-contract/spec.md:7`, closing the "Neutral
Mentor Behavior Parity" requirement:

> "Regional voice is governed by the independent region axis, not by the style."

That is true for `gentle` and false for `neutral`, which under Decision 5 has no region at all.
The rest of the file is clean — it never pairs `neutral` with `user-language` and never documents
the legacy aliases, so nothing else needs removing.

- [x] Rewrite that sentence: `neutral` is REGIONLESS — it carries no region and receives no composed regional directive; the region axis applies to `gentle` only. Also added the matching `AND` line to the "Neutral receives the same mentor contract" scenario, so the regionless claim is asserted, not only stated.
- [x] Add the legacy alias targets to this spec (currently absent): `gentleman` → `gentle`, `gentleman-neutral-artifacts` → `neutral`. Landed as a new requirement, "Legacy Persona Alias Resolution", with three scenarios (each alias plus idempotency of an already-migrated state).
- [x] No test required (doc edit). Re-read the file end to end for leftover old-posture wording.

**Found in the end-to-end re-read, beyond the two items above:**

- "Canonical Tone Channel" scenario said "persona `gentleman` or `neutral`" — updated to "style `gentle` or `neutral`". Missed by WU-10.1 item 6.
- "Safe Persona Fallback Semantics" named no concrete fallback ("a default-safe style"), which left the requirement unverifiable once the code had to pick one. It now names `neutral` with no region, states that a missing state.json counts as unreadable, and requires implementations NOT to collapse readable-empty with unreadable-empty — both arrive carrying an empty persona value, and only readability separates them. This tracks the fallback fix committed alongside this unit.
- Deliberately LEFT as-is: the `gentleman` mentions at the "Gentleman and Neutral reconciliation" requirement refer to the asset filenames `claude/persona-gentleman.md` and `claude/output-style-gentleman.md`, which still exist under those names. The mention at the "Gentle style with the Rioplatense region" scenario is a negative historical reference ("not from a `gentleman` persona variant") and is still accurate.

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

## Work Unit 12 — Compatibility matrix: N=2 idempotency + #1712 order-independence

**Satisfies**: R1 (migration), R2 (idempotency + merge-order independence)

Requested directly by the maintainer on issue #912 (comment `5112515766`): a migration
compatibility matrix resolving sequencing with #1702, plus idempotency tests. WU-4's existing
idempotency test covers a single sync path; it does NOT cover every matrix row, and nothing
today proves order-independence against PR #1712.

### WU-12.1 — Test: N=2 idempotency across every matrix row

- [x] New file `internal/cli/sync_persona_migration_test.go`.
- [x] Table over the five R1 rows (absent, `gentleman`, `gentleman-neutral-artifacts`, `neutral`, `custom`). For each:
  - Build a legacy `state.json` fixture — written as RAW JSON, not through `state.Write`, because the "absent persona field" row is a shape `InstallState` cannot express (marshalling always emits the key).
  - Run migrate + sync twice.
  - Assert the resolved tuple after run 2 equals run 1.
  - Assert the written `state.json` bytes after run 2 equal run 1.
  - Assert the injected file bytes after run 2 equal run 1.
  - Additionally assert the `persona` value LEFT IN `state.json`: if a legacy alias survives the first sync, every later run re-enters the migration path, and byte-identical output would be masking a state that never converged.
- [x] Run `go test ./internal/cli/...` — **no genuine RED was available**: WU-3/WU-4 were already implemented, so the tests pass on first run. Non-vacuity was established by mutation instead (see below).

### WU-12.2 — Test: #1712 cross-sequence convergence

- [x] In the same file, model both merge orders for `"persona": "gentleman-neutral-artifacts"`:
  - **Sequence A** — apply #1712's remap first (rewrites the value to `"neutral"`), then this change's migration.
  - **Sequence B** — apply this change's migration directly to the original legacy value.
- [x] Assert both sequences produce the identical resolved tuple `{neutral, "", true}`.
- [x] Assert both produce byte-identical `state.json` and byte-identical injected content.
- [x] Assert a subsequent sync on either result takes the plain `neutral` row, never an alias path.
- [x] Run `go test ./internal/cli/...` — GREEN.

**Acceptance**: The test file IS the compatibility matrix. Its table is the artifact quoted in
the reply on issue #912 — if the reply and the test ever disagree, the test is authoritative.

### WU-12.3 — Two findings from writing it, both worth keeping

**1. Reading only `CLAUDE.md` proves nothing about the reply voice.** The first draft compared
`CLAUDE.md` bytes and asserted "no Rioplatense here" for the hybrid. That assertion passed for the
wrong reason: Claude Code delivers its voice through the OUTPUT STYLE, not the persona section
(`voiceLivesInOutputStyle` in `internal/components/persona/inject.go`), so `CLAUDE.md` contains no
regional directive under ANY persona. Caught only because the test carries a vacuity guard asserting
the `gentleman` side DOES contain "Rioplatense" — that guard failed and exposed the mistake. The
capture now reads the whole persona surface: `CLAUDE.md` plus every file in `~/.claude/output-styles/`,
sorted, with explicit absence markers.

**2. The marker assertions guard the ASSET, not the routing.** Verified by mutation: injecting a
Rioplatense clause into `ComposeLanguageDirective`'s empty-region branch leaves them GREEN, because
the neutral path never calls the composer on any agent — `output-style-neutral.md` is written
verbatim. What they do catch is regional voice baked back into the neutral asset itself, which is
WU-5's concern verified end to end. Documented in the test so nobody reads them as covering
Decision 5 routing.

**Mutation results** (each applied to `HEAD`, run, then reverted):

| Mutation | Outcome |
|---|---|
| Hybrid alias resolves to `gentle`+`argentina` again (#1702 defect 1) | Kills all three tests |
| `stateReadable` ignored — unreadable collapses into readable-empty | Kills the WU-4 fallback tests |
| Write-back stops persisting the migrated persona (alias survives) | Kills the matrix + convergence tests |
| Empty region emits the Rioplatense clause | **SURVIVES** — see finding 2 |
| Rioplatense baked into `claude/output-style-neutral.md` | Kills the hybrid test |

Note on tooling: `sd` silently failed to apply a multi-line mutation pattern, which first read as
"the test does not catch this". Always confirm a mutation actually landed before trusting a survival.

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
WU-5 (stripRegionalVoiceDirective + assets) [CAN start after WU-1.2 — no direct WU-2/3/4 dependency]
  └─► WU-6 (inject)                     [depends on WU-5 + WU-4]
WU-7 (TUI persona screen)               [depends on WU-1: PersonaGentle]
  └─► WU-8 (TUI language screen)        [depends on WU-1: RegionID/RegionMap, WU-7: ScreenPersonaLanguage route]
        └─► WU-9 (TUI review)           [depends on WU-2: Selection.Region, WU-8]
              └─► WU-11 (state persist) [depends on WU-4 + WU-8 + WU-2]
WU-10 (spec edit)                       [INDEPENDENT — can run at any point]
WU-12 (compatibility matrix tests)      [depends on WU-3 + WU-4 rework: alias + matrix targets]
```

**Parallelizable after WU-1.2 green**:
- WU-2 + WU-5 can start simultaneously.
- WU-3 can start after WU-2 (or after WU-1 alone, since it only needs the const).
- WU-7 can start after WU-1.2 (only needs `PersonaGentle` const).
- WU-10 is fully independent.

---

## Review Workload Forecast — MEASURED, superseding the pre-implementation estimate

The original forecast projected ~1080-1130 total changed lines and five PRs each under 400. Both
numbers were wrong. Real total against `origin/main` is **~5000 changed lines**. The estimate
counted production lines and consistently under-counted tests, goldens, and the SDD artifacts.

### Delivered chain (feature-branch-chain, user-selected 2026-07-31)

`additions + deletions`, the metric `Check PR Cognitive Load` uses.

| Link | Tip | Contents | Changed lines | Budget |
|---|---|---|---|---|
| `...-pr0` | `8590a7e4` | SDD artifacts (proposal, spec, design, tasks) | 1148 | exception |
| `...-pr1` | `c65d4cfd` | WU-1, WU-2, WU-3 | 628 | exception |
| `...-pr2` | `b62ea6aa` | WU-4, WU-5 | 707 | exception |
| `...-pr3` | `957b443c` | WU-6 | 621 | exception |
| `...-pr4` | `fa04ac87` | WU-7 through WU-11 | 855 | exception |
| `...-pr5` | `dd6fb1d8` | D5 rework, fixture pins, safe fallback, WU-12, doc corrections | 1584 | exception |

Base order: pr0 on `main`, then each link on the previous.

### Why `size:exception` rather than a finer cut

Re-cutting boundaries cannot reach 400, because **four individual commits already exceed it on
their own**: the SDD artifacts (1148), WU-1/2/3 (628), WU-4/5 (707), and WU-6 (621). A PR cannot be
smaller than the commit it carries. Reaching 400 would require decomposing those commits internally,
which produces intermediate states that were never compiled or tested as a unit. That is the same
objection that blocks redistributing the Decision 5 rework, and it is not worth trading proven code
for a smaller diff.

An eight-link cut was measured before choosing this: pr0 1148, pr1 628, pr2 707, pr3 621,
pr4a 647, pr4b 208, pr5a 828, pr5b 830. Only one link of eight lands under 400.

Worth noting for reviewers: the TUI commits were never the problem. WU-7 is 134 lines, WU-9 is 112,
WU-10 is 87, WU-11 is 121. The weight sits in the model/state/inject layers and their tests.

### Rationale to record on each PR

- **pr0** — documentation only, no executable content. It reads far faster than its line count
  suggests, and splitting planning prose across PRs would leave each implementation link
  referencing a spec that does not exist yet.
- **pr1, pr2, pr3** — one atomic work-unit group each, already the smallest shippable slice. Tests
  are the bulk (pr1 carries ~452 lines of test alone), and separating tests from the code they
  cover would leave a link that is green for the wrong reason.
- **pr4** — five small commits (134, 401, 112, 87, 121) that only exceed the budget when grouped.
  Splittable if a maintainer prefers: WU-7/8/9 at 647 and WU-10/11 at 208.
- **pr5** — carries the Decision 5 rework, which cannot move earlier (see below), plus the safe
  fallback fix and the WU-12 matrix. Splittable into 828 + 830 if preferred.

The `size:exception` label is maintainer-applied. Contributors request it with rationale; they do
not set it themselves.

### Decision 5 rework placement — the original plan was NOT achievable

The pre-implementation plan said each reworked unit would ship inside the PR that introduced it, so
that no PR contradicts the spec at its own merge point. That is impossible without rewriting history,
because the rework commit was authored against the FINAL tree and several of its files are modified
by LATER work units:

| File in the rework | Modified later by | Cannot ship before |
|---|---|---|
| `internal/model/region.go`, `region_test.go` | WU-6 | pr3 |
| `internal/components/persona/inject.go` | WU-6 | pr3 |
| `internal/cli/validate.go`, `internal/tui/model.go` | WU-8 | pr4 |
| `internal/cli/sync.go`, `sync_test.go` | WU-11 | pr4 |

Only `internal/model/types.go` and `internal/cli/persona_language_contract_test.go` depend solely on
the first commit. **Accepted consequence: pr1 through pr4 carry the pre-Decision-5 posture
(`gentleman-neutral-artifacts` resolving toward `gentle + argentina`) and pr5 corrects it.** Reviewers
of the earlier links should read them as the history that produced the fix, not as the final contract.
