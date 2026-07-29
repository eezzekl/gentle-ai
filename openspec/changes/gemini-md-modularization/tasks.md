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

Sequential dependency (PR1 → {PR2,PR3} → PR4) fits **feature-branch-chain** (tracker + child branches); user selected it over stacked-to-main on 2026-07-28.

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

- [ ] 4.1 RED: assembled root >12000 chars → warning names actual count + 12,000 limit, run does not fail; <12000 → no warning (spec: Byte-Budget Warning)
- [ ] 4.2 GREEN: add check at `RenderSyncReport` (`internal/cli/sync.go:~1570`) or its call site, and at `RenderInstallManualActions` (`internal/cli/run.go:1546`) for the install path
- [ ] 4.3 RED: integration test — persona+sdd+engram inject for gemini and antigravity → assembled `GEMINI.md` `len(bytes) < 12000` (spec: Root Assembly Byte Budget, Golden and Byte-Budget Test Coverage)
- [ ] 4.4 RED: migration test — seed ~54KB legacy inline `GEMINI.md` fixture, `sync` → converts to stub+references, persona-first order preserved, unrelated agent content untouched (spec: Migration and Idempotency — legacy install migrates)
- [ ] 4.5 RED: idempotency test — second `sync` over stub+references layout → byte-identical output (spec: Second sync is a no-op)
- [ ] 4.6 RED: convergence test — sync gemini-then-antigravity vs antigravity-then-gemini → identical root + reference bytes (design D3)
- [ ] 4.7 RED: uninstall test — `session-bootstrap`/`sdd-stub` blocks + markers removed; orphaned reference files never read (spec: Uninstall/Revert Safety)
- [ ] 4.8 GREEN: implement/verify all Phase 4 assertions pass; add tests to `internal/components/golden_test.go` or new `internal/components/integration_test.go`
- [ ] 4.9 (optional, low-priority) fix stale comment at `internal/components/sdd/inject.go:260-262`

## Completion Criteria (spec traceability)

Root <12k both adapters (1.x/2.x/3.x) · stub wording invariants (1.3-1.6) · reference files verbatim, no test-marker blocks (2.1-2.3, 3.4-3.6) · migration/idempotency/convergence (4.4-4.6) · byte-budget warning never fails (4.1-4.2) · uninstall leaves no active references (4.7) · antigravity shared-references warning (1.7-1.8) · `go test ./...` green after every phase.
