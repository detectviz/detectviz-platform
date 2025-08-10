#!/bin/bash

# Setup script for Detectviz Platform git hooks
# Installs pre-commit hooks to ensure contract consistency

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
HOOKS_DIR="$REPO_ROOT/.github/hooks"
GIT_HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "🔧 Setting up Detectviz Platform git hooks..."

# Check if we're in a git repository
if [[ ! -d "$REPO_ROOT/.git" ]]; then
    echo "❌ Not in a git repository"
    exit 1
fi

# Create .git/hooks directory if it doesn't exist
mkdir -p "$GIT_HOOKS_DIR"

# Install pre-commit hook
if [[ -f "$HOOKS_DIR/pre-commit" ]]; then
    cp "$HOOKS_DIR/pre-commit" "$GIT_HOOKS_DIR/pre-commit"
    chmod +x "$GIT_HOOKS_DIR/pre-commit"
    echo "✅ Pre-commit hook installed"
else
    echo "❌ Pre-commit hook source not found at $HOOKS_DIR/pre-commit"
    exit 1
fi

# Provide usage information
echo ""
echo "🎉 Git hooks setup complete!"
echo ""
echo "The pre-commit hook will now:"
echo "  • Validate contracts when proto files change"
echo "  • Check Go code formatting and compilation"
echo "  • Verify Python imports and basic syntax"
echo "  • Regenerate contract stubs if needed"
echo ""
echo "Environment variables you can set:"
echo "  DETECTVIZ_AUTO_STAGE_CONTRACTS=true   # Auto-stage generated files"
echo ""
echo "To bypass hooks (emergency only):"
echo "  git commit --no-verify"
echo ""

# Offer to run tool installation
read -p "🛠️  Install required tools now? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📦 Installing tools..."
    cd "$REPO_ROOT/contracts"
    make install-tools
    echo "✅ Tools installation complete"
fi

echo "🎊 Setup finished! Your next commit will be validated automatically."