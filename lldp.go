package main

type LLDPOutput struct {
	LLDP []LLDPInterfaceGroup `json:"lldp"`
}

type LLDPInterfaceGroup struct {
	Interfaces []LLDPInterface `json:"interface"`
}

type LLDPInterface struct {
	Name    string        `json:"name"`
	Via     string        `json:"via"`
	RID     string        `json:"rid"`
	Age     string        `json:"age"`
	Chassis []LLDPChassis `json:"chassis"`
	Ports   []LLDPPort    `json:"port"`
}

type LLDPChassis struct {
	IDs                  []LLDPIdentifier `json:"id"`
	Names                []LLDPValue      `json:"name"`
	Descriptions         []LLDPValue      `json:"descr"`
	ManagementIPs        []LLDPValue      `json:"mgmt-ip"`
	ManagementInterfaces []LLDPValue      `json:"mgmt-iface"`
	Capabilities         []LLDPCapability `json:"capability"`
}

type LLDPPort struct {
	IDs  []LLDPIdentifier `json:"id"`
	TTLs []LLDPValue      `json:"ttl"`
}

type LLDPIdentifier struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type LLDPValue struct {
	Value string `json:"value"`
}

type LLDPCapability struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}
