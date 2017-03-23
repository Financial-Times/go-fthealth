package v2

import (
	"errors"
	"math/rand"
	"testing"
	"time"
)

type testCase struct {
	name         string
	count        int
	delay        time.Duration
	parallel     bool
	specialCheck specialCheck
}

type specialCheck struct {
	wanted bool
	check  Check
}

func createHealthCheck(count int, checkDuration time.Duration, specialCheck specialCheck, parallel bool) HC {
	checks := make([]Check, count)
	for i := range checks {
		checks[i].Checker = func() (string, error) {
			time.Sleep(checkDuration)
			return "", nil
		}
		checks[i].Severity = uint8((i % 3) + 1)
	}

	if specialCheck.wanted {
		randomIndex := rand.Intn(count)
		checks[randomIndex] = specialCheck.check
	}

	if parallel {
		return HealthCheck{"up-mam", "Methode Article Mapper", "This mapps methode articles to internal UPP format.", checks}
	} else {
		return HealthCheckSerial{HealthCheck{"up-mam", "Methode Article Mapper", "This mapps methode articles to internal UPP format.", checks}}
	}
}

func verifyChecksAreOK(result HealthResult, tcName string, t *testing.T) {
	for _, check := range result.Checks {
		if check.Ok != true {
			t.Errorf("TC name: %s, Error was: one check was not OK!", tcName)
		}
	}
}

func verifyTimePassedOK(expDur time.Duration, actualDur time.Duration, tcName string, t *testing.T) {
	expSec := expDur.Nanoseconds() / 1000000000
	actualSec := actualDur.Nanoseconds() / 1000000000
	if expSec != actualSec {
		t.Errorf("TC name: %s, Error was: expected duration is %ds but actual was %ds \n", tcName, expSec, actualSec)
	}
}

func verifyResultOK(result HealthResult, expectedOverallSeverity uint8, tcName string, t *testing.T) {
	expectedOK := expectedOverallSeverity == 0
	if result.Ok != expectedOK {
		t.Errorf("TC name: %s, Error was: expected overall status %b but actual was %b \n", tcName, true, result.Ok)
	}
	if result.Severity != expectedOverallSeverity {
		t.Errorf("TC name: %s, Error was: expected overall severity %d but actual was %d \n", tcName, expectedOverallSeverity, result.Severity)
	}
}

func TestHealthCheckSequentialAndParallel(t *testing.T) {
	testCases := [...]testCase{
		{name: "Happy flow, sequential checks", count: 10, delay: time.Millisecond * 200, parallel: false, specialCheck: specialCheck{}},
		{name: "Happy flow, parallel checks", count: 10, delay: time.Second * 1, parallel: true, specialCheck: specialCheck{}},
	}

	for _, el := range testCases {
		hc := createHealthCheck(el.count, el.delay, el.specialCheck, el.parallel)

		start := time.Now()
		result := RunCheck(hc)
		actualDur := time.Now().Sub(start)

		expDur := 10 * el.delay
		if el.parallel {
			expDur = el.delay
		}

		verifyChecksAreOK(result, el.name, t)
		verifyTimePassedOK(expDur, actualDur, el.name, t)
	}
}

func TestResultStatusAndSeverityForSequentialAndParallel(t *testing.T) {
	testCases := [...]testCase{
		{name: "Overall status and severity, happy flow, sequential", count: 3, delay: time.Millisecond * 1, parallel: false,
			specialCheck: specialCheck{true, Check{Severity: 2, Checker: func() (string, error) {
				time.Sleep(time.Millisecond * 1)
				return "", errors.New("Failure")
			}}}},
		{name: "Overall status and severity, happy flow, parallel", count: 3, delay: time.Millisecond * 1, parallel: true,
			specialCheck: specialCheck{true, Check{Severity: 2, Checker: func() (string, error) {
				time.Sleep(time.Millisecond * 1)
				return "", errors.New("Failure")
			}}}},
		{name: "Overall status and severity, with check error, sequential", count: 3, delay: time.Millisecond * 1, parallel: false, specialCheck: specialCheck{}},
		{name: "Overall status and severity, with check error, parallel", count: 3, delay: time.Millisecond * 1, parallel: true, specialCheck: specialCheck{}},
	}

	for _, el := range testCases {
		hc := createHealthCheck(el.count, el.delay, el.specialCheck, el.parallel)
		result := RunCheck(hc)
		verifyResultOK(result, el.specialCheck.check.Severity, el.name, t)
	}
}
