package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type StigmergyApi struct {
	config     *Config
	logger     *slog.Logger
	httpClient *http.Client
}

func NewStigmergyApi(logger *slog.Logger, config *Config, httpClient *http.Client) *StigmergyApi {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return fmt.Errorf("API redirects are not accepted") },
		}
	}

	return &StigmergyApi{
		config:     config,
		logger:     logger,
		httpClient: httpClient,
	}
}

// UploadMachineReport submits a report matching Stigmergy's API model.
func (s *StigmergyApi) UploadMachineReport(report MachineReport) error {
	jsonBytes, err := json.Marshal(report)
	if err != nil {
		s.logger.Error("failed to marshal report", "error", err)
		return err
	}

	bodyBuffer := bytes.NewBuffer(jsonBytes)

	request, err := http.NewRequest(http.MethodPost, s.config.APIEndpoint+"/api/v1alpha1/machine-reports", bodyBuffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := s.do(request)
	if err != nil {
		s.logger.Error("failed to submit machine report", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit machine report: Stigmergy returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	// _, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Printf("Error reading response body: %v\n", err)
	// 	return err
	// }

	s.logger.Info("machine report submitted", "status", resp.Status)
	return nil
}

func (s *StigmergyApi) FindServerFromLocation(location ServerMachineSelector) (*Server, error) {
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1alpha1/servers", s.config.APIEndpoint), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.do(request)
	if err != nil {
		s.logger.Error("failed to get servers", "error", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list servers: Stigmergy returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var servers ServerList
	if err := json.NewDecoder(resp.Body).Decode(&servers); err != nil {
		s.logger.Error("failed to decode servers response", "error", err)
		return nil, err
	}

	target := canonicalMachineLocation(location.Location)
	if target.LLDPPort == "" || target.SwitchMAC == "" {
		return nil, fmt.Errorf("cannot find server without a complete LLDP location")
	}
	var match *Server
	for _, server := range servers.Items {
		if canonicalMachineLocation(server.Spec.MachineSelector.Location) == target {
			if match != nil {
				return nil, fmt.Errorf("multiple servers found for LLDP location %s/%s", target.SwitchMAC, target.LLDPPort)
			}
			match = &server
		}
	}
	if match != nil {
		return match, nil
	}

	return nil, fmt.Errorf("no server found for location: %+v", location)
}

func (s *StigmergyApi) do(request *http.Request) (*http.Response, error) {
	if s.config.APIToken != "" {
		request.Header.Set("Authorization", "Bearer "+s.config.APIToken)
	}
	return s.httpClient.Do(request)
}

// MachineLocationFromLLDP derives the same switch/port identity that
// Stigmergy's MachineReport controller derives from a submitted report.
func MachineLocationFromLLDP(groups []MachineReportLLDPInterfaceGroup) (MachineLocation, error) {
	var location MachineLocation
	for _, group := range groups {
		for _, iface := range group.Interface {
			if location.LLDPPort == "" {
				for _, port := range iface.Ports {
					for _, id := range port.IDs {
						if id.Value != "" {
							location.LLDPPort = id.Value
							break
						}
					}
				}
			}
			if location.SwitchMAC == "" {
				for _, chassis := range iface.Chassis {
					for _, id := range chassis.IDs {
						if strings.EqualFold(id.Type, "mac") && id.Value != "" {
							location.SwitchMAC = id.Value
							break
						}
					}
				}
			}
			if location.LLDPPort != "" && location.SwitchMAC != "" {
				location = canonicalMachineLocation(location)
				if location.LLDPPort != "" && location.SwitchMAC != "" {
					return location, nil
				}
				return MachineLocation{}, fmt.Errorf("machine report does not contain a complete LLDP location")
			}
		}
	}

	return MachineLocation{}, fmt.Errorf("machine report does not contain a complete LLDP location")
}

func canonicalMachineLocation(location MachineLocation) MachineLocation {
	return MachineLocation{
		LLDPPort:  strings.TrimSpace(location.LLDPPort),
		SwitchMAC: strings.ToLower(strings.TrimSpace(location.SwitchMAC)),
	}
}
