package task

import (
	"testing"
)

func TestContext_GetCurrentEnv_nil(t *testing.T) {
	var c *Context
	if got := c.GetCurrentEnv(); got != "" {
		t.Errorf("nil context GetCurrentEnv() = %q, want \"\"", got)
	}
}

func TestContext_GetCurrentEnv_noParent(t *testing.T) {
	c := &Context{Env: "prod"}
	if got := c.GetCurrentEnv(); got != "prod" {
		t.Errorf("GetCurrentEnv() = %q, want \"prod\"", got)
	}
}

func TestContext_GetCurrentEnv_withParent(t *testing.T) {
	parent := &Context{Env: "parent-env"}
	child := &Context{Parent: parent, Env: "child-env"}
	// GetCurrentEnv traverses to the root parent.
	if got := child.GetCurrentEnv(); got != "parent-env" {
		t.Errorf("GetCurrentEnv() = %q, want \"parent-env\"", got)
	}
}

func TestContext_UseEnv_nil(t *testing.T) {
	var c *Context
	if got := c.UseEnv("anything"); got != nil {
		t.Errorf("UseEnv on nil context should return nil, got %v", got)
	}
}

func TestContext_UseEnv_setsEnv(t *testing.T) {
	c := &Context{
		Env:         "old",
		EnvProvider: EnvProvider{},
	}
	result := c.UseEnv("new")
	if result == nil {
		t.Fatal("UseEnv returned nil")
	}
	if result.Env != "new" {
		t.Errorf("UseEnv result.Env = %q, want \"new\"", result.Env)
	}
	// Original must be unchanged.
	if c.Env != "old" {
		t.Error("UseEnv mutated the original context")
	}
}

func TestContext_Clone_nil(t *testing.T) {
	var c *Context
	if got := c.Clone(); got != nil {
		t.Errorf("Clone of nil should be nil, got %v", got)
	}
}

func TestContext_Clone_deep(t *testing.T) {
	orig := &Context{
		Env:         "myenv",
		EnvProvider: EnvProvider{"prod": Environment{"HOST": "prod.host"}},
	}
	clone := orig.Clone()
	if clone.Env != orig.Env {
		t.Errorf("Clone.Env = %q, want %q", clone.Env, orig.Env)
	}
	// Modifying clone's provider must not affect original.
	clone.EnvProvider["dev"] = Environment{"HOST": "dev.host"}
	if _, ok := orig.EnvProvider["dev"]; ok {
		t.Error("modifying clone's EnvProvider affected original")
	}
}

func TestContext_GetEnv_found(t *testing.T) {
	c := &Context{
		EnvProvider: EnvProvider{
			"prod": Environment{"HOST": "prod.example.com"},
		},
	}
	env, found := c.GetEnv("prod")
	if !found {
		t.Fatal("GetEnv('prod') should find env")
	}
	if env["HOST"] != "prod.example.com" {
		t.Errorf("env HOST = %q, want \"prod.example.com\"", env["HOST"])
	}
}

func TestContext_GetEnv_notFound(t *testing.T) {
	c := &Context{EnvProvider: EnvProvider{}}
	_, found := c.GetEnv("missing")
	if found {
		t.Error("GetEnv('missing') should return found=false")
	}
}

func TestContext_GetEnv_parentFallback(t *testing.T) {
	parent := &Context{
		EnvProvider: EnvProvider{
			"shared": Environment{"KEY": "from-parent"},
		},
	}
	child := &Context{
		Parent:      parent,
		EnvProvider: EnvProvider{},
	}
	env, found := child.GetEnv("shared")
	if !found {
		t.Fatal("GetEnv should fall back to parent")
	}
	if env["KEY"] != "from-parent" {
		t.Errorf("env KEY = %q, want \"from-parent\"", env["KEY"])
	}
}

func TestContext_GetEnv_childOverridesParent(t *testing.T) {
	parent := &Context{
		EnvProvider: EnvProvider{
			"env1": Environment{"KEY": "parent-value"},
		},
	}
	child := &Context{
		Parent: parent,
		EnvProvider: EnvProvider{
			"env1": Environment{"KEY": "child-value"},
		},
	}
	env, found := child.GetEnv("env1")
	if !found {
		t.Fatal("GetEnv should find env1")
	}
	if env["KEY"] != "child-value" {
		t.Errorf("child should override parent, got %q", env["KEY"])
	}
}
