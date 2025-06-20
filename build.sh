#!/bin/bash

# Build script for goBodyFile
# This script builds the application for multiple platforms

echo "Building goBodyFile..."

# Build for current platform
echo "Building for current platform..."
go build -o gobodyfile

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o gobodyfile-linux-amd64

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o gobodyfile-windows-amd64.exe

# Build for macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -o gobodyfile-darwin-amd64

echo "Build complete!"
echo "Generated binaries:"
ls -la gobodyfile* 