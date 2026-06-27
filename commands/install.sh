#!/usr/bin/env sh
set -eu
mkdir -p "$AIFT_ROOT"
cat > "$AIFT_ROOT/aift-os.sh" <<LAUNCH
#!/usr/bin/env sh
exec "\$HOME/AIFT/AIFT-OS/aift-os.sh" "\$@"
LAUNCH
cat > "$AIFT_ROOT/aift" <<LAUNCH
#!/usr/bin/env sh
exec "\$HOME/AIFT/AIFT-OS/bin/aift" "\$@"
LAUNCH
chmod +x "$AIFT_ROOT/aift-os.sh" "$AIFT_ROOT/aift" "$AIFT_OS_HOME/aift-os.sh" "$AIFT_OS_HOME/bin/aift"
echo "Installed launchers:"
echo "  $AIFT_ROOT/aift"
echo "  $AIFT_ROOT/aift-os.sh"
