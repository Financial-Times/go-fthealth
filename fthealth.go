package fthealth

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthCheck struct {
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

func (ch *HealthCheck) health() (result HealthResult) {
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
    w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	err := enc.Encode(health)
	if err != nil {
		panic("write this bit")
	}
}

type checkHandler struct {
	HealthCheck
}

func Handler(name, description string, checks ...Check) func(w http.ResponseWriter, r *http.Request) {
	ch := &checkHandler{HealthCheck{name, description, checks}}
	return ch.handle
}
