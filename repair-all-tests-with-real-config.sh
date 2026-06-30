#!/data/data/com.termux/files/usr/bin/bash
set -u

OS="${AIFT_OS:-$HOME/AIFT/AIFT-OS}"
cd "$OS" || exit 1

echo "Repairing all tests that guessed config.Config fields"

mkdir -p repair
mkdir -p reports

cat > repair/repair-go-tests-real-config.py <<'PY'
from pathlib import Path
import re

root = Path.cwd()
test_files = sorted(root.glob("**/*_test.go"))

bad_field_patterns = [
    r"\bconfig\.Config\s*\{\s*OS\s*:",
    r"\bconfig\.Config\s*\{\s*Root\s*:",
    r"\bconfig\.Config\s*\{[^}]*\bOS\s*:",
    r"\bconfig\.Config\s*\{[^}]*\bRoot\s*:",
]

changed = []

for path in test_files:
    if any(part in {"node_modules", ".git", "dist", "build", ".next", "vendor"} for part in path.parts):
        continue

    text = path.read_text()

    original = text

    text = re.sub(
        r"cfg\s*:=\s*config\.Config\s*\{[^}]*\}",
        "cfg := config.Load()",
        text,
        flags=re.DOTALL,
    )

    text = re.sub(
        r"config\.Config\s*\{[^}]*\}",
        "config.Load()",
        text,
        flags=re.DOTALL,
    )

    if text != original:
        path.write_text(text)
        changed.append(str(path))

report = root / "reports" / "repair-all-tests-real-config.txt"
report.write_text("\n".join(changed) + "\n")
print("Changed files:")
for item in changed:
    print(item)
PY

python repair/repair-go-tests-real-config.py

echo "Formatting Go files"
gofmt -w $(find . -name '*.go' -not -path './.git/*' -not -path './node_modules/*')

echo "Running tests"
go test ./...

echo "Building native CLI"
go build -o "$HOME/.local/bin/aift" ./cmd/aift
hash -r

echo "Running AIFT doctor"
aift doctor || true

echo "Running AIFT verify"
aift verify

echo "Git status"
git status --short

echo "Done"
