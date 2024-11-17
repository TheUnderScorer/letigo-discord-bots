#!/bin/bash

# Ensure script stops on errors
set -e

# Read the version from package.json
VERSION=$(jq -r '.version' ../package.json)

# Check if VERSION is not empty
if [ -z "$VERSION" ]; then
    echo "Error: VERSION could not be read from package.json"
    exit 1
fi

# Print the version for confirmation
echo "Building with version: $VERSION"

# Build the Go application
CGO_ENABLED=0 GOOS=linux go build -ldflags "-X app/metadata.Version=${VERSION}" -o ./app