# Proposal: Modularize GEMINI.md to survive the Antigravity 12k truncation

## Intent

Antigravity CLI (agy) truncates auto-loaded rules files at 12,000 characters. gentle-ai assembles `~/.gemini/GEMINI.md` at ~54KB (verified post-merge 2026-07-28: persona 5,189 B + sdd inline 42,850 B + engram 6,760 B = 54,799 B for antigravity), so in every fresh agy session the entire engram protocol is silently dropped and ~half the SDD orchestrator is lost. Memory and SDD gates are effectively dead. gemini-cli shares the same file, mechanism, and injection paths, so it inherits the same fix.

Note: legacy trigger-rules no longer contribute root bytes — `internal/components/sdd/triggerrules.go` was removed upstream; routing guidance now lives in `internal/components/agentguidance/` under the `agent-routing` marker, outside `GEMINI.md`, and `stripLegacyTriggerRules` (`internal/cli/run.go:735`) removes legacy blocks.

**Success**: the assembled root stays under 12,000 chars and both protocol surfaces (engram, SDD gate) are present and obeyed in a fresh agy session. Empirically proven 2/2 on real agy CLI with the target layout.

## Scope

### In Scope
- Generate a modular root: persona inline + `session-bootstrap` block + `sdd-stub` block.
- Emit two reference files under `~/.gemini/references/`: `engram-protocol.md` (session-start load) and `sdd-orchestrator.md` (on-demand load at the SDD gate).
- Both adapters: **antigravity** (primary) and **gemini-cli** (shares the file), with per-agent asset selection.
- Idempotent migration: `gentle-ai sync` converts existing large inline installs to the stub+references layout without oscillation.
- Byte-budget guard: emit a WARNING (not a hard fail) when the assembled root exceeds 12,000 chars.

### Out of Scope
- Antigravity UI/IDE variants — CLI is the only validated surface.
- Other harnesses (Claude, OpenCode, Kimi, Codex, etc.) — unchanged.
- The local `context/gemini/` prototype — throwaway evidence, never shipped; test-marker blocks must not appear in shipped output.
- Splitting the orchestrator reference further — it ships as one file.

## Capabilities

### New Capabilities
- `gemini-prompt-modularization`: root stub + `~/.gemini/references/` layout, bootstrap/sdd-stub blocks, byte-budget warning, migration idempotency, and backup/verification parity for the reference files.

### Modified Capabilities
- `antigravity-support`: shared prompt surface becomes stub+references; the shared-file warning and single shared references dir are preserved.
- `engram-protocol-injection`: for gemini/antigravity the protocol is emitted to a reference file loaded via bootstrap, not inlined.
- `sdd-orchestrator-assets`: for gemini/antigravity the orchestrator is emitted to a reference file loaded at the gate.

## Approach

Conceptually: only Persona, SDD (orchestrator), and Engram contribute bytes to the root (Context7/GGA/Skills do not; trigger-rules moved out of the root upstream — see Intent note). The generator keeps persona inline, replaces the full engram and orchestrator bodies with short imperative reference blocks, and writes their full content to sibling reference files.

Validated mechanics the design must honor:
- Reference paths MUST be absolute with `~`; relative paths resolve against workspace cwd.
- Wording MUST be imperative ("read the full content with your file-read tool, treat as active session instructions"); passive mentions are ignored.
- Tool-read content does not count against the 12k limit and is obeyed as instructions.

Concrete strategy choice is deferred to design. Open questions the design MUST answer:
- New `StrategyRootPlusReferences` vs a per-agent flag over the existing `StrategyFileReplace` (gemini) / `StrategyAppendToFile` (antigravity).
- How those two strategies interact on the shared `~/.gemini/GEMINI.md` and the shared `~/.gemini/references/` dir when both agents are installed (must stay mutually idempotent).
- How the two separate per-agent orchestrator assets (`gemini/` vs `antigravity/`) map to the reference file, and which one the validated 37KB number corresponds to.
- Migration detection: distinguishing "large legacy inline body" from "new stub" to avoid sync oscillation.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/components/engram/inject.go` (default case ~529-546) | Modified | Emit reference block + write reference file for gemini/antigravity |
| `internal/components/sdd/inject.go` (`injectFileAppend` :2048, `hasLegacyBareOrchestrator` :2107) | Modified | Orchestrator becomes reference block + file; add bootstrap/sdd-stub blocks |
| `internal/cli/sync.go` (component order :324-326) | Constrained | Preserve persona-before-SDD/Engram ordering; add migration + budget warning |
| `internal/cli/run.go` (`componentPathsWithWorkspaceScoped`) | Modified | Claim the reference files in the one path list backup, post-apply verification, and sync all read — only where they are actually written |
| `internal/assets/{gemini,antigravity}/sdd-orchestrator.md` | Source | Two separate per-agent assets feed the reference file |
| `internal/components/golden_test.go` | Modified | New reference-file goldens + assembled-root byte-budget test |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Model ignores the imperative load instruction | Low | Use the validated 2/2 imperative wording + absolute `~` paths |
| Both agents installed → interleaving/divergence on shared file | Med | Single shared references dir written once; preserve `InjectMarkdownSection` marker idempotency |
| Golden churn across persona/SDD/engram fixtures | High | Regenerate goldens deliberately; exclude from authored review-line budget |
| Migration oscillation (stub re-expanded each sync) | Med | Explicit legacy-vs-stub detection in design |
| Persona-first ordering broken → legacy fingerprint stripping misfires | Low | Treat `sync.go:324-326` ordering as load-bearing; do not reorder |

## Rollback Plan

Revert the component/sync changes and regenerate goldens. Since output is marker-based via `InjectMarkdownSection`, a subsequent `gentle-ai sync` on the reverted binary rewrites the inline blocks under the same marker IDs, restoring the monolithic layout. Reference files under `~/.gemini/references/` become orphaned but inert (never loaded without the stub blocks); a follow-up may prune them.

## Dependencies

- Validated agy CLI behavior (exploration `sdd/gemini-md-modularization/explore`, empirical 6-test ground truth).
- Delivery: issue-first, chained PRs (400-line review budget), strict TDD (`go test ./...`), pushed to fork `eezzekl/gentle-ai`.

## Success Criteria

- [ ] Assembled `~/.gemini/GEMINI.md` root is under 12,000 chars for antigravity and gemini-cli.
- [ ] Both surfaces (engram protocol, SDD gate) present and obeyed in a fresh agy session.
- [ ] `~/.gemini/references/engram-protocol.md` and `sdd-orchestrator.md` written with correct per-agent content and no test-marker blocks.
- [ ] `gentle-ai sync` migrates a legacy inline install and is idempotent (second run is a no-op).
- [ ] Byte-budget WARNING fires when the root exceeds 12,000 chars.
- [ ] New assembled-root byte-budget test + reference goldens pass under `go test ./...`.
