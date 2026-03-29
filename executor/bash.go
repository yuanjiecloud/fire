package executor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type bashExecutor struct {
	Binary string
	in     io.Reader
	out    io.Writer

	cmd *exec.Cmd
}

func NewBashExecutor(env map[string]string, scripts []string) IExecutor {
	in := bytes.NewBuffer(nil)
	if len(env) > 0 {
		for k, v := range env {
			in.WriteString(fmt.Sprintf("export %s=%s\n", k, shellQuote(v)))
		}
	}
	for _, line := range scripts {
		if len(line) == 0 {
			continue
		}
		in.WriteString(line + "\n")
	}
	result := &bashExecutor{
		Binary: "/bin/bash",
		in:     in,
		out:    os.Stdout,
	}
	return result
}

func (t *bashExecutor) prepare(args ...string) error {
	t.cmd = exec.Command(t.Binary, args...)
	t.cmd.Stdin = t.in
	t.cmd.Stdout = t.out
	t.cmd.Stderr = os.Stderr
	return nil
}

func (t *bashExecutor) Start(args ...string) error {
	err := t.prepare(args...)
	if err != nil {
		return err
	}
	return t.cmd.Start()
}

func (t *bashExecutor) StartAndWait(args ...string) error {
	err := t.prepare(args...)
	if err != nil {
		return err
	}
	return t.cmd.Run()
}

func (t *bashExecutor) Wait() error {
	if t.cmd == nil {
		return errors.New("no prepare")
	}
	return t.cmd.Wait()
}

// shellQuote wraps a value in single quotes, escaping any single quotes within.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
