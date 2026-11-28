#!/usr/bin/env bash
# Disposable x86_64 UEFI/OVMF M1 smoke proof. Does not touch host NVRAM or disks.
# no-harness: isolated virtual boot fixture writes only operator-selected artifacts and temporary OVMF VARS; no federation or repository mutation.
set -euo pipefail
export LC_ALL=C
for tool in grub-mkrescue cpio gzip qemu-system-x86_64 sha256sum timeout grep find sort file mktemp tr; do
  command -v "$tool" >/dev/null || { echo "MISSING_DEPENDENCY: $tool" >&2; exit 2; }
done
: "${AIFT_M1_KERNEL:?Set AIFT_M1_KERNEL to a locally installed x86_64 Linux bzImage}"
: "${AIFT_M1_BUSYBOX:?Set AIFT_M1_BUSYBOX to a statically linked x86_64 busybox binary}"
: "${AIFT_M1_OVMF_CODE:?Set AIFT_M1_OVMF_CODE to a read-only OVMF CODE fd image}"
: "${AIFT_M1_OVMF_VARS:?Set AIFT_M1_OVMF_VARS to an OVMF VARS template}"
for f in "$AIFT_M1_KERNEL" "$AIFT_M1_BUSYBOX" "$AIFT_M1_OVMF_CODE" "$AIFT_M1_OVMF_VARS"; do
  test -f "$f" || { echo "MISSING_FILE: $f" >&2; exit 2; }
done
if ! file "$AIFT_M1_BUSYBOX" | grep -q 'statically linked'; then
  echo "AIFT_M1_BUSYBOX must be statically linked" >&2; exit 2
fi
output="${AIFT_M1_OUTPUT:-$PWD/artifacts/m1-uefi}"
mkdir -p "$output"
output="$(cd "$output" && pwd)"
# Remove stale proof files so a failed boot cannot reuse a previous successful marker.
rm -f "$output/serial.log" "$output/qemu.log" "$output/SHA256SUMS"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/initramfs/bin" "$work/iso/boot/grub" "$work/iso/boot" "$work/initramfs"/{proc,sys,dev}
cp "$AIFT_M1_BUSYBOX" "$work/initramfs/bin/busybox"
cat > "$work/initramfs/init" <<'INIT'
#!/bin/busybox sh
/bin/busybox mount -t proc proc /proc
/bin/busybox mount -t sysfs sysfs /sys
echo AIFT_M1_BOOT_OK >/dev/console
echo AIFT_M1_BOOT_OK
/bin/busybox sync
/bin/busybox poweroff -f
/bin/busybox sleep 2
exit 1
INIT
chmod 0755 "$work/initramfs/init" "$work/initramfs/bin/busybox"
(
  cd "$work/initramfs"
  find . -print0 | sort -z | cpio --null -o -H newc --quiet | gzip -n -9 > "$work/iso/boot/initramfs.cpio.gz"
)
cp "$AIFT_M1_KERNEL" "$work/iso/boot/vmlinuz"
cat > "$work/iso/boot/grub/grub.cfg" <<'GRUB'
set timeout=0
set default=0
terminal_output console
menuentry 'AIFT M1 Linux (virtual proof only)' {
  linux /boot/vmlinuz console=ttyS0,115200 earlyprintk=serial,ttyS0 rdinit=/init panic=10
  initrd /boot/initramfs.cpio.gz
}
GRUB
grub-mkrescue -o "$output/aift-m1-x86_64.iso" "$work/iso" >"$output/build.log" 2>&1
cp "$AIFT_M1_OVMF_VARS" "$work/OVMF_VARS.fd"
set +e
timeout --signal=TERM 90s qemu-system-x86_64 \
  -machine q35 -m 1024 -smp 1 -nodefaults -display none \
  -serial file:"$output/serial.log" -monitor none -no-reboot \
  -drive if=pflash,format=raw,readonly=on,file="$AIFT_M1_OVMF_CODE" \
  -drive if=pflash,format=raw,file="$work/OVMF_VARS.fd" \
  -cdrom "$output/aift-m1-x86_64.iso" >"$output/qemu.log" 2>&1
rc=$?
set -e
if [ "$rc" -ne 0 ] || ! (tr -d '\r' < "$output/serial.log" | grep -Fxq 'AIFT_M1_BOOT_OK'); then
  printf 'M1 FAIL: qemu_exit=%s; inspect serial.log and qemu.log\n' "$rc" >&2
  exit 1
fi
sha256sum "$output/aift-m1-x86_64.iso" > "$output/SHA256SUMS"
printf 'M1 PASS: QEMU/OVMF serial boot marker observed; artifact %s\n' "$output/aift-m1-x86_64.iso"
cat "$output/SHA256SUMS"
