#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"

git checkout phase7-doctor-git-housekeeping

mkdir -p internal/doctor

cat > internal/doctor/doctor_test.go <<'GO'
package doctor

import (
"os"
"path/filepath"
"testing"

"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

func TestDoctorRunDoesNotPanic(t *testing.T) {
root := t.TempDir()
osDir := filepath.Join(root, "AIFT-OS")

if err := os.MkdirAll(osDir, 0755); err != nil {
t.Fatal(err)
}

cfg := config.Config{
Root: root,
OS:   osDir,
}

if err := Run(cfg); err != nil {
t.Fatalf("Run returned error: %v", err)
}
}

func TestDoctorRepairDoesNotPanic(t *testing.T) {
root := t.TempDir()
osDir := filepath.Join(root, "AIFT-OS")

if err := os.MkdirAll(filepath.Join(osDir, "registry"), 0755); err != nil {
t.Fatal(err)
}

cfg := config.Config{
Root: root,
OS:   osDir,
}

if err := Repair(cfg); err != nil {
t.Fatalf("Repair returned error: %v", err)
}
}

func TestDoctorGitDoesNotPanic(t *testing.T) {
root := t.TempDir()
osDir := filepath.Join(root, "AIFT-OS")

if err := os.MkdirAll(osDir, 0755); err != nil {
t.Fatal(err)
}

cfg := config.Config{
Root: root,
OS:   osDir,
}

if err := Git(cfg); err != nil {
t.Fatalf("Git returned error: %v", err)
}
}

func TestDoctorFullDoesNotPanic(t *testing.T) {
root := t.TempDir()
osDir := filepath.Join(root, "AIFT-OS")

if err := os.MkdirAll(filepath.Join(osDir, "registry"), 0755); err != nil {
t.Fatal(err)
}

cfg := config.Config{
Root: root,
OS:   osDir,
}

if err := Full(cfg); err != nil {
t.Fatalf("Full returned error: %v", err)
}
}
GO

gofmt -w internal/doctor/doctor_test.go

go test ./...
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r
aift doctor full

git add internal/doctor/doctor_test.go
git commit -m "test: add doctor coverage"
git push origin phase7-doctor-git-housekeeping

echo "Done. Re-run CI on PR 20."
