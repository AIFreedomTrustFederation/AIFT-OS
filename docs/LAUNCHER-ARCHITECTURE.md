# AIFT-OS Launcher Architecture

AIFT-OS launcher flow:

1. ~/AIFT/aift
2. ~/AIFT/AIFT-OS/bin/aift
3. ~/AIFT/AIFT-OS/aift-os.sh
4. ~/AIFT/AIFT-OS/bin/aiftd

Files:

- ~/AIFT/aift is the workspace launcher.
- bin/aift is the repository shell launcher.
- aift-os.sh is the bootstrap compatibility launcher.
- bin/aiftd is the compiled Go control-plane executable.

The Go binary is named aiftd so it does not conflict with the aift shell launcher.
