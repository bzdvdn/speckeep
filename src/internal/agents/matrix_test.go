package agents

import (
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func expectedPathsForTarget(target string, commands []CommandDefinition) []string {
	base := targetSkillDirs[target]
	paths := []string{base + "/sdd/SKILL.md"}
	for _, cmd := range commands {
		paths = append(paths, base+"/spk-"+cmd.Name+"/SKILL.md")
	}
	if flat, ok := targetFlatCommandDirs[target]; ok {
		for _, cmd := range commands {
			paths = append(paths, flat.dir+"/spk-"+cmd.Name+".md")
		}
	}
	if target == "gemini" {
		for _, cmd := range commands {
			paths = append(paths, ".gemini/commands/spk-"+cmd.Name+".toml")
		}
	}
	if target == "aider" {
		paths = append(paths, ".aider/CONVENTIONS.md")
	}
	sort.Strings(paths)
	return paths
}

func frontmatterName(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}
	for _, line := range strings.Split(content, "\n")[1:] {
		if line == "---" {
			break
		}
		if strings.HasPrefix(line, "name:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
	}
	return ""
}

// TestAgentArtifactMatrix is the golden matrix: for every supported target
// crossed with both languages and both shells, the generated file set must
// match exactly, and the per-file invariants must hold. This is the guard
// against target drift introduced by the skills-first / flat-command /
// skills-only changes.
func TestAgentArtifactMatrix(t *testing.T) {
	commands := DefaultCommands("sh")

	legacy := map[string]struct{}{}
	legacyLists := [][]string{
		LegacyPrefixPaths(commands),
		LegacyCommandWrapperPaths(commands),
		LegacySkillPhasePaths(commands),
		LegacyOpenCodeCommandPaths(commands),
	}
	for _, list := range legacyLists {
		for _, p := range list {
			legacy[filepath.ToSlash(p)] = struct{}{}
		}
	}

	for _, target := range SupportedTargets() {
		for _, lang := range []string{"en", "ru"} {
			for _, shell := range []string{"sh", "powershell"} {
				name := target + "/" + lang + "/" + shell
				files, err := FilesForTarget(target, lang, shell)
				if err != nil {
					t.Fatalf("%s: FilesForTarget error: %v", name, err)
				}

				byPath := make(map[string]string, len(files))
				got := make([]string, 0, len(files))
				for _, f := range files {
					if _, dup := byPath[f.Path]; dup {
						t.Fatalf("%s: duplicate generated path %s", name, f.Path)
					}
					byPath[f.Path] = f.Content
					got = append(got, f.Path)
					if f.Content == "" {
						t.Fatalf("%s: empty content for %s", name, f.Path)
					}
					if _, bad := legacy[f.Path]; bad {
						t.Fatalf("%s: generated a legacy artifact %s", name, f.Path)
					}
					if strings.HasPrefix(f.Path, ".opencode/commands/") {
						t.Fatalf("%s: opencode must be skills-only, got %s", name, f.Path)
					}
				}
				sort.Strings(got)
				want := expectedPathsForTarget(target, commands)
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("%s: generated paths mismatch\n got: %v\nwant: %v", name, got, want)
				}

				base := targetSkillDirs[target]
				root := byPath[base+"/sdd/SKILL.md"]
				if frontmatterName(root) != "sdd" {
					t.Fatalf("%s: root skill name=%q, want sdd", name, frontmatterName(root))
				}
				if !strings.Contains(root, checkReadyScript(shell)) {
					t.Fatalf("%s: root skill must reference %s", name, checkReadyScript(shell))
				}
				if target == "opencode" && !strings.Contains(root, "Вспомогательные") && !strings.Contains(root, "Auxiliary") {
					t.Fatalf("%s: root skill must separate auxiliary commands", name)
				}

				for _, cmd := range commands {
					path := base + "/spk-" + cmd.Name + "/SKILL.md"
					content := byPath[path]
					if got := frontmatterName(content); got != "spk-"+cmd.Name {
						t.Fatalf("%s: %s name=%q, want spk-%s", name, path, got, cmd.Name)
					}
					advertised := strings.Contains(content, "check-ready")
					if advertised != hasReadyCheck(cmd.Name) {
						t.Fatalf("%s: %s readiness advertised=%v, hasCheck=%v", name, path, advertised, hasReadyCheck(cmd.Name))
					}
					if hasReadyCheck(cmd.Name) && !strings.Contains(content, checkReadyScript(shell)+" "+cmd.Name) {
						t.Fatalf("%s: %s must reference %s %s", name, path, checkReadyScript(shell), cmd.Name)
					}
				}

				if flat, ok := targetFlatCommandDirs[target]; ok {
					for _, cmd := range commands {
						path := flat.dir + "/spk-" + cmd.Name + ".md"
						content := byPath[path]
						if strings.HasPrefix(content, "---") {
							t.Fatalf("%s: flat command %s must not carry skill frontmatter", name, path)
						}
						if !strings.Contains(content, flat.invokePrefix+"spk-"+cmd.Name) {
							t.Fatalf("%s: flat command %s missing invocation", name, path)
						}
					}
				}
			}
		}
	}
}
