package executor

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// DockerOptions configures how scripts are executed inside Docker.
//
// Exactly one of Image or Container must be set:
//   - Image: runs a fresh container via "docker run --rm -i".
//   - Container: executes inside an already-running container via "docker exec -i".
type DockerOptions struct {
	// Image is the Docker image to use with "docker run". Mutually exclusive with Container.
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// Container is the name or ID of a running container for "docker exec".
	// Mutually exclusive with Image.
	Container string `json:"container,omitempty" yaml:"container,omitempty"`
	// WorkDir sets the working directory inside the container (-w).
	WorkDir string `json:"workDir,omitempty" yaml:"work-dir,omitempty"`
	// Volumes is a list of volume mount specs passed as -v flags (e.g. "/host:/container").
	// Only used in "docker run" mode.
	Volumes []string `json:"volumes,omitempty" yaml:"volumes,omitempty"`
	// Network sets the network mode for the container (--network).
	// Only used in "docker run" mode.
	Network string `json:"network,omitempty" yaml:"network,omitempty"`
	// Shell is the shell binary to invoke inside the container (default: sh).
	Shell string `json:"shell,omitempty" yaml:"shell,omitempty"`
}

func (o *DockerOptions) shell() string {
	if o.Shell != "" {
		return o.Shell
	}
	return "sh"
}

type dockerExecutor struct {
	in      io.Reader
	out     io.Writer
	options *DockerOptions
	env     map[string]string

	cmd *exec.Cmd
}

// NewDockerExecutor creates an executor that runs scripts inside a Docker container.
// Env vars are forwarded as -e flags; scripts are piped to the container shell via stdin.
func NewDockerExecutor(env map[string]string, scripts []string, options *DockerOptions) IExecutor {
	in := bytes.NewBuffer(nil)
	for _, line := range scripts {
		if len(line) == 0 {
			continue
		}
		in.WriteString(line + "\n")
	}
	return &dockerExecutor{
		in:      in,
		out:     os.Stdout,
		options: options,
		env:     env,
	}
}

func (d *dockerExecutor) prepare(extraArgs ...string) error {
	if d.options == nil {
		return errors.Errorf("docker options are required")
	}
	if d.options.Image == "" && d.options.Container == "" {
		return errors.Errorf("docker executor requires either image or container to be set")
	}
	if d.options.Image != "" && d.options.Container != "" {
		return errors.Errorf("docker executor: image and container are mutually exclusive")
	}

	args := []string{"docker"}

	if d.options.Container != "" {
		// docker exec mode
		args = append(args, "exec", "-i")
		for k, v := range d.env {
			args = append(args, "-e", k+"="+v)
		}
		if d.options.WorkDir != "" {
			args = append(args, "-w", d.options.WorkDir)
		}
		args = append(args, d.options.Container, d.options.shell())
	} else {
		// docker run mode
		args = append(args, "run", "--rm", "-i")
		for k, v := range d.env {
			args = append(args, "-e", k+"="+v)
		}
		if d.options.WorkDir != "" {
			args = append(args, "-w", d.options.WorkDir)
		}
		for _, vol := range d.options.Volumes {
			args = append(args, "-v", vol)
		}
		if d.options.Network != "" {
			args = append(args, "--network", d.options.Network)
		}
		args = append(args, d.options.Image, d.options.shell())
	}

	args = append(args, extraArgs...)
	d.cmd = exec.Command(args[0], args[1:]...)
	d.cmd.Stdin = d.in
	d.cmd.Stdout = d.out
	d.cmd.Stderr = os.Stderr
	return nil
}

func (d *dockerExecutor) Start(args ...string) error {
	if err := d.prepare(args...); err != nil {
		return err
	}
	return d.cmd.Start()
}

func (d *dockerExecutor) StartAndWait(args ...string) error {
	if err := d.prepare(args...); err != nil {
		return err
	}
	return d.cmd.Run()
}
