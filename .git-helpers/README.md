# Git Helper Scripts

## create-and-merge-pr.sh

Automatically creates a feature branch, commits changes, creates a PR, and merges it to main.

### Usage

```bash
# From main branch with changes
./.git-helpers/create-and-merge-pr.sh <feature-name>

# Example
./.git-helpers/create-and-merge-pr.sh add-trip-service

# Or from a feature branch
git checkout -b feature/my-feature
# ... make changes ...
./.git-helpers/create-and-merge-pr.sh
```

### What it does

1. Creates a feature branch (if on main)
2. Shows current changes
3. Commits all changes
4. Pushes branch to remote
5. Creates a GitHub PR
6. Merges the PR automatically
7. Updates local main branch

### Requirements

- GitHub CLI (`gh`) installed and authenticated
- Changes staged or unstaged (will be added automatically)
