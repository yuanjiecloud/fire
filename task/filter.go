package task

import "sync"

// Filter is a concurrency-safe set of strings used to deduplicate dependency
// resolution across goroutines.
type Filter struct {
	mu   sync.RWMutex
	seen map[string]bool
}

var globalResolverFilter = NewFilter()

func NewFilter() *Filter {
	return &Filter{seen: make(map[string]bool)}
}

func (f *Filter) Add(depend string) {
	f.mu.Lock()
	f.seen[depend] = true
	f.mu.Unlock()
}

func (f *Filter) Contains(depend string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.seen[depend]
}
