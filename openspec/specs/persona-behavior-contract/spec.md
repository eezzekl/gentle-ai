# persona-behavior-contract Specification

## Requirements

### Requirement: Neutral Mentor Behavior Parity

The system MUST treat `neutral` as a level-neutral parity of the `gentle` mentor behavior contract. Neutral persona content MUST preserve the same senior mentor expectations as `gentle`, including concise answers, direct correction after verification, concept-first teaching, careful technical reasoning, and user-growth-oriented guidance, while MUST NOT bake any marked regional voice, regional slang, voseo, or style branding into the base asset. Regional voice is governed by the independent region axis, not by the style.

#### Scenario: Neutral receives the same mentor contract without regional voice

- GIVEN an agent persona asset is rendered with style `neutral`
- WHEN the generated instruction content is inspected
- THEN it includes the same mentor behavior expectations as `gentle` for brevity, verification, concept-first explanation, and constructive correction
- AND it does not bake Rioplatense Spanish, regional slang, voseo, or style branding into the base asset

#### Scenario: Gentle style with the Rioplatense region injects the voseo directive

- GIVEN an agent persona asset is rendered with style `gentle` and region `argentina`
- WHEN the generated instruction content is inspected
- THEN it preserves the `gentle` mentor behavior contract
- AND it includes exactly one composed language directive instructing Rioplatense Spanish (voseo) for conversational replies
- AND the regional voice comes from the language axis, not from a `gentleman` persona variant

---

### Requirement: Neutral Interaction Discipline

The neutral persona contract MUST require disciplined interaction defaults across supported agent consumers: short default answers, at most one question at a time, stopping after asking a question, no option menus or exhaustive alternatives unless a real tradeoff exists, and verification before accepting or correcting a user claim.

#### Scenario: Neutral defaults to brief replies

- GIVEN a neutral persona instruction is installed for an agent
- WHEN the agent answers a normal user request that does not require extensive detail
- THEN the instruction requires the minimum useful response
- AND it permits expansion only when the user asks or the task genuinely requires it

#### Scenario: Neutral asks one question and stops

- GIVEN a neutral persona instruction needs clarification from the user
- WHEN it asks a question
- THEN it asks at most one question
- AND it instructs the agent to stop and wait for the user's answer

#### Scenario: Neutral avoids unnecessary menus

- GIVEN a neutral persona instruction describes how to present alternatives
- WHEN there is no real fork with meaningful tradeoffs
- THEN it prohibits option menus, exhaustive lists, and multiple approaches by default

#### Scenario: Neutral verifies before agreeing or correcting

- GIVEN a user makes a technical claim
- WHEN a neutral persona instruction governs the response
- THEN it requires verification against code, docs, or other evidence before agreeing with the claim
- AND it requires explaining why the claim is wrong when evidence disproves it

---

### Requirement: Artifact Language Independence

Persona style and region voice MUST govern only direct chat replies to the user. Artifact language is governed by the `artifactsInEnglish` checkbox, not by the selected style or region. When `artifactsInEnglish` is true (the default), generated technical artifacts MUST be in English and neutral professional wording regardless of the selected style or region. When `artifactsInEnglish` is false, generated artifacts MAY use the selected reply language when the user explicitly requests it or the existing project artifact convention requires it.

#### Scenario: Neutral keeps generated artifacts in English when the checkbox is ON

- GIVEN style `neutral` is active with `artifactsInEnglish` set to true
- WHEN the system generates code, identifiers, comments, UI copy, documentation, commit messages, PR descriptions, SDD artifacts, or tests
- THEN the generated artifact content defaults to English and neutral professional wording

#### Scenario: Gentle voice does not leak into artifacts when the checkbox is ON

- GIVEN style `gentle` with region `argentina` and `artifactsInEnglish` set to true
- WHEN the system generates a technical artifact without an explicit request for regional language or tone
- THEN the artifact does not include Rioplatense slang, voseo, or regional persona voice
- AND the artifact is in English

#### Scenario: Artifacts may follow the reply language when the checkbox is OFF

- GIVEN style `gentle` with region `mexico` and `artifactsInEnglish` set to false
- WHEN the user explicitly requests artifacts in the selected reply language
- THEN the injected artifact-language directive permits generated artifacts in that language
- AND it differs from the English-only directive injected when the checkbox is ON

---

### Requirement: Claude Neutral Output Style Contract

Claude-specific neutral output-style content MUST be meaningful and MUST NOT fall back to a generic default assistant character. It MUST encode the neutral mentor behavior contract, interaction discipline, verification-first rule, and artifact language independence without regional voice.

#### Scenario: Claude neutral output-style is not default assistant behavior

- GIVEN Claude assets are generated with persona `neutral`
- WHEN the neutral output-style content is inspected
- THEN it contains explicit neutral mentor behavior instructions
- AND it contains brevity, one-question, no-menu, verification-first, and artifact-language constraints
- AND it does not describe or imply an unstyled default assistant character

#### Scenario: Claude explicit gentle output-style remains honored

- GIVEN Claude assets are generated with style `gentle`
- WHEN the output-style content is inspected
- THEN it preserves the `gentle`-specific mentor instructions
- AND the output-style content is region-neutral; regional voice is carried by the composed persona language directive, not by the output-style
- AND it is not replaced by neutral output-style content

---

### Requirement: Kimi Neutral Output Style Content

Kimi neutral output-style module content MUST be meaningful, non-empty, and semantically aligned with the generic neutral behavior contract. Empty files, placeholder text, or whitespace-only injected output-style content MUST NOT be accepted for neutral.

#### Scenario: Kimi neutral output-style is meaningful

- GIVEN Kimi assets are generated or injected with style `neutral`
- WHEN the `output-style.md` content is inspected
- THEN it is non-empty after trimming whitespace
- AND it includes neutral mentor behavior, interaction discipline, verification-first, and artifact-language constraints
- AND it excludes marked regional voice instructions baked into the base asset

#### Scenario: Kimi neutral output-style rejects placeholder-only content

- GIVEN the Kimi neutral output-style source contains only placeholder or whitespace-only content
- WHEN the asset is prepared for injection
- THEN the system treats the content as invalid for neutral parity
- AND implementation MUST provide meaningful neutral output-style content instead

---

### Requirement: Generic Neutral Asset Parity

All neutral consumers that are not covered by an agent-specific override MUST receive parity through the generic neutral persona or output-style asset. Agent-specific assets MAY adapt wording to platform mechanics, but MUST NOT weaken the neutral behavior contract. For adapters with an active output-style channel (Claude Code, Kimi), parity with the generic neutral requirements MUST be evaluated over the COMBINED persona-residual + output-style channel, not over the persona file in isolation — the residual file alone deliberately omits tone/language/philosophy content that now lives exclusively in the output style.

#### Scenario: Non-agent-specific consumers receive generic neutral parity

- GIVEN an agent or surface consumes the generic neutral persona asset
- WHEN neutral instructions are rendered for that consumer
- THEN the rendered content includes the neutral mentor behavior contract
- AND it includes brevity, one-question, no-menu, verification-first, and artifact-language constraints

#### Scenario: Agent-specific neutral assets do not weaken generic behavior

- GIVEN an agent has its own neutral persona or output-style asset
- WHEN that asset is compared against the generic neutral behavior contract
- THEN it preserves all generic neutral requirements
- AND any agent-specific differences are limited to platform-accurate wording or installation mechanics

#### Scenario: Output-style-capable adapters are evaluated on the combined channel

- GIVEN Claude Code or Kimi has a residual neutral persona section plus a reconciled neutral output-style asset
- WHEN that adapter's neutral behavior contract is compared against the generic neutral requirements
- THEN the combined content of the residual persona section and the output-style asset preserves all generic neutral requirements
- AND the residual persona section alone is not required to independently preserve tone/language/philosophy content

---

### Requirement: Safe Persona Fallback Semantics

When persisted persona state is unreadable or holds an unknown/invalid value, sync and persona resolution MUST NOT silently introduce a regional voice the user did not choose. The fallback MUST be a default-safe style that does not inject an unselected regional voice. Known legacy migration cases are distinct: an empty or absent `persona` field in a readable prior-version state.json migrates per the migration contract (R1) to `gentle` style with the `argentina` (Rioplatense) region, because that reproduces the prior gentleman default the user already had.

#### Scenario: Absent legacy persona field migrates to gentle + Rioplatense

- GIVEN a readable prior-version state.json with an empty or absent `persona` field
- WHEN sync resolves the persona to apply
- THEN it migrates to style `gentle` with region `argentina` (Rioplatense) per the migration contract
- AND `artifactsInEnglish` resolves to `true`

#### Scenario: Invalid persisted persona does not inject an unselected regional voice

- GIVEN persisted persona state contains an unknown or invalid value not covered by migration aliases
- WHEN sync resolves the persona to apply
- THEN it does not select a regional voice implicitly
- AND it applies a default-safe style without an unselected regional voice

#### Scenario: Unreadable persisted persona does not inject an unselected regional voice

- GIVEN persisted persona state cannot be read
- WHEN sync resolves the persona to apply
- THEN it does not select a regional voice implicitly
- AND it applies a default-safe style without an unselected regional voice
- AND it may surface a warning if the sync command already reports recoverable configuration issues

---

### Requirement: Explicit Persona Selection Preservation

Explicit style and region selections MUST remain authoritative. When the user explicitly selects style `gentle` with a region, the system MUST apply the `gentle` mentor behavior and inject that region's composed voice directive; when the user explicitly selects `neutral`, the system MUST apply neutral parity behavior with no marked regional voice baked into the base asset.

#### Scenario: Explicit gentle + region selection remains honored during sync

- GIVEN the user has explicitly selected style `gentle` with region `argentina`
- WHEN sync resolves and applies persona assets
- THEN the `gentle` persona assets are selected
- AND the composed Rioplatense (voseo) language directive remains present

#### Scenario: Explicit neutral selection remains honored during sync

- GIVEN the user has explicitly selected style `neutral`
- WHEN sync resolves and applies persona assets
- THEN neutral persona assets are selected
- AND the rendered content includes neutral parity behavior with no marked regional voice in the base asset

#### Scenario: Fallback does not override an explicit selection

- GIVEN a valid explicit persona selection exists
- WHEN sync applies persona assets
- THEN fallback logic is not used to replace that explicit selection
- AND the selected persona remains the source of truth for rendered persona content

---

### Requirement: Canonical Tone Channel for Output-Style-Capable Adapters

For any adapter with an active output-style channel — Claude Code (gated by `SupportsOutputStyles()`) and Kimi (explicit carve-out) — the output-style asset MUST be the exclusive source of persona tone, language, and philosophy. The system-prompt persona section for these adapters MUST be reduced to a residual block containing only `## Rules`, `## Expertise`, `## Contextual Skill Loading (MANDATORY)`, a one-line pointer to the output style, plus any agent-native tooling section identified in the design's disposition tables (Kimi: `## Kimi-native notes`). The residual MUST NOT duplicate tone/language/philosophy content.

#### Scenario: Claude and Kimi residual sections carry no tone content

- GIVEN Claude or Kimi assets are generated with persona `gentleman` or `neutral`
- WHEN the CLAUDE.md or KIMI.md-included persona section is inspected
- THEN it contains only Rules, Expertise, Contextual Skill Loading, a pointer to the output style, and any agent-native tooling section identified in the design's disposition tables (Kimi: `## Kimi-native notes`)
- AND it contains no tone, language, or philosophy prose

---

### Requirement: Persona Section Parity for Non-Output-Style Adapters

Adapters without an active output-style channel MUST continue to receive the full persona section — tone, language, philosophy, rules, expertise, and skill loading — unchanged by this change.

#### Scenario: Non-output-style adapter keeps full persona section

- GIVEN an adapter other than Claude Code or Kimi is generated with a persona
- WHEN its persona asset is inspected
- THEN it contains the full persona section, including tone/language/philosophy, as before this change

---

### Requirement: Persona Drift Reconciliation

Where the system-prompt and output-style assets for a persona variant have drifted into independently authored paraphrases, the output-style asset MUST be reconciled to contain the union of normative tone rules found in either copy. Reconciliation MUST NOT drop a rule present in only one of the two drifted copies.

#### Scenario: Gentleman and Neutral reconciliation preserve rules from both copies

- GIVEN `claude/persona-{variant}.md` and `claude/output-style-{variant}.md` contain divergent normative tone rules, with Neutral using a different section structure than Gentleman
- WHEN the reconciled output-style asset for each variant is produced
- THEN it includes every normative rule that existed in either source for that variant
- AND no rule present in either source is silently dropped

---

### Requirement: Idempotent Convergence on Re-Injection

Upgrading an installation that has an existing full (pre-slim) persona section MUST converge the on-disk system-prompt persona section to the residual block via the marker-delimited section-replace mechanism. Repeated injection MUST be idempotent.

#### Scenario: Existing full section converges to residual, then stays stable

- GIVEN an installation has a full, pre-slim persona section on disk
- WHEN injection runs after this change, then runs again
- THEN the first run replaces the section with the residual block, leaving no orphaned tone content outside the marker boundaries
- AND the second run leaves the on-disk content unchanged

---

### Requirement: Legacy Fingerprint Continuity

Uninstall and auto-heal fingerprint checks that rely on literal markers from the persona section (e.g., `## Rules`, `Senior Architect`) MUST continue to match on-disk content after tone/philosophy content is removed, or the fingerprint constants MUST be updated so uninstall and auto-heal continue to function correctly.

#### Scenario: Uninstall and auto-heal fingerprints still match after slimming

- GIVEN a Claude or Kimi installation has the residual persona section on disk
- WHEN the uninstall cleaner and auto-heal each check their persona-section fingerprint
- THEN both checks match the residual on-disk content
- AND auto-heal does not false-positive a repair

---

### Requirement: Kimi Non-Duplication and Documentation Accuracy

Kimi's `KIMI.md` module includes MUST NOT deliver tone/language/philosophy content twice per session. `docs/agents.md` MUST accurately describe Kimi's output-style status as an active canonical tone channel, not as unsupported.

#### Scenario: Kimi delivers tone content once per session

- GIVEN Kimi assets are installed with an active persona
- WHEN the rendered session content is inspected
- THEN tone/language/philosophy content appears exactly once, sourced from the output-style module

#### Scenario: docs/agents.md reflects Kimi's real output-style status

- GIVEN `docs/agents.md` documents the Output Styles column per adapter
- WHEN the Kimi row is inspected
- THEN it accurately reflects that Kimi has an active output-style channel used as the canonical tone source

---

### Requirement: Per-Adapter Test and Golden Coverage

Automated tests and golden fixtures MUST cover the residual persona section, the reconciled canonical output-style content, and unaffected non-output-style adapter behavior, and MUST be updated in strict-TDD lockstep with implementation changes.

#### Scenario: Claude goldens capture the residual section; Kimi is covered by unit assertions

- GIVEN Claude golden fixtures are regenerated after this change
- WHEN the persona section of the golden is inspected
- THEN it matches the residual block, not the prior full section
- AND, since no Kimi golden fixtures exist, Kimi's residual persona section and reconciled output-style content are covered by `inject_test.go` unit assertions instead

#### Scenario: Non-output-style adapter goldens remain unchanged

- GIVEN a golden fixture for a non-output-style adapter existed before this change
- WHEN the same fixture is regenerated after this change
- THEN its persona section content is unchanged
