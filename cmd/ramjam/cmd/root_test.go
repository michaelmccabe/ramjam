package cmd

import (
	"bytes"
	"testing"
)

func TestRootCmd(t *testing.T) {
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)

	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestRootCmdUsage(t *testing.T) {
	if rootCmd.Use != "ramjam" {
		t.Errorf("Use = %v, want %v", rootCmd.Use, "ramjam")
	}

	if rootCmd.Short == "" {
		t.Error("Short description should not be empty")
	}

	if rootCmd.Long == "" {
		t.Error("Long description should not be empty")
	}
}

func TestRootCmdVerboseFlag(t *testing.T) {
	flag := rootCmd.PersistentFlags().Lookup("verbose")
	if flag == nil {
		t.Fatal("verbose flag not found")
	}

	if flag.Shorthand != "v" {
		t.Errorf("verbose shorthand = %v, want %v", flag.Shorthand, "v")
	}

	if flag.DefValue != "false" {
		t.Errorf("verbose default value = %v, want %v", flag.DefValue, "false")
	}
}

func TestRootCmdHelp(t *testing.T) {
	var stdout bytes.Buffer
	rootCmd.SetOut(&stdout)

	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() with --help error = %v", err)
	}

	output := stdout.String()
	if output == "" {
		t.Error("Help output should not be empty")
	}
}

func TestRootCmdVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}
