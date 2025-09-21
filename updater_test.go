package main

import (
	"testing"
)

func TestVersionComparison(t *testing.T) {
	config := DefaultConfig()
	uc := &UpdateChecker{
		config:         config,
		currentVersion: "1.2.1",
	}

	tests := []struct {
		remote   string
		current  string
		expected bool
		name     string
	}{
		{"1.2.2", "1.2.1", true, "patch version upgrade"},
		{"1.3.0", "1.2.1", true, "minor version upgrade"},
		{"2.0.0", "1.2.1", true, "major version upgrade"},
		{"1.2.1", "1.2.1", false, "same version"},
		{"1.2.0", "1.2.1", false, "downgrade"},
		{"v1.2.2", "v1.2.1", true, "version with v prefix"},
		{"1.2.10", "1.2.9", true, "double digit version"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := uc.isNewerVersion(test.remote, test.current)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != test.expected {
				t.Errorf("Expected %v for %s vs %s, got %v", test.expected, test.remote, test.current, result)
			}
		})
	}
}

func TestGetAssetNameForPlatform(t *testing.T) {
	uc := &UpdateChecker{}

	tests := []struct {
		version  string
		expected string
		name     string
	}{
		{"1.2.1", "nfcuid_linux_amd64_1.2.1.tar.gz", "linux asset name"},
		{"v2.0.0", "nfcuid_linux_amd64_2.0.0.tar.gz", "version with v prefix"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := uc.getAssetNameForPlatform(test.version)
			if result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}
