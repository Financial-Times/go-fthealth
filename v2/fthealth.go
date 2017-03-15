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
	ch.setResultBasicInfo(&result)
	ch.doChecks(ch.Parallel, &result)
	setResultGlobalOK(&result)
	return
}

func (ch *HealthCheck) setResultBasicInfo(result *HealthResult) {
	result.SchemaVersion = 1
	result.SystemCode = ch.SystemCode
	result.Name = ch.Name
	result.Description = ch.Description
}

func (ch *HealthCheck) doChecks(parallel bool, result *HealthResult) {
	if parallel {
		result.Checks = make([]CheckResult, len(ch.Checks))
		wg := sync.WaitGroup{}
		for i := 0; i < len(ch.Checks); i++ {
			wg.Add(1)
			go func(i int) {
				result.Checks[i] = runChecker(ch.Checks[i])
				wg.Done()
			}(i)
		}
		wg.Wait()
	} else {
		for _, checker := range ch.Checks {
			result.Checks = append(result.Checks, runChecker(checker))
		}
	}
}

func runChecker(ch Check) CheckResult {
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

func setResultGlobalOK(result *HealthResult) {
	result.Ok = getOverallStatus(result)
	if result.Ok == false {
		result.Severity = getOverallSeverity(result)
	}
}

func getOverallStatus(result *HealthResult) bool {
	for _, check := range result.Checks {
		if !check.Ok {
			return false
		}
	}
	return true
}

func getOverallSeverity(result *HealthResult) uint8 {
	var severity uint8 = 3
	for _, check := range result.Checks {
		if check.Ok == false && check.Severity < severity {
			severity = check.Severity
		}
	}
	return severity
}
