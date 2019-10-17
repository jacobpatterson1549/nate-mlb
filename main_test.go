package main

import (
	"testing"
)

func TestStartupFuncs_initialCap(t *testing.T) {
	mainFlags := mainFlags{}
	startupFuncs := startupFuncs(mainFlags)
	if len(startupFuncs) != cap(startupFuncs) {
		t.Errorf("expected length of startupFuncs (%v) to equal capacity (%v) when no environment variables/flags are set", len(startupFuncs), cap(startupFuncs))
	}
}

func TestStartupFuncs_optionalFlags(t *testing.T) {
	mainFlags := mainFlags{}
	startupFuncsDefault := startupFuncs(mainFlags)
	mainFlags.playerTypesCsv = "1"
	mainFlags.adminPassword = "test_password17"
	startupFuncsAll := startupFuncs(mainFlags)
	if len(startupFuncsDefault)+2 != len(startupFuncsAll) {
		t.Errorf("expected certain params to be default, but did not appear to be: wanted %v, got %v", len(startupFuncsDefault)+2, len(startupFuncsAll))
	}
}
