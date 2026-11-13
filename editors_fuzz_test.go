package editorlocator

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func FuzzParseVSWhere(f *testing.F) {
	f.Add([]byte(`[
		{"displayName":"Visual Studio Community 2022","productPath":"C:\\VS\\2022\\devenv.exe","catalog":{"productDisplayVersion":"17.12.3"}},
		{"displayName":"Visual Studio Professional 2019","productPath":"C:\\VS\\2019\\devenv.exe","catalog":{"productDisplayVersion":"16.11.40"}},
		{"displayName":"Build Tools","productPath":""}
	]`))
	f.Add([]byte("not json"))
	f.Add([]byte("[]"))
	f.Add([]byte("null"))
	f.Fuzz(func(t *testing.T, data []byte) {
		var installs []struct {
			DisplayName string `json:"displayName"`
			ProductPath string `json:"productPath"`
			Catalog     struct {
				Version string `json:"productDisplayVersion"`
			} `json:"catalog"`
		}
		err := json.Unmarshal(data, &installs)
		got := parseVSWhere(data)
		if err != nil {
			if got != nil {
				t.Fatalf("invalid JSON listed %v", got)
			}
			return
		}
		i := 0
		for _, install := range installs {
			if KindOf(strings.ReplaceAll(install.ProductPath, `\`, "/")) != VisualStudio {
				continue
			}
			if i >= len(got) {
				t.Fatalf("missing %q", install.ProductPath)
			}
			wantName := strings.TrimSpace(install.DisplayName + " " + install.Catalog.Version)
			if got[i].Name != wantName || got[i].Program != install.ProductPath {
				t.Fatalf("got %+v, want %q %q", got[i], wantName, install.ProductPath)
			}
			i++
		}
		if i != len(got) {
			t.Fatalf("extra editors %v", got[i:])
		}
		encoded, err := json.Marshal(gotEditorsJSON(got))
		if err != nil {
			t.Fatal(err)
		}
		again := parseVSWhere(encoded)
		if len(again) != len(got) {
			t.Fatalf("round-trip length %d, want %d", len(again), len(got))
		}
		for j := range got {
			if again[j].Program != got[j].Program || again[j].Name != got[j].Name {
				t.Fatalf("round-trip %v vs %v", again[j], got[j])
			}
		}
	})
}

func gotEditorsJSON(editors []Editor) []struct {
	DisplayName string `json:"displayName"`
	ProductPath string `json:"productPath"`
} {
	rows := make([]struct {
		DisplayName string `json:"displayName"`
		ProductPath string `json:"productPath"`
	}, len(editors))
	for i, e := range editors {
		rows[i].DisplayName = e.Name
		rows[i].ProductPath = e.Program
	}
	return rows
}

func FuzzCommandExe(f *testing.F) {
	f.Add(`"C:\Program Files\Rider\rider64.exe" "%1"`)
	f.Add(`C:\Program Files\Rider\rider64.EXE "%1"`)
	f.Add(`rundll32 shell32,OpenAs`)
	f.Fuzz(func(t *testing.T, command string) {
		got := commandExe(command)
		trimmed := strings.TrimSpace(command)
		if rest, ok := strings.CutPrefix(trimmed, `"`); ok {
			exe, _, _ := strings.Cut(rest, `"`)
			if got != exe {
				t.Fatalf("quoted: got %q, want %q", got, exe)
			}
			return
		}
		at := -1
		for i := 0; i+len(".exe") <= len(trimmed); i++ {
			if strings.EqualFold(trimmed[i:i+len(".exe")], ".exe") {
				at = i
				break
			}
		}
		if at < 0 {
			if got != "" {
				t.Fatalf("no .exe: got %q", got)
			}
			return
		}
		want := trimmed[:at+len(".exe")]
		if got != want {
			t.Fatalf("unquoted: got %q, want %q", got, want)
		}
	})
}

func FuzzKindOf(f *testing.F) {
	f.Add("/usr/bin/code")
	f.Add("/usr/bin/code-insiders")
	f.Add("/usr/local/bin/cursor")
	f.Add("/var/lib/flatpak/exports/bin/com.vscodium.codium")
	f.Add("/opt/rider/bin/rider.sh")
	f.Add("C:/VS/2022/Common7/IDE/devenv.exe")
	f.Add("/usr/bin/monodevelop")
	f.Fuzz(func(t *testing.T, program string) {
		got := KindOf(program)
		if got < Other || got > VisualStudio {
			t.Fatalf("kind %d outside the known families", got)
		}
		if KindOf(filepath.Base(program)) != got {
			t.Fatalf("directory changed kind of %q", program)
		}
		base := strings.ToLower(filepath.Base(program))
		rest, namedCode := strings.CutPrefix(base, "code")
		var want Kind
		switch {
		case strings.HasPrefix(base, "rider"), strings.HasPrefix(base, "com.jetbrains.rider"):
			want = JetBrains
		case strings.HasPrefix(base, "devenv"):
			want = VisualStudio
		case namedCode && (rest == "" || rest[0] == '.' || rest[0] == '-'),
			strings.HasPrefix(base, "codium"), strings.HasPrefix(base, "cursor"),
			strings.HasPrefix(base, "com.visualstudio.code"), strings.HasPrefix(base, "com.vscodium.codium"):
			want = VSCodeFamily
		}
		if got != want {
			t.Fatalf("%q: got %v, want %v", program, got, want)
		}
	})
}
