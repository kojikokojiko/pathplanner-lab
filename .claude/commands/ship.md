Commit all changes and push to the current branch.

1. Run `git status` to see what changed
2. Run `git diff` to review the changes
3. Stage relevant files (avoid .env, secrets)
4. Write a concise commit message in Japanese or English based on the changes
5. Commit and push

Format:
```
git add <specific files>
git commit -m "$(cat <<'EOF'
<type>: <summary>

<detail if needed>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
EOF
)"
git push
```

Types: feat / fix / refactor / chore / docs
