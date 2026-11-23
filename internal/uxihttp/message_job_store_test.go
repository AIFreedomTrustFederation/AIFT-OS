package uxihttp

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMessageJobStorePersistsAcrossReopen(t *testing.T) {
	root := t.TempDir()

	first, err := newMessageJobStore(root)
	if err != nil {
		t.Fatal(err)
	}

	original := messageJob{
		ID:        "0123456789abcdef0123456789abcdef",
		SessionID: "session-test",
		Status:    "completed",
		CreatedAt: time.Now().UTC(),
	}

	if err := first.save(original); err != nil {
		t.Fatal(err)
	}

	second, err := newMessageJobStore(root)
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(second.dir, original.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}

	recovered, err := second.load(original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.ID != original.ID ||
		recovered.SessionID != original.SessionID ||
		recovered.Status != original.Status ||
		!recovered.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf("recovered job does not match original: %+v", recovered)
	}

	if len(data) == 0 {
		t.Fatal("persisted job is empty")
	}

	info, err := os.Stat(filepath.Join(second.dir, original.ID+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("job permissions = %o, want 600", info.Mode().Perm())
	}
}

func TestMessageJobStoreRejectsInvalidID(t *testing.T) {
	store, err := newMessageJobStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	if err := store.save(messageJob{
		ID:     "../outside",
		Status: "running",
	}); err == nil {
		t.Fatal("expected invalid job ID to be rejected")
	}
}
