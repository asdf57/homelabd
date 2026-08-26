package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
			Timeout: 30 * time.Second,
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

	resp, err := s.httpClient.Post(s.config.APIEndpoint+"/api/v1alpha1/machine-reports", "application/json", bodyBuffer)
	if err != nil {
		s.logger.Error("failed to submit machine report", "error", err)
		return err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return err
	}

	s.logger.Info("machine report submitted", "status", resp.Status, "body", string(respBytes))
	return nil
}
