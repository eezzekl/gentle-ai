# Delta for Antigravity Support

## MODIFIED Requirements

### Requirement: Antigravity shares the Gemini global prompt surface

The system MUST write global prompt/persona content for Antigravity to `~/.gemini/GEMINI.md`, and MUST write reference files for Antigravity to the shared `~/.gemini/references/` directory also used by gemini-cli.
(Previously: only covered the shared root file, not the shared references directory introduced by the stub+references layout.)

#### Scenario: Antigravity and Gemini CLI are selected together

- GIVEN both `gemini-cli` and `antigravity` are selected
- WHEN the installer applies SDD prompt content
- THEN the installer warns that both agents share `~/.gemini/GEMINI.md`
- AND the installer warns that both agents share `~/.gemini/references/`
