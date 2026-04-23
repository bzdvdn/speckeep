package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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
		SpecsDir:     "specs",
		ArchiveDir:   "archive",
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
		ArchivePrompt:      "prompts/archive.md",
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
	ArchivePrompt      string `yaml:"archive_prompt"`
	VerifyPrompt       string `yaml:"verify_prompt"`
}

type Scripts struct {
	// RunSpeckeep is the canonical wrapper entrypoint script name.
	RunSpeckeep string `yaml:"run_speckeep"`

	// RunDraftspec is a deprecated alias for RunSpeckeep kept for legacy configs.
	// It is ignored when RunSpeckeep is set and is cleared on Save().
	RunDraftspec        string `yaml:"run_draftspec,omitempty"`
	InspectSpec         string `yaml:"inspect_spec"`
	CheckConstitution   string `yaml:"check_constitution"`
	CheckSpecReady      string `yaml:"check_spec_ready"`
	CheckInspectReady   string `yaml:"check_inspect_ready"`
	CheckPlanReady      string `yaml:"check_plan_ready"`
	CheckTasksReady     string `yaml:"check_tasks_ready"`
	CheckImplementReady string `yaml:"check_implement_ready"`
	CheckArchiveReady   string `yaml:"check_archive_ready"`
	CheckVerifyReady    string `yaml:"check_verify_ready"`
	VerifyTaskState     string `yaml:"verify_task_state"`
	ListOpenTasks       string `yaml:"list_open_tasks"`
	LinkAgents          string `yaml:"link_agents"`
	ListSpecs           string `yaml:"list_specs"`
	ShowSpec            string `yaml:"show_spec"`
}

func NormalizeShell(shell string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(shell))
	switch value {
	case "sh", "powershell":
		return value, nil
	default:
		return "", fmt.Errorf("unsupported shell %q, expected sh or powershell", shell)
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
		RunSpeckeep:         "run-speckeep" + ext,
		InspectSpec:         "inspect-spec" + ext,
		CheckConstitution:   "check-constitution" + ext,
		CheckSpecReady:      "check-spec-ready" + ext,
		CheckInspectReady:   "check-inspect-ready" + ext,
		CheckPlanReady:      "check-plan-ready" + ext,
		CheckTasksReady:     "check-tasks-ready" + ext,
		CheckImplementReady: "check-implement-ready" + ext,
		CheckArchiveReady:   "check-archive-ready" + ext,
		CheckVerifyReady:    "check-verify-ready" + ext,
		VerifyTaskState:     "verify-task-state" + ext,
		ListOpenTasks:       "list-open-tasks" + ext,
		LinkAgents:          "link-agents" + ext,
		ListSpecs:           "list-specs" + ext,
		ShowSpec:            "show-spec" + ext,
	}
}

func Default() Config {
	cfg := defaultConfig
	cfg.applyDefaults()
	return cfg
}

func Load(root string) (Config, error) {
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
				filepath.Join(root, ".speckeep", "specgate.yaml"), // legacy config filename (pre-rename)
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

func Save(root string, cfg Config) error {
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
	if c.Project.Name == "" {
		c.Project.Name = defaultConfig.Project.Name
	}
	if c.Project.ConstitutionFile == "" {
		c.Project.ConstitutionFile = defaultConfig.Project.ConstitutionFile
	}
	if c.Runtime.Shell == "" {
		c.Runtime.Shell = defaultConfig.Runtime.Shell
	}
	if c.Paths.SpecsDir == "" {
		c.Paths.SpecsDir = defaultConfig.Paths.SpecsDir
	}
	if c.Paths.ArchiveDir == "" {
		c.Paths.ArchiveDir = defaultConfig.Paths.ArchiveDir
	}
	if c.Paths.TemplatesDir == "" {
		c.Paths.TemplatesDir = defaultConfig.Paths.TemplatesDir
	}
	if c.Paths.ScriptsDir == "" {
		c.Paths.ScriptsDir = defaultConfig.Paths.ScriptsDir
	}
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
	if c.Agents.AgentsFile == "" {
		c.Agents.AgentsFile = defaultConfig.Agents.AgentsFile
	}
	if !c.Agents.UpdateAgentsMD {
		c.Agents.UpdateAgentsMD = defaultConfig.Agents.UpdateAgentsMD
	}
	if c.Templates.Spec == "" {
		c.Templates.Spec = defaultConfig.Templates.Spec
	}
	if c.Templates.Plan == "" {
		c.Templates.Plan = defaultConfig.Templates.Plan
	}
	if c.Templates.Tasks == "" {
		c.Templates.Tasks = defaultConfig.Templates.Tasks
	}
	if c.Templates.DataModel == "" {
		c.Templates.DataModel = defaultConfig.Templates.DataModel
	}
	if c.Templates.ContractsAPI == "" {
		c.Templates.ContractsAPI = defaultConfig.Templates.ContractsAPI
	}
	if c.Templates.ContractsEvents == "" {
		c.Templates.ContractsEvents = defaultConfig.Templates.ContractsEvents
	}
	if c.Templates.ArchiveSummary == "" {
		c.Templates.ArchiveSummary = defaultConfig.Templates.ArchiveSummary
	}
	if c.Templates.InspectReport == "" {
		c.Templates.InspectReport = defaultConfig.Templates.InspectReport
	}
	if c.Templates.VerifyReport == "" {
		c.Templates.VerifyReport = defaultConfig.Templates.VerifyReport
	}
	if c.Templates.Constitution == "" {
		c.Templates.Constitution = defaultConfig.Templates.Constitution
	}
	if c.Templates.ConstitutionPrompt == "" {
		c.Templates.ConstitutionPrompt = defaultConfig.Templates.ConstitutionPrompt
	}
	if c.Templates.SpecPrompt == "" {
		c.Templates.SpecPrompt = defaultConfig.Templates.SpecPrompt
	}
	if c.Templates.InspectPrompt == "" {
		c.Templates.InspectPrompt = defaultConfig.Templates.InspectPrompt
	}
	if c.Templates.PlanPrompt == "" {
		c.Templates.PlanPrompt = defaultConfig.Templates.PlanPrompt
	}
	if c.Templates.TasksPrompt == "" {
		c.Templates.TasksPrompt = defaultConfig.Templates.TasksPrompt
	}
	if c.Templates.ImplementPrompt == "" {
		c.Templates.ImplementPrompt = defaultConfig.Templates.ImplementPrompt
	}
	if c.Templates.ArchivePrompt == "" {
		c.Templates.ArchivePrompt = defaultConfig.Templates.ArchivePrompt
	}
	if c.Templates.VerifyPrompt == "" {
		c.Templates.VerifyPrompt = defaultConfig.Templates.VerifyPrompt
	}
	defaultScripts := ScriptDefaultsForShell(c.Runtime.Shell)
	if c.Scripts.RunSpeckeep == "" && c.Scripts.RunDraftspec != "" {
		c.Scripts.RunSpeckeep = c.Scripts.RunDraftspec
		c.Scripts.RunDraftspec = ""
	}
	if c.Scripts.RunSpeckeep == "" {
		c.Scripts.RunSpeckeep = defaultScripts.RunSpeckeep
	}
	if c.Scripts.InspectSpec == "" {
		c.Scripts.InspectSpec = defaultScripts.InspectSpec
	}
	if c.Scripts.CheckConstitution == "" {
		c.Scripts.CheckConstitution = defaultScripts.CheckConstitution
	}
	if c.Scripts.CheckSpecReady == "" {
		c.Scripts.CheckSpecReady = defaultScripts.CheckSpecReady
	}
	if c.Scripts.CheckInspectReady == "" {
		c.Scripts.CheckInspectReady = defaultScripts.CheckInspectReady
	}
	if c.Scripts.CheckPlanReady == "" {
		c.Scripts.CheckPlanReady = defaultScripts.CheckPlanReady
	}
	if c.Scripts.CheckTasksReady == "" {
		c.Scripts.CheckTasksReady = defaultScripts.CheckTasksReady
	}
	if c.Scripts.CheckImplementReady == "" {
		c.Scripts.CheckImplementReady = defaultScripts.CheckImplementReady
	}
	if c.Scripts.CheckArchiveReady == "" {
		c.Scripts.CheckArchiveReady = defaultScripts.CheckArchiveReady
	}
	if c.Scripts.CheckVerifyReady == "" {
		c.Scripts.CheckVerifyReady = defaultScripts.CheckVerifyReady
	}
	if c.Scripts.VerifyTaskState == "" {
		c.Scripts.VerifyTaskState = defaultScripts.VerifyTaskState
	}
	if c.Scripts.ListOpenTasks == "" {
		c.Scripts.ListOpenTasks = defaultScripts.ListOpenTasks
	}
	if c.Scripts.LinkAgents == "" {
		c.Scripts.LinkAgents = defaultScripts.LinkAgents
	}
	if c.Scripts.ListSpecs == "" {
		c.Scripts.ListSpecs = defaultScripts.ListSpecs
	}
	if c.Scripts.ShowSpec == "" {
		c.Scripts.ShowSpec = defaultScripts.ShowSpec
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
