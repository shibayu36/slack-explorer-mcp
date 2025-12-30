#!/bin/bash

set -e

# Check if version is provided
if [ -z "$1" ]; then
    echo "Usage: ./release.sh <version>"
    echo "Example: ./release.sh 1.0.1"
    exit 1
fi

VERSION=$1

# Validate version format (basic semantic versioning)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Version must be in format X.Y.Z (e.g., 1.0.1)"
    exit 1
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "Error: You have uncommitted changes. Please commit or stash them first."
    exit 1
fi

# Update version constant in main.go
echo "Updating version to $VERSION..."
sed -i '' "s/const Version = \".*\"/const Version = \"$VERSION\"/" main.go

# Update version in server.json
echo "Updating server.json version..."
sed -i '' "s/\"version\": \"[^\"]*\"/\"version\": \"$VERSION\"/" server.json
sed -i '' "s|ghcr.io/shibayu36/slack-explorer-mcp:[^\"]*|ghcr.io/shibayu36/slack-explorer-mcp:$VERSION|" server.json

# Run tests to ensure everything works
echo "Running tests..."
go test ./...

# Build to ensure compilation
echo "Building..."
go build -o slack-explorer-mcp

# Commit version change
echo "Committing version change..."
git add main.go server.json
git commit -m "Release v$VERSION"

# Create annotated tag
echo "Creating tag v$VERSION..."
git tag -a "v$VERSION" -m "Release v$VERSION"

# Push commits and tags
echo "Pushing to remote..."
git push
git push --tags

echo "✅ Release v$VERSION completed successfully!"
echo ""
echo "Next steps:"
echo "1. Create a GitHub release from the tag"
echo "2. Build binaries for different platforms if needed"
echo "3. Update any documentation or changelogs"
