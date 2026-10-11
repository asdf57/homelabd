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

func TestMachineReportIncludesReadOnlyPartitionObservations(t *testing.T) {
	mount := "/"
	disk := LSBLKDevice{Name: "sda", Size: 10000000000, ReadOnly: true, Children: []LSBLKDevice{
		{Type: "part", Path: "/dev/sda7", Size: 2000000000, PartitionNumber: 7, StartBytes: 1048576, Filesystem: "ext4", UUID: "root", PartUUID: "partition", MountPoints: []*string{nil, &mount}},
	}}
	report, err := newMachineReport(time.Now(), []LSBLKDevice{disk}, IdentityInfo{}, CPUInfo{}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	observed := report.Spec.Storage[0]
	if !observed.ReadOnly || len(observed.Partitions) != 1 || observed.Partitions[0].Number != 7 || observed.Partitions[0].MountPoints[0] != "/" {
		t.Fatal("partition observations lost")
	}
}
