package executor

import "github.com/pkg/errors"

type Type string

const (
	TypeBash = Type("bash")
	TypeSh   = Type("sh")
	TypeSsh  = Type("ssh")
)

type IExecutor interface {
	Start(args ...string) error
	StartAndWait(args ...string) error
}

// New creates an executor for the given type. env is the set of environment
// variables to export before running scripts. opts is only used when t == TypeSsh.
func New(t Type, env map[string]string, scripts []string, opts *SshOptions) (IExecutor, error) {
	switch t {
	case TypeBash:
		return NewBashExecutor(env, scripts), nil
	case TypeSh:
		return NewShExecutor(env, scripts), nil
	case TypeSsh:
		return NewSshExecutor(env, scripts, opts), nil
	default:
		return nil, errors.Errorf("unknown executor type: %s", t)
	}
}
