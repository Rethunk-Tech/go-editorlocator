# AGENTS.md

Package `editorlocator` finds the code editors and IDEs installed on a computer and builds the
arguments that open a project, a solution or a file at a line in one. A caller that launches an
editor imports it; it imports no consumer package and knows no product, setting or path belonging
to one. Its only exports are the two argument builders, the solution lookup, the finder, and the
`Editor` and `Kind` types.

## Gate

`gate` is the command. Each override in `.gate.toml` is there because a runner
cannot resolve that check from PATH. The `windows-latest` test job is the only place
Windows-only behaviour runs.

`TestAssociatedEditorsReadsTheProgramsRegisteredForAFileType` is the test only that job runs: it
writes a throwaway ProgID under `HKCU\Software\Classes` and reads it back through
`associatedEditors`. Every other test, including the injected-host discovery test, runs on every
OS.

## Shape

- `editorHost` stays unexported with its three fields injected, and `Find` is the only thing that
  fills them from the real machine. The discovery rules are worth testing against a fake
  filesystem, and a package-level hook would let two tests in one binary disagree about `PATH`.
- `Find` is the single path to a list of editors. Anything else that wants the list calls it, so
  deduplication and `Kind` assignment happen once, in `editorHost.add`.
- `KindOf` reads the program's file name and nothing else, because a caller may hold a path from a
  setting that no longer exists, and a `Kind` must still be answerable for it.
- `ProjectArgs` and `FileArgs` never return an empty slice: every program can at least be handed a
  path, so the `Other` fallback is the floor and the per-family cases are the extras.
- `SolutionIn` returning `""` on several candidates is policy, not a bound. One solution file is a
  project; two are a choice the caller has to make, and picking one here would open the wrong one.
- The `Kind` zero value is `Other`, and the names are exported constants in the order
  `KindOf` distinguishes them. Adding a family means adding a case to `KindOf` and a case to each
  argument builder; the `exhaustive` linter checks the latter.
- `editorSites` order is preference order: `Find`'s first entry is the one a caller that takes
  `[0]` should get. PATH beats an install folder, and the newest version of a folder pattern beats
  older ones.
- `//go:build windows` splits the platform discovery. Any file gaining a build tag adds its GOOS to
  the `crosscompile` gate in `.gate.toml`.

## Files

| File | Holds |
| --- | --- |
| `doc.go` | Package doc. |
| `editors.go` | `Editor`, `Kind`, `Find`, `KindOf`, `ProjectArgs`, `FileArgs`, `SolutionIn`, the install-site table, and the unexported helpers `editorHost`, `versioned`, `parseVSWhere` and `commandExe`. |
| `editors_other.go` | `platformEditors` on everything but Windows, where it adds nothing. |
| `editors_windows.go` | `platformEditors` on Windows: vswhere, the registry lookups and the program label. |
