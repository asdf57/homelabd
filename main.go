package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/asdf57/homelabd/utils"
	"github.com/vishvananda/netlink"
)

var pciBDFPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{4}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}\.[0-7]$`,
)

type MachineReport struct {
	ObservedAt time.Time          `json:"observed_at"`
	Storage    []LSBLKDevice      `json:"storage"`
	System     IdentityInfo       `json:"system"`
	Cpu        CPUInfo            `json:"cpu"`
	Interfaces []NetworkInterface `json:"interfaces"`
}

type NetworkInterface struct {
	Name      string           `json:"name"`
	Type      string           `json:"type"`
	Index     int              `json:"index"`
	MAC       string           `json:"mac"`
	MTU       int              `json:"mtu"`
	State     string           `json:"state"`
	Addresses []NetworkAddress `json:"addresses"`
}

type NetworkAddress struct {
	Address      string `json:"address"`
	PrefixLength int    `json:"prefix_length"`
	Family       string `json:"family"`
}

type OSInfo struct {
	Distro     string `json:"distro"`
	DistroVer  string `json:"distro_version"`
	KernelName string `json:"kernel_name"`
	KernelRel  string `json:"kernel_release"`
	Arch       string `json:"arch"`
}

type LSBLKOutput struct {
	BlockDevices []LSBLKDevice `json:"blockdevices"`
}

// CPUCore mirrors one "processor : N" block from /proc/cpuinfo.
type CPUCore struct {
	Processor      int      `json:"processor"`
	VendorID       string   `json:"vendor_id"`
	CPUFamily      int      `json:"cpu_family"`
	Model          int      `json:"model"`
	ModelName      string   `json:"model_name"`
	Stepping       int      `json:"stepping"`
	Microcode      string   `json:"microcode"`
	MHz            float64  `json:"mhz"`
	CacheSizeKB    int      `json:"cache_size_kb"`
	PhysicalID     int      `json:"physical_id"`
	Siblings       int      `json:"siblings"`
	CoreID         int      `json:"core_id"`
	CPUCores       int      `json:"cpu_cores"`
	ApicID         int      `json:"apicid"`
	InitialApicID  int      `json:"initial_apicid"`
	Flags          []string `json:"flags"`
	Bugs           []string `json:"bugs"`
	Bogomips       float64  `json:"bogomips"`
	ClflushSize    int      `json:"clflush_size"`
	CacheAlignment int      `json:"cache_alignment"`
	AddressSizes   string   `json:"address_sizes"`
	PowerMgmt      []string `json:"power_management"`
}

// CPUInfo is the aggregate view: identity fields that are the same across
// all cores, plus the full per-core slice for anything that varies
// (freq, apicid, core id) and for socket/core/thread derivation.
type CPUInfo struct {
	VendorID      string    `json:"vendor_id"`
	ModelName     string    `json:"model_name"`
	Sockets       int       `json:"sockets"`        // distinct physical_id
	PhysicalCores int       `json:"physical_cores"` // distinct (physical_id, core_id) pairs
	LogicalCPUs   int       `json:"logical_cpus"`   // len(Cores)
	Cores         []CPUCore `json:"cores"`
}

type MemInfo struct {
	TotalKB     uint64 `json:"total_kb"`
	AvailableKB uint64 `json:"available_kb"`
	SwapTotalKB uint64 `json:"swap_total_kb"`
}

type CPUMemInfo struct {
	CPU CPUInfo `json:"cpu"`
	Mem MemInfo `json:"mem"`
}

type LSBLKDevice struct {
	Name       string `json:"name"`
	KernelName string `json:"kname"`
	Path       string `json:"path"`
	Type       string `json:"type"`
	Size       uint64 `json:"size"`
	Rotational bool   `json:"rota"`
	Model      string `json:"model"`
	Serial     string `json:"serial"`
	WWN        string `json:"wwn"`
	Transport  string `json:"tran"`
}

type LSBLKEnrichedDevice struct {
	Name       string `json:"name"`
	KernelName string `json:"kname"`
	Path       string `json:"path"`
	Type       string `json:"type"`
	Size       uint64 `json:"size"`
	Rotational bool   `json:"rota"`
	Model      string `json:"model"`
	Serial     string `json:"serial"`
	WWN        string `json:"wwn"`
	Transport  string `json:"tran"`
}

const dmiPath = "/sys/class/dmi/id/"

type IdentityInfo struct {
	BiosDate        string `json:"bios_date"`
	BiosRelease     string `json:"bios_release"`
	BiosVendor      string `json:"bios_vendor"`
	BIOSVersion     string `json:"bios_version"`
	BoardAssetTag   string `json:"board_asset_tag"`
	BoardName       string `json:"board_name"`
	BoardSerial     string `json:"board_serial"`
	BoardVendor     string `json:"board_vendor"`
	BoardVersion    string `json:"board_version"`
	ChassisAssetTag string `json:"chassis_asset_tag"`
	ChassisSerial   string `json:"chassis_serial"`
	ChassisType     string `json:"chassis_type"`
	ChassisVendor   string `json:"chassis_vendor"`
	ChassisVersion  string `json:"chassis_version"`
	Modalias        string `json:"modalias"`
	ProductFamily   string `json:"product_family"`
	ProductName     string `json:"product_name"`
	ProductSerial   string `json:"product_serial"`
	ProductSKU      string `json:"product_sku"`
	ProductVersion  string `json:"product_version"`
	ProductUUID     string `json:"product_uuid"`
	SysVendor       string `json:"sys_vendor"`
}

func DiscoverSysInfo() (IdentityInfo, error) {
	read := func(f string) string {
		b, _ := os.ReadFile(filepath.Join(dmiPath, f))
		return strings.TrimSpace(string(b))
	}

	return IdentityInfo{
		BiosDate:        read("bios_date"),
		BiosRelease:     read("bios_release"),
		BiosVendor:      read("bios_vendor"),
		BIOSVersion:     read("bios_version"),
		BoardAssetTag:   read("board_asset_tag"),
		BoardName:       read("board_name"),
		BoardSerial:     read("board_serial"),
		BoardVendor:     read("board_vendor"),
		BoardVersion:    read("board_version"),
		ChassisAssetTag: read("chassis_asset_tag"),
		ChassisSerial:   read("chassis_serial"),
		ChassisType:     read("chassis_type"),
		ChassisVendor:   read("chassis_vendor"),
		ChassisVersion:  read("chassis_version"),
		Modalias:        read("modalias"),
		ProductFamily:   read("product_family"),
		ProductName:     read("product_name"),
		ProductSerial:   read("product_serial"),
		ProductSKU:      read("product_sku"),
		ProductVersion:  read("product_version"),
		ProductUUID:     read("product_uuid"),
		SysVendor:       read("sys_vendor"),
	}, nil
}

// func cStrToString(data []byte) string {
// 	res := bytes.IndexByte(data, 0)
// 	if res == -1 {
// 		res =
// 	}
// }

// func DiscoverOSInfo() (OSInfo, error) {
// 	var u unix.Utsname
// 	if err := unix.Uname(&u); err != nil {
// 		return OSInfo{}, err
// 	}

// }

func DiscoverDisks() ([]LSBLKDevice, error) {
	cmd := exec.Command(
		"lsblk",
		"--json",
		"--bytes",
		"--output",
		"NAME,KNAME,PATH,TYPE,SIZE,ROTA,MODEL,SERIAL,WWN,TRAN",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run lsblk: %w", err)
	}

	var result LSBLKOutput
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("decode lsblk output: %w", err)
	}

	var disks []LSBLKDevice
	for _, device := range result.BlockDevices {
		if device.Type == "disk" {
			disks = append(disks, device)
		}
	}

	return disks, nil
}

func pciAddress(blockName string) (string, error) {
	path := filepath.Join("/sys/class/block", blockName, "device")

	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}

	for _, component := range strings.Split(resolved, string(filepath.Separator)) {
		if pciBDFPattern.MatchString(component) {
			return component, nil
		}
	}

	return "", nil
}

func ReadCpuInfo() (*CPUInfo, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return nil, fmt.Errorf("open /proc/cpuinfo: %w", err)
	}
	defer file.Close()

	return ParseCPUInfo(file)
}

type parsedCPUCore struct {
	core          CPUCore
	hasProcessor  bool
	hasPhysicalID bool
	hasCoreID     bool
}

// ParseCPUInfo parses the blank-line-separated records used by
// /proc/cpuinfo. Unknown fields are ignored so the parser remains compatible
// with kernel additions and vendor-specific fields.
func ParseCPUInfo(reader io.Reader) (*CPUInfo, error) {
	scanner := bufio.NewScanner(reader)
	// Flags lines can be large on newer CPUs and inside some virtual machines.
	scanner.Buffer(make([]byte, 4096), 1024*1024)

	var (
		parsed  []parsedCPUCore
		current parsedCPUCore
		lineNo  int
	)

	appendCurrent := func() {
		if current.hasProcessor {
			parsed = append(parsed, current)
		}
		current = parsedCPUCore{}
	}

	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			appendCurrent()
			continue
		}

		key, value, ok := strings.Cut(line, ":")
		if !ok {
			// Ignore non key/value lines. Architectures and kernel versions can
			// add informational lines that do not follow the common format.
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		parseInt := func() (int, error) {
			n, err := strconv.Atoi(value)
			if err != nil {
				return 0, fmt.Errorf("parse /proc/cpuinfo line %d field %q: %w", lineNo, key, err)
			}
			return n, nil
		}
		parseFloat := func() (float64, error) {
			n, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return 0, fmt.Errorf("parse /proc/cpuinfo line %d field %q: %w", lineNo, key, err)
			}
			return n, nil
		}

		var parseErr error
		switch key {
		case "processor":
			// A missing blank line should not merge two processor records.
			if current.hasProcessor {
				parsed = append(parsed, current)
				current = parsedCPUCore{}
			}
			current.core.Processor, parseErr = parseInt()
			current.hasProcessor = parseErr == nil
		case "vendor_id":
			current.core.VendorID = value
		case "cpu family":
			current.core.CPUFamily, parseErr = parseInt()
		case "model":
			current.core.Model, parseErr = parseInt()
		case "model name":
			current.core.ModelName = value
		case "stepping":
			current.core.Stepping, parseErr = parseInt()
		case "microcode":
			current.core.Microcode = value
		case "cpu MHz":
			current.core.MHz, parseErr = parseFloat()
		case "cache size":
			current.core.CacheSizeKB, parseErr = parseCacheSizeKB(value)
		case "physical id":
			current.core.PhysicalID, parseErr = parseInt()
			current.hasPhysicalID = parseErr == nil
		case "siblings":
			current.core.Siblings, parseErr = parseInt()
		case "core id":
			current.core.CoreID, parseErr = parseInt()
			current.hasCoreID = parseErr == nil
		case "cpu cores":
			current.core.CPUCores, parseErr = parseInt()
		case "apicid":
			current.core.ApicID, parseErr = parseInt()
		case "initial apicid":
			current.core.InitialApicID, parseErr = parseInt()
		case "flags":
			current.core.Flags = strings.Fields(value)
		case "bugs":
			current.core.Bugs = strings.Fields(value)
		case "bogomips":
			current.core.Bogomips, parseErr = parseFloat()
		case "clflush size":
			current.core.ClflushSize, parseErr = parseInt()
		case "cache_alignment":
			current.core.CacheAlignment, parseErr = parseInt()
		case "address sizes":
			current.core.AddressSizes = value
		case "power management":
			current.core.PowerMgmt = strings.Fields(value)
		}
		if parseErr != nil {
			return nil, parseErr
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read /proc/cpuinfo: %w", err)
	}
	appendCurrent()
	if len(parsed) == 0 {
		return nil, fmt.Errorf("parse /proc/cpuinfo: no processor records found")
	}

	info := &CPUInfo{
		VendorID:  parsed[0].core.VendorID,
		ModelName: parsed[0].core.ModelName,
		Cores:     make([]CPUCore, 0, len(parsed)),
	}
	sockets := make(map[int]struct{})
	physicalCores := make(map[[2]int]struct{})
	topologyAvailable := true
	for _, entry := range parsed {
		info.Cores = append(info.Cores, entry.core)
		if entry.hasPhysicalID {
			sockets[entry.core.PhysicalID] = struct{}{}
		}
		if entry.hasPhysicalID && entry.hasCoreID {
			physicalCores[[2]int{entry.core.PhysicalID, entry.core.CoreID}] = struct{}{}
		} else {
			topologyAvailable = false
		}
	}

	info.LogicalCPUs = len(info.Cores)
	info.Sockets = len(sockets)
	if info.Sockets == 0 {
		info.Sockets = 1
	}
	if topologyAvailable {
		info.PhysicalCores = len(physicalCores)
	} else if parsed[0].core.CPUCores > 0 {
		info.PhysicalCores = parsed[0].core.CPUCores * info.Sockets
	} else {
		info.PhysicalCores = info.LogicalCPUs
	}

	return info, nil
}

func parseCacheSizeKB(value string) (int, error) {
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty cache size")
	}

	size, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse cache size %q: %w", value, err)
	}
	if len(parts) > 1 {
		switch strings.ToUpper(parts[1]) {
		case "KB", "KIB":
		case "MB", "MIB":
			size *= 1024
		case "B":
			size /= 1024
		default:
			return 0, fmt.Errorf("parse cache size %q: unknown unit %q", value, parts[1])
		}
	}
	return int(size), nil
}

func DiscoverNetworkInfo() ([]NetworkInterface, error) {
	var ifaces []NetworkInterface

	links, err := netlink.LinkList()
	if err != nil {
		return []NetworkInterface{}, err
	}

	for _, link := range links {
		attrs := link.Attrs()
		iface := NetworkInterface{
			Name:  attrs.Name,
			Type:  link.Type(),
			Index: attrs.Index,
			MAC:   attrs.HardwareAddr.String(),
			MTU:   attrs.MTU,
			State: attrs.OperState.String(),
		}

		addrs, err := netlink.AddrList(link, netlink.FAMILY_ALL)
		if err != nil {
			return []NetworkInterface{}, err
		}

		iface.Addresses = make([]NetworkAddress, 0, len(addrs))
		for _, addr := range addrs {
			parsed, ok := networkAddressFromNetlink(addr)
			if ok {
				iface.Addresses = append(iface.Addresses, parsed)
			}
		}

		ifaces = append(ifaces, iface)
	}

	return ifaces, nil
}

func networkAddressFromNetlink(addr netlink.Addr) (NetworkAddress, bool) {
	if addr.IPNet == nil || addr.IP == nil {
		return NetworkAddress{}, false
	}

	prefixLength, bits := addr.Mask.Size()
	if prefixLength < 0 {
		return NetworkAddress{}, false
	}

	family := "ipv6"
	if addr.IP.To4() != nil {
		family = "ipv4"
	} else if bits != net.IPv6len*8 {
		return NetworkAddress{}, false
	}

	return NetworkAddress{
		Address:      addr.IP.String(),
		PrefixLength: prefixLength,
		Family:       family,
	}, true
}

func prettyPrint(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func BuildMachineReport(logger *slog.Logger) (MachineReport, error) {
	diskInfo, err := DiscoverDisks()
	if err != nil {
		logger.Error("failed to discover disks",
			"error", err,
		)
		return MachineReport{}, err
	}

	sysInfo, err := DiscoverSysInfo()
	if err != nil {
		logger.Error("failed to discover system info",
			"error", err,
		)
		return MachineReport{}, err
	}

	cpuInfo, err := ReadCpuInfo()
	if err != nil {
		logger.Error("failed to discover cpu info",
			"error", err,
		)
		return MachineReport{}, err
	}

	ifaceInfo, err := DiscoverNetworkInfo()
	if err != nil {
		logger.Error("failed to discover network info",
			"error", err,
		)
		return MachineReport{}, err
	}

	return MachineReport{
		ObservedAt: time.Now().UTC(),
		Storage:    diskInfo,
		System:     sysInfo,
		Cpu:        *cpuInfo,
		Interfaces: ifaceInfo,
	}, nil

}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	config, err := utils.LoadConfig(logger)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	scanDone := make(chan struct{})

	go func() {
		ticker := time.NewTicker(config.PollingInterval)

		defer close(scanDone)
		defer ticker.Stop()

		// Give the scan one initial run
		_, err := BuildMachineReport(logger)
		if err != nil {
			logger.Error("failed to build machine report", "error", err)
		}

		logger.Info("performed initial scan")

		for {
			select {
			case t := <-ticker.C:

				logger.Info("event time", "tickTime", t.String())
				report, err := BuildMachineReport(logger)
				if err != nil {
					logger.Error("failed to build machine report", "error", err)
				}

				// Send to API server
				jsonBytes, err := json.Marshal(report)
				if err != nil {
					logger.Error("failed to marshal report", "error", err)
					continue
				}

				bodyBuffer := bytes.NewBuffer(jsonBytes)

				resp, err := http.Post(config.APIEndpoint+"/api/v1alpha1/machine-reports", "application/json", bodyBuffer)
				if err != nil {
					logger.Error("failed to submit machine report", "error", err)
					continue
				}
				defer resp.Body.Close()

				respBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading response body: %v\n", err)
					return
				}

				logger.Info("machine report submitted", "status", resp.Status, "body", string(respBytes))

			case <-ctx.Done():
				logger.Info("shutdown requested")
				return
			}
		}
	}()

	<-ctx.Done()
	<-scanDone
}
