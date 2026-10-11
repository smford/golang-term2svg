#!/usr/bin/env bash
# scripts/update-docs-version.sh
# Injects the latest release version into documentation files.
set -euo pipefail

DOCS_DIR="${1:-docs}"
INDEX_FILE="${DOCS_DIR}/index.html"
VERSION_ARG="${2:-}"

if [ -n "${VERSION_ARG}" ]; then
  TAG="${VERSION_ARG}"
else
  # Query latest release via gh CLI or git tags
  TAG=$(gh release view --json tagName -q .tagName 2>/dev/null || git tag -l 'v*' | sort -V | tail -n 1)
fi

if [ -z "${TAG}" ]; then
  TAG="v0.1.0"
fi

VERSION="${TAG#v}"

echo "Updating documentation to version ${TAG} (${VERSION})..."

if [ -f "${INDEX_FILE}" ]; then
  # Replace badge and release version placeholders
  sed -i.bak -E \
    -e "s/v[0-9]+\.[0-9]+\.[0-9]+/${TAG}/g" \
    "${INDEX_FILE}"
  rm -f "${INDEX_FILE}.bak"
  echo "Updated ${INDEX_FILE} to ${TAG}"
else
  echo "Warning: ${INDEX_FILE} not found."
fi
