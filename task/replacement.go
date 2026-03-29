package task

import (
	"fmt"
	"strings"

	"github.com/yuanjiecloud/fire/log"
)

type Replacement struct {
	Package    string  `json:"package" yaml:"package"`
	Version    Version `json:"version" yaml:"version"`
	Repository string  `json:"repository" yaml:"repository"`
}

func (t Replacement) IsLocal() bool {
	log.Debug(fmt.Sprintf("pkg: %s, ver: %v, repos: %s", t.Package, t.Version, t.Repository))
	return !strings.HasPrefix(t.Repository, "http://") &&
		!strings.HasPrefix(t.Repository, "https://") &&
		!strings.HasPrefix(t.Repository, "git@")
}
