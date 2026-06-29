package integration

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestHelpCoversAllHandlers verifies every command listed in `help` output
// has a corresponding handler in main.go's switch statement.
func TestHelpCoversAllHandlers(t *testing.T) {
	// Get help output
	out, err := run(t, t.TempDir(), "help")
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}

	// Parse command names from help output (lines starting with "  <word>")
	helpCommands := parseHelpCommands(out)
	if len(helpCommands) == 0 {
		t.Fatal("parsed zero commands from help")
	}

	// Read main.go to find switch cases
	root := repoRoot()
	data, err := os.ReadFile(filepath.Join(root, "cmd", "aift", "main.go"))
	if err != nil {
		t.Fatalf("cannot read main.go: %v", err)
	}
	src := string(data)

	for _, cmd := range helpCommands {
		// Skip meta-commands that don't need a case
		if cmd == "help" || cmd == "-h" || cmd == "--help" {
			continue
		}
		// Check that the command appears as a case in the switch
		if !strings.Contains(src, `case "`+cmd+`"`) {
			t.Errorf("help lists %q but no case in main.go switch", cmd)
		}
	}
}

// TestAllSwitchCasesHaveHelp verifies every case in main.go's switch
// has corresponding help text.
func TestAllSwitchCasesHaveHelp(t *testing.T) {
	root := repoRoot()
	data, err := os.ReadFile(filepath.Join(root, "cmd", "aift", "main.go"))
	if err != nil {
		t.Fatalf("cannot read main.go: %v", err)
	}
	src := string(data)

	// Extract case labels from the main switch
	caseRe := regexp.MustCompile(`case "([a-z][-a-z]*)"`)
	matches := caseRe.FindAllStringSubmatch(src, -1)

	// Get help output
	out, _ := run(t, t.TempDir(), "help")

	for _, m := range matches {
		cmd := m[1]
		if !strings.Contains(out, cmd) {
			t.Errorf("main.go has case %q but it's not in help output", cmd)
		}
	}
}

// TestNoDuplicateHelpCommands verifies no command appears twice in help output.
func TestNoDuplicateHelpCommands(t *testing.T) {
	out, err := run(t, t.TempDir(), "help")
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}

	counts := map[string]int{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "Commands:" || strings.HasPrefix(line, "AIFT-OS") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) > 0 {
			counts[parts[0]]++
		}
	}

	for cmd, count := range counts {
		if count > 1 {
			t.Errorf("command %q appears %d times in help (expected once)", cmd, count)
		}
	}
}

// parseHelpCommands extracts the primary command name from each help line.
func parseHelpCommands(helpOutput string) []string {
	var commands []string
	seen := map[string]bool{}
	for _, line := range strings.Split(helpOutput, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "Commands:" || strings.HasPrefix(line, "AIFT-OS") {
			continue
		}
		// Extract first word as the command name
		parts := strings.Fields(line)
		if len(parts) > 0 {
			cmd := parts[0]
			if !seen[cmd] {
				seen[cmd] = true
				commands = append(commands, cmd)
			}
		}
	}
	return commands
}
