package cli

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPlanUpgrade(t *testing.T) {
	oldUp := planUpgrade("v1.0.0", "v1.0.0")
	if oldUp.Update {
		t.Fatal("same version should not flip update=true")
	}
	outdated := planUpgrade("v0.8.1", "v1.0.0")
	if !outdated.Update || outdated.Latest != "v1.0.0" {
		t.Fatalf("unexpected outdated status: %+v", outdated)
	}
	dev := planUpgrade("dev", "v1.0.0")
	if !dev.Update || !dev.IsDev {
		t.Fatalf("dev build should be treated as updatable: %+v", dev)
	}
	unknown := planUpgrade("v1.0.0", "")
	if unknown.Update {
		t.Fatalf("unknown latest should not update: %+v", unknown)
	}
}

func TestTagNameFromReleaseJSON(t *testing.T) {
	tag, ok := tagNameFromReleaseJSON([]byte(`{"tag_name":"v2.0.0","name":"two"}`))
	if !ok || tag != "v2.0.0" {
		t.Fatalf("unexpected tag: %q ok=%v", tag, ok)
	}
	if _, ok := tagNameFromReleaseJSON([]byte(`{"foo":1}`)); ok {
		t.Fatal("expected missing tag_name to fail")
	}
}

func TestSelfUpgradeInPlaceRoundtrip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("binary swap is not supported on Windows in-process")
	}

	tmp := t.TempDir()
	exe := filepath.Join(tmp, "speckeep")
	if err := os.WriteFile(exe, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile(old): %v", err)
	}
	oldResolve := resolveBinaryPath
	resolveBinaryPath = func() (string, error) { return exe, nil }
	defer func() { resolveBinaryPath = oldResolve }()

	newContent := []byte("brand-new-binary")
	var tarBuf bytes.Buffer
	gz := gzip.NewWriter(&tarBuf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Name: "speckeep", Mode: 0o755, Size: int64(len(newContent))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := tw.Write(newContent); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tw.Close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gz.Close: %v", err)
	}

	tag := "v9.9.9"
	asset := fmt.Sprintf("speckeep_%s_%s_%s.tar.gz", tag, runtime.GOOS, runtime.GOARCH)
	sum := sha256.Sum256(tarBuf.Bytes())
	sums := fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), asset)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "sha256sum.txt"):
			_, _ = w.Write([]byte(sums))
		case strings.HasSuffix(r.URL.Path, ".tar.gz"):
			_, _ = w.Write(tarBuf.Bytes())
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	oldBase := upgradeDownloadBase
	upgradeDownloadBase = srv.URL + "/releases/download"
	defer func() { upgradeDownloadBase = oldBase }()

	path, err := upgradeInPlace(context.Background(), tag)
	if err != nil {
		t.Fatalf("upgradeInPlace returned error: %v", err)
	}
	if path != exe {
		t.Fatalf("expected replaced path %s, got %s", exe, path)
	}
	got, err := os.ReadFile(exe)
	if err != nil {
		t.Fatalf("read replaced binary: %v", err)
	}
	if !bytes.Equal(got, newContent) {
		t.Fatalf("binary not replaced: got %q want %q", got, newContent)
	}
}

func TestSelfUpgradeRejectsChecksumMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not applicable on Windows")
	}
	tmp := t.TempDir()
	exe := filepath.Join(tmp, "speckeep")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	oldResolve := resolveBinaryPath
	resolveBinaryPath = func() (string, error) { return exe, nil }
	defer func() { resolveBinaryPath = oldResolve }()

	archiveBytes := []byte("not-a-real-tarball")
	tag := "v1.0.0"
	asset := fmt.Sprintf("speckeep_%s_%s_%s.tar.gz", tag, runtime.GOOS, runtime.GOARCH)
	// Wrong checksum on purpose.
	sums := fmt.Sprintf("0000deadbeef  %s\n", asset)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "sha256sum.txt"):
			_, _ = w.Write([]byte(sums))
		case strings.HasSuffix(r.URL.Path, ".tar.gz"):
			_, _ = w.Write(archiveBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	oldBase := upgradeDownloadBase
	upgradeDownloadBase = srv.URL + "/releases/download"
	defer func() { upgradeDownloadBase = oldBase }()

	if _, err := upgradeInPlace(context.Background(), tag); err == nil {
		t.Fatal("expected checksum mismatch to fail the upgrade")
	}
	// Original binary must be untouched.
	if got, _ := os.ReadFile(exe); string(got) != "old" {
		t.Fatalf("binary should be untouched on checksum failure, got %q", got)
	}
}
