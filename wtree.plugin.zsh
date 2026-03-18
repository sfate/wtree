# Git worktree helper plugin
#
# ~/.oh-my-zsh/custom/plugins/wtree/wtree.plugin.zsh

# Base directory for worktrees
_worktree_base_dir="$HOME/.worktrees"
_worktree_env_file=".wtree_env"
_ticket_prefix="LM-"
_branch_prefix="ob-lm-"

# Get project paths
_wtree_get_project_dir() {
  local project_git_dir="$(git rev-parse --path-format=absolute --git-common-dir 2> /dev/null)"
  if [[ "$project_git_dir" == "" ]]; then
    echo "\nError: not a git repository." >&2
    return 1
  fi

  echo "$(dirname $project_git_dir)"
  return 0
}

# Derive branch name from <ref>
_wtree_derive_branch() {
  local ref="$1"

  if [[ "$ref" != ${_ticket_prefix}* ]]; then
    return 1
  fi

  local num="${ref#$_ticket_prefix}"
  local branch_name="$_branch_prefix$num"
  local branch=$(git show-ref --heads | grep -E "$branch_name" | head -1 | awk '{print $2}' | sed 's|refs/heads/||')

  if [[ "$branch" == "" ]]; then
    branch="$branch_name"
  fi

  echo "$branch"
  return 0
}

# Derive base branch
_wtree_get_base_branch() {
  local base_branch_ref=$(git symbolic-ref refs/remotes/${remote:-origin}/HEAD 2>/dev/null)
  if [[ "$base_branch_ref" == "" ]]; then
    echo "\nError: not a git repository." >&2
    return 1
  else
    echo "${base_branch_ref##*/}"
  fi
}

# Check if branch exists
_wtree_branch_exists() {
  local branch="$1"
  local exists=$(git show-ref --heads | grep -E "$branch" | head -1 | awk '{print $2}' | sed 's|refs/heads/||')
  [[ "$exists" != "" ]]
}

# Create branch if needed
_wtree_ensure_branch() {
  local branch="$1"
  local base_branch="$2"

  if ! _wtree_branch_exists "$branch"; then
    echo "Branch does not exist.. creating from: $base_branch."
    git branch "$branch" "$base_branch"
  fi
}

# Command: List worktrees
_wtree_list() {
  local project_dir
  project_dir=$(_wtree_get_project_dir) || return 1
  local name_project="$(basename $project_dir)"
  local worktree_project_dir="$_worktree_base_dir/$name_project"

  if [[ ! -d "$worktree_project_dir" ]]; then
    echo "\nNo worktrees found for project: $name_project"
    return 0
  fi

  echo "\nWorktrees for project: $name_project"
  echo "────────────────────────────────────"

  for worktree in "$worktree_project_dir"/*(/N); do
    local ref="$(basename "$worktree")"
    local branch=$(cd "$worktree" 2>/dev/null && git branch --show-current 2>/dev/null)
    if [[ "$branch" != "" ]]; then
      local last_activity=$(cd "$worktree" 2>/dev/null && git log -1 --format="%cr" "$branch" 2>/dev/null)
      if [[ "$last_activity" != "" ]]; then
        echo "$ref -> $branch  ($last_activity)"
      else
        echo "$ref -> $branch"
      fi
    fi
  done
}

# Command: Delete worktree
_wtree_delete() {
  local ref="$1"

  if [[ "$ref" == "" ]]; then
    echo "\nError: <ref> is required for --delete." >&2
    echo "Usage: wtree --delete <ref>"
    return 1
  fi

  local project_dir
  project_dir=$(_wtree_get_project_dir) || return 1
  local name_project="$(basename $project_dir)"
  local worktree_project_dir="$_worktree_base_dir/$name_project"
  local worktree_dir="$worktree_project_dir/$ref"

  if [[ ! -d "$worktree_dir" ]]; then
    echo "\nError: worktree not found: $ref" >&2
    return 1
  fi

  echo "Removing worktree: $ref"
  cd "$project_dir"
  git worktree remove "$worktree_dir" --force
  echo "Worktree removed."

  _wtree_custom_post_delete "$ref"
}

# Command: Clean all worktrees
_wtree_clean() {
  local project_dir
  project_dir=$(_wtree_get_project_dir) || return 1
  local name_project="$(basename $project_dir)"
  local worktree_project_dir="$_worktree_base_dir/$name_project"

  if [[ ! -d "$worktree_project_dir" ]]; then
    echo "\nNo worktrees to clean for project: $name_project"
    return 0
  fi

  echo "Cleaning all worktrees for project: $name_project"
  cd "$project_dir"

  for worktree in "$worktree_project_dir"/*(/N); do
    local ref="$(basename "$worktree")"
    echo "Removing: $ref"
    git worktree remove "$worktree" --force

    _wtree_custom_post_delete "$ref"
  done

  # Remove empty worktree project directory
  if [[ -d "$worktree_project_dir" && -z "$(ls -A "$worktree_project_dir")" ]]; then
    rmdir "$worktree_project_dir"
  fi

  echo "All worktrees cleaned."
}

# Command: Navigate to project root
_wtree_root() {
  local project_dir
  project_dir=$(_wtree_get_project_dir) || return 1

  echo "Navigate to project: $project_dir"
  cd "$project_dir"
}

# Command: Show help
_wtree_help() {
  echo "\nGit Worktree Helper"
  echo "═══════════════════"
  echo "\nUsage:"
  echo "  wtree <ref> [branch] [base_branch]   Create/switch to worktree"
  echo "  wtree --list                         List worktrees for project"
  echo "  wtree --delete <ref>                 Delete worktree by <ref>"
  echo "  wtree --clean/--clear                Remove all project worktrees"
  echo "  wtree --root                         Navigate to project root"
  echo "  wtree --help/-h                      Show this help"
  echo "\nExamples:"
  echo "  wtree LM-1234                        Auto-detect branch for ticket"
  echo "  wtree feature-x my-branch            Create worktree with custom branch"
  echo "  wtree --list                         Show all worktrees"
  echo "  wtree --delete LM-1234               Remove specific worktree"
}

# Command: Create/switch to worktree
_wtree_create() {
  local ref="$1"
  local branch="$2"
  local base_branch="$3"

  if [[ "$ref" == "" ]]; then
    _wtree_help
    return 1
  fi

  local project_dir
  project_dir=$(_wtree_get_project_dir) || return 1
  local name_project="$(basename $project_dir)"
  local worktree_project_dir="$_worktree_base_dir/$name_project"
  local worktree_dir="$worktree_project_dir/$ref"

  # Always navigate to project dir to avoid git worktree errors
  echo "Navigate to project: $project_dir"
  cd "$project_dir"

  # Create worktree base dir if it doesn't exist
  mkdir -p "$worktree_dir"

  if [[ "$(git worktree list | grep -F "$worktree_dir")" != "" ]]; then
    # echo "Worktree already exists. Switching to: $worktree_dir"
    # cd "$worktree_dir"
    echo "Worktree already exists."
    _wtree_navigate "$ref" "$name_project" "$worktree_dir"
    return 0
  fi

  # Derive branch name from ref if not provided
  if [[ "$branch" == "" ]]; then
    branch=$(_wtree_derive_branch "$ref")
    if [[ "$branch" == "" ]]; then
      echo "\nError: [branch name] is required, once <ref> does not match $_ticket_prefix* pattern." >&2
      echo "Usage: wtree <ref> [branch] [base_branch]" >&2
      return 1
    fi
  fi

  # Derive base branch if not provided
  if [[ "$base_branch" == "" ]]; then
    base_branch=$(_wtree_get_base_branch)
  fi

  # Create branch if it doesn't exist
  echo "Using branch name: $branch"
  _wtree_ensure_branch "$branch" "$base_branch"

  # Create or switch to worktree
  git worktree add --force "$worktree_dir" "$branch"
  # echo "Switching to worktree: $worktree_dir"
  # cd "$worktree_dir"
  _wtree_navigate "$ref" "$name_project" "$worktree_dir"
}

# Navigate to worktree directory
_wtree_navigate() {
  local ref="$1"
  local name_project="$2"
  local worktree_dir="$3"

  echo "Switching to worktree: $worktree_dir"
  cd "$worktree_dir"

  _wtree_custom_post_navigation "$ref" "$name_project"
}

# Custom initialization for worktree (e.g. database setup)
_wtree_custom_post_navigation() {
  # This function is called after a worktree is created or switched to, and can be used to perform any necessary setup tasks, such as loading environment variables or preparing a test database.

  local ref="$1"
  local name_project="$2"

  if [[ ! -f ".env" ]]; then
    echo "\nWarning: .env file not found in worktree. Skipping environment variable setup." >&2
  else
    echo "Loading environment variables from .env file..."
    export $(cat .env | xargs)
  fi

  if [[ "$name_project" == "veer-api" ]]; then
    # check certs (it is required for app to work, but it's under .gitignore, so it may be missing)
    mkdir -p certs
    if [[ -z "$(ls -A certs)" ]]; then
      echo "No TLS certs found in certs/ directory. Generating self-signed certs for local development..."
      make selfcerts > /dev/null 2>&1 || echo " * Failed to generate TLS certs. Continuing without TLS configuration."
    fi
  fi

  local dbhost="localhost"
  local dbuser="postgres"
  local dbpass="postgres"
  local dbname="db-test-$ref"
  local dbport=""

  if [[ "$name_project" == "veer-api" ]]; then
    dbport="5432"
  # else if [[ "$name_project" == "api-reporting-service" ]]; then
  #   dbport="5433"
  fi

  export COMPOSE_PROJECT_NAME="$name_project"

  local new_env="false"
  local _wtree_env="$_worktree_env_file"
  if [[ ! -f "$PWD/$_wtree_env" ]]; then
    echo "\nAdding $_wtree_env with set of varibles for worktree." >&2
    echo "export COMPOSE_PROJECT_NAME=\"$COMPOSE_PROJECT_NAME\"" > $_wtree_env
    new_env="true"
  fi

  if [[ "$dbport" != "" && "$new_env" == "true" ]]; then
    # VEER_TEST_DB_URL is used by tests to connect to the test database
    export VEER_TEST_DB_URL="dbname=$dbname user=$dbuser password=$dbpass host=$dbhost sslmode=disable"

    echo "export VEER_TEST_DB_URL=\"$VEER_TEST_DB_URL\"" >> $_wtree_env
    echo "export DBURL=\"$VEER_TEST_DB_URL\"" >> $_wtree_env

    if PGPASSWORD=$dbpass psql -h "$dbhost" -p "$dbport" -U "$dbuser" -lqt | cut -d \| -f 1 | grep -qw "$dbname"; then
      echo "Database '$dbname' already exists"
    else
      echo "Creating database '$dbname'..."
      PGPASSWORD=$dbpass createdb -h $dbhost -p $dbport -U $dbuser -w $dbname > /dev/null
    fi

    echo "Running database migrations for '$dbname'..."
    DBURL="$VEER_TEST_DB_URL" go run cmd/migrate/migrate.go up > /dev/null
  fi
}

# Custom cleanup for worktree (e.g. database teardown)
_wtree_custom_post_delete() {
  # This function is called after a worktree is deleted, and can be used to perform any necessary cleanup tasks, such as dropping the associated test database.

  local ref="$1"

  local dbhost="localhost"
  local dbport="5432"
  local dbuser="postgres"
  local dbpass="postgres"
  local dbname="db-test-$ref"

  echo "Dropping database for worktree..."
  PGPASSWORD=$dbpass dropdb --if-exists -h $dbhost -p $dbport -U $dbuser -w $dbname > /dev/null
}

# Main function - router
function wtree() {
  case "$1" in
    --list)
      _wtree_list
      ;;
    --delete)
      _wtree_delete "$2"
      ;;
    --clean|--clear)
      _wtree_clean
      ;;
    --root)
      _wtree_root
      ;;
    --help|-h)
      _wtree_help
      ;;
    *)
      _wtree_create "$@"
      ;;
  esac
}
