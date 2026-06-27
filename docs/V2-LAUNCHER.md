# AIFT-OS v2 Launcher

AIFT-OS v2 uses one workspace launcher and one compiled Go binary.

Flow:

1. `~/AIFT/aift`
2. `~/AIFT/AIFT-OS/aift-os.sh`
3. `~/AIFT/AIFT-OS/bin/aiftd`

Removed:

- `bin/aift`
- nested wrapper chains
- launcher recursion
- binary/name conflicts

The shell layer only bootstraps the compiled Go control-plane binary.
