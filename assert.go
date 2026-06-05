package charmtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

// RequireView asserts that the simulator's current view matches the expected
// string exactly. ANSI escape sequences are stripped before comparison.
func RequireView(t *testing.T, s *Simulator, expected string) {
	t.Helper()
	actual := ansi.Strip(s.View())
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)
	if actual != expected {
		t.Errorf("view mismatch:\n--- expected ---\n%s\n--- actual ---\n%s", expected, actual)
	}
}

// RequireViewContains asserts that the stripped view contains the given
// substring.
func RequireViewContains(t *testing.T, s *Simulator, substr string) {
	t.Helper()
	actual := ansi.Strip(s.View())
	if !strings.Contains(actual, substr) {
		t.Errorf("view does not contain %q:\n%s", substr, actual)
	}
}

// RequireViewNotContains asserts that the stripped view does NOT contain
// the given substring.
func RequireViewNotContains(t *testing.T, s *Simulator, substr string) {
	t.Helper()
	actual := ansi.Strip(s.View())
	if strings.Contains(actual, substr) {
		t.Errorf("view should not contain %q but does:\n%s", substr, actual)
	}
}

// RequireViewLines asserts that the view has exactly the expected number of
// non-empty lines.
func RequireViewLines(t *testing.T, s *Simulator, expected int) {
	t.Helper()
	actual := ansi.Strip(s.View())
	lines := 0
	for _, line := range strings.Split(actual, "\n") {
		if strings.TrimSpace(line) != "" {
			lines++
		}
	}
	if lines != expected {
		t.Errorf("expected %d non-empty lines, got %d", expected, lines)
	}
}

// ViewString returns the current view with ANSI sequences stripped,
// useful for debugging in tests.
func ViewString(s *Simulator) string {
	return ansi.Strip(s.View())
}

// DumpView prints the current view (ANSI-stripped) to the test log for
// debugging purposes.
func DumpView(t *testing.T, s *Simulator) {
	t.Helper()
	t.Logf("Current view:\n%s", ViewString(s))
}

// DumpMessages prints all messages processed by the simulator to the test
// log for debugging.
func DumpMessages(t *testing.T, s *Simulator) {
	t.Helper()
	var b strings.Builder
	for i, msg := range s.Messages() {
		fmt.Fprintf(&b, "  [%d] %T: %v\n", i, msg, msg)
	}
	t.Logf("Messages (%d):\n%s", len(s.Messages()), b.String())
}
