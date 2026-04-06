package restapi

import (
	"errors"
	"testing"
)

func TestIsLoopbackBindAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
		wantErr bool
	}{
		{
			name:    "localhost host",
			address: "localhost:8080",
			want:    true,
		},
		{
			name:    "ipv4 loopback",
			address: "127.0.0.1:8080",
			want:    true,
		},
		{
			name:    "ipv6 loopback",
			address: "[::1]:8080",
			want:    true,
		},
		{
			name:    "wildcard shorthand",
			address: ":8080",
			want:    false,
		},
		{
			name:    "wildcard ipv4",
			address: "0.0.0.0:8080",
			want:    false,
		},
		{
			name:    "named host is treated as non-loopback",
			address: "api.example.com:8080",
			want:    false,
		},
		{
			name:    "missing port is invalid",
			address: "localhost",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isLoopbackBindAddress(tt.address)
			if tt.wantErr {
				if !errors.Is(err, errInvalidBindAddress) {
					t.Fatalf("expected invalid bind address error, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestValidateRESTServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		address string
		apiKey  string
		wantErr error
	}{
		{
			name:    "loopback without api key remains allowed",
			address: "127.0.0.1:8080",
		},
		{
			name:    "wildcard without api key is rejected",
			address: ":8080",
			wantErr: errRESTServerRequiresAPIKey,
		},
		{
			name:    "named host without api key is rejected",
			address: "api.example.com:8080",
			wantErr: errRESTServerRequiresAPIKey,
		},
		{
			name:    "wildcard with api key is allowed",
			address: ":8080",
			apiKey:  "secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRESTServerConfig(tt.address, tt.apiKey)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidateOllamaServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr error
	}{
		{
			name:    "loopback remains allowed",
			address: "localhost:11434",
		},
		{
			name:    "wildcard is rejected",
			address: ":11434",
			wantErr: errOllamaServerRequiresLoopback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOllamaServerConfig(tt.address)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
