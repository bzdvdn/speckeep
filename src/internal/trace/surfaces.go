package trace

import (
	"path"
	"regexp"
	"sort"
	"strings"
)

// TaskSurface is the declared footprint of one task: what it plans to touch
// and which proof entries it claims.
type TaskSurface struct {
	ID      string
	Closed  bool
	ACs     []string
	Touches []string
	Proofs  []Entry
}

// ParseTaskSurfaces extracts per-task Touches/Proof/AC declarations from
// tasks.md. Open tasks are included because their declared footprint still
// matters for drift detection.
func ParseTaskSurfaces(content string) []TaskSurface {
	var (
		surfaces []TaskSurface
		current  *TaskSurface
	)

	flush := func() {
		if current == nil {
			return
		}
		current.ACs = uniqueSorted(current.ACs)
		current.Touches = uniqueSorted(current.Touches)
		surfaces = append(surfaces, *current)
		current = nil
	}

	for _, line := range strings.Split(content, "\n") {
		if matches := checkboxTokenPattern.FindStringSubmatch(line); len(matches) > 0 {
			taskID := firstMatch(taskIDPattern, line)
			if taskID == "" {
				continue
			}
			flush()
			current = &TaskSurface{ID: taskID, Closed: matches[1] == "x"}
			current.ACs = append(current.ACs, findACIDs(line)...)
		}
		if current == nil {
			continue
		}
		if refs := parseTouches(line); len(refs) > 0 {
			current.Touches = append(current.Touches, refs...)
		}
		if proof := proofLinePattern.FindStringSubmatch(line); len(proof) > 0 {
			current.Proofs = append(current.Proofs, Entry{
				TaskID: current.ID,
				Kind:   proof[1],
				File:   proof[2],
				Anchor: proof[3],
			})
		}
	}
	flush()
	return surfaces
}

func parseTouches(line string) []string {
	idx := strings.Index(line, "Touches:")
	if idx < 0 {
		return nil
	}
	raw := line[idx+len("Touches:"):]
	// Trailing requirement/task IDs are commonly written as "(AC-001, RQ-002)".
	if cut := strings.Index(raw, " ("); cut >= 0 {
		raw = raw[:cut]
	}
	var refs []string
	for _, part := range strings.Split(raw, ",") {
		ref := strings.Trim(strings.TrimSpace(part), "`")
		ref = strings.TrimSuffix(ref, ".")
		ref = filepathToSlash(ref)
		if ref == "" || strings.HasPrefix(ref, "<") {
			continue
		}
		if idOnlyRef.MatchString(ref) {
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}

var idOnlyRef = regexp.MustCompile(`^(?:AC|RQ)-[0-9]{3}\b|^T[0-9]+\.[0-9]+$`)

func findACIDs(line string) []string {
	return acceptanceIDPattern.FindAllString(line, -1)
}

// DeclaredSurfaces returns the union of Touches/Proof paths for a task list.
func DeclaredSurfaces(surfaces []TaskSurface) []string {
	seen := map[string]bool{}
	var out []string
	for _, surface := range surfaces {
		for _, ref := range surface.Touches {
			if !seen[ref] {
				seen[ref] = true
				out = append(out, ref)
			}
		}
		for _, proof := range surface.Proofs {
			file := filepathToSlash(strings.Trim(proof.File, "`"))
			if file != "" && !seen[file] {
				seen[file] = true
				out = append(out, file)
			}
		}
	}
	sort.Strings(out)
	return out
}

func surfaceRefs(surface TaskSurface) []string {
	refs := append([]string{}, surface.Touches...)
	for _, proof := range surface.Proofs {
		refs = append(refs, filepathToSlash(strings.Trim(proof.File, "`")))
	}
	return uniqueSorted(refs)
}

// MatchSurface maps a repo-relative file to the tasks, ACs and anchors that
// declared it.
func MatchSurface(file string, surfaces []TaskSurface) (tasks, acs, anchors []string) {
	file = filepathToSlash(file)
	for _, surface := range surfaces {
		matched := false
		for _, ref := range surface.Touches {
			if refMatches(file, ref) {
				matched = true
				break
			}
		}
		if !matched {
			for _, proof := range surface.Proofs {
				if filepathToSlash(strings.Trim(proof.File, "`")) == file {
					matched = true
					if proof.Anchor != "" {
						anchors = append(anchors, proof.Anchor)
					}
				}
			}
		}
		if matched {
			tasks = append(tasks, surface.ID)
			acs = append(acs, surface.ACs...)
			for _, proof := range surface.Proofs {
				if proof.Anchor != "" && filepathToSlash(strings.Trim(proof.File, "`")) == file {
					anchors = append(anchors, proof.Anchor)
				}
			}
		}
	}
	return uniqueSorted(tasks), uniqueSorted(acs), uniqueSorted(anchors)
}

// MatchesRef reports whether a repo-relative file matches a declared surface
// reference (exact path, directory prefix, or glob).
func MatchesRef(file, ref string) bool {
	return refMatches(file, ref)
}

func refMatches(file, ref string) bool {
	ref = strings.Trim(ref, "`")
	ref = filepathToSlash(ref)
	if ref == "" {
		return false
	}
	if strings.ContainsAny(ref, "*?[") {
		ok, err := path.Match(ref, file)
		return err == nil && ok
	}
	if strings.HasSuffix(ref, "/") {
		return strings.HasPrefix(file, ref)
	}
	return file == ref || strings.HasPrefix(file, ref+"/")
}

func filepathToSlash(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")
}

func uniqueSorted(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
