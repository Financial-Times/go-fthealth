package fthealth

import (
	"testing"
	"time"
)

func TestHealthCheck(t *testing.T) {

	const count = 10
	var delay = time.Millisecond * 20 * count

	checks := make([]Check, count)
	for i, _ := range checks {
		checks[i].Checker = func() error {
			time.Sleep(delay)
			return nil
		}
	}

	hc := &healthCheck{"hc name", "hc desc", checks, false}

	start := time.Now()

	result := hc.health()

	for i := 0; i < count; i++ {
		if result.Checks[0].Ok != true {
			t.Error("expected ok")
		}
	}

	expDur := count * delay
	dur := time.Now().Sub(start)

	// round down to seconds
	expSeconds := expDur.Nanoseconds() / 1000000000
	actualSeconds := dur.Nanoseconds() / 1000000000
	if expSeconds != actualSeconds {
		t.Errorf("expected duration is %ds but actual was %ds \n", expSeconds, actualSeconds)
	}
}

// this test mostly exists to exercise the parallel code and make "go test -race" useful
func TestHealthCheckParallel(t *testing.T) {

	const count = 10
	var delay = time.Second * 1

	checks := make([]Check, count)
	for i, _ := range checks {
		checks[i].Checker = func() error {
			time.Sleep(delay)
			return nil
		}
	}

	hc := &healthCheck{"hc name", "hc desc", checks, true}

	start := time.Now()

	result := hc.health()

	for i := 0; i < count; i++ {
		if result.Checks[0].Ok != true {
			t.Error("expected ok")
		}
	}

	expDur := delay
	dur := time.Now().Sub(start)

	// round down to seconds
	expSeconds := expDur.Nanoseconds() / 1000000000
	actualSeconds := dur.Nanoseconds() / 1000000000
	if expSeconds != actualSeconds {
		t.Errorf("expected duration is %ds but actual was %ds \n", expSeconds, actualSeconds)
	}
}
