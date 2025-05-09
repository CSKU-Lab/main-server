#!/bin/sh

load_env_context() {
  local context_dir="$1"
  local env_file="${context_dir%/}/.env"

  if [ ! -f "$env_file" ]; then
    echo "No .env file found in context: $context_dir" >&2
    return 1
  fi

  # Export each key=value from .env
  export $(grep -v '^#' "$env_file" | xargs)
}
