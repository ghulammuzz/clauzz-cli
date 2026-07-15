package cli

import (
	"strings"
	"testing"
)

func TestResumeArgs(t *testing.T) {
	got := resumeArgs("abc-123", false)
	if strings.Join(got, " ") != "claude --resume abc-123" {
		t.Errorf("normal mode argv = %v", got)
	}

	got = resumeArgs("abc-123", true)
	if strings.Join(got, " ") != "claude --resume abc-123 --dangerously-skip-permissions" {
		t.Errorf("dasp argv = %v", got)
	}
}
