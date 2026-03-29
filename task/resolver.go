package task

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/pkg/errors"
	"github.com/yuanjiecloud/fire/log"
)

type Resolver struct {
	dependencies  []string
	replaceMapper map[string]Replacement
}

func NewResolver(dependencies []string, replace []Replacement) *Resolver {
	ret := &Resolver{
		dependencies:  dependencies,
		replaceMapper: make(map[string]Replacement),
	}
	for _, item := range replace {
		ret.replaceMapper[item.Package] = item
	}
	return ret
}

func (t *Resolver) checkout(repositoryUrl string, namespace, name, branch string, repositoryPath string) error {
	stat, err := os.Stat(repositoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("repository directory does not exist: %v", repositoryPath)
		}
		return fmt.Errorf("stat repository path %q: %w", repositoryPath, err)
	}
	if !stat.IsDir() {
		return errors.Errorf("%s is not a directory", repositoryPath)
	}
	if branch == "" {
		if strings.Contains(repositoryUrl, "github.com") {
			branch = "main"
		} else {
			branch = "master"
		}
	}
	cloneDir := path.Join(repositoryPath, namespace, name)
	cloneDest := path.Join(cloneDir, branch)
	if CheckIfExists(cloneDest) {
		if CheckIfExists(path.Join(cloneDest, ".git")) {
			log.Info("ignore: ", fmt.Sprintf("%s/%s@%s", namespace, name, branch))
			return nil
		}
		// Destination exists but is not a git repo — remove it so we can clone cleanly.
		if err = os.RemoveAll(cloneDest); err != nil {
			return fmt.Errorf("remove stale clone destination %q: %w", cloneDest, err)
		}
	}
	log.Debug(cloneDir, "=>", repositoryUrl, branch)
	if err = os.MkdirAll(cloneDir, 0775); err != nil {
		return fmt.Errorf("init cache dir %q: %w", cloneDir, err)
	}
	gitCommand := exec.Command("git", "clone", "-b", branch, repositoryUrl, branch)
	gitCommand.Stderr = os.Stderr
	gitCommand.Stdout = os.Stdout
	gitCommand.Dir = cloneDir
	if err = gitCommand.Run(); err != nil {
		return err
	}
	return t.resolveDirectory(cloneDest)
}

func (t *Resolver) resolveDirectory(dir string) error {
	wd := Getwd()
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("enter directory %q: %w", dir, err)
	}
	if path.IsAbs(dir) {
		log.Info("enter dir: ", dir)
	} else {
		log.Info("enter dir: ", path.Join(wd, dir))
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			log.Fatal(err)
		}
		log.Info("goback dir: ", wd)
	}()
	if !CheckIfExists(DefaultConfigFile) {
		return errors.Errorf("%s is an invalid repository: %s not found", dir, DefaultConfigFile)
	}
	pipeline, err := Parse(DefaultConfigFile)
	if err != nil {
		return fmt.Errorf("invalid repository %q: %w", dir, err)
	}
	log.Debug("start resolving dependency pipeline: ", path.Join(dir, DefaultConfigFile))
	return pipeline.Resolve()
}

func (t *Resolver) Start() error {
	reposDir, err := GetGlobalReposDir()
	if err != nil {
		return fmt.Errorf("get global repos dir: %w", err)
	}
	log.Debug("dependencies count: ", len(t.dependencies))
	for _, depend := range t.dependencies {
		var (
			namespace string
			name      string
			version   string
			err       error
		)
		if globalResolverFilter.Contains(depend) {
			log.Debug("ignore dependency: ", depend)
			continue
		}
		globalResolverFilter.Add(depend)
		log.Info("resolving: ", depend)
		replacement, found := t.replaceMapper[depend]
		if found {
			namespace, name, _, err = SplitPackageName(replacement.Package)
			if err != nil {
				log.Error(err)
				return errors.Errorf("resolve replacement failed: %v", depend)
			}
			version = replacement.Version.String()
			if replacement.IsLocal() {
				log.Debug("detected a local repos: ", depend)
			if CheckIfExists(replacement.Repository) {
				log.Info("ignore local repository: ", depend)
				if err = t.resolveDirectory(replacement.Repository); err != nil {
					return fmt.Errorf("resolve local repository %q: %w", replacement.Repository, err)
				}
				continue
			}
			}
			err = t.checkout(replacement.Repository, namespace, name, version, reposDir)
			if err != nil {
				log.Error(err)
				return errors.Errorf("checkout replacement failed: %v", depend)
			}
		} else {
			namespace, name, version, err = SplitPackageName(depend)
			if err != nil {
				log.Error(err)
				return errors.Errorf("resolve dependencies failed")
			}
			repositoryUrl := fmt.Sprintf("https://github.com/%s/%s.git", namespace, name)
			err = t.checkout(repositoryUrl, namespace, name, version, reposDir)
			if err != nil {
				log.Error(err)
				return errors.Errorf("checkout package failed: %v", depend)
			}
		}
	}
	return nil
}
