# wtree — Git Worktree Helper

A CLI for managing git worktrees organised by ticket reference or arbitrary name. Worktrees are stored in a central directory (`~/.worktrees/<project>/`) and can be listed, created, and cleaned up with simple commands.

Navigation is built-in: `wtree` spawns a new shell session inside the target worktree directory. Exit the shell to return to where you started.

## Install

```bash
# Build from source
make build
mv wtree /usr/local/bin/

# Or with go install
go install github.com/sfate/wtree@latest
```

## Configuration

On first run inside a git project, wtree automatically adds an entry for that project to `~/.config/wtree/config.yml` and exits, prompting you to fill in any required fields.

```yaml
# ~/.config/wtree/config.yml

base_dir: ~/.worktrees   # optional — override the default worktree root

projects:
  - name: project-abc
    path: ~/code/project-abc
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
- `base_dir` defaults to `~/.worktrees` when omitted
- `ticket_prefix` is optional; if omitted, `[branch]` must always be passed explicitly
- `branch_prefix` is optional; only meaningful when `ticket_prefix` is set
- `hooks` are optional; each entry is a path to an executable script
- Project `name` and `path` must each be unique across all entries

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

`make release` fetches the latest tag, bumps the version, pushes the tag, and cross-compiles binaries into `dist/`.
