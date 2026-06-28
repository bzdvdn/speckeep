package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrUnsupportedShell = errors.New("unsupported shell, expected sh or powershell")
)

const speckeepDirName = ".speckeep"

var defaultConfig = Config{
	Version: 1,
	Project: Project{
		Name:             "my-project",
		ConstitutionFile: "CONSTITUTION.md",
	},
	Runtime: Runtime{
		Shell: "sh",
	},
	Paths: Paths{
		SpecsDir:     "specs/active",
		ArchiveDir:   "specs/archived",
		TemplatesDir: ".speckeep/templates",
		ScriptsDir:   ".speckeep/scripts",
	},
	Language: Language{
		Default:  "en",
		Docs:     "en",
		Agent:    "en",
		Comments: "en",
	},
	Agents: Agents{
		UpdateAgentsMD: true,
		AgentsFile:     "AGENTS.md",
		Targets:        nil,
	},
	Templates: Templates{
		Spec:               "spec.md",
		Plan:               "plan.md",
		Tasks:              "tasks.md",
		DataModel:          "data-model.md",
		ContractsAPI:       "contracts/api.md",
		ContractsEvents:    "contracts/events.md",
		ArchiveSummary:     "archive/summary.md",
		InspectReport:      "inspect.md",
		VerifyReport:       "verify.md",
		Constitution:       "constitution.md",
		ConstitutionPrompt: "prompts/constitution.md",
		SpecPrompt:         "prompts/spec.md",
		InspectPrompt:      "prompts/inspect.md",
		PlanPrompt:         "prompts/plan.md",
		TasksPrompt:        "prompts/tasks.md",
		ImplementPrompt:    "prompts/implement.md",
		VerifyPrompt:       "prompts/verify.md",
	},
	Scripts: ScriptDefaultsForShell("sh"),
}

type Config struct {
	Version   int       `yaml:"version"`
	Project   Project   `yaml:"project"`
	Runtime   Runtime   `yaml:"runtime"`
	Paths     Paths     `yaml:"paths"`
	Language  Language  `yaml:"language"`
	Agents    Agents    `yaml:"agents"`
	Templates Templates `yaml:"templates"`
	Scripts   Scripts   `yaml:"scripts"`
	Workflow  Workflow  `yaml:"workflow,omitempty"`
}

type Workflow struct {
	Schema string `yaml:"schema,omitempty"`
}

type Project struct {
	Name             string `yaml:"name"`
	ConstitutionFile string `yaml:"constitution_file"`
}

type Runtime struct {
	Shell string `yaml:"shell"`
}

type Paths struct {
	SpecsDir     string `yaml:"specs_dir"`
	ArchiveDir   string `yaml:"archive_dir"`
	TemplatesDir string `yaml:"templates_dir"`
	ScriptsDir   string `yaml:"scripts_dir"`
}

type Language struct {
	Default  string `yaml:"default"`
	Docs     string `yaml:"docs"`
	Agent    string `yaml:"agent"`
	Comments string `yaml:"comments"`
}

type Agents struct {
	UpdateAgentsMD bool     `yaml:"update_agents_md"`
	AgentsFile     string   `yaml:"agents_file"`
	Targets        []string `yaml:"targets,omitempty"`
}

type Templates struct {
	Spec               string `yaml:"spec"`
	Plan               string `yaml:"plan"`
	Tasks              string `yaml:"tasks"`
	DataModel          string `yaml:"data_model"`
	ContractsAPI       string `yaml:"contracts_api"`
	ContractsEvents    string `yaml:"contracts_events"`
	ArchiveSummary     string `yaml:"archive_summary"`
	InspectReport      string `yaml:"inspect_report"`
	VerifyReport       string `yaml:"verify_report"`
	Constitution       string `yaml:"constitution"`
	ConstitutionPrompt string `yaml:"constitution_prompt"`
	SpecPrompt         string `yaml:"spec_prompt"`
	InspectPrompt      string `yaml:"inspect_prompt"`
	PlanPrompt         string `yaml:"plan_prompt"`
	TasksPrompt        string `yaml:"tasks_prompt"`
	ImplementPrompt    string `yaml:"implement_prompt"`
	VerifyPrompt       string `yaml:"verify_prompt"`
}

type Scripts struct {
	// RunSpeckeep is the canonical wrapper entrypoint script name.
	RunSpeckeep string `yaml:"run_speckeep"`

	// RunDraftspec is a deprecated alias for RunSpeckeep kept for legacy configs.
	// It is ignored when RunSpeckeep is set and is cleared on Save().
	RunDraftspec string `yaml:"run_draftspec,omitempty"`

	CheckReady        string `yaml:"check_ready"`
	InspectSpec       string `yaml:"inspect_spec"`
	CheckConstitution string `yaml:"check_constitution"`
	VerifyTaskState   string `yaml:"verify_task_state"`
	ListOpenTasks     string `yaml:"list_open_tasks"`
	LinkAgents        string `yaml:"link_agents"`
	ListSpecs         string `yaml:"list_specs"`
	ShowSpec          string `yaml:"show_spec"`
}

func NormalizeShell(shell string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(shell))
	switch value {
	case "sh", "powershell":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported shell %q: %w", shell, ErrUnsupportedShell)
	}
}

func ScriptDefaultsForShell(shell string) Scripts {
	normalized, err := NormalizeShell(shell)
	if err != nil {
		normalized = "sh"
	}

	ext := ".sh"
	if normalized == "powershell" {
		ext = ".ps1"
	}

	return Scripts{
		RunSpeckeep:       "run-speckeep" + ext,
		CheckReady:        "check-ready" + ext,
		InspectSpec:       "inspect-spec" + ext,
		CheckConstitution: "check-constitution" + ext,
		VerifyTaskState:   "verify-task-state" + ext,
		ListOpenTasks:     "list-open-tasks" + ext,
		LinkAgents:        "link-agents" + ext,
		ListSpecs:         "list-specs" + ext,
		ShowSpec:          "show-spec" + ext,
	}
}

func Default() Config {
	cfg := defaultConfig
	cfg.applyDefaults()
	return cfg
}

func Load(ctx context.Context, root string) (Config, error) {
	return load(ctx, root)
}

func load(_ context.Context, root string) (Config, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	configPath := filepath.Join(root, speckeepDirName, "speckeep.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			legacyPaths := []string{
				filepath.Join(root, ".speckeep", "specgate.yaml"),
			}
			for _, legacyPath := range legacyPaths {
				legacyContent, legacyErr := os.ReadFile(legacyPath)
				if legacyErr != nil {
					continue
				}
				if err := yaml.Unmarshal(legacyContent, &cfg); err != nil {
					return Config{}, fmt.Errorf("parse speckeep config: %w", err)
				}
				cfg.applyDefaults()
				return cfg, nil
			}
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read speckeep config: %w", err)
	}

	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse speckeep config: %w", err)
	}

	cfg.applyDefaults()
	return cfg, nil
}

func Save(ctx context.Context, root string, cfg Config) error {
	return save(ctx, root, cfg)
}

func save(_ context.Context, root string, cfg Config) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	cfg.applyDefaults()
	configPath := filepath.Join(root, speckeepDirName, "speckeep.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal speckeep config: %w", err)
	}
	return os.WriteFile(configPath, content, 0o644)
}

func (c Config) DraftspecDir(root string) (string, error) { return resolve(root, speckeepDirName) }
func (c Config) ConfigPath(root string) (string, error) {
	draftspecDir, err := c.DraftspecDir(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(draftspecDir, "speckeep.yaml"), nil
}
func (c Config) SpecsDir(root string) (string, error)     { return resolve(root, c.Paths.SpecsDir) }
func (c Config) ArchiveDir(root string) (string, error)   { return resolve(root, c.Paths.ArchiveDir) }
func (c Config) TemplatesDir(root string) (string, error) { return resolve(root, c.Paths.TemplatesDir) }
func (c Config) ScriptsDir(root string) (string, error)   { return resolve(root, c.Paths.ScriptsDir) }

func (c *Config) applyDefaults() {
	if c.Version == 0 {
		c.Version = defaultConfig.Version
	}

	// Language: fallback chain (Default -> Docs/Agent/Comments)
	if c.Language.Default == "" {
		c.Language.Default = defaultConfig.Language.Default
	}
	if c.Language.Docs == "" {
		c.Language.Docs = c.Language.Default
	}
	if c.Language.Agent == "" {
		c.Language.Agent = c.Language.Default
	}
	if c.Language.Comments == "" {
		c.Language.Comments = c.Language.Default
	}

	// Scripts: migrate deprecated RunDraftspec before reflection defaults
	if c.Scripts.RunSpeckeep == "" && c.Scripts.RunDraftspec != "" {
		c.Scripts.RunSpeckeep = c.Scripts.RunDraftspec
		c.Scripts.RunDraftspec = ""
	}

	// Reflection-based defaulting for all remaining string fields
	applyDefaultsReflect(reflect.ValueOf(c).Elem(), reflect.ValueOf(defaultConfig))
}

// applyDefaultsReflect recursively sets empty exported string fields from defaults.
// Fields with yaml:"...,omitempty" tags, int fields, bool fields, and slice fields
// are skipped to preserve explicit zero-values and optional/legacy configuration.
func applyDefaultsReflect(target, defaultVal reflect.Value) {
	t := target.Type()
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}

		tag := ft.Tag.Get("yaml")
		if strings.Contains(tag, "omitempty") {
			continue
		}

		tf := target.Field(i)
		df := defaultVal.Field(i)

		if !tf.CanSet() {
			continue
		}

		switch tf.Kind() {
		case reflect.String:
			if tf.String() == "" {
				tf.SetString(df.String())
			}
		case reflect.Struct:
			applyDefaultsReflect(tf, df)
		}
	}
}

func resolve(root, configuredPath string) (string, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	if filepath.IsAbs(configuredPath) {
		return configuredPath, nil
	}
	return filepath.Join(root, filepath.FromSlash(configuredPath)), nil
}
