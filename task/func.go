package task

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/yuanjiecloud/fire/log"
)

var (
	globalCacheDir   string
	globalReposDir   string
	globalFireConfig string
)

func GetGlobalConfigDir() (fireConfigDir string, err error) {
	return GetGlobalCacheDir()
}

func GetGlobalCacheDir() (fireCacheDir string, err error) {
	if globalCacheDir != "" {
		return globalCacheDir, nil
	}
	fireCacheDir, err = os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir failed: %w", err)
	}
	fireCacheDir = path.Join(fireCacheDir, ".config", "fire")
	_, err = os.Stat(fireCacheDir)
	if err == nil {
		globalCacheDir = fireCacheDir
		return fireCacheDir, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat cache dir %q: %w", fireCacheDir, err)
	}
	if err = os.MkdirAll(fireCacheDir, 0775); err != nil {
		return "", fmt.Errorf("create cache dir %q: %w", fireCacheDir, err)
	}
	globalCacheDir = fireCacheDir
	return fireCacheDir, nil
}

func GetGlobalReposDir() (reposDir string, err error) {
	if globalReposDir != "" {
		return globalReposDir, nil
	}
	cacheDir, err := GetGlobalCacheDir()
	if err != nil {
		return "", fmt.Errorf("get repos dir failed: %w", err)
	}
	reposDir = path.Join(cacheDir, "repos")
	_, err = os.Stat(reposDir)
	if err == nil {
		globalReposDir = reposDir
		return reposDir, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat repos dir %q: %w", reposDir, err)
	}
	if err = os.MkdirAll(reposDir, 0775); err != nil {
		return "", fmt.Errorf("create repos dir %q: %w", reposDir, err)
	}
	globalReposDir = reposDir
	return reposDir, nil
}

func GetGlobalFireConfig() (configFile string, err error) {
	if globalFireConfig != "" {
		return globalFireConfig, nil
	}
	cacheDir, err := GetGlobalCacheDir()
	if err != nil {
		return "", err
	}
	configFile = path.Join(cacheDir, "fire.yaml")
	globalFireConfig = configFile
	return
}

// SplitPackageName parses a package spec of the form "namespace/name[@version]".
// Both namespace and name are required; omitting the namespace would produce an
// ambiguous GitHub clone URL ("https://github.com//name.git"), so single-segment
// names without a slash are rejected.
func SplitPackageName(packageName string) (namespace, name, version string, err error) {
	if packageName == "" {
		err = fmt.Errorf("invalid package: %q (expected namespace/name[@version])", packageName)
		return
	}
	parts := strings.SplitN(packageName, "@", 2)
	if len(parts) == 2 {
		version = parts[1]
		if version == "" {
			err = fmt.Errorf("invalid package: %q (version after '@' is empty)", packageName)
			return
		}
	}
	segments := strings.Split(parts[0], "/")
	switch len(segments) {
	case 2:
		namespace = segments[0]
		name = segments[1]
		if namespace == "" || name == "" {
			err = fmt.Errorf("invalid package: %q (namespace and name must not be empty)", packageName)
		}
	default:
		err = fmt.Errorf("invalid package: %q (expected namespace/name[@version])", packageName)
	}
	return
}

func CheckIfExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func Getwd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}

func CreateRepositoryLocation(namespace, repositoryName string) string {
	if repositoryName == "" {
		log.Fatal("invalid repository name: ", repositoryName)
	}
	globalReposDir, err := GetGlobalReposDir()
	if err != nil {
		log.Fatal(err)
	}
	l := []string{globalReposDir}
	if namespace == "" {
		l = append(l, DefaultNamespace)
	} else {
		l = append(l, namespace)
	}
	l = append(l, repositoryName)
	return path.Join(l...)
}

func CreateRepositoryLocationSpecificVersion(namespace, repositoryName, version string) string {
	if repositoryName == "" {
		log.Fatal("invalid repository name: ", repositoryName)
	}
	globalReposDir, err := GetGlobalReposDir()
	if err != nil {
		log.Fatal(err)
	}
	l := []string{globalReposDir}
	if namespace == "" {
		l = append(l, DefaultNamespace)
	} else {
		l = append(l, namespace)
	}
	l = append(l, repositoryName)
	if version == "" {
		l = append(l, DefaultVersionBranch)
	} else {
		l = append(l, version)
	}
	return path.Join(l...)
}

func CheckIfNestedRepository(dir string) bool {
	basedir := path.Dir(dir)
	return CheckIfExists(path.Join(basedir, DefaultConfigFile))
}

func CheckIfGitRepository(dir string) bool {
	return CheckIfExists(path.Join(dir, ".git"))
}

func GitFetchAndUpdate(dir string) error {
	if !CheckIfGitRepository(dir) {
		return fmt.Errorf("invalid git repository: %v", dir)
	}
	wd := Getwd()
	err := os.Chdir(dir)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = os.Chdir(wd)
		if err != nil {
			log.Fatal(err)
		}
	}()
	cmd := exec.Command("git", "pull")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func showTitle(title string) {
	holder := []byte("==========================================================================")
	titleData := []byte(fmt.Sprintf(" %s ", title))
	offset := (len(holder) - len(title)) / 2
	if offset < 0 {
		log.Info(title)
	} else {
		for i := 0; i < len(titleData); i++ {
			holder[i+offset] = titleData[i]
		}
		log.Info(string(holder))
	}
}
