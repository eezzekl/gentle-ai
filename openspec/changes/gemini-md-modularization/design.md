# Design: Modularize GEMINI.md under the 12k Antigravity budget

## Post-Merge Constraints (main merged 2026-07-28, d53e9578)

- **Module path is `/v2`**: all new code and tests import `github.com/gentleman-programming/gentle-ai/v2/...`.
- **Capability decision — do NOT touch `capabilitymanifest`**: adapters now derive `Supports*` booleans from `internal/agents/capabilitymanifest` (canonical, digested; `CanonicalJSON`/`Digest` feed RAR authority in `internal/reviewtransaction/`). Adding a field to `AgentFeatureClaims` changes the canonical JSON and digest of every agent. Path/layout methods (`OutputStyleDir`, `CommandsDir`, `EmbeddedSubAgentsDir`) remain plain adapter methods, so `ReferencesDir` + the local `SharedReferenceLayout` interface stays in that family. Promote to a manifest claim only if a third adapter adopts shared references.
- **Trigger-rules left the root**: `internal/components/sdd/triggerrules.go` was deleted; routing guidance now lives in `internal/components/agentguidance/` under the `agent-routing` marker, outside `GEMINI.md`, and `stripLegacyTriggerRules` (`internal/cli/run.go:735`) strips legacy blocks. The root this change assembles is persona + sdd stub + engram stub only.
- **Orchestrator assets fully rewritten** (delegation topology + Native Checking Contract): antigravity 392 lines / 35,508 B, gemini 361 lines / 32,841 B. All goldens regenerate against this post-merge content.

## Technical Approach

Keep both adapters' existing strategies untouched. Add an optional adapter capability interface `SharedReferenceLayout { ReferencesDir(homeDir string) string }` implemented by the gemini and antigravity adapters (returns `~/.gemini/references`). The Engram and SDD components type-assert it (precedent: `bootstrapper`, `codexModelResolver` in `sdd/inject.go`); when present they write full bodies to `~/.gemini/references/` and inject short imperative stubs under the EXISTING marker IDs. Persona stays inline, unchanged. Migration is unconditional convergent replacement — no detection heuristic exists or is needed.

## Architecture Decisions

### D1 (Q1): Capability interface, not a new SystemPromptStrategy

**Choice**: optional adapter interface; no new strategy constant.
**Alternatives**: `StrategyRootPlusReferences` (Kimi-style).
**Rationale**: gemini=`StrategyFileReplace`, antigravity=`StrategyAppendToFile` is a load-bearing distinction in `persona/inject.go` (117-181 vs append path) — one new constant collapses it; two doubles churn. Strategy switches use grouped cases (`sdd/inject.go:262`) and `default:` fall-throughs (`engram/inject.go:529`); a new constant risks silent missing-case regressions at every switch. Codebase convention for capability variance is optional interfaces + agent switches.

### D2 (Q4 + marker IDs): Reuse existing marker IDs; migration needs no detection

**Choice**: stub content written under the SAME IDs `gentle-ai:engram-protocol` and `gentle-ai:sdd-orchestrator` — NOT new `session-bootstrap`/`sdd-stub` IDs.
**Rationale**: `InjectMarkdownSection` replaces marker content unconditionally → legacy→stub converts on first sync; stub→stub is byte-identical (`WriteFileAtomic` Changed=false); fresh install appends stub. Oscillation is impossible: the new binary has no full-body write path for these agents. Fingerprint/size/version heuristics rejected as unnecessary state.
New IDs were rejected because they require paired remove+add and break the proposal's rollback plan: a reverted binary rewrites full bodies under the same IDs; with new IDs, stub blocks would persist as orphans beside restored bodies.
**Collision check**: no new IDs introduced; persona legacy-strip fingerprints ("## Personality", "Senior Architect") do not appear in stub wording; `hasLegacyBareOrchestrator` strip (`sdd/inject.go:2107`) still runs first.
**Flag**: deviates from the prototype's block NAMES only; validated content is preserved verbatim — validation was content-based, marker comments are inert.

### D3 (Q2): Shared-file idempotency via byte-invariance

Every write to shared paths is agent-invariant or deterministically selected:
- Root `GEMINI.md`: persona (same generic asset), sdd stub (agent-neutral bytes), engram stub (agent-neutral). Both agents write identical bytes → "last writer wins" degenerates to no-op; order-independent.
- `references/engram-protocol.md`: `protocolFull()` for both (slim is Claude-only, `protocol.go:43-50`) → identical bytes.
- `references/sdd-orchestrator.md`: deterministic selection (D4).
Sync twice, any agent order → byte-identical output.

### D4 (Q3): Single shared orchestrator reference, antigravity-priority selection

**Choice**: one `~/.gemini/references/sdd-orchestrator.md`; content = antigravity asset when antigravity ∈ selected agents, else gemini asset. New `SelectedAgents []model.AgentID` on `sdd.InjectOptions`, populated at the `sdd.InjectOptions` build sites — `sync.go:944` (ComponentSDD case) and `run.go:1374` (both loops already hold the adapter set).
**Evidence**: assets are NOT reconcilable — re-verified post-merge on the rewritten assets: antigravity (392 lines, 35,508 B) mandates Mission Control `define_subagent`/`invoke_subagent` dynamic delegation; gemini (361 lines, 32,841 B) is the generic coordinator without those tools. (The 431-line/37,439 B figures were the pre-merge prototype validation content.)
**Alternatives**: per-agent filenames — rejected: the stub lives under the shared `sdd-orchestrator` marker, so per-agent stub text reintroduces last-writer mismatch (agy could load the gemini stub), needs unvalidated conditional wording, and conflicts with proposal scope.
**Residual risk**: syncing gemini alone on a machine where antigravity is installed but unselected writes gemini content (same divergence exists today inline); gemini-cli obeying antigravity content references tools it lacks — accepted per locked antigravity priority.

### D5: Stub content generated in Go

`renderSessionBootstrapStub()` (engram pkg) and `renderSDDGateStub()` (sdd pkg), like `RenderTriggerRules` — wording invariants unit-testable, no embed plumbing. No test-marker blocks in shipped output.

Both stubs MUST contain an imperative verb, "with your file-read tool", and the absolute `~/.gemini/references/...` path. Beyond that the two stubs carry different wording, because they gate different things:

| | `session-bootstrap` (engram) | `sdd-stub` (sdd) |
|---|---|---|
| When it must fire | before the model's first reply | before any SDD command or its natural-language equivalent |
| Lists SDD commands | no | yes |
| "in your home directory, NOT the workspace" | required | not required |
| "treat as active session instructions" | required | not required |

The bootstrap stub carries the extra two clauses because it fires unprompted at session start, where the model has no task context to disambiguate a `~`-anchored path from a workspace-relative one, and where nothing else tells it the content is instructions rather than reference material. The SDD stub fires in response to an explicit SDD command, so the surrounding request already supplies that framing.

This split is the contract the spec states: *Session-Bootstrap Block Wording* requires the workspace clarification and the active-session-instructions framing; *SDD-Stub Block Wording* requires the command list, the imperative read, and no orchestration detail. An earlier draft of this decision required all clauses in every stub; the spec and the implementation both settled on the split above, and this section was corrected to match rather than the other way round.

### D6: Ordering unchanged; block position no longer matters

Persona→SDD→Engram preserved (component order at `sync.go:324-326` is load-bearing for legacy fingerprint stripping). Bootstrap lands last on fresh installs (append semantics). Justified: position only mattered under truncation; the root is now <12k so the whole file loads, and the 2/2 prototype validation succeeded with the SDD stub near EOF.

### D7: Byte-budget warning

Check at end of sync (`RenderSyncReport`, `sync.go:~1570`, or its call site) and at install completion (`RenderInstallManualActions`, `run.go:1546` — already the home for non-fatal completion actions): read the assembled `SystemPromptFile` for shared-root agents; `len(bytes) > 12000` → WARNING (bytes ≥ chars, conservative). Message: path, actual size, 12,000 limit, "content past the limit is silently truncated in Antigravity sessions", remediation hint. Never fails.

### D8 (discovered in Phase 4): persona must own a marker-bound section

**Problem**: the Phase 4 convergence test (task 4.6) proved D3 false for the root. `persona.Inject` on `StrategyFileReplace` adapters wrote the raw persona asset as the ENTIRE prompt file whenever a Gentleman variant was selected, because `preserveManagedSections` returned early for `isGentlemanConversationPersona`. Two consequences, both reproduced:
- **Order divergence**: gemini-first yielded `[sdd][engram][persona]` while antigravity-first yielded `[persona][sdd][engram]`. Antigravity's `StrategyAppendToFile` branch strips the unmarked persona via `StripLegacyPersonaBlock` and re-appends it marked at the end, so the same inputs produced different bytes.
- **Silent stub loss (the serious one)**: a persona-only sync for gemini wiped BOTH stubs from the shared root. Pre-split that destroyed inline bodies; post-split the root is the only pointer to `~/.gemini/references/`, so it disconnects engram and the SDD orchestrator with no error and no warning.

**Choice**: every adapter whose prompt file is shared with another component now writes exactly one marker-bound `gentle-ai:persona` section and preserves managed sections, regardless of persona variant — the shape OpenCode already had via a special case. That special case is gone; the path is uniform across all three affected strategies:
- `StrategyFileReplace` — gemini-cli, cursor, codex, qwen, kilocode, opencode.
- `StrategyInstructionsFile` (VS Code) and `StrategySteeringFile` (Kiro) — same shape, plus the YAML header those two require. `wrapInstructionsFile`/`wrapSteeringFile` are replaced by frontmatter constants plus `ensureFrontmatter`, which seeds the header only when the file has none and never rewrites an existing one.

`StrategyAppendToFile` (antigravity, windsurf, pi) already had the correct shape and is untouched.

**Frontmatter ownership**: persona and sdd each hold their own copy of the header constants, and both seed the same file, so drift between them would make the file's metadata depend on component order. They are now byte-identical, and persona adopted sdd's existing `description` wording so the shipped `sdd-vscode-instructions.golden` did not have to change. `TestSharedPromptFileConvergesAcrossComponentOrder` pins the invariant on rendered bytes rather than on the unexported constants.

**Evidence**: `internal/components/persona/shared_root_preservation_test.go` (managed-section preservation, marker-bound shape, idempotency, user-content safety, adapter-order convergence, and header seeding/duplication for the wrapped strategies), `TestEnsureFrontmatterAddsKiroHeaderWithoutDuplicating` in `internal/components/persona/inject_test.go` (replaces the former `wrapSteeringFile` test, same guarantee), plus `TestSyncOrderConvergence` and `TestSharedPromptFileConvergesAcrossComponentOrder` in `internal/components/reference_layout_integration_test.go`. Golden churn was one file: `persona-kiro-gentleman.golden` gained the two persona markers. No persona goldens exist for gemini, cursor, codex, qwen or kilocode.

**What is deliberately NOT asserted**: byte-equality of a wrapped prompt file across arbitrary *component* orders. The pipeline fixes persona before SDD at `internal/cli/sync.go:329-331` and that order is load-bearing (persona's `StripLegacyPersonaBlock` would otherwise treat freshly written managed sections as legacy), so "SDD first" is a state the product never produces. Block order across *agents* sharing one root is asserted, because both orders are real there.

**Scope note**: this supersedes D6's claim that block position no longer matters — it did not matter for truncation, but it did make D3's convergence claim false. Ordering is now stable because persona occupies a replaceable section instead of the whole file.

## Data Flow

    persona.Inject ──→ GEMINI.md [persona]
    sdd.Inject ──→ references/sdd-orchestrator.md (priority asset)
               └─→ GEMINI.md [sdd-orchestrator stub]
    engram.Inject ──→ references/engram-protocol.md
                  └─→ GEMINI.md [engram-protocol bootstrap stub]
    sync summary ──→ byte-budget check ──→ WARNING

Budget: the prototype figure (≈10.5KB shipped, 11,115 B minus test block) included inline trigger-rules; post-merge the root no longer carries them, so headroom grows beyond ~1.4KB. Exact number is re-measured by the task 4.3 integration test.

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/agents/gemini/adapter.go`, `internal/agents/antigravity/adapter.go` | Modify | Add `ReferencesDir` |
| `internal/components/engram/inject.go` (default case 529-546) | Modify | Reference branch: write file + bootstrap stub |
| `internal/components/engram/` bootstrap render + test | Create | `renderSessionBootstrapStub` |
| `internal/components/sdd/inject.go` (`injectFileAppend` :2048) | Modify | Stub + reference write; `SelectedAgents` priority |
| `internal/components/sdd/` gate-stub render + test | Create | `renderSDDGateStub` |
| `internal/cli/sync.go`, `internal/cli/run.go` | Modify | Populate `SelectedAgents`; budget warning |
| `internal/components/golden_test.go` + testdata | Modify | Regen sdd-gemini/sdd-antigravity/engram-antigravity against the post-merge rewritten assets; new reference-file goldens; persona goldens unchanged |

## Testing Strategy (strict TDD — RED first, `go test ./...`)

| Layer | What | Approach |
|-------|------|----------|
| Unit | Stub wording invariants (D5) | Assert required phrases present, test markers absent |
| Golden | Reference files + root stubs, both agents | New goldens for `references/*.md`; regen root goldens |
| Integration | Assembled-root byte budget | persona+sdd+engram inject → assert `len < 12000` |
| Integration | Migration idempotency | Seed legacy ~54KB inline file (verified current: 54,799 B for antigravity — persona 5,189 + sdd 42,850 + engram 6,760) → inject → stub layout; second pass byte-identical |
| Integration | Both-agents convergence | gemini→antigravity and reverse orders produce identical bytes |

## Threat Matrix

N/A — no routing, shell, subprocess, VCS/PR automation, executable-file classification, or process-integration boundary; changes are marker-based file injection.

## Migration / Rollout

Automatic on first sync (D2); no flags. Rollback: revert + sync rewrites full bodies under the same marker IDs (enabled by D2); `~/.gemini/references/` files become orphaned but inert — proposal-consistent.

## Open Questions

None blocking. Noted: stale comment at `sdd/inject.go:260-262` claims the generic persona asset embeds the orchestrator — verified false (73 lines, no SDD content); safe to ignore or fix opportunistically.
Watch item: active change `organic-rdd-recovery` on main carries a delta spec on `sdd-orchestrator-assets` (same capability as this change). Base spec is currently unchanged; if that change archives first, rebase this change's delta.
