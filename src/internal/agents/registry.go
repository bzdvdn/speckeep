package agents

import (
	"errors"
	"fmt"
)

var ErrUnsupportedTarget = errors.New("unsupported agent target")

type Adapter interface {
	Target() string
	Render(commands []CommandDefinition, language string) ([]File, error)
	Paths(commands []CommandDefinition, language string) ([]string, error)
}

var orderedTargets = []string{"aider", "claude", "codex", "copilot", "cursor", "kilocode", "opencode", "roocode", "trae", "windsurf"}

var adapterRegistry = map[string]Adapter{
	"aider":    aiderAdapter{},
	"claude":   claudeAdapter{},
	"codex":    codexAdapter{},
	"copilot":  copilotAdapter{},
	"cursor":   cursorAdapter{},
	"kilocode": kilocodeAdapter{},
	"opencode": opencodeAdapter{},
	"roocode":  roocodeAdapter{},
	"trae":     traeAdapter{},
	"windsurf": windsurfAdapter{},
}

func SupportedTargets() []string {
	return append([]string(nil), orderedTargets...)
}

func adapterForTarget(target string) (Adapter, error) {
	adapter, ok := adapterRegistry[target]
	if !ok {
		return nil, fmt.Errorf("unsupported agent target %q: %w", target, ErrUnsupportedTarget)
	}
	return adapter, nil
}
