package manual

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/events"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/workspace"
)

type Status struct {
	ManualSource string `json:"manualSource"`
	PDFBuild     string `json:"pdfBuild"`
	WebPublish   string `json:"webPublish"`
}

type Builder struct {
	Repo       string `json:"repo"`
	Capability string `json:"capability"`
	Status     string `json:"status"`
}

type Contract struct {
	Repo        string   `json:"repo"`
	Title       string   `json:"title"`
	SourcePath  string   `json:"sourcePath"`
	AssetsPath  string   `json:"assetsPath"`
	Builder     Builder  `json:"builder"`
	Status      Status   `json:"status"`
	ManualType  string   `json:"manualType"`
	Format      string   `json:"format"`
	Sections    []string `json:"sections"`
	GeneratedAt string   `json:"generatedAt"`
}

type Registry struct {
	GeneratedAt string     `json:"generatedAt"`
	Manuals     []Contract `json:"manuals"`
}

func InitAll(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	for _, r := range repos {
		if err := InitRepo(cfg, r); err != nil {
			return err
		}
	}

	return Scan(cfg)
}

func InitRepo(cfg config.Config, r workspace.Repo) error {
	base := filepath.Join(r.Path, "docs", "manual")
	source := filepath.Join(base, "source")
	assets := filepath.Join(base, "assets")

	dirs := []string{
		filepath.Join(source, "man0"),
		filepath.Join(source, "man1"),
		filepath.Join(source, "man2"),
		filepath.Join(source, "man3"),
		filepath.Join(source, "man4"),
		filepath.Join(source, "man5"),
		filepath.Join(source, "man6"),
		filepath.Join(source, "man7"),
		filepath.Join(source, "man8"),
		filepath.Join(source, "man9"),
		assets,
		filepath.Join(r.Path, ".aift"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	writeIfMissing(filepath.Join(base, "README.md"), manualReadme(r.Name))
	writeIfMissing(filepath.Join(source, "index.md"), manualIndex(r.Name))
	writeIfMissing(filepath.Join(source, "man0", "00-introduction.md"), introPage(r.Name))
	writeIfMissing(filepath.Join(source, "man7", "modularity.md"), modularityPage(r.Name))
	writeIfMissing(filepath.Join(source, "man7", "truthfulness.md"), truthfulnessPage(r.Name))
	writeIfMissing(filepath.Join(source, "man7", "booksmith-pipeline.md"), booksmithPage(r.Name))

	contract := BuildContract(cfg, r)
	data, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(r.Path, ".aift", "manual.json"), append(data, '\n'), 0644)
}

func Scan(cfg config.Config) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}

	reg := Registry{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Manuals:     []Contract{},
	}

	for _, r := range repos {
		reg.Manuals = append(reg.Manuals, BuildContract(cfg, r))
	}

	if err := writeRegistry(cfg, reg); err != nil {
		return err
	}
	if err := writeReport(cfg, reg); err != nil {
		return err
	}

	return events.Emit(cfg, "manual.scan", "manual", "federation manual scan complete", map[string]string{
		"manuals": fmt.Sprint(len(reg.Manuals)),
	})
}

func Repo(cfg config.Config, name string) error {
	repos, err := workspace.FindRepos(cfg)
	if err != nil {
		return err
	}
	for _, r := range repos {
		if r.Name == name {
			c := BuildContract(cfg, r)
			fmt.Println("Repository:", c.Repo)
			fmt.Println("Title:", c.Title)
			fmt.Println("Source:", c.SourcePath)
			fmt.Println("Assets:", c.AssetsPath)
			fmt.Println("Builder:", c.Builder.Repo)
			fmt.Println("Builder capability:", c.Builder.Capability)
			fmt.Println("Builder status:", c.Builder.Status)
			fmt.Println("Manual source:", c.Status.ManualSource)
			fmt.Println("PDF build:", c.Status.PDFBuild)
			fmt.Println("Web publish:", c.Status.WebPublish)
			return nil
		}
	}
	return fmt.Errorf("repository not found: %s", name)
}

func Report(cfg config.Config) error {
	path := filepath.Join(cfg.OSHome, "reports", "manuals.md")
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

func BuildContract(cfg config.Config, r workspace.Repo) Contract {
	sourceRel := "docs/manual/source"
	assetsRel := "docs/manual/assets"

	sourceStatus := "planned"
	if dirExists(filepath.Join(r.Path, sourceRel)) {
		sourceStatus = "ready"
	}

	builderStatus := "planned"
	booksmithPath := filepath.Join(cfg.Root, "booksmith-ai")
	if !dirExists(booksmithPath) {
		booksmithPath = filepath.Join(cfg.Root, "BookSmith-Federation-OS")
	}
	if fileExists(filepath.Join(booksmithPath, ".aift", "commands", "manual-build.sh")) {
		builderStatus = "ready"
	}

	return Contract{
		Repo:       r.Name,
		Title:      titleFor(r.Name),
		SourcePath: sourceRel,
		AssetsPath: assetsRel,
		Builder: Builder{
			Repo:       "booksmith-ai",
			Capability: "manual.build.pdf",
			Status:     builderStatus,
		},
		Status: Status{
			ManualSource: sourceStatus,
			PDFBuild:     builderStatus,
			WebPublish:   "planned",
		},
		ManualType:  "unix-style-federation-manual",
		Format:      "markdown-source-booksmith-built-pdf",
		Sections:    []string{"man0", "man1", "man2", "man3", "man4", "man5", "man6", "man7", "man8", "man9"},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

func writeRegistry(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "registry", "manuals.json")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(reg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(out, append(data, '\n'), 0644); err != nil {
		return err
	}
	fmt.Println("Wrote", out)
	return nil
}

func writeReport(cfg config.Config, reg Registry) error {
	out := filepath.Join(cfg.OSHome, "reports", "manuals.md")
	if err := os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}

	var b strings.Builder
	b.WriteString("# Federation Manual Contracts\n\n")
	b.WriteString("Every repo owns manual source. BookSmith owns PDF/build/publishing. AIFT-OS owns discovery, truth, contracts, and reports.\n\n")
	b.WriteString("| Repository | Source | PDF Build | Web Publish | Builder |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, m := range reg.Manuals {
		b.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | `%s` | `%s:%s` |\n",
			m.Repo, m.Status.ManualSource, m.Status.PDFBuild, m.Status.WebPublish, m.Builder.Repo, m.Builder.Capability))
	}

	if err := os.WriteFile(out, []byte(b.String()), 0644); err != nil {
		return err
	}
	fmt.Println("Wrote", out)
	return nil
}

func writeIfMissing(path string, content string) error {
	if fileExists(path) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func manualReadme(repo string) string {
	return "# " + titleFor(repo) + " Manual\n\nThis directory contains the UNIX-style manual source for this repository.\n\nManual source belongs to this repo. PDF and publication builds belong to BookSmith.\n\n"
}

func manualIndex(repo string) string {
	return "# " + titleFor(repo) + " Manual Index\n\n- `man0/` — Doctrine and introduction\n- `man1/` — User commands\n- `man2/` — System calls and internal operations\n- `man3/` — Libraries and APIs\n- `man4/` — Devices and providers\n- `man5/` — File formats\n- `man6/` — Federation applications\n- `man7/` — Concepts, doctrines, standards\n- `man8/` — Administration\n- `man9/` — Kernel/internal interfaces\n\n"
}

func introPage(repo string) string {
	return "# " + titleFor(repo) + " Introduction\n\n## NAME\n\n" + repo + " manual introduction\n\n## DESCRIPTION\n\nThis manual page defines the role, purpose, and federation contract for this repository.\n\n## STATUS\n\nManual source: ready\n\nPDF build: planned until BookSmith exposes `manual.build.pdf` as a verified capability.\n\n"
}

func modularityPage(repo string) string {
	return "# Modularity Doctrine\n\n## NAME\n\nmodularity — replaceable modules and provider-agnostic architecture\n\n## DESCRIPTION\n\nNo dependency is sacred. Only the contract is sacred.\n\nThis repository should declare modules, providers, capabilities, events, and replacement paths through `.aift/` contracts.\n\n"
}

func truthfulnessPage(repo string) string {
	return "# Truthfulness Doctrine\n\n## NAME\n\ntruthfulness — evidence before orchestration\n\n## DESCRIPTION\n\nAIFT-OS must never claim this repository can perform an action until that capability is verified.\n\nCapabilities move through planned, detected, ready, v1, broken, and deprecated states.\n\n"
}

func booksmithPage(repo string) string {
	return "# BookSmith Manual Pipeline\n\n## NAME\n\nbooksmith-pipeline — federation manual PDF and publication pipeline\n\n## DESCRIPTION\n\nThis repository owns its manual source. BookSmith owns compilation, PDF generation, proofing, publishing packet generation, and static web export.\n\n## SOURCE\n\n`docs/manual/source/`\n\n## CONTRACT\n\n`.aift/manual.json`\n\n## STATUS\n\nBookSmith PDF build remains planned until a verified `.aift/commands/manual-build.sh` capability exists in BookSmith.\n\n"
}

func titleFor(repo string) string {
	return strings.ReplaceAll(repo, "-", " ") + " UNIX Manual"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
