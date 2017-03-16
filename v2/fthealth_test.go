package v2

import (
	"errors"
	"testing"
	"time"
	"math/rand"
)

func createHealthCheck(count int, checkDuration time.Duration, parallel bool, checkWithError bool, checkWithErrorSeverity int) *HealthCheck {
	checks := make([]Check, count)
	for i := range checks {
		checks[i].Checker = func() (string, error) {
			time.Sleep(checkDuration)
			return "", nil
		}
		checks[i].Severity = uint8((i % 3) + 1);
	}

	if (checkWithError) {
		randomIndex := rand.Intn(count);
		checks[randomIndex].Checker = func() (string, error) {
			time.Sleep(checkDuration)
			return "", errors.New("Failure");
		}
		checks[randomIndex].Severity = uint8(checkWithErrorSeverity);
	}

	return &HealthCheck{"up-mam", "Methode Article Mapper", "This mapps methode articles to internal UPP format.", checks, parallel}
}

func verifyChecksAreOK(result HealthResult, t *testing.T) {
	for _, check := range result.Checks {
		if check.Ok != true {
			t.Error("Check was not OK!")
		}
	}
}

func verifyTimePassedOK(expDur time.Duration, actualDur time.Duration, t *testing.T) {
	expSec := expDur.Nanoseconds() / 1000000000
	actualSec := actualDur.Nanoseconds() / 1000000000
	if expSec != actualSec {
		t.Errorf("expected duration is %ds but actual was %ds \n", expSec, actualSec)
	}
}

func verifyResultOK(result HealthResult, expectedOverallSeverity int, t *testing.T) {
	expectedOK := expectedOverallSeverity == 0;
	if result.Ok != expectedOK {
		t.Errorf("expected overall status %b but actual was %b \n", true, result.Ok)
	}
	if result.Severity != uint8(expectedOverallSeverity) {
		t.Errorf("expected overall severity %d but actual was %d \n", expectedOverallSeverity, result.Severity)
	}
}

func TestHealthCheckSequential(t *testing.T) {
	const count = 10
	delay := time.Millisecond * 20 * count

	hc := createHealthCheck(count, delay, false, false, 0);

	start := time.Now()
	result := hc.health()

	verifyChecksAreOK(result, t);

	expDur := time.Duration(count * delay)
	actualDur := time.Now().Sub(start)

	verifyTimePassedOK(expDur, actualDur, t);
}

func TestHealthCheckParallel(t *testing.T) {
	const count = 10
	delay := time.Second * 1

	hc := createHealthCheck(count, delay, true, false, 0);

	start := time.Now()
	result := hc.health()

	verifyChecksAreOK(result, t);

	expDur := delay
	actualDur := time.Now().Sub(start)

	verifyTimePassedOK(expDur, actualDur, t);
}

func TestNonHealthyCheckForOverallStatusAndSeverityForSequential(t *testing.T) {
	const count = 3
	delay := time.Millisecond * 1
	checkErrorSeverity := 2;

	hc := createHealthCheck(count, delay, false, true, checkErrorSeverity);

	result := hc.health()

	verifyResultOK(result, checkErrorSeverity, t);
}

func TestNonHealthyCheckForOverallStatusAndSeverityForParallel(t *testing.T) {
	const count = 3
	delay := time.Millisecond * 1
	checkErrorSeverity := 2;

	hc := createHealthCheck(count, delay, true, true, checkErrorSeverity);

	result := hc.health()

	verifyResultOK(result, checkErrorSeverity, t);
}

func TestHealthyCheckForOverallStatusAndSeverityForSequential(t *testing.T) {
	const count = 3
	delay := time.Millisecond * 1
	checkErrorSeverity := 0;

	hc := createHealthCheck(count, delay, false, false, checkErrorSeverity);

	result := hc.health()

	verifyResultOK(result, checkErrorSeverity, t);
}

func TestHealthyCheckForOverallStatusAndSeverityForParallel(t *testing.T) {
	const count = 3
	delay := time.Millisecond * 1
	checkErrorSeverity := 0;

	hc := createHealthCheck(count, delay, true, false, checkErrorSeverity);

	result := hc.health()

	verifyResultOK(result, checkErrorSeverity, t);
}
