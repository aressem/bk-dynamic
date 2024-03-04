package main

import (
	"os"
	"testing"
)

func TestIsPullRequest(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{"BUILDKITE_PULL_REQUEST true should return true", "true", true},
		{"BUILDKITE_PULL_REQUEST false should return false", "false", false},
		{"BUILDKITE_PULL_REQUEST empty should return true", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BUILDKITE_PULL_REQUEST", tt.args)
			if got := isPullRequest(); got != tt.want {
				t.Errorf("isPullrequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMain_test(t *testing.T) {
	os.Setenv("VESPA_VERSION", "9.0.0")
	os.Setenv("BUILDKITE_PULL_REQUEST", "false")
	main()
}
