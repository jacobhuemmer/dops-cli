#!/bin/sh
set -eu

MESSAGE="${MESSAGE:?message is required}"
LOG_FILE="${LOG_FILE:-/tmp/dops-notes.log}"

main() {
  timestamp=$(date '+%Y-%m-%d %H:%M:%S')
  entry="${timestamp}  ${MESSAGE}"

  echo "${entry}" >> "${LOG_FILE}"

  echo "==> Note appended"
  echo "    ${entry}"
  echo "    File: ${LOG_FILE}"
  echo ""
  echo "✓ Done"
}

main "$@"
