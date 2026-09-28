// Package kuma imports a Kuma JSON export into projects and services.
package kuma

import "encoding/json"

// Export is the subset of a Kuma export we understand. Unknown fields are
// ignored on purpose so newer exports still import.
type Export struct {
	Version       string         `json:"version"`
	Monitors      []Monitor      `json:"monitors"`
	Notifications []Notification `json:"notifications"`
}

// Monitor is a Kuma monitor.
type Monitor struct {
	ID                 int             `json:"id"`
	Name               string          `json:"name"`
	Type               string          `json:"type"`
	URL                string          `json:"url"`
	Hostname           string          `json:"hostname"`
	Port               *int            `json:"port"`
	Interval           int             `json:"interval"`
	Timeout            int             `json:"timeout"`
	MaxRetries         int             `json:"maxretries"`
	Parent             *int            `json:"parent"`
	Active             *bool           `json:"active"`
	NotificationIDList map[string]bool `json:"notificationIDList"`
}

// Notification is a Kuma notification provider. Config may contain secrets and
// is never copied into a report or a log.
type Notification struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	IsDefault bool   `json:"isDefault"`
	Config    string `json:"config"`
}

// Report summarises what an import did or would do.
type Report struct {
	Version  string          `json:"version"`
	Projects []string        `json:"projects"`
	Services []ServiceReport `json:"services"`
	Skipped  []Skip          `json:"skipped"`
	Warnings []string        `json:"warnings"`
}

// ServiceReport is one service in a report.
type ServiceReport struct {
	Name    string `json:"name"`
	Project string `json:"project"`
	Type    string `json:"type"`
	Action  string `json:"action"`
}

// Skip records a monitor or notification that was not imported.
type Skip struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

// Parse decodes a Kuma export.
func Parse(data []byte) (Export, error) {
	var export Export
	if err := json.Unmarshal(data, &export); err != nil {
		return Export{}, err
	}
	return export, nil
}

// supportedTypes are the monitor types the product can check.
var supportedTypes = map[string]bool{"http": true, "tcp": true, "dns": true}

// groupType marks a Kuma monitor that represents a group of monitors.
const groupType = "group"
