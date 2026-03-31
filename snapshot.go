package charmtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

const snapshotDir = "testdata"

// RequireSnapshot compares the current view against a golden file stored
// in testdata/<testname>.golden. If the golden file does not exist, it is
// created automatically (the test passes on first run, then validates on
// subsequent runs).
//
// Set the CHARM_TEST_UPDATE=1 environment variable to overwrite existing
// golden files when the expected output changes intentionally.
//
// Usage:
//
//	func TestMyView(t *testing.T) {
//	    sim := charmtest.New(NewMyModel())
//	    sim.SendKey("j")
//	    charmtest.RequireSnapshot(t, sim)
//	}
func RequireSnapshot(t *testing.T, s *Simulator) {
	t.Helper()
	requireSnapshotNamed(t, s, t.Name())
}

// RequireSnapshotNamed is like RequireSnapshot but uses a custom name for
// the golden file instead of the test name. Useful when a single test
// captures multiple snapshots at different points.
func RequireSnapshotNamed(t *testing.T, s *Simulator, name string) {
	t.Helper()
	requireSnapshotNamed(t, s, name)
}

func requireSnapshotNamed(t *testing.T, s *Simulator, name string) {
	t.Helper()

	// Sanitize the name for use as a filename
	safeName := strings.ReplaceAll(name, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	safeName = strings.ReplaceAll(safeName, " ", "_")

	goldenPath := filepath.Join(snapshotDir, safeName+".golden")
	actual := ansi.Strip(s.View())

	// Check if we should update snapshots
	update := os.Getenv("CHARM_TEST_UPDATE") == "1"

	// If golden file doesn't exist or update is requested, create/overwrite it
	if _, err := os.Stat(goldenPath); os.IsNotExist(err) || update {
		if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
			t.Fatalf("failed to create snapshot directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0o644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		if update {
			t.Logf("snapshot updated: %s", goldenPath)
		} else {
			t.Logf("snapshot created: %s (run again to validate)", goldenPath)
		}
		return
	}

	// Read and compare
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	if actual != string(expected) {
		t.Errorf("snapshot mismatch for %s:\n--- expected (golden) ---\n%s\n--- actual ---\n%s\n\nRun with CHARM_TEST_UPDATE=1 to update the snapshot.",
			goldenPath, string(expected), actual)
	}
}
