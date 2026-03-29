package executor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func NewSshExecutor(env map[string]string, scripts []string, options *SshOptions) Executor {
	in := bytes.NewBuffer(nil)
	if len(env) > 0 {
		for k, v := range env {
			in.WriteString(fmt.Sprintf("export %s=%s\n", k, shellQuote(v)))
		}
	}
	if options != nil && options.RemotePath != "" {
		in.WriteString(fmt.Sprintf("cd %s\n", shellQuote(options.RemotePath)))
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
		return errors.New("invalid ssh options")
	}
	if t.options.Host == "" {
		return errors.New("host is empty")
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
		prefix = append(prefix, "-p", fmt.Sprintf("%v", t.options.Port))
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

