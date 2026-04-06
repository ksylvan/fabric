package restapi

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	errInvalidBindAddress           = errors.New("invalid bind address")
	errRESTServerRequiresAPIKey     = errors.New("non-loopback REST API binding requires --api-key")
	errOllamaServerRequiresLoopback = errors.New("ollama compatibility server must bind to a loopback address")
)

func validateRESTServerConfig(address, apiKey string) error {
	isLoopback, err := isLoopbackBindAddress(address)
	if err != nil {
		return err
	}
	if isLoopback || apiKey != "" {
		return nil
	}
	return fmt.Errorf("%w: %q", errRESTServerRequiresAPIKey, address)
}

func validateOllamaServerConfig(address string) error {
	isLoopback, err := isLoopbackBindAddress(address)
	if err != nil {
		return err
	}
	if isLoopback {
		return nil
	}
	return fmt.Errorf("%w: %q", errOllamaServerRequiresLoopback, address)
}

func isLoopbackBindAddress(address string) (bool, error) {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return false, fmt.Errorf("%w: %q", errInvalidBindAddress, address)
	}
	if host == "" || host == "*" {
		return false, nil
	}
	if strings.EqualFold(host, "localhost") {
		return true, nil
	}

	if zoneIndex := strings.Index(host, "%"); zoneIndex >= 0 {
		host = host[:zoneIndex]
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return false, nil
	}
	return ip.IsLoopback(), nil
}
