package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTree writes files (path → content) under root.
func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// fixture is a main module in app/ that links example.com/lib (with a
// subpackage under its own license) everywhere and example.com/webonly only
// on js/wasm, and uses example.com/testonly only in a test. extra files are
// added to the tree.
func fixture(t *testing.T, extra map[string]string) string {
	t.Helper()
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOWORK", "off")
	root := t.TempDir()
	files := map[string]string{
		"app/go.mod": `module example.com/app

go 1.25

require (
	example.com/lib v1.2.3
	example.com/testonly v1.0.0
	example.com/webonly v0.1.0
)

replace (
	example.com/lib => ../lib
	example.com/testonly => ../testonly
	example.com/webonly => ../webonly
)
`,
		"app/main.go": `package main

import (
	_ "example.com/lib"
	_ "example.com/lib/vendored"
)

func main() {}
`,
		"app/main_js.go": `package main

import _ "example.com/webonly"
`,
		"app/main_test.go": `package main

import _ "example.com/testonly"
`,
		"lib/go.mod":               "module example.com/lib\n",
		"lib/lib.go":               "package lib\n",
		"lib/LICENSE":              "lib license\r\n",
		"lib/vendored/vendored.go": "package vendored\n",
		"lib/vendored/COPYING.txt": "vendored license  \n",
		"lib/unused/LICENSE":       "not linked\n",
		"lib/unused/unused.go":     "package unused\n",
		"webonly/go.mod":           "module example.com/webonly\n",
		"webonly/webonly.go":       "package webonly\n",
		"webonly/LICENSE.md":       "webonly license\n",
		"testonly/go.mod":          "module example.com/testonly\n",
		"testonly/testonly.go":     "package testonly\n",
		"testonly/LICENSE":         "testonly license\n",
		"goroot/LICENSE":           "go license\n",
	}
	for k, v := range extra {
		files[k] = v
	}
	writeTree(t, root, files)
	return root
}

var testTargets = []target{{"linux", "amd64", false}, {"js", "wasm", false}}

func TestGenerate(t *testing.T) {
	root := fixture(t, nil)
	entries, err := generate(filepath.Join(root, "app"), testTargets, filepath.Join(root, "goroot"))
	if err != nil {
		t.Fatal(err)
	}
	want := []entry{
		{Module: "example.com/lib", Version: "v1.2.3", License: "lib license\n\nvendored/COPYING.txt:\n\nvendored license"},
		{Module: "example.com/webonly", Version: "v0.1.0", License: "webonly license"},
		{Module: goEntry, License: "go license"},
	}
	if len(entries) != len(want) {
		t.Fatalf("entries = %+v, want %+v", entries, want)
	}
	for i := range want {
		if entries[i] != want[i] {
			t.Errorf("entries[%d] = %+v, want %+v", i, entries[i], want[i])
		}
	}
}

func TestGenerateMissingLicense(t *testing.T) {
	root := fixture(t, map[string]string{
		"app/go.mod": `module example.com/app

go 1.25

require example.com/nolicense v1.0.0

replace example.com/nolicense => ../nolicense
`,
		"app/main.go":            "package main\n\nimport _ \"example.com/nolicense\"\n\nfunc main() {}\n",
		"app/main_js.go":         "package main\n",
		"app/main_test.go":       "package main\n",
		"nolicense/go.mod":       "module example.com/nolicense\n",
		"nolicense/nolicense.go": "package nolicense\n",
		"nolicense/README":       "no license here\n",
	})
	_, err := generate(filepath.Join(root, "app"), testTargets, filepath.Join(root, "goroot"))
	if err == nil || !strings.Contains(err.Error(), "example.com/nolicense") {
		t.Fatalf("err = %v, want one naming example.com/nolicense", err)
	}
}

func TestCheckFile(t *testing.T) {
	entries := []entry{{Module: "example.com/lib", Version: "v1.2.3", License: "lib license"}}
	data, err := encode(entries)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "go_modules.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkFile(path, data); err != nil {
		t.Errorf("unchanged file: %v", err)
	}

	entries[0].Version = "v1.2.4"
	changed, err := encode(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkFile(path, changed); err == nil || !strings.Contains(err.Error(), path) {
		t.Errorf("changed entry: err = %v, want one naming %s", err, path)
	}

	missing := filepath.Join(t.TempDir(), "go_modules.json")
	if err := checkFile(missing, data); err == nil || !strings.Contains(err.Error(), missing) {
		t.Errorf("missing file: err = %v, want one naming %s", err, missing)
	}
}
