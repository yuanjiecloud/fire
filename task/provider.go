package task

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/yuanjiecloud/fire/log"
)

var (
	pipelineMu               sync.RWMutex
	mapFromPipelineToLocation = make(map[string]string)
	pipelineMapper            = make(map[string]*Pipeline)
)

func AddPipeline(pipelineWithVersion string, dir string) (*Pipeline, error) {
	// Resolve to absolute path before any Chdir so configFile is always correct.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve path %q: %v", dir, err)
	}
	workdirBackup := Getwd()
	if err = os.Chdir(absDir); err != nil {
		return nil, fmt.Errorf("cannot enter directory %q: %v", absDir, err)
	}
	log.Debug("enter dir: ", absDir)
	defer func() {
		if err = os.Chdir(workdirBackup); err != nil {
			log.Fatal(err)
		}
		log.Debug("goback: ", workdirBackup)
	}()
	configFile := filepath.Join(absDir, DefaultConfigFile)
	pipeline, err := Parse(configFile)
	if err != nil {
		return nil, fmt.Errorf("invalid fire project: %s", absDir)
	}
	log.Debug("add pipeline: ", pipelineWithVersion, " => ", absDir)
	pipelineMu.Lock()
	mapFromPipelineToLocation[pipelineWithVersion] = absDir
	pipelineMapper[pipelineWithVersion] = pipeline
	pipelineMu.Unlock()
	return pipeline, nil
}

func FindPipeline(pipelineWithVersion string) (pipeline *Pipeline, found bool) {
	pipelineMu.RLock()
	pipeline, found = pipelineMapper[pipelineWithVersion]
	pipelineMu.RUnlock()
	return
}

func CheckIfContainPipeline(pipelineWithVersion string) bool {
	pipelineMu.RLock()
	_, found := pipelineMapper[pipelineWithVersion]
	pipelineMu.RUnlock()
	return found
}

func FindPipelineReposDir(pipelineWithVersion string) (dir string, found bool) {
	pipelineMu.RLock()
	dir, found = mapFromPipelineToLocation[pipelineWithVersion]
	pipelineMu.RUnlock()
	return
}
