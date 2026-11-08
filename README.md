<h1 align="center">go-editorlocator</h1>

<div align="center">

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.27.3+-00ADD8?logo=go&logoColor=white)](go.mod)

</div>

---

Find installed code editors and IDEs and open a folder or file in one. Package `editorlocator`, module `github.com/Rethunk-Tech/go-editorlocator`.

Each editor's command line differs: one takes a folder and a `file:line:column`, another a solution and a separate
line argument, a third only a file. This package answers both halves in one place — which editors are installed on
this computer, and the arguments that open a path in the one the caller picked — so a caller never carries its own
table of editor flags.

## Quick start

```bash
go get github.com/Rethunk-Tech/go-editorlocator
```

A worked example and the contracts: [HUMANS.md](HUMANS.md).

## Features

- `Find` lists every editor and IDE it can see: each program on `PATH`, the VS Code, VSCodium, Rider and Cursor
  install folders, and on Windows the Visual Studios from the installer's own tool plus the programs registered for
  `.sln`, `.slnx` and `.csproj`.
- One entry per program, and one per installed JetBrains version, named with the version it records.
- Every `Editor` carries its `Kind`, so a caller can group or filter the list without re-reading the path.
- `ProjectArgs` and `FileArgs` build the arguments for a folder, a solution or a file at a line and column, in the
  form that editor takes.
- `SolutionIn` reports a folder's single solution file and returns nothing when there is none or a choice of several.
- The discovery table and the injected lookups are unexported, so the search runs against a fake filesystem in tests.

## Documentation

| Doc | Audience |
| --- | --- |
| [HUMANS.md](HUMANS.md) | Running and using the package |
| [AGENTS.md](AGENTS.md) | Internals, invariants, and why |
| [CONTRIBUTING.md](CONTRIBUTING.md) | Process — commits, tests, docs |
| [SECURITY.md](SECURITY.md) | Vulnerability disclosure |
| [CHANGELOG.md](CHANGELOG.md) | Release notes |

## License

MIT — see [LICENSE](LICENSE). © 2026 Rethunk.Tech, LLC.
