Run checks and commit for the specdrift project. Execute all of the following steps in order. Stop at the first failure.

1. **gofmt**: `gofmt -l .` — if any files are listed, run `gofmt -w .` to fix them and report which files were formatted
2. **go vet**: `go vet ./...`
3. **test**: `go test ./...`
4. **specdrift check**: `go run . check 'docs/spec/*.md'`
   - If DRIFT is detected, do NOT just run update. First read the drifted spec section and the changed source code, then determine whether the spec text needs to be revised to reflect the source change. Update the spec content if needed, then run `go run . update 'docs/spec/*.md'` to sync the hashes.
5. **coverage check**: `go run . coverage --src 'internal/*.go' --src '*.go' 'docs/spec/*.md'`
   - If there are "Not covered" files, show the list to the user and ask whether to add documentation for any of them before committing.
   - If the user wants to add documentation, update the relevant spec files, run `go run . update 'docs/spec/*.md'` to sync hashes, then re-run steps 3-4 before proceeding.
   - If the user decides no documentation is needed, proceed to commit.
6. **commit**: If all checks pass, stage all changes and create a commit. Write a concise commit message in Japanese summarizing the changes. End with `Co-Authored-By: Claude Opus 4.6 (1M context) <noreply@anthropic.com>`.
