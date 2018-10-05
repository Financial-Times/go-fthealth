package v1_1

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
	timeout      time.Duration
	parallel     bool
	specialCheck specialCheck
	assertFunc   func(t *testing.T, hc HC, el testCase) (result HealthResult)
}

type specialCheck struct {
	wanted bool
	check  Check
}

func createHealthCheck(count int, checkDuration, timeout time.Duration, specialCheck specialCheck, parallel bool) HC {
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

	hc := HealthCheck{"up-mam", "Methode Article Mapper", "This mapps methode articles to internal UPP format.", checks}
	if parallel {
		if timeout != time.Duration(0) {
			return TimedHealthCheck{HealthCheck: hc, Timeout: timeout}
		}
		return hc
	} else {
		return HealthCheckSerial{hc}
	}
}

func createFeedbackHealthChecks(count int, checkDuration, timeout time.Duration, specialCheck specialCheck, parallel bool, fb chan bool) HC {
	hc := createHealthCheck(count, checkDuration, timeout, specialCheck, parallel)
	return FeedbackHealthCheck{hc, fb}
}

func verifyChecksAreOK(result HealthResult, tcName string, t *testing.T) {
	for _, check := range result.Checks {
		if check.Ok != true {
			t.Errorf("TC name: %s, Error was: one check was not OK!", tcName)
		}
	}
}

func verifyChecksAreNOTOk(result HealthResult, testCaseName string, t *testing.T) {
	for _, check := range result.Checks {
		if check.Ok != false {
			t.Errorf("Testcase: %s, Error was: one check was OK!", testCaseName)
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
		t.Errorf("TC name: %s, Error was: expected overall status %t but actual was %t \n", tcName, true, result.Ok)
	}
	if result.Severity != expectedOverallSeverity {
		t.Errorf("TC name: %s, Error was: expected overall severity %d but actual was %d \n", tcName, expectedOverallSeverity, result.Severity)
	}
}

func TestHealthCheckSequentialAndParallelAndTimed(t *testing.T) {
	testCases := [...]testCase{
		{name: "Happy flow, sequential checks", count: 10, delay: time.Millisecond * 200, parallel: false, specialCheck: specialCheck{}},
		{name: "Happy flow, parallel checks", count: 10, delay: time.Second * 1, parallel: true, specialCheck: specialCheck{}},
		{name: "Happy flow, parallel checks, timed", count: 3, delay: time.Millisecond * 200, timeout: 1 * time.Second, parallel: true, specialCheck: specialCheck{}},
	}

	for _, el := range testCases {
		hc := createHealthCheck(el.count, el.delay, el.timeout, el.specialCheck, el.parallel)
		assertSequentialAndParallelAndTimed(t, hc, el)
	}
}

func assertSequentialAndParallelAndTimed(t *testing.T, hc HC, el testCase) (result HealthResult) {
	start := time.Now()
	result = RunCheck(hc)
	actualDur := time.Now().Sub(start)

	expDur := time.Duration(el.count) * el.delay
	if el.parallel {
		expDur = el.delay
	}

	verifyChecksAreOK(result, el.name, t)
	verifyTimePassedOK(expDur, actualDur, el.name, t)
	return result
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
		{name: "Overall status and severity, happy flow, parallel and timed", count: 3, delay: time.Millisecond * 1, parallel: true, timeout: 3 * time.Second,
			specialCheck: specialCheck{true, Check{Severity: 2, Checker: func() (string, error) {
				time.Sleep(time.Millisecond * 1)
				return "", errors.New("Failure")
			}}}},
		{name: "Overall status and severity, with check error, parallel and timed", count: 3, delay: time.Millisecond * 1, parallel: true, timeout: 3 * time.Second, specialCheck: specialCheck{}},
		{name: "Checker throws a panic, parallel", count: 1, parallel: true, timeout: 3 * time.Second,
			specialCheck: specialCheck{true, Check{Severity: 2, Checker: func() (string, error) {
				panic("Checker did something unexpected")
			}}}},
		{name: "Checker throws a panic, sequential", count: 1, parallel: false, timeout: 3 * time.Second,
			specialCheck: specialCheck{true, Check{Severity: 2, Checker: func() (string, error) {
				panic("Checker did something unexpected")
			}}}},
	}

	for _, el := range testCases {
		hc := createHealthCheck(el.count, el.delay, el.timeout, el.specialCheck, el.parallel)
		assertHealthCheckStatusAndSeverityForSequentialAndParallel(t, hc, el)
	}
}

func assertHealthCheckStatusAndSeverityForSequentialAndParallel(t *testing.T, hc HC, el testCase) (result HealthResult) {
	result = RunCheck(hc)
	verifyResultOK(result, el.specialCheck.check.Severity, el.name, t)
	return result
}

func TestTimedHealthCheck(t *testing.T) {
	testCases := [...]testCase{
		{name: "Happy flow, parallel checks, timed", count: 3, delay: time.Millisecond * 500, timeout: 100 * time.Millisecond, parallel: true, specialCheck: specialCheck{}},
	}

	for _, el := range testCases {
		hc := createHealthCheck(el.count, el.delay, el.timeout, el.specialCheck, el.parallel)
		assertTimedHealthCheck(t, hc, el)
	}
}

func TestFeedbackHealthCheck(t *testing.T) {
	testCases := [...]testCase{
		{name: "Happy flow, parallel checks, timed", count: 3, delay: time.Millisecond * 500, timeout: 100 * time.Millisecond, parallel: true, specialCheck: specialCheck{}, assertFunc: assertTimedHealthCheck},
		{name: "Overall status and severity, happy flow, sequential", count: 3, delay: time.Millisecond * 1, parallel: false, assertFunc: assertHealthCheckStatusAndSeverityForSequentialAndParallel,
			specialCheck: specialCheck{true, Check{Severity: 2, Checker: func() (string, error) {
				time.Sleep(time.Millisecond * 1)
				return "", errors.New("Failure")
			}}}},
		{name: "Happy flow, sequential checks", count: 10, delay: time.Millisecond * 200, parallel: false, specialCheck: specialCheck{}, assertFunc: assertSequentialAndParallelAndTimed},
		{name: "Happy flow, parallel checks", count: 10, delay: time.Second * 1, parallel: true, specialCheck: specialCheck{}, assertFunc: assertSequentialAndParallelAndTimed},
		{name: "Happy flow, parallel checks, timed", count: 3, delay: time.Millisecond * 200, timeout: 1 * time.Second, parallel: true, specialCheck: specialCheck{}, assertFunc: assertSequentialAndParallelAndTimed},
	}
	for _, el := range testCases {
		fb := make(chan bool, 1)
		hc := createFeedbackHealthChecks(el.count, el.delay, el.timeout, el.specialCheck, el.parallel, fb)
		result := el.assertFunc(t, hc, el)
		if len(fb) != 1 {
			close(fb)
			t.Errorf("Expected 1 item on queue, got %d\n", len(fb))
		}

		st := <-fb
		close(fb)
		if st != result.Ok {
			t.Errorf("Status was %t and result.ok was %t \n", st, result.Ok)
		}
	}
}

func assertTimedHealthCheck(t *testing.T, hc HC, el testCase) (result HealthResult) {
	start := time.Now()
	result = RunCheck(hc)
	actualDur := time.Now().Sub(start)

	verifyChecksAreNOTOk(result, el.name, t)
	verifyTimePassedOK(el.timeout, actualDur, el.name, t)
	return result
}
