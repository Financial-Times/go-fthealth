package fthealth

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
	"time"
)

type healthCheck struct {
	name        string
	description string
	checks      []Check
}

type Check struct {
	BusinessImpact   string
	Name             string
	PanicGuide       string
	Severity         uint8 //TODO: enumerate
	TechnicalSummary string
	Checker          func() error
}

func (ch *healthCheck) health() (result HealthResult) {
	result.Name = ch.name
	result.Description = ch.description
	result.SchemaVersion = 1
	for _, checker := range ch.checks {
		result.Checks = append(result.Checks, runChecker(checker))
	}
	return
}

func runChecker(ch Check) CheckResult {
	result := CheckResult{
		BusinessImpact:   ch.BusinessImpact,
		LastUpdated:      time.Now(),
		Name:             ch.Name,
		PanicGuide:       ch.PanicGuide,
		Severity:         ch.Severity,
		TechnicalSummary: ch.TechnicalSummary,
	}
	err := ch.Checker()
	if err != nil {
		result.Ok = false
		result.Output = err.Error()
	} else {
		result.Ok = true
	}
	return result
}

func (ch *checkHandler) handle(w http.ResponseWriter, r *http.Request) {
	health := ch.health()

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		err := writeHTMLResp(w, health)
		if err == nil {
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err := enc.Encode(health)
	if err != nil {
		panic("write this bit")
	}
}

type checkHandler struct {
	healthCheck
}

func Handler(name, description string, checks ...Check) func(w http.ResponseWriter, r *http.Request) {
	ch := &checkHandler{healthCheck{name, description, checks}}
	return ch.handle
}

func writeHTMLResp(w http.ResponseWriter, health HealthResult) error {
	w.Header().Set("Content-Type", "text/html")
	t := template.New("healthchecks")
	t, err := t.Parse(` <!DOCTYPE html>
                            <head>
                                <title>Healthchecks</title>
                            </head>

                            <body>
                                <h2>Checks</h2>
                                <ul>
                                    {{ range $key, $value := . }}
                                        <li>
                                            <strong> {{ $value.Name }} </strong>
                                            <ul>
                                                <li> <strong> Ok: {{ $value.Ok }} </strong> </li>
                                                <li> Business impact: {{ $value.BusinessImpact }} </li>
                                                <li> Output: {{ $value.Output }} </li>
                                                <li> Last updated: {{ $value.LastUpdated }} </li>
                                                <li> Panic guide: <a href="{{ $value.PanicGuide }}">{{ $value.PanicGuide }}</a> </li>
                                                <li> Severity: {{ $value.Severity }} </li>
                                                <li> Technical summary: {{ $value.TechnicalSummary }} </li>
                                            </ul>
                                        </li>
                                    {{ end }}
                                </ul>
                            </body>`)
	if err == nil {
		t.Execute(w, health.Checks)
	}
	return err
}
