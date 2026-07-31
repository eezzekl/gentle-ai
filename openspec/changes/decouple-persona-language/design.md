# Design: Decouple persona style from language/region

Split the coupled persona axis into two orthogonal axes — **style** (`gentle | neutral | custom`) and **language/region** — that converge at a single Go composition point. The regional voice becomes ONE directive line composed at build time from a curated region map plus free text. No per-region asset files, no new runtime template engine. This document resolves the 4 open technical decisions and maps the components, data flow, and ADRs that the tasks phase will turn into steps.

## Quick path (the architecture in 5 moves)

1. **Model**: add `PersonaGentle`, `RegionID` + curated `map[RegionID]Label`, and a region-directive builder. Drop `PersonaGentlemanNeutralArtifacts`.
2. **Compose**: `personaContent()` returns a region-NEUTRAL persona body; a new `composeLanguageDirective(region, artifactsInEnglish)` appends ONE language line. Same line feeds every agent.
3. **Persist**: `InstallState` gains `Region string` + `ArtifactsInEnglish bool`; `MergeAgents` carries both; migration maps legacy persona values explicitly (no zero-value traps).
4. **Inject**: build-time append for all non-Kimi agents; Kimi gets the directive as a dedicated `language.md` Jinja module. Output-style assets become language-agnostic (single source of truth = persona content).
5. **Select**: TUI gains a region screen (radios + "Idioma del usuario" + free text) and an artifacts-in-English checkbox; `gentle`/`neutral` route through it, `custom` skips it.

## The 4 open decisions (RESOLVED)

### Decision 1 — Language line placement: persona content ONLY

**Chosen**: Inject the composed language directive into the **persona content block only**. Strip the `## Language` / `## Language Rules` section from BOTH the persona asset AND the Claude/Kimi output-style assets, making the output-style govern tone/personality only.

**Rationale**: Single source of truth. Today the regional voice is duplicated — it lives in the persona asset's `## Language` section (verified: `claude/persona-gentleman.md:37-42`) AND in the output-style asset's `## Language Rules` section (verified: `claude/output-style-gentleman.md:40-48`). Two copies of a runtime behavior rule is a consistency bug waiting to happen the moment the region is parameterized: a user on Mexican voice would get a persona block saying "Mexicano (tuteo)" and an output-style block still hardcoding Rioplatense voseo. Collapsing the directive into one layer eliminates that contradiction by construction.

**Why persona content and not output-style**: The persona content block is the universal layer — every agent strategy (`MarkdownSections`, `FileReplace`, `AppendToFile`, `InstructionsFile`, `SteeringFile`, `JinjaModules`) writes it. The output-style layer is Claude/Kimi-only. Putting the single source of truth in the universal layer means one code path (`composeLanguageDirective` → append to `personaContent` return value) covers all agents; the Claude/Kimi output-style files simply stop carrying language rules.

**Tradeoffs**:
- (+) One composition point; no cross-layer drift; adding a region is a one-line map entry.
- (+) Output-style assets shrink and stop duplicating the Persona Scope / Language blocks.
- (−) The Claude output-style file loses its self-contained language rule, so reading it in isolation no longer shows the language contract. Mitigation: the output-style retains the `## Persona Scope` English-artifacts guard (it is a tone/scope concern, not a regional-voice concern) and a one-line pointer comment that reply language is governed by the managed persona block.
- (−) Requires editing both asset families in slice 1 for Claude/generic. Accepted: it is the same strip operation twice.

**Rejected**: Inject into both layers. Rejected because keeping two parameterized copies in sync is exactly the failure mode the refactor exists to remove.

### Decision 2 — Strip the `## Language` section; KEEP the `gentleman.md` output-style filename

Two sub-decisions:

**2a. Strip vs. rename assets → STRIP (compose at build time), do not create `persona-gentle.md` asset files.**

The persona assets already isolate the regional voice in a dedicated `## Language` section with a stable header (verified across `claude/persona-gentleman.md:37`, and the proposal confirms all six agent families share the structure). Stripping that one section at build time and appending the composed directive is a clean, testable transform on a known boundary. Creating parallel `persona-gentle.md` files would duplicate ~70 lines × 6 agents of persona body that is 100% identical except for the stripped section — a maintenance liability for zero behavioral gain.

- Implementation: keep reading the existing `persona-gentleman.md` assets, but run them through `stripLanguageSection(content)` before appending `composeLanguageDirective(...)`. The strip targets the `## Language` H2 section by header, removing from that header up to the next H2 (or EOF).
- Guard: a unit test asserts (a) the composed directive is the ONLY regional-voice line in the final string, and (b) the `## Persona Scope`, `## Personality`, `## Tone`, `## Philosophy`, `## Behavior` sections survive byte-for-byte. This blindajes against the "strip removed more than the regional line" risk.

**Asset filenames on disk stay `persona-gentleman.md`** in `internal/assets/...`. These are embedded source assets, never written to the user's disk under that name, so renaming them buys nothing and would churn every `assets.MustRead` call site. The `gentle` identity lives in the model/code layer, not the asset filename.

**2b. Output-style filename on the USER'S disk → KEEP `gentleman.md` (do not migrate to `gentle.md`).**

This is the ORPHAN-risk decision. Today inject writes `~/.claude/.../output-styles/gentleman.md` and sets `"outputStyle": "Gentleman"` (verified `inject.go:349,363`); uninstall removes exactly `gentleman.md` (verified `service.go:454`); non-gentleman sync cleanup removes exactly `gentleman.md` and the settings value `"Gentleman"` (verified `inject.go:404,417`).

If we rename the written file to `gentle.md` and the settings value to `"Gentle"`, then **every existing install that already wrote `gentleman.md` + `"Gentleman"` becomes an orphan**: the new uninstall/cleanup code removes `gentle.md` and leaves the old `gentleman.md` on disk forever, plus a dangling `"outputStyle": "Gentleman"` in settings.json pointing at a file the tool no longer manages. That directly violates the migration-safety constraint (#1118): existing users must not silently degrade.

**Chosen**: Keep the written filename `gentleman.md` and the settings value `"Gentleman"`. The user-facing persona is now "gentle" but the on-disk output-style artifact name is an implementation detail that is NOT worth a migration. The internal `PersonaGentle` ID maps to the existing `gentleman.md` / `"Gentleman"` output-style artifact.

**Tradeoffs**:
- (+) Zero orphan risk; uninstall and sync cleanup keep working unchanged for both legacy and migrated installs.
- (+) Byte-identical output-style artifact for a migrated user → satisfies the golden round-trip idempotency requirement (#1118 trap 2).
- (−) A small naming inconsistency: model says `gentle`, disk says `gentleman.md`. Documented here and in a code comment so it is intentional, not an oversight.

**Rejected**: Rename to `gentle.md` and teach uninstall/cleanup to remove BOTH names. Rejected because it adds a dual-name cleanup matrix, an orphaned-settings-value migration, and new tests for negligible benefit — the filename is invisible to users. If a future change ever surfaces the filename, it can be migrated then with a dedicated, tested rename path. (Pre-existing note: uninstall does not currently remove `neutral.md`; that is an existing gap, out of scope here.)

### Decision 3 — `artifactsInEnglish`: always-on opt-out checkbox, injects a distinct directive

**Chosen**: Confirm the checkbox is a real, always-surfaced opt-out, default **ON**, that toggles WHICH artifact-language directive is composed — not a no-op.

**Why it is meaningful despite defaulting ON**: Today the `## Persona Scope` block hardcodes "Generated technical artifacts default to English regardless of the active persona or conversation language" (verified `claude/persona-gentleman.md:33`). That is an unconditional English mandate. The checkbox makes that mandate **conditional and user-controlled**:

- `artifactsInEnglish = true` (default) → compose the existing English-artifacts directive (preserves today's behavior exactly).
- `artifactsInEnglish = false` → compose an in-language artifacts directive: "Generated technical artifacts (code comments, docs, commit messages, UI copy) follow the selected reply language; use neutral/professional Spanish for the {region} variant." This is the opt-out path for users extending a non-English project.

**Validation against the model/state/inject flow**:
- **Model**: `Selection.ArtifactsInEnglish bool`. Because Go's zero value is `false` and the product default is `true`, the field MUST be set explicitly everywhere a Selection is built (TUI default, migration, sync load) — never left to default. This is migration trap 1 (#1118) and is called out as a hard invariant below.
- **State**: `InstallState.ArtifactsInEnglish bool` with `json:"artifactsInEnglish,omitempty"`. CAUTION: `omitempty` on a bool drops `false` from JSON. That is acceptable here ONLY because `false` is the non-default and we always reconstruct intent through migration/normalization, but the safer choice is **no `omitempty`** so the persisted value is unambiguous. **Chosen: omit `omitempty`** for `artifactsInEnglish` to make the persisted bool explicit and round-trippable. (Contrast: `region` keeps `omitempty` since empty string is a meaningful "no region / follow user".)
- **Inject**: `composeLanguageDirective` takes `artifactsInEnglish` and selects the artifacts clause. The persona asset's `## Persona Scope` block keeps the structural guard (artifacts are NOT styled by persona voice); the composed directive supplies the specific language target.

**Tradeoffs**:
- (+) Genuine user control; the checkbox is not inert.
- (+) Default ON preserves current behavior and migration target for every legacy persona.
- (−) Two artifact-language directive variants to maintain and test. Accepted: two short string branches, both unit-tested.

**Rejected**: Drop the checkbox and keep English artifacts unconditional. Rejected because the proposal/product round locked it as a first-class opt-out (#1115 decision 3), and non-English project users are a real audience.

### Decision 4 — Kimi language as a dedicated `language.md` Jinja module

**Chosen**: For Kimi (the only `StrategyJinjaModules` agent), the composed language directive goes into a **dedicated `language.md` module**, included by `KIMI.md`, NOT folded into the `persona.md` module.

**Rationale**: Kimi already separates concerns into modules — `persona.md` and `output-style.md` are written as distinct includes (verified `inject.go:285-309`). The whole point of this refactor is to make language an orthogonal axis; modeling it as its own module mirrors that orthogonality and keeps the persona module byte-stable relative to the non-language content. A third `{% include "language.md" %}` line in the static `KIMI.md` template makes the seam explicit and lets the language directive evolve without touching the persona module.

**Why not fold into `persona.md`**: Folding would re-couple language into the persona body for exactly one agent, contradicting the universal "language is a separate concern" model and making Kimi's persona module diverge from the clean strip-and-compose shape used everywhere else. A dedicated module keeps Kimi consistent with the architecture.

**Shape** (designed now, BUILT in a deferred slice — Kimi is not in the Motor first slice):
- `KIMI.md` static template gains `{% include "language.md" %}` (one line, alongside the existing persona/output-style includes).
- The `StrategyJinjaModules` case writes a third module file `language.md` whose content is `composeLanguageDirective(region, artifactsInEnglish)` — the SAME function that produces the appended line for all other agents. One composition source, two delivery shapes (append vs. module file).
- Kimi `output-style-gentleman.md` / `output-style-neutral.md` get the same Language-section strip as Claude (Decision 1), so the directive is not duplicated in Kimi's output-style module either.

**Tradeoffs**:
- (+) Clean orthogonality; persona module stays language-free; one shared composer.
- (+) Future language-only changes touch one Kimi module, not the persona body.
- (−) One extra module file + one template include line. Negligible.

**Rejected**: Embed the directive in `persona.md`. Rejected for re-coupling the axes for a single agent and breaking architectural symmetry.

## Component map and data flow

```
                 TUI selection                     CLI sync (regenerate)
        ┌───────────────────────────┐        ┌───────────────────────────┐
        │ ScreenPersona (gentle/    │        │ state.json                │
        │   neutral/custom)         │        │  persona, region,         │
        │ ScreenPersonaLanguage     │        │  artifactsInEnglish        │
        │  - region radios          │        └─────────────┬─────────────┘
        │  - "Idioma del usuario"   │                      │ BuildSyncSelection
        │  - "Otro…" free text      │                      │ + applyResolvedPersona
        │  - [x] artifacts in EN    │                      │   (MIGRATION here)
        └─────────────┬─────────────┘                      │
                      │ Selection{Persona, Region,         │
                      │           ArtifactsInEnglish}      │
                      └──────────────┬─────────────────────┘
                                     ▼
                       internal/model: composeLanguageDirective(region, artInEn)
                                     │  (region map lookup OR free-text OR follow-user)
                                     ▼
              internal/components/persona/inject.go
                personaContent(agent, persona)            // region-NEUTRAL body
                  └─ stripLanguageSection(asset)
                content = body + "\n" + directive          // single append point
                                     │
              ┌──────────────────────┴───────────────────────┐
              ▼ (all non-Kimi: build-time append)             ▼ (Kimi: JinjaModules)
   MarkdownSections / FileReplace / AppendToFile /   persona.md (body, no language)
   InstructionsFile / SteeringFile write `content`   language.md (directive)  ← NEW
   Output-style assets: tone/personality only        output-style.md (no language)
   (gentleman.md filename KEPT; "Gentleman" value)
```

### Integration points

| Layer | File(s) | Change |
|-------|---------|--------|
| Persona IDs | `internal/model/types.go` | Add `PersonaGentle = "gentle"`; remove `PersonaGentlemanNeutralArtifacts`. Keep `PersonaNeutral`, `PersonaCustom`. |
| Region model | `internal/model/types.go` | New `RegionID string` + curated `map[RegionID]Label` (AR/MX/CO/ES/CL, gentilicio-first labels) + sentinel IDs for "Idioma del usuario" and free-text. |
| Directive builder | `internal/model` (new func `composeLanguageDirective`) | Pure func: `(region, artifactsInEnglish) → string`. Region map → directive; free text → raw; "Idioma del usuario" → follow-user clause. Appends the artifacts clause. |
| Selection | `internal/model/selection.go` | Add `Region string`, `ArtifactsInEnglish bool`. |
| State | `internal/state/state.go` | Add `Region string json:"region,omitempty"`, `ArtifactsInEnglish bool json:"artifactsInEnglish"` (NO omitempty). `MergeAgents` carries both. |
| Validation | `internal/cli/validate.go` | `normalizePersona()` accepts `gentle|neutral|custom` + back-compat aliases `gentleman`, `gentleman-neutral-artifacts` → `gentle`. |
| Sync + migration | `internal/cli/sync.go` | `applyResolvedPersona()` migration matrix (below); `BuildSyncSelection` loads `Region` + `ArtifactsInEnglish`. |
| Injection | `internal/components/persona/inject.go` | `stripLanguageSection()` + append `composeLanguageDirective()`; `isGentlemanConversationPersona()` → `isGentlePersona()` (matches `PersonaGentle`); Kimi `language.md` module; output-style dispatch unchanged filenames. |
| Persona assets | `internal/assets/{claude,generic}/persona-gentleman.md` (slice 1) | Strip `## Language` section only. Other agents deferred. |
| Output-style assets | `internal/assets/claude/output-style-gentleman.md`, `output-style-neutral.md` (slice 1) | Strip `## Language Rules` section; keep `## Persona Scope`. |
| TUI | `internal/tui/screens/persona.go`, NEW `persona_language.go`, `model.go`, `router.go`, `screens/review.go` | Drop hybrid option; add region screen + checkbox; route gentle/neutral → language screen, custom skips; review shows region + flag. |
| Spec | `openspec/specs/persona-behavior-contract/spec.md` | Reword to decoupled axes. |
| Uninstall | `internal/components/uninstall/service.go` | NO CHANGE to filenames (Decision 2b keeps `gentleman.md`). |

## Migration (mandatory safety net — #1118)

`applyResolvedPersona()` in `sync.go` is the single migration choke point. The matrix:

| legacy `state.json` persona | migrates to |
|---|---|
| `gentleman` | `gentle` + region=Rioplatense + artifactsInEnglish=`true` |
| `gentleman-neutral-artifacts` | `gentle` + region=Rioplatense + artifactsInEnglish=`true` (CONVERGES with `gentleman` — test asserts byte-identical output, proving the hybrid was a no-op) |
| `neutral` | `neutral` + region="Idioma del usuario" + artifactsInEnglish=`true` |
| `custom` | `custom` (no region, no injection) unchanged |
| empty / absent | `gentle` + Rioplatense + artifactsInEnglish=`true` (preserve today's fallback) |

**Two hard invariants the tests MUST blindar**:
1. **No bool zero-value trap**: migration sets `ArtifactsInEnglish = true` EXPLICITLY for every legacy case. Never rely on the Go zero value. A legacy state.json with no `artifactsInEnglish` key must NOT flip artifacts to Spanish on sync.
2. **Idempotency / golden round-trip**: legacy state → migrate → inject → BYTE-IDENTICAL files vs. pre-change baseline. The strongest test is a golden comparison, not field-by-field assertion. An existing user runs `sync` and sees ZERO spurious diff and no degraded persona. The Decision 2b filename-keep is what makes this byte-identical possible.

## Testability (Strict TDD active — `go test ./...`)

Every touchpoint is a pure function or a deterministic file transform:

- `composeLanguageDirective(region, artifactsInEnglish)` — pure; table test across AR/MX/CO/ES/CL + "Idioma del usuario" + free text × {true,false}.
- `stripLanguageSection(content)` — pure transform; assert only `## Language` removed, all other H2 sections survive.
- `personaContent(agent, persona)` — assert returned string contains the composed directive and is region-neutral before composition.
- `isGentlePersona(persona)` — trivial pure func.
- `normalizePersona()` — table test including legacy aliases.
- `applyResolvedPersona()` — the full migration matrix above, one case per row.
- State round-trip — `InstallState` JSON marshal/unmarshal with new fields; assert `artifactsInEnglish=false` survives (the `omitempty` removal).
- Golden idempotency — legacy state → inject → compare against committed golden files for Claude + generic.
- TUI render — `RenderPersonaLanguage()` shows curated radios + "Idioma del usuario" + free-text area + checkbox; `RenderPersona()` no longer lists `gentleman-neutral-artifacts`; review shows region + flag.

## Risks and assumptions

| Risk / assumption | Handling |
|---|---|
| `stripLanguageSection` over-strips (removes more than the regional line). | Header-bounded strip (`## Language` → next H2/EOF) + survival test for all other sections. Slice-1 scoped to Claude/generic so the transform is proven before fanning out. |
| `personaContent()` signature/return change ripples to all callers and tests. | Keep the signature; change only the return value (append after strip). The composition happens inside `personaContent` or in the single caller before write — no new params on the hot path. Validate with existing inject tests first. |
| Bool zero-value flips artifacts to Spanish on sync. | Explicit `true` in every migration branch; no `omitempty` on `artifactsInEnglish`; round-trip test for `false`. |
| Output-style filename rename would orphan legacy files. | Decision 2b: keep `gentleman.md` + `"Gentleman"`. No rename, no orphan. |
| Kimi shape designed but unbuilt could drift from the universal composer. | Kimi `language.md` MUST call the same `composeLanguageDirective` — enforced by sharing the function, not re-implementing. |
| Assumption: all six persona assets share the `## Language` H2 boundary. | Verified for `claude/persona-gentleman.md`; proposal asserts the rest. Deferred-slice tasks must re-verify per agent before stripping. |
| `neutral` + curated region produces coherent directive. | Same template for all styles; `neutral` simply omits the mentor voice. Covered by `composeLanguageDirective` table test. |

## Next step

Proceed to `sdd-tasks` once the spec is also ready. Tasks must encode the migration matrix and the golden idempotency test as first-class, non-optional acceptance steps, and scope the first slice to Claude + generic per the proposal.
