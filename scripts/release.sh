#!/bin/bash

# Release script for arena package
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if version is provided
if [ $# -eq 0 ]; then
    print_error "Please provide a version number (e.g., v1.0.0)"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    print_error "Version must be in format vX.Y.Z (e.g., v1.0.0)"
    exit 1
fi

print_status "Preparing release $VERSION"

# Check if we're on main branch
BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$BRANCH" != "main" ]; then
    print_warning "You're not on main branch (current: $BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check for uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    print_error "You have uncommitted changes. Please commit or stash them first."
    exit 1
fi

# Run tests
print_status "Running tests..."
go test -v -race ./...

# Run benchmarks
print_status "Running benchmarks..."
go test -bench=. -benchmem ./...

# Check formatting
print_status "Checking code formatting..."
if [ -n "$(gofmt -l .)" ]; then
    print_error "Code is not properly formatted. Run 'go fmt ./...'"
    exit 1
fi

# Run go vet
print_status "Running go vet..."
go vet ./...

# Update go.mod
print_status "Updating go.mod..."
go mod tidy

# Create git tag
print_status "Creating git tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"

print_status "Release $VERSION is ready!"
print_status "To publish:"
print_status "  1. git push origin main"
print_status "  2. git push origin $VERSION"
print_status "  3. Create GitHub release at: https://github.com/pavanmanishd/arena/releases/new"

print_warning "Don't forget to update CHANGELOG.md if needed!"
