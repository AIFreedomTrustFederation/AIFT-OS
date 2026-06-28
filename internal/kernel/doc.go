// Package kernel provides a planned command registration framework for AIFT-OS.
//
// Status: planned
//
// The kernel package defines a Command type and Kernel struct that will
// eventually replace the switch-case dispatch in cmd/aift/main.go with
// a registry-based pattern. Each command would be registered via
// kernel.Register() and dispatched through the Kernel.
//
// This package is part of the intended architecture but not yet wired
// into the runtime. It is imported by cmd/aift to prevent orphaning
// and to make the planned integration path visible.
package kernel
