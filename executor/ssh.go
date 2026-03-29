package executor

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type SshOptions struct {
	Host           string `json:"host,omitempty" yaml:"host,omitempty"`
	User           string `json:"user,omitempty" yaml:"user,omitempty"`
	Port           int    `json:"port,omitempty" yaml:"port,omitempty"`
	IdentifierFile string `json:"identifierFile,omitempty" yaml:"identifier-file,omitempty"`
	RemotePath     string `json:"remotePath,omitempty" yaml:"remote-path,omitempty"`
}

type sshExecutor struct {
	Binary  string
	in      io.Reader
	out     io.Writer
	options *SshOptions

	cmd *exec.Cmd
}

func sshShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func NewSshExecutor(env map[string]string, scripts []string, options *SshOptions) IExecutor {
	in := bytes.NewBuffer(nil)
	if len(env) > 0 {
		for k, v := range env {
			in.WriteString(fmt.Sprintf("export %s=%s\n", k, sshShellQuote(v)))
		}
	}
	if options != nil && options.RemotePath != "" {
		in.WriteString(fmt.Sprintf("cd %s\n", sshShellQuote(options.RemotePath)))
	}
	for _, line := range scripts {
		if len(line) == 0 {
			continue
		}
		in.WriteString(line + "\n")
	}
	result := &sshExecutor{
		Binary:  "ssh",
		in:      in,
		out:     os.Stdout,
		options: options,
	}
	return result
}

func (t *sshExecutor) prepare(args ...string) error {
	if t.options == nil {
		return errors.Errorf("invalid ssh options")
	}
	if t.options.Host == "" {
		return errors.Errorf("host is empty")
	}
	host := t.options.Host
	if t.options.User != "" {
		host = t.options.User + "@" + host
	}
	prefix := make([]string, 0)
	if t.options.IdentifierFile != "" {
		prefix = append(prefix, "-i", t.options.IdentifierFile)
	}
	if t.options.Port > 0 {
		prefix = append(prefix, "-P", fmt.Sprintf("%v", t.options.Port))
	}
	prefix = append(prefix, host, "sh")
	t.cmd = exec.Command(t.Binary, append(prefix, args...)...)
	t.cmd.Stdin = t.in
	t.cmd.Stdout = t.out
	t.cmd.Stderr = os.Stderr
	return nil
}

func (t *sshExecutor) Start(args ...string) error {
	err := t.prepare(args...)
	if err != nil {
		return err
	}
	return t.cmd.Start()
}

func (t *sshExecutor) StartAndWait(args ...string) error {
	err := t.prepare(args...)
	if err != nil {
		return err
	}

	return t.cmd.Run()
}

func (t *sshExecutor) Wait() error {
	if t.cmd == nil {
		return errors.New("no prepare")
	}
	return t.cmd.Wait()
}
