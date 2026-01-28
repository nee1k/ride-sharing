#!/bin/bash
# Helper script to create a feature branch, commit changes, create PR, and merge it

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get feature name from argument or generate from branch name
FEATURE_NAME=${1:-$(git branch --show-current | sed 's/feature\///' | sed 's/\//-/g')}

if [ -z "$FEATURE_NAME" ] || [ "$FEATURE_NAME" = "main" ]; then
    echo -e "${YELLOW}⚠️  Please provide a feature name or run from a feature branch${NC}"
    echo "Usage: $0 <feature-name>"
    echo "Example: $0 add-trip-service"
    exit 1
fi

BRANCH_NAME="feature/${FEATURE_NAME}"

echo -e "${BLUE}🚀 Creating and merging PR for: ${FEATURE_NAME}${NC}"

# Check if we're on main
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" = "main" ]; then
    echo -e "${BLUE}📦 Creating feature branch: ${BRANCH_NAME}${NC}"
    git checkout -b "$BRANCH_NAME"
else
    BRANCH_NAME="$CURRENT_BRANCH"
    echo -e "${BLUE}📦 Using current branch: ${BRANCH_NAME}${NC}"
fi

# Check for changes
if [ -z "$(git status --porcelain)" ]; then
    echo -e "${YELLOW}⚠️  No changes to commit${NC}"
    exit 1
fi

# Show status
echo -e "${BLUE}📋 Current changes:${NC}"
git status --short

# Commit changes
echo -e "${BLUE}💾 Committing changes...${NC}"
read -p "Enter commit message (or press Enter for default): " COMMIT_MSG

if [ -z "$COMMIT_MSG" ]; then
    COMMIT_MSG="feat: ${FEATURE_NAME}"
fi

git add .
git commit -m "$COMMIT_MSG"

# Push branch
echo -e "${BLUE}📤 Pushing branch to remote...${NC}"
git push -u origin "$BRANCH_NAME"

# Create PR
echo -e "${BLUE}📝 Creating pull request...${NC}"
PR_NUMBER=$(gh pr create --title "$COMMIT_MSG" --body "## Changes

This PR implements: ${FEATURE_NAME}

### Changes Made
- See commit message for details

### Testing
- [ ] Tested locally
- [ ] Ready for review" --base main --json number --jq '.number')

if [ -z "$PR_NUMBER" ]; then
    echo -e "${YELLOW}⚠️  Failed to create PR${NC}"
    exit 1
fi

echo -e "${GREEN}✅ PR #${PR_NUMBER} created: https://github.com/$(git remote get-url origin | sed 's/.*github.com[:/]\(.*\)\.git/\1/')/pull/${PR_NUMBER}${NC}"

# Merge PR
echo -e "${BLUE}🔀 Merging PR...${NC}"
gh pr merge "$PR_NUMBER" --merge --delete-branch

# Switch back to main and pull
echo -e "${BLUE}🔄 Updating local main branch...${NC}"
git checkout main
git pull origin main

echo -e "${GREEN}✅ Successfully merged PR #${PR_NUMBER} to main!${NC}"
