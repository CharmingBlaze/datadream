package sdk

import (
	"runtime"
	"testing"
)

func TestPlatformKey(t *testing.T) {
	key := PlatformKey()
	if key == "" {
		t.Fatal("empty platform key")
	}
	switch runtime.GOOS {
	case "windows", "linux", "darwin":
	default:
		t.Fatalf("unexpected GOOS %q", runtime.GOOS)
	}
}

func TestRaylibLinkLibsNonEmpty(t *testing.T) {
	flags := RaylibLinkLibs()
	if len(flags) == 0 {
		t.Fatal("expected link flags")
	}
	switch runtime.GOOS {
	case "windows":
		found := false
		for _, f := range flags {
			if f == "-lwinmm" || f == "-lopengl32" {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected Windows system libs in %v", flags)
		}
	case "darwin":
		found := false
		for _, f := range flags {
			if f == "-framework" {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected framework flags on darwin, got %v", flags)
		}
	case "linux":
		found := false
		for _, f := range flags {
			if f == "-lX11" || f == "-lraylib" || f == "-lGL" {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected Linux link libs, got %v", flags)
		}
	}
}

func TestDoctorClangFlavorOKWhenCompilerWorks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows flavor checks covered by toolchain tests")
	}
	st := Doctor()
	if st.ClangOK && !st.ClangFlavorOK {
		t.Fatalf("ClangFlavorOK should follow ClangOK on %s", runtime.GOOS)
	}
}

func TestRaylibAssetForCurrentPlatform(t *testing.T) {
	asset, _, err := raylibAssetForPlatform()
	if err != nil {
		t.Fatal(err)
	}
	if asset == "" {
		t.Fatal("empty raylib asset name")
	}
}

func TestClangAssetForCurrentPlatform(t *testing.T) {
	asset, url, kind, err := clangAssetForPlatform()
	if err != nil {
		t.Fatal(err)
	}
	if asset == "" || url == "" || kind == "" {
		t.Fatalf("incomplete clang asset info: %q %q %q", asset, url, kind)
	}
}
