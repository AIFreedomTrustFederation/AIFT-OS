package gitx

import (
	"bytes"
	"os/exec"
	"strings"
)

func Run(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func Branch(repo string) string {
	out, err := Run(repo, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil || out == "" {
		return "unknown"
	}
	return out
}

func Remote(repo string) string {
	out, err := Run(repo, "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	return out
}

func Dirty(repo string) bool {
	out, err := Run(repo, "status", "--porcelain")
	return err == nil && strings.TrimSpace(out) != ""
}
