// Package scheduler provides a planned tick-based scheduling loop for AIFT-OS.
//
// Status: planned
//
// The scheduler runs registry generation, dashboard updates, and dependency
// reports on a configurable interval. It is the intended replacement for
// the direct jobs.RunAll call in runtime.Tick.
//
// This package is part of the intended architecture but not yet wired
// into the active runtime path. It is imported by cmd/aift to prevent
// orphaning and to make the planned integration path visible.
package scheduler
