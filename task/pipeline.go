package task

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"

	"errors"

	"github.com/yuanjiecloud/fire/datatype"
	"github.com/yuanjiecloud/fire/log"
	"gopkg.in/yaml.v3"
)

type Pipeline struct {
	configfile string

	Version      Version       `json:"version,omitempty" yaml:"version,omitempty"`
	Environments EnvProvider   `json:"environments,omitempty" yaml:"environments,omitempty"`
	Tasks        []Task        `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Dependencies []string      `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Replace      []Replacement `json:"replace,omitempty" yaml:"replace,omitempty"`
	// Parallel runs all tasks concurrently when true.
	// Individual task errors are collected and returned together.
	Parallel bool `json:"parallel,omitempty" yaml:"parallel,omitempty"`
}

func Parse(file string) (c *Pipeline, err error) {
	log.Debug("parsing file: ", file, " wd:", Getwd())
	_, err = os.Stat(file)
	if os.IsNotExist(err) {
		return nil, err
	}
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c = &Pipeline{}
	err = yaml.Unmarshal(fileContent, c)
	if err != nil {
		return
	}
	c.configfile = file
	return
}

func (t *Pipeline) ToJson() string {
	data, err := json.Marshal(t)
	if err != nil {
		log.Error("failed to marshal pipeline to JSON: ", err)
		return "{}"
	}
	return string(data)
}

func (t *Pipeline) FindTask(taskName string) (result Task, found bool) {
	for _, item := range t.Tasks {
		if item.Name == taskName {
			return item, true
		}
	}
	return result, false
}

func (t *Pipeline) CreateContext(ctx *Context) *Context {
	result := &Context{
		EnvProvider: t.Environments.Clone(),
	}
	if ctx != nil && ctx.EnvProvider != nil {
		result.Parent = ctx
		result.EnvProvider = result.EnvProvider.MergeKeepExisting(ctx.EnvProvider)
	}
	return result
}

func (t *Pipeline) RunAll(ctx *Context) error {
	newContext := t.CreateContext(ctx)
	if t.Parallel {
		return t.runAllParallel(newContext)
	}
	return t.runAllSequential(newContext)
}

func (t *Pipeline) runAllSequential(ctx *Context) error {
	for _, item := range t.Tasks {
		showTitle(fmt.Sprintf("start task: %s", item.Name))
		log.Debug("task env: ", item.Env)
		err := item.Exec(ctx.UseEnv(item.Env))
		if err != nil {
			return err
		}
		showTitle(fmt.Sprintf("end(%s)", item.Name))
	}
	return nil
}

// runAllParallel runs every task in its own goroutine and collects all errors.
func (t *Pipeline) runAllParallel(ctx *Context) error {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []string
	)
	for _, item := range t.Tasks {
		item := item // capture
		wg.Add(1)
		go func() {
			defer wg.Done()
			showTitle(fmt.Sprintf("start task (parallel): %s", item.Name))
			if err := item.Exec(ctx.UseEnv(item.Env)); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("task %q: %v", item.Name, err))
				mu.Unlock()
			} else {
				showTitle(fmt.Sprintf("end(%s)", item.Name))
			}
		}()
	}
	wg.Wait()
	if len(errs) == 0 {
		return nil
	}
	msg := fmt.Sprintf("%d task(s) failed:", len(errs))
	for _, s := range errs {
		msg += "\n  " + s
	}
	return errors.New(msg)
}

func (t *Pipeline) RunTask(name string, ctx *Context) error {
	selected, found := t.FindTask(name)
	if !found {
		log.Debug("task not found: ", name)
		pipeline, b := FindPipeline(name)
		if !b {
			log.Debug("pipeline not found: ", name)
			return fmt.Errorf("task not found: %s", name)
		}
		return pipeline.RunAll(t.CreateContext(ctx))
	}
	showTitle(fmt.Sprintf("start task: %s", name))
	defer func() {
		showTitle(fmt.Sprintf("end(%s)", name))
	}()
	return selected.Exec(t.CreateContext(ctx).UseEnv(selected.Env))
}

func (t *Pipeline) Resolve() error {
	resolver := NewResolver(t.Dependencies, t.Replace)
	return resolver.Start()
}

func (t *Pipeline) GetAllowTaskList() (taskList datatype.SortableStringList) {
	if t == nil {
		return make(datatype.SortableStringList, 0)
	}
	filter := make(map[string]bool)
	for _, item := range t.Tasks {
		if _, b := filter[item.Name]; !b {
			taskList = append(taskList, item.Name)
			filter[item.Name] = true
			if item.Pipeline != "" {
				pipeline, pipelineExists := FindPipeline(item.Pipeline)
				if !pipelineExists {
					continue
				}
				for _, tn := range pipeline.GetAllowTaskList() {
					if _, b := filter[tn]; !b {
						filter[tn] = true
						taskList = append(taskList, tn)
					}
				}
			}
		}
	}
	sort.Sort(taskList)
	return
}

func (t *Pipeline) Preload() error {
	enter(t.Getwd())
	defer goback()
	replaceMapper := make(map[string]Replacement)
	for _, replace := range t.Replace {
		replaceMapper[replace.Package] = replace
	}
	for _, depend := range t.Dependencies {
		log.Debug("resolving dependency: ", depend)
		replacement, found := replaceMapper[depend]
		namespace, name, version, err := SplitPackageName(depend)
		if err != nil {
			return fmt.Errorf("invalid dependency %q: %w", depend, err)
		}
		if found {
			if replacement.Repository == "" {
				return fmt.Errorf("replacement for %q has empty repository", depend)
			}
			var repositoryDir string
			if replacement.IsLocal() {
				repositoryDir = path.Join(Getwd(), replacement.Repository)
			} else {
				repositoryDir = CreateRepositoryLocationSpecificVersion(namespace, name, replacement.Version.String())
			}
			pipeline, err := AddPipeline(depend, repositoryDir)
			if err != nil {
				return fmt.Errorf("add pipeline %q: %w", depend, err)
			}
			if err = pipeline.Preload(); err != nil {
				return err
			}
		} else {
			log.Debug("no replacement dependency: ", name)
			repositoryDir := CreateRepositoryLocationSpecificVersion(namespace, name, version)
			pipeline, err := AddPipeline(depend, repositoryDir)
			if err != nil {
				return fmt.Errorf("add pipeline %q: %w", depend, err)
			}
			if err = pipeline.Preload(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Pipeline) Getwd() string {
	return path.Dir(t.configfile)
}

func (t *Pipeline) CleanDependencies() error {
	for _, depend := range t.Dependencies {
		reposDir, found := FindPipelineReposDir(depend)
		if !found {
			continue
		}
		pipeline, found := FindPipeline(depend)
		if !found {
			continue
		}
		err := pipeline.CleanDependencies()
		log.CheckAndFatal(err)
		log.Debug("found dependency: ", reposDir)
		if !CheckIfNestedRepository(reposDir) {
			log.Info("rm ", reposDir)
			err = os.RemoveAll(reposDir)
			log.CheckAndFatal(err)
		}
	}
	return nil
}

func (t *Pipeline) UpdateDependencies() error {
	for _, depend := range t.Dependencies {
		reposDir, found := FindPipelineReposDir(depend)
		if !found {
			continue
		}
		pipeline, found := FindPipeline(depend)
		if !found {
			continue
		}
		if err := pipeline.UpdateDependencies(); err != nil {
			return err
		}
		if CheckIfGitRepository(reposDir) {
			log.Info("update: ", reposDir)
			if err := GitFetchAndUpdate(reposDir); err != nil {
				return fmt.Errorf("git update %q: %w", reposDir, err)
			}
		} else {
			log.Info("ignore local repository dir: ", reposDir)
		}
	}
	return nil
}
