package v2

import (
	"time"
)

type CheckResult struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Ok               bool      `json:"ok"`
	Severity         uint8     `json:"severity"`
	BusinessImpact   string    `json:"businessImpact"`
	TechnicalSummary string    `json:"technicalSummary"`
	PanicGuide       string    `json:"panicGuide"`
	CheckOutput      string    `json:"checkOutput"`
	LastUpdated      time.Time `json:"lastUpdated"`
	Ack              string    `json:"ack,omitempty"`
}

type HealthResult struct {
	SchemaVersion float64       `json:"schemaVersion"`
	SystemCode    string        `json:"systemCode"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Checks        []CheckResult `json:"checks"`
	Ok            bool          `json:"ok"`
	Severity      uint8         `json:"severity,omitempty"`
}
