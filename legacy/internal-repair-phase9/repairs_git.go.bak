package repair

import (
	"os/exec"
	"strings"
)

type GitRepair struct{}

func (GitRepair) Name() string {
	return "git"
}

func (GitRepair) Detect(ctx Context) []Issue {

	if !inGitRepo(ctx.Config.OSHome) {
		return nil
	}

	out, err := exec.Command(
		"git",
		"-C",
		ctx.Config.OSHome,
		"status",
		"--short",
	).CombinedOutput()

	if err != nil {
		return []Issue{
			{
				ID:      "git-status",
				Repair:  "git",
				Message: strings.TrimSpace(string(out)),
				Safety:  Blocked,
			},
		}
	}

	var issues []Issue

	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {

		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.Contains(line, ".aift/") ||
			strings.Contains(line, "registry/") ||
			strings.Contains(line, "var/events/") {

			issues = append(issues, Issue{
				ID:      "generated-state",
				Repair:  "git",
				Message: line,
				Safety:  Safe,
			})

		} else {

			issues = append(issues, Issue{
				ID:      "source-change",
				Repair:  "git",
				Message: line,
				Safety:  Review,
			})

		}
	}

	return issues
}

func (GitRepair) Apply(ctx Context, issue Issue) Action {

	if !inGitRepo(ctx.Config.OSHome) {
		return Action{
			Repair:  "git",
			Message: "skipped git repair outside git worktree",
		}
	}

	restore := []string{
		".aift/capabilities.json",
		".aift/providers.json",
		".aift/repos.json",
		".aift/workflows.json",
		"var/events/events.jsonl",
	}

	for _, f := range restore {
		_ = run(ctx.Config.OSHome, "git", "restore", f)
	}

	return Action{
		Repair:  "git",
		Message: "restored tracked generated runtime state",
	}
}

func (GitRepair) Verify(ctx Context, issue Issue) error {
	return nil
}
