package patchengine

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Operation struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Path        string            `json:"path"`
	Status      string            `json:"status"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}

type Plan struct {
	SchemaVersion string      `json:"schemaVersion"`
	GeneratedAt   string      `json:"generatedAt"`
	Root          string      `json:"root"`
	Operations    []Operation `json:"operations"`
}

type Result struct {
	SchemaVersion string      `json:"schemaVersion"`
	GeneratedAt   string      `json:"generatedAt"`
	Status        string      `json:"status"`
	Root          string      `json:"root"`
	Operations    []Operation `json:"operations"`
	Checks        []Operation `json:"checks"`
	Message       string      `json:"message"`
}

func Inspect(cfg config.Config) error {
	plan, err := BuildPlan(cfg)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func PlanCommand(cfg config.Config) error {
	plan, err := BuildPlan(cfg)
	if err != nil {
		return err
	}
	if err := WritePlan(cfg, plan); err != nil {
		return err
	}
	return WritePlanReport(cfg, plan)
}

func Validate(cfg config.Config) error {
	result, err := ValidateTree(cfg)
	if err != nil {
		if writeErr := WriteResult(cfg, result); writeErr != nil {
			fmt.Fprintf(os.Stderr, "patch-engine: failed to write validation result: %v\n", writeErr)
		}
		return err
	}
	if err := WriteResult(cfg, result); err != nil {
		return err
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func BuildPlan(cfg config.Config) (Plan, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	root := cfg.OSHome

	ops := []Operation{}

	for _, path := range discoverFiles(root, []string{".go", ".sh", ".json", ".md"}) {
		kind := classify(path)
		ops = append(ops, Operation{
			ID:          safeID(kind + "." + rel(root, path)),
			Kind:        kind,
			Path:        rel(root, path),
			Status:      "detected",
			Description: "Patchable source artifact discovered.",
			Metadata: map[string]string{
				"absolutePath": path,
			},
		})
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].Path < ops[j].Path
	})

	return Plan{
		SchemaVersion: "aift.patch.plan.v1",
		GeneratedAt:   now,
		Root:          root,
		Operations:    ops,
	}, nil
}

func ValidateTree(cfg config.Config) (Result, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result := Result{
		SchemaVersion: "aift.patch.result.v1",
		GeneratedAt:   now,
		Status:        "ready",
		Root:          cfg.OSHome,
		Operations:    []Operation{},
		Checks:        []Operation{},
		Message:       "Patch engine validation passed.",
	}

	checks := []struct {
		id      string
		command []string
		desc    string
	}{
		{"gofmt", []string{"gofmt", "-w", "cmd", "internal"}, "Format Go source."},
		{"go-test", []string{"go", "test", "./..."}, "Run Go tests."},
		{"go-build", []string{"go", "build", "-o", "bin/aiftd", "./cmd/aift"}, "Build AIFT CLI."},
	}

	for _, check := range checks {
		op := Operation{
			ID:          check.id,
			Kind:        "validation",
			Path:        cfg.OSHome,
			Status:      "ready",
			Description: check.desc,
			Metadata: map[string]string{
				"command": strings.Join(check.command, " "),
			},
		}

		if output, err := run(cfg.OSHome, check.command...); err != nil {
			op.Status = "failed"
			op.Metadata["output"] = output
			result.Status = "failed"
			result.Message = fmt.Sprintf("validation failed at %s", check.id)
			result.Checks = append(result.Checks, op)
			return result, err
		} else {
			op.Metadata["output"] = output
		}

		result.Checks = append(result.Checks, op)
	}

	return result, nil
}

func WritePlan(cfg config.Config, plan Plan) error {
	out := filepath.Join(cfg.OSHome, "registry", "patch-plan.json")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(out, append(data, '\n'), 0644)
}

func WriteResult(cfg config.Config, result Result) error {
	out := filepath.Join(cfg.OSHome, "registry", "patch-result.json")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(out, append(data, '\n'), 0644)
}

func WritePlanReport(cfg config.Config, plan Plan) error {
	out := filepath.Join(cfg.OSHome, "reports", "patch-plan.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	var b strings.Builder
	b.WriteString("# AIFT Patch Engine Plan\n\n")
	b.WriteString("This plan lists patchable source artifacts discovered from the repository. It does not modify files.\n\n")
	b.WriteString("| Path | Kind | Status |\n")
	b.WriteString("|---|---|---|\n")
	for _, op := range plan.Operations {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` |\n", op.Path, op.Kind, op.Status))
	}
	return os.WriteFile(out, []byte(b.String()), 0644)
}

func discoverFiles(root string, suffixes []string) []string {
	out := []string{}
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			switch name {
			case ".git", "node_modules", ".repair-backups", "registry", "reports", "bin":
				return filepath.SkipDir
			}
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(path, suffix) {
				out = append(out, path)
				break
			}
		}
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "patch-engine: failed to walk directory %s: %v\n", root, err)
	}
	return out
}

func classify(path string) string {
	switch {
	case strings.HasSuffix(path, ".go"):
		return "go-source"
	case strings.HasSuffix(path, ".sh"):
		return "shell-script"
	case strings.HasSuffix(path, ".json"):
		return "json-document"
	case strings.HasSuffix(path, ".md"):
		return "markdown-document"
	default:
		return "file"
	}
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func rel(root, path string) string {
	value, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return value
}

func safeID(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer("/", ".", " ", "-", "_", "-", ":", "-", "\\", ".")
	value = replacer.Replace(value)
	return strings.Trim(value, ".-")
}
