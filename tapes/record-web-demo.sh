#!/bin/sh
# Record the dops web UI demo as a GIF.
#
# Prerequisites:
#   - dops built (make build)
#   - ffmpeg installed
#   - npx available (Node.js)
#
# Usage:
#   ./tapes/record-web-demo.sh

set -e

# Source nvm if available (needed for npx/node).
if [ -s "$HOME/.nvm/nvm.sh" ]; then
  . "$HOME/.nvm/nvm.sh"
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PORT=3000
OUTPUT="$ROOT_DIR/assets/web-demo.gif"
VIDEO_DIR="$ROOT_DIR/tapes/web-demo-results"

echo "==> Building dops..."
cd "$ROOT_DIR"
make build

echo "==> Starting dops web server on port $PORT..."
"$ROOT_DIR/bin/dops" open --no-browser --port "$PORT" &
DOPS_PID=$!

cleanup() {
  echo "==> Cleaning up..."
  kill "$DOPS_PID" 2>/dev/null || true
  rm -rf "$SCRIPT_DIR/web-demo-results"
}
trap cleanup EXIT

# Wait for server to be ready.
echo "==> Waiting for server..."
for i in $(seq 1 30); do
  if curl -s "http://localhost:$PORT/api/catalogs" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

echo "==> Recording web UI demo..."
cd "$SCRIPT_DIR"
./node_modules/.bin/playwright test --config playwright.config.ts 2>&1 || true
cd "$ROOT_DIR"

# Find the recorded video (Playwright saves as .webm).
VIDEO=$(find "$SCRIPT_DIR/web-demo-results" -name "*.webm" -type f 2>/dev/null | head -1)

if [ -z "$VIDEO" ]; then
  echo "ERROR: No video recorded. Check Playwright output above."
  exit 1
fi

echo "==> Converting video to GIF..."
# Two-pass: generate palette, then encode with it for high-quality GIF.
PALETTE="/tmp/dops-web-demo-palette.png"
ffmpeg -y -i "$VIDEO" \
  -vf "fps=12,scale=900:-1:flags=lanczos,palettegen=stats_mode=diff" \
  "$PALETTE" 2>/dev/null

ffmpeg -y -i "$VIDEO" -i "$PALETTE" \
  -lavfi "fps=12,scale=900:-1:flags=lanczos[x];[x][1:v]paletteuse=dither=bayer:bayer_scale=3" \
  "$OUTPUT" 2>/dev/null

echo "==> Done! GIF saved to $OUTPUT"
ls -lh "$OUTPUT"
