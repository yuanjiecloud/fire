package task

import "github.com/yuanjiecloud/fire/log"

type Context struct {
	Parent      *Context
	EnvProvider EnvProvider
	Env         string
}

func (t *Context) Clone() *Context {
	if t == nil {
		return nil
	}
	return &Context{
		Parent:      t.Parent,
		EnvProvider: t.EnvProvider.Clone(),
		Env:         t.Env,
	}
}

func (t *Context) GetEnv(name string) (env Environment, found bool) {
	log.Debug("query: ", name)
	env, found = t.EnvProvider[name]
	if !found && t.Parent != nil {
		return t.Parent.GetEnv(name)
	}
	return
}

func (t *Context) GetCurrentEnv() string {
	if t == nil {
		return ""
	}
	if t.Parent != nil {
		return t.Parent.GetCurrentEnv()
	}
	return t.Env
}

func (t *Context) UseEnv(env string) *Context {
	if t == nil {
		return nil
	}
	result := t.Clone()
	result.Env = env
	return result
}
