package apkg

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"
)

// rezip returns a copy of a zip archive with one file's content replaced.
func rezip(t *testing.T, data []byte, name string, content []byte) []byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, f := range zr.File {
		w, _ := zw.Create(f.Name)
		if f.Name == name {
			w.Write(content)
			continue
		}
		rc, _ := f.Open()
		io.Copy(w, rc)
		rc.Close()
	}
	zw.Close()
	return buf.Bytes()
}
