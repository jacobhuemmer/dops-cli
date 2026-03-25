#!/bin/sh
set -eu

TARGET="${TARGET:?target is required}"

main() {
  echo "==> Checking path: ${TARGET}"
  echo ""

  if [ ! -e "${TARGET}" ]; then
    echo "✗ Path does not exist" >&2
    exit 1
  fi

  if [ -d "${TARGET}" ]; then
    echo "    Type: directory"
    count=$(ls -1 "${TARGET}" | wc -l | tr -d ' ')
    echo "    Entries: ${count}"
  elif [ -f "${TARGET}" ]; then
    echo "    Type: file"
    size=$(wc -c < "${TARGET}" | tr -d ' ')
    echo "    Size: ${size} bytes"
  else
    echo "    Type: other"
  fi

  perms=$(ls -ld "${TARGET}" | cut -d' ' -f1)
  echo "    Permissions: ${perms}"
  echo ""
  echo "✓ Done"
}

main "$@"
