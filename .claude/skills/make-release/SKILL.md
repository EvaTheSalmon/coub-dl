---
name: make-release
description: "Use this skill when the user wants to cut a release of coub-dl. Checks the working tree, proposes an annotated-tag message from recent commits, gets explicit approval, then tags (with --cleanup=whitespace), pushes, and monitors the release workflow. Never tags or pushes without explicit approval."
---

# Skill: make-release

Releases are driven by pushing a `v*` tag. `.github/workflows/release.yml` cross-compiles
the binaries (linux/darwin/windows × amd64/arm64), archives them, and publishes the GitHub
release. The annotated tag message becomes the release body.

## Critical rule

**Never create the tag or push without the user explicitly approving the version and the
message.** Propose, wait, then execute.

## Step 1: Check the working tree

```bash
git status
git diff --stat
```

`master` is protected — direct pushes are rejected, changes land via PR. Do not tag a dirty
tree. If the release needs unmerged work, get it merged first.

## Step 2: Determine the version

Ask if not given. Format `vMAJOR.MINOR.PATCH`; prepend `v` to a bare number. Show recent tags:

```bash
git tag --sort=-version:refname | head -5
```

Tag the latest `origin/master`:

```bash
git fetch origin && git switch master && git pull --ff-only
```

## Step 3: Propose the tag message

Read commits since the last tag and draft concise notes. Present them as an editable
proposal — do not make the user write from scratch.

```bash
git log --oneline $(git describe --tags --abbrev=0 2>/dev/null)..HEAD 2>/dev/null || git log --oneline -10
```

Three quirks to honor (all learned the hard way on v1.0.0):

1. **`git tag` strips lines starting with `#` by default** — its default cleanup mode is
   `strip`, which treats `#` lines as comments and removes them, so markdown `##` headings
   vanish. ALWAYS tag with `--cleanup=whitespace` (Step 4).
2. **`release.yml` pipes the body through `sed '/^$/d'`**, deleting every blank line. Even
   with headings preserved, section spacing collapses. `##` headings and bullet lists still
   render (they interrupt paragraphs in GitHub markdown), but consecutive prose paragraphs
   merge. So draft with `##` sections + bullet lists, avoid multi-paragraph prose, and do
   not end a section on prose that follows a list. The real fix is removing that `sed` from
   the workflow.
3. **Do not repeat the version as the first body line** — the workflow sets the release
   `name:` to the tag (e.g. `v1.0.0`), so the page heading already shows it.

## Step 4: Get approval and execute

After the user approves the version and message:

```bash
git tag -a <version> --cleanup=whitespace -F - <<'EOF'
<approved message>
EOF
git push origin <version>
```

`--cleanup=whitespace` is mandatory — without it the `##` headings are stripped.

## Step 5: Monitor the workflow

```bash
gh run list --workflow=release.yml --limit 1
gh run watch <run-id> --exit-status
gh release view <version>
```

- Pass: report `https://github.com/EvaTheSalmon/coub-dl/releases/tag/<version>` and confirm
  the 5 assets are attached.
- Fail: show the failing step and logs, ask how to proceed.

## Token note

The local `gh` PAT can push tags (it has `contents`) but **cannot create/edit PRs or edit a
published release** — those return `403 Resource not accessible by personal access token`.
The release itself still publishes (Actions uses its own `GITHUB_TOKEN`). If the published
body needs a post-publish fix, edit it in the web UI:
`https://github.com/EvaTheSalmon/coub-dl/releases/edit/<version>`.

## Redo a release (delete and retag)

With user approval:

```bash
gh release delete <version> --yes   # may 403 with the local PAT — delete via web if so
git push origin --delete <version>
git tag -d <version>
```

Then re-run from Step 3 with the corrected version or message.

## Test the cross-compile locally (optional)

Mirror the workflow without publishing:

```bash
for t in linux/amd64 darwin/arm64 windows/amd64; do
  GOOS=${t%/*} GOARCH=${t#*/} CGO_ENABLED=0 \
    go build -trimpath -ldflags "-s -w" -o /tmp/coub-dl-${t//\//-} . && echo "OK $t"
done
```

## Reminders

- The release workflow triggers on any `v*` tag push (`.github/workflows/release.yml`).
- Always annotated tags (`git tag -a`), never lightweight.
- Always `--cleanup=whitespace`, so `#` headings survive.
- ffmpeg is a runtime dependency; the binaries do not bundle it.
