package task

import (
	"sync"
	"testing"
)

func TestFilter_AddAndContains(t *testing.T) {
	f := NewFilter()
	if f.Contains("dep/a") {
		t.Error("fresh filter should not contain 'dep/a'")
	}
	f.Add("dep/a")
	if !f.Contains("dep/a") {
		t.Error("filter should contain 'dep/a' after Add")
	}
	if f.Contains("dep/b") {
		t.Error("filter should not contain 'dep/b'")
	}
}

func TestFilter_Idempotent(t *testing.T) {
	f := NewFilter()
	f.Add("x")
	f.Add("x")
	if !f.Contains("x") {
		t.Error("repeated Add should not break Contains")
	}
}

func TestFilter_ConcurrentSafe(t *testing.T) {
	f := NewFilter()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		key := "key"
		go func() {
			defer wg.Done()
			f.Add(key)
		}()
		go func() {
			defer wg.Done()
			f.Contains(key)
		}()
	}
	wg.Wait()
}

func TestNewFilter_notNil(t *testing.T) {
	f := NewFilter()
	if f == nil {
		t.Error("NewFilter should not return nil")
	}
}
