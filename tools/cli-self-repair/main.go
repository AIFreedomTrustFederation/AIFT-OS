package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {
	path := "cmd/aift/main.go"

	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	src := string(b)

	src = removeDuplicateCaseBlocks(src)

	commands := extractCaseCommands(src)
	if len(commands) == 0 {
		panic("no CLI commands discovered from switch cases")
	}

	src = replaceHelpLikeFunction(src, commands)

	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		panic(err)
	}

	if err := os.MkdirAll("registry/cli", 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll("reports", 0755); err != nil {
		panic(err)
	}

	report := "# AIFT CLI Self-Repair Report\n\n## Commands\n\n"
	for _, cmd := range commands {
		report += "- " + cmd + "\n"
	}

	if err := os.WriteFile("reports/cli-self-repair.md", []byte(report), 0644); err != nil {
		panic(err)
	}

	json := "{\n  \"commands\": [\n"
	for i, cmd := range commands {
		comma := ","
		if i == len(commands)-1 {
			comma = ""
		}
		json += fmt.Sprintf("    %q%s\n", cmd, comma)
	}
	json += "  ]\n}\n"

	if err := os.WriteFile("registry/cli/cli-self-repair.json", []byte(json), 0644); err != nil {
		panic(err)
	}

	fmt.Println("CLI self-repair complete")
	fmt.Println("commands:", len(commands))
}

func extractCaseCommands(src string) []string {
	re := regexp.MustCompile(`case\s+"([^"]+)"\s*:`)
	matches := re.FindAllStringSubmatch(src, -1)

	seen := map[string]bool{}
	var commands []string

	for _, match := range matches {
		cmd := match[1]
		if cmd == "" || seen[cmd] {
			continue
		}
		seen[cmd] = true
		commands = append(commands, cmd)
	}

	sort.Strings(commands)

	return commands
}

func removeDuplicateCaseBlocks(src string) string {
	lines := strings.Split(src, "\n")
	seen := map[string]bool{}
	var out []string

	for i := 0; i < len(lines); {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, `case "`) && strings.HasSuffix(trimmed, `":`) {
			cmd := strings.Split(trimmed, `"`)[1]

			if seen[cmd] {
				i++
				for i < len(lines) {
					next := strings.TrimSpace(lines[i])
					if strings.HasPrefix(next, `case "`) || strings.HasPrefix(next, "default:") {
						break
					}
					i++
				}
				continue
			}

			seen[cmd] = true
		}

		out = append(out, line)
		i++
	}

	return strings.Join(out, "\n")
}

func replaceHelpLikeFunction(src string, commands []string) string {
	names := []string{"help", "printHelp", "usage", "printUsage"}

	for _, name := range names {
		repaired, ok := replaceFunction(src, name, commands)
		if ok {
			return repaired
		}
	}

	// Fallback: replace every fmt.Println("  command") line with one generated block
	// at the first help-looking print location.
	lines := strings.Split(src, "\n")
	var out []string
	inserted := false
	insideOldHelpPrints := false

	for _, line := range lines {
		if strings.Contains(line, `fmt.Println("Commands:")`) && !inserted {
			out = append(out, line)
			out = append(out, generatedHelpPrintLines(commands)...)
			inserted = true
			insideOldHelpPrints = true
			continue
		}

		if insideOldHelpPrints {
			if regexp.MustCompile(`fmt\.Println\("  [^"]+"\)`).MatchString(line) {
				continue
			}
			insideOldHelpPrints = false
		}

		out = append(out, line)
	}

	if inserted {
		return strings.Join(out, "\n")
	}

	panic("could not find help, printHelp, usage, printUsage, or Commands print block")
}

func replaceFunction(src string, name string, commands []string) (string, bool) {
	startNeedle := "func " + name + "("
	start := strings.Index(src, startNeedle)
	if start == -1 {
		return src, false
	}

	brace := strings.Index(src[start:], "{")
	if brace == -1 {
		return src, false
	}
	brace += start

	depth := 0
	end := -1

	for i := brace; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}

	if end == -1 {
		return src, false
	}

	body := "func " + name + "() {\n"
	body += "\tfmt.Println(\"AIFT-OS Federation Control Plane\")\n"
	body += "\tfmt.Println(\"\")\n"
	body += "\tfmt.Println(\"Commands:\")\n"
	for _, cmd := range commands {
		body += "\tfmt.Println(\"  " + cmd + "\")\n"
	}
	body += "}\n"

	return src[:start] + body + src[end:], true
}

func generatedHelpPrintLines(commands []string) []string {
	var lines []string
	for _, cmd := range commands {
		lines = append(lines, "\tfmt.Println(\"  "+cmd+"\")")
	}
	return lines
}
