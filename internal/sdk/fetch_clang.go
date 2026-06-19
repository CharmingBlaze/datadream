package sdk

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallClang downloads and installs a bundled Clang toolchain into sdk/toolchain/clang/.
func InstallClang(root string) error {
	if root == "" {
		root = Root()
	}
	if root == "" {
		return fmt.Errorf("cannot find DataDream root — set DATADREAM_ROOT")
	}

	dest := filepath.Join(root, "sdk", "toolchain", "clang")
	if clang := bundledClang(root); clang != "" {
		fmt.Printf("✓ Clang already installed: %s\n", clang)
		return nil
	}

	asset, url, kind, err := clangAssetForPlatform()
	if err != nil {
		return err
	}

	fmt.Printf("↓ Downloading Clang toolchain: %s\n", asset)
	fmt.Println("  (this is a large download — may take several minutes)")

	tmp, err := os.MkdirTemp("", "datadream-clang-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	archive := filepath.Join(tmp, asset)
	if err := downloadFile(url, archive); err != nil {
		return err
	}

	extractDir := filepath.Join(tmp, "extract")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return err
	}

	var rootDir string
	switch kind {
	case "zip":
		rootDir, err = extractZipArchive(archive, extractDir)
	case "tar.xz":
		rootDir, err = extractTarXzArchive(archive, extractDir)
	default:
		return fmt.Errorf("unsupported archive kind: %s", kind)
	}
	if err != nil {
		return err
	}

	src := filepath.Join(extractDir, rootDir)
	if !fileExists(filepath.Join(src, "bin", clangBinName())) {
		return fmt.Errorf("downloaded toolchain missing bin/%s under %s", clangBinName(), src)
	}

	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("cannot replace existing toolchain dir: %w", err)
	}
	if err := copyDir(src, dest); err != nil {
		return fmt.Errorf("cannot install toolchain: %w", err)
	}

	installed := filepath.Join(dest, "bin", clangBinName())
	fmt.Printf("✓ Clang installed to %s\n", dest)
	fmt.Printf("  Compiler: %s\n", installed)
	return nil
}

func clangBinName() string {
	if runtime.GOOS == "windows" {
		return "clang.exe"
	}
	return "clang"
}

func clangAssetForPlatform() (asset, url, kind string, err error) {
	switch PlatformKey() {
	case "windows-amd64":
		asset = fmt.Sprintf("llvm-mingw-%s-ucrt-x86_64.zip", LLVMMingwVersion)
		url = fmt.Sprintf("https://github.com/mstorsjo/llvm-mingw/releases/download/%s/%s", LLVMMingwVersion, asset)
		return asset, url, "zip", nil
	case "windows-386":
		asset = fmt.Sprintf("llvm-mingw-%s-ucrt-i686.zip", LLVMMingwVersion)
		url = fmt.Sprintf("https://github.com/mstorsjo/llvm-mingw/releases/download/%s/%s", LLVMMingwVersion, asset)
		return asset, url, "zip", nil
	case "windows-arm64":
		asset = fmt.Sprintf("llvm-mingw-%s-ucrt-aarch64.zip", LLVMMingwVersion)
		url = fmt.Sprintf("https://github.com/mstorsjo/llvm-mingw/releases/download/%s/%s", LLVMMingwVersion, asset)
		return asset, url, "zip", nil
	case "linux-amd64":
		asset = fmt.Sprintf("clang+llvm-%s-x86_64-linux-gnu-ubuntu-22.04.tar.xz", LLVMOrgVersion)
		url = fmt.Sprintf("https://github.com/llvm/llvm-project/releases/download/llvmorg-%s/%s", LLVMOrgVersion, asset)
		return asset, url, "tar.xz", nil
	case "linux-arm64":
		asset = fmt.Sprintf("clang+llvm-%s-aarch64-linux-gnu.tar.xz", LLVMOrgVersion)
		url = fmt.Sprintf("https://github.com/llvm/llvm-project/releases/download/llvmorg-%s/%s", LLVMOrgVersion, asset)
		return asset, url, "tar.xz", nil
	case "darwin-arm64":
		asset = fmt.Sprintf("clang+llvm-%s-arm64-apple-darwin22.0.tar.xz", LLVMOrgVersion)
		url = fmt.Sprintf("https://github.com/llvm/llvm-project/releases/download/llvmorg-%s/%s", LLVMOrgVersion, asset)
		return asset, url, "tar.xz", nil
	case "darwin-amd64":
		asset = fmt.Sprintf("clang+llvm-%s-x86_64-apple-darwin22.0.tar.xz", LLVMOrgVersion)
		url = fmt.Sprintf("https://github.com/llvm/llvm-project/releases/download/llvmorg-%s/%s", LLVMOrgVersion, asset)
		return asset, url, "tar.xz", nil
	default:
		return "", "", "", fmt.Errorf("unsupported platform for sdk install clang: %s", PlatformKey())
	}
}

func extractZipArchive(archive, destDir string) (rootName string, err error) {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return "", err
	}
	defer r.Close()

	prefix := zipRootPrefix(r.File)
	rootName = strings.TrimSuffix(prefix, "/")
	if rootName == "" {
		rootName = "."
	}

	for _, f := range r.File {
		target := filepath.Join(destDir, filepath.FromSlash(f.Name))
		if err := extractZipFile(f, target); err != nil {
			return "", err
		}
	}
	return rootName, nil
}

func zipRootPrefix(files []*zip.File) string {
	for _, f := range files {
		name := strings.TrimSuffix(f.Name, "/")
		if name == "" {
			continue
		}
		if i := strings.Index(name, "/"); i > 0 {
			return name[:i+1]
		}
		return name + "/"
	}
	return ""
}

func extractTarXzArchive(archive, destDir string) (rootName string, err error) {
	tarPath, err := exec.LookPath("tar")
	if err != nil {
		return "", fmt.Errorf("tar not found — install tar or download Clang manually into sdk/toolchain/clang/")
	}
	cmd := exec.Command(tarPath, "-xJf", archive, "-C", destDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("extract failed: %w", err)
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return entries[0].Name(), nil
	}
	return ".", nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
