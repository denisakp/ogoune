package main

import (
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/v4/disk"

	"github.com/denisakp/ogoune/pkg/agentwire"
)

func part(device, mount, fstype string) disk.PartitionStat {
	return disk.PartitionStat{Device: device, Mountpoint: mount, Fstype: fstype}
}

// A mount table is not a list of disks. This is the case that motivated the
// rule, taken from a real machine: 434 btrfs subvolumes on one device, every one
// reporting the same 188G / 145G / 77%.
func TestRepresentativeMounts_CollapsesSubvolumesOntoOneFilesystem(t *testing.T) {
	parts := []disk.PartitionStat{
		part("/dev/vdb1", "/", "btrfs"),
		part("/dev/vdb1", "/opt/orbstack-guest/data", "btrfs"),
		part("/dev/vdb1", "/mnt/machines/ogoune-agent", "btrfs"),
	}
	for i := 0; i < 400; i++ {
		parts = append(parts, part("/dev/vdb1", fmt.Sprintf("/var/lib/containers/subvol-%03d", i), "btrfs"))
	}
	parts = append(parts, part("mac", "/mnt/mac", "virtiofs"), part("orbstack", "/mnt/data", "overlay"))

	got := representativeMounts(parts)

	want := []string{"/", "/mnt/data", "/mnt/mac"}
	if len(got) != len(want) {
		t.Fatalf("got %d filesystems %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("position %d = %q, want %q", i, got[i], want[i])
		}
	}
}

// "/" is more use to an operator than "/opt/vendor/data/subvol".
func TestRepresentativeMounts_PrefersTheShallowestPath(t *testing.T) {
	// Deepest first, so a first-wins implementation would fail this.
	parts := []disk.PartitionStat{
		part("/dev/sda1", "/var/lib/docker/overlay2/deep/path", "ext4"),
		part("/dev/sda1", "/var/lib", "ext4"),
		part("/dev/sda1", "/", "ext4"),
	}
	got := representativeMounts(parts)
	if len(got) != 1 || got[0] != "/" {
		t.Errorf("got %v, want [/]", got)
	}
}

// Mount order must not decide the answer, and neither must map iteration.
func TestRepresentativeMounts_IsDeterministic(t *testing.T) {
	parts := []disk.PartitionStat{
		part("/dev/sdb", "/data", "ext4"),
		part("/dev/sda", "/", "ext4"),
		part("/dev/sdc", "/backup", "xfs"),
	}
	first := representativeMounts(parts)
	for i := 0; i < 50; i++ {
		again := representativeMounts(parts)
		if len(again) != len(first) {
			t.Fatalf("length changed on iteration %d", i)
		}
		for j := range first {
			if again[j] != first[j] {
				t.Fatalf("order changed on iteration %d: %v vs %v", i, again, first)
			}
		}
	}
}

func TestRepresentativeMounts_SkipsWhatIsNotAFilesystem(t *testing.T) {
	parts := []disk.PartitionStat{
		part("proc", "/proc", "proc"),
		part("tmpfs", "/tmp", "tmpfs"),
		part("udev", "/dev", "devtmpfs"),
		part("", "/nameless", "ext4"),
		part("/dev/sda1", "/", "ext4"),
	}
	got := representativeMounts(parts)
	if len(got) != 1 || got[0] != "/" {
		t.Errorf("got %v, want only [/]", got)
	}
}

// Distinct devices are distinct filesystems, however many there are.
func TestRepresentativeMounts_KeepsGenuinelyDifferentDevices(t *testing.T) {
	parts := []disk.PartitionStat{
		part("/dev/sda1", "/", "ext4"),
		part("/dev/sdb1", "/data", "xfs"),
		part("pool/dataset", "/tank", "zfs"),
	}
	if got := representativeMounts(parts); len(got) != 3 {
		t.Errorf("got %d filesystems %v, want 3", len(got), got)
	}
}

// The cap keeps what somebody will be paged about, not the alphabetical head.
func TestCapAndSort_KeepsTheFullestFilesystems(t *testing.T) {
	var in []agentwire.DiskUsage
	for i := 0; i < maxReportedFilesystems+10; i++ {
		// Alphabetically ascending mounts, usage descending: a naive cap would
		// keep the emptiest.
		in = append(in, agentwire.DiskUsage{
			Mount:   fmt.Sprintf("/mnt/vol-%03d", i),
			UsedPct: float64(100 - i),
		})
	}

	out := capAndSort(in)

	if len(out) != maxReportedFilesystems {
		t.Fatalf("got %d entries, want %d", len(out), maxReportedFilesystems)
	}
	for _, d := range out {
		if d.UsedPct < float64(100-maxReportedFilesystems) {
			t.Errorf("%s at %.0f%% survived the cap while fuller ones were dropped", d.Mount, d.UsedPct)
		}
	}
	// And displayed in a stable order.
	for i := 1; i < len(out); i++ {
		if out[i-1].Mount > out[i].Mount {
			t.Errorf("result is not ordered by mount: %q before %q", out[i-1].Mount, out[i].Mount)
		}
	}
}

func TestCapAndSort_LeavesAShortListAlone(t *testing.T) {
	in := []agentwire.DiskUsage{
		{Mount: "/data", UsedPct: 10},
		{Mount: "/", UsedPct: 90},
	}
	out := capAndSort(in)
	if len(out) != 2 || out[0].Mount != "/" || out[1].Mount != "/data" {
		t.Errorf("got %v, want [/ /data]", out)
	}
}
