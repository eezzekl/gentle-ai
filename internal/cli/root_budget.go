package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
)

// rootBudgetLimit is the number of characters Antigravity reads from an
// auto-loaded rules file before it silently truncates. Content past this point
// never reaches the model, and nothing reports that it was dropped — which is
// why the whole stub+references layout exists.
const rootBudgetLimit = 12000

// sharedReferenceRootAdapter is the adapter capability that marks the shared,
// truncating root: gemini-cli and antigravity both write ~/.gemini/GEMINI.md and
// both keep their large bodies under ReferencesDir. Every other agent owns its
// own prompt file with no such limit, so the budget must not be applied to it.
type sharedReferenceRootAdapter interface {
	ReferencesDir(homeDir string) string
}

// rootBudgetWarnings reports one warning per oversized shared root. It reads the
// assembled file rather than summing what this run wrote, because the overflow
// that matters is the total the agent will actually load — including content
// written by earlier runs, other components, and the user.
//
// Byte length is used as the character count: it is never smaller than the
// character count, so the check is conservative and cannot under-report an
// overflow. Missing and unreadable roots are not overflows and stay silent.
func rootBudgetWarnings(homeDir, workspaceDir string, scope InstallScope, adapters []agents.Adapter) []string {
	warnings := make([]string, 0)
	seen := make(map[string]struct{}, len(adapters))

	for _, adapter := range adapters {
		if _, ok := adapter.(sharedReferenceRootAdapter); !ok {
			continue
		}
		if !adapter.SupportsSystemPrompt() {
			continue
		}

		path := adapter.SystemPromptFile(componentInjectionDirScoped(homeDir, workspaceDir, scope, adapter))
		if path == "" {
			continue
		}
		// gemini-cli and antigravity resolve to the same root; warn once.
		if _, duplicate := seen[path]; duplicate {
			continue
		}
		seen[path] = struct{}{}

		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		if size := info.Size(); size > rootBudgetLimit {
			warnings = append(warnings, fmt.Sprintf(
				"WARNING: %s is %d characters, over the %s limit. Antigravity silently truncates auto-loaded rules at %s characters, so everything past that point never reaches the model. Move large sections into ~/.gemini/references/ and point at them from the root, or remove hand-added content from the file.",
				path, size, formatRootBudgetLimit(), formatRootBudgetLimit(),
			))
		}
	}

	if len(warnings) == 0 {
		return nil
	}
	return warnings
}

// formatRootBudgetLimit renders the limit the way the requirement names it, with
// a thousands separator, so the warning text matches what users read in the docs.
//
// Derived from rootBudgetLimit rather than written out: the warning tells the
// user which threshold was crossed, so a changed constant with a hard-coded
// string would have the message quote a limit the check no longer enforces.
func formatRootBudgetLimit() string {
	digits := strconv.Itoa(rootBudgetLimit)

	var b strings.Builder
	for i, digit := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(digit)
	}
	return b.String()
}
