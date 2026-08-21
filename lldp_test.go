package main

import (
	"encoding/json"
	"testing"
)

func TestLLDPOutputUnmarshal(t *testing.T) {
	const input = `{
		"lldp": [{
			"interface": [{
				"name": "enp7s0",
				"via": "LLDP",
				"rid": "1",
				"age": "0 day, 14:31:03",
				"chassis": [{
					"id": [{"type": "mac", "value": "d4:01:c3:27:91:67"}],
					"name": [{"value": "MikroTik"}],
					"descr": [{"value": "MikroTik RouterOS"}],
					"mgmt-ip": [{"value": "10.0.2.1"}],
					"mgmt-iface": [{"value": "3"}],
					"capability": [{"type": "Router", "enabled": true}]
				}],
				"port": [{
					"id": [{"type": "ifname", "value": "bridge/ether3"}],
					"ttl": [{"value": "120"}]
				}]
			}]
		}]
	}`

	var output LLDPOutput
	if err := json.Unmarshal([]byte(input), &output); err != nil {
		t.Fatalf("unmarshal LLDP output: %v", err)
	}

	if len(output.LLDP) != 1 || len(output.LLDP[0].Interfaces) != 1 {
		t.Fatalf("unexpected interface groups: %#v", output.LLDP)
	}

	iface := output.LLDP[0].Interfaces[0]
	if iface.Name != "enp7s0" || iface.Chassis[0].ManagementIPs[0].Value != "10.0.2.1" {
		t.Fatalf("unexpected interface: %#v", iface)
	}
	if iface.Ports[0].IDs[0].Value != "bridge/ether3" || iface.Ports[0].TTLs[0].Value != "120" {
		t.Fatalf("unexpected port: %#v", iface.Ports[0])
	}
}
