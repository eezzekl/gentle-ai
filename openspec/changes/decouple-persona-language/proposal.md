# Proposal: Decouple persona style from language/region

## Intent

Split today's single coupled persona axis (`gentleman | gentleman-neutral-artifacts | neutral | custom`) into TWO independent axes — personality **style** (`gentle | neutral | custom`) and **language/region** — and persist both across installs and syncs.

Today personality and regional voice are tangled in one selector, with Rioplatense voseo hardcoded inside each persona asset's `## Language` section. The `gentleman-neutral-artifacts` hybrid is the symptom: it exists only to keep the gentleman teaching voice while forcing English artifacts, but it injects content IDENTICAL to `gentleman` (verified: both land in the same `personaContent()` default case, and the gentleman asset already mandates English artifacts). A user who wants the gentle mentor style but a Mexican voice — or a neutral style with a Colombian voice — cannot express that. The axes are conceptually orthogonal but structurally fused.

After this change a user picks a personality style and, independently, a reply language/region. The regional voice is injected as ONE composed directive line built in Go from a curated region map plus free text — no per-region asset files, no runtime template engine.

## Locked decisions (chosen approach)

These were resolved in the design and product-decision rounds and are encoded here as the proposal's chosen direction. They are not reopened.

| # | Decision |
|---|----------|
| 1 | Drop `gentleman-neutral-artifacts` (verified zero behavioral delta). The style axis becomes `gentle | neutral | custom`. |
| 2 | Language/region is a SECOND axis: a TUI radio list of curated regions + an always-present "Idioma del usuario" option + an always-present free-text "Otro…" box. The injected language directive is ONE composed line appended in Go. No per-region `.md` assets. Curated regions live in a Go `map[RegionID]Label`. |
| 3 | Curated regions v1: Argentina, Mexico, Colombia, Spain, Chile. Labels are gentilicio-first with the regional variant in parens: "Argentino (rioplatense, voseo)", "Mexicano (tuteo)", "Colombiano (paisa)", "Español (España)", "Chileno". Sub-national dialects (e.g. yucateco) go through free text, NOT curated radios. |
| 4 | `artifactsInEnglish` becomes a CHECKBOX in the language section (not a persona variant). Default ON (opt-out) — matches current behavior. It toggles which artifact-language directive is injected. |
| 5 | Default region for a fresh `gentle` install = Rioplatense (voseo) → preserves today's behavior exactly. |
| 6 | `neutral` style ALSO shows the region selector (axes are orthogonal). `custom` keeps "no injection / follow user" → no region selector for custom. |
| 7 | Persist `region` (id or free-text string) + `artifactsInEnglish` (bool) in `.gentle-ai/state.json` on the same mechanism as the existing `persona` field; sync regenerates via `InjectForSync`; `MergeAgents` must carry both new fields. |
| 8 | Back-compat: existing `"persona": "gentleman-neutral-artifacts"` and `"persona": "gentleman"` in state.json MUST migrate to `persona=gentle + artifactsInEnglish=true` (explicit + tested) so users do not silently degrade to neutral. |

## Goals

- Make personality style and reply language/region independently selectable at install time and persisted across syncs.
- Remove the `gentleman-neutral-artifacts` hybrid and the dead `default`-case ambiguity it relied on.
- Inject regional voice as a single composed directive line built from a Go region map + free text, so adding a region is a one-line map entry, not an N×M asset matrix.
- Preserve today's behavior exactly for existing users (gentle + Rioplatense + English artifacts).
- Keep all generated artifacts (specs, code, comments, UI identifiers) English by default; only the TUI region LABELS shown to the user stay in Spanish.

## Non-goals

- Per-region curated multi-line voice packs or per-region asset files. The LLM produces regional Spanish from a named variant; curated packs are over-engineering.
- A runtime template engine. Injection stays Go build-time composition (only Kimi keeps its existing Jinja includes).
- Changing the gentle teaching contract (concepts-first, voseo for Rioplatense, brevity, one-question-at-a-time) or the neutral behavior contract beyond decoupling the region.
- Adding new curated regions beyond v1 (AR/MX/CO/ES/CL); anything else goes through free text.
- Validating free-text region quality. Free text is injected as-is into the directive.

## Scope

### In scope

The full verified touchpoint set from the exploration:

| Area | Change |
|------|--------|
| `internal/model/types.go` | Add `PersonaGentle`; drop `PersonaGentlemanNeutralArtifacts`; add `RegionID` type + curated `map[RegionID]Label`; add region + artifacts-in-English to the selection model. |
| `internal/model/selection.go` | Carry `Region` (id or free string) and `ArtifactsInEnglish` in `Selection`. |
| `internal/state/state.go` | Add `Region string` + `ArtifactsInEnglish bool` to `InstallState`; `MergeAgents` must carry both new fields. |
| `internal/cli/validate.go` | `normalizePersona()` accepts new IDs + back-compat aliases (`gentleman`, `gentleman-neutral-artifacts`). |
| `internal/cli/sync.go` | `applyResolvedPersona()` migrates old persona values to `gentle + artifactsInEnglish=true`; `BuildSyncSelection` loads the new state fields. |
| `internal/components/persona/inject.go` | Compose the one-line language directive into persona content; simplify `isGentlemanConversationPersona()` to gentle; route artifacts-in-English to the right directive; output-style dispatch handles the renamed/new persona; Kimi module writes the composed line. |
| Persona assets (`claude`, `generic`, `opencode`, `kimi`, `kiro`, `hermes`) | Make the `gentle` base region-neutral (strip the hardcoded `## Language` Rioplatense line) so the composed directive is the single source of regional voice. |
| Output-style assets (Claude gentle/neutral, Kimi twins) | Strip the hardcoded Language Rules section so they govern tone/personality only (the composed directive lives in persona content). |
| TUI region screen + checkbox (`internal/tui/screens/persona.go`, `model.go`, `router.go`, `review.go`) | Drop `gentleman-neutral-artifacts`; add region radio list + "Idioma del usuario" + "Otro…" free text + artifacts-in-English checkbox; route gentle/neutral → region screen, skip it for custom; review shows region + artifacts flag. |
| `openspec/specs/persona-behavior-contract/spec.md` | Update wording: persona axis is `gentle | neutral | custom`; regional voice is the language axis, not a persona variant; artifact-language is the checkbox. |
| `internal/components/uninstall/service.go` | Remove the output-style file under its new name (`gentle.md` if renamed); no per-region files exist to remove. |
| Tests | Update/replace persona-language contract tests, preset tests, inject tests, state round-trip, and add explicit migration tests. |

### Out of scope

- Adding curated regions beyond v1.
- Localizing artifacts or UI beyond the existing Spanish region labels.
- Any production code in this proposal phase (planning only).

## Approach

Treat style and language/region as two orthogonal selections that converge at injection time.

1. **Model**: `gentle | neutral | custom` for style; `RegionID` (curated map) or free-text string for region; `ArtifactsInEnglish bool` (default true) for artifact language. Custom skips region entirely.
2. **Compose, don't multiply**: `personaContent()` returns a region-neutral persona body and appends ONE composed directive built from the region (e.g. Argentina → `Reply to the user in warm, natural Argentine Spanish (Rioplatense, voseo).`). Free text and "Idioma del usuario" feed the same template (free text raw; "Idioma del usuario" → follow how the user writes, no forced region). The artifacts-in-English checkbox toggles which artifact-language directive is injected.
3. **One code path for all agents**: build-time composition for non-Kimi agents (append to the persona string before write); the existing Kimi Jinja module writes the same composed line. No runtime templating added.
4. **Persist on the proven rail**: `region` + `artifactsInEnglish` live in `state.json` next to `persona`; `MergeAgents` carries them; sync regenerates via `InjectForSync`.
5. **Migrate explicitly**: old `gentleman` and `gentleman-neutral-artifacts` persona values map to `gentle + artifactsInEnglish=true` with a tested migration case, so no existing user silently degrades to neutral.

## Recommended first slice

Land the change with **Claude + generic agent coverage first**, then stage the remaining agents.

**First slice (one reviewable PR):**
- Model + state + validate + sync (migration) + `inject.go` composition.
- Region-neutral persona + output-style assets for **Claude and generic** only.
- TUI region screen + artifacts checkbox + review.
- `persona-behavior-contract` spec update.
- Migration tests, state round-trip tests, inject tests for Claude/generic, TUI render tests.

**Deferred slice(s):**
- OpenCode/Kilocode, Kimi (Jinja module), Kiro, and Hermes asset coverage for the region-neutral base + composed directive.

Rationale: the model/state/inject/TUI core is the load-bearing, highest-risk surface and is agent-independent. Claude+generic proves the composition end-to-end. The remaining agents are mechanical asset edits that repeat the same strip-and-compose pattern; staging them keeps the first PR within review budget and isolates the Kimi-specific Jinja path. The migration logic ships in slice 1 so no user degrades regardless of which agents are covered.

## Open design decisions (deferred to design phase)

These are resolved enough to propose but their exact mechanics belong to design:

1. **Language directive placement in the output-style layer**: persona content block only (recommended single source of truth), or also reflected in the Claude gentle/neutral + Kimi output-style assets. Recommendation: persona content only; output-style governs tone/personality.
2. **Strip vs. rename for assets**: prefer stripping the hardcoded `## Language` section and composing at build time over renaming files. The rename-vs-strip choice and the output-style filename migration (`gentleman.md` → `gentle.md` vs. keep `gentleman.md`) are design decisions, with the uninstall/cleanup migration risk called out.
3. **`artifactsInEnglish` meaningfulness**: since it defaults ON for all non-custom personas and current neutral already mandates English artifacts, confirm in design whether the checkbox is surfaced as a true opt-out and which directive each state injects.
4. **Kimi language module shape**: compose the line into the existing `persona.md` module vs. a dedicated `language.md` module included by `KIMI.md`.

## Affected internal packages

| Area | Impact |
|------|--------|
| `internal/model` | Persona IDs, `RegionID` + curated map, selection fields. |
| `internal/state` | New `Region` + `ArtifactsInEnglish` fields; `MergeAgents` carry. |
| `internal/cli` | `validate.go` normalization + aliases; `sync.go` migration + sync selection. |
| `internal/components/persona` | Injection composition, gentle simplification, output-style dispatch, Kimi module. |
| `internal/components/uninstall` | Output-style filename removal under new name. |
| `internal/assets/{claude,generic,opencode,kimi,kiro,hermes}` | Region-neutral persona + output-style assets. |
| `internal/tui` | Region screen, checkbox, routing, review. |
| `openspec/specs/persona-behavior-contract` | Spec wording for the decoupled axes. |

## Risks and mitigations

| Risk | Mitigation |
|------|------------|
| Migration gap: old `gentleman-neutral-artifacts`/`gentleman` state silently degrades to neutral, losing the teaching voice. | Explicit, tested migration case mapping both to `gentle + artifactsInEnglish=true`, shipped in the first slice. |
| `personaContent()` signature change ripples to all callers/tests. | Land the signature change with Claude/generic coverage first; treat remaining agents as mechanical follow-ups against the stabilized signature. |
| Output-style filename migration (`gentleman.md` → `gentle.md`) leaves orphaned files on uninstall/sync. | Keep rename-vs-strip a design decision; if renamed, update uninstall to remove both old and new names; cover with sync/uninstall tests. |
| `artifactsInEnglish` is effectively always-on, making the checkbox feel inert. | Frame it as an explicit opt-out in design; verify each state injects a distinct artifact-language directive. |
| Stripping `## Language` from assets accidentally removes more than the regional line, weakening the persona. | Strip only the regional language section; add asset guards/tests asserting the composed directive is the sole regional source and the teaching contract survives. |
| Region label drift for ultra-niche dialects (e.g. yucateco). | Out of scope by design; free-text box covers it; no curated radio promised. |
| Changed-line count exceeds the practical ~400-line review budget. | First-slice boundary (Claude+generic) plus staged agent follow-ups; request/record `size:exception` only if a single slice still exceeds budget. |

## Rollback plan

This is prompt/config/state behavior, so rollback is largely file-level:

1. Revert the model/state/inject commit(s) and the asset strip/compose commit(s) from the same work unit.
2. Restore the previous `applyResolvedPersona` fallback and `normalizePersona` aliases.
3. Restore the original persona + output-style assets (with the hardcoded `## Language` sections).
4. State migration is additive (new optional fields); on rollback, the new `region`/`artifactsInEnglish` fields are ignored by old code — no destructive data migration. If a filename rename shipped, restore the original output-style filename and uninstall list.
5. Re-run install/sync tests after rollback to confirm generated prompts match the restored baseline.

## Success criteria

- [ ] Personality style (`gentle | neutral | custom`) and language/region are selected independently in the TUI and persisted in `state.json`.
- [ ] `gentleman-neutral-artifacts` is removed; old state values migrate to `gentle + artifactsInEnglish=true` (tested).
- [ ] Fresh `gentle` install defaults to Rioplatense (voseo) and reproduces today's behavior exactly.
- [ ] `neutral` shows the region selector; `custom` does not.
- [ ] Curated regions v1 (AR/MX/CO/ES/CL) render with the gentilicio-first labels; "Idioma del usuario" and free-text "Otro…" are always present.
- [ ] The regional voice is injected as one composed directive line built in Go; no per-region asset files exist.
- [ ] `artifactsInEnglish` is a checkbox defaulting ON; unticking it injects the in-language artifact directive.
- [ ] Sync regenerates persona + region + artifacts flag via `InjectForSync`; `MergeAgents` carries the new fields.
- [ ] `persona-behavior-contract` spec reflects the decoupled axes.
- [ ] `go test ./...` and `go vet ./...` pass (strict TDD: tests first).
