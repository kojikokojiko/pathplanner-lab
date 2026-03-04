Create a pull request for the current branch.

1. Run `git status` and `git diff main...HEAD` to understand all changes
2. Run `git log main...HEAD --oneline` to see commits
3. Draft a clear PR title (under 60 chars) and summary based on the actual changes
4. Push the branch if not already pushed: `git push -u origin HEAD`
5. Create PR with `gh pr create` using this format:

```
gh pr create --title "<title>" --body "$(cat <<'EOF'
## 概要
<変更内容を箇条書き>

## テスト方法
- [ ] `make dev` でローカル起動確認
- [ ] 該当機能の動作確認

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Return the PR URL when done.
