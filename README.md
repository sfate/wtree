# wtree — Git Worktree Helper

A CLI for managing git worktrees organised by ticket reference or arbitrary name. By default, worktrees are stored in a `.wtree/` subdirectory inside each repository and can be listed, created, and cleaned up with simple commands.

## Install

```bash
go install github.com/sfate/wtree@latest
```

`go install` places the binary in `$(go env GOBIN)` (falling back to `$(go env GOPATH)/bin`). Make sure that directory is in your `PATH`:

```bash
export PATH="$(go env GOBIN):$PATH"
```

If you use `asdf` for Go, refresh the shim after installation:

```bash
asdf reshim golang $(asdf current golang | awk 'NR==2 {print $2}')
```

## Shell integration

`wtree` needs shell integration to `cd` into worktree directories. Add to your shell config:

**zsh** (`~/.zshrc`):
```zsh
eval "$(wtree --shell-init zsh)"
```

**bash** (`~/.bashrc`):
```bash
eval "$(wtree --shell-init bash)"
```

> **Powerlevel10k users:** place the `eval` line **before** the p10k instant prompt block in `~/.zshrc`.

**How it works:** navigation commands (`wtree <ref>`, `wtree --root`) print only the target path to stdout. The shell function captures that output and calls `cd`. All other output (branch names, status messages) goes to stderr and displays in your terminal normally. Non-navigation commands (`--list`, `--delete`, etc.) bypass the wrapper and run directly.

## Configuration

On first run inside a git project, wtree automatically adds an entry for that project to `~/.config/wtree/config.yml` and exits, prompting you to fill in any required fields.

```yaml
# ~/.config/wtree/config.yml

projects:
  - name: project-abc
    path: ~/code/project-abc
    base_dir: ~/.wtree/project-abc   # optional — default is ~/code/project-abc/.wtree
    ticket_prefix: ABC-   # optional — enables automatic branch derivation from refs like ABC-1234
    branch_prefix: ob-    # optional — prepended to ticket prefix: ob- + abc- → ob-abc-1234
    hooks:
      post_navigation: ~/.config/wtree/hooks/project-abc/post_navigation.sh
      post_delete: ~/.config/wtree/hooks/project-abc/post_delete.sh

  - name: personal-site
    path: ~/code/personal-site
    # no ticket_prefix — branch must always be supplied explicitly
```

**Rules:**
- the YAML root contains `projects` only
- `base_dir` is optional per project; when omitted it defaults to `<project>/.wtree`
- `ticket_prefix` is optional; if omitted, `[branch]` must always be passed explicitly
- `branch_prefix` is optional; only meaningful when `ticket_prefix` is set
- `hooks` are optional; each entry is a path to an executable script
- Project `name` and `path` must each be unique across all entries

**Git ignore:**

If you use the default in-project base dir, add `.wtree/` to your repository’s `.gitignore`:

```gitignore
.wtree/
```

**Hook arguments:**

| Hook | Arguments |
|---|---|
| `post_navigation` | `<ref> <project_name> <worktree_dir>` |
| `post_delete` | `<ref>` |

**Branch derivation:**

| `ticket_prefix` | `branch_prefix` | ref `ABC-1234` → branch |
|---|---|---|
| `ABC-` | `ob-` | `ob-abc-1234` |
| `ABC-` | _(absent)_ | `abc-1234` |
| _(absent)_ | any | `[branch]` required |

## Commands

### Create / switch to a worktree

```bash
wtree <ref> [branch] [base_branch]
```

If `ticket_prefix` is configured and `<ref>` starts with it, the branch name is derived automatically. Otherwise `[branch]` is required.

```bash
wtree ABC-1234                        # auto-derives branch from ticket number
wtree feature-x my-feature-branch    # explicit branch
wtree feature-x my-branch develop    # explicit branch and base branch
```

### List worktrees

```bash
wtree --list
```

Prints an ASCII table with Ref, Branch, and Last Activity columns, sorted by most recent activity.

### Delete a worktree

```bash
wtree --delete <ref>
```

### Navigate to project root

```bash
wtree --root
```

### Remove all worktrees for the current project

```bash
wtree --clean
# or
wtree --clear
```

### Remove stale worktrees

```bash
wtree --clean-stale
```

Lists worktrees with no commit activity in the last two weeks, shows a confirmation prompt, then removes them.

### Help

```bash
wtree --help
wtree -h
```

### Version

```bash
wtree --version
```

The value comes from the repository [VERSION](/Users/oleksiibobyriev/blackholesun/wtree/VERSION) file by default, and release builds override the embedded value with the same version string.

## Development

```bash
make build       # compile binary
make test        # run tests
make lint        # go vet + golangci-lint
make audit       # govulncheck
make release           # tag and push patch release (v1.2.3 → v1.2.4)
make release BUMP=minor  # v1.2.3 → v1.3.0
make release BUMP=major  # v1.2.3 → v2.0.0
make clean       # remove binary and dist/
```

`make release` requires a clean worktree, verifies that [VERSION](/Users/oleksiibobyriev/blackholesun/wtree/VERSION) matches the latest tag, bumps the version, commits the `VERSION` change, pushes the commit, and then pushes the new tag.
