# M1 — disposable x86-64 UEFI Linux boot proof (draft)

This harness creates a minimal **Linux kernel + static BusyBox initramfs** ISO, boots it using **QEMU/OVMF**, checks an actual serial-console marker and writes an ISO SHA-256 digest. It does not install AIFT services or constitute a usable desktop OS. AIFT-OS remains the federation *userspace control plane*, not a Linux kernel.

## Prerequisites (on a disposable Linux CI/VM host)

Install distribution-provided `qemu-system-x86`, `ovmf`, `grub-common`, `grub-pc-bin`, `grub-efi-amd64-bin`, `xorriso`, `mtools`, `cpio`, `gzip`, `file`, `coreutils`, and a **static x86-64 BusyBox** binary. Supply a compatible **x86-64 Linux bzImage** and OVMF CODE/VARS pair from the same firmware package. No installer, elevated host changes or NVRAM modifications are performed by the script.

```sh
export AIFT_M1_KERNEL=/boot/vmlinuz-<installed-x86_64-version>
export AIFT_M1_BUSYBOX=/usr/bin/busybox
export AIFT_M1_OVMF_CODE=/usr/share/OVMF/OVMF_CODE.fd
export AIFT_M1_OVMF_VARS=/usr/share/OVMF/OVMF_VARS.fd
bash -n scripts/m1_uefi_qemu_smoke.sh
bash scripts/m1_uefi_qemu_smoke.sh
cat artifacts/m1-uefi/serial.log
sha256sum -c artifacts/m1-uefi/SHA256SUMS
```

**Important:** These paths are examples, not assertions that the files exist. Verify `file "$AIFT_M1_BUSYBOX"` reports *statically linked*. QEMU is invoked without KVM so it can run on an unprivileged x86-64 host. A temporary copy of OVMF VARS is used and deleted afterward; the host's firmware/NVRAM and physical disks are never changed. The harness deliberately powers off after the init marker; this is **M1 smoke evidence only**, not M2 boot selection or M3 service supervision.

## Acceptance and limitations

Only label **VERIFIED BOOT** if QEMU exits successfully, `serial.log` contains `AIFT_M1_BOOT_OK`, `SHA256SUMS` verifies the ISO and the exact commit, host architecture, OVMF/kernel package versions and commands are recorded. The generated ISO is not a signed release and does not establish Secure Boot, Windows/macOS compatibility, persistence, a desktop, or hardware readiness.

A missing dependency, unavailable OVMF pair, absent static BusyBox, QEMU timeout or missing serial marker is **BLOCKED WITH EVIDENCE**, not a successful boot. Save `build.log`, `qemu.log` and `serial.log` for review. Rollback is deletion of the draft branch or reverting the harness commit; generated artifacts stay outside source control.

## Machine-readable evidence template

```json
{
  "schema": "aift.os.milestone-evidence.v1",
  "milestone": "M1",
  "target": "qemu-x86_64-ovmf",
  "status": "NOT_VERIFIED",
  "commit": null,
  "artifact_sha256": null,
  "serial_marker": "AIFT_M1_BOOT_OK",
  "command": "bash scripts/m1_uefi_qemu_smoke.sh",
  "rollback": "delete disposable output and revert draft changes",
  "next": "run on a host with QEMU, OVMF, GRUB EFI, x86-64 kernel and static BusyBox"
}
```
