package providers

import (
	"fmt"
	"path/filepath"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/jsonfile"
)

type Provider struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func Defaults() []Provider {
	return []Provider{
		{Name: "local-git", Type: "git", Status: "enabled", Description: "Local Git repository provider"},
		{Name: "github", Type: "git-host", Status: "configured-by-repo-remotes", Description: "GitHub remotes discovered from sovereign repositories"},
		{Name: "aift-forge", Type: "forge", Status: "planned", Description: "Future local-first federation forge provider"},
		{Name: "ollama", Type: "ai", Status: "planned", Description: "Local Ollama AI runtime provider"},
		{Name: "llamacpp", Type: "ai", Status: "planned", Description: "Local llama.cpp provider"},
		{Name: "vllm", Type: "ai", Status: "planned", Description: "Local/network vLLM provider"},
		{Name: "openai-compatible", Type: "ai", Status: "disabled-by-default", Description: "OpenAI-compatible endpoint provider"},
	}
}

func WriteRegistry(cfg config.Config) error {
	return jsonfile.Write(filepath.Join(cfg.OSHome, "registry", "providers.json"), Defaults(), true)
}

func List(cfg config.Config) error {
	if err := WriteRegistry(cfg); err != nil {
		return err
	}

	fmt.Printf("%-22s %-14s %-24s %s\n", "PROVIDER", "TYPE", "STATUS", "DESCRIPTION")
	for _, p := range Defaults() {
		fmt.Printf("%-22s %-14s %-24s %s\n", p.Name, p.Type, p.Status, p.Description)
	}

	return nil
}
