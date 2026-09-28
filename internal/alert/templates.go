// Package alert renders and delivers Slack alerts.
package alert

import "strings"

// Trigger names.
const (
	TriggerDown      = "down"
	TriggerRecovered = "recovered"
)

// Built-in templates, used when neither a service nor project template exists.
const (
	BuiltinDown = "DOWN: {{service.name}}\n" +
		"Project: {{project.name}}\n" +
		"URL: {{service.url}}\n" +
		"Status: {{status}}\n" +
		"Duration: {{duration}}\n" +
		"Error: {{error}}"

	BuiltinRecovered = "RECOVERED: {{service.name}}\n" +
		"Project: {{project.name}}\n" +
		"URL: {{service.url}}\n" +
		"Status: {{status}}\n" +
		"Was down for: {{duration}}"
)

// Vars are the values available to a template.
type Vars struct {
	ServiceName string
	ServiceURL  string
	ProjectName string
	Status      string
	Duration    string
	Error       string
}

// Builtin returns the built-in template for a trigger, or "" if unknown.
func Builtin(trigger string) string {
	switch trigger {
	case TriggerDown:
		return BuiltinDown
	case TriggerRecovered:
		return BuiltinRecovered
	default:
		return ""
	}
}

// Render substitutes the documented template variables.
func Render(tmpl string, v Vars) string {
	replacer := strings.NewReplacer(
		"{{service.name}}", v.ServiceName,
		"{{service.url}}", v.ServiceURL,
		"{{project.name}}", v.ProjectName,
		"{{status}}", v.Status,
		"{{duration}}", v.Duration,
		"{{error}}", v.Error,
	)
	return replacer.Replace(tmpl)
}
