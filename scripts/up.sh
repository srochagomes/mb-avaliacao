#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

# Em primeiro plano, Ctrl+C (ou saída do compose) remove containers e a rede.
# Com -d/--detach o script só sobe e deixa rodando; use ./scripts/down.sh para parar.
detached=0
for arg in "$@"; do
  case "$arg" in
    -d|--detach) detached=1 ;;
  esac
done

if [[ "$detached" -eq 0 ]]; then
  cleanup() {
    docker compose down --remove-orphans
  }
  trap cleanup EXIT INT TERM
fi

docker compose up --build "$@"
