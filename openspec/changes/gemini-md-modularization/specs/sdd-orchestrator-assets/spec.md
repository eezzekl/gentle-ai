# Delta for SDD Orchestrator Assets

## ADDED Requirements

### Requirement: Orchestrator Reference File Emission for Gemini/Antigravity

For gemini-cli and antigravity, the per-agent SDD orchestrator asset content MUST be written to `~/.gemini/references/sdd-orchestrator.md` on install and sync instead of being inlined in the root `GEMINI.md`. The root `sdd-stub` block MUST instruct reading that file before any SDD command.

#### Scenario: Orchestrator content emitted to reference file, not inlined

- GIVEN gentle-ai installs or syncs for gemini-cli or antigravity
- WHEN the SDD component runs
- THEN the full per-agent orchestrator body is written to `~/.gemini/references/sdd-orchestrator.md`, and the root `GEMINI.md` contains only the `sdd-stub` block

### Requirement: Per-Agent Orchestrator Reference Content Selection

Syncing the SDD orchestrator component for gemini-cli MUST write `internal/assets/gemini/sdd-orchestrator.md` content into `~/.gemini/references/sdd-orchestrator.md`. Syncing for antigravity MUST write `internal/assets/antigravity/sdd-orchestrator.md` content into that same shared file.
(Note: because both agents write the same shared reference file, when both are installed the file reflects whichever agent synced most recently. Defining stable, non-oscillating behavior for that dual-install case is a design responsibility, not resolved at spec level.)

#### Scenario: gemini-cli sync writes the gemini asset

- GIVEN gemini-cli sync runs
- WHEN the reference file is written
- THEN `~/.gemini/references/sdd-orchestrator.md` contains the gemini-cli orchestrator asset content

#### Scenario: antigravity sync writes the antigravity asset

- GIVEN antigravity sync runs
- WHEN the reference file is written
- THEN `~/.gemini/references/sdd-orchestrator.md` contains the antigravity orchestrator asset content
