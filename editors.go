package editorlocator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

// Kind is the family a program belongs to, which decides what it is given to open.
type Kind int

// The families KindOf tells apart.
const (
	// Other is a program no family below claims: it is given a path and told nothing else.
	Other Kind = iota
	// VSCodeFamily opens a folder, and a file at a line and column.
	VSCodeFamily
	// JetBrains opens a solution, and a file at a line.
	JetBrains
	// VisualStudio opens a solution, and a file without a line.
	VisualStudio
)

// Editor is one installed program a project can open in; each installed version is its own.
type Editor struct {
	Name    string `json:"name"`
	Program string `json:"program"`
	Kind    Kind   `json:"kind"`
}

// Where each editor installs, as globs under environment variables; the order is the order Find tries them.
var editorSites = []struct {
	name    string
	bin     string
	windows []string
	unix    []string
}{
	{
		name: "VS Code",
		bin:  "code",
		windows: []string{
			"${LOCALAPPDATA}/Programs/Microsoft VS Code/bin/code.cmd",
			"${ProgramFiles}/Microsoft VS Code/bin/code.cmd",
		},
		unix: []string{
			"/snap/bin/code",
			"/var/lib/flatpak/exports/bin/com.visualstudio.code",
			"${HOME}/.local/share/flatpak/exports/bin/com.visualstudio.code",
		},
	},
	{
		name: "VS Code Insiders",
		bin:  "code-insiders",
		windows: []string{
			"${LOCALAPPDATA}/Programs/Microsoft VS Code Insiders/bin/code-insiders.cmd",
			"${ProgramFiles}/Microsoft VS Code Insiders/bin/code-insiders.cmd",
		},
		unix: []string{"/snap/bin/code-insiders"},
	},
	{
		name: "VSCodium",
		bin:  "codium",
		windows: []string{
			"${LOCALAPPDATA}/Programs/VSCodium/bin/codium.cmd",
			"${ProgramFiles}/VSCodium/bin/codium.cmd",
		},
		unix: []string{
			"/snap/bin/codium",
			"/var/lib/flatpak/exports/bin/com.vscodium.codium",
			"${HOME}/.local/share/flatpak/exports/bin/com.vscodium.codium",
		},
	},
	{
		name: "Rider",
		bin:  "rider",
		windows: []string{
			"${LOCALAPPDATA}/JetBrains/Toolbox/apps/Rider/bin/rider64.exe",
			"${LOCALAPPDATA}/JetBrains/Toolbox/apps/Rider/ch-*/*/bin/rider64.exe",
			"${LOCALAPPDATA}/Programs/Rider*/bin/rider64.exe",
			"${ProgramFiles}/JetBrains/JetBrains Rider*/bin/rider64.exe",
		},
		unix: []string{
			"${HOME}/.local/share/JetBrains/Toolbox/apps/rider/bin/rider.sh",
			"${HOME}/.local/share/JetBrains/Toolbox/apps/rider/ch-*/*/bin/rider.sh",
			"${HOME}/.local/share/JetBrains/Toolbox/apps/Rider/ch-*/*/bin/rider.sh",
			"/opt/rider*/bin/rider.sh",
			"/opt/JetBrains Rider*/bin/rider.sh",
			"/snap/bin/rider",
			"/var/lib/flatpak/exports/bin/com.jetbrains.Rider",
			"${HOME}/.local/share/flatpak/exports/bin/com.jetbrains.Rider",
		},
	},
	{
		name:    "Cursor",
		bin:     "cursor",
		windows: []string{"${LOCALAPPDATA}/Programs/cursor/resources/app/bin/cursor.cmd"},
	},
}

// editorHost is what finding editors reads from the computer, as fields so a test can stand in for it.
type editorHost struct {
	windows  bool
	getenv   func(string) string
	lookPath func(string) (string, error)
}

// Find lists the editors installed on this computer, each installed version on its own. The first entry is the one
// the table's order prefers: PATH before an install folder, and the newest version of a folder pattern before older
// ones.
func Find() []Editor {
	host := editorHost{windows: runtime.GOOS == "windows", getenv: os.Getenv, lookPath: exec.LookPath}
	out := host.find()
	for _, e := range platformEditors() {
		out = host.add(out, e)
	}
	return out
}

// find lists each known editor on PATH and in its install folders, the newest version of a folder pattern first.
func (h editorHost) find() []Editor {
	var out []Editor
	for _, site := range editorSites {
		if path, err := h.lookPath(site.bin); err == nil {
			out = h.add(out, Editor{Name: site.name, Program: path})
		}
		patterns := site.unix
		if h.windows {
			patterns = site.windows
		}
		for _, pattern := range patterns {
			found, err := filepath.Glob(filepath.FromSlash(os.Expand(pattern, h.getenv)))
			if err != nil {
				continue
			}
			slices.Reverse(found)
			for _, path := range found {
				out = h.add(out, Editor{Name: versioned(site.name, path), Program: path})
			}
		}
	}
	return out
}

// add keeps one entry per program: PATH often holds a link to a folder that is searched too. It also fills Kind, so
// every editor this package returns carries the family its arguments are built from.
func (h editorHost) add(list []Editor, e Editor) []Editor {
	key := func(path string) string {
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		if h.windows {
			return strings.ToLower(path)
		}
		return path
	}
	if slices.ContainsFunc(list, func(have Editor) bool { return key(have.Program) == key(e.Program) }) {
		return list
	}
	e.Kind = KindOf(e.Program)
	return append(list, e)
}

// versioned adds the version a JetBrains install records, so two versions of an editor can be told apart.
func versioned(name, program string) string {
	data, err := os.ReadFile(filepath.Join(filepath.Dir(filepath.Dir(program)), "product-info.json"))
	if err != nil {
		return name
	}
	var info struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(data, &info) != nil || info.Version == "" {
		return name
	}
	return name + " " + info.Version
}

// parseVSWhere reads vswhere's JSON: one entry per Visual Studio installed, versions side by side.
func parseVSWhere(data []byte) []Editor {
	var installs []struct {
		DisplayName string `json:"displayName"`
		ProductPath string `json:"productPath"`
		Catalog     struct {
			Version string `json:"productDisplayVersion"`
		} `json:"catalog"`
	}
	if json.Unmarshal(data, &installs) != nil {
		return nil
	}
	out := make([]Editor, 0, len(installs))
	for _, install := range installs {
		if KindOf(install.ProductPath) != VisualStudio {
			continue
		}
		name := strings.TrimSpace(install.DisplayName + " " + install.Catalog.Version)
		out = append(out, Editor{Name: name, Program: install.ProductPath})
	}
	return out
}

// commandExe is the program a Windows shell\open\command value runs: quoted, or up to its ".exe" since an unquoted
// path may hold spaces.
func commandExe(command string) string {
	command = strings.TrimSpace(command)
	if rest, ok := strings.CutPrefix(command, `"`); ok {
		exe, _, _ := strings.Cut(rest, `"`)
		return exe
	}
	// .exe is ASCII; ToLower can grow a string (İ → i̇), so an index into the
	// lowercased copy is not a valid slice of command.
	for i := 0; i+len(".exe") <= len(command); i++ {
		if strings.EqualFold(command[i:i+len(".exe")], ".exe") {
			return command[:i+len(".exe")]
		}
	}
	return ""
}

// KindOf reports the family a program is in, read from its name alone. An installer may move the program or rename
// the folder around it, but the name of the program itself is what decides how it reads its arguments.
func KindOf(program string) Kind {
	base := strings.ToLower(filepath.Base(program))
	switch {
	case strings.Contains(base, "rider"):
		return JetBrains
	case strings.HasPrefix(base, "devenv"):
		return VisualStudio
	case strings.HasPrefix(base, "code"), strings.HasPrefix(base, "codium"), strings.HasPrefix(base, "cursor"),
		strings.HasPrefix(base, "com.visualstudio.code"), strings.HasPrefix(base, "com.vscodium.codium"):
		return VSCodeFamily
	}
	return Other
}

// ProjectArgs is the arguments that open a project in program: the VS Code family works on folders, and the rest
// open the solution when there is one.
func ProjectArgs(program, dir, solution string) []string {
	if KindOf(program) == VSCodeFamily || solution == "" {
		return []string{dir}
	}
	return []string{solution}
}

// FileArgs is the arguments that open a file in program at a line and column, in the form that program takes.
func FileArgs(program, file string, line, column int) []string {
	switch KindOf(program) {
	case VSCodeFamily:
		return []string{"--goto", fmt.Sprintf("%s:%d:%d", file, line, column)}
	case JetBrains:
		return []string{"--line", fmt.Sprint(line), file}
	case VisualStudio:
		return []string{"/Edit", file}
	case Other:
	}
	return []string{file}
}

// SolutionIn is the folder's one solution file, or "" when it holds none or a choice of several.
func SolutionIn(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	var found []string
	for _, e := range entries {
		switch strings.ToLower(filepath.Ext(e.Name())) {
		case ".sln", ".slnx":
			found = append(found, filepath.Join(dir, e.Name()))
		}
	}
	if len(found) != 1 {
		return ""
	}
	return found[0]
}
