package executor

import (
	"bytes"
	"fmt"
	"os"
)

func NewShExecutor(env map[string]string, scripts []string) IExecutor {
	in := bytes.NewBuffer(nil)
	for k, v := range env {
		in.WriteString(fmt.Sprintf("export %s=%s\n", k, shellQuote(v)))
	}
	for _, line := range scripts {
		if len(line) == 0 {
			continue
		}
		in.WriteString(line + "\n")
	}
	return &bashExecutor{
		Binary: "/bin/sh",
		in:     in,
		out:    os.Stdout,
	}
}
