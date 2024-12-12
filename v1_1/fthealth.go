package v1_1

import (
	"sync"
	"time"
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

// TimedHealthCheck New type for fail-safe backward compatibility
type TimedHealthCheck struct {
	HealthCheck
	Timeout time.Duration
}

type FeedbackHealthCheck struct {
	HC
	feedback chan<- bool
}

func NewFeedbackHealthCheck(hc HC, fb chan<- bool) FeedbackHealthCheck {
	return FeedbackHealthCheck{hc, fb}
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

func (fch FeedbackHealthCheck) initResult(result *HealthResult) {
	fch.HC.initResult(result)
}

func (fch FeedbackHealthCheck) doChecks(result *HealthResult) {
	fch.HC.doChecks(result)
	fch.feedback <- ComputeOverallStatus(result)
}

func (ch TimedHealthCheck) doChecks(result *HealthResult) {
	lc := len(ch.Checks)
	result.Checks = make([]CheckResult, lc)
	wg := sync.WaitGroup{}
	wg.Add(lc)
	for i, c := range ch.Checks {
		go func(i int, c Check) {
			c.Timeout = ch.Timeout
			result.Checks[i] = c.runChecker()
			wg.Done()
		}(i, c)
	}
	wg.Wait()
}
