
function wtree() {
  case "$1" in
    --shell-init|-h|--help|--list|--delete|--clean|--clear|--clean-stale|--version)
      command wtree "$@"
      ;;
    *)
      local output exit_code dir shell_code
      output=$(command wtree "$@")
      exit_code=$?
      dir=$(printf '%s' "$output" | head -1)
      shell_code=$(printf '%s' "$output" | tail -n +2)
      if [[ -n "$dir" ]]; then
        cd "$dir"
        [[ -n "$shell_code" ]] && eval "$shell_code"
      fi
      return $exit_code
      ;;
  esac
}
