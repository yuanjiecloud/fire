package task

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// SplitPackageName
// ---------------------------------------------------------------------------

func TestSplitPackageName_valid(t *testing.T) {
	tests := []struct {
		input          string
		wantNS, wantN  string
		wantVer        string
	}{
		{"org/repo", "org", "repo", ""},
		{"org/repo@v1.2.3", "org", "repo", "v1.2.3"},
		{"org/repo@main", "org", "repo", "main"},
		{"firerepos/hello@v1.0.0", "firerepos", "hello", "v1.0.0"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			ns, name, ver, err := SplitPackageName(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ns != tc.wantNS {
				t.Errorf("namespace = %q, want %q", ns, tc.wantNS)
			}
			if name != tc.wantN {
				t.Errorf("name = %q, want %q", name, tc.wantN)
			}
			if ver != tc.wantVer {
				t.Errorf("version = %q, want %q", ver, tc.wantVer)
			}
		})
	}
}

func TestSplitPackageName_invalid(t *testing.T) {
	tests := []struct {
		input   string
		errSnip string
	}{
		{"", "invalid package"},
		{"nonamespace", "expected namespace/name"},
		{"a/b/c", "expected namespace/name"},
		{"org/repo@", "version after '@' is empty"},
		{"/repo", "namespace and name must not be empty"},
		{"org/", "namespace and name must not be empty"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, _, _, err := SplitPackageName(tc.input)
			if err == nil {
				t.Fatalf("expected error for %q, got nil", tc.input)
			}
			if !strings.Contains(err.Error(), tc.errSnip) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.errSnip)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// CheckIfExists
// ---------------------------------------------------------------------------

func TestCheckIfExists(t *testing.T) {
	dir := t.TempDir()
	if !CheckIfExists(dir) {
		t.Errorf("CheckIfExists(%q) should be true for existing dir", dir)
	}
	if CheckIfExists(filepath.Join(dir, "nonexistent")) {
		t.Error("CheckIfExists should be false for non-existent path")
	}
}

// ---------------------------------------------------------------------------
// CheckIfGitRepository
// ---------------------------------------------------------------------------

func TestCheckIfGitRepository(t *testing.T) {
	dir := t.TempDir()
	if CheckIfGitRepository(dir) {
		t.Error("empty dir should not be a git repo")
	}
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if !CheckIfGitRepository(dir) {
		t.Error("dir with .git should be a git repo")
	}
}

// ---------------------------------------------------------------------------
// CheckIfNestedRepository
// ---------------------------------------------------------------------------

func TestCheckIfNestedRepository(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "child")
	if err := os.Mkdir(child, 0755); err != nil {
		t.Fatal(err)
	}
	// No fire.yaml in parent yet.
	if CheckIfNestedRepository(child) {
		t.Error("should not be nested without fire.yaml in parent")
	}
	// Create fire.yaml in parent.
	f, err := os.Create(filepath.Join(parent, DefaultConfigFile))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	if !CheckIfNestedRepository(child) {
		t.Error("should be nested when parent has fire.yaml")
	}
}

// ---------------------------------------------------------------------------
// Parse
// ---------------------------------------------------------------------------

func TestParse_validYAML(t *testing.T) {
	dir := t.TempDir()
	content := `
version: v0.1.0
name: test-pipeline
tasks:
  - name: hello
    type: bash
    scripts:
      - echo hello
`
	file := filepath.Join(dir, "fire.yaml")
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := Parse(file)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(p.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(p.Tasks))
	}
	if p.Tasks[0].Name != "hello" {
		t.Errorf("task name = %q, want \"hello\"", p.Tasks[0].Name)
	}
}

func TestParse_nonexistentFile(t *testing.T) {
	_, err := Parse("/nonexistent/path/fire.yaml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestParse_invalidYAML(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "fire.yaml")
	if err := os.WriteFile(file, []byte(":\tinvalid:\tyaml:"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Parse(file)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestParse_storesConfigfile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "fire.yaml")
	if err := os.WriteFile(file, []byte("version: v1\n"), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := Parse(file)
	if err != nil {
		t.Fatal(err)
	}
	if p.Getwd() != dir {
		t.Errorf("Getwd() = %q, want %q", p.Getwd(), dir)
	}
}

// ---------------------------------------------------------------------------
// Pipeline.ToJson
// ---------------------------------------------------------------------------

func TestPipeline_ToJson(t *testing.T) {
	p := &Pipeline{Version: "v0.1.0"}
	j := p.ToJson()
	if !strings.Contains(j, "v0.1.0") {
		t.Errorf("ToJson() = %q, expected to contain version", j)
	}
}

// ---------------------------------------------------------------------------
// Pipeline.FindTask
// ---------------------------------------------------------------------------

func TestPipeline_FindTask(t *testing.T) {
	p := &Pipeline{
		Tasks: []Task{
			{Name: "build"},
			{Name: "test"},
		},
	}
	task, found := p.FindTask("test")
	if !found {
		t.Fatal("FindTask('test') should be found")
	}
	if task.Name != "test" {
		t.Errorf("task.Name = %q, want \"test\"", task.Name)
	}
	_, found = p.FindTask("nonexistent")
	if found {
		t.Error("FindTask('nonexistent') should return found=false")
	}
}

// ---------------------------------------------------------------------------
// Pipeline.GetAllowTaskList
// ---------------------------------------------------------------------------

func TestPipeline_GetAllowTaskList_nil(t *testing.T) {
	var p *Pipeline
	list := p.GetAllowTaskList()
	if len(list) != 0 {
		t.Errorf("nil pipeline should return empty list, got %v", list)
	}
}

func TestPipeline_GetAllowTaskList_sorted(t *testing.T) {
	p := &Pipeline{
		Tasks: []Task{
			{Name: "zebra"},
			{Name: "apple"},
			{Name: "mango"},
		},
	}
	list := p.GetAllowTaskList()
	if len(list) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(list))
	}
	if list[0] != "apple" || list[1] != "mango" || list[2] != "zebra" {
		t.Errorf("list not sorted: %v", list)
	}
}

func TestPipeline_GetAllowTaskList_deduplication(t *testing.T) {
	p := &Pipeline{
		Tasks: []Task{
			{Name: "build"},
			{Name: "build"}, // duplicate
			{Name: "test"},
		},
	}
	list := p.GetAllowTaskList()
	if len(list) != 2 {
		t.Errorf("expected 2 unique tasks, got %d: %v", len(list), list)
	}
}

// ---------------------------------------------------------------------------
// Pipeline.RunAll (sequential) with real bash tasks
// ---------------------------------------------------------------------------

// taskEnv is a convenience named environment for tests so that runScripts
// finds a matching env and doesn't return "unset env".
var taskEnvProvider = EnvProvider{"default": Environment{}}

func TestPipeline_RunAll_sequential(t *testing.T) {
	p := &Pipeline{
		Environments: taskEnvProvider,
		Tasks: []Task{
			{Name: "t1", Type: "bash", Env: "default", Scripts: []string{"exit 0"}},
			{Name: "t2", Type: "bash", Env: "default", Scripts: []string{"exit 0"}},
		},
	}
	if err := p.RunAll(nil); err != nil {
		t.Errorf("RunAll sequential failed: %v", err)
	}
}

func TestPipeline_RunAll_noScripts(t *testing.T) {
	// Tasks with no scripts don't need a named env.
	p := &Pipeline{
		Tasks: []Task{
			{Name: "noop1"},
			{Name: "noop2"},
		},
	}
	if err := p.RunAll(nil); err != nil {
		t.Errorf("RunAll with no-script tasks failed: %v", err)
	}
}

func TestPipeline_RunAll_parallel(t *testing.T) {
	p := &Pipeline{
		Parallel:     true,
		Environments: taskEnvProvider,
		Tasks: []Task{
			{Name: "p1", Type: "bash", Env: "default", Scripts: []string{"exit 0"}},
			{Name: "p2", Type: "bash", Env: "default", Scripts: []string{"exit 0"}},
		},
	}
	if err := p.RunAll(nil); err != nil {
		t.Errorf("RunAll parallel failed: %v", err)
	}
}

func TestPipeline_RunAll_parallelCollectsErrors(t *testing.T) {
	p := &Pipeline{
		Parallel:     true,
		Environments: taskEnvProvider,
		Tasks: []Task{
			{Name: "ok", Type: "bash", Env: "default", Scripts: []string{"exit 0"}},
			{Name: "fail", Type: "bash", Env: "default", Scripts: []string{"exit 1"}},
		},
	}
	err := p.RunAll(nil)
	if err == nil {
		t.Fatal("expected error from failing parallel task")
	}
	if !strings.Contains(err.Error(), "1 task(s) failed") {
		t.Errorf("error message should mention failure count, got: %q", err.Error())
	}
}

func TestPipeline_RunTask_found(t *testing.T) {
	p := &Pipeline{
		Environments: taskEnvProvider,
		Tasks: []Task{
			{Name: "greet", Type: "bash", Env: "default", Scripts: []string{"echo hello"}},
		},
	}
	if err := p.RunTask("greet", nil); err != nil {
		t.Errorf("RunTask failed: %v", err)
	}
}

func TestPipeline_RunTask_notFound(t *testing.T) {
	p := &Pipeline{}
	err := p.RunTask("missing", nil)
	if err == nil {
		t.Fatal("expected error for missing task")
	}
	if !strings.Contains(err.Error(), "missing") {
		t.Errorf("error should mention task name, got: %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// Pipeline.CreateContext
// ---------------------------------------------------------------------------

func TestPipeline_CreateContext_noParent(t *testing.T) {
	p := &Pipeline{
		Environments: EnvProvider{
			"prod": Environment{"HOST": "prod.host"},
		},
	}
	ctx := p.CreateContext(nil)
	if ctx == nil {
		t.Fatal("CreateContext returned nil")
	}
	env, found := ctx.GetEnv("prod")
	if !found {
		t.Fatal("context should have 'prod' env")
	}
	if env["HOST"] != "prod.host" {
		t.Errorf("HOST = %q, want \"prod.host\"", env["HOST"])
	}
}

func TestPipeline_CreateContext_withParent(t *testing.T) {
	parent := &Context{
		EnvProvider: EnvProvider{
			"shared": Environment{"KEY": "parent-val"},
		},
	}
	p := &Pipeline{
		Environments: EnvProvider{
			"local": Environment{"KEY": "local-val"},
		},
	}
	ctx := p.CreateContext(parent)
	if ctx.Parent == nil {
		t.Fatal("created context should have parent")
	}
	// 'local' env should come from the pipeline itself.
	env, found := ctx.GetEnv("local")
	if !found {
		t.Fatal("'local' env not found")
	}
	if env["KEY"] != "local-val" {
		t.Errorf("KEY = %q, want \"local-val\"", env["KEY"])
	}
}
