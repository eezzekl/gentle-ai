# Spec Delta: Decouple persona style from language/region

This delta defines what MUST be true after the `decouple-persona-language` change is
applied. It describes WHAT the system must guarantee, not HOW to implement it. It is the
Motor-first slice: model, state, validation, sync + migration, injection composition, and
Claude + generic assets. Deferred adapters (OpenCode/Kilocode, Kimi Jinja module, Kiro,
Hermes) are out of scope here and are listed in "Deferred follow-up".

The two mandatory safety-net requirements — back-compat migration and sync idempotency —
are first-class requirements below (R1, R2), not footnotes.

## Terminology

| Term | Meaning |
|------|---------|
| Style axis | The personality dimension: `gentle` (teaching-first), `neutral` (no marked personality), `custom` (no injection, follow user). |
| Region axis | The reply language/region dimension, orthogonal to style. A curated `RegionID`, the always-present "Idioma del usuario", or free text. |
| `artifactsInEnglish` | Boolean toggle: generated artifacts (code, comments, commits, docs) in English (true) vs. the selected reply language (false). Default true (opt-out). |
| Composed directive | The single language line built in Go from the region selection and appended into persona content at inject time. |
| Curated regions v1 | Argentina, Mexico, Colombia, Spain, Chile (locked, decision #1115). |

## Curated regions v1 (locked)

| RegionID | TUI label (Spanish, gentilicio-first) | Composed directive intent (English) |
|----------|----------------------------------------|--------------------------------------|
| `argentina` | Argentino (rioplatense, voseo) | Reply in warm, natural Argentine Spanish (Rioplatense, voseo). |
| `mexico` | Mexicano (tuteo) | Reply in warm, natural Mexican Spanish (tuteo). |
| `colombia` | Colombiano (paisa) | Reply in warm, natural Colombian Spanish (paisa). |
| `spain` | Español (España) | Reply in warm, natural Spanish from Spain. |
| `chile` | Chileno | Reply in warm, natural Chilean Spanish. |
| `user-language` | Idioma del usuario | Reply in the language the user writes in; do not force a region. |
| free text ("Otro…") | Otro… (texto libre) | Reply using the user-provided free-text language/region instruction, injected as-is. |

The TUI labels MAY be Spanish. All other generated artifacts (the injected directive,
identifiers, comments, docs) MUST be English.

---

## Requirements

### Requirement: R1 — Back-compat state migration (MANDATORY safety net)

Every existing `state.json` persona value MUST map deterministically onto the new two-axis
model (`style` + `region` + `artifactsInEnglish`). Migration MUST be explicit and tested.
No existing user may silently degrade to `neutral` or lose the teaching voice on upgrade.

The migration matrix is authoritative:

| Old `persona` value | → `style` | → `region` | → `artifactsInEnglish` |
|---------------------|-----------|------------|------------------------|
| `gentleman` | `gentle` | `argentina` (Rioplatense) | `true` |
| `gentleman-neutral-artifacts` | `gentle` | `argentina` (Rioplatense) | `true` |
| `neutral` | `neutral` | `user-language` (Idioma del usuario) | `true` |
| `custom` | `custom` | none (no region) | n/a (no injection) |
| empty / absent | `gentle` | `argentina` (Rioplatense) | `true` |

`gentleman-neutral-artifacts` is a verified zero-delta variant of `gentleman` (both land in
the same `personaContent()` default case; the gentleman asset already mandates English
artifacts). It MUST converge with `gentleman` — both map to the identical target tuple.

`artifactsInEnglish` MUST be set to `true` explicitly during migration for all non-custom
cases. It MUST NOT be left to fall to the Go bool zero-value (`false`), which would flip
artifacts to the selected language on the next sync. This is a silent-degradation trap and
is a required test.

#### Scenario: gentleman migrates to gentle + Rioplatense + English artifacts

- GIVEN a `state.json` with `"persona": "gentleman"` and no `region` or `artifactsInEnglish` fields
- WHEN persona resolution/migration runs
- THEN the resolved style is `gentle`
- AND the resolved region is `argentina` (Rioplatense, voseo)
- AND `artifactsInEnglish` is `true`

#### Scenario: gentleman-neutral-artifacts converges with gentleman (proves hybrid was a no-op)

- GIVEN one `state.json` with `"persona": "gentleman"` and another with `"persona": "gentleman-neutral-artifacts"`
- WHEN migration runs on both
- THEN both resolve to the identical tuple `{style: gentle, region: argentina, artifactsInEnglish: true}`
- AND the injected persona content for both is byte-identical

#### Scenario: neutral migrates to neutral + user-language + English artifacts

- GIVEN a `state.json` with `"persona": "neutral"` and no `region` or `artifactsInEnglish` fields
- WHEN migration runs
- THEN the resolved style is `neutral`
- AND the resolved region is `user-language` ("Idioma del usuario")
- AND `artifactsInEnglish` is `true`

#### Scenario: custom migrates unchanged with no region and no injection

- GIVEN a `state.json` with `"persona": "custom"`
- WHEN migration runs
- THEN the resolved style is `custom`
- AND no region is set
- AND no persona/language directive is injected

#### Scenario: empty or absent persona preserves the gentle Rioplatense default

- GIVEN a `state.json` with an empty or absent `persona` field
- WHEN migration runs
- THEN the resolved style is `gentle`
- AND the resolved region is `argentina` (Rioplatense)
- AND `artifactsInEnglish` is `true`

#### Scenario: artifactsInEnglish is never left at the bool zero-value during migration

- GIVEN a legacy `state.json` that has no `artifactsInEnglish` field
- WHEN migration resolves a non-custom persona
- THEN `artifactsInEnglish` is explicitly `true`
- AND it is NOT the Go bool zero-value (`false`) inherited by omission
- AND a subsequent sync does not flip generated artifacts to the selected language

---

### Requirement: R2 — Sync idempotency (MANDATORY safety net)

Running sync repeatedly (TUI or CLI) on the same resolved selection MUST NOT drift, MUST NOT
re-inject duplicate language directives, and MUST NOT degrade the persona. The strongest form
of this guarantee is a round-trip: an existing user's old state → migrate → inject → sync
produces output byte-identical to the prior baseline, with zero spurious diff.

#### Scenario: Repeated sync produces identical injected output

- GIVEN a resolved selection (style + region + artifactsInEnglish)
- WHEN sync injects persona content N times in succession
- THEN the injected file content after each run is byte-identical
- AND no duplicate language directive lines accumulate
- AND no duplicate injection markers accumulate

#### Scenario: Existing gentleman user runs sync after upgrade with zero spurious diff

- GIVEN an existing `state.json` with `"persona": "gentleman"` produced by the pre-change version
- WHEN the user upgrades and runs sync
- THEN the migrated selection is `{gentle, argentina, artifactsInEnglish: true}`
- AND the injected persona content is byte-identical to the pre-change gentleman baseline
- AND the user sees no degraded persona and no spurious file diff

#### Scenario: Migrated state persists the new fields on first sync without re-migrating

- GIVEN a legacy `state.json` migrated in-memory to the two-axis model
- WHEN sync writes state back
- THEN `region` and `artifactsInEnglish` are persisted explicitly in `state.json`
- AND a second sync reads the already-migrated fields directly without re-applying legacy aliases
- AND the second sync output is byte-identical to the first

---

### Requirement: R3 — Two independent axes in the selection model

The system MUST represent personality style and language/region as two independent
selections. Style MUST be one of `gentle`, `neutral`, `custom`. The `gentleman` and
`gentleman-neutral-artifacts` persona IDs MUST be removed from the selectable set (retained
only as back-compat migration aliases per R1).

#### Scenario: Style axis exposes exactly gentle, neutral, custom

- GIVEN the persona style options are enumerated
- WHEN the selectable style set is inspected
- THEN it is exactly `{gentle, neutral, custom}`
- AND it does not include `gentleman` or `gentleman-neutral-artifacts` as selectable options

#### Scenario: Region selection is independent of style

- GIVEN a selected style of `gentle` or `neutral`
- WHEN a region is selected
- THEN any curated region, "Idioma del usuario", or free text is selectable regardless of which of the two styles was chosen
- AND the region selection does not change the style selection

#### Scenario: Custom style carries no region and triggers no injection

- GIVEN a selected style of `custom`
- WHEN the selection is finalized
- THEN no region is associated with the selection
- AND no persona content or language directive is injected

---

### Requirement: R4 — Composed language directive (single source of regional voice)

Regional voice MUST be injected as ONE composed directive line built in Go from the region
selection, appended into the persona content. There MUST be no per-region asset files. The
curated regions MUST live in a Go map keyed by `RegionID`. The persona base assets MUST be
region-neutral (no hardcoded Rioplatense `## Language` line); the composed directive is the
single source of the regional voice.

#### Scenario: Curated region injects its composed directive

- GIVEN a selected style of `gentle` and region `mexico`
- WHEN persona content is composed
- THEN the content contains exactly one language directive instructing reply in warm, natural Mexican Spanish (tuteo)
- AND the base gentle asset contains no hardcoded Rioplatense/voseo language line

#### Scenario: Idioma del usuario injects a no-forced-region directive

- GIVEN region `user-language` ("Idioma del usuario")
- WHEN persona content is composed
- THEN the directive instructs replying in the language the user writes in without forcing a region

#### Scenario: Free-text region is injected as-is

- GIVEN a free-text region value (e.g. "yucateco")
- WHEN persona content is composed
- THEN the free-text value is injected into the directive verbatim, without curation or validation

#### Scenario: No per-region asset files exist

- GIVEN the asset tree after the change
- WHEN persona/output-style assets are enumerated
- THEN there are no per-region `.md` files
- AND adding a new curated region requires only a new entry in the Go region map

#### Scenario: Gentle + Rioplatense reproduces today's regional voice exactly

- GIVEN a selected style of `gentle` and region `argentina`
- WHEN persona content is composed
- THEN the regional voice directive corresponds to warm, natural Argentine Spanish (Rioplatense, voseo)
- AND the resulting behavior matches the pre-change gentleman regional voice

---

### Requirement: R5 — Artifacts-in-English checkbox

`artifactsInEnglish` MUST be a checkbox in the language section, defaulting ON (opt-out).
When ON, generated artifacts (code, comments, commits, docs) MUST default to English
regardless of the selected reply region. When OFF, the injected artifact-language directive
MUST permit artifacts in the selected reply language. The two states MUST inject distinct
directives so the toggle is observable.

#### Scenario: Checkbox defaults ON for a fresh install

- GIVEN a fresh install reaching the language section
- WHEN the artifacts-in-English checkbox is rendered
- THEN it defaults to ON (checked)
- AND the user may untick it to opt out

#### Scenario: ON keeps artifacts in English regardless of region

- GIVEN `artifactsInEnglish` is ON with region `mexico`
- WHEN persona content is composed
- THEN the artifact-language directive requires generated artifacts to default to English
- AND the reply-language directive still requests Mexican Spanish for chat replies

#### Scenario: OFF permits artifacts in the selected language

- GIVEN `artifactsInEnglish` is OFF with region `mexico`
- WHEN persona content is composed
- THEN the injected artifact-language directive permits generated artifacts in the selected reply language
- AND it differs from the directive injected when the checkbox is ON

---

### Requirement: R6 — Persistence of region and artifactsInEnglish

`region` (curated id or free-text string) and `artifactsInEnglish` (bool) MUST be persisted
in `.gentle-ai/state.json` on the same rail as the existing `persona` field. State merge
(`MergeAgents`) MUST carry both new fields. A round-trip marshal/unmarshal MUST preserve
both fields and their values.

#### Scenario: New state fields round-trip through marshal and unmarshal

- GIVEN an install state with `style: gentle`, `region: chile`, `artifactsInEnglish: false`
- WHEN the state is marshaled to JSON and unmarshaled back
- THEN `region` is `chile` and `artifactsInEnglish` is `false`
- AND no field is lost or reset to a zero-value

#### Scenario: State merge carries region and artifactsInEnglish

- GIVEN an existing state with persisted `region` and `artifactsInEnglish`
- WHEN agent state is merged
- THEN the merged state preserves both `region` and `artifactsInEnglish`

#### Scenario: Sync regenerates persona, region, and artifacts flag from persisted state

- GIVEN persisted state with style, region, and artifactsInEnglish
- WHEN sync builds its selection and regenerates persona content
- THEN the regenerated content reflects all three persisted values

---

### Requirement: R7 — Validation accepts new IDs and back-compat aliases

Persona normalization MUST accept the new style IDs (`gentle`, `neutral`, `custom`) and MUST
accept the legacy values (`gentleman`, `gentleman-neutral-artifacts`) as back-compat aliases
that normalize to their migration targets (per R1). Unknown values MUST be handled without
silently selecting a regional voice the user did not choose.

#### Scenario: New style IDs validate

- GIVEN a persona value of `gentle`, `neutral`, or `custom`
- WHEN normalization runs
- THEN the value is accepted as valid

#### Scenario: Legacy aliases normalize to migration targets

- GIVEN a persona value of `gentleman` or `gentleman-neutral-artifacts`
- WHEN normalization runs
- THEN it normalizes to `gentle` (with the migration tuple from R1)
- AND it is not rejected as invalid

#### Scenario: Unknown persona does not silently inject regional voice

- GIVEN an unknown persona value not covered by R1 aliases
- WHEN normalization/resolution runs
- THEN it resolves to a safe default that does not introduce an unselected regional voice

---

### Requirement: R8 — TUI region screen and review

The TUI MUST drop the `gentleman-neutral-artifacts` option, present the style choice as
`gentle | neutral | custom`, and present a language/region section for `gentle` and `neutral`
containing the curated region radios, the always-present "Idioma del usuario" option, the
always-present free-text "Otro…" input, and the artifacts-in-English checkbox. The region
section MUST be skipped for `custom`. The review screen MUST show the selected region and the
artifacts-in-English flag.

#### Scenario: Style options no longer include the hybrid

- GIVEN the persona style screen is rendered
- WHEN the options are listed
- THEN they are `gentle`, `neutral`, `custom`
- AND `gentleman-neutral-artifacts` is absent

#### Scenario: Region section shows curated radios plus always-present options

- GIVEN style `gentle` or `neutral` is selected
- WHEN the language/region section is rendered
- THEN it shows the five curated region radios with gentilicio-first Spanish labels
- AND it shows "Idioma del usuario"
- AND it shows a free-text "Otro… (texto libre)" input
- AND it shows the artifacts-in-English checkbox defaulting ON

#### Scenario: Region section is skipped for custom

- GIVEN style `custom` is selected
- WHEN routing proceeds
- THEN the language/region section is skipped
- AND no region or artifacts flag is collected

#### Scenario: Review shows region and artifacts flag

- GIVEN a finalized selection with a region and artifacts flag
- WHEN the review screen is rendered
- THEN it displays the selected region label
- AND it displays the artifacts-in-English state

---

### Requirement: R9 — Claude and generic asset coverage (this slice)

For this slice, the Claude and generic persona and output-style assets MUST be region-neutral
for the `gentle` style (no hardcoded Rioplatense `## Language` line) while preserving the full
teaching/mentor behavior contract. The composed directive MUST be the sole source of regional
voice. The neutral behavior contract MUST be preserved without regional voice. Stripping the
language section MUST NOT remove or weaken any non-language part of the persona contract.

#### Scenario: Gentle Claude/generic asset is region-neutral but keeps the teaching contract

- GIVEN the Claude or generic `gentle` persona asset is inspected
- WHEN its content is examined
- THEN it contains no hardcoded Rioplatense/voseo language line
- AND it preserves the concepts-first teaching, brevity, one-question, and verification-first contract

#### Scenario: Composed directive is the sole regional source for Claude/generic

- GIVEN Claude or generic content is composed for `gentle` + a region
- WHEN the rendered content is inspected
- THEN the only regional voice instruction is the composed directive line
- AND there is no second hardcoded regional instruction in the base asset or output-style asset

#### Scenario: Neutral Claude/generic asset keeps parity without regional voice

- GIVEN the Claude or generic `neutral` asset is rendered with a region directive
- WHEN its content is inspected
- THEN it preserves the neutral mentor contract (brevity, one-question, no-menu, verification-first, artifact-language)
- AND it contains no marked regional slang baked into the base asset

---

## Delta to existing spec: persona-behavior-contract

`openspec/specs/persona-behavior-contract/spec.md` currently ties regional voice to the
`gentleman` persona. Under the new model the style axis is `gentle | neutral | custom` and
regional voice is the language axis, not a persona variant. The following deltas MUST be
applied to that spec when this change lands.

| Existing element | Required delta |
|------------------|----------------|
| Requirement "Neutral Mentor Behavior Parity" | Replace "variant of the Gentleman mentor behavior contract" framing so neutral is parity of the `gentle` mentor contract; keep "no regional voice in the base asset". |
| Scenario "Gentleman keeps regional mentor behavior when explicitly selected" | Replace with: "Gentle persona with the Rioplatense region injects the Rioplatense (voseo) directive" — regional voice now comes from the language axis, not a `gentleman` persona. |
| Requirement "Artifact Language Independence" + scenario "Gentleman voice does not leak into artifacts" | Re-anchor to the `artifactsInEnglish` checkbox; replace `gentleman` references with `gentle`; artifact language is governed by the checkbox, not a persona variant. |
| Requirement "Claude Neutral Output Style Contract" scenario "Claude explicit Gentleman output-style remains honored" | Replace `gentleman` with `gentle`; the output-style governs tone/personality only and is region-neutral (regional voice is in the composed persona directive). |
| Requirement "Safe Persona Fallback Semantics" | Keep "do not silently reactivate regional Gentleman voice"; update naming to the `gentle` style + region axis; absent state resolves per R1 (fresh default `gentle` + Rioplatense) only for migration of a known prior gentleman selection, while truly unknown/invalid values stay safe-default without an unselected regional voice. |
| Requirement "Explicit Persona Selection Preservation" scenarios referencing `gentleman` | Replace `gentleman` with `gentle` + explicit region; explicit selections of style and region MUST remain authoritative across sync. |

These deltas are flagged here; the actual edit to the existing spec file is an implementation
task (sdd-tasks/apply), kept consistent with R1–R9 above.

---

## Testability notes (Strict TDD active — `go test ./...`)

Every requirement maps to unit-testable boundaries:

| Requirement | Test shape |
|-------------|------------|
| R1 migration | Table-driven test over the migration matrix; assert resolved tuple per old value; explicit assertion that `gentleman` and `gentleman-neutral-artifacts` converge byte-identical; explicit `artifactsInEnglish == true` (not zero-value) assertion. |
| R2 idempotency | Inject N times; assert byte-identical output and no duplicate markers/directives. Golden/round-trip: old state → migrate → inject vs. pre-change baseline. |
| R3 axes | Enumerate style options; assert region orthogonality; assert custom carries no region/injection. |
| R4 directive | Pure-function test on `personaContent()` per (style, region); assert single directive line; assert no per-region asset files; assert base asset has no hardcoded Rioplatense line. |
| R5 checkbox | Assert default ON; assert distinct directives for ON vs OFF. |
| R6 persistence | State JSON marshal/unmarshal round-trip with new fields; `MergeAgents` carry test. |
| R7 validation | Table-driven `normalizePersona()` over new IDs + legacy aliases + unknown. |
| R8 TUI | Direct `Model.Update()` state-transition tests; render assertions for region radios, options, checkbox, review; skip-region-for-custom routing test. |
| R9 assets | Asset-content guard tests: gentle base region-neutral, teaching contract survives, composed directive sole regional source. |

---

## Deferred follow-up (out of this spec's scope)

These adapters reuse the same strip-and-compose pattern and ship after the Motor slice:

- OpenCode / Kilocode region-neutral base + composed directive.
- Kimi Jinja module shape (compose directive into `persona.md` vs. dedicated `language.md`).
- Kiro steering-file asset coverage.
- Hermes persona/output-style asset coverage.

The migration logic (R1) and sync idempotency (R2) ship in THIS slice, so no existing user
degrades regardless of which adapters are covered yet.
