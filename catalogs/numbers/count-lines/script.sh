#!/bin/sh
set -eu

START_LINE="${START_LINE:?start_line is required}"
END_LINE="${END_LINE:--1}"

main() {
  echo "==> Counting lines"
  echo "    Range: ${START_LINE} to ${END_LINE}"

  total=100
  if [ "${END_LINE}" -lt 0 ]; then
    end=$((total + END_LINE + 1))
  else
    end="${END_LINE}"
  fi

  if [ "${START_LINE}" -lt 0 ]; then
    start=$((total + START_LINE + 1))
  else
    start="${START_LINE}"
  fi

  count=$((end - start + 1))
  echo "    Lines in range: ${count} (of ${total} total)"
  echo ""
  echo "✓ Done"
}

main "$@"
