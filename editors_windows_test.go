//go:build windows

package editorlocator

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"golang.org/x/sys/windows/registry"
)

func TestMain(m *testing.M) {
	if os.Getenv("EDITORLOCATOR_FAKE_VSWHERE") == "1" {
		_, _ = os.Stdout.WriteString(os.Getenv("EDITORLOCATOR_FAKE_VSWHERE_JSON"))
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestAssociatedEditorsReadsTheProgramsRegisteredForAFileType(t *testing.T) {
	const (
		fileType = ".editorlocatortest"
		progID   = "EditorLocatorTest.File"
	)
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	classes := `Software\Classes\`
	create := func(path string) registry.Key {
		key, _, err := registry.CreateKey(registry.CURRENT_USER, classes+path, registry.SET_VALUE)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			_ = key.Close()
			_ = registry.DeleteKey(registry.CURRENT_USER, classes+path)
		})
		return key
	}
	// Cleanups run last to first, so each key goes before the parent it was created under.
	create(fileType)
	if err := create(fileType+`\OpenWithProgids`).SetStringValue(progID, ""); err != nil {
		t.Fatal(err)
	}
	create(progID)
	create(progID + `\shell`)
	create(progID + `\shell\open`)
	if err := create(progID+`\shell\open\command`).SetStringValue("", `"`+exe+`" "%1"`); err != nil {
		t.Fatal(err)
	}
	found := associatedEditors([]string{fileType})
	if !slices.ContainsFunc(found, func(e Editor) bool { return e.Program == exe }) {
		t.Fatalf("%s not found in %v", exe, found)
	}
}

func TestAssociatedEditorsFindsRegisteredPrograms(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := `"` + exe + `" "%1"`
	tests := []struct {
		name     string
		fileType string
		setup    func(t *testing.T, fileType string) []string
	}{
		{
			name:     "OpenWithProgids",
			fileType: ".elwa",
			setup: func(t *testing.T, fileType string) []string {
				id := "EditorLocator.OpenWithProgids"
				createClassKey(t, registry.CLASSES_ROOT, fileType)
				if err := createClassKey(t, registry.CLASSES_ROOT, fileType+`\OpenWithProgids`).SetStringValue(id, ""); err != nil {
					t.Fatal(err)
				}
				setOpenCommand(t, id, command)
				return []string{exe}
			},
		},
		{
			name:     "two OpenWithProgids",
			fileType: ".elwb",
			setup: func(t *testing.T, fileType string) []string {
				exe2 := cloneExecutable(t, "eltest2.exe")
				id1, id2 := "EditorLocator.ProgA", "EditorLocator.ProgB"
				createClassKey(t, registry.CLASSES_ROOT, fileType)
				progIDs := createClassKey(t, registry.CLASSES_ROOT, fileType+`\OpenWithProgids`)
				if err := progIDs.SetStringValue(id1, ""); err != nil {
					t.Fatal(err)
				}
				if err := progIDs.SetStringValue(id2, ""); err != nil {
					t.Fatal(err)
				}
				setOpenCommand(t, id1, command)
				setOpenCommand(t, id2, `"`+exe2+`" "%1"`)
				return []string{exe, exe2}
			},
		},
		{
			name:     "OpenWithList applications",
			fileType: ".elwc",
			setup: func(t *testing.T, fileType string) []string {
				exe2 := cloneExecutable(t, "elowl2.exe")
				app1, app2 := "elowl1.exe", "elowl2.exe"
				createClassKey(t, registry.CLASSES_ROOT, fileType)
				createClassKey(t, registry.CLASSES_ROOT, fileType+`\OpenWithList`)
				createClassKey(t, registry.CLASSES_ROOT, fileType+`\OpenWithList\`+app1)
				createClassKey(t, registry.CLASSES_ROOT, fileType+`\OpenWithList\`+app2)
				setOpenCommand(t, `Applications\`+app1, command)
				setOpenCommand(t, `Applications\`+app2, `"`+exe2+`" "%1"`)
				return []string{exe, exe2}
			},
		},
		{
			name:     "FileExts OpenWithList skips MRUList",
			fileType: ".elwd",
			setup: func(t *testing.T, fileType string) []string {
				app := "elmru.exe"
				list := fileExtsKey + fileType + `\OpenWithList`
				createClassKey(t, registry.CURRENT_USER, fileExtsKey+fileType)
				key := createClassKey(t, registry.CURRENT_USER, list)
				if err := key.SetStringValue("a", app); err != nil {
					t.Fatal(err)
				}
				if err := key.SetStringValue("MRUList", "a"); err != nil {
					t.Fatal(err)
				}
				setOpenCommand(t, `Applications\`+app, command)
				return []string{exe}
			},
		},
		{
			name:     "REG_EXPAND_SZ open command",
			fileType: ".elwe",
			setup: func(t *testing.T, fileType string) []string {
				id := "EditorLocator.Expand"
				t.Setenv("EDITORLOCATOR_TEST_EXE", exe)
				createClassKey(t, registry.CLASSES_ROOT, fileType)
				if err := createClassKey(t, registry.CLASSES_ROOT, fileType+`\OpenWithProgids`).SetStringValue(id, ""); err != nil {
					t.Fatal(err)
				}
				createClassKey(t, registry.CLASSES_ROOT, id)
				createClassKey(t, registry.CLASSES_ROOT, id+`\shell`)
				createClassKey(t, registry.CLASSES_ROOT, id+`\shell\open`)
				if err := createClassKey(t, registry.CLASSES_ROOT, id+`\shell\open\command`).SetExpandStringValue("", `"%EDITORLOCATOR_TEST_EXE%" "%1"`); err != nil {
					t.Fatal(err)
				}
				return []string{exe}
			},
		},
		{
			name:     "FileExts UserChoice",
			fileType: ".elwf",
			setup: func(t *testing.T, fileType string) []string {
				id := "EditorLocator.UserChoice"
				choice := fileExtsKey + fileType + `\UserChoice`
				createClassKey(t, registry.CURRENT_USER, fileExtsKey+fileType)
				if err := createClassKey(t, registry.CURRENT_USER, choice).SetStringValue("ProgId", id); err != nil {
					t.Fatal(err)
				}
				setOpenCommand(t, id, command)
				return []string{exe}
			},
		},
		{
			name:     "default class ProgID",
			fileType: ".elwg",
			setup: func(t *testing.T, fileType string) []string {
				id := "EditorLocator.DefaultClass"
				if err := createClassKey(t, registry.CLASSES_ROOT, fileType).SetStringValue("", id); err != nil {
					t.Fatal(err)
				}
				setOpenCommand(t, id, command)
				return []string{exe}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			want := tc.setup(t, tc.fileType)
			found := associatedEditors([]string{tc.fileType})
			for _, w := range want {
				if !slices.ContainsFunc(found, func(e Editor) bool { return e.Program == w }) {
					t.Fatalf("%s not found in %v", w, found)
				}
			}
		})
	}
}

func createClassKey(t *testing.T, root registry.Key, path string) registry.Key {
	t.Helper()
	key, _, err := registry.CreateKey(root, path, registry.SET_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = key.Close()
		_ = registry.DeleteKey(root, path)
	})
	return key
}

func setOpenCommand(t *testing.T, class, command string) {
	t.Helper()
	createClassKey(t, registry.CLASSES_ROOT, class)
	createClassKey(t, registry.CLASSES_ROOT, class+`\shell`)
	createClassKey(t, registry.CLASSES_ROOT, class+`\shell\open`)
	if err := createClassKey(t, registry.CLASSES_ROOT, class+`\shell\open\command`).SetStringValue("", command); err != nil {
		t.Fatal(err)
	}
}

func TestAssociatedEditorsNamesTheProgramFromItsFamily(t *testing.T) {
	for _, tc := range []struct {
		fileType, exeName, wantName string
	}{
		{".elna", "rider64.exe", "Rider"},
		{".elnb", "devenv.exe", "Visual Studio"},
	} {
		t.Run(tc.wantName, func(t *testing.T) {
			program := cloneExecutable(t, tc.exeName)
			id := "EditorLocator.Label." + tc.wantName
			createClassKey(t, registry.CLASSES_ROOT, tc.fileType)
			if err := createClassKey(t, registry.CLASSES_ROOT, tc.fileType+`\OpenWithProgids`).SetStringValue(id, ""); err != nil {
				t.Fatal(err)
			}
			setOpenCommand(t, id, `"`+program+`" "%1"`)
			found := associatedEditors([]string{tc.fileType})
			if !slices.ContainsFunc(found, func(e Editor) bool { return e.Program == program && e.Name == tc.wantName }) {
				t.Fatalf("got %v, want %s at %s", found, tc.wantName, program)
			}
		})
	}
}

func TestAssociatedEditorsOmitsTheSolutionPicker(t *testing.T) {
	fileType := ".elomit"
	launcher := cloneExecutable(t, "VSLauncher.exe")
	id := "EditorLocator.VSLauncher"
	createClassKey(t, registry.CLASSES_ROOT, fileType)
	if err := createClassKey(t, registry.CLASSES_ROOT, fileType+`\OpenWithProgids`).SetStringValue(id, ""); err != nil {
		t.Fatal(err)
	}
	setOpenCommand(t, id, `"`+launcher+`" "%1"`)
	found := associatedEditors([]string{fileType})
	if len(found) != 0 {
		t.Fatalf("VSLauncher listed: %v", found)
	}
}

func TestFindAddsEachVisualStudioVswhereLists(t *testing.T) {
	root := t.TempDir()
	t.Setenv("ProgramFiles(x86)", root)
	devenv := cloneExecutable(t, "devenv.exe")
	t.Setenv("EDITORLOCATOR_FAKE_VSWHERE", "1")
	t.Setenv("EDITORLOCATOR_FAKE_VSWHERE_JSON", fmt.Sprintf(
		`[{"displayName":"Visual Studio Community 2022","productPath":%q,"catalog":{"productDisplayVersion":"17.12.3"}}]`,
		devenv,
	))
	src, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(src) //nolint:gosec // G304: src is this test binary from os.Executable.
	if err != nil {
		t.Fatal(err)
	}
	vswhere := filepath.Join(root, "Microsoft Visual Studio", "Installer", "vswhere.exe")
	if err := os.MkdirAll(filepath.Join(root, "Microsoft Visual Studio", "Installer"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(vswhere, data, 0o700); err != nil { //nolint:gosec // G703: vswhere is under t.TempDir().
		t.Fatal(err)
	}
	found := Find()
	if !slices.ContainsFunc(found, func(e Editor) bool {
		return e.Program == devenv && e.Name == "Visual Studio Community 2022 17.12.3" && e.Kind == VisualStudio
	}) {
		t.Fatalf("vswhere install missed: %v", found)
	}
}

func cloneExecutable(t *testing.T, name string) string {
	t.Helper()
	src, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), name)
	data, err := os.ReadFile(src) //nolint:gosec // G304: src is this test binary from os.Executable.
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0o600); err != nil { //nolint:gosec // G703: dst is under t.TempDir().
		t.Fatal(err)
	}
	return dst
}
