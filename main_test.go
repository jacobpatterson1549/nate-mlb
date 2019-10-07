package main

import (
	"errors"
	"testing"
)

func TestStartupFuncs_initialCap(t *testing.T) {
	mainVars := mainVars{}
	startupFuncs := startupFuncs(mainVars)
	if len(startupFuncs) != cap(startupFuncs) {
		t.Errorf("expected length of startupFuncs (%v) to equal capacity (%v) when no environment variables/flags are set", len(startupFuncs), cap(startupFuncs))
	}
}

func TestStartupFuncs_optionalFlags(t *testing.T) {
	mainVars := mainVars{}
	startupFuncsDefault := startupFuncs(mainVars)
	mainVars.playerTypesCsv = "1"
	mainVars.adminPassword = "test_password17"
	startupFuncsAll := startupFuncs(mainVars)
	if len(startupFuncsDefault)+2 != len(startupFuncsAll) {
		t.Errorf("expected certain params to be default, but did not appear to be: wanted %v, got %v", len(startupFuncsDefault)+2, len(startupFuncsAll))
	}
}

func TestWaitForDb_numTries(t *testing.T) {
	waitForDbTests := []struct {
		successfulConnectTry int
		numFibonacciTries    int
		wantError            bool
	}{
		{ // should not fail when not attempted to connect
		},
		{
			successfulConnectTry: 1,
			numFibonacciTries:    1,
		},
		{
			successfulConnectTry: 2,
			numFibonacciTries:    3,
		},
		{
			successfulConnectTry: 4,
			numFibonacciTries:    3,
			wantError:            true,
		},
	}
	for i, test := range waitForDbTests {
		dbCheckCount := 0
		dbCheckFunc := func() error {
			dbCheckCount++
			if dbCheckCount != test.successfulConnectTry {
				return errors.New("check failed")
			}
			return nil
		}
		sleepFunc := func(waitTime int) {}
		err := waitForDb(dbCheckFunc, sleepFunc, test.numFibonacciTries)
		gotError := err != nil
		if test.wantError != gotError {
			t.Errorf("Test %v: wantedError = %v, gotError = %v", i, test.wantError, gotError)
		}
	}
}

func TestWaitForDb_fibonacci(t *testing.T) {
	wantFibonacciSleepSeconds := []int{0, 1, 1, 2, 3, 5, 8}
	dbCheckCount := 0
	dbCheckFunc := func() error {
		dbCheckCount++
		return errors.New("check failed")
	}
	i := 0
	sleepFunc := func(sleepSeconds int) {
		if wantFibonacciSleepSeconds[i] != sleepSeconds {
			t.Errorf("unexpected %vth wait time: wanted %v, got %v", i, wantFibonacciSleepSeconds[i], sleepSeconds)
		}
		i++
	}
	numFibonacciTries := len(wantFibonacciSleepSeconds)
	err := waitForDb(dbCheckFunc, sleepFunc, numFibonacciTries)
	if err == nil {
		t.Error("expected db wait check to error out")
	}
	if numFibonacciTries != i {
		t.Errorf("expected to wait for db to start %v times, got %v", numFibonacciTries, i)
	}
	if numFibonacciTries != dbCheckCount {
		t.Errorf("expected to check the db %v times, got %v", numFibonacciTries, dbCheckCount)
	}
}
