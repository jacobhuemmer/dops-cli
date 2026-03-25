#!/bin/sh
set -eu

RESOURCE="${RESOURCE:?resource is required}"
FORMAT="${FORMAT:-table}"

main() {
  echo "==> Describing resource: ${RESOURCE}"
  echo ""

  # Parse resource type from common ID patterns.
  case "${RESOURCE}" in
    arn:*)
      provider="aws"
      service=$(echo "${RESOURCE}" | cut -d: -f3)
      ;;
    projects/*)
      provider="gcp"
      service=$(echo "${RESOURCE}" | cut -d/ -f3)
      ;;
    /subscriptions/*)
      provider="azure"
      service=$(echo "${RESOURCE}" | cut -d/ -f5)
      ;;
    *)
      provider="unknown"
      service="unknown"
      ;;
  esac

  case "${FORMAT}" in
    json)
      printf '{"resource":"%s","provider":"%s","service":"%s","status":"active"}\n' \
        "${RESOURCE}" "${provider}" "${service}"
      ;;
    yaml)
      printf 'resource: %s\nprovider: %s\nservice: %s\nstatus: active\n' \
        "${RESOURCE}" "${provider}" "${service}"
      ;;
    *)
      echo "    Resource: ${RESOURCE}"
      echo "    Provider: ${provider}"
      echo "    Service:  ${service}"
      echo "    Status:   active"
      ;;
  esac

  echo ""
  echo "✓ Done"
}

main "$@"
