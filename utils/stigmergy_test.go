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
