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
