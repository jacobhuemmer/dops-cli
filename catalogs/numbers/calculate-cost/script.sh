#!/bin/sh
set -eu

HOURS="${HOURS:?hours is required}"
RATE="${RATE:?rate is required}"

main() {
  echo "==> Estimating compute cost"
  echo "    Hours: ${HOURS}"
  echo "    Rate:  \$${RATE}/hr"
  echo ""

  cost=$(printf '%.2f' "$(echo "${HOURS} * ${RATE}" | bc -l)")
  echo "    Estimated cost: \$${cost}"
  echo ""
  echo "✓ Done"
}

main "$@"
