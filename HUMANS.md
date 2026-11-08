# HUMANS.md

Run and use `go-editorlocator`. Internals: [AGENTS.md](AGENTS.md).

## Quick start

```bash
go get github.com/Rethunk-Tech/go-editorlocator
```

```go
func open(ctx context.Context, dir, file string, line int) error {
 editors := editorlocator.Find() // nil or empty means nothing was found
 if len(editors) == 0 {
  return errors.New("no editor installed")
 }
 chosen := editors[0]
 project := exec.CommandContext(ctx, chosen.Program,
  editorlocator.ProjectArgs(chosen.Program, dir, editorlocator.SolutionIn(dir))...)
 if err := project.Start(); err != nil {
  return err
 }
 at := exec.CommandContext(ctx, chosen.Program,
  editorlocator.FileArgs(chosen.Program, file, line, 2)...)
 return at.Start()
}
```

## Contracts

| Entry point | What it guarantees |
| --- | --- |
| `Find() []Editor` | Returns one entry per program, plus one per installed JetBrains version. Every entry has a non-empty `Name`, a `Program` that exists where it was found, and the `Kind` `KindOf` reports for that program. Returns nil, never an error, when nothing is installed. |
| `Editor` | `Name` is for display only; it carries the version where the install records one. `Program` is the path to run. `Kind` decides which arguments the program accepts. |
| `KindOf(program)` | Reads the program's file name only, so it works on a path that no longer exists. Never panics; anything it does not recognise is `Other`. |
| `ProjectArgs(program, dir, solution)` | Passes a folder for the VS Code family, and the solution when there is one for every other kind. Uses `dir` when `solution` is empty, so a caller may pass `SolutionIn`'s result unchecked. |
| `FileArgs(program, file, line, column)` | Always returns at least the file. `Other` gets the file alone, so a column is dropped rather than refused on a program that has no syntax for it. |
| `SolutionIn(dir)` | Returns the path of the folder's only `.sln` or `.slnx` file, and `""` for none, several, or a folder that cannot be read. |

## Requirements

Go 1.27.3 or newer, plus `lefthook` and `gitleaks` on PATH if you want the git hooks.

```bash
lefthook install
```

That is per clone; git does not commit hooks. Pre-commit scans staged changes for secrets, and
pre-push runs `gate`.

## Verify

```bash
gate
```
