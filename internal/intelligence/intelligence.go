package intelligence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/capabilities"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/fsutil"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/gitx"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/sliceutil"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type RepoIntelligence struct {
	Name              string            `json:"name"`
	Path              string            `json:"path"`
	Branch            string            `json:"branch"`
	Remote            string            `json:"remote"`
	Dirty             bool              `json:"dirty"`
	DetectedFiles     []string          `json:"detectedFiles"`
	Languages         []string          `json:"languages"`
	Frameworks        []string          `json:"frameworks"`
	Role              string            `json:"role"`
	Maturity          string            `json:"maturity"`
	Score             int               `json:"score"`
	Confidence        int               `json:"confidence"`
	CapabilitySummary map[string]string `json:"capabilitySummary"`
	Ready             []string          `json:"ready"`
	V1                []string          `json:"v1"`
	Detected          []string          `json:"detected"`
	Planned           []string          `json:"planned"`
	Broken            []string          `json:"broken"`
	Recommendations   []string          `json:"recommendations"`
}

type FederationIntelligence struct {
	GeneratedAt string             `json:"generatedAt"`
	Repos       []RepoIntelligence `json:"repos"`
	Summary     Summary            `json:"summary"`
}

type Summary struct {
	Repos          int `json:"repos"`
	Dirty          int `json:"dirty"`
	ReadyOrV1Repos int `json:"readyOrV1Repos"`
	BrokenRepos    int `json:"brokenRepos"`
	AverageScore   int `json:"averageScore"`
}

func Scan(cfg config.Config) error {
	if err := capabilities.Scan(cfg); err != nil {
		return err
	}

	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	out := FederationIntelligence{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Repos:       []RepoIntelligence{},
	}

	total := 0
	for _, r := range repos {
		ri := analyzeRepo(r)
		out.Repos = append(out.Repos, ri)
		out.Summary.Repos++
		total += ri.Score
		if ri.Dirty {
			out.Summary.Dirty++
		}
		if len(ri.Ready)+len(ri.V1) > 0 {
			out.Summary.ReadyOrV1Repos++
		}
		if len(ri.Broken) > 0 {
			out.Summary.BrokenRepos++
		}
	}

	if out.Summary.Repos > 0 {
		out.Summary.AverageScore = total / out.Summary.Repos
	}

	if err := writeRegistry(cfg, out); err != nil {
		return err
	}
	if err := writeReport(cfg, out); err != nil {
		return err
	}

	return events.Emit(cfg, "intelligence.scan", "intelligence", "real federation intelligence scan complete", map[string]string{
		"repos": fmt.Sprint(out.Summary.Repos),
		"score": fmt.Sprint(out.Summary.AverageScore),
	})
}

func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "intelligence.md")
	data, err := os.ReadFile(path)
	if err != nil {
		if err := Scan(cfg); err != nil {
			return err
		}
		data, err = os.ReadFile(path)
		if err != nil {
			return err
		}
	}
	fmt.Print(string(data))
	return nil
}

func Repo(cfg config.Config, name string) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}
	for _, r := range repos {
		if r.Name == name {
			ri := analyzeRepo(r)
			printRepo(ri)
			return nil
		}
	}
	return fmt.Errorf("repository not found: %s", name)
}

func Roadmap(cfg config.Config) error {
	var fi FederationIntelligence
	data, err := os.ReadFile(filepath.Join(cfg.OSHome, "registry", "intelligence.json"))
	if err != nil {
		if err := Scan(cfg); err != nil {
			return err
		}
		data, err = os.ReadFile(filepath.Join(cfg.OSHome, "registry", "intelligence.json"))
		if err != nil {
			return err
		}
	}
	if err := json.Unmarshal(data, &fi); err != nil {
		return err
	}

	fmt.Println("# Federation Roadmap")
	fmt.Println()
	for _, r := range fi.Repos {
		if len(r.Recommendations) == 0 {
			continue
		}
		fmt.Println("##", r.Name)
		for _, rec := range r.Recommendations {
			fmt.Println("-", rec)
		}
		fmt.Println()
	}
	return nil
}

func analyzeRepo(r workspace.Repo) RepoIntelligence {
	ri := RepoIntelligence{
		Name:              r.Name,
		Path:              r.Path,
		Branch:            gitx.Branch(r.Path),
		Remote:            gitx.Remote(r.Path),
		Dirty:             gitx.Dirty(r.Path),
		CapabilitySummary: map[string]string{},
	}

	ri.DetectedFiles = detectFiles(r.Path)
	ri.Languages = detectLanguages(r.Path)
	ri.Frameworks = detectFrameworks(r.Path)
	ri.Role = classifyRole(r.Name, ri)
	ri.CapabilitySummary, ri.Ready, ri.V1, ri.Detected, ri.Planned, ri.Broken = loadCapabilities(r.Path)
	ri.Score, ri.Maturity, ri.Confidence = score(ri)
	ri.Recommendations = recommendations(ri)

	return ri
}

func detectFiles(path string) []string {
	candidates := []string{
		"README.md", "AGENTS.md", "go.mod", "package.json", "pnpm-lock.yaml",
		"package-lock.json", "yarn.lock", "Makefile", "Dockerfile",
		"docker-compose.yml", "next.config.js", "next.config.ts",
		"tsconfig.json", "tailwind.config.js", "tailwind.config.ts",
		".aift/repo.json", ".aift/capabilities.json",
	}
	var found []string
	for _, c := range candidates {
		if fsutil.Exists(filepath.Join(path, c)) {
			found = append(found, c)
		}
	}
	return found
}

func detectLanguages(path string) []string {
	set := map[string]bool{}
	if fsutil.Exists(filepath.Join(path, "go.mod")) {
		set["Go"] = true
	}
	if fsutil.Exists(filepath.Join(path, "package.json")) || fsutil.Exists(filepath.Join(path, "tsconfig.json")) {
		set["TypeScript/JavaScript"] = true
	}
	if fsutil.Exists(filepath.Join(path, "README.md")) || fsutil.Exists(filepath.Join(path, "docs")) {
		set["Markdown/Docs"] = true
	}
	return sliceutil.SortedBoolMapKeys(set)
}

func detectFrameworks(path string) []string {
	set := map[string]bool{}
	if fsutil.Exists(filepath.Join(path, "next.config.js")) || fsutil.Exists(filepath.Join(path, "next.config.ts")) {
		set["Next.js"] = true
	}
	if fsutil.Exists(filepath.Join(path, "tailwind.config.js")) || fsutil.Exists(filepath.Join(path, "tailwind.config.ts")) {
		set["Tailwind"] = true
	}
	if fsutil.Exists(filepath.Join(path, "go.mod")) {
		set["Go CLI/Service"] = true
	}
	if fsutil.Exists(filepath.Join(path, "Dockerfile")) || fsutil.Exists(filepath.Join(path, "docker-compose.yml")) {
		set["Container"] = true
	}
	return sliceutil.SortedBoolMapKeys(set)
}

func classifyRole(name string, ri RepoIntelligence) string {
	n := strings.ToLower(name)
	switch {
	case strings.Contains(n, "aift-os"):
		return "federation-control-plane"
	case strings.Contains(n, "forge"):
		return "forge-repository-platform"
	case strings.Contains(n, "booksmith"):
		return "authoring-publishing-system"
	case strings.Contains(n, "freedom-trust"):
		return "doctrine-governance-trust"
	case strings.Contains(n, "aether") || strings.Contains(n, "coin"):
		return "economic-trust-layer"
	case strings.Contains(n, "vps"):
		return "infrastructure-node-layer"
	case strings.Contains(n, "www") || strings.Contains(n, "github.io"):
		return "public-web-portal"
	case sliceutil.Contains(ri.Frameworks, "Next.js"):
		return "web-application"
	case sliceutil.Contains(ri.Languages, "Markdown/Docs"):
		return "documentation-repository"
	default:
		return "unknown-sovereign-repository"
	}
}

func loadCapabilities(path string) (map[string]string, []string, []string, []string, []string, []string) {
	summary := map[string]string{}
	var ready, v1, detected, planned, broken []string

	data, err := os.ReadFile(filepath.Join(path, ".aift", "capabilities.json"))
	if err != nil {
		return summary, ready, v1, detected, planned, broken
	}

	var rc struct {
		Capabilities []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"capabilities"`
	}
	if json.Unmarshal(data, &rc) != nil {
		return summary, ready, v1, detected, planned, broken
	}

	for _, c := range rc.Capabilities {
		summary[c.Name] = c.Status
		switch c.Status {
		case "ready":
			ready = append(ready, c.Name)
		case "v1":
			v1 = append(v1, c.Name)
		case "detected":
			detected = append(detected, c.Name)
		case "planned":
			planned = append(planned, c.Name)
		case "broken":
			broken = append(broken, c.Name)
		}
	}

	sort.Strings(ready)
	sort.Strings(v1)
	sort.Strings(detected)
	sort.Strings(planned)
	sort.Strings(broken)

	return summary, ready, v1, detected, planned, broken
}

func score(ri RepoIntelligence) (int, string, int) {
	score := 0
	confidence := 35

	if sliceutil.Contains(ri.DetectedFiles, "README.md") {
		score += 10
		confidence += 5
	}
	if sliceutil.Contains(ri.DetectedFiles, ".aift/repo.json") {
		score += 10
		confidence += 10
	}
	if sliceutil.Contains(ri.DetectedFiles, ".aift/capabilities.json") {
		score += 10
		confidence += 10
	}
	if len(ri.Ready) > 0 {
		score += len(ri.Ready) * 8
		confidence += 10
	}
	if len(ri.V1) > 0 {
		score += len(ri.V1) * 12
		confidence += 15
	}
	if len(ri.Detected) > 0 {
		score += len(ri.Detected) * 3
	}
	if len(ri.Broken) > 0 {
		score -= len(ri.Broken) * 10
		confidence += 10
	}
	if len(ri.Frameworks) > 0 {
		score += 5
		confidence += 5
	}
	if ri.Remote != "" {
		score += 5
	}
	if ri.Dirty {
		score -= 5
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	if confidence > 100 {
		confidence = 100
	}

	maturity := "unknown"
	switch {
	case score >= 80:
		maturity = "orchestratable"
	case score >= 60:
		maturity = "runtime-ready"
	case score >= 40:
		maturity = "integrating"
	case score >= 20:
		maturity = "discovered"
	default:
		maturity = "planned"
	}

	return score, maturity, confidence
}

func recommendations(ri RepoIntelligence) []string {
	var rec []string

	if !sliceutil.Contains(ri.DetectedFiles, ".aift/repo.json") {
		rec = append(rec, "Add `.aift/repo.json` manifest.")
	}
	if !sliceutil.Contains(ri.DetectedFiles, ".aift/capabilities.json") {
		rec = append(rec, "Run `aift capabilities scan` to generate truthful capability state.")
	}
	if !sliceutil.Contains(ri.Ready, "verify") && !sliceutil.Contains(ri.V1, "verify") {
		rec = append(rec, "Add `.aift/commands/verify.sh` before enabling orchestration.")
	}
	if !sliceutil.Contains(ri.Ready, "build") && !sliceutil.Contains(ri.V1, "build") && sliceutil.Contains(ri.Detected, "build") {
		rec = append(rec, "Convert detected build support into `.aift/commands/build.sh`.")
	}
	if !sliceutil.Contains(ri.Ready, "test") && !sliceutil.Contains(ri.V1, "test") && sliceutil.Contains(ri.Detected, "test") {
		rec = append(rec, "Convert detected test support into `.aift/commands/test.sh`.")
	}
	if len(ri.Broken) > 0 {
		rec = append(rec, "Fix broken capabilities before promotion or orchestration.")
	}
	if ri.Dirty {
		rec = append(rec, "Review and commit or intentionally preserve local changes.")
	}

	return rec
}

func writeRegistry(cfg config.Config, fi FederationIntelligence) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "intelligence.json"), fi, true)
}

func writeReport(cfg config.Config, fi FederationIntelligence) error {
	out := filepath.Join(cfg.OSHome, "reports", "intelligence.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Federation Intelligence Report\n\n")
	b.WriteString(fmt.Sprintf("- Repositories: %d\n", fi.Summary.Repos))
	b.WriteString(fmt.Sprintf("- Dirty repositories: %d\n", fi.Summary.Dirty))
	b.WriteString(fmt.Sprintf("- Repos with ready/v1 capabilities: %d\n", fi.Summary.ReadyOrV1Repos))
	b.WriteString(fmt.Sprintf("- Repos with broken capabilities: %d\n", fi.Summary.BrokenRepos))
	b.WriteString(fmt.Sprintf("- Average maturity score: %d\n\n", fi.Summary.AverageScore))

	for _, r := range fi.Repos {
		b.WriteString("## " + r.Name + "\n\n")
		b.WriteString(fmt.Sprintf("- Role: `%s`\n", r.Role))
		b.WriteString(fmt.Sprintf("- Maturity: `%s`\n", r.Maturity))
		b.WriteString(fmt.Sprintf("- Score: `%d`\n", r.Score))
		b.WriteString(fmt.Sprintf("- Confidence: `%d`\n", r.Confidence))
		b.WriteString(fmt.Sprintf("- Dirty: `%v`\n", r.Dirty))
		b.WriteString(fmt.Sprintf("- Languages: `%s`\n", strings.Join(r.Languages, ", ")))
		b.WriteString(fmt.Sprintf("- Frameworks: `%s`\n", strings.Join(r.Frameworks, ", ")))
		b.WriteString(fmt.Sprintf("- Ready: `%s`\n", strings.Join(r.Ready, ", ")))
		b.WriteString(fmt.Sprintf("- V1: `%s`\n", strings.Join(r.V1, ", ")))
		b.WriteString(fmt.Sprintf("- Detected: `%s`\n", strings.Join(r.Detected, ", ")))
		b.WriteString(fmt.Sprintf("- Planned: `%s`\n", strings.Join(r.Planned, ", ")))
		b.WriteString(fmt.Sprintf("- Broken: `%s`\n\n", strings.Join(r.Broken, ", ")))
		if len(r.Recommendations) > 0 {
			b.WriteString("Recommendations:\n")
			for _, rec := range r.Recommendations {
				b.WriteString("- " + rec + "\n")
			}
			b.WriteString("\n")
		}
	}

	if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
		return err
	}
	fmt.Println("Wrote", out)
	return nil
}

func printRepo(ri RepoIntelligence) {
	fmt.Println("Repository:", ri.Name)
	fmt.Println("Role:", ri.Role)
	fmt.Println("Maturity:", ri.Maturity)
	fmt.Println("Score:", ri.Score)
	fmt.Println("Confidence:", ri.Confidence)
	fmt.Println("Dirty:", ri.Dirty)
	fmt.Println("Languages:", strings.Join(ri.Languages, ", "))
	fmt.Println("Frameworks:", strings.Join(ri.Frameworks, ", "))
	fmt.Println("Ready:", strings.Join(ri.Ready, ", "))
	fmt.Println("V1:", strings.Join(ri.V1, ", "))
	fmt.Println("Detected:", strings.Join(ri.Detected, ", "))
	fmt.Println("Planned:", strings.Join(ri.Planned, ", "))
	fmt.Println("Broken:", strings.Join(ri.Broken, ", "))
	if len(ri.Recommendations) > 0 {
		fmt.Println("Recommendations:")
		for _, rec := range ri.Recommendations {
			fmt.Println("-", rec)
		}
	}
}


