# Delta for Persona Behavior Contract

## ADDED Requirements

### Requirement: Marker-Bound Persona Section in Shared Prompt Files

For every adapter whose persona lands in a prompt file that other components also write into, persona content MUST be written as a single marker-delimited `gentle-ai:persona` section rather than as the whole file body. Persona injection MUST NOT overwrite, reorder, or drop marker-delimited sections owned by other components, and MUST NOT depend on which persona variant was selected.

This closes a real defect rather than restating existing behavior. Before this change, `StrategyFileReplace` adapters wrote the raw persona asset as the entire prompt file whenever a Gentleman persona was selected, because the managed-section preservation helper returned early for those variants. Under the stub+references layout the `sdd-orchestrator` and `engram-protocol` sections are the only pointer to `~/.gemini/references/`, so overwriting the file silently disconnected both the engram protocol and the SDD orchestrator contract with no error and no warning.

#### Scenario: Persona injection preserves sections owned by other components

- GIVEN a prompt file already contains the managed `sdd-orchestrator` and `engram-protocol` sections
- WHEN the persona component injects any persona variant, including both Gentleman variants
- THEN both sections and their bodies survive unchanged, and the persona content is present in its own `gentle-ai:persona` section

#### Scenario: Persona section shape does not vary by persona variant

- GIVEN a fresh install for an adapter that shares its prompt file with other components
- WHEN persona injection runs for a Gentleman variant and for a non-Gentleman variant
- THEN both produce exactly one marker-delimited `gentle-ai:persona` section

#### Scenario: User-authored content survives persona injection

- GIVEN a prompt file contains text the user wrote that gentle-ai has never owned
- WHEN persona injection runs
- THEN that text is preserved; only gentle-ai's own legacy unmarked persona output may be replaced

### Requirement: Single Frontmatter Header in Wrapped Prompt Files

For adapters whose prompt file opens with a YAML frontmatter header (VS Code instructions files, Kiro steering files), persona injection MUST seed that header only when the file does not already have one, and MUST NOT rewrite or duplicate an existing header. Every component that seeds the header for a given adapter MUST write byte-identical header content, so the file's metadata does not depend on which component ran first.

A second header further down the file is not metadata: it renders as a literal `---` divider inside the loaded prompt.

#### Scenario: Header is seeded once on a fresh install

- GIVEN a fresh install for an adapter whose prompt file uses a YAML header
- WHEN persona injection runs
- THEN the file opens with exactly one header, followed by the marker-bound persona section

#### Scenario: Existing header is preserved, not replaced

- GIVEN the prompt file already opens with a YAML header, whether seeded by another component or authored by the user
- WHEN persona injection runs
- THEN that header is preserved byte-for-byte and no second header is added

#### Scenario: Header content does not depend on which component seeded it

- GIVEN both the persona and SDD components can seed the header for the same adapter
- WHEN each runs first on an otherwise empty file
- THEN both produce byte-identical header content

### Requirement: Shared-Root Convergence Across Adapter Order

For adapters that write the same shared root prompt file, the assembled result MUST NOT depend on the order in which those adapters are synced. Syncing gemini-cli then antigravity MUST produce the same root bytes as syncing antigravity then gemini-cli.

#### Scenario: Shared root converges regardless of sync order

- GIVEN gemini-cli and antigravity are both selected, and both write `~/.gemini/GEMINI.md`
- WHEN persona, SDD, and engram are injected for one adapter and then the other, in either order
- THEN the resulting root file and both reference files are byte-identical between the two orders

## MODIFIED Requirements

### Requirement: Per-Adapter Test and Golden Coverage

Persona coverage MUST additionally assert, for every adapter whose prompt file is shared with other components, that persona injection preserves those components' marker-delimited sections and writes exactly one marker-bound persona section.
(Previously: coverage asserted per-adapter persona content and goldens without any assertion that a persona sync leaves another component's sections intact, which is why the overwrite went undetected.)

#### Scenario: Shared-file preservation is covered per adapter

- GIVEN the persona test suite runs
- WHEN an adapter writes its persona into a prompt file shared with other components
- THEN a test asserts the other components' sections survive persona injection, for both Gentleman and non-Gentleman variants
