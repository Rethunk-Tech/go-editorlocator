# Security

## Supported versions

The latest tagged release. There are no long-lived release branches, so a fix ships in the next tag
rather than being backported.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting on this repository: <https://github.com/Rethunk-Tech/go-editorlocator/security/advisories/new>, or the **Security** tab, then **Report a vulnerability**. The report stays private between you and us until a fix ships.

If private reporting is unavailable, email [oss@rethunk.tech](mailto:oss@rethunk.tech) with **[SECURITY]** in the subject.

Please do not open a public issue, pull request or discussion for a suspected vulnerability.

A useful report has the affected repository and version (or commit), steps to reproduce, what an attacker gets and under which conditions, and whether anyone else knows.

## What counts as a vulnerability

A vulnerability is a flaw that lets someone cross a boundary they should not: run code, read or change data they have no right to, escape a sandbox or path restriction, or leak a credential. A bug is anything else that misbehaves, such as a crash on your own input, wrong output or a slow path; open an ordinary issue for those. If you are unsure which one you have, report it privately and we will say.

## Coordinated disclosure

We handle the report inside the private advisory, so you see each status change and can comment on the fix. The advisory goes public when the fix ships, or 90 days after your report, whichever comes first. If a fix is close, either of us can ask for more time in the thread. You are credited in the advisory unless you would rather not be.

## Response timelines

- We acknowledge your report within 3 business days.
- We give you an assessment (accepted, needs more information or declined, with reasons) within 10 business days.
- We aim to ship a fix within 90 days of the report.

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

## Safe harbour

We will not pursue legal action against good-faith research that avoids privacy violations, data destruction and disruption of services, stops at a proof of concept, and keeps the details private until the advisory is published. Rethunk.Tech, LLC maintains this repository.
