package cli

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Upgrade targets that can be overridden for mirrors and tests.
var (
	upgradeAPIBase      = "https://api.github.com/repos/bzdvdn/speckeep/releases/latest"
	upgradeDownloadBase = "https://github.com/bzdvdn/speckeep/releases/download"
	upgradeRepoOwner    = "bzdvdn"
	upgradeRepoName     = "speckeep"
)

type upgradeStatus struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	IsDev   bool   `json:"is_dev"`
	Update  bool   `json:"update"`
	Hint    string `json:"hint,omitempty"`
}

// fetchLatestTag reads the tag_name of the latest GitHub release.
func fetchLatestTag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upgradeAPIBase, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("reach GitHub releases: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github releases responded %s", resp.Status)
	}
	tag, ok := tagNameFromReleaseJSON(body)
	if !ok {
		return "", fmt.Errorf("release JSON has no tag_name")
	}
	return tag, nil
}

func planUpgrade(current, latest string) upgradeStatus {
	isDev := current == "" || current == "dev"
	if latest == "" {
		return upgradeStatus{Current: current, IsDev: isDev, Update: false, Hint: "latest version unknown"}
	}
	if !isDev && current == latest {
		return upgradeStatus{Current: current, Latest: latest, IsDev: false, Update: false}
	}
	return upgradeStatus{Current: current, Latest: latest, IsDev: isDev, Update: true}
}

func (s upgradeStatus) done() string {
	if s.Update {
		return fmt.Sprintf("update: %s → %s", s.Current, s.Latest)
	}
	if s.Latest == "" {
		return fmt.Sprintf("current: %s (latest unknown)", s.Current)
	}
	return fmt.Sprintf("up to date: %s", s.Current)
}

// resolveBinaryPath is overridable for tests.
var resolveBinaryPath = func() (string, error) { return os.Executable() }

// upgradeInPlace downloads, verifies, and replaces the running speckeep binary.
func upgradeInPlace(ctx context.Context, tag string) (string, error) {
	exe, err := resolveBinaryPath()
	if err != nil {
		return "", fmt.Errorf("locate current binary: %w", err)
	}
	exePath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		exePath = exe
	}
	arch := runtime.GOOS + "_" + runtime.GOARCH
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	asset := fmt.Sprintf("speckeep_%s_%s%s", tag, arch, ext)

	dir, err := os.MkdirTemp("", "speckeep-upgrade-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	archivePath := filepath.Join(dir, asset)
	if err := download(ctx, fmt.Sprintf("%s/%s/%s", upgradeDownloadBase, tag, asset), archivePath); err != nil {
		return "", err
	}
	if err := verifyChecksum(ctx, tag, asset, archivePath); err != nil {
		return "", err
	}
	binPath, err := extractBinary(archivePath, dir)
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		return "", fmt.Errorf("cannot replace a running Windows binary in place — install manually: scoop install speckeep (or overwrite %s after closing)", exePath)
	}
	info, err := os.Stat(exePath)
	if err != nil {
		return "", err
	}
	perm := info.Mode().Perm()
	tmp := exePath + ".speckeep-new"
	if err := copyFileMode(binPath, tmp, 0o755); err != nil {
		return "", fmt.Errorf("stage new binary: %w", err)
	}
	if err := os.Rename(tmp, exePath); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("replace binary at %s (try sudo or a package manager): %w", exePath, err)
	}
	_ = os.Chmod(exePath, perm)
	return exePath, nil
}

func download(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: responded %s", url, resp.Status)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("save %s: %w", dest, err)
	}
	return nil
}

func verifyChecksum(ctx context.Context, tag, asset, archivePath string) error {
	sumsURL := fmt.Sprintf("%s/%s/sha256sum.txt", upgradeDownloadBase, tag)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sumsURL, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil // checksum unavailable — do not block upgrade
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil
	}
	want := ""
	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == asset {
			want = strings.ToLower(strings.TrimSpace(parts[0]))
			break
		}
	}
	if want == "" {
		return nil
	}
	have, err := fileSHA256(archivePath)
	if err != nil {
		return err
	}
	if have != want {
		return fmt.Errorf("checksum mismatch for %s (got %s, want %s) — refusing to replace the binary", asset, have, want)
	}
	return nil
}

func extractBinary(archivePath, destDir string) (string, error) {
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractTarGzBinary(archivePath, destDir)
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZipBinary(archivePath, destDir)
	}
	return "", fmt.Errorf("unsupported archive format: %s", archivePath)
}

func extractTarGzBinary(archivePath, destDir string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != binaryName() {
			continue
		}
		bin := filepath.Join(destDir, binaryName())
		if err := writeReaderFile(bin, tr, 0o755); err != nil {
			return "", err
		}
		return bin, nil
	}
	return "", fmt.Errorf("archive does not contain %s", binaryName())
}

func extractZipBinary(archivePath, destDir string) (string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if filepath.Base(f.Name) != binaryName() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		bin := filepath.Join(destDir, binaryName())
		if err := writeReaderFile(bin, rc, 0o755); err != nil {
			return "", err
		}
		return bin, nil
	}
	return "", fmt.Errorf("archive does not contain %s", binaryName())
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "speckeep.exe"
	}
	return "speckeep"
}

func writeReaderFile(path string, r io.Reader, mode os.FileMode) error {
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, r)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

func copyFileMode(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	return err
}

// packageManagerHint helps users on wrapper-managed installs.
func packageManagerHint() string {
	if _, err := exec.LookPath("brew"); err == nil {
		return "  run: brew upgrade speckeep"
	}
	if _, err := exec.LookPath("sudo"); err == nil {
		return "  run: sudo speckeep self upgrade"
	}
	return "  run: speckeep self upgrade (sudo, if the binary is root-owned)"
}
