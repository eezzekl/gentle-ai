# Gemini Prompt Modularization Specification

## Purpose

Define the stub+references layout for `~/.gemini/GEMINI.md` (gemini-cli, antigravity) so the root stays under Antigravity's 12,000-character truncation limit while engram protocol, SDD orchestrator, and persona all remain effective. (Trigger-rules left the root upstream: `triggerrules.go` was deleted and routing guidance now lives in `internal/components/agentguidance/` outside `GEMINI.md`.)

## Requirements

### Requirement: Root Assembly Byte Budget

The assembled root `~/.gemini/GEMINI.md` MUST stay under 12,000 characters for both the gemini-cli and antigravity adapters. Persona content MUST be present inline. The full engram-protocol body and the full SDD orchestrator body MUST NOT be inline.

#### Scenario: Root under budget for antigravity

- GIVEN `gentle-ai sync` runs for antigravity
- WHEN the root `GEMINI.md` is assembled
- THEN persona is present inline, the root is under 12,000 characters, and no full engram-protocol or SDD-orchestrator body is inline

#### Scenario: Root under budget for gemini-cli

- GIVEN `gentle-ai sync` runs for gemini-cli
- WHEN the root `GEMINI.md` is assembled
- THEN the root is under 12,000 characters and no full engram-protocol or SDD-orchestrator body is inline

### Requirement: Session-Bootstrap Block Wording

The generated `session-bootstrap` block MUST imperatively instruct the model to read `~/.gemini/references/engram-protocol.md` (absolute, `~`-anchored path) with its file-read tool, to treat that content as active session instructions, and to load it before its first reply. The block MUST clarify the path is not resolved against the workspace.

#### Scenario: Bootstrap block wording is validated

- GIVEN the root is assembled
- WHEN the `session-bootstrap` block is rendered
- THEN it imperatively instructs a file-read of `~/.gemini/references/engram-protocol.md`, states the path is `~`-absolute and NOT the workspace, and requires loading before the first reply

### Requirement: SDD-Stub Block Wording

The generated `sdd-stub` block MUST list SDD commands and imperatively instruct reading `~/.gemini/references/sdd-orchestrator.md` before any SDD command or natural-language equivalent. The stub MUST NOT contain orchestration detail.

#### Scenario: Stub gates SDD commands

- GIVEN the root is assembled
- WHEN the `sdd-stub` block is rendered
- THEN it lists SDD commands, imperatively instructs reading `~/.gemini/references/sdd-orchestrator.md` first, and contains no orchestration detail

### Requirement: Reference File Generation

`~/.gemini/references/engram-protocol.md` and `~/.gemini/references/sdd-orchestrator.md` MUST be written on both install and sync, with content preserved verbatim from source assets (modulo existing template rendering). Shipped output MUST NOT contain test-marker blocks.

#### Scenario: Reference files written on install and sync

- GIVEN gentle-ai runs install or sync for gemini-cli or antigravity
- WHEN reference files are written
- THEN both reference files exist with source content preserved and no test-marker blocks present

### Requirement: Reference File Backup and Verification Parity

The reference files are managed write targets, so every deployment path MUST derive them from the same decision the writer uses, rather than repeating it. `~/.gemini/references/engram-protocol.md` and `~/.gemini/references/sdd-orchestrator.md` MUST appear in the install backup set, the sync backup set, post-apply verification, and post-sync verification whenever they are written. They MUST NOT appear in any of those sets when they are not written: a workspace-scoped run keeps both bodies inline and writes no reference file, so claiming the paths there would fail verification on a file the run was never asked to create.

#### Scenario: Reference files are backed up and verified on a global run

- GIVEN gemini-cli or antigravity is selected with the Engram or SDD component
- WHEN install or sync resolves its managed paths under global scope
- THEN both reference-file paths are present in the backup set and in the verification set, so a rewrite has a snapshot to roll back to and a broken write is caught

#### Scenario: Workspace-scoped run claims no reference file

- GIVEN the same selection under workspace scope
- WHEN install or sync resolves its managed paths
- THEN neither reference-file path is claimed, because both bodies stay inline and no reference file is written

### Requirement: Migration and Idempotency

Syncing over a legacy monolithic `GEMINI.md` (inline engram/orchestrator bodies) MUST convert it to the stub+references layout. A second sync MUST produce byte-identical output to the first (no oscillation). Persona-first ordering MUST be preserved, and other agents' content in the file MUST remain untouched.

#### Scenario: Legacy install migrates

- GIVEN a legacy `GEMINI.md` has inline engram-protocol and orchestrator bodies
- WHEN `gentle-ai sync` runs
- THEN the file converts to stub+references, persona-first ordering is preserved, and unrelated agent content is untouched

#### Scenario: Second sync is a no-op

- GIVEN a `GEMINI.md` already in stub+references layout
- WHEN `gentle-ai sync` runs again
- THEN the output is byte-identical to the prior run

### Requirement: Byte-Budget Warning

When the assembled root exceeds 12,000 characters, the system MUST emit a warning, not a failure, naming the actual character count and the 12,000 limit.

#### Scenario: Warning on overflow

- GIVEN the assembled root exceeds 12,000 characters
- WHEN sync or install completes assembly
- THEN a warning is emitted naming the actual count and the limit, and the run does not fail

### Requirement: Uninstall/Revert Safety

Uninstall or revert MUST remove the `session-bootstrap` and `sdd-stub` blocks together with their markers, leaving no active references. Orphaned reference files under `~/.gemini/references/` MUST remain inert (never loaded without the stub blocks).

#### Scenario: Uninstall removes stub blocks

- GIVEN gemini-cli or antigravity is uninstalled
- WHEN uninstall runs
- THEN the `session-bootstrap` and `sdd-stub` blocks and their markers are removed, and no active reference remains

#### Scenario: Orphaned reference files stay inert

- GIVEN reference files remain on disk after uninstall
- WHEN a future session starts
- THEN the orphaned files are never loaded because no stub block references them

### Requirement: Golden and Byte-Budget Test Coverage

Reference-file goldens MUST exist for both `engram-protocol.md` and `sdd-orchestrator.md`, for both gemini-cli and antigravity. An automated test MUST assert the assembled root stays under 12,000 characters for both adapters.

#### Scenario: Byte-budget regression test exists

- GIVEN `go test ./...` runs
- WHEN component tests execute
- THEN a test asserts the assembled root is under 12,000 characters for gemini-cli and antigravity, and reference-file goldens pass for both adapters
