package task

import (
	"github.com/pkg/errors"
	"github.com/yuanjiecloud/fire/executor"
)

type ExecutorType string

const (
	ExecutorTypeBash = ExecutorType("bash")
	ExecutorTypeSsh  = ExecutorType("ssh")
)

type Task struct {
	Name         string        `json:"name,omitempty" yaml:"name,omitempty"`
	Environments Environment   `json:"environments,omitempty" yaml:"environments,omitempty"`
	Type         executor.Type `json:"type,omitempty" yaml:"type,omitempty"`
	Env          string        `json:"env,omitempty" yaml:"env,omitempty"`
	Pipeline     string        `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`
	Scripts      []string      `json:"scripts,omitempty" yaml:"scripts,omitempty"`
}

func (t *Task) Exec(ctx *Context) error {
	var err error
	err = t.runPipeline(ctx)
	if err != nil {
		return err
	}
	err = t.runScripts(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (t *Task) runPipeline(ctx *Context) error {
	if t.Pipeline == "" {
		return nil
	}
	return nil
}

func (t *Task) getCurrentEnv(ctx *Context) (result Environment, found bool) {
	if t.Environments != nil {
		result = t.Environments
	} else {
		result = make(Environment)
	}
	if ctx.Env != "" {
		env, b := ctx.GetEnv(ctx.Env)
		if b {
			result = result.MergeIgnoreDuplicated(env)
			found = true
		}
	}
	if t.Env != "" {
		env, b := ctx.GetEnv(t.Env)
		if b {
			result = result.MergeIgnoreDuplicated(env)
			found = true
		}
	}
	return
}

func (t *Task) runScripts(ctx *Context) (err error) {
	if len(t.Scripts) == 0 {
		return nil
	}
	env, found := t.getCurrentEnv(ctx)
	if !found {
		return errors.Errorf("unset env")
	}
	if t.Type == executor.TypeBash {
		return executor.NewBashExecutor(env, t.Scripts).StartAndWait()
	} else {
		return errors.Errorf("unknown executor type: %v", t.Type)
	}
}
