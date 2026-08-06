package uxi

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Engine struct {
	Store     *Store
	AIFTRoot  string
	Completer Completer
}

func NewEngine(store *Store, root string, completer Completer) (*Engine, error) {
	if store == nil {
		return nil, errors.New("store is required")
	}
	if strings.TrimSpace(root) == "" {
		return nil, errors.New("AIFT root is required")
	}
	return &Engine{Store: store, AIFTRoot: root, Completer: completer}, nil
}

func (e *Engine) Repositories() ([]Repository, error) {
	return DiscoverRepositories(e.AIFTRoot)
}

func (e *Engine) HandleMessage(ctx context.Context, sessionID, content string) (Session, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return Session{}, errors.New("message content is required")
	}
	session, err := e.Store.GetSession(sessionID)
	if err != nil {
		return Session{}, err
	}
	repos, err := e.Repositories()
	if err != nil {
		return Session{}, err
	}
	if _, err := e.Store.AppendTurn(sessionID, Turn{Role: "user", Content: content}); err != nil {
		return Session{}, err
	}

	var answer string
	var evidence []Evidence
	metadata := map[string]any{"mode": "local-inference"}

	if target, ok := inspectCommand(content); ok {
		answer, evidence = inspectAnswer(target, repos)
		metadata["mode"] = "deterministic-inspection"
	} else {
		system := systemPrompt(repos)
		messages := conversationMessages(session.Turns, content, 24)
		if e.Completer != nil {
			answer, err = e.Completer.Complete(ctx, system, messages)
		}
		if e.Completer == nil || err != nil {
			answer = degradedAnswer(repos, err)
			metadata["mode"] = "degraded"
			if err != nil {
				metadata["inference_error"] = err.Error()
			}
		}
	}

	return e.Store.AppendTurn(sessionID, Turn{
		Role:      "assistant",
		Content:   answer,
		Evidence:  evidence,
		Metadata:  metadata,
		CreatedAt: time.Now().UTC(),
	})
}

func conversationMessages(turns []Turn, current string, limit int) []ChatMessage {
	start := 0
	if limit > 1 && len(turns) > limit-1 {
		start = len(turns) - (limit - 1)
	}
	messages := make([]ChatMessage, 0, len(turns)-start+1)
	for _, turn := range turns[start:] {
		if turn.Role != "user" && turn.Role != "assistant" {
			continue
		}
		if strings.TrimSpace(turn.Content) == "" {
			continue
		}
		messages = append(messages, ChatMessage{Role: turn.Role, Content: turn.Content})
	}
	return append(messages, ChatMessage{Role: "user", Content: current})
}

func inspectCommand(content string) (string, bool) {
	fields := strings.Fields(content)
	if len(fields) >= 1 && (fields[0] == "/inspect" || strings.EqualFold(fields[0], "inspect")) {
		if len(fields) == 1 {
			return "", true
		}
		return strings.Join(fields[1:], " "), true
	}
	return "", false
}

func inspectAnswer(target string, repos []Repository) (string, []Evidence) {
	if strings.TrimSpace(target) == "" {
		names := make([]string, 0, len(repos))
		for _, repo := range repos {
			names = append(names, repo.Name)
		}
		return fmt.Sprintf("I discovered %d Git repositories: %s. Use `/inspect <repository>` for an evidence-backed view.", len(repos), strings.Join(names, ", ")), nil
	}
	for _, repo := range repos {
		if strings.EqualFold(repo.Name, target) || strings.EqualFold(repo.ID, target) {
			capStatuses := map[string]int{}
			for _, capability := range repo.Capabilities {
				capStatuses[capability.Status]++
			}
			var statusParts []string
			for status, count := range capStatuses {
				statusParts = append(statusParts, fmt.Sprintf("%s=%d", status, count))
			}
			sort.Strings(statusParts)
			capSummary := "no capability manifest"
			if len(statusParts) > 0 {
				capSummary = strings.Join(statusParts, ", ")
			}
			return fmt.Sprintf("%s is classified as `%s` and currently reports `%s`. Languages: %s. Capabilities: %s. Path: %s", repo.Name, repo.Role, repo.Status, valueOr(repo.Languages, "not detected"), capSummary, filepath.Clean(repo.Path)), repo.Evidence
		}
	}
	return fmt.Sprintf("I could not find `%s` under the configured AIFT root. The available repositories are: %s.", target, repositoryNames(repos)), nil
}

func systemPrompt(repos []Repository) string {
	return "You are Aetherion inside MoBox UXI. Be truthful, local-first, concise, and evidence-aware. " +
		"Do not claim that an action ran unless an execution result proves it. Distinguish observed, inferred, planned, and blocked states. " +
		"The currently discovered repositories are: " + repositoryNames(repos) + ". " +
		"This first release is read-only. Recommend plans, but do not imply that files, deployments, money, identity, or external systems were changed."
}

func degradedAnswer(repos []Repository, inferenceErr error) string {
	message := fmt.Sprintf("MoBox UXI is operating in truthful read-only mode. I discovered %d repositories: %s.", len(repos), repositoryNames(repos))
	if inferenceErr != nil {
		message += " The local model endpoint is not available, so I did not fabricate an AI response. Use `/inspect <repository>` while the model runtime is offline."
	}
	return message
}

func repositoryNames(repos []Repository) string {
	names := make([]string, 0, len(repos))
	for _, repo := range repos {
		names = append(names, repo.Name)
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}

func valueOr(values []string, fallback string) string {
	if len(values) == 0 {
		return fallback
	}
	return strings.Join(values, ", ")
}
