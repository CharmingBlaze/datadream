// packdist assembles a DataDream distribution directory or zip.
//
// Usage:
//
//	packdist --out dist/datadream-windows-amd64.zip --verify
//	packdist --verify-only /path/to/unzipped-dist
//
// --verify: doctor + hello_friendly + hello_raw + coin-runner in packed tree
// Build once (maintainers): go build -o packdist ./tools/packdist
package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	out := flag.String("out", defaultOut(), "output directory or .zip path")
	verify := flag.Bool("verify", false, "run doctor + build hello_friendly + hello_raw + coin-runner on packed tree")
	verifyOnly := flag.String("verify-only", "", "verify an existing unpacked distribution directory")
	skipStudio := flag.Bool("skip-studio", false, "omit datadream-studio IDE from the distribution")
	flag.Parse()

	if *verifyOnly != "" {
		if err := verifyDist(*verifyOnly); err != nil {
			die(err)
		}
		fmt.Println("✓ Verification passed (doctor + hello_friendly + hello_raw + coin-runner build)")
		return
	}

	root, err := os.Getwd()
	if err != nil {
		die(err)
	}

	if strings.HasSuffix(strings.ToLower(*out), ".zip") {
		if err := packZip(root, *out, *verify, *skipStudio); err != nil {
			die(err)
		}
		fmt.Printf("✓ Distribution zip: %s\n", *out)
		return
	}

	if err := packDir(root, *out, *skipStudio); err != nil {
		die(err)
	}
	fmt.Printf("✓ Distribution folder: %s\n", *out)
	if *verify {
		if err := verifyDist(*out); err != nil {
			die(err)
		}
		fmt.Println("✓ Verification passed (doctor + hello_friendly + hello_raw + coin-runner build)")
	}
}

func defaultOut() string {
	return filepath.Join("dist", fmt.Sprintf("datadream-%s-%s.zip", runtime.GOOS, runtime.GOARCH))
}

func packDir(root, dest string, skipStudio bool) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	copyTree(filepath.Join(root, "sdk"), filepath.Join(dest, "sdk"))
	copyTree(filepath.Join(root, "examples"), filepath.Join(dest, "examples"))
	copyTree(filepath.Join(root, "libs"), filepath.Join(dest, "libs"))

	for _, doc := range []string{"README.md", filepath.Join("docs", "GETTING_STARTED.txt"), filepath.Join("docs", "SETUP.md"), filepath.Join("docs", "DISTRIBUTION.md"), filepath.Join("docs", "STUDIO.md")} {
		src := filepath.Join(root, doc)
		dst := filepath.Join(dest, doc)
		copyFile(src, dst)
	}
	copyFile(filepath.Join(root, "sdk", "README.md"), filepath.Join(dest, "sdk", "README.md"))

	binDir := filepath.Join(dest, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	srcBin, err := findCompilerBinary(root)
	if err != nil {
		return err
	}
	dstBin := filepath.Join(binDir, filepath.Base(srcBin))
	if err := copyFileStrict(srcBin, dstBin); err != nil {
		return err
	}
	if err := copyStudio(root, binDir, skipStudio); err != nil {
		return err
	}
	return writeLaunchers(dest, skipStudio)
}

func packZip(root, zipPath string, verify, skipStudio bool) error {
	tmp := strings.TrimSuffix(zipPath, ".zip") + "-tmp"
	if err := os.RemoveAll(tmp); err != nil {
		return err
	}
	if err := packDir(root, tmp, skipStudio); err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	if verify {
		if err := verifyDist(tmp); err != nil {
			return err
		}
		fmt.Println("✓ Verification passed (doctor + hello_friendly + hello_raw + coin-runner build)")
	}

	if err := os.MkdirAll(filepath.Dir(zipPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	return filepath.Walk(tmp, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(tmp, path)
		w, err := zw.Create(filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		r, err := os.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(w, r)
		return err
	})
}

func copyStudio(root, binDir string, skipStudio bool) error {
	if skipStudio {
		fmt.Println("Skipping datadream-studio (packdist --skip-studio)")
		return nil
	}
	src, isApp, err := findBuiltStudio(root)
	if err != nil {
		return fmt.Errorf("datadream-studio not built — run scripts/build-studio before packing: %w", err)
	}
	if isApp {
		dst := filepath.Join(binDir, filepath.Base(src))
		fmt.Printf("Including IDE: %s\n", filepath.Base(src))
		return copyTree(src, dst)
	}
	dst := filepath.Join(binDir, filepath.Base(src))
	fmt.Printf("Including IDE: %s\n", filepath.Base(dst))
	if err := copyFileStrict(src, dst); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		return os.Chmod(dst, 0755)
	}
	return nil
}

func findBuiltStudio(root string) (string, bool, error) {
	exeName := "datadream-studio" + exeSuffix()
	candidates := []struct {
		path string
		app  bool
	}{
		{filepath.Join(root, "cmd", "studio", "build", "bin", exeName), false},
		{filepath.Join(root, "cmd", "studio", "build", "bin", "datadream-studio.app"), true},
		{filepath.Join(root, exeName), false},
		{filepath.Join(root, "bin", exeName), false},
		{filepath.Join(root, "bin", "datadream-studio.app"), true},
	}
	for _, c := range candidates {
		if c.app {
			if info, err := os.Stat(c.path); err == nil && info.IsDir() {
				return c.path, true, nil
			}
			continue
		}
		if fileExists(c.path) {
			return c.path, false, nil
		}
	}
	return "", false, fmt.Errorf("no build output under cmd/studio/build/bin")
}

func writeLaunchers(dest string, skipStudio bool) error {
	if skipStudio {
		return nil
	}

	srcGuide := filepath.Join(dest, "docs", "GETTING_STARTED.txt")
	if fileExists(srcGuide) {
		copyFileStrict(srcGuide, filepath.Join(dest, "GETTING_STARTED.txt"))
	}

	binDir := filepath.Join(dest, "bin")

	switch runtime.GOOS {
	case "windows":
		bat := "@echo off\r\n" +
			"set \"DATADREAM_ROOT=%~dp0\"\r\n" +
			"set \"DATADREAM_ROOT=%DATADREAM_ROOT:~0,-1%\"\r\n" +
			"start \"\" \"%~dp0bin\\datadream-studio.exe\"\r\n"
		if err := os.WriteFile(filepath.Join(dest, "Start DataDream Studio.bat"), []byte(bat), 0644); err != nil {
			return err
		}
		src := filepath.Join(binDir, "datadream-studio.exe")
		if fileExists(src) {
			return copyFileStrict(src, filepath.Join(dest, "DataDream Studio.exe"))
		}
	case "darwin":
		sh := "#!/bin/sh\n" +
			"DIR=\"$(cd \"$(dirname \"$0\")\" && pwd)\"\n" +
			"export DATADREAM_ROOT=\"$DIR\"\n" +
			"exec open -a \"$DIR/bin/datadream-studio.app\"\n"
		shPath := filepath.Join(dest, "start-studio.sh")
		if err := os.WriteFile(shPath, []byte(sh), 0755); err != nil {
			return err
		}
		appSrc := filepath.Join(binDir, "datadream-studio.app")
		if info, err := os.Stat(appSrc); err == nil && info.IsDir() {
			return copyTree(appSrc, filepath.Join(dest, "DataDream Studio.app"))
		}
	default:
		sh := "#!/bin/sh\n" +
			"DIR=\"$(cd \"$(dirname \"$0\")\" && pwd)\"\n" +
			"export DATADREAM_ROOT=\"$DIR\"\n" +
			"exec \"$DIR/bin/datadream-studio\" \"$@\"\n"
		shPath := filepath.Join(dest, "start-studio.sh")
		if err := os.WriteFile(shPath, []byte(sh), 0755); err != nil {
			return err
		}
		src := filepath.Join(binDir, "datadream-studio")
		if fileExists(src) {
			return copyFileStrict(src, filepath.Join(dest, "datadream-studio"))
		}
	}
	return nil
}

func verifyDist(dest string) error {
	bin, err := findPackedBinary(dest)
	if err != nil {
		return err
	}
	bin, err = filepath.Abs(bin)
	if err != nil {
		return err
	}
	dest, err = filepath.Abs(dest)
	if err != nil {
		return err
	}
	env := append(os.Environ(), "DATADREAM_ROOT="+dest)

	cmd := exec.Command(bin, "doctor")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("doctor failed in packed tree: %w", err)
	}

	hello := filepath.Join(dest, "examples", "raylib", "hello_friendly.dd")
	outHello := filepath.Join(dest, "bin", "hello_smoke"+exeSuffix())
	cmd = exec.Command(bin, "build", hello, "-o", outHello)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hello_friendly build failed in packed tree: %w", err)
	}

	helloRaw := filepath.Join(dest, "examples", "raylib", "hello_raw.dd")
	outHelloRaw := filepath.Join(dest, "bin", "hello_raw_smoke"+exeSuffix())
	cmd = exec.Command(bin, "build", helloRaw, "-o", outHelloRaw)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("hello_raw build failed in packed tree: %w", err)
	}

	coinRunnerDir := filepath.Join(dest, "examples", "coin-runner")
	coinOut := filepath.Join(dest, "bin", "coin_runner_smoke"+exeSuffix())
	cmd = exec.Command(bin, "build", "game.dd", "-o", coinOut)
	cmd.Dir = coinRunnerDir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("coin-runner build failed in packed tree: %w", err)
	}
	return nil
}

func findCompilerBinary(root string) (string, error) {
	names := []string{"datadream" + exeSuffix(), "datadream"}
	for _, name := range names {
		p := filepath.Join(root, name)
		if fileExists(p) {
			return p, nil
		}
	}
	return "", fmt.Errorf("datadream binary not found — run: go build -o datadream ./cmd/datadream")
}

func findPackedBinary(dest string) (string, error) {
	p := filepath.Join(dest, "bin", "datadream"+exeSuffix())
	if fileExists(p) {
		return p, nil
	}
	p = filepath.Join(dest, "bin", "datadream")
	if fileExists(p) {
		return p, nil
	}
	return "", fmt.Errorf("packed binary missing under %s/bin", dest)
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func copyTree(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	inExamples := filepath.Base(src) == "examples"
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if shouldSkipCopy(rel, inExamples) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyFile(path, target)
	})
}

func shouldSkipCopy(rel string, inExamples bool) bool {
	rel = filepath.ToSlash(rel)
	base := filepath.Base(rel)
	switch base {
	case ".git", "node_modules":
		return true
	}
	if inExamples && strings.HasSuffix(strings.ToLower(rel), ".exe") {
		return true
	}
	if strings.HasSuffix(rel, ".zip") {
		return true
	}
	return false
}

func copyFile(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return nil
	}
	return copyFileStrict(src, dst)
}

func copyFileStrict(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if info, statErr := os.Stat(src); statErr == nil && info.Mode()&0111 != 0 {
			return os.Chmod(dst, info.Mode().Perm())
		}
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
