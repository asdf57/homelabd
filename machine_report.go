package main

import (
	"fmt"
	"math"
	"time"

	"github.com/asdf57/homelabd/utils"
)

func newMachineReport(
	observedAt time.Time,
	storage []LSBLKDevice,
	system IdentityInfo,
	cpu CPUInfo,
	interfaces []NetworkInterface,
	lldp []LLDPInterfaceGroup,
) (utils.MachineReport, error) {
	apiStorage := make([]utils.MachineReportStorageDevice, len(storage))
	for i, device := range storage {
		if device.Size > math.MaxInt64 {
			return utils.MachineReport{}, fmt.Errorf("disk %q size %d exceeds Stigmergy's int64 limit", device.Name, device.Size)
		}
		apiStorage[i] = utils.MachineReportStorageDevice{
			KName:  device.KernelName,
			Model:  device.Model,
			Name:   device.Name,
			Path:   device.Path,
			Rota:   device.Rotational,
			Serial: device.Serial,
			Size:   int64(device.Size),
			Tran:   device.Transport,
			Type:   device.Type,
			WWN:    device.WWN,
		}
	}

	apiCores := make([]utils.MachineReportCPUCore, len(cpu.Cores))
	for i, core := range cpu.Cores {
		apiCores[i] = utils.MachineReportCPUCore{
			AddressSizes:    core.AddressSizes,
			APICID:          core.ApicID,
			Bogomips:        float32(core.Bogomips),
			Bugs:            core.Bugs,
			CacheAlignment:  core.CacheAlignment,
			CacheSizeKB:     core.CacheSizeKB,
			ClflushSize:     core.ClflushSize,
			CoreID:          core.CoreID,
			CPUCores:        core.CPUCores,
			CPUFamily:       core.CPUFamily,
			Flags:           core.Flags,
			InitialAPICID:   core.InitialApicID,
			MHz:             float32(core.MHz),
			Microcode:       core.Microcode,
			Model:           core.Model,
			ModelName:       core.ModelName,
			PhysicalID:      core.PhysicalID,
			PowerManagement: core.PowerMgmt,
			Processor:       core.Processor,
			Siblings:        core.Siblings,
			Stepping:        core.Stepping,
			VendorID:        core.VendorID,
		}
	}

	apiInterfaces := make([]utils.MachineReportNetworkInterface, len(interfaces))
	for i, iface := range interfaces {
		addresses := make([]utils.MachineReportNetworkAddress, len(iface.Addresses))
		for j, address := range iface.Addresses {
			addresses[j] = utils.MachineReportNetworkAddress{
				Address:      address.Address,
				Family:       address.Family,
				PrefixLength: address.PrefixLength,
			}
		}
		apiInterfaces[i] = utils.MachineReportNetworkInterface{
			Addresses: addresses,
			Index:     iface.Index,
			MAC:       iface.MAC,
			MTU:       iface.MTU,
			Name:      iface.Name,
			State:     iface.State,
			Type:      iface.Type,
		}
	}

	return utils.MachineReport{
		APIVersion: utils.APIVersionV1Alpha1,
		Kind:       utils.KindMachineReport,
		Metadata:   utils.Metadata{Name: "report"},
		Spec: utils.MachineReportSpec{
			CPU: utils.MachineReportCPU{
				Cores:         apiCores,
				LogicalCPUs:   cpu.LogicalCPUs,
				ModelName:     cpu.ModelName,
				PhysicalCores: cpu.PhysicalCores,
				Sockets:       cpu.Sockets,
				VendorID:      cpu.VendorID,
			},
			Interfaces: apiInterfaces,
			LLDPInfo:   machineReportLLDP(lldp),
			ObservedAt: observedAt,
			Storage:    apiStorage,
			System: utils.MachineReportSystem{
				BIOSDate:        system.BiosDate,
				BIOSRelease:     system.BiosRelease,
				BIOSVendor:      system.BiosVendor,
				BIOSVersion:     system.BIOSVersion,
				BoardAssetTag:   system.BoardAssetTag,
				BoardName:       system.BoardName,
				BoardSerial:     system.BoardSerial,
				BoardVendor:     system.BoardVendor,
				BoardVersion:    system.BoardVersion,
				ChassisAssetTag: system.ChassisAssetTag,
				ChassisSerial:   system.ChassisSerial,
				ChassisType:     system.ChassisType,
				ChassisVendor:   system.ChassisVendor,
				ChassisVersion:  system.ChassisVersion,
				Modalias:        system.Modalias,
				ProductFamily:   system.ProductFamily,
				ProductName:     system.ProductName,
				ProductSerial:   system.ProductSerial,
				ProductSKU:      system.ProductSKU,
				ProductUUID:     system.ProductUUID,
				ProductVersion:  system.ProductVersion,
				SysVendor:       system.SysVendor,
			},
		},
	}, nil
}

func machineReportLLDP(groups []LLDPInterfaceGroup) []utils.MachineReportLLDPInterfaceGroup {
	result := make([]utils.MachineReportLLDPInterfaceGroup, len(groups))
	for i, group := range groups {
		interfaces := make([]utils.MachineReportLLDPInterface, len(group.Interfaces))
		for j, iface := range group.Interfaces {
			chassis := make([]utils.MachineReportLLDPChassis, len(iface.Chassis))
			for k, item := range iface.Chassis {
				chassis[k] = utils.MachineReportLLDPChassis{
					Capabilities:         machineReportLLDPCapabilities(item.Capabilities),
					Descriptions:         machineReportLLDPValues(item.Descriptions),
					IDs:                  machineReportLLDPIdentifiers(item.IDs),
					ManagementInterfaces: machineReportLLDPValues(item.ManagementInterfaces),
					ManagementIPs:        machineReportLLDPValues(item.ManagementIPs),
					Names:                machineReportLLDPValues(item.Names),
				}
			}

			ports := make([]utils.MachineReportLLDPPort, len(iface.Ports))
			for k, port := range iface.Ports {
				ports[k] = utils.MachineReportLLDPPort{
					IDs:  machineReportLLDPIdentifiers(port.IDs),
					TTLs: machineReportLLDPValues(port.TTLs),
				}
			}

			interfaces[j] = utils.MachineReportLLDPInterface{
				Age:     iface.Age,
				Chassis: chassis,
				Name:    iface.Name,
				Ports:   ports,
				RID:     iface.RID,
				Via:     iface.Via,
			}
		}
		result[i] = utils.MachineReportLLDPInterfaceGroup{Interface: interfaces}
	}
	return result
}

func machineReportLLDPIdentifiers(values []LLDPIdentifier) []utils.MachineReportLLDPIdentifier {
	result := make([]utils.MachineReportLLDPIdentifier, len(values))
	for i, value := range values {
		result[i] = utils.MachineReportLLDPIdentifier{Type: value.Type, Value: value.Value}
	}
	return result
}

func machineReportLLDPValues(values []LLDPValue) []utils.MachineReportLLDPValue {
	result := make([]utils.MachineReportLLDPValue, len(values))
	for i, value := range values {
		result[i] = utils.MachineReportLLDPValue{Value: value.Value}
	}
	return result
}

func machineReportLLDPCapabilities(values []LLDPCapability) []utils.MachineReportLLDPCapability {
	result := make([]utils.MachineReportLLDPCapability, len(values))
	for i, value := range values {
		result[i] = utils.MachineReportLLDPCapability{Enabled: value.Enabled, Type: value.Type}
	}
	return result
}
