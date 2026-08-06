package uxi

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrMutationDisabled = errors.New("mutating adapters are disabled")

type AdapterDescriptor struct {
	Kind        string `json:"kind"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Mutating    bool   `json:"mutating"`
	Risk        string `json:"risk"`
}

type AdapterRequest struct {
	SessionID  string         `json:"session_id"`
	ActionID   string         `json:"action_id"`
	Target     string         `json:"target"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type AdapterResult struct {
	Summary  string         `json:"summary"`
	Evidence []Evidence     `json:"evidence,omitempty"`
	Data     map[string]any `json:"data,omitempty"`
}

type Adapter interface {
	Descriptor() AdapterDescriptor
	Invoke(context.Context, AdapterRequest) (AdapterResult, error)
}

type AdapterRegistry struct {
	adapters map[string]Adapter
}

func NewAdapterRegistry(adapters ...Adapter) (*AdapterRegistry, error) {
	registry := &AdapterRegistry{adapters: map[string]Adapter{}}
	for _, adapter := range adapters {
		if adapter == nil {
			return nil, errors.New("adapter is nil")
		}
		descriptor := adapter.Descriptor()
		descriptor.Kind = strings.TrimSpace(descriptor.Kind)
		if descriptor.Kind == "" {
			return nil, errors.New("adapter kind is required")
		}
		if _, exists := registry.adapters[descriptor.Kind]; exists {
			return nil, fmt.Errorf("duplicate adapter kind %q", descriptor.Kind)
		}
		registry.adapters[descriptor.Kind] = adapter
	}
	return registry, nil
}

func NewDefaultAdapterRegistry(aiftRoot string) (*AdapterRegistry, error) {
	return NewAdapterRegistry(
		RepositoryInspectAdapter{AIFTRoot: aiftRoot},
		ForgeMissionInspectAdapter{AIFTRoot: aiftRoot},
	)
}

func (r *AdapterRegistry) List() []AdapterDescriptor {
	if r == nil {
		return []AdapterDescriptor{}
	}
	descriptors := make([]AdapterDescriptor, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		descriptors = append(descriptors, adapter.Descriptor())
	}
	sort.Slice(descriptors, func(i, j int) bool { return descriptors[i].Kind < descriptors[j].Kind })
	return descriptors
}

func (r *AdapterRegistry) Lookup(kind string) (Adapter, bool) {
	if r == nil {
		return nil, false
	}
	adapter, ok := r.adapters[strings.TrimSpace(kind)]
	return adapter, ok
}

type RepositoryInspectAdapter struct{ AIFTRoot string }

func (a RepositoryInspectAdapter) Descriptor() AdapterDescriptor {
	return AdapterDescriptor{Kind: "repository.inspect", Version: "v1", Description: "Read evidence-backed repository status.", Mutating: false, Risk: "low"}
}

func (a RepositoryInspectAdapter) Invoke(_ context.Context, request AdapterRequest) (AdapterResult, error) {
	repositories, err := DiscoverRepositories(a.AIFTRoot)
	if err != nil {
		return AdapterResult{}, err
	}
	for _, repository := range repositories {
		if strings.EqualFold(repository.Name, request.Target) || strings.EqualFold(repository.ID, request.Target) {
			return AdapterResult{
				Summary:  fmt.Sprintf("Repository %s inspected from local evidence", repository.Name),
				Evidence: repository.Evidence,
				Data:     map[string]any{"repository": repository},
			}, nil
		}
	}
	return AdapterResult{}, fmt.Errorf("repository not found: %s", request.Target)
}

type ForgeMissionInspectAdapter struct{ AIFTRoot string }

func (a ForgeMissionInspectAdapter) Descriptor() AdapterDescriptor {
	return AdapterDescriptor{Kind: "forge.mission.inspect", Version: "v1", Description: "Read persisted Forge mission evidence.", Mutating: false, Risk: "low"}
}

func (a ForgeMissionInspectAdapter) Invoke(_ context.Context, request AdapterRequest) (AdapterResult, error) {
	mission, evidence, err := InspectForgeMission(a.AIFTRoot)
	if err != nil {
		return AdapterResult{}, err
	}
	return AdapterResult{
		Summary:  forgeMissionAnswer(mission),
		Evidence: evidence,
		Data:     map[string]any{"mission": mission, "requested_target": request.Target},
	}, nil
}
