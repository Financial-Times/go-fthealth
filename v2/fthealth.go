package v2

import (
	"sync"
	"time"
)

type HealthCheck struct {
	SystemCode  string
	Name        string
	Description string
	Checks      []Check
	Parallel    bool
}

type Check struct {
	ID               string
	Name             string
	Severity         uint8
	BusinessImpact   string
	TechnicalSummary string
	PanicGuide       string
	Checker          func() (string, error)
}

func RunCheck(hc *HealthCheck) HealthResult {
	return hc.health()
}

func (ch *HealthCheck) health() (result HealthResult) {
	result.SchemaVersion = 1
	result.SystemCode = ch.SystemCode
	result.Name = ch.Name
	result.Description = ch.Description

	ch.doChecks(&result)

	result.Ok = ComputeOverallStatus(&result)
	if result.Ok == false {
		result.Severity = ComputeOverallSeverity(&result)
	}
	return
}

func (ch *HealthCheck) doChecks(result *HealthResult) {
	if !ch.Parallel {
		for _, checker := range ch.Checks {
			result.Checks = append(result.Checks, checker.runChecker())
		}
	} else {
		result.Checks = make([]CheckResult, len(ch.Checks))
		wg := sync.WaitGroup{}
		for i := 0; i < len(ch.Checks); i++ {
			wg.Add(1)
			go func(i int) {
				result.Checks[i] = ch.Checks[i].runChecker()
				wg.Done()
			}(i)
		}
		wg.Wait()
	}
}

func (ch *Check) runChecker() CheckResult {
	result := CheckResult{
		ID:               ch.ID,
		Name:             ch.Name,
		Severity:         ch.Severity,
		BusinessImpact:   ch.BusinessImpact,
		TechnicalSummary: ch.TechnicalSummary,
		PanicGuide:       ch.PanicGuide,
		LastUpdated:      time.Now(),
	}
	out, err := ch.Checker()
	if err != nil {
		result.Ok = false
		result.CheckOutput = err.Error()
	} else {
		result.Ok = true
		result.CheckOutput = out
	}
	return result
}
