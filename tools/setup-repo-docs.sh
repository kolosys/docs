#!/bin/bash

# Kolosys Documentation Setup Script
# Usage: ./setup-repo-docs.sh <repository-name> [go-version]

set -e

REPO_NAME="$1"
GO_VERSION="${2:-1.22}"

if [ -z "$REPO_NAME" ]; then
    echo "❌ Usage: $0 <repository-name> [go-version]"
    echo "   Example: $0 ion 1.22"
    exit 1
fi

echo "🚀 Setting up Kolosys documentation for $REPO_NAME..."

# Create GitHub workflow
echo "📝 Creating GitHub workflow..."
mkdir -p .github/workflows

cat > .github/workflows/docs.yml << EOF
name: Generate Documentation

on:
  push:
    branches: [main]
    paths: 
      - '**/*.go'
      - 'docs/**'
      - 'docs-templates/**'
      - 'examples/**'
  pull_request:
    branches: [main]
    paths:
      - '**/*.go'
      - 'docs/**'
      - 'docs-templates/**'
      - 'examples/**'
  workflow_dispatch:

# Permissions required for auto-committing generated documentation
permissions:
  contents: write

jobs:
  generate-docs:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0
          token: \${{ secrets.GITHUB_TOKEN }}

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '$GO_VERSION'

      - name: Download Kolosys documentation tools
        run: |
          echo "📥 Downloading Kolosys documentation tools..."
          mkdir -p .kolosys-docs
          
          # Download the shared documentation generator
          curl -sSL https://raw.githubusercontent.com/kolosys/docs/main/shared/scripts/generate-docs.go \\
            -o .kolosys-docs/generate-docs.go

      - name: Generate documentation configuration
        run: |
          if [ ! -f "kolosys-docs.json" ]; then
            echo "📝 Creating default configuration..."
            cat > kolosys-docs.json << 'CONFIG_EOF'
          {
            "repository": {
              "name": "$REPO_NAME",
              "owner": "kolosys",
              "description": "Documentation for $REPO_NAME"
            },
            "packages": [
              {
                "name": "$REPO_NAME",
                "description": "Main package",
                "priority": 1
              }
            ],
            "docs": {
              "root_dir": ".",
              "docs_dir": "docs"
            },
            "output": {
              "generate_combined_api": true,
              "generate_examples": true,
              "verbose": true
            }
          }
          CONFIG_EOF
          fi

      - name: Build documentation
        run: |
          echo "🚀 Building $REPO_NAME documentation..."
          go run .kolosys-docs/generate-docs.go

      - name: Verify generated documentation
        run: |
          echo "📋 Checking generated documentation files..."
          ls -la docs/ || echo "No docs directory found"
          if [ -d "docs" ]; then
            find docs -name "*.md" -type f | head -10
            echo "📊 Total markdown files: \$(find docs -name "*.md" -type f | wc -l)"
          fi

      - name: Cleanup
        run: rm -rf .kolosys-docs

      - name: Commit generated documentation
        if: github.event_name == 'push' && github.ref == 'refs/heads/main'
        uses: stefanzweifel/git-auto-commit-action@v5
        with:
          commit_message: |
            📚 Auto-generate documentation from Go source code
            
            - Updated API documentation
            - Generated examples from source
            - Updated references
            
            [skip ci]
          file_pattern: 'docs/**'
          commit_user_name: 'github-actions[bot]'
          commit_user_email: 'github-actions[bot]@users.noreply.github.com'
          commit_author: 'github-actions[bot] <github-actions[bot]@users.noreply.github.com>'
EOF

# Create default configuration
echo "⚙️ Creating documentation configuration..."
cat > kolosys-docs.json << EOF
{
  "repository": {
    "name": "$REPO_NAME",
    "owner": "kolosys",
    "description": "Documentation for $REPO_NAME"
  },
  "packages": [
    {
      "name": "$REPO_NAME",
      "description": "Main package",
      "priority": 1
    }
  ],
  "docs": {
    "root_dir": ".",
    "docs_dir": "docs"
  },
  "output": {
    "generate_combined_api": true,
    "generate_examples": true,
    "verbose": true
  }
}
EOF

# Create GitBook configuration
echo "📖 Creating GitBook configuration..."
cat > .gitbook.yaml << EOF
root: ./docs

structure:
  introduction: INTRODUCTION.md
  summary: SUMMARY.md

redirects:
  previous/page: new-folder/page.md
EOF

echo "✅ Documentation setup complete for $REPO_NAME!"
echo ""
echo "📁 Created files:"
echo "   .github/workflows/docs.yml"
echo "   kolosys-docs.json"
echo "   .gitbook.yaml"
echo ""
echo "🔗 Next steps:"
echo "1. Commit these files to your repository"
echo "2. Push to trigger documentation generation"
echo "3. Configure GitBook to sync with your repository"
echo ""
echo "🚀 Your documentation will be automatically generated on every push!"
