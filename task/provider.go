package task

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/yuanjiecloud/fire/log"
)

var mapFromPipelineToLocation = make(map[string]string)
var pipelineMapper = make(map[string]*Pipeline)

func AddPipeline(pipelineWithVersion string, dir string) (*Pipeline, error) {
	// Resolve to absolute path before any Chdir so configFile is always correct.
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, errors.Errorf("cannot resolve path %q: %v", dir, err)
	}
	workdirBackup := Getwd()
	if err = os.Chdir(absDir); err != nil {
		return nil, errors.Errorf("cannot enter directory %q: %v", absDir, err)
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
		return nil, errors.Errorf("invalid fire project: %s", absDir)
	}
	log.Debug("add pipeline: ", pipelineWithVersion, " => ", absDir)
	mapFromPipelineToLocation[pipelineWithVersion] = absDir
	pipelineMapper[pipelineWithVersion] = pipeline
	return pipeline, nil
}

func FindPipeline(pipelineWithVersion string) (pipeline *Pipeline, found bool) {
	pipeline, found = pipelineMapper[pipelineWithVersion]
	return
}

func CheckIfContainPipeline(pipelineWithVersion string) bool {
	_, found := pipelineMapper[pipelineWithVersion]
	return found
}

func FindPipelineReposDir(pipelineWithVersion string) (dir string, found bool) {
	dir, found = mapFromPipelineToLocation[pipelineWithVersion]
	return
}
