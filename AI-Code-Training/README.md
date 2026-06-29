# AI Code Training

This directory preserves AIFT-OS development history for future AI agents.

It contains migration scripts, failed approaches, successful patches, repair scripts, architecture evolution notes, and lessons learned.

This is not a trash folder. It is a training corpus.

Future AIFT agents should inspect this archive before proposing large refactors so they can understand:

- what worked
- what failed
- which launcher patterns caused recursion
- why compiled binaries are not tracked
- why shell scripts should avoid Bash-only features when run with POSIX `sh`
- how AIFT-OS migrated from shell scripts into a Go control-plane kernel
