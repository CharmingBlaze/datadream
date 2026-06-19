package ide

import (
	"os"
	"path/filepath"

	"datadream/internal/sdk"
)

// DistributionRoot returns the DataDream install folder (with sdk/manifest.json).
// Checks DATADREAM_ROOT, then auto-detection from the running binary.
func DistributionRoot() string {
	if v := os.Getenv("DATADREAM_ROOT"); v != "" {
		if hasDistribution(v) {
			return filepath.Clean(v)
		}
	}
	if r := sdk.Root(); r != "" {
		return r
	}
	return ""
}

// EnsureDistributionRoot sets DATADREAM_ROOT and returns the install folder.
// Falls back to preferred when it contains the SDK, otherwise uses DistributionRoot().
func EnsureDistributionRoot(preferred string) string {
	if preferred != "" && hasDistribution(preferred) {
		_ = os.Setenv("DATADREAM_ROOT", filepath.Clean(preferred))
		return filepath.Clean(preferred)
	}
	if r := DistributionRoot(); r != "" {
		_ = os.Setenv("DATADREAM_ROOT", r)
		return r
	}
	if preferred != "" {
		_ = os.Setenv("DATADREAM_ROOT", filepath.Clean(preferred))
		return filepath.Clean(preferred)
	}
	return ""
}

func hasDistribution(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "sdk", "manifest.json"))
	return err == nil
}
