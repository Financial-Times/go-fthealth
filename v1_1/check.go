package v1_1

import (
	"strings"
	"time"
)

type Check struct {
	ID               string
	Name             string
	Severity         uint8
	BusinessImpact   string
	TechnicalSummary string
	PanicGuide       string
	Checker          func() (string, error)
}

func (ch *Check) runChecker() CheckResult {
	result := CheckResult{
		ID:               ch.ID,
		Name:             ch.Name,
		Severity:         ch.Severity,
		BusinessImpact:   ch.BusinessImpact,
		TechnicalSummary: ch.TechnicalSummary,
		PanicGuide:       ch.PanicGuide,
		PanicGuideIsLink: strings.HasPrefix(ch.PanicGuide, "http"),
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
