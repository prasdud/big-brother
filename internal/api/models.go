package api

import (
	"strings"

	"github.com/prasdud/big-brother/internal/store"
)

// splitTags parses the stored comma-separated tag list.
func splitTags(raw string) []string {
	out := []string{}
	for _, part := range strings.Split(raw, ",") {
		if tag := strings.TrimSpace(part); tag != "" {
			out = append(out, tag)
		}
	}
	return out
}

// joinTags normalizes a tag list for storage.
func joinTags(tags []string) string {
	clean := make([]string, 0, len(tags))
	for _, tag := range tags {
		if trimmed := strings.TrimSpace(tag); trimmed != "" {
			clean = append(clean, trimmed)
		}
	}
	return strings.Join(clean, ",")
}

type projectView struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	DefaultChannelID string `json:"default_channel_id"`
	CreatedAt        string `json:"created_at"`
}

type serviceView struct {
	ID                string   `json:"id"`
	ProjectID         string   `json:"project_id"`
	Name              string   `json:"name"`
	Slug              string   `json:"slug"`
	Type              string   `json:"type"`
	URL               string   `json:"url"`
	Hostname          string   `json:"hostname"`
	Port              int64    `json:"port"`
	IntervalSeconds   int64    `json:"interval_seconds"`
	TimeoutSeconds    int64    `json:"timeout_seconds"`
	FailureThreshold  int64    `json:"failure_threshold"`
	Enabled           bool     `json:"enabled"`
	ChannelID         string   `json:"channel_id"`
	TemplateDown      string   `json:"template_down"`
	TemplateRecovered string   `json:"template_recovered"`
	Tags              []string `json:"tags"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

func newProjectView(p store.Project) projectView {
	return projectView{
		ID:               p.ID,
		Name:             p.Name,
		Slug:             p.Slug,
		DefaultChannelID: p.DefaultChannelID,
		CreatedAt:        p.CreatedAt,
	}
}

func newServiceView(s store.Service) serviceView {
	return serviceView{
		ID:                s.ID,
		ProjectID:         s.ProjectID,
		Name:              s.Name,
		Slug:              s.Slug,
		Type:              s.Type,
		URL:               s.Url,
		Hostname:          s.Hostname,
		Port:              s.Port,
		IntervalSeconds:   s.IntervalSeconds,
		TimeoutSeconds:    s.TimeoutSeconds,
		FailureThreshold:  s.FailureThreshold,
		Enabled:           s.Enabled != 0,
		ChannelID:         s.ChannelID,
		TemplateDown:      s.TemplateDown,
		TemplateRecovered: s.TemplateRecovered,
		Tags:              splitTags(s.Tags),
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}

type nameRequest struct {
	Name string `json:"name"`
}

// serviceRequest is the create and update payload for a service.
type serviceRequest struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	URL               string   `json:"url"`
	Hostname          string   `json:"hostname"`
	Port              int64    `json:"port"`
	IntervalSeconds   int64    `json:"interval_seconds"`
	TimeoutSeconds    int64    `json:"timeout_seconds"`
	FailureThreshold  int64    `json:"failure_threshold"`
	TemplateDown      string   `json:"template_down"`
	TemplateRecovered string   `json:"template_recovered"`
	Tags              []string `json:"tags"`
}
