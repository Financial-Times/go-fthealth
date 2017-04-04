package fthealth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
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

func TestNonHealthyCheckForOverallStatusAndSeverityForSequential(t *testing.T) {

	const count = 3

	checks := make([]Check, count)
	checks[0].Checker = func() error {
		return nil
	}
	checks[0].Severity = 3

	checks[1].Checker = func() error {
		return errors.New("Failure")
	}
	checks[1].Severity = 2

	checks[2].Checker = func() error {
		return nil
	}
	checks[2].Severity = 1
	hc := &healthCheck{"hc name", "hc desc", checks, false}

	result := hc.health()

	if result.Ok != false {
		t.Errorf("expected overall status %b but actual was %b \n", false, result.Ok)
	}
	if result.Severity != 2 {
		t.Errorf("expected overall severity %d but actual was %d \n", 2, result.Severity)
	}
}

func TestNonHealthyCheckForOverallStatusAndSeverityForParallel(t *testing.T) {

	const count = 3

	checks := make([]Check, count)
	checks[0].Checker = func() error {
		return nil
	}
	checks[0].Severity = 3

	checks[1].Checker = func() error {
		return errors.New("Failure")
	}
	checks[1].Severity = 2

	checks[2].Checker = func() error {
		return nil
	}
	checks[2].Severity = 1
	hc := &healthCheck{"hc name", "hc desc", checks, true}

	result := hc.health()

	if result.Ok != false {
		t.Errorf("expected overall status %b but actual was %b \n", false, result.Ok)
	}
	if result.Severity != 2 {
		t.Errorf("expected overall severity %d but actual was %d \n", 2, result.Severity)
	}
}

func TestHealthyCheckForOverallStatusAndSeverityForSequential(t *testing.T) {

	const count = 3

	checks := make([]Check, count)
	checks[0].Checker = func() error {
		return nil
	}
	checks[0].Severity = 3

	checks[1].Checker = func() error {
		return nil
	}
	checks[1].Severity = 2

	checks[2].Checker = func() error {
		return nil
	}
	checks[2].Severity = 1
	hc := &healthCheck{"hc name", "hc desc", checks, false}

	result := hc.health()

	if result.Ok != true {
		t.Errorf("expected overall status %b but actual was %b \n", true, result.Ok)
	}
	if result.Severity != 0 {
		t.Errorf("expected overall severity %d but actual was %d \n", 0, result.Severity)
	}
}

func TestHealthyCheckForOverallStatusAndSeverityForParallel(t *testing.T) {

	const count = 3

	checks := make([]Check, count)
	checks[0].Checker = func() error {
		return nil
	}
	checks[0].Severity = 3

	checks[1].Checker = func() error {
		return nil
	}
	checks[1].Severity = 2

	checks[2].Checker = func() error {
		return nil
	}
	checks[2].Severity = 1
	hc := &healthCheck{"hc name", "hc desc", checks, true}

	result := hc.health()

	if result.Ok != true {
		t.Errorf("expected overall status %b but actual was %b \n", true, result.Ok)
	}
	if result.Severity != 0 {
		t.Errorf("expected overall severity %d but actual was %d \n", 0, result.Severity)
	}
}

func TestCoalesce(t *testing.T) {

	delta := make(chan int)

	most := make(chan int)

	go func() {
		var running int
		var mostRunning int
		for i := range delta {
			running += i
			if running > mostRunning {
				mostRunning = running
			}
		}
		if running != 0 {
			panic("bug in test")
		}
		most <- mostRunning
	}()

	slowCheck := Check{
		BusinessImpact:   "blah",
		Name:             "My check",
		PanicGuide:       "Don't panic",
		Severity:         1,
		TechnicalSummary: "Something technical",
		Checker: func() error {
			delta <- 1
			time.Sleep(100 * time.Millisecond)
			delta <- -1
			return nil
		},
	}

	h := http.HandlerFunc(Handler("name", "desc", slowCheck))

	ts := httptest.NewServer(h)
	defer ts.Close()

	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			_, err := http.Get(ts.URL)
			if err != nil {
				t.Error(err)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(delta)

	m := <-most
	if m != 1 {
		t.Errorf("concurrency exceeded 1 : %d", m)
	}
}
