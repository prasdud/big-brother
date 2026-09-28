package api

import "errors"

const (
	defaultIntervalSeconds  = 60
	defaultTimeoutSeconds   = 10
	defaultFailureThreshold = 3
)

func (req *serviceRequest) applyDefaults() {
	if req.IntervalSeconds == 0 {
		req.IntervalSeconds = defaultIntervalSeconds
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = defaultTimeoutSeconds
	}
	if req.FailureThreshold == 0 {
		req.FailureThreshold = defaultFailureThreshold
	}
}

func (req serviceRequest) validate() error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	switch req.Type {
	case "http":
		if req.URL == "" {
			return errors.New("url is required for http services")
		}
	case "tcp":
		if req.Hostname == "" {
			return errors.New("hostname is required for tcp services")
		}
		if req.Port < 1 || req.Port > 65535 {
			return errors.New("port must be between 1 and 65535 for tcp services")
		}
	case "dns":
		if req.Hostname == "" {
			return errors.New("hostname is required for dns services")
		}
	default:
		return errors.New("type must be one of http, tcp, dns")
	}
	if req.IntervalSeconds < 1 {
		return errors.New("interval_seconds must be positive")
	}
	if req.TimeoutSeconds < 1 {
		return errors.New("timeout_seconds must be positive")
	}
	if req.FailureThreshold < 1 {
		return errors.New("failure_threshold must be positive")
	}
	return nil
}
