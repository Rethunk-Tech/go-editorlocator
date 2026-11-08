# Changelog

Notable changes to `go-editorlocator`. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-11-08

### Added

- `Find` lists the installed code editors and IDEs, one entry per program and one per installed
  JetBrains version.
- `Kind` and `KindOf` name the family a program belongs to.
- `ProjectArgs` builds the arguments that open a folder or a solution in a given program.
- `FileArgs` builds the arguments that open a file at a line and column in a given program.
- `SolutionIn` reports a folder's single solution file.

[Unreleased]: https://github.com/Rethunk-Tech/go-editorlocator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/Rethunk-Tech/go-editorlocator/releases/tag/v0.1.0
