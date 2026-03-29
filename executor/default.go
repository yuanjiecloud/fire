package executor

import "fmt"

type Type string

const (
	TypeBash   = Type("bash")
	TypeSh     = Type("sh")
	TypeSsh    = Type("ssh")
	TypeDocker = Type("docker")
	TypeBatch  = Type("batch")
)

type Executor interface {
	Start(args ...string) error
	StartAndWait(args ...string) error
}

// Options holds optional configuration for executor types that need it.
// Only the field matching the chosen Type is used.
type Options struct {
	SSH    *SshOptions
	Docker *DockerOptions
	Batch  *BatchOptions
}

// New creates an executor for the given type.
// env is the set of environment variables to forward into the execution context.
// opts carries type-specific configuration (SSH or Docker); unused fields are ignored.
func New(t Type, env map[string]string, scripts []string, opts Options) (Executor, error) {
	switch t {
	case TypeBash:
		return NewBashExecutor(env, scripts), nil
	case TypeSh:
		return NewShExecutor(env, scripts), nil
	case TypeSsh:
		return NewSshExecutor(env, scripts, opts.SSH), nil
	case TypeDocker:
		return NewDockerExecutor(env, scripts, opts.Docker), nil
	case TypeBatch:
		return NewBatchExecutor(env, scripts, opts.Batch), nil
	default:
		return nil, fmt.Errorf("unknown executor type: %s", t)
	}
}
