#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail
cd "${AIFT_ROOT:-$HOME/AIFT}/AIFT-OS"

python <<'PY'
from pathlib import Path

p = Path("cmd/aift/main.go")
s = p.read_text()

imp = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/doctor"'
if imp not in s:
    marker = '"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"'
    s = s.replace(marker, marker + "\n\t" + imp)

if 'case "doctor":' in s and 'doctor.Git' not in s:
    old = '''case "doctor":
\t\terr = doctor.Run(cfg)'''
    new = '''case "doctor":
\t\tif len(args) > 0 && args[0] == "repair" {
\t\t\terr = doctor.Repair(cfg)
\t\t} else if len(args) > 0 && args[0] == "git" {
\t\t\terr = doctor.Git(cfg)
\t\t} else if len(args) > 0 && args[0] == "full" {
\t\t\terr = doctor.Full(cfg)
\t\t} else {
\t\t\terr = doctor.Run(cfg)
\t\t}'''
    s = s.replace(old, new)

if 'doctor [repair|git|full]' not in s:
    s = s.replace('fmt.Println("  doctor")', 'fmt.Println("  doctor [repair|git|full]")')

p.write_text(s)
PY
