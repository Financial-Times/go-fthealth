package v1_1

import (
	"sync"
)

type HC interface {
	initResult(result *HealthResult)
	doChecks(result *HealthResult)
}

type HealthCheck struct {
	SystemCode  string
	Name        string
	Description string
	Checks      []Check
}

type HealthCheckSerial struct {
	HealthCheck
}

func RunCheck(hc HC) (result HealthResult) {
	hc.initResult(&result)
	hc.doChecks(&result)

	result.Ok = ComputeOverallStatus(&result)
	if result.Ok == false {
		result.Severity = ComputeOverallSeverity(&result)
	}
	return
}

func (ch HealthCheck) initResult(result *HealthResult) {
	result.SchemaVersion = 1
	result.SystemCode = ch.SystemCode
	result.Name = ch.Name
	result.Description = ch.Description
}

func (ch HealthCheck) doChecks(result *HealthResult) {
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

func (chs HealthCheckSerial) doChecks(result *HealthResult) {
	for _, checker := range chs.Checks {
		result.Checks = append(result.Checks, checker.runChecker())
	}
}