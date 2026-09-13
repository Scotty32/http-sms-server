#!/bin/sh
set -e

if ! command -v pnpm >/dev/null 2>&1; then
    echo "==> Installing pnpm..."
    npm install -g pnpm@9
fi

echo "==> Installing dependencies..."
pnpm config set network-concurrency 5
pnpm install --frozen-lockfile

echo "==> Starting application..."
exec "$@"
