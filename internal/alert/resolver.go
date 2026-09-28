package alert

import (
	"github.com/prasdud/big-brother/internal/store"
)

// ResolveChannelID returns the product channel id for a service: its override,
// else the project default, else "" (no workspace fallback).
func ResolveChannelID(svc store.Service, project store.Project) string {
	if svc.ChannelID != "" {
		return svc.ChannelID
	}
	return project.DefaultChannelID
}

// ResolveTemplate returns the message body for a trigger: service override,
// else project template, else the built-in default. projectTemplate may be "".
func ResolveTemplate(svc store.Service, projectTemplate, trigger string) string {
	switch trigger {
	case TriggerDown:
		if svc.TemplateDown != "" {
			return svc.TemplateDown
		}
	case TriggerRecovered:
		if svc.TemplateRecovered != "" {
			return svc.TemplateRecovered
		}
	}
	if projectTemplate != "" {
		return projectTemplate
	}
	return Builtin(trigger)
}

// TriggerFor maps a state transition to an alert trigger, or "" when the
// transition is not alertable.
func TriggerFor(from, to string) string {
	if to == "down" {
		return TriggerDown
	}
	if from == "down" && to == "up" {
		return TriggerRecovered
	}
	return ""
}
