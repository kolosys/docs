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
      - 'kolosys-docs.json'
  pull_request:
    branches: [main]
    paths:
      - '**/*.go'
      - 'docs/**'
      - 'docs-templates/**'
      - 'examples/**'
      - 'kolosys-docs.json'
  workflow_dispatch:

jobs:
  generate-docs:
    uses: kolosys/docs/.github/workflows/docs-workflow.yml@main
    with:
      repository_name: "$REPO_NAME"
      go_version: "$GO_VERSION"
      generate_examples: true
      skip_commit: false
      create_pr: true  # Use PR approach for protected branches
    permissions:
      contents: write
      pull-requests: write
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
