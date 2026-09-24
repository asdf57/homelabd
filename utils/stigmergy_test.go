package utils

import (
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestUploadMachineReportUsesInjectedHTTPClient(t *testing.T) {
	called := false
	client := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			called = true
			if got, want := request.URL.String(), "http://stigmergy.test/api/v1alpha1/machine-reports"; got != want {
				t.Fatalf("request URL = %q, want %q", got, want)
			}
			if got, want := request.Header.Get("Content-Type"), "application/json"; got != want {
				t.Fatalf("Content-Type = %q, want %q", got, want)
			}

			return &http.Response{
				Status:     "201 Created",
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	api := NewStigmergyApi(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		&Config{APIEndpoint: "http://stigmergy.test"},
		client,
	)

	if err := api.UploadMachineReport(MachineReport{
		APIVersion: APIVersionV1Alpha1,
		Kind:       KindMachineReport,
		Metadata:   Metadata{Name: "report"},
	}); err != nil {
		t.Fatalf("UploadMachineReport() error = %v", err)
	}
	if !called {
		t.Fatal("injected HTTP client was not used")
	}
}

func TestMachineLocationFromLLDP(t *testing.T) {
	location, err := MachineLocationFromLLDP([]MachineReportLLDPInterfaceGroup{{
		Interface: []MachineReportLLDPInterface{{
			Chassis: []MachineReportLLDPChassis{{
				IDs: []MachineReportLLDPIdentifier{{Type: "MAC", Value: " D4:01:C3:27:91:67 "}},
			}},
			Ports: []MachineReportLLDPPort{{
				IDs: []MachineReportLLDPIdentifier{{Type: "ifname", Value: " bridge/ether3 "}},
			}},
		}},
	}})
	if err != nil {
		t.Fatalf("MachineLocationFromLLDP() error = %v", err)
	}
	if want := (MachineLocation{LLDPPort: "bridge/ether3", SwitchMAC: "d4:01:c3:27:91:67"}); location != want {
		t.Fatalf("MachineLocationFromLLDP() = %+v, want %+v", location, want)
	}
}

func TestMachineLocationFromLLDPRequiresPortAndSwitchMAC(t *testing.T) {
	_, err := MachineLocationFromLLDP([]MachineReportLLDPInterfaceGroup{{
		Interface: []MachineReportLLDPInterface{{
			Ports: []MachineReportLLDPPort{{IDs: []MachineReportLLDPIdentifier{{Value: "ether3"}}}},
		}},
	}})
	if err == nil {
		t.Fatal("MachineLocationFromLLDP() error = nil, want incomplete-location error")
	}
}

func TestFindServerFromLocationDecodesServerList(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if got, want := request.URL.String(), "http://stigmergy.test/api/v1alpha1/servers"; got != want {
			t.Fatalf("request URL = %q, want %q", got, want)
		}
		return &http.Response{
			Status:     "200 OK",
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`{
				"apiVersion":"homelab.io/v1alpha1",
				"kind":"ServerList",
				"metadata":{"resourceVersion":"7"},
				"items":[{"apiVersion":"homelab.io/v1alpha1","kind":"Server","metadata":{"name":"desktop"},"spec":{"machineSelector":{"location":{"lldp_port":"bridge/ether3","switch_mac":"d4:01:c3:27:91:67"}}}}]
			}`)),
			Header: make(http.Header),
		}, nil
	})}
	api := NewStigmergyApi(slog.New(slog.NewTextHandler(io.Discard, nil)), &Config{APIEndpoint: "http://stigmergy.test"}, client)

	server, err := api.FindServerFromLocation(ServerMachineSelector{Location: MachineLocation{
		LLDPPort: " bridge/ether3 ", SwitchMAC: "D4:01:C3:27:91:67",
	}})
	if err != nil {
		t.Fatalf("FindServerFromLocation() error = %v", err)
	}
	if server.Metadata.Name != "desktop" {
		t.Fatalf("FindServerFromLocation() name = %q, want desktop", server.Metadata.Name)
	}
}
