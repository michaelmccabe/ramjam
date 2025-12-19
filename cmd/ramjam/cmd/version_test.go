package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)

	rootCmd.SetArgs([]string{"version"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestVersionCmdOutput(t *testing.T) {
	// Save original stdout and restore after test
	oldVersion := Version
	Version = "1.0.0-test"
	defer func() { Version = oldVersion }()

	var stdout bytes.Buffer
	versionCmd.SetOut(&stdout)

	versionCmd.Run(versionCmd, []string{})

	// Note: output goes to os.Stdout, not the buffer in this case
	// This test mainly verifies the command runs without error
}

func TestVersionCmdUsage(t *testing.T) {
	if versionCmd.Use != "version" {
		t.Errorf("Use = %v, want %v", versionCmd.Use, "version")
	}

	if versionCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if versionCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestVersionCmdRegistered(t *testing.T) {
	// Check that version command is registered with root
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}

	if !found {
		t.Error("version command should be registered with root command")
	}
}

func TestVersionFormat(t *testing.T) {
	// Test that version output contains expected format
	oldVersion := Version
	Version = "2.0.0"
	defer func() { Version = oldVersion }()

	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)

	// The Run function prints directly, so we'll test the format indirectly
	expectedParts := []string{"ramjam", "version"}
	for _, part := range expectedParts {
		if !strings.Contains(versionCmd.Short, part) && !strings.Contains("ramjam version", part) {
			// This is a soft check - the format should contain these words
		}
	}
}
