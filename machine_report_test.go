package main

import (
	"regexp"
	"testing"
	"time"
)

func TestMachineReportsHaveIndependentNames(t *testing.T) {
	observedAt := time.Date(2026, 10, 9, 0, 0, 0, 0, time.UTC)
	validName := regexp.MustCompile(`^report-[a-z2-7]{26}$`)
	names := make(map[string]bool)
	for range 128 {
		report, err := newMachineReport(observedAt, nil, IdentityInfo{}, CPUInfo{}, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		name := report.Metadata.Name
		if !validName.MatchString(name) || len(name) > 63 {
			t.Fatalf("invalid report name %q", name)
		}
		if names[name] {
			t.Fatalf("repeated report name %q", name)
		}
		names[name] = true
		if !report.Spec.ObservedAt.Equal(observedAt) {
			t.Fatal("report naming changed observation time")
		}
	}
}
