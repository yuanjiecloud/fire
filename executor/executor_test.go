package executor

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// shellQuote
// ---------------------------------------------------------------------------

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "'hello'"},
		{"", "''"},
		{"with space", "'with space'"},
		{"it's", "'it'\\''s'"},
		{"a'b'c", "'a'\\''b'\\''c'"},
		{"no-special_chars.123", "'no-special_chars.123'"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := shellQuote(tc.input)
			if got != tc.want {
				t.Errorf("shellQuote(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// replacePlaceholder
// ---------------------------------------------------------------------------

func TestReplacePlaceholder(t *testing.T) {
	tests := []struct {
		s, ph, val, want string
	}{
		{"echo {}", "{}", "hello", "echo hello"},
		{"cp {} {}.bak", "{}", "file.txt", "cp file.txt file.txt.bak"},
		{"no placeholder", "{}", "x", "no placeholder"},
		{"", "{}", "x", ""},
		{"{{}} test", "{{}}", "X", "X test"},
		{"a{}b{}c", "{}", "Z", "aZbZc"},
	}
	for _, tc := range tests {
		t.Run(tc.s, func(t *testing.T) {
			got := replacePlaceholder(tc.s, tc.ph, tc.val)
			if got != tc.want {
				t.Errorf("replacePlaceholder(%q,%q,%q) = %q, want %q",
					tc.s, tc.ph, tc.val, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// New factory
// ---------------------------------------------------------------------------

func TestNew_knownTypes(t *testing.T) {
	types := []Type{TypeBash, TypeSh}
	for _, typ := range types {
		t.Run(string(typ), func(t *testing.T) {
			ex, err := New(typ, nil, nil, Options{})
			if err != nil {
				t.Fatalf("New(%q) returned error: %v", typ, err)
			}
			if ex == nil {
				t.Fatalf("New(%q) returned nil executor", typ)
			}
		})
	}
}

func TestNew_unknownType(t *testing.T) {
	_, err := New(Type("nonexistent"), nil, nil, Options{})
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error message %q should mention the unknown type", err.Error())
	}
}

func TestNew_sshMissingOptions(t *testing.T) {
	ex, err := New(TypeSsh, nil, nil, Options{SSH: nil})
	if err != nil {
		t.Fatalf("New(TypeSsh) construction error: %v", err)
	}
	// Calling Start/StartAndWait with nil options should return an error.
	err = ex.StartAndWait()
	if err == nil {
		t.Fatal("expected error when SSH options are nil")
	}
}

func TestNew_sshMissingHost(t *testing.T) {
	ex, err := New(TypeSsh, nil, nil, Options{SSH: &SshOptions{Host: ""}})
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	err = ex.StartAndWait()
	if err == nil {
		t.Fatal("expected error for empty host")
	}
	if !strings.Contains(err.Error(), "host is empty") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNew_dockerMissingOptions(t *testing.T) {
	ex, err := New(TypeDocker, nil, nil, Options{Docker: nil})
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	err = ex.StartAndWait()
	if err == nil {
		t.Fatal("expected error when docker options are nil")
	}
}

func TestNew_dockerMissingImageAndContainer(t *testing.T) {
	ex, err := New(TypeDocker, nil, nil, Options{Docker: &DockerOptions{}})
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	err = ex.StartAndWait()
	if err == nil {
		t.Fatal("expected error when neither image nor container is set")
	}
}

func TestNew_dockerBothImageAndContainer(t *testing.T) {
	ex, err := New(TypeDocker, nil, nil, Options{
		Docker: &DockerOptions{Image: "img", Container: "cnt"},
	})
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	err = ex.StartAndWait()
	if err == nil {
		t.Fatal("expected error when both image and container are set")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestNew_batchMissingOptions(t *testing.T) {
	ex, err := New(TypeBatch, nil, nil, Options{Batch: nil})
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	err = ex.StartAndWait()
	if err == nil {
		t.Fatal("expected error when batch options are nil")
	}
}

func TestNew_batchEmptyItems(t *testing.T) {
	ex, err := New(TypeBatch, nil, []string{"echo hi"}, Options{
		Batch: &BatchOptions{Items: []string{}},
	})
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	// Empty items should succeed immediately (no-op).
	if err = ex.StartAndWait(); err != nil {
		t.Errorf("empty items should not error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// BatchOptions helpers
// ---------------------------------------------------------------------------

func TestBatchOptions_placeholder_default(t *testing.T) {
	o := &BatchOptions{}
	if o.placeholder() != "{}" {
		t.Errorf("default placeholder should be '{}', got %q", o.placeholder())
	}
}

func TestBatchOptions_placeholder_custom(t *testing.T) {
	o := &BatchOptions{Placeholder: "%%"}
	if o.placeholder() != "%%" {
		t.Errorf("custom placeholder should be '%%', got %q", o.placeholder())
	}
}

func TestBatchOptions_executorType_default(t *testing.T) {
	o := &BatchOptions{}
	if o.executorType() != TypeBash {
		t.Errorf("default executor type should be bash, got %q", o.executorType())
	}
}

func TestBatchOptions_executorType_custom(t *testing.T) {
	o := &BatchOptions{Executor: TypeSh}
	if o.executorType() != TypeSh {
		t.Errorf("custom executor type should be sh, got %q", o.executorType())
	}
}

// ---------------------------------------------------------------------------
// DockerOptions.shell helper
// ---------------------------------------------------------------------------

func TestDockerOptions_shell_default(t *testing.T) {
	o := &DockerOptions{}
	if o.shell() != "sh" {
		t.Errorf("default shell should be 'sh', got %q", o.shell())
	}
}

func TestDockerOptions_shell_custom(t *testing.T) {
	o := &DockerOptions{Shell: "/bin/bash"}
	if o.shell() != "/bin/bash" {
		t.Errorf("custom shell should be '/bin/bash', got %q", o.shell())
	}
}

// ---------------------------------------------------------------------------
// SshOptions port flag direction (verify -p flag is used, not -P)
// ---------------------------------------------------------------------------

func TestSshExecutor_portFlagIsLowercase(t *testing.T) {
	ex := NewSshExecutor(nil, nil, &SshOptions{
		Host: "testhost",
		Port: 2222,
	})
	se, ok := ex.(*sshExecutor)
	if !ok {
		t.Fatal("expected *sshExecutor")
	}
	if err := se.prepare(); err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	args := se.cmd.Args
	for i, a := range args {
		if a == "-P" {
			t.Errorf("found uppercase -P flag at index %d (should be lowercase -p): %v", i, args)
		}
	}
	found := false
	for i, a := range args {
		if a == "-p" && i+1 < len(args) && args[i+1] == "2222" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("could not find '-p 2222' in ssh args: %v", args)
	}
}

// ---------------------------------------------------------------------------
// Batch sequential execution (uses real bash — requires /bin/bash)
// ---------------------------------------------------------------------------

func TestBatchExecutor_sequential(t *testing.T) {
	opts := &BatchOptions{
		Items:    []string{"alpha", "beta"},
		Parallel: false,
	}
	ex := NewBatchExecutor(nil, []string{"echo {}"}, opts)
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("sequential batch failed: %v", err)
	}
}

func TestBatchExecutor_parallel(t *testing.T) {
	opts := &BatchOptions{
		Items:       []string{"one", "two", "three"},
		Parallel:    true,
		MaxParallel: 2,
	}
	ex := NewBatchExecutor(nil, []string{"echo {}"}, opts)
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("parallel batch failed: %v", err)
	}
}

func TestBatchExecutor_customPlaceholder(t *testing.T) {
	opts := &BatchOptions{
		Items:       []string{"world"},
		Placeholder: "%%",
	}
	ex := NewBatchExecutor(nil, []string{"echo %%"}, opts)
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("custom placeholder batch failed: %v", err)
	}
}

func TestBatchExecutor_envInjected(t *testing.T) {
	opts := &BatchOptions{Items: []string{"X"}}
	ex := NewBatchExecutor(
		map[string]string{"MY_VAR": "hello"},
		[]string{"test \"$MY_VAR\" = hello", "test \"$FIRE_BATCH_ITEM\" = X"},
		opts,
	)
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("env injection test failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// BashExecutor runs a real script
// ---------------------------------------------------------------------------

func TestBashExecutor_runScript(t *testing.T) {
	ex := NewBashExecutor(nil, []string{"exit 0"})
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("bash executor exit 0 failed: %v", err)
	}
}

func TestBashExecutor_envExported(t *testing.T) {
	ex := NewBashExecutor(
		map[string]string{"FIRE_TEST_VAR": "fire_value"},
		[]string{`test "$FIRE_TEST_VAR" = fire_value`},
	)
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("bash env export failed: %v", err)
	}
}

func TestBashExecutor_envSpecialChars(t *testing.T) {
	ex := NewBashExecutor(
		map[string]string{"MSG": "hello world & 'quotes'"},
		[]string{`test "$MSG" = "hello world & 'quotes'"`},
	)
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("bash env special chars failed: %v", err)
	}
}

func TestShExecutor_runScript(t *testing.T) {
	ex := NewShExecutor(nil, []string{"exit 0"})
	if err := ex.StartAndWait(); err != nil {
		t.Errorf("sh executor exit 0 failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// batchError helper
// ---------------------------------------------------------------------------

func TestBatchError_noErrors(t *testing.T) {
	var be batchError
	if be.err() != nil {
		t.Error("empty batchError should return nil")
	}
}

func TestBatchError_multipleErrors(t *testing.T) {
	var be batchError
	be.add("item1", errStr("fail1"))
	be.add("item2", errStr("fail2"))
	err := be.err()
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "2 item(s) failed") {
		t.Errorf("error message should mention count, got: %q", msg)
	}
	if !strings.Contains(msg, "item1") || !strings.Contains(msg, "item2") {
		t.Errorf("error message should contain item names, got: %q", msg)
	}
}

// errStr is a minimal error for tests.
type errStr string

func (e errStr) Error() string { return string(e) }
