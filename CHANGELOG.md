# Changelog

Notable changes to `go-editorlocator`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.2] - 2026-11-21

### Fixed

- Install folders are ordered by numeric version.
- `KindOf` no longer matches `rider` as a substring, treats any basename starting with `code` as VS Code, and recognizes `vswhere` `devenv` paths on non-Windows.
- A directory named `.sln` is not a solution, and `SolutionIn("")` no longer means this folder.
- Visual Studio Build Tools is not listed as an editor.

## [0.1.1] - 2026-11-12

### Fixed

- A command line holding a character that changes length when lowercased no longer yields a wrong executable path.

## [0.1.0] - 2026-11-08

### Added

- `Find` lists the installed code editors and IDEs, one entry per program and one per installed
  JetBrains version.
- `Kind` and `KindOf` name the family a program belongs to.
- `ProjectArgs` builds the arguments that open a folder or a solution in a given program.
- `FileArgs` builds the arguments that open a file at a line and column in a given program.
- `SolutionIn` reports a folder's single solution file.

[Unreleased]: https://github.com/Rethunk-Tech/go-editorlocator/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/Rethunk-Tech/go-editorlocator/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/Rethunk-Tech/go-editorlocator/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Rethunk-Tech/go-editorlocator/releases/tag/v0.1.0
