#!/usr/bin/env bash
# Pre-tool validation: enforce bun in ts/, block edits to generated files.
INPUT=$(cat)
TOOL=$(echo "$INPUT" | jq -r '.tool_name // empty')

case "$TOOL" in
  Bash)
    CMD=$(echo "$INPUT" | jq -r '.tool_input.command // empty')
    [[ -z "$CMD" ]] && exit 0

    if echo "$CMD" | grep -qE '^\s*(npm|yarn|pnpm)\s'; then
      echo "Use bun instead (bun install, bun add, bun run, bun test)" >&2
      exit 2
    fi
    if echo "$CMD" | grep -qE '^\s*npx\s'; then
      echo "Use bunx instead of npx" >&2
      exit 2
    fi
    ;;

  Edit|Write)
    FILE=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')
    [[ -z "$FILE" ]] && exit 0

    case "$FILE" in
      */go/client.gen.go)
        echo "Blocked: client.gen.go is generated. Run 'just gen' to regenerate." >&2
        exit 2
        ;;
      */ts/src/client/*)
        echo "Blocked: ts/src/client/ is generated. Run 'just gen' to regenerate." >&2
        exit 2
        ;;
      *package-lock.json|*yarn.lock|*pnpm-lock.yaml)
        echo "Blocked: use bun.lockb" >&2
        exit 2
        ;;
    esac
    ;;
esac

exit 0
