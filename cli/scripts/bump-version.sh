#!/bin/bash
# Auto-increment patch version
# Usage: ./scripts/bump-version.sh

VERSION_FILE="VERSION"

if [ ! -f "$VERSION_FILE" ]; then
    echo "1.0.0" > "$VERSION_FILE"
fi

# Read current version
CURRENT=$(cat "$VERSION_FILE")

# Split version into parts
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

# Increment patch version
PATCH=$((PATCH + 1))

# Create new version
NEW_VERSION="$MAJOR.$MINOR.$PATCH"

# Write back to file
echo "$NEW_VERSION" > "$VERSION_FILE"

# Output new version (for use in Makefile)
echo "$NEW_VERSION"
