package trace

import (
	"reflect"
	"testing"
)

const surfacesContent = `# Tasks

## Phase 1

- [x] T1.1 Add export (AC-001, AC-002). Touches: src/a.go, src/dir/
  Proof: code src/a.go RunA
- [ ] T2.1 Later work. Touches: src/later.go
`

func TestParseTaskSurfaces(t *testing.T) {
	surfaces := ParseTaskSurfaces(surfacesContent)
	if len(surfaces) != 2 {
		t.Fatalf("got %d surfaces, want 2: %+v", len(surfaces), surfaces)
	}
	if !surfaces[0].Closed || surfaces[1].Closed {
		t.Errorf("closed flags = %v, %v; want true, false", surfaces[0].Closed, surfaces[1].Closed)
	}
	if !reflect.DeepEqual(surfaces[0].ACs, []string{"AC-001", "AC-002"}) {
		t.Errorf("ACs = %v", surfaces[0].ACs)
	}
	if !reflect.DeepEqual(surfaces[0].Touches, []string{"src/a.go", "src/dir/"}) {
		t.Errorf("Touches = %v", surfaces[0].Touches)
	}
	if len(surfaces[0].Proofs) != 1 || surfaces[0].Proofs[0].Anchor != "RunA" {
		t.Errorf("Proofs = %+v", surfaces[0].Proofs)
	}
}

func TestMatchSurface(t *testing.T) {
	surfaces := ParseTaskSurfaces(surfacesContent)

	tasks, acs, anchors := MatchSurface("src/a.go", surfaces)
	if !reflect.DeepEqual(tasks, []string{"T1.1"}) {
		t.Errorf("tasks = %v", tasks)
	}
	if !reflect.DeepEqual(acs, []string{"AC-001", "AC-002"}) {
		t.Errorf("acs = %v", acs)
	}
	if !reflect.DeepEqual(anchors, []string{"RunA"}) {
		t.Errorf("anchors = %v", anchors)
	}

	tasks, _, _ = MatchSurface("src/dir/nested/file.go", surfaces)
	if !reflect.DeepEqual(tasks, []string{"T1.1"}) {
		t.Errorf("directory touch tasks = %v", tasks)
	}

	tasks, _, _ = MatchSurface("src/later.go", surfaces)
	if !reflect.DeepEqual(tasks, []string{"T2.1"}) {
		t.Errorf("open task touch = %v", tasks)
	}

	tasks, _, _ = MatchSurface("src/other.go", surfaces)
	if len(tasks) != 0 {
		t.Errorf("unexpected match: %v", tasks)
	}
}

func TestDeclaredNotChangedSkipsOpenTasks(t *testing.T) {
	surfaces := ParseTaskSurfaces(surfacesContent)
	got := declaredNotChanged(surfaces, []string{"src/a.go"})
	if !reflect.DeepEqual(got, []string{"src/dir/"}) {
		t.Errorf("declaredNotChanged = %v", got)
	}
	for _, ref := range got {
		if ref == "src/later.go" {
			t.Errorf("open task surface reported as drift: %v", got)
		}
	}
}

func TestParseTouchesStripsTrailingIDs(t *testing.T) {
	content := "- [ ] T1.1 Add (AC-001). Touches: server/report_export (AC-001, AC-002, RQ-002)\n- [ ] T2.1 Next. Touches: src/a.go, src/b.go (RQ-003)\n"
	surfaces := ParseTaskSurfaces(content)
	if got, want := surfaces[0].Touches, []string{"server/report_export"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Touches = %v, want %v", got, want)
	}
	if got, want := surfaces[1].Touches, []string{"src/a.go", "src/b.go"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Touches = %v, want %v", got, want)
	}
}
