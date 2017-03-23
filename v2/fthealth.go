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
	return hc.health(true)
}

func RunCheckSerial(hc *HealthCheck) HealthResult {
	return hc.health(false)
}

func (ch *HealthCheck) health(parallel bool) (result HealthResult) {
	result.SchemaVersion = 1
	result.SystemCode = ch.SystemCode
	result.Name = ch.Name
	result.Description = ch.Description

	ch.doChecks(&result, parallel)

	result.Ok = ComputeOverallStatus(&result)
	if result.Ok == false {
		result.Severity = ComputeOverallSeverity(&result)
	}
	return
}

func (ch *HealthCheck) doChecks(result *HealthResult, parallel bool) {
	if !parallel {
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
