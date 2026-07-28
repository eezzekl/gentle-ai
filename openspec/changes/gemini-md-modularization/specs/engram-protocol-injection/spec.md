# Delta for Engram Protocol Injection

## ADDED Requirements

### Requirement: Reference-File Emission for Gemini/Antigravity

For the gemini-cli and antigravity adapters, the reference file MUST be the verified redundant channel required by the existing per-adapter slimming requirement: the full engram protocol content MUST be written to `~/.gemini/references/engram-protocol.md` on install and sync (not inlined in the root default-case body), and the root `session-bootstrap` block MUST act as the pointer that requirement calls for.

#### Scenario: Full protocol written to the reference file

- GIVEN gentle-ai installs or syncs for gemini-cli or antigravity
- WHEN the engram component runs
- THEN `~/.gemini/references/engram-protocol.md` contains the full protocol content verbatim (modulo template rendering) with no test-marker content

#### Scenario: Root bootstrap block is the required pointer

- GIVEN the reference file has been written
- WHEN the root `GEMINI.md` is assembled
- THEN the `session-bootstrap` block is the pointer to the full protocol location for this verified-channel adapter pair
