package utils

import "time"

const (
	APIVersionV1Alpha1 = "homelab.io/v1alpha1"
	KindMachineReport  = "MachineReport"
	KindServer         = "Server"
)

type Metadata struct {
	Annotations       *map[string]string `json:"annotations,omitempty"`
	CreationTimestamp *time.Time         `json:"creationTimestamp,omitempty"`
	DeletionTimestamp *time.Time         `json:"deletionTimestamp,omitempty"`
	Finalizers        *[]string          `json:"finalizers,omitempty"`
	Generation        *int64             `json:"generation,omitempty"`
	Labels            *map[string]string `json:"labels,omitempty"`
	Name              string             `json:"name"`
	ResourceVersion   *string            `json:"resourceVersion,omitempty"`
	UID               *string            `json:"uid,omitempty"`
}

type ResourceReference struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type MachineLocation struct {
	LLDPPort  string `json:"lldp_port"`
	SwitchMAC string `json:"switch_mac"`
}

type MachineReport struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   Metadata          `json:"metadata"`
	Spec       MachineReportSpec `json:"spec"`
}

type MachineReportSpec struct {
	CPU        MachineReportCPU                  `json:"cpu"`
	Interfaces []MachineReportNetworkInterface   `json:"interfaces"`
	LLDPInfo   []MachineReportLLDPInterfaceGroup `json:"lldp_info"`
	ObservedAt time.Time                         `json:"observed_at"`
	Storage    []MachineReportStorageDevice      `json:"storage"`
	System     MachineReportSystem               `json:"system"`
}

type MachineReportCPU struct {
	Cores         []MachineReportCPUCore `json:"cores"`
	LogicalCPUs   int                    `json:"logical_cpus"`
	ModelName     string                 `json:"model_name"`
	PhysicalCores int                    `json:"physical_cores"`
	Sockets       int                    `json:"sockets"`
	VendorID      string                 `json:"vendor_id"`
}

type MachineReportCPUCore struct {
	AddressSizes    string   `json:"address_sizes"`
	APICID          int      `json:"apicid"`
	Bogomips        float32  `json:"bogomips"`
	Bugs            []string `json:"bugs"`
	CacheAlignment  int      `json:"cache_alignment"`
	CacheSizeKB     int      `json:"cache_size_kb"`
	ClflushSize     int      `json:"clflush_size"`
	CoreID          int      `json:"core_id"`
	CPUCores        int      `json:"cpu_cores"`
	CPUFamily       int      `json:"cpu_family"`
	Flags           []string `json:"flags"`
	InitialAPICID   int      `json:"initial_apicid"`
	MHz             float32  `json:"mhz"`
	Microcode       string   `json:"microcode"`
	Model           int      `json:"model"`
	ModelName       string   `json:"model_name"`
	PhysicalID      int      `json:"physical_id"`
	PowerManagement []string `json:"power_management"`
	Processor       int      `json:"processor"`
	Siblings        int      `json:"siblings"`
	Stepping        int      `json:"stepping"`
	VendorID        string   `json:"vendor_id"`
}

type MachineReportStorageDevice struct {
	KName  string `json:"kname"`
	Model  string `json:"model"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Rota   bool   `json:"rota"`
	Serial string `json:"serial"`
	Size   int64  `json:"size"`
	Tran   string `json:"tran"`
	Type   string `json:"type"`
	WWN    string `json:"wwn"`
}

type MachineReportSystem struct {
	BIOSDate        string `json:"bios_date"`
	BIOSRelease     string `json:"bios_release"`
	BIOSVendor      string `json:"bios_vendor"`
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
	ProductUUID     string `json:"product_uuid"`
	ProductVersion  string `json:"product_version"`
	SysVendor       string `json:"sys_vendor"`
}

type MachineReportNetworkInterface struct {
	Addresses []MachineReportNetworkAddress `json:"addresses"`
	Index     int                           `json:"index"`
	MAC       string                        `json:"mac"`
	MTU       int                           `json:"mtu"`
	Name      string                        `json:"name"`
	State     string                        `json:"state"`
	Type      string                        `json:"type"`
}

type MachineReportNetworkAddress struct {
	Address      string `json:"address"`
	Family       string `json:"family"`
	PrefixLength int    `json:"prefix_length"`
}

type MachineReportLLDPCapability struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
}

type MachineReportLLDPChassis struct {
	Capabilities         []MachineReportLLDPCapability `json:"capability"`
	Descriptions         []MachineReportLLDPValue      `json:"descr"`
	IDs                  []MachineReportLLDPIdentifier `json:"id"`
	ManagementInterfaces []MachineReportLLDPValue      `json:"mgmt-iface"`
	ManagementIPs        []MachineReportLLDPValue      `json:"mgmt-ip"`
	Names                []MachineReportLLDPValue      `json:"name"`
}

type MachineReportLLDPIdentifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type MachineReportLLDPInterface struct {
	Age     string                     `json:"age"`
	Chassis []MachineReportLLDPChassis `json:"chassis"`
	Name    string                     `json:"name"`
	Ports   []MachineReportLLDPPort    `json:"port"`
	RID     string                     `json:"rid"`
	Via     string                     `json:"via"`
}

type MachineReportLLDPInterfaceGroup struct {
	Interface []MachineReportLLDPInterface `json:"interface"`
}

type MachineReportLLDPPort struct {
	IDs  []MachineReportLLDPIdentifier `json:"id"`
	TTLs []MachineReportLLDPValue      `json:"ttl"`
}

type MachineReportLLDPValue struct {
	Value string `json:"value"`
}

type Server struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Metadata   Metadata      `json:"metadata"`
	Spec       ServerSpec    `json:"spec"`
	Status     *ServerStatus `json:"status,omitempty"`
}

type ServerSpec struct {
	DomainName      *string                   `json:"domainName,omitempty"`
	FeatureFlags    *ServerFeatureFlags       `json:"featureFlags,omitempty"`
	Groups          *[]string                 `json:"groups,omitempty"`
	HostName        *string                   `json:"hostName,omitempty"`
	MachineSelector ServerMachineSelector     `json:"machineSelector"`
	Networking      *ServerNetworkingSpec     `json:"networking,omitempty"`
	OperatingSystem *ServerOperatingSystem    `json:"operatingSystem,omitempty"`
	Packages        *[]string                 `json:"packages,omitempty"`
	Provisioning    *ServerProvisioningSpec   `json:"provisioning,omitempty"`
	Reconciliation  *ServerReconciliationSpec `json:"reconciliation,omitempty"`
	Sysctls         *map[string]string        `json:"sysctls,omitempty"`
	Users           *[]ServerUser             `json:"users,omitempty"`
}

type ServerMachineSelector struct {
	Location MachineLocation `json:"location"`
}

type ServerFeatureFlags struct {
	Backup     *ServerFeatureToggle     `json:"backup,omitempty"`
	Daemon     *ServerFeatureToggle     `json:"daemon,omitempty"`
	Kubernetes *ServerKubernetesFeature `json:"kubernetes,omitempty"`
	Monitoring *ServerFeatureToggle     `json:"monitoring,omitempty"`
}

type ServerFeatureToggle struct {
	Enabled bool `json:"enabled"`
}

type ServerKubernetesFeature struct {
	Enabled bool    `json:"enabled"`
	Role    *string `json:"role,omitempty"`
}

type ServerNetworkingSpec struct {
	Management *ServerManagementNetworkSpec `json:"management,omitempty"`
}

type ServerManagementNetworkSpec struct {
	AddressSelector   ServerManagementAddressSelector   `json:"addressSelector"`
	InterfaceSelector ServerManagementInterfaceSelector `json:"interfaceSelector"`
}

type ServerManagementAddressSelector struct {
	Family string `json:"family"`
	Subnet string `json:"subnet"`
}

type ServerManagementInterfaceSelector struct {
	AttachedAtMachineLocation bool `json:"attachedAtMachineLocation"`
}

type ServerOperatingSystem struct {
	Architecture       string                    `json:"architecture"`
	BootMode           string                    `json:"bootMode"`
	Distribution       string                    `json:"distribution"`
	InstallationSource *ServerInstallationSource `json:"installationSource,omitempty"`
	KernelArguments    *[]string                 `json:"kernelArguments,omitempty"`
	Locale             *string                   `json:"locale,omitempty"`
	Timezone           *string                   `json:"timezone,omitempty"`
	Version            string                    `json:"version"`
}

type ServerInstallationSource struct {
	Checksum ServerChecksum `json:"checksum"`
	URL      string         `json:"url"`
}

type ServerChecksum struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

type ServerProvisioningSpec struct {
	Enabled bool `json:"enabled"`
}

type ServerReconciliationSpec struct {
	Paused bool `json:"paused"`
}

type ServerUser struct {
	Groups *[]string `json:"groups,omitempty"`
	Locked *bool     `json:"locked,omitempty"`
	Name   string    `json:"name"`
	Shell  *string   `json:"shell,omitempty"`
}

type ServerStatus struct {
	Agent              *ServerAgentStatus        `json:"agent,omitempty"`
	Conditions         *[]ServerCondition        `json:"conditions,omitempty"`
	Features           *ServerFeatureStatuses    `json:"features,omitempty"`
	FQDN               *string                   `json:"fqdn,omitempty"`
	MachineRef         *ResourceReference        `json:"machineRef,omitempty"`
	Networking         *ServerNetworkingStatus   `json:"networking,omitempty"`
	ObservedAddresses  *[]string                 `json:"observedAddresses,omitempty"`
	ObservedGeneration *int64                    `json:"observedGeneration,omitempty"`
	Phase              *string                   `json:"phase,omitempty"`
	Provisioning       *ServerProvisioningStatus `json:"provisioning,omitempty"`
	ResourceSummary    *ServerResourceSummary    `json:"resourceSummary,omitempty"`
	SSH                *ServerSSHStatus          `json:"ssh,omitempty"`
	SystemStats        *ServerSystemStats        `json:"systemStats,omitempty"`
}

type ServerAgentStatus struct {
	LastSeenTime *time.Time `json:"lastSeenTime,omitempty"`
	Reachable    bool       `json:"reachable"`
	Version      *string    `json:"version,omitempty"`
}

type ServerCondition struct {
	LastTransitionTime *time.Time `json:"lastTransitionTime,omitempty"`
	Message            *string    `json:"message,omitempty"`
	ObservedGeneration *int64     `json:"observedGeneration,omitempty"`
	Reason             string     `json:"reason"`
	Status             string     `json:"status"`
	Type               string     `json:"type"`
}

type ServerFeatureStatus struct {
	Healthy   bool `json:"healthy"`
	Installed bool `json:"installed"`
}

type ServerFeatureStatuses struct {
	Backup     *ServerFeatureStatus `json:"backup,omitempty"`
	Daemon     *ServerFeatureStatus `json:"daemon,omitempty"`
	Kubernetes *ServerFeatureStatus `json:"kubernetes,omitempty"`
	Monitoring *ServerFeatureStatus `json:"monitoring,omitempty"`
}

type ServerNetworkingStatus struct {
	Management *ServerManagementNetworkStatus `json:"management,omitempty"`
}

type ServerManagementNetworkStatus struct {
	Address   ServerManagementAddressStatus   `json:"address"`
	Interface ServerManagementInterfaceStatus `json:"interface"`
	Reason    string                          `json:"reason"`
}

type ServerManagementAddressStatus struct {
	Address      string `json:"address"`
	Family       string `json:"family"`
	PrefixLength int    `json:"prefixLength"`
}

type ServerManagementInterfaceStatus struct {
	MAC  string `json:"mac"`
	Name string `json:"name"`
}

type ServerProvisioningStatus struct {
	AttemptID                *string    `json:"attemptID,omitempty"`
	AttemptNumber            *int       `json:"attemptNumber,omitempty"`
	BackendRunID             *string    `json:"backendRunID,omitempty"`
	CompletedAt              *time.Time `json:"completedAt,omitempty"`
	CurrentStage             *string    `json:"currentStage,omitempty"`
	Message                  *string    `json:"message,omitempty"`
	ObservedServerGeneration *int64     `json:"observedServerGeneration,omitempty"`
	Phase                    *string    `json:"phase,omitempty"`
	Provisioned              bool       `json:"provisioned"`
	StartedAt                *time.Time `json:"startedAt,omitempty"`
}

type ServerResourceSummary struct {
	Addresses         *int `json:"addresses,omitempty"`
	DiskClaims        *int `json:"diskClaims,omitempty"`
	Interfaces        *int `json:"interfaces,omitempty"`
	NetworkPortClaims *int `json:"networkPortClaims,omitempty"`
}

type ServerSSHAuthorizedKeyStatus struct {
	AccessGrantRef ResourceReference `json:"accessGrantRef"`
	Fingerprint    string            `json:"fingerprint"`
	LoginUser      string            `json:"loginUser"`
	PublicKey      string            `json:"publicKey"`
}

type ServerSSHStatus struct {
	AuthorizedKeys []ServerSSHAuthorizedKeyStatus `json:"authorizedKeys"`
}

type ServerSystemStats struct {
	CPUPercent    *float32 `json:"cpuPercent,omitempty"`
	LoadAverage1m *float32 `json:"loadAverage1m,omitempty"`
	MemoryPercent *float32 `json:"memoryPercent,omitempty"`
}
