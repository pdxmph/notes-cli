#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from argument or prompt
VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo -n "Enter version (e.g., v0.1.0): "
    read VERSION
fi

# Validate version format
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Version must be in format vX.Y.Z${NC}"
    exit 1
fi

echo -e "${GREEN}🚀 Starting release process for $VERSION${NC}"

# Step 1: Ensure all changes are committed
echo -e "\n${YELLOW}Step 1: Checking for uncommitted changes...${NC}"
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo -e "${RED}Error: Uncommitted changes found. Please commit or stash them first.${NC}"
    exit 1
fi

# Step 2: Create and push tag
echo -e "\n${YELLOW}Step 2: Creating tag...${NC}"
if git tag | grep -q "^${VERSION}$"; then
    echo -e "${RED}Error: Tag $VERSION already exists${NC}"
    exit 1
else
    echo "Enter release notes (press Ctrl-D when done):"
    RELEASE_NOTES=$(cat)
    git tag -a "$VERSION" -m "$RELEASE_NOTES"
fi

# Step 3: Push tag to GitHub
echo -e "\n${YELLOW}Step 3: Pushing tag to GitHub...${NC}"
git push origin "$VERSION"

# Step 4: Build binaries for multiple platforms
echo -e "\n${YELLOW}Step 4: Building binaries...${NC}"
rm -rf dist
mkdir -p dist

# Build for multiple platforms
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    OS="${PLATFORM%/*}"
    ARCH="${PLATFORM#*/}"
    
    echo "Building for $OS/$ARCH..."
    OUTPUT_NAME="notes-cli"
    if [ "$OS" = "windows" ]; then
        OUTPUT_NAME="notes-cli.exe"
    fi
    
    GOOS=$OS GOARCH=$ARCH go build -o "dist/${OUTPUT_NAME}" .
    
    # Create archive
    ARCHIVE_NAME="notes-cli_${VERSION}_${OS}_${ARCH}.tar.gz"
    cd dist
    cp ../README.md ../LICENSE .
    
    # Copy completions
    mkdir -p completions
    cp ../completions/* completions/
    
    FILES_TO_ARCHIVE="$OUTPUT_NAME README.md LICENSE completions"
    tar czf "$ARCHIVE_NAME" $FILES_TO_ARCHIVE
    
    # Clean up
    rm -f "$OUTPUT_NAME" README.md LICENSE
    rm -rf completions
    cd ..
done

# Create checksums
cd dist
shasum -a 256 *.tar.gz > checksums.txt
cd ..

# Step 5: Create GitHub release
echo -e "\n${YELLOW}Step 5: Creating GitHub release...${NC}"
GITHUB_REPO="pdxmph/notes-cli"

if ! command -v gh &> /dev/null; then
    echo -e "${RED}Error: GitHub CLI (gh) not found. Install with: brew install gh${NC}"
    echo -e "${YELLOW}Alternatively, create the release manually at:${NC}"
    echo "https://github.com/$GITHUB_REPO/releases/new?tag=$VERSION"
    echo "Upload these files:"
    ls -la dist/*.tar.gz dist/checksums.txt
else
    gh release create "$VERSION" \
      --repo "$GITHUB_REPO" \
      --title "$VERSION" \
      --notes "$RELEASE_NOTES" \
      dist/*.tar.gz \
      dist/checksums.txt
fi

echo -e "\n${GREEN}✅ Release $VERSION completed successfully!${NC}"
echo -e "${GREEN}View at: https://github.com/$GITHUB_REPO/releases/tag/$VERSION${NC}"

# Show installation instructions
echo -e "\n${YELLOW}Installation instructions:${NC}"
echo "macOS (Intel):"
echo "  curl -L https://github.com/pdxmph/notes-cli/releases/download/$VERSION/notes-cli_${VERSION}_darwin_amd64.tar.gz | tar xz"
echo "  sudo mv notes-cli /usr/local/bin/"
echo ""
echo "macOS (Apple Silicon):"
echo "  curl -L https://github.com/pdxmph/notes-cli/releases/download/$VERSION/notes-cli_${VERSION}_darwin_arm64.tar.gz | tar xz"
echo "  sudo mv notes-cli /usr/local/bin/"
echo ""
echo "Linux (amd64):"
echo "  curl -L https://github.com/pdxmph/notes-cli/releases/download/$VERSION/notes-cli_${VERSION}_linux_amd64.tar.gz | tar xz"
echo "  sudo mv notes-cli /usr/local/bin/"
echo ""
echo "After installation, install completions with:"
echo "  sudo cp completions/_notes-cli /usr/local/share/zsh/site-functions/"
echo "  sudo cp completions/notes-cli.bash /usr/local/etc/bash_completion.d/"
