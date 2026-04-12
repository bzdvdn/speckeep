package agents

import "fmt"

type Adapter interface {
	Target() string
	Render(commands []CommandDefinition, language string) ([]File, error)
	Paths(commands []CommandDefinition, language string) ([]string, error)
}

var orderedTargets = []string{"aider", "claude", "codex", "copilot", "cursor", "kilocode", "roocode", "trae", "windsurf"}

var adapterRegistry = map[string]Adapter{
	"aider":    aiderAdapter{},
	"claude":   claudeAdapter{},
	"codex":    codexAdapter{},
	"copilot":  copilotAdapter{},
	"cursor":   cursorAdapter{},
	"kilocode": kilocodeAdapter{},
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
		return nil, fmt.Errorf("unsupported agent target %q", target)
	}
	return adapter, nil
}
