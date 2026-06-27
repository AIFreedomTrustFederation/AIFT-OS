package kernel

import (
	"fmt"

	"github.com/AIFreedomTrustFederation/AIFT-OS/internal/config"
)

type Command struct {
	Name string
	Help string
	Run  func(config.Config, []string) error
}

type Kernel struct {
	Config   config.Config
	Commands map[string]Command
}

func New(cfg config.Config) *Kernel {
	return &Kernel{
		Config:   cfg,
		Commands: map[string]Command{},
	}
}

func (k *Kernel) Register(c Command) {
	k.Commands[c.Name] = c
}

func (k *Kernel) Help() {
	fmt.Println("AIFT-OS Federation Control Plane")
	fmt.Println()
	fmt.Println("Commands:")
	for _, c := range k.Commands {
		fmt.Printf("  %-10s %s\n", c.Name, c.Help)
	}
}
