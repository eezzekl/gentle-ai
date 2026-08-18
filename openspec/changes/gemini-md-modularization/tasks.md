# Tasks: Modularize GEMINI.md under the 12k Antigravity budget

> **Post-merge note (main merged 2026-07-28, d53e9578)**: all new code/tests import `github.com/gentleman-programming/gentle-ai/v2/...`. Orchestrator assets were fully rewritten upstream (antigravity 35,508 B, gemini 32,841 B) — regenerate goldens against that content. Trigger-rules no longer render into the root (`triggerrules.go` deleted; `agentguidance` owns routing outside `GEMINI.md`). Do NOT add fields to `capabilitymanifest.AgentFeatureClaims` — see design "Post-Merge Constraints".

## Review Workload Forecast

| Field | Value |
|---|---|
| Estimated authored changed lines | ~525 (goldens excluded) |
| 400-line budget risk | Medium (per-PR ≤400; total ~525 across 4 PRs) |
| Chained PRs recommended | Yes |
| Delivery strategy | force-chained |
| Chain strategy | feature-branch-chain (user-selected 2026-07-28) |

Decision needed before apply: No — resolved
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: Medium

Sequential dependency fits **feature-branch-chain** (tracker + child branches); user selected it over stacked-to-main on 2026-07-28. The chain as built is fully linear — PR1 → PR2 → PR3 → PR4, each branch parented on the previous one. The `PR1 → {PR2,PR3} → PR4` fan-out written here originally was planned but never materialized in git.

### Suggested Work Units

| Unit | Goal | PR | Focused test | Runtime harness | Rollback boundary |
|---|---|---|---|---|---|
| WU-1 | `SharedReferenceLayout` capability on gemini/antigravity adapters | PR1 | `go test ./internal/agents/...` | N/A — pure adapter method, no I/O | Revert 2 adapter files + tests |
| WU-2 | `renderSessionBootstrapStub()` + wording test | PR1 | `go test ./internal/components/engram/...` | N/A — pure render func | Revert added func + test |
| WU-4 | `renderSDDGateStub()` + wording test | PR1 | `go test ./internal/components/sdd/...` | N/A — pure render func | Revert added func + test |
| WU-8 | Extend antigravity collision warning to `~/.gemini/references/` | PR1 | `go test ./internal/cli/... -run Antigravity` | `agy install` w/ both agents selected → warning text | Revert warning string edit |
| WU-3 | Wire engram default-case to write reference file + bootstrap stub | PR2 | `go test ./internal/components/engram/... ./internal/components/...` | `agy sync` (gemini+antigravity) → inspect `~/.gemini/references/engram-protocol.md` | Revert default-case branch; engram falls back to inline |
| WU-5 | `SelectedAgents` + antigravity-priority selection (D4) | PR3 | `go test ./internal/components/sdd/...` | N/A — pure selection logic | Revert field + selection func |
| WU-6 | Wire `injectFileAppend` to write reference file + gate stub | PR3 | `go test ./internal/components/sdd/... ./internal/components/... ./internal/cli/...` | `agy sync` (gemini+antigravity) → inspect `~/.gemini/references/sdd-orchestrator.md` | Revert wiring; sdd falls back to inline |
| WU-7 | Byte-budget WARNING at sync/install completion (D7) | PR4 | `go test ./internal/cli/... -run Budget` | `agy sync` w/ oversized root → warning printed, exit 0 | Revert warning block in sync.go/run.go |
| WU-9 | Migration/idempotency/convergence/budget integration tests | PR4 | `go test ./internal/components/...` | `agy sync` twice over seeded legacy fixture → byte-identical 2nd pass | Revert new integration test file |

## Phase 1: Foundation (PR1)

- [x] 1.1 RED: `internal/agents/gemini/adapter_test.go`, `internal/agents/antigravity/adapter_test.go` — assert `ReferencesDir(home) == filepath.Join(home, ".gemini", "references")`
- [x] 1.2 GREEN: add `ReferencesDir` to `internal/agents/gemini/adapter.go`, `internal/agents/antigravity/adapter.go` (spec: gemini-prompt-modularization Reference File Generation)
- [x] 1.3 RED: engram test — `renderSessionBootstrapStub()` contains file-read imperative, absolute `~/.gemini/references/engram-protocol.md`, "NOT the workspace", "active session instructions"; no test-marker text (spec: Session-Bootstrap Block Wording)
- [x] 1.4 GREEN: implement `renderSessionBootstrapStub()` + local `SharedReferenceLayout` interface in `internal/components/engram/inject.go`
- [x] 1.5 RED: sdd test — `renderSDDGateStub()` lists SDD commands, imperatively instructs reading `~/.gemini/references/sdd-orchestrator.md` first, "before handling any SDD command or request", no orchestration detail (spec: SDD-Stub Block Wording)
- [x] 1.6 GREEN: implement `renderSDDGateStub()` + local `SharedReferenceLayout` interface in `internal/components/sdd/inject.go`
- [x] 1.7 RED: `internal/cli/antigravity_collision_test.go` (tasks.md said `run_test.go`, drifted — real location) — `antigravityCollisionCheck` message includes `~/.gemini/references/` when both agents selected (spec: antigravity-support MODIFIED)
- [x] 1.8 GREEN: extend warning string in `antigravityCollisionCheck`, `internal/cli/run.go:2183` (appended to checks at :2104); required an adjacent `// refusal:by-design human-authority` marker to satisfy `TestEveryProductionRefusalNamesResolutionOrDeclaresByDesign` after the message text changed

## Phase 2: Engram Reference Wiring (PR2)

- [x] 2.1 RED: `internal/components/engram/reference_file_test.go` (in-package, so `protocolFull()` verbatim equality is directly assertable) + golden tests in `internal/components/golden_test.go` — `engram.Inject` for gemini and antigravity write `~/.gemini/references/engram-protocol.md` (verbatim `protocolFull()`, no section markers) and inject the bootstrap stub under `gentle-ai:engram-protocol` in root; idempotency covered here too (spec: engram-protocol-injection Reference-File Emission)
- [x] 2.2 GREEN: in `internal/components/engram/inject.go` default case, type-assert `SharedReferenceLayout`; when present write reference file + stub instead of inline body
- [x] 2.3 [generated] regen `testdata/golden/engram-antigravity-rulesmd.golden` (6760 → 386 B); add `testdata/golden/engram-gemini-rulesmd.golden`, `testdata/golden/engram-gemini-referencefile.golden`, `testdata/golden/engram-antigravity-referencefile.golden` — gemini/antigravity pairs are byte-identical, confirming D3 convergence
- [x] 2.4 (unplanned, required) `internal/components/delivery_guarantee_installed_test.go` — the issue-#1042 antigravity subtest asserted the delivery-guarantee wording on `GEMINI.md`; that body now lives in the reference file. Moved the assertions to `~/.gemini/references/engram-protocol.md` and added a root-points-at-reference assertion so the guarantee still has an unbroken path to the model.

## Phase 3: SDD Orchestrator Reference Wiring (PR3)

- [x] 3.1 RED: `internal/components/sdd/reference_file_test.go` — `referenceOrchestratorAgent`: `[gemini]` → gemini asset; `[gemini,antigravity]`, `[antigravity]`, and antigravity anywhere in the list → antigravity asset; empty selection → injected adapter's own asset (spec: sdd-orchestrator-assets Per-Agent Content Selection, D4)
- [x] 3.2 GREEN: add `SelectedAgents []model.AgentID` to `InjectOptions`; add `referenceOrchestratorAgent` + `sddReferenceFileName`, `userHomeDir`/`SetUserHomeDirForTest`, `homeAnchoredReferencePath` in `internal/components/sdd/inject.go` (same home-anchoring idiom PR2 introduced in engram)
- [x] 3.3 GREEN: populate `SelectedAgents` at both `sdd.InjectOptions` build sites — `internal/cli/sync.go` (ComponentSDD case) and `internal/cli/run.go` (ComponentSDD case); required a new `selectedAgentIDs(adapters)` helper in `internal/cli/run.go` (no adapters→IDs mapper existed). Guarded by an unplanned but required end-to-end wiring test, `internal/cli/sdd_reference_selection_test.go`: without the population the unit-level D4 selection is dead code, since every real caller passes through these two sites.
- [x] 3.4 RED: golden tests in `internal/components/golden_test.go` (`TestGoldenSDD_Gemini`, `TestGoldenSDD_Antigravity`) + behavior tests in `reference_file_test.go` — reference file holds the selected asset verbatim with no section markers, root carries only the gate stub under `gentle-ai:sdd-orchestrator`; workspace-scoped inline fallback, idempotency, and marker-level legacy migration covered too (spec: Orchestrator Reference File Emission)
- [x] 3.5 GREEN: rewire `injectFileAppend` to type-assert `SharedReferenceLayout` and branch to reference-file + stub path, taken ONLY when the written path equals the home-anchored path the stub advertises; `InjectionResult.Files` now reports the reference file as well
- [x] 3.6 [generated] regen `testdata/golden/sdd-gemini-geminimd.golden` and `testdata/golden/sdd-antigravity-rulesmd.golden` (both ~33KB → 349 B, and now byte-identical to each other, confirming D3 root convergence); add `testdata/golden/sdd-gemini-referencefile.golden` (40,110 B) and `testdata/golden/sdd-antigravity-referencefile.golden` (42,777 B) — these two differ by design (D4: the assets are not reconcilable)

## Phase 4: Budget Warning, Migration & Convergence (PR4)

- [x] 4.1 RED: `internal/cli/root_budget_test.go` — over budget the warning names the actual count, the path, the 12,000 limit and the truncation consequence; under budget, missing root, and non-shared-root agents stay silent; `RunSyncWithSelection` over a 60,000-character root warns and does NOT fail (spec: Byte-Budget Warning)
- [x] 4.2 GREEN: `internal/cli/root_budget.go` — `rootBudgetWarnings` measures the assembled `SystemPromptFile` for adapters on the shared reference layout, deduplicated so gemini+antigravity warn once. Computed in `RunSyncWithSelection` and `RunInstall` right after assembly (deliberately before the verification early-return, so an oversized root is still reported on a run that fails for another reason) and surfaced through `RootBudgetWarnings` on both results. `RenderSyncReport` routes every branch — including the no-op one — through `renderWithRootBudgetWarnings`, because an idempotent re-sync is exactly when an already-oversized root would otherwise stay hidden forever
- [x] 4.3 RED: `internal/components/reference_layout_integration_test.go` — persona+sdd+engram in pipeline order for both adapters → root under 12,000 bytes, persona content inline, neither large body inline, both reference files present. NOTE: the requirement says persona *content* must be inline, not that a `gentle-ai:persona` marker must exist; the assertion compares against the longest real rendered persona line from a clean persona-only injection rather than a hand-copied phrase that would rot
- [x] 4.4 RED: migration test — ~54KB legacy inline fixture (both bodies under their markers, plus a foreign block), sync → converts to stub+references, root under budget, persona-first order preserved, foreign block byte-identical, each managed marker exactly once. The fixture is synthetic on purpose so it stays a fixed legacy shape when the shipped assets are rewritten upstream
- [x] 4.5 RED: idempotency test — second full assembly pass leaves root and both reference files byte-identical
- [x] 4.6 RED: convergence test — gemini-then-antigravity vs antigravity-then-gemini → identical root + reference bytes. **This test failed and exposed a real pre-existing defect in the persona component; see 4.10 and design.md D8.** It passes now
- [x] 4.7 RED: uninstall test in `internal/components/uninstall/service_test.go` — after inject + uninstall of ComponentSDD and ComponentEngram, no `sdd-orchestrator`/`engram-protocol` marker and no live `~/.gemini/references/` pointer survives in the root, while both reference files REMAIN on disk. Per the spec requirement *Uninstall/Revert Safety* the orphans are kept inert, not deleted (user-confirmed 2026-07-29, superseding the deletion patch proposed by review finding R3-uninstall-orphans-reference-file). The test accepts a removed root as well: uninstall legitimately drops the file once the stubs were its only managed content
- [x] 4.8 GREEN: all Phase 4 assertions pass; `go test ./...` green with a cleared test cache
- [x] 4.9 fixed the stale comment in the `sdd.Inject` FileReplace/AppendToFile case — it still claimed the orchestrator "is included in the generic persona asset", which stopped being true upstream and is doubly wrong now that the body lives in a reference file
- [x] 4.10 (unplanned, required by 4.6) GREEN: `internal/components/persona/inject.go` — every adapter whose prompt file is shared with another component now writes one marker-bound `gentle-ai:persona` section and preserves managed sections, for all persona variants. Previously the Gentleman variants bypassed `preserveManagedSections` entirely and wrote the raw asset over the whole prompt file, which (a) made D3 convergence false by block order and (b) let a persona-only sync silently wipe both stubs — the only pointer to the reference files. The OpenCode special case that already had the correct shape is gone; the path is uniform. Guarded by `internal/components/persona/shared_root_preservation_test.go` (spec delta: persona-behavior-contract — Marker-Bound Persona Section in Shared Prompt Files, Shared-Root Convergence Across Adapter Order)
- [x] 4.11 (scope extension, user-requested) GREEN: extended 4.10 to `StrategyInstructionsFile` (VS Code) and `StrategySteeringFile` (Kiro), which had the identical wipe. `wrapInstructionsFile`/`wrapSteeringFile` are replaced by frontmatter constants plus `ensureFrontmatter`, which seeds the YAML header only when the file lacks one and never rewrites an existing header — a second header renders as a literal `---` divider inside the prompt, not as metadata. Persona adopted sdd's existing `description` wording so the two components' header constants are byte-identical and the shipped `sdd-vscode-instructions.golden` did not change. `TestWrapSteeringFileAddsKiroFrontmatter` was replaced by `TestEnsureFrontmatterAddsKiroHeaderWithoutDuplicating` (same guarantee, plus the seeding and duplication cases). Golden churn: `persona-kiro-gentleman.golden` only, which gained the two persona markers. `StrategyAppendToFile` (antigravity, windsurf, pi) already had the correct shape and is untouched (spec delta: persona-behavior-contract — Single Frontmatter Header in Wrapped Prompt Files)

- [x] 4.12 (D9, added after review feedback on issue #1989) RED: `TestManagedReferencePathMatchesWhatInjectWrites` in both `internal/components/engram/reference_file_test.go` and `internal/components/sdd/reference_file_test.go`, plus `internal/cli/reference_backup_parity_test.go` — the two reference files must appear in `componentPathsWithWorkspaceScoped` and `syncBackupTargets` under global scope, and must NOT appear under workspace scope. GREEN: each component exports `ManagedReferencePath(adapter any, injectionDir string) string`, which returns the path only where `Inject` actually writes one; both `inject.go` files now derive their own write branch from it, so writer and path list cannot drift. `internal/cli/run.go` calls it in the `ComponentEngram` and `ComponentSDD` arms. **D9 as originally written prescribed an unconditional append**, which would have listed two files that no workspace-scoped install ever writes and turned a backup gap into a post-apply/post-sync verification failure on every workspace install — the home-anchoring guard is not optional, and `TestInjectKeepsProtocolInlineWhenTargetIsNotHome`/`TestInjectKeepsOrchestratorInlineWhenTargetIsNotHome` (both pre-existing) are what proved it. Design updated to match

### Notes and deliberate non-goals

- Byte-equality of a wrapped prompt file across arbitrary *component* orders is deliberately NOT asserted. `internal/cli/sync.go:329-331` fixes persona before SDD and that order is load-bearing, so "SDD first" is a state the product never produces. `TestSharedPromptFileConvergesAcrossComponentOrder` therefore asserts the header is single, top-of-file and identical across components, and that both sections survive — not full byte-equality. Block order across *agents* sharing one root IS asserted (4.6), because both orders are real there.
- Cosmetic follow-up (not a defect of this change): the shared VS Code header says `description: Gentleman persona with SDD orchestration and Engram protocol`, which is inaccurate when a neutral persona is installed. Left as the shipped wording to avoid churning `sdd-vscode-instructions.golden` inside this change.

## Completion Criteria (spec traceability)

Root <12k both adapters (1.x/2.x/3.x) · stub wording invariants (1.3-1.6) · reference files verbatim, no test-marker blocks (2.1-2.3, 3.4-3.6) · migration/idempotency/convergence (4.4-4.6) · byte-budget warning never fails (4.1-4.2) · uninstall leaves no active references (4.7) · reference files backed up and verified only where written (4.12) · antigravity shared-references warning (1.7-1.8) · `go test ./...` green after every phase.
