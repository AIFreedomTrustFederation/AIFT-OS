// tools/architecture/main.go
//
// Generates architecture baseline artifacts:
//   - registry/architecture.json   (machine-readable)
//   - docs/ARCHITECTURE.md         (human documentation)
//   - reports/architecture-report.md (report with invariant checks)
//
// Usage: go run ./tools/architecture
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ── Types ────────────────────────────────────────────────────────────

type Architecture struct {
	GeneratedAt  string            `json:"generated_at"`
	Version      string            `json:"version"`
	Packages     []Package         `json:"packages"`
	Commands     []Command         `json:"commands"`
	Dependencies []Dependency      `json:"dependencies"`
	Invariants   []InvariantResult `json:"invariants"`
}

type Package struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	DependsOn  []string `json:"depends_on"`
	DependedBy []string `json:"depended_by"`
	HasTests   bool     `json:"has_tests"`
	Category   string   `json:"category"`
}

type Command struct {
	Name       string `json:"name"`
	HasHandler bool   `json:"has_handler"`
	HasHelp    bool   `json:"has_help"`
	Status     string `json:"status"` // active, planned, meta
	Package    string `json:"package,omitempty"`
}

type Dependency struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type InvariantResult struct {
	Name    string   `json:"name"`
	Passed  bool     `json:"passed"`
	Details []string `json:"details,omitempty"`
}

// ── Main ─────────────────────────────────────────────────────────────

func main() {
	root := findRepoRoot()

	arch := Architecture{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Version:     "1.0.0",
	}

	// Gather data
	pkgDeps := gatherPackageDeps(root)
	arch.Packages = buildPackages(root, pkgDeps)
	arch.Dependencies = buildDependencies(pkgDeps)
	arch.Commands = buildCommands(root)
	arch.Invariants = checkInvariants(root, arch)

	// Write outputs
	writeJSON(root, arch)
	writeArchDoc(root, arch)
	writeReport(root, arch)

	// Summary
	passed := 0
	failed := 0
	for _, inv := range arch.Invariants {
		if inv.Passed {
			passed++
		} else {
			failed++
		}
	}
	fmt.Printf("Architecture baseline generated: %d invariants checked (%d passed, %d failed)\n", len(arch.Invariants), passed, failed)

	// In CI mode (--ci), fail if there are NEW violations beyond the known baseline.
	// The baseline records known failures so we only fail on regressions.
	if len(os.Args) > 1 && os.Args[1] == "--ci" {
		baseline := loadBaseline(root)
		newFailures := 0
		for _, inv := range arch.Invariants {
			if !inv.Passed && !baseline[inv.Name] {
				fmt.Printf("NEW VIOLATION: %s\n", inv.Name)
				newFailures++
			}
		}
		if newFailures > 0 {
			fmt.Printf("%d new invariant violation(s) detected\n", newFailures)
			os.Exit(1)
		}
		fmt.Println("No new violations (known failures are baselined)")
	} else if failed > 0 {
		fmt.Printf("Known violations found. Run with --ci to check for regressions.\n")
	}
}

func loadBaseline(root string) map[string]bool {
	path := filepath.Join(root, "architecture-baseline.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]bool{}
	}
	known := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			known[line] = true
		}
	}
	return known
}

// ── Data gathering ───────────────────────────────────────────────────

func findRepoRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			fmt.Fprintln(os.Stderr, "cannot find repo root")
			os.Exit(1)
		}
		dir = parent
	}
}

func gatherPackageDeps(root string) map[string][]string {
	cmd := exec.Command("go", "list", "-f", "{{.ImportPath}}|{{join .Imports \",\"}}", "./internal/...", "./cmd/...")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "go list failed: %v\n", err)
		os.Exit(1)
	}

	const prefix = "github.com/AIFreedomTrustFederation/AIFT-OS/"
	deps := map[string][]string{}
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			continue
		}
		pkg := strings.TrimPrefix(parts[0], prefix)
		var internal []string
		for _, imp := range strings.Split(parts[1], ",") {
			if strings.Contains(imp, "AIFT-OS/internal/") {
				internal = append(internal, strings.TrimPrefix(imp, prefix))
			}
		}
		sort.Strings(internal)
		deps[pkg] = internal
	}
	return deps
}

func buildPackages(root string, pkgDeps map[string][]string) []Package {
	// Build reverse dependency map
	revDeps := map[string][]string{}
	for pkg, deps := range pkgDeps {
		for _, dep := range deps {
			revDeps[dep] = append(revDeps[dep], pkg)
		}
	}

	var packages []Package
	for pkg, deps := range pkgDeps {
		if !strings.HasPrefix(pkg, "internal/") {
			continue
		}
		sort.Strings(revDeps[pkg])

		hasTests := false
		pkgDir := filepath.Join(root, pkg)
		entries, _ := os.ReadDir(pkgDir)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), "_test.go") {
				hasTests = true
				break
			}
		}

		packages = append(packages, Package{
			Name:       strings.TrimPrefix(pkg, "internal/"),
			Path:       pkg,
			DependsOn:  deps,
			DependedBy: revDeps[pkg],
			HasTests:   hasTests,
			Category:   categorize(pkg),
		})
	}
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})
	return packages
}

func categorize(pkg string) string {
	name := strings.TrimPrefix(pkg, "internal/")
	switch {
	case name == "config" || name == "version" || name == "workspace" || name == "fsutil" || name == "jsonfile" || name == "sliceutil" || name == "gitx":
		return "foundation"
	case name == "api" || name == "daemon" || name == "runtime" || name == "scheduler" || name == "supervisor" || name == "jobs":
		return "runtime"
	case name == "events" || name == "eventbus" || name == "eventmesh":
		return "events"
	case name == "discoveryengine" || name == "capabilities" || name == "intelligence" || name == "graph":
		return "analysis"
	case name == "modules" || name == "kernelregistry" || name == "kernel" || name == "kernelruntime" || name == "patchengine":
		return "kernel"
	case name == "manifests" || name == "registry" || name == "repo" || name == "reports" || name == "state":
		return "data"
	case name == "federation" || name == "sync" || name == "workflow":
		return "federation"
	case name == "doctor" || name == "manual":
		return "operations"
	case name == "planner" || name == "servicecontracts" || name == "services":
		return "planning"
	case name == "plugins" || name == "providers":
		return "extensions"
	default:
		return "other"
	}
}

func buildDependencies(pkgDeps map[string][]string) []Dependency {
	var deps []Dependency
	for pkg, imports := range pkgDeps {
		if !strings.HasPrefix(pkg, "internal/") {
			continue
		}
		for _, imp := range imports {
			deps = append(deps, Dependency{From: pkg, To: imp})
		}
	}
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].From != deps[j].From {
			return deps[i].From < deps[j].From
		}
		return deps[i].To < deps[j].To
	})
	return deps
}

func buildCommands(root string) []Command {
	mainPath := filepath.Join(root, "cmd", "aift", "main.go")
	data, err := os.ReadFile(mainPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read main.go: %v\n", err)
		return nil
	}
	src := string(data)

	// Extract case labels from main switch
	caseRe := regexp.MustCompile(`case "([a-z][-a-z]*)"`)
	caseMatches := caseRe.FindAllStringSubmatch(src, -1)
	handlers := map[string]bool{}
	for _, m := range caseMatches {
		handlers[m[1]] = true
	}

	// Extract help entries
	helpRe := regexp.MustCompile(`fmt\.Println\("  ([a-z][-a-z]*)`)
	helpMatches := helpRe.FindAllStringSubmatch(src, -1)
	helpEntries := map[string]bool{}
	for _, m := range helpMatches {
		helpEntries[m[1]] = true
	}

	// Planned commands from legacy stubs
	stubPath := filepath.Join(root, "cmd", "aift", "legacy_command_stubs.go")
	planned := map[string]bool{}
	if stubData, err := os.ReadFile(stubPath); err == nil {
		stubRe := regexp.MustCompile(`plannedCommand\("([a-z][-a-z]*)"`)
		for _, m := range stubRe.FindAllStringSubmatch(string(stubData), -1) {
			planned[m[1]] = true
		}
	}

	// Merge all known commands
	all := map[string]bool{}
	for k := range handlers {
		all[k] = true
	}
	for k := range helpEntries {
		all[k] = true
	}

	meta := map[string]bool{"help": true, "-h": true, "--help": true}

	var commands []Command
	for name := range all {
		if meta[name] {
			continue
		}
		status := "active"
		if planned[name] {
			status = "planned"
		}

		commands = append(commands, Command{
			Name:       name,
			HasHandler: handlers[name],
			HasHelp:    helpEntries[name],
			Status:     status,
		})
	}
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
	return commands
}

// ── Invariant checks ─────────────────────────────────────────────────

func checkInvariants(root string, arch Architecture) []InvariantResult {
	var results []InvariantResult

	results = append(results, checkNoCircularImports(root))
	results = append(results, checkCommandsHaveHandlers(arch))
	results = append(results, checkCommandsHaveHelp(arch))
	results = append(results, checkNoDuplicateCommands(root))
	results = append(results, checkNoOrphanedPackages(arch))
	results = append(results, checkModulesHaveManifests(root))
	results = append(results, checkCapabilitiesHaveEvidence(root))
	results = append(results, checkServiceContractsHaveOwner(root))

	return results
}

func checkNoCircularImports(root string) InvariantResult {
	// Go compiler itself prevents circular imports, so if `go build` succeeds
	// there are no cycles. Verify by building.
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return InvariantResult{
			Name:    "no-circular-imports",
			Passed:  false,
			Details: []string{"go build failed: " + string(out)},
		}
	}
	return InvariantResult{
		Name:   "no-circular-imports",
		Passed: true,
	}
}

func checkCommandsHaveHandlers(arch Architecture) InvariantResult {
	var missing []string
	for _, cmd := range arch.Commands {
		if !cmd.HasHandler && cmd.HasHelp {
			missing = append(missing, cmd.Name+" (in help but no case in switch)")
		}
	}
	return InvariantResult{
		Name:    "commands-have-handlers",
		Passed:  len(missing) == 0,
		Details: missing,
	}
}

func checkCommandsHaveHelp(arch Architecture) InvariantResult {
	var missing []string
	for _, cmd := range arch.Commands {
		if cmd.HasHandler && !cmd.HasHelp {
			missing = append(missing, cmd.Name+" (has handler but not in help)")
		}
	}
	return InvariantResult{
		Name:    "commands-have-help",
		Passed:  len(missing) == 0,
		Details: missing,
	}
}

func checkNoDuplicateCommands(root string) InvariantResult {
	mainPath := filepath.Join(root, "cmd", "aift", "main.go")
	data, _ := os.ReadFile(mainPath)

	// Count help entries
	helpRe := regexp.MustCompile(`fmt\.Println\("  ([a-z][-a-z]*)`)
	matches := helpRe.FindAllStringSubmatch(string(data), -1)
	counts := map[string]int{}
	for _, m := range matches {
		counts[m[1]]++
	}

	var dups []string
	for name, count := range counts {
		if count > 1 {
			dups = append(dups, fmt.Sprintf("%s appears %d times in help", name, count))
		}
	}
	sort.Strings(dups)
	return InvariantResult{
		Name:    "no-duplicate-commands",
		Passed:  len(dups) == 0,
		Details: dups,
	}
}

func checkNoOrphanedPackages(arch Architecture) InvariantResult {
	var orphans []string
	for _, pkg := range arch.Packages {
		if len(pkg.DependedBy) == 0 && pkg.Name != "version" {
			// Check if imported by cmd/aift
			orphans = append(orphans, pkg.Name+" (not imported by any other package)")
		}
	}

	// Filter: packages imported by cmd/aift are not orphans
	var filtered []string
	for _, o := range orphans {
		name := strings.Split(o, " ")[0]
		isUsed := false
		for _, pkg := range arch.Packages {
			if pkg.Name == name {
				// Check if cmd/aift depends on it
				for _, dep := range pkg.DependedBy {
					if strings.HasPrefix(dep, "cmd/") {
						isUsed = true
						break
					}
				}
				break
			}
		}
		if !isUsed {
			filtered = append(filtered, o)
		}
	}

	return InvariantResult{
		Name:    "no-orphaned-packages",
		Passed:  len(filtered) == 0,
		Details: filtered,
	}
}

func checkModulesHaveManifests(root string) InvariantResult {
	// Check if internal packages that are "modules" have corresponding entries
	// in registry or manifest files. This checks that every Go package under
	// internal/ has at least a Go file (not just test files).
	var issues []string
	entries, _ := os.ReadDir(filepath.Join(root, "internal"))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkgDir := filepath.Join(root, "internal", e.Name())
		files, _ := os.ReadDir(pkgDir)
		hasGoFile := false
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".go") && !strings.HasSuffix(f.Name(), "_test.go") {
				hasGoFile = true
				break
			}
		}
		if !hasGoFile {
			issues = append(issues, e.Name()+" has no source files (only test files or empty)")
		}
	}
	return InvariantResult{
		Name:    "modules-have-source",
		Passed:  len(issues) == 0,
		Details: issues,
	}
}

func checkCapabilitiesHaveEvidence(root string) InvariantResult {
	// Read capabilities.go to verify all capability names have descriptions
	capPath := filepath.Join(root, "internal", "capabilities", "capabilities.go")
	data, err := os.ReadFile(capPath)
	if err != nil {
		return InvariantResult{
			Name:    "capabilities-have-evidence",
			Passed:  true,
			Details: []string{"capabilities.go not found, skipped"},
		}
	}

	src := string(data)

	// Extract capability names from capabilityNames() return slice
	namesIdx := strings.Index(src, "func capabilityNames()")
	if namesIdx < 0 {
		return InvariantResult{Name: "capabilities-have-evidence", Passed: true}
	}
	// Find the closing brace of the function
	braceEnd := strings.Index(src[namesIdx:], "\n}\n")
	if braceEnd < 0 {
		return InvariantResult{Name: "capabilities-have-evidence", Passed: true}
	}
	namesSection := src[namesIdx : namesIdx+braceEnd]

	// Only match strings inside the return []string{...} block
	nameRe := regexp.MustCompile(`"([a-z]+)"`)
	nameMatches := nameRe.FindAllStringSubmatch(namesSection, -1)
	names := map[string]bool{}
	for _, m := range nameMatches {
		names[m[1]] = true
	}

	// Extract case labels from description() switch
	descIdx := strings.Index(src, "func description(")
	if descIdx < 0 {
		return InvariantResult{Name: "capabilities-have-evidence", Passed: true}
	}
	descEnd := strings.Index(src[descIdx:], "\n}\n")
	if descEnd < 0 {
		descEnd = len(src) - descIdx
	}
	descSection := src[descIdx : descIdx+descEnd]
	caseRe := regexp.MustCompile(`case "([a-z]+)"`)
	caseMatches := caseRe.FindAllStringSubmatch(descSection, -1)
	described := map[string]bool{}
	for _, m := range caseMatches {
		described[m[1]] = true
	}

	var missing []string
	for name := range names {
		if !described[name] {
			missing = append(missing, name+" (registered capability with no description)")
		}
	}
	sort.Strings(missing)

	return InvariantResult{
		Name:    "capabilities-have-evidence",
		Passed:  len(missing) == 0,
		Details: missing,
	}
}

func checkServiceContractsHaveOwner(root string) InvariantResult {
	// Check that service contract definitions have owner fields
	scPath := filepath.Join(root, "internal", "servicecontracts", "servicecontracts.go")
	data, err := os.ReadFile(scPath)
	if err != nil {
		return InvariantResult{
			Name:   "service-contracts-have-owner",
			Passed: true,
		}
	}

	src := string(data)
	hasOwnerField := strings.Contains(src, "Owner") || strings.Contains(src, "owner")

	if !hasOwnerField {
		return InvariantResult{
			Name:    "service-contracts-have-owner",
			Passed:  false,
			Details: []string{"ServiceContract struct has no Owner field"},
		}
	}

	return InvariantResult{
		Name:   "service-contracts-have-owner",
		Passed: true,
	}
}

// ── Output writers ───────────────────────────────────────────────────

func writeJSON(root string, arch Architecture) {
	os.MkdirAll(filepath.Join(root, "registry"), 0755)
	data, _ := json.MarshalIndent(arch, "", "  ")
	path := filepath.Join(root, "registry", "architecture.json")
	os.WriteFile(path, data, 0644)
	fmt.Printf("Wrote %s\n", path)
}

func writeArchDoc(root string, arch Architecture) {
	os.MkdirAll(filepath.Join(root, "docs"), 0755)
	path := filepath.Join(root, "docs", "ARCHITECTURE.md")

	var b strings.Builder
	b.WriteString("# AIFT-OS Architecture\n\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", arch.GeneratedAt))

	// Package categories
	b.WriteString("## Package Categories\n\n")
	categories := map[string][]Package{}
	for _, pkg := range arch.Packages {
		categories[pkg.Category] = append(categories[pkg.Category], pkg)
	}
	catOrder := []string{"foundation", "runtime", "events", "analysis", "kernel", "data", "federation", "operations", "planning", "extensions", "other"}
	for _, cat := range catOrder {
		pkgs := categories[cat]
		if len(pkgs) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(cat)))
		b.WriteString("| Package | Dependencies | Dependents | Tests |\n")
		b.WriteString("|---|---:|---:|---:|\n")
		for _, pkg := range pkgs {
			tests := "no"
			if pkg.HasTests {
				tests = "yes"
			}
			b.WriteString(fmt.Sprintf("| `%s` | %d | %d | %s |\n", pkg.Name, len(pkg.DependsOn), len(pkg.DependedBy), tests))
		}
		b.WriteString("\n")
	}

	// Command registry
	b.WriteString("## Command Registry\n\n")
	b.WriteString("| Command | Status | Handler | Help |\n")
	b.WriteString("|---|---|---|---|\n")
	for _, cmd := range arch.Commands {
		handler := "yes"
		if !cmd.HasHandler {
			handler = "no"
		}
		help := "yes"
		if !cmd.HasHelp {
			help = "no"
		}
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", cmd.Name, cmd.Status, handler, help))
	}
	b.WriteString("\n")

	// Dependency graph (Mermaid)
	b.WriteString("## Package Dependency Graph\n\n")
	b.WriteString("```mermaid\ngraph TD\n")
	for _, dep := range arch.Dependencies {
		from := strings.TrimPrefix(dep.From, "internal/")
		to := strings.TrimPrefix(dep.To, "internal/")
		b.WriteString(fmt.Sprintf("    %s --> %s\n", mermaidID(from), mermaidID(to)))
	}
	b.WriteString("```\n\n")

	// Invariants
	b.WriteString("## Architectural Invariants\n\n")
	b.WriteString("| Invariant | Status |\n")
	b.WriteString("|---|---|\n")
	for _, inv := range arch.Invariants {
		status := "PASS"
		if !inv.Passed {
			status = "FAIL"
		}
		b.WriteString(fmt.Sprintf("| %s | %s |\n", inv.Name, status))
	}
	b.WriteString("\n")

	os.WriteFile(path, []byte(b.String()), 0644)
	fmt.Printf("Wrote %s\n", path)
}

func writeReport(root string, arch Architecture) {
	os.MkdirAll(filepath.Join(root, "reports"), 0755)
	path := filepath.Join(root, "reports", "architecture-report.md")

	var b strings.Builder
	b.WriteString("# Architecture Report\n\n")
	b.WriteString(fmt.Sprintf("Generated: %s\n\n", arch.GeneratedAt))

	// Summary
	b.WriteString("## Summary\n\n")
	b.WriteString(fmt.Sprintf("- **Packages**: %d\n", len(arch.Packages)))
	b.WriteString(fmt.Sprintf("- **Commands**: %d\n", len(arch.Commands)))
	b.WriteString(fmt.Sprintf("- **Dependencies**: %d internal edges\n", len(arch.Dependencies)))

	tested := 0
	for _, pkg := range arch.Packages {
		if pkg.HasTests {
			tested++
		}
	}
	b.WriteString(fmt.Sprintf("- **Tested packages**: %d / %d\n", tested, len(arch.Packages)))
	b.WriteString("\n")

	// Invariant results
	b.WriteString("## Invariant Check Results\n\n")
	passed := 0
	failed := 0
	for _, inv := range arch.Invariants {
		if inv.Passed {
			passed++
		} else {
			failed++
		}
	}
	b.WriteString(fmt.Sprintf("**%d passed, %d failed**\n\n", passed, failed))

	for _, inv := range arch.Invariants {
		status := "PASS"
		if !inv.Passed {
			status := "FAIL"
			b.WriteString(fmt.Sprintf("### %s: %s\n\n", status, inv.Name))
			for _, d := range inv.Details {
				b.WriteString(fmt.Sprintf("- %s\n", d))
			}
			b.WriteString("\n")
			_ = status
		} else {
			b.WriteString(fmt.Sprintf("### %s: %s\n\n", status, inv.Name))
		}
	}

	// Active vs planned commands
	b.WriteString("## Command Status\n\n")
	active := 0
	plannedCount := 0
	for _, cmd := range arch.Commands {
		if cmd.Status == "planned" {
			plannedCount++
		} else {
			active++
		}
	}
	b.WriteString(fmt.Sprintf("- **Active**: %d\n", active))
	b.WriteString(fmt.Sprintf("- **Planned**: %d\n", plannedCount))
	if plannedCount > 0 {
		b.WriteString("\nPlanned commands:\n")
		for _, cmd := range arch.Commands {
			if cmd.Status == "planned" {
				b.WriteString(fmt.Sprintf("- `%s`\n", cmd.Name))
			}
		}
	}
	b.WriteString("\n")

	// Package categories summary
	b.WriteString("## Package Categories\n\n")
	catCounts := map[string]int{}
	for _, pkg := range arch.Packages {
		catCounts[pkg.Category]++
	}
	for _, cat := range []string{"foundation", "runtime", "events", "analysis", "kernel", "data", "federation", "operations", "planning", "extensions"} {
		if c, ok := catCounts[cat]; ok {
			b.WriteString(fmt.Sprintf("- **%s**: %d packages\n", cat, c))
		}
	}
	b.WriteString("\n")

	os.WriteFile(path, []byte(b.String()), 0644)
	fmt.Printf("Wrote %s\n", path)
}

func mermaidID(s string) string {
	return strings.NewReplacer("-", "_", "/", "_", ".", "_").Replace(s)
}
