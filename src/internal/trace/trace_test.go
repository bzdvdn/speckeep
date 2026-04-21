package trace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTracePattern(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantType string
		wantTask string
		wantAC   string
		wantDesc string
	}{
		{
			name:     "task with description and AC",
			line:     "// @sk-task T1.1: Add something (AC-001)",
			wantType: "task",
			wantTask: "T1.1",
			wantAC:   "AC-001",
			wantDesc: "Add something",
		},
		{
			name:     "test with name and AC",
			line:     "// @sk-test T1.1: TestSomething (AC-002)",
			wantType: "test",
			wantTask: "T1.1",
			wantAC:   "AC-002",
			wantDesc: "TestSomething",
		},
		{
			name:     "task without AC",
			line:     "// @sk-task T2.2: Simple description",
			wantType: "task",
			wantTask: "T2.2",
			wantAC:   "",
			wantDesc: "Simple description",
		},
		{
			name:     "task with no colon",
			line:     "// @sk-task T3.3 (AC-003)",
			wantType: "task",
			wantTask: "T3.3",
			wantAC:   "AC-003",
			wantDesc: "",
		},
		{
			name:     "multiline comment style",
			line:     "/* @sk-task T4.4: In block comment (AC-004) */",
			wantType: "task",
			wantTask: "T4.4",
			wantAC:   "AC-004",
			wantDesc: "In block comment",
		},
		{
			name:     "shell comment style",
			line:     "# @sk-task T5.5: In shell script (AC-005)",
			wantType: "task",
			wantTask: "T5.5",
			wantAC:   "AC-005",
			wantDesc: "In shell script",
		},
		{
			name:     "legacy task format accepted",
			line:     "// @ds-task T6.6: Legacy annotation (AC-006)",
			wantType: "task",
			wantTask: "T6.6",
			wantAC:   "AC-006",
			wantDesc: "Legacy annotation",
		},
		{
			name:     "namespaced task id accepted",
			line:     "// @sk-task my-spec#T7.1: Namespaced (AC-007)",
			wantType: "task",
			wantTask: "my-spec#T7.1",
			wantAC:   "AC-007",
			wantDesc: "Namespaced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := tracePattern.FindStringSubmatch(tt.line)
			if len(matches) == 0 {
				t.Errorf("tracePattern.FindStringSubmatch() did not match line: %s", tt.line)
				return
			}
			if matches[1] != tt.wantType {
				t.Errorf("match type = %v, want %v", matches[1], tt.wantType)
			}
			if matches[2] != tt.wantTask {
				t.Errorf("match task = %v, want %v", matches[2], tt.wantTask)
			}
			if matches[4] != tt.wantAC {
				t.Errorf("match AC = %v, want %v", matches[4], tt.wantAC)
			}
		})
	}
}

func TestScan(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "trace-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a mock file with annotations
	mockFile := filepath.Join(tmpDir, "service.go")
	content := `package mock
// @sk-task T1.1: Implementation (AC-001)
func Do() {}

// @sk-test T1.1: TestDo (AC-001)
func TestDo() {}
`
	if err := os.WriteFile(mockFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a file to skip
	skipDir := filepath.Join(tmpDir, "node_modules")
	if err := os.Mkdir(skipDir, 0o755); err != nil {
		t.Fatal(err)
	}
	skipFile := filepath.Join(skipDir, "skipped.js")
	if err := os.WriteFile(skipFile, []byte("// @sk-task T9.9: Should skip"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := Scan(tmpDir)
	if err != nil {
		t.Errorf("Scan() error = %v", err)
		return
	}

	if len(result.Findings) != 2 {
		t.Errorf("Scan() found %d findings, want 2", len(result.Findings))
	}

	foundTask := false
	foundTest := false
	for _, f := range result.Findings {
		if f.Type == "task" && f.TaskID == "T1.1" && f.ACID == "AC-001" {
			foundTask = true
		}
		if f.Type == "test" && f.TaskID == "T1.1" && f.ACID == "AC-001" {
			foundTest = true
		}
	}

	if !foundTask {
		t.Errorf("Scan() did not find the task annotation")
	}
	if !foundTest {
		t.Errorf("Scan() did not find the test annotation")
	}
}

func TestFilterBySlug(t *testing.T) {
	findings := []Finding{
		{Type: "task", TaskID: "T1.1", ACID: "AC-001"},
		{Type: "task", TaskID: "T2.1", ACID: "AC-002"},
		{Type: "test", TaskID: "T1.1", ACID: "AC-001"},
	}

	taskIDs := map[string]struct{}{
		"T1.1": {},
	}

	filtered := FilterBySlug(findings, taskIDs)

	if len(filtered) != 2 {
		t.Errorf("FilterBySlug() len = %d, want 2", len(filtered))
	}

	for _, f := range filtered {
		if f.TaskID != "T1.1" {
			t.Errorf("FilterBySlug() included wrong task ID: %s", f.TaskID)
		}
	}
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"node_modules", true},
		{".git", true},
		{".speckeep", true},
		{"src", false},
		{"vendor", true},
		{".hidden", true},
	}

	for _, tt := range tests {
		if got := shouldSkip(tt.path); got != tt.want {
			t.Errorf("shouldSkip(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
