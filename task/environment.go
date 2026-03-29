package task

type Environment map[string]string

func (t Environment) Clone() Environment {
	result := make(Environment)
	for k, v := range t {
		result[k] = v
	}
	return result
}

func (t Environment) OverridePatch(env Environment) (result Environment) {
	result = t.Clone()
	for k, v := range env {
		result[k] = v
	}
	return result
}

func (t Environment) MergeKeepExisting(env Environment) (result Environment) {
	result = t.Clone()
	for k, v := range env {
		if _, b := result[k]; !b {
			result[k] = v
		}
	}
	return
}

type EnvProvider map[string]Environment

func NewEnvProvider() EnvProvider {
	result := make(EnvProvider)
	return result
}

func (t EnvProvider) Clone() EnvProvider {
	if t == nil {
		return NewEnvProvider()
	}
	result := NewEnvProvider()
	for k, v := range t {
		result[k] = v
	}
	return result
}

func (t EnvProvider) OverridePatch(provider EnvProvider) EnvProvider {
	result := t.Clone()
	for k, env := range provider {
		result[k] = env
	}
	return result
}

func (t EnvProvider) MergeKeepExisting(p EnvProvider) EnvProvider {
	result := t.Clone()
	for k, v := range p {
		if _, b := result[k]; !b {
			result[k] = v
		}
	}
	return result
}
