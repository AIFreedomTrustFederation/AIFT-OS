package intelligence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyRole(t *testing.T) {
	tests := []struct {
		name string
		ri   RepoIntelligence
		want string
	}{
		{"aift-os", RepoIntelligence{}, "federation-control-plane"},
		{"AIFT-OS", RepoIntelligence{}, "federation-control-plane"},
		{"my-forge-app", RepoIntelligence{}, "forge-repository-platform"},
		{"booksmith-ai", RepoIntelligence{}, "authoring-publishing-system"},
		{"freedom-trust-docs", RepoIntelligence{}, "doctrine-governance-trust"},
		{"aether-coin", RepoIntelligence{}, "economic-trust-layer"},
		{"my-vps-node", RepoIntelligence{}, "infrastructure-node-layer"},
		{"www-portal", RepoIntelligence{}, "public-web-portal"},
		{"my-github.io", RepoIntelligence{}, "public-web-portal"},
		{"webapp", RepoIntelligence{Frameworks: []string{"Next.js"}}, "web-application"},
		{"docs-repo", RepoIntelligence{Languages: []string{"Markdown/Docs"}}, "documentation-repository"},
		{"random-repo", RepoIntelligence{}, "unknown-sovereign-repository"},
	}
	for _, tt := range tests {
		got := classifyRole(tt.name, tt.ri)
		if got != tt.want {
			t.Errorf("classifyRole(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		name     string
		ri       RepoIntelligence
		wantMin  int
		wantMax  int
		maturity string
	}{
		{
			name:     "empty repo",
			ri:       RepoIntelligence{},
			wantMin:  0,
			wantMax:  0,
			maturity: "planned",
		},
		{
			name: "repo with README only",
			ri: RepoIntelligence{
				DetectedFiles: []string{"README.md"},
			},
			wantMin:  10,
			wantMax:  10,
			maturity: "planned",
		},
		{
			name: "repo with aift manifests",
			ri: RepoIntelligence{
				DetectedFiles: []string{"README.md", ".aift/repo.json", ".aift/capabilities.json"},
			},
			wantMin:  30,
			wantMax:  30,
			maturity: "discovered",
		},
		{
			name: "repo with ready capabilities",
			ri: RepoIntelligence{
				DetectedFiles: []string{"README.md", ".aift/repo.json", ".aift/capabilities.json"},
				Ready:         []string{"verify", "build"},
			},
			wantMin:  46,
			wantMax:  46,
			maturity: "integrating",
		},
		{
			name: "mature repo with v1 capabilities",
			ri: RepoIntelligence{
				DetectedFiles: []string{"README.md", ".aift/repo.json", ".aift/capabilities.json"},
				Ready:         []string{"verify"},
				V1:            []string{"build", "test", "status"},
				Frameworks:    []string{"Go CLI/Service"},
				Remote:        "origin",
			},
			wantMin:  80,
			wantMax:  100,
			maturity: "orchestratable",
		},
		{
			name: "broken repo penalized",
			ri: RepoIntelligence{
				DetectedFiles: []string{"README.md"},
				Broken:        []string{"build", "test"},
			},
			wantMin:  0,
			wantMax:  0,
			maturity: "planned",
		},
		{
			name: "dirty repo penalized",
			ri: RepoIntelligence{
				DetectedFiles: []string{"README.md", ".aift/repo.json"},
				Dirty:         true,
			},
			wantMin:  15,
			wantMax:  15,
			maturity: "planned",
		},
	}
	for _, tt := range tests {
		s, maturity, _ := score(tt.ri)
		if s < tt.wantMin || s > tt.wantMax {
			t.Errorf("%s: score = %d, want [%d, %d]", tt.name, s, tt.wantMin, tt.wantMax)
		}
		if maturity != tt.maturity {
			t.Errorf("%s: maturity = %q, want %q", tt.name, maturity, tt.maturity)
		}
	}
}

func TestScoreClampedTo100(t *testing.T) {
	ri := RepoIntelligence{
		DetectedFiles: []string{"README.md", ".aift/repo.json", ".aift/capabilities.json"},
		Ready:         []string{"a", "b", "c", "d", "e"},
		V1:            []string{"f", "g", "h", "i", "j"},
		Frameworks:    []string{"Go CLI/Service"},
		Remote:        "origin",
	}
	s, _, _ := score(ri)
	if s > 100 {
		t.Errorf("score should be clamped to 100, got %d", s)
	}
}

func TestScoreConfidenceCap(t *testing.T) {
	ri := RepoIntelligence{
		DetectedFiles: []string{"README.md", ".aift/repo.json", ".aift/capabilities.json"},
		Ready:         []string{"a", "b"},
		V1:            []string{"c", "d"},
		Broken:        []string{"e"},
	}
	_, _, conf := score(ri)
	if conf > 100 {
		t.Errorf("confidence should be capped at 100, got %d", conf)
	}
}

func TestRecommendations(t *testing.T) {
	ri := RepoIntelligence{
		DetectedFiles: []string{"README.md"},
		Detected:      []string{"build", "test"},
		Dirty:         true,
	}
	recs := recommendations(ri)

	expectContains := []string{
		"Add `.aift/repo.json` manifest",
		"Run `aift capabilities scan`",
		"Add `.aift/commands/verify.sh`",
		"Convert detected build support",
		"Convert detected test support",
		"Review and commit",
	}
	for _, want := range expectContains {
		found := false
		for _, rec := range recs {
			if containsStr(rec, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("recommendations missing: %q", want)
		}
	}
}

func TestRecommendationsNoVerifyIfReady(t *testing.T) {
	ri := RepoIntelligence{
		DetectedFiles: []string{"README.md", ".aift/repo.json", ".aift/capabilities.json"},
		Ready:         []string{"verify"},
	}
	recs := recommendations(ri)
	for _, rec := range recs {
		if containsStr(rec, "verify.sh") {
			t.Error("should not recommend verify.sh when verify is already ready")
		}
	}
}

func TestRecommendationsBroken(t *testing.T) {
	ri := RepoIntelligence{
		DetectedFiles: []string{"README.md", ".aift/repo.json", ".aift/capabilities.json"},
		Ready:         []string{"verify"},
		Broken:        []string{"build"},
	}
	recs := recommendations(ri)
	found := false
	for _, rec := range recs {
		if containsStr(rec, "Fix broken") {
			found = true
			break
		}
	}
	if !found {
		t.Error("should recommend fixing broken capabilities")
	}
}

func TestDetectLanguages(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# hi"), 0644)

	langs := detectLanguages(dir)
	if !contains(langs, "Go") {
		t.Error("expected Go language")
	}
	if !contains(langs, "TypeScript/JavaScript") {
		t.Error("expected TypeScript/JavaScript language")
	}
	if !contains(langs, "Markdown/Docs") {
		t.Error("expected Markdown/Docs language")
	}
}

func TestDetectLanguagesEmpty(t *testing.T) {
	dir := t.TempDir()
	langs := detectLanguages(dir)
	if len(langs) != 0 {
		t.Errorf("expected no languages in empty dir, got %v", langs)
	}
}

func TestDetectFrameworks(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "next.config.js"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "tailwind.config.js"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine"), 0644)

	frameworks := detectFrameworks(dir)
	expected := []string{"Go CLI/Service", "Next.js", "Tailwind", "Container"}
	for _, want := range expected {
		if !contains(frameworks, want) {
			t.Errorf("expected framework %q, got %v", want, frameworks)
		}
	}
}

func TestDetectFrameworksEmpty(t *testing.T) {
	dir := t.TempDir()
	frameworks := detectFrameworks(dir)
	if len(frameworks) != 0 {
		t.Errorf("expected no frameworks in empty dir, got %v", frameworks)
	}
}

func TestDetectFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# hi"), 0644)
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "Makefile"), []byte("all:"), 0644)

	files := detectFiles(dir)
	if !contains(files, "README.md") {
		t.Error("expected README.md")
	}
	if !contains(files, "go.mod") {
		t.Error("expected go.mod")
	}
	if !contains(files, "Makefile") {
		t.Error("expected Makefile")
	}
}

func TestContains(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	if !contains(items, "banana") {
		t.Error("contains should find banana")
	}
	if contains(items, "grape") {
		t.Error("contains should not find grape")
	}
	if contains(nil, "anything") {
		t.Error("contains should return false for nil slice")
	}
}

func TestKeys(t *testing.T) {
	m := map[string]bool{"b": true, "a": true, "c": true}
	got := keys(m)
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("keys returned %d items, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("keys[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestKeysEmpty(t *testing.T) {
	m := map[string]bool{}
	got := keys(m)
	if len(got) != 0 {
		t.Errorf("keys(empty) returned %d items, want 0", len(got))
	}
}

func TestLoadCapabilities(t *testing.T) {
	dir := t.TempDir()
	aiftDir := filepath.Join(dir, ".aift")
	os.MkdirAll(aiftDir, 0755)
	os.WriteFile(filepath.Join(aiftDir, "capabilities.json"), []byte(`{
		"capabilities": [
			{"name": "build", "status": "ready"},
			{"name": "test", "status": "v1"},
			{"name": "deploy", "status": "planned"},
			{"name": "health", "status": "detected"},
			{"name": "sync", "status": "broken"}
		]
	}`), 0644)

	summary, ready, v1, detected, planned, broken := loadCapabilities(dir)

	if len(summary) != 5 {
		t.Errorf("summary should have 5 entries, got %d", len(summary))
	}
	if len(ready) != 1 || ready[0] != "build" {
		t.Errorf("ready = %v, want [build]", ready)
	}
	if len(v1) != 1 || v1[0] != "test" {
		t.Errorf("v1 = %v, want [test]", v1)
	}
	if len(planned) != 1 || planned[0] != "deploy" {
		t.Errorf("planned = %v, want [deploy]", planned)
	}
	if len(detected) != 1 || detected[0] != "health" {
		t.Errorf("detected = %v, want [health]", detected)
	}
	if len(broken) != 1 || broken[0] != "sync" {
		t.Errorf("broken = %v, want [sync]", broken)
	}
}

func TestLoadCapabilitiesMissingFile(t *testing.T) {
	dir := t.TempDir()
	summary, ready, v1, detected, planned, broken := loadCapabilities(dir)
	if len(summary) != 0 || len(ready) != 0 || len(v1) != 0 || len(detected) != 0 || len(planned) != 0 || len(broken) != 0 {
		t.Error("missing file should return empty results")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
