package editorlocator

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func placeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestFindLabelsEveryProgramAndListsItOnce(t *testing.T) {
	found := Find()
	seen := map[string]bool{}
	for _, e := range found {
		if e.Name == "" {
			t.Errorf("unnamed editor at %s", e.Program)
		}
		if e.Kind != KindOf(e.Program) {
			t.Errorf("%s: kind %v, want %v", e.Program, e.Kind, KindOf(e.Program))
		}
		if seen[e.Program] {
			t.Errorf("%s listed twice", e.Program)
		}
		seen[e.Program] = true
	}
}

func TestEditorsFindEachInstalledVersionOnce(t *testing.T) {
	root := t.TempDir()
	local := filepath.Join(root, "Local")
	programs := filepath.Join(root, "Program Files")
	older := filepath.Join(programs, "JetBrains", "JetBrains Rider 2024.3", "bin", "rider64.exe")
	newer := filepath.Join(programs, "JetBrains", "JetBrains Rider 2025.1", "bin", "rider64.exe")
	code := filepath.Join(local, "Programs", "Microsoft VS Code", "bin", "code.cmd")
	for _, path := range []string{older, newer, code} {
		placeFile(t, path, "")
	}
	placeFile(t, filepath.Join(programs, "JetBrains", "JetBrains Rider 2025.1", "product-info.json"), `{"version":"2025.1.2"}`)
	env := map[string]string{"LOCALAPPDATA": local, "ProgramFiles": programs}
	host := editorHost{
		windows: true,
		getenv:  func(name string) string { return env[name] },
		lookPath: func(bin string) (string, error) {
			if bin == "code" {
				return code, nil
			}
			return "", errors.New("not on PATH")
		},
	}
	got := host.find()
	want := []Editor{
		{Name: "VS Code", Program: code, Kind: VSCodeFamily},
		{Name: "Rider 2025.1.2", Program: newer, Kind: JetBrains},
		{Name: "Rider", Program: older, Kind: JetBrains},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseVSWhereListsEveryInstall(t *testing.T) {
	got := parseVSWhere([]byte(`[
		{"displayName":"Visual Studio Community 2022","productPath":"C:/VS/2022/devenv.exe","catalog":{"productDisplayVersion":"17.12.3"}},
		{"displayName":"Visual Studio Professional 2019","productPath":"C:/VS/2019/devenv.exe","catalog":{"productDisplayVersion":"16.11.40"}},
		{"displayName":"Build Tools","productPath":""},
		{"displayName":"Visual Studio Build Tools 2019","productPath":"C:/BuildTools/Common7/Tools/LaunchDevCmd.bat","catalog":{"productDisplayVersion":"16.8.4"}}
	]`))
	want := []Editor{
		{Name: "Visual Studio Community 2022 17.12.3", Program: "C:/VS/2022/devenv.exe"},
		{Name: "Visual Studio Professional 2019 16.11.40", Program: "C:/VS/2019/devenv.exe"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v", got)
	}
	if parseVSWhere([]byte("not json")) != nil {
		t.Fatal("broken output should list nothing")
	}
}

func TestCommandExe(t *testing.T) {
	for command, want := range map[string]string{
		`"C:\Program Files\Rider\rider64.exe" "%1"`: `C:\Program Files\Rider\rider64.exe`,
		`C:\Program Files\Rider\rider64.EXE "%1"`:   `C:\Program Files\Rider\rider64.EXE`,
		`rundll32 shell32,OpenAs`:                   "",
	} {
		if got := commandExe(command); got != want {
			t.Errorf("%s: got %q", command, got)
		}
	}
}

func TestKindOfReadsTheProgramNameAlone(t *testing.T) {
	for program, want := range map[string]Kind{
		"/usr/bin/code":          VSCodeFamily,
		"/usr/bin/code-insiders": VSCodeFamily,
		"/usr/bin/codelite":      Other,
		"/usr/local/bin/cursor":  VSCodeFamily,
		"/var/lib/flatpak/exports/bin/com.vscodium.codium": VSCodeFamily,
		"/opt/rider/bin/rider.sh":                          JetBrains,
		"C:/VS/2022/Common7/IDE/devenv.exe":                VisualStudio,
		"/usr/bin/monodevelop":                             Other,
		"/usr/bin/provider.sh":                             Other,
		"/var/lib/flatpak/exports/bin/com.jetbrains.Rider": JetBrains,
	} {
		if got := KindOf(program); got != want {
			t.Errorf("%s: got %v, want %v", program, got, want)
		}
	}
}

func TestArguments(t *testing.T) {
	const (
		dir = "/mods/ShopDrop"
		sln = "/mods/ShopDrop/ShopDrop.slnx"
		src = "/mods/ShopDrop/Plugin.cs"
	)
	for program, want := range map[string]struct{ project, file []string }{
		"/usr/bin/code":      {[]string{dir}, []string{"--goto", src + ":7:2"}},
		"/opt/r/rider.sh":    {[]string{sln}, []string{"--line", "7", src}},
		"/vs/devenv.exe":     {[]string{sln}, []string{"/Edit", src}},
		"/usr/bin/monodevel": {[]string{sln}, []string{src}},
	} {
		if got := ProjectArgs(program, dir, sln); !slices.Equal(got, want.project) {
			t.Errorf("%s project: %v", program, got)
		}
		if got := FileArgs(program, src, 7, 2); !slices.Equal(got, want.file) {
			t.Errorf("%s file: %v", program, got)
		}
	}
	if got := ProjectArgs("/opt/r/rider.sh", dir, ""); !slices.Equal(got, []string{dir}) {
		t.Errorf("no solution: %v", got)
	}
}

func TestSolutionInNeedsExactlyOne(t *testing.T) {
	dir := t.TempDir()
	if got := SolutionIn(dir); got != "" {
		t.Fatalf("none: %q", got)
	}
	placeFile(t, filepath.Join(dir, "A.slnx"), "")
	if got := SolutionIn(dir); filepath.Base(got) != "A.slnx" {
		t.Fatalf("one: %q", got)
	}
	placeFile(t, filepath.Join(dir, "B.sln"), "")
	if got := SolutionIn(dir); got != "" {
		t.Fatalf("two: %q", got)
	}
}

func TestSolutionInIgnoresDirectories(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "Old.sln"), 0o750); err != nil {
		t.Fatal(err)
	}
	placeFile(t, filepath.Join(dir, "App.sln"), "")
	if got := SolutionIn(dir); filepath.Base(got) != "App.sln" {
		t.Fatalf("got %q", got)
	}
}

func TestSolutionInEmptyDirIsNotCwd(t *testing.T) {
	dir := t.TempDir()
	placeFile(t, filepath.Join(dir, "App.sln"), "")
	t.Chdir(dir)
	if got := SolutionIn(""); got != "" {
		t.Fatalf("empty dir: %q", got)
	}
	bracket := filepath.Join(t.TempDir(), "proj[1]")
	placeFile(t, filepath.Join(bracket, "App.sln"), "")
	if got := SolutionIn(bracket); filepath.Base(got) != "App.sln" {
		t.Fatalf("bracketed: %q", got)
	}
}

func TestFindOrdersVersionsNumerically(t *testing.T) {
	programs := t.TempDir()
	var want []Editor
	for _, v := range []string{"2024.9", "2024.10", "2024.2"} {
		placeFile(t, filepath.Join(programs, "JetBrains", "JetBrains Rider "+v, "bin", "rider64.exe"), "")
	}
	for _, v := range []string{"2024.10", "2024.9", "2024.2"} {
		want = append(want, Editor{Name: "Rider", Program: filepath.Join(programs, "JetBrains", "JetBrains Rider "+v, "bin", "rider64.exe"), Kind: JetBrains})
	}
	host := editorHost{
		windows:  true,
		getenv:   func(name string) string { return map[string]string{"ProgramFiles": programs}[name] },
		lookPath: func(string) (string, error) { return "", errors.New("not on PATH") },
	}
	if got := host.find(); !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
