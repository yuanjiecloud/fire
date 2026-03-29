# fire

A YAML-driven task pipeline runner for local, remote, and containerized execution.

---

## Features

- **Multiple executor types** — `bash`, `sh`, `ssh`, `docker`, `batch`
- **Pipeline dependencies** — declare Git-backed pipelines; `fire install` clones them automatically
- **Named environments** — define `environments` blocks and bind tasks to them
- **Parallel execution** — tasks in a pipeline can run concurrently with `parallel: true`
- **Batch execution** — run the same scripts against a list of items, sequentially or in parallel
- **Dependency replace** — swap a remote dependency for a local path during development

---

## Installation

```bash
go install github.com/yuanjiecloud/fire@latest
```

Or build from source:

```bash
git clone https://github.com/yuanjiecloud/fire.git
cd fire
go build -o fire .
```

Requires Go 1.23 or later.

---

## Quick start

Create a `fire.yaml` in your project root, then run:

```bash
fire run          # run all tasks in order
fire run <task>   # run a specific task
fire list         # list available tasks
fire install      # clone/checkout declared dependencies
fire update       # git-pull all cached dependencies
fire clean        # remove cached dependency directories
fire version      # print the build version
```

### Global flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--workdir` | `-w` | current dir | Working directory containing `fire.yaml` |
| `--verbose` | `-v` | false | Print debug log |
| `--printstack` | | false | Print call stack on fatal errors |
| `--global` | `-g` | false | Operate on global repos cache (used by `clean`) |
| `--task` | `-t` | | Task name to run (only for `run`) |

---

## `fire.yaml` reference

```yaml
version: v0.1.0          # optional, free-form string
name: my-project          # optional project name
parallel: false           # set true to run all tasks concurrently

environments:             # named environment blocks
  prod:
    HOST: prod.example.com
    USER: deploy
  dev:
    HOST: localhost
    USER: dev

tasks:
  - name: <task-name>
    type: bash | sh | ssh | docker | batch
    env: <environment-name>          # bind a named environment
    scripts:
      - <shell command>
    pipeline: <dependency-name>      # delegate to another fire.yaml
    # type-specific options (see below)

dependencies:
  - "namespace/repo@version"         # cloned from GitHub

replace:
  - package: "namespace/repo"
    repository: ./local/path         # local path override
    # or:
    repository: https://custom.git
    version: main
```

---

## Executor types

### `bash` / `sh`

Runs scripts through `/bin/bash` or `/bin/sh`. Environment variables are exported before the scripts run.

```yaml
tasks:
  - name: build
    type: bash
    env: prod
    scripts:
      - go build -o bin/app ./cmd/app
      - echo "build complete"
```

### `ssh`

Executes scripts on a remote host over SSH. Scripts are piped to `sh` on the remote side.

```yaml
tasks:
  - name: deploy
    type: ssh
    ssh-options:
      host: prod.example.com
      user: deploy
      port: 22
      identifier-file: ~/.ssh/id_rsa
      remote-path: /opt/app       # cd to this directory before running scripts
    scripts:
      - ./restart.sh
```

| `ssh-options` field | Description |
|---------------------|-------------|
| `host` | Remote hostname or IP (required) |
| `user` | SSH username |
| `port` | SSH port (default: system default 22) |
| `identifier-file` | Path to private key (`-i`) |
| `remote-path` | Working directory on the remote host |

### `docker`

Runs scripts inside a Docker container. Supports two modes:

**`docker run`** — starts a fresh container for each task execution:

```yaml
tasks:
  - name: test-in-container
    type: docker
    docker-options:
      image: golang:1.23
      work-dir: /workspace
      volumes:
        - ".:/workspace"
      network: host
      shell: bash              # optional, defaults to sh
    scripts:
      - go test ./...
```

**`docker exec`** — executes inside an already-running container:

```yaml
tasks:
  - name: reload
    type: docker
    docker-options:
      container: my-app
      work-dir: /app
    scripts:
      - ./reload.sh
```

| `docker-options` field | Description |
|------------------------|-------------|
| `image` | Docker image (`docker run` mode, mutually exclusive with `container`) |
| `container` | Container name or ID (`docker exec` mode) |
| `work-dir` | Working directory inside the container (`-w`) |
| `volumes` | Volume mount specs, e.g. `".:/workspace"` (run mode only) |
| `network` | Network mode, e.g. `host` (run mode only) |
| `shell` | Shell binary inside the container (default: `sh`) |

### `batch`

Iterates over a list of items and runs the same scripts for each, replacing a placeholder with the item value.

```yaml
tasks:
  - name: process-files
    type: batch
    batch-options:
      items:
        - file_001.txt
        - file_002.txt
        - file_003.txt
      parallel: true      # run all items concurrently
      max-parallel: 4     # limit to 4 concurrent goroutines
      placeholder: "{}"   # default, replaced with each item value
      executor: bash      # bash or sh, default: bash
    scripts:
      - wc -l {}
      - echo "done: {}"
```

Each item execution also receives a `FIRE_BATCH_ITEM` environment variable set to the current item value.

| `batch-options` field | Description |
|-----------------------|-------------|
| `items` | List of values to iterate over |
| `parallel` | Run items concurrently (default: false) |
| `max-parallel` | Max concurrent goroutines; 0 = unlimited |
| `placeholder` | Token replaced with item value (default: `{}`) |
| `executor` | Shell to use per item: `bash` or `sh` (default: `bash`) |

---

## Pipeline dependencies

Declare external `fire.yaml` projects as dependencies:

```yaml
dependencies:
  - "myorg/deploy-scripts@v2.1.0"   # cloned from github.com/myorg/deploy-scripts
  - "myorg/shared-tasks"            # defaults to main branch

replace:
  - package: "myorg/shared-tasks"
    repository: ../shared-tasks     # use local path instead
```

Run `fire install` to clone all dependencies into `~/.config/fire/repos/`.

A task can delegate to a dependency pipeline:

```yaml
tasks:
  - name: deploy
    pipeline: myorg/deploy-scripts@v2.1.0
    env: prod
```

---

## Parallel task execution

Set `parallel: true` at the pipeline level to run all tasks concurrently. Errors from all tasks are collected and returned together.

```yaml
parallel: true
tasks:
  - name: lint
    type: bash
    env: default
    scripts: [golangci-lint run]
  - name: test
    type: bash
    env: default
    scripts: [go test ./...]
  - name: build
    type: bash
    env: default
    scripts: [go build ./...]
```

---

## Example: PDF translation pipeline

A complete example combining `batch` (parallel OCR and translation) with `bash` (PDF split and merge):

```yaml
version: v0.1.0
name: pdf-translate

tasks:
  - name: split
    type: bash
    scripts:
      - mkdir -p pages
      - pdftoppm -png input.pdf pages/page

  - name: ocr
    type: batch
    batch-options:
      items: [pages/page-1.png, pages/page-2.png, pages/page-3.png]
      parallel: true
      max-parallel: 4
    scripts:
      - tesseract {} {}.txt -l chi_sim+eng

  - name: translate
    type: batch
    batch-options:
      items: [pages/page-1.png.txt, pages/page-2.png.txt, pages/page-3.png.txt]
      parallel: true
      max-parallel: 4
    scripts:
      - python3 translate.py {} {}.zh.txt

  - name: merge
    type: bash
    scripts:
      - python3 merge_pdf.py pages/ output_translated.pdf
```

---

## Project layout

```
fire.yaml               # project pipeline definition
~/.config/fire/
  fire.yaml             # global pipeline (used when no local fire.yaml found)
  repos/                # cached dependency checkouts
    <namespace>/
      <name>/
        <version>/      # cloned git repository
```

---

## Development

```bash
# run tests
go test ./...

# run tests with coverage
go test ./... -cover

# build
go build -o fire .
```

**Test coverage** (current):

| Package | Coverage |
|---------|----------|
| `datatype` | 100% |
| `executor` | 71% |
| `log` | 50% |
| `task` | 40% |

---

## License

See [LICENSE](LICENSE).
