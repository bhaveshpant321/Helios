#!/bin/bash

# Helios API Quick Start Script
# This script helps you set up and run the Helios API server

set -e

echo "🚀 Helios API Quick Start"
echo "=========================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

echo "✅ Go version: $(go version)"
echo ""

# Check if .env exists
if [ ! -f ".env" ]; then
    echo "⚠️  .env file not found. Creating from .env.example..."
    cp .env.example .env
    echo "📝 Please edit .env file with your database credentials"
    echo "   Then run this script again."
    exit 0
fi

echo "✅ Found .env configuration"
echo ""

# Check if go.mod exists
if [ ! -f "go.mod" ]; then
    echo "❌ go.mod not found. Please run this script from the api directory."
    exit 1
fi

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod download
echo "✅ Dependencies downloaded"
echo ""

# Build the application
echo "🔨 Building application..."
go build -o helios-api main.go
echo "✅ Build successful"
echo ""

# Run the application
echo "🚀 Starting Helios API server..."
echo "   Press Ctrl+C to stop the server"
echo ""
./helios-api
