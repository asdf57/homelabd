package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asdf57/homelabd/utils"
	"github.com/vishvananda/netlink"
)

func TestParseCPUInfo(t *testing.T) {
	const input = `processor       : 0
vendor_id       : AuthenticAMD
cpu family      : 25
model           : 33
model name      : AMD Ryzen 7 5800X 8-Core Processor
stepping        : 2
microcode       : 0xa201205
cpu MHz         : 1754.308
cache size      : 512 KB
physical id     : 0
siblings        : 16
core id         : 0
cpu cores       : 8
apicid          : 0
initial apicid  : 0
flags           : fpu vme avx2
bugs            : spectre_v1 spectre_v2
bogomips        : 7600.39
clflush size    : 64
cache_alignment : 64
address sizes   : 48 bits physical, 48 bits virtual
power management: ts ttp tm hwpstate cpb

processor       : 1
vendor_id       : AuthenticAMD
model name      : AMD Ryzen 7 5800X 8-Core Processor
cpu MHz         : 3864.117
cache size      : 512 KB
physical id     : 0
core id         : 1
cpu cores       : 8

processor       : 8
vendor_id       : AuthenticAMD
model name      : AMD Ryzen 7 5800X 8-Core Processor
cpu MHz         : 3858.868
cache size      : 512 KB
physical id     : 0
core id         : 0
cpu cores       : 8
`

	info, err := ParseCPUInfo(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseCPUInfo() error = %v", err)
	}

	if info.VendorID != "AuthenticAMD" {
		t.Errorf("VendorID = %q, want AuthenticAMD", info.VendorID)
	}
	if info.ModelName != "AMD Ryzen 7 5800X 8-Core Processor" {
		t.Errorf("ModelName = %q", info.ModelName)
	}
	if info.Sockets != 1 || info.PhysicalCores != 2 || info.LogicalCPUs != 3 {
		t.Errorf("topology = %d sockets, %d physical cores, %d logical CPUs", info.Sockets, info.PhysicalCores, info.LogicalCPUs)
	}
	if len(info.Cores) != 3 {
		t.Fatalf("len(Cores) = %d, want 3", len(info.Cores))
	}

	first := info.Cores[0]
	if first.Processor != 0 || first.CPUFamily != 25 || first.Model != 33 || first.Stepping != 2 {
		t.Errorf("first core identity fields were not parsed: %+v", first)
	}
	if first.MHz != 1754.308 || first.CacheSizeKB != 512 || first.Bogomips != 7600.39 {
		t.Errorf("first core numeric fields were not parsed: %+v", first)
	}
	if got := strings.Join(first.Flags, " "); got != "fpu vme avx2" {
		t.Errorf("Flags = %q", got)
	}
	if got := strings.Join(first.PowerMgmt, " "); got != "ts ttp tm hwpstate cpb" {
		t.Errorf("PowerMgmt = %q", got)
	}
}

func TestParseCPUInfoWithoutBlankLines(t *testing.T) {
	const input = `processor : 0
physical id : 0
core id : 0
processor : 1
physical id : 0
core id : 0`

	info, err := ParseCPUInfo(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseCPUInfo() error = %v", err)
	}
	if info.LogicalCPUs != 2 || info.PhysicalCores != 1 {
		t.Errorf("topology = %d physical cores, %d logical CPUs", info.PhysicalCores, info.LogicalCPUs)
	}
}

func TestParseCPUInfoDerivesSMTTopology(t *testing.T) {
	var input strings.Builder
	for processor := 0; processor < 16; processor++ {
		fmt.Fprintf(&input, "processor : %d\nphysical id : 0\ncore id : %d\ncpu cores : 8\n\n", processor, processor%8)
	}

	info, err := ParseCPUInfo(strings.NewReader(input.String()))
	if err != nil {
		t.Fatalf("ParseCPUInfo() error = %v", err)
	}
	if info.Sockets != 1 || info.PhysicalCores != 8 || info.LogicalCPUs != 16 {
		t.Errorf("topology = %d sockets, %d physical cores, %d logical CPUs", info.Sockets, info.PhysicalCores, info.LogicalCPUs)
	}
}

func TestParseCPUInfoRejectsInvalidNumbers(t *testing.T) {
	_, err := ParseCPUInfo(strings.NewReader("processor : nope\n"))
	if err == nil || !strings.Contains(err.Error(), `line 1 field "processor"`) {
		t.Fatalf("ParseCPUInfo() error = %v, want field and line context", err)
	}
}

func TestParseCPUInfoRequiresProcessorRecord(t *testing.T) {
	_, err := ParseCPUInfo(strings.NewReader("vendor_id : AuthenticAMD\n"))
	if err == nil {
		t.Fatal("ParseCPUInfo() error = nil, want no-records error")
	}
}

func TestParseCacheSizeKB(t *testing.T) {
	for _, test := range []struct {
		input string
		want  int
	}{
		{input: "512 KB", want: 512},
		{input: "32 MB", want: 32768},
		{input: "1048576 B", want: 1024},
	} {
		got, err := parseCacheSizeKB(test.input)
		if err != nil {
			t.Errorf("parseCacheSizeKB(%q) error = %v", test.input, err)
		} else if got != test.want {
			t.Errorf("parseCacheSizeKB(%q) = %d, want %d", test.input, got, test.want)
		}
	}
}

func TestNetworkAddressFromNetlink(t *testing.T) {
	for _, test := range []struct {
		name      string
		cidr      string
		want      NetworkAddress
		wantValid bool
	}{
		{
			name: "IPv4",
			cidr: "192.0.2.10/24",
			want: NetworkAddress{
				Address:      "192.0.2.10",
				PrefixLength: 24,
				Family:       "ipv4",
			},
			wantValid: true,
		},
		{
			name: "IPv6",
			cidr: "2001:db8::10/64",
			want: NetworkAddress{
				Address:      "2001:db8::10",
				PrefixLength: 64,
				Family:       "ipv6",
			},
			wantValid: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ip, network, err := net.ParseCIDR(test.cidr)
			if err != nil {
				t.Fatal(err)
			}
			network.IP = ip

			got, valid := networkAddressFromNetlink(netlink.Addr{IPNet: network})
			if valid != test.wantValid {
				t.Fatalf("valid = %v, want %v", valid, test.wantValid)
			}
			if got != test.want {
				t.Errorf("networkAddressFromNetlink() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestNetworkInterfaceJSONUsesAPIPrimitives(t *testing.T) {
	iface := NetworkInterface{
		Name:  "eth0",
		Type:  "device",
		Index: 2,
		MAC:   "02:42:c0:00:02:0a",
		MTU:   1500,
		State: "up",
		Addresses: []NetworkAddress{
			{
				Address:      "192.0.2.10",
				PrefixLength: 24,
				Family:       "ipv4",
			},
		},
	}

	data, err := json.Marshal(iface)
	if err != nil {
		t.Fatal(err)
	}

	got := string(data)
	for _, want := range []string{
		`"mac":"02:42:c0:00:02:0a"`,
		`"state":"up"`,
		`"addresses":[{"address":"192.0.2.10","prefix_length":24,"family":"ipv4"}]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("JSON %s does not contain %s", got, want)
		}
	}
}

func TestInstallSSHKeys(t *testing.T) {
	directory := t.TempDir()
	stalePath := filepath.Join(directory, "stale")
	if err := os.WriteFile(stalePath, []byte("stale key\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	server := &utils.Server{Status: &utils.ServerStatus{SSH: &utils.ServerSSHStatus{
		AuthorizedKeys: []utils.ServerSSHAuthorizedKeyStatus{
			{LoginUser: "ansible", PublicKey: "ssh-ed25519 first first@example"},
			{LoginUser: "ansible", PublicKey: "ssh-ed25519 second second@example\n"},
			{LoginUser: "operator", PublicKey: "ssh-ed25519 third third@example"},
		},
	}}}

	if err := installSSHKeys(testLogger(), server, directory); err != nil {
		t.Fatalf("installSSHKeys() error = %v", err)
	}

	want := "ssh-ed25519 first first@example\nssh-ed25519 second second@example\n"
	assertFileContentAndMode(t, filepath.Join(directory, "ansible"), want, 0o600)
	assertFileContentAndMode(t, filepath.Join(directory, "operator"), "ssh-ed25519 third third@example\n", 0o600)
	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Errorf("stale key file still exists; os.Stat() error = %v", err)
	}

	if err := installSSHKeys(testLogger(), &utils.Server{}, directory); err != nil {
		t.Fatalf("installSSHKeys() with no status error = %v", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("key directory contains %d entries after empty status, want 0", len(entries))
	}
}

func TestInstallSSHKeysRejectsUnsafeLoginUser(t *testing.T) {
	server := &utils.Server{Status: &utils.ServerStatus{SSH: &utils.ServerSSHStatus{
		AuthorizedKeys: []utils.ServerSSHAuthorizedKeyStatus{
			{LoginUser: "../root", PublicKey: "ssh-ed25519 bad"},
		},
	}}}

	err := installSSHKeys(testLogger(), server, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "invalid SSH login user") {
		t.Fatalf("installSSHKeys() error = %v, want invalid-login-user error", err)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertFileContentAndMode(t *testing.T, path, want string, wantMode os.FileMode) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != want {
		t.Errorf("%s content = %q, want %q", path, content, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != wantMode {
		t.Errorf("%s mode = %o, want %o", path, info.Mode().Perm(), wantMode)
	}
}
