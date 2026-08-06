package uxi

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeIntegrationRepo(t *testing.T, root, name, sourceID, sourcePath, status string) string {
	t.Helper()
	repo := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(repo, ".aift"), 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := `{"schema":"aift.uxi.integration.v1","repository":"` + name + `","sources":[{"id":"` + sourceID + `","kind":"json","path":"` + sourcePath + `","status":"` + status + `"}]}`
	if err := os.WriteFile(filepath.Join(repo, ".aift", "uxi.json"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestDiscoverAndReadIntegrationSource(t *testing.T) {
	root := t.TempDir()
	repo := writeIntegrationRepo(t, root, "booksmith-ai", "booksmith.library", "library/book-registry.json", "ready")
	if err := os.MkdirAll(filepath.Join(repo, "library"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "library", "book-registry.json"), []byte(`{"books":[{"slug":"one"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	sources, evidence, err := DiscoverIntegrationSources(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 || len(evidence) != 1 {
		t.Fatalf("sources=%#v evidence=%#v", sources, evidence)
	}
	data, sourceEvidence, err := readIntegrationJSON(sources[0])
	if err != nil {
		t.Fatal(err)
	}
	if data["sha256"] == "" || len(sourceEvidence) != 2 {
		t.Fatalf("data=%#v evidence=%#v", data, sourceEvidence)
	}
}

func TestIntegrationRejectsTraversalAndSymlinkEscape(t *testing.T) {
	root := t.TempDir()
	_ = writeIntegrationRepo(t, root, "unsafe", "unsafe.source", "../outside.json", "ready")
	sources, evidence, err := DiscoverIntegrationSources(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 0 || len(evidence) != 1 || evidence[0].Status != "failed" {
		t.Fatalf("sources=%#v evidence=%#v", sources, evidence)
	}

	root2 := t.TempDir()
	repo2 := writeIntegrationRepo(t, root2, "linked", "linked.source", "data.json", "ready")
	outside := filepath.Join(root2, "outside.json")
	if err := os.WriteFile(outside, []byte(`{"secret":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(repo2, "data.json")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	sources, _, err = DiscoverIntegrationSources(root2)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := readIntegrationJSON(sources[0]); err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}

func TestSourceInspectAdapterProducesArtifact(t *testing.T) {
	root := t.TempDir()
	repo := writeIntegrationRepo(t, root, "TheMindofAll", "models.local", "models/manifest.json", "ready")
	if err := os.MkdirAll(filepath.Join(repo, "models"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "models", "manifest.json"), []byte(`{"models":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	store, _ := NewStore(t.TempDir())
	engine, err := NewEngine(store, root, nil)
	if err != nil {
		t.Fatal(err)
	}
	session, _ := store.CreateSession("source")
	session, err = store.ProposeAction(session.ID, Action{Kind: "source.inspect", Target: "models.local", Risk: "low"})
	if err != nil {
		t.Fatal(err)
	}
	session, err = engine.InvokeAction(context.Background(), session.ID, session.Actions[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if session.Jobs[0].Status != StatusSucceeded || len(session.Artifacts) != 1 {
		t.Fatalf("session=%#v", session)
	}
}
