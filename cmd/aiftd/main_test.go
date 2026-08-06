package main

import "testing"

func TestRequireLoopback(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:8787", "[::1]:8787", "localhost:8787"} {
		if err := requireLoopback(addr); err != nil {
			t.Fatalf("requireLoopback(%q): %v", addr, err)
		}
	}
	for _, addr := range []string{"0.0.0.0:8787", "192.168.1.10:8787"} {
		if err := requireLoopback(addr); err == nil {
			t.Fatalf("requireLoopback(%q) unexpectedly passed", addr)
		}
	}
}
