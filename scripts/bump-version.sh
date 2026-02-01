#!/bin/bash

# Bump version script
# Usage: ./scripts/bump-version.sh [major|minor|patch]

VERSION_FILE="cmd/api/version.go"

# Get current version
CURRENT_VERSION=$(grep 'Version.*=' $VERSION_FILE | sed 's/.*"\(.*\)"/\1/')

if [ -z "$CURRENT_VERSION" ]; then
    echo "Error: Could not read current version"
    exit 1
fi

# Parse version
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Bump version based on argument
case $1 in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
    *)
        echo "Usage: $0 [major|minor|patch]"
        echo "Current version: $CURRENT_VERSION"
        exit 1
        ;;
esac

NEW_VERSION="$MAJOR.$MINOR.$PATCH"

# Update version.go
sed -i "s/Version.*=.*/Version   = \"$NEW_VERSION\"/" $VERSION_FILE

echo "Version bumped: $CURRENT_VERSION -> $NEW_VERSION"

# Output new version for use in other scripts
echo $NEW_VERSION
