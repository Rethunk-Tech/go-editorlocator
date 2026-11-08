# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's **private vulnerability reporting** on this repository (Security → Report a
vulnerability). It stays private until a fix ships.

Please do not open a public issue for a suspected vulnerability first.

## What is in scope

This package reads the machine it runs on and returns strings; it never runs a program itself.

- **Install paths and the environment.** `Find` expands `${VARS}` in its own install patterns
  through `os.Getenv` and globs the result. A variable holding a path with a glob metacharacter
  (`*`, `?`, `[`) must not escape the pattern's prefix: a malformed pattern is skipped, not
  retried, and `filepath.Glob` returning an error ends that pattern rather than the scan.
- **The Windows registry.** On Windows, `Find` reads `shell\open\command` values for project file
  types. A crafted command must not make it return anything but a program path: `commandExe` cuts
  at the closing quote or the first `.exe`, and a program that does not exist on disk is dropped
  before it reaches the result.
- **`vswhere` output.** `parseVSWhere` decodes JSON from an external program. Malformed or
  oversized input must return no editors, not panic or loop; an entry with no `productPath` is
  skipped.
- **Editor metadata.** `versioned` reads a JetBrains install's `product-info.json`. A missing,
  unreadable or malformed file must leave the plain name rather than fail the scan.
- **`KindOf` and the argument builders.** All four are pure string functions. No input — including
  an empty path, a name with separators, or a negative line — may panic, and `FileArgs` must always
  return at least the file, so a caller that shells out cannot silently start an editor with no
  path.
- **`SolutionIn`.** A directory holding symlinks named `*.sln` must not make it follow or resolve
  them; it reports the name the glob matched, and the caller decides whether to use it.

## What is out of scope

- **What the caller does with the strings.** This package builds an argument vector; it does not
  quote for a shell, and it does not start a process. A caller that joins the arguments into a
  shell command, rather than passing them to `exec.Command`, carries the quoting risk itself.
- **The programs it reports.** `Find` reports editors the user installed and registered. An
  editor that is malicious or misconfigured is the user's choice, not a flaw here.
- **A `Program` that changes after `Find` returns.** The path is not held open or revalidated; a
  caller launching it later executes whatever is there then.
- **Path traversal through arguments.** `FileArgs` passes the file through unchanged. A caller that
  built that path from untrusted input has already made that decision.
