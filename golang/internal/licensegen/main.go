// Command licensegen writes the license texts of the Go modules linked into
// the app to assets/licenses/go_modules.json, which the app's license page
// shows. It is not part of the app.
//
// Run it in golang/ after adding, removing or updating a Go module:
//
//	go run ./internal/licensegen          write the file
//	go run ./internal/licensegen -check   fail if the file is out of date
//
// Both fail when a linked module has no license file.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// target is one platform fluffy builds the Go library for.
type target struct {
	goos, goarch string
	cgo          bool
}

// targets are the platforms fluffy builds for, with the environment it
// builds with (fluffy's build.go: plain `go build -trimpath`, no tags). A
// new fluffy target or build flag has to be added here too.
var targets = []target{
	{"linux", "amd64", true},
	{"windows", "amd64", true},
	{"darwin", "arm64", true},
	{"ios", "arm64", true},
	{"android", "arm64", true},
	{"js", "wasm", false},
}

// entry is one component on the license page.
type entry struct {
	Module  string `json:"module"`
	Version string `json:"version"`
	License string `json:"license"`
}

// goEntry is the name of the entry for the Go standard library, which also
// covers wasm_exec.js. It has no version, so that a Go patch release does
// not change the list.
const goEntry = "Go"

func main() {
	log.SetFlags(0)
	log.SetPrefix("licensegen: ")
	check := flag.Bool("check", false, "fail if the output file is out of date instead of writing it")
	out := flag.String("o", "../assets/licenses/go_modules.json", "output file")
	flag.Parse()

	goroot, err := goEnv(".", "GOROOT")
	if err != nil {
		log.Fatal(err)
	}
	entries, err := generate(".", targets, goroot)
	if err != nil {
		log.Fatal(err)
	}
	data, err := encode(entries)
	if err != nil {
		log.Fatal(err)
	}
	if *check {
		if err := checkFile(*out, data); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, data, 0o644); err != nil {
		log.Fatal(err)
	}
}

// listedPackage is the part of `go list -json` output licensegen reads.
type listedPackage struct {
	ImportPath string
	Dir        string
	Standard   bool
	Module     *listedModule
}

type listedModule struct {
	Path    string
	Version string
	Dir     string
	Main    bool
	Replace *listedModule
}

// module is a linked module and the directories of its linked packages.
type module struct {
	path, version, dir string
	pkgDirs            map[string]bool
}

// generate returns the entries for the modules the main package in dir
// links on any of targets, sorted by module path, and the Go entry from
// goroot's LICENSE.
func generate(dir string, targets []target, goroot string) ([]entry, error) {
	mods := map[string]*module{}
	for _, t := range targets {
		pkgs, err := goList(dir, t)
		if err != nil {
			return nil, err
		}
		for _, p := range pkgs {
			m := p.Module
			if p.Standard || m == nil || m.Main {
				continue
			}
			if r := m.Replace; r != nil {
				// A directory replacement has no version of its own.
				version := m.Version
				if r.Version != "" {
					version = r.Version
				}
				m = &listedModule{Path: m.Path, Version: version, Dir: r.Dir}
			}
			mod := mods[m.Path]
			if mod == nil {
				mod = &module{path: m.Path, version: m.Version, dir: m.Dir, pkgDirs: map[string]bool{}}
				mods[m.Path] = mod
			}
			if p.Dir != "" {
				mod.pkgDirs[p.Dir] = true
			}
		}
	}

	var entries []entry
	for _, mod := range mods {
		if mod.dir == "" {
			d, err := download(dir, mod.path, mod.version)
			if err != nil {
				return nil, err
			}
			mod.dir = d
		}
		text, err := moduleLicense(mod)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry{Module: mod.path, Version: mod.version, License: text})
	}
	slices.SortFunc(entries, func(a, b entry) int { return strings.Compare(a.Module, b.Module) })

	goLicense, err := readGoLicense(goroot)
	if err != nil {
		return nil, err
	}
	return append(entries, entry{Module: goEntry, License: normalize(goLicense)}), nil
}

// goLicensePaths are where Go's LICENSE is looked for, after GOROOT: Arch
// Linux and its derivatives move it out of GOROOT.
var goLicensePaths = []string{"/usr/share/licenses/go/LICENSE"}

func readGoLicense(goroot string) ([]byte, error) {
	paths := append([]string{filepath.Join(goroot, "LICENSE")}, goLicensePaths...)
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("the Go standard library's license is in none of %s", strings.Join(paths, ", "))
}

// goList lists the packages the main package in dir links on t.
func goList(dir string, t target) ([]listedPackage, error) {
	cgo := "0"
	if t.cgo {
		cgo = "1"
	}
	cmd := exec.Command("go", "list", "-deps", "-json", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS="+t.goos, "GOARCH="+t.goarch, "CGO_ENABLED="+cgo, "GOFLAGS=")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list for %s/%s: %w\n%s", t.goos, t.goarch, err, stderr.Bytes())
	}
	var pkgs []listedPackage
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var p listedPackage
		if err := dec.Decode(&p); errors.Is(err, io.EOF) {
			return pkgs, nil
		} else if err != nil {
			return nil, fmt.Errorf("go list for %s/%s: %w", t.goos, t.goarch, err)
		}
		pkgs = append(pkgs, p)
	}
}

// download fetches a module that is not in the module cache yet and
// returns its directory.
func download(dir, path, version string) (string, error) {
	cmd := exec.Command("go", "mod", "download", "-json", path+"@"+version)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go mod download %s@%s: %w", path, version, err)
	}
	var m struct{ Dir string }
	if err := json.Unmarshal(out, &m); err != nil || m.Dir == "" {
		return "", fmt.Errorf("go mod download %s@%s: no directory in %q", path, version, out)
	}
	return m.Dir, nil
}

// moduleLicense returns the license files of mod's root directory and of
// the directories of its linked packages, which may hold code under its own
// license. Files outside the root are introduced by their path.
func moduleLicense(mod *module) (string, error) {
	var parts []string
	root, err := licenseFiles(mod.dir)
	if err != nil {
		return "", err
	}
	for _, f := range root {
		parts = append(parts, f.text)
	}
	dirs := make([]string, 0, len(mod.pkgDirs))
	for d := range mod.pkgDirs {
		if filepath.Clean(d) != filepath.Clean(mod.dir) {
			dirs = append(dirs, d)
		}
	}
	slices.Sort(dirs)
	for _, d := range dirs {
		files, err := licenseFiles(d)
		if err != nil {
			return "", err
		}
		for _, f := range files {
			rel, err := filepath.Rel(mod.dir, filepath.Join(d, f.name))
			if err != nil {
				return "", err
			}
			parts = append(parts, filepath.ToSlash(rel)+":\n\n"+f.text)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("module %s@%s has no license file (LICENSE, LICENCE, COPYING or NOTICE) in %s", mod.path, mod.version, mod.dir)
	}
	return strings.Join(parts, "\n\n"), nil
}

type licenseFile struct{ name, text string }

// licenseFiles returns the license files directly in dir, sorted by name.
func licenseFiles(dir string) ([]licenseFile, error) {
	des, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []licenseFile
	for _, de := range des {
		if de.IsDir() || !isLicenseFile(de.Name()) {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, de.Name()))
		if err != nil {
			return nil, err
		}
		files = append(files, licenseFile{de.Name(), normalize(b)})
	}
	return files, nil
}

func isLicenseFile(name string) bool {
	if strings.HasSuffix(name, ".go") {
		return false
	}
	upper := strings.ToUpper(name)
	for _, prefix := range []string{"LICENSE", "LICENCE", "COPYING", "NOTICE"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

// normalize makes a license text independent of the platform it was read
// on: LF line ends, no trailing blanks.
func normalize(b []byte) string {
	s := strings.ReplaceAll(string(b), "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.Trim(strings.Join(lines, "\n"), "\n")
}

func encode(entries []entry) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// checkFile fails, naming path, when its content is not want.
func checkFile(path string, want []byte) error {
	got, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%s is out of date: run `go run ./internal/licensegen` in golang/ and commit the result", path)
	}
	return nil
}

func goEnv(dir, key string) (string, error) {
	cmd := exec.Command("go", "env", key)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go env %s: %w", key, err)
	}
	return strings.TrimSpace(string(out)), nil
}
