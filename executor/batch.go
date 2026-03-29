package executor

import (
	"fmt"
	"os"
	"sync"

	"github.com/pkg/errors"
)

// BatchOptions configures the batch executor.
//
// The batch executor iterates over Items, expanding the placeholder
// (default "{}") in every script line with the current item value,
// then runs all resulting script sets — either sequentially or in
// parallel depending on Parallel.
//
// Example fire.yaml:
//
//	tasks:
//	  - name: ocr-pages
//	    type: batch
//	    batch-options:
//	      items:
//	        - page_001.png
//	        - page_002.png
//	        - page_003.png
//	      parallel: true
//	      placeholder: "{}"
//	      executor: bash
//	    scripts:
//	      - tesseract {} {}.txt
//	      - echo "OCR done: {}"
type BatchOptions struct {
	// Items is the list of values to iterate over.
	Items []string `json:"items,omitempty" yaml:"items,omitempty"`

	// Parallel runs one goroutine per item when true; defaults to sequential.
	Parallel bool `json:"parallel,omitempty" yaml:"parallel,omitempty"`

	// Placeholder is the token replaced with each item value in script lines.
	// Defaults to "{}".
	Placeholder string `json:"placeholder,omitempty" yaml:"placeholder,omitempty"`

	// Executor chooses which shell to use for each item's scripts (bash, sh).
	// Defaults to bash.
	Executor Type `json:"executor,omitempty" yaml:"executor,omitempty"`

	// MaxParallel limits concurrent goroutines when Parallel is true.
	// 0 means unlimited.
	MaxParallel int `json:"maxParallel,omitempty" yaml:"max-parallel,omitempty"`
}

func (o *BatchOptions) placeholder() string {
	if o.Placeholder != "" {
		return o.Placeholder
	}
	return "{}"
}

func (o *BatchOptions) executorType() Type {
	if o.Executor != "" {
		return o.Executor
	}
	return TypeBash
}

// batchError collects per-item errors from parallel runs.
type batchError struct {
	mu   sync.Mutex
	errs []string
}

func (e *batchError) add(item string, err error) {
	e.mu.Lock()
	e.errs = append(e.errs, fmt.Sprintf("item %q: %v", item, err))
	e.mu.Unlock()
}

func (e *batchError) err() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.errs) == 0 {
		return nil
	}
	msg := fmt.Sprintf("%d item(s) failed:", len(e.errs))
	for _, s := range e.errs {
		msg += "\n  " + s
	}
	return errors.New(msg)
}

type batchExecutor struct {
	env     map[string]string
	scripts []string
	options *BatchOptions
}

// NewBatchExecutor creates an executor that runs scripts once per item in
// BatchOptions.Items, substituting the placeholder with the item value.
func NewBatchExecutor(env map[string]string, scripts []string, options *BatchOptions) IExecutor {
	return &batchExecutor{
		env:     env,
		scripts: scripts,
		options: options,
	}
}

func (b *batchExecutor) Start(args ...string) error {
	return b.run(false)
}

func (b *batchExecutor) StartAndWait(args ...string) error {
	return b.run(true)
}

func (b *batchExecutor) run(wait bool) error {
	if b.options == nil {
		return errors.New("batch options are required")
	}
	if len(b.options.Items) == 0 {
		return nil
	}

	if b.options.Parallel {
		return b.runParallel()
	}
	return b.runSequential()
}

func (b *batchExecutor) runSequential() error {
	for _, item := range b.options.Items {
		if err := b.execItem(item); err != nil {
			return fmt.Errorf("batch item %q failed: %w", item, err)
		}
	}
	return nil
}

func (b *batchExecutor) runParallel() error {
	items := b.options.Items
	maxWorkers := b.options.MaxParallel
	if maxWorkers <= 0 {
		maxWorkers = len(items)
	}

	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	var be batchError

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(it string) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := b.execItem(it); err != nil {
				be.add(it, err)
			}
		}(item)
	}
	wg.Wait()
	return be.err()
}

// execItem expands the placeholder with item in every script line and runs
// the resulting script via the configured inner executor type.
func (b *batchExecutor) execItem(item string) error {
	expanded := make([]string, 0, len(b.scripts))
	ph := b.options.placeholder()
	for _, line := range b.scripts {
		expanded = append(expanded, replacePlaceholder(line, ph, item))
	}

	// build per-item env: merge global env + FIRE_BATCH_ITEM convenience var
	env := make(map[string]string, len(b.env)+1)
	for k, v := range b.env {
		env[k] = v
	}
	env["FIRE_BATCH_ITEM"] = item

	inner, err := New(b.options.executorType(), env, expanded, Options{})
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "==> batch item: %s\n", item)
	return inner.StartAndWait()
}

// replacePlaceholder replaces all occurrences of ph in s with value.
// It avoids importing strings at the call site and keeps the expansion logic central.
func replacePlaceholder(s, ph, value string) string {
	result := make([]byte, 0, len(s))
	phBytes := []byte(ph)
	sBytes := []byte(s)
	vBytes := []byte(value)
	for i := 0; i < len(sBytes); {
		if i+len(phBytes) <= len(sBytes) && string(sBytes[i:i+len(phBytes)]) == ph {
			result = append(result, vBytes...)
			i += len(phBytes)
		} else {
			result = append(result, sBytes[i])
			i++
		}
	}
	return string(result)
}
