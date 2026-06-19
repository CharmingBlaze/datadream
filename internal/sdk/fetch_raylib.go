package sdk

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// InstallRaylib downloads official raylib 6.0 prebuilt binaries for the current platform.
func InstallRaylib(root string) error {
	if root == "" {
		root = Root()
	}
	if root == "" {
		return fmt.Errorf("cannot find DataDream root — set DATADREAM_ROOT")
	}

	asset, subdir, err := raylibAssetForPlatform()
	if err != nil {
		return err
	}

	url := RaylibReleaseURL + "/" + asset
	fmt.Printf("↓ Downloading raylib %s: %s\n", RaylibVersion, asset)

	tmp, err := os.MkdirTemp("", "datadream-raylib-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	archive := filepath.Join(tmp, asset)
	if err := downloadFile(url, archive); err != nil {
		return err
	}

	destBase := filepath.Join(root, "sdk", "raylib", RaylibVersion)
	includeDest := filepath.Join(destBase, "include")
	libDest := filepath.Join(destBase, "lib", PlatformKey())

	if err := os.MkdirAll(includeDest, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(libDest, 0755); err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(asset, ".zip"):
		if err := extractRaylibZip(archive, includeDest, libDest, subdir); err != nil {
			return err
		}
	case strings.HasSuffix(asset, ".tar.gz"):
		if err := extractRaylibTarGz(archive, includeDest, libDest, subdir); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported archive: %s", asset)
	}

	// Also refresh headers from official tag if missing.
	if !fileExists(filepath.Join(includeDest, "raylib.h")) {
		if err := fetchRaylibHeaders(includeDest); err != nil {
			return err
		}
	}

	fmt.Printf("✓ raylib %s installed to %s\n", RaylibVersion, destBase)
	return nil
}

func raylibAssetForPlatform() (asset, subdir string, err error) {
	switch PlatformKey() {
	case "windows-amd64":
		return "raylib-6.0_win64_mingw-w64.zip", "raylib-6.0_win64_mingw-w64", nil
	case "windows-386":
		return "raylib-6.0_win32_mingw-w64.zip", "raylib-6.0_win32_mingw-w64", nil
	case "linux-amd64":
		return "raylib-6.0_linux_amd64.tar.gz", "raylib-6.0_linux_amd64", nil
	case "linux-arm64":
		return "raylib-6.0_linux_arm64.tar.gz", "raylib-6.0_linux_arm64", nil
	case "darwin-amd64", "darwin-arm64":
		return "raylib-6.0_macos.tar.gz", "raylib-6.0_macos", nil
	default:
		return "", "", fmt.Errorf("unsupported platform for raylib prebuilt: %s", PlatformKey())
	}
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s (%s)", url, resp.Status)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func extractRaylibZip(archive, includeDest, libDest, prefix string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name := f.Name
		if prefix != "" && strings.HasPrefix(name, prefix+"/") {
			name = strings.TrimPrefix(name, prefix+"/")
		}
		if strings.HasPrefix(name, "include/") {
			target := filepath.Join(includeDest, strings.TrimPrefix(name, "include/"))
			if err := extractZipFile(f, target); err != nil {
				return err
			}
		}
		if strings.HasPrefix(name, "lib/") {
			target := filepath.Join(libDest, strings.TrimPrefix(name, "lib/"))
			if err := extractZipFile(f, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractZipFile(f *zip.File, target string) error {
	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0755)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, rc)
	return err
}

func extractRaylibTarGz(archive, includeDest, libDest, prefix string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := hdr.Name
		if prefix != "" && strings.HasPrefix(name, prefix+"/") {
			name = strings.TrimPrefix(name, prefix+"/")
		}
		var target string
		switch {
		case strings.HasPrefix(name, "include/"):
			target = filepath.Join(includeDest, strings.TrimPrefix(name, "include/"))
		case strings.HasPrefix(name, "lib/"):
			target = filepath.Join(libDest, strings.TrimPrefix(name, "lib/"))
		default:
			continue
		}
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	return nil
}

func fetchRaylibHeaders(dest string) error {
	headers := []string{"raylib.h", "raymath.h", "rlgl.h"}
	base := "https://raw.githubusercontent.com/raysan5/raylib/" + RaylibVersion + "/src/"
	for _, h := range headers {
		url := base + h
		path := filepath.Join(dest, h)
		fmt.Printf("↓ Fetching header %s\n", h)
		if err := downloadFile(url, path); err != nil {
			return err
		}
	}
	fmt.Printf("↓ Fetching header raygui.h\n")
	if err := downloadFile("https://raw.githubusercontent.com/raysan5/raygui/master/src/raygui.h", filepath.Join(dest, "raygui.h")); err != nil {
		return err
	}
	return nil
}

// InstallRaylibHeaders fetches official raylib 6.0 headers into the SDK (no libs).
func InstallRaylibHeaders(root string) error {
	if root == "" {
		root = Root()
	}
	dest := filepath.Join(root, "sdk", "raylib", RaylibVersion, "include")
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	return fetchRaylibHeaders(dest)
}
