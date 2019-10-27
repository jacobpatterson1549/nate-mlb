package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"strings"
	"testing"
)

func TestStartupFuncs_initialCap(t *testing.T) {
	mainFlags := new(mainFlags)
	log := log.New(ioutil.Discard, "test", log.LstdFlags)
	startupFuncs := startupFuncs(mainFlags, log)
	if len(startupFuncs) != cap(startupFuncs) {
		t.Errorf("expected length of startupFuncs (%v) to equal capacity (%v) when no environment variables/flags are set", len(startupFuncs), cap(startupFuncs))
	}
}

func TestStartupFuncs_optionalFlags(t *testing.T) {
	mainFlags := new(mainFlags)
	log := log.New(ioutil.Discard, "test", log.LstdFlags)
	startupFuncsDefault := startupFuncs(mainFlags, log)
	mainFlags.playerTypesCsv = "1"
	mainFlags.adminPassword = "test_password17"
	startupFuncsAll := startupFuncs(mainFlags, log)
	if len(startupFuncsDefault)+2 != len(startupFuncsAll) {
		t.Errorf("expected certain params to be default, but did not appear to be: wanted %v, got %v", len(startupFuncsDefault)+2, len(startupFuncsAll))
	}
}

func TestInitFlags(t *testing.T) {
	programName := "TestInitFlags"
	fs, mainFlags := initFlags(programName)
	args := strings.Fields("-ap=pass123 -n=cool_app -ds=user:pass@host/db -p=8080 -pt=2,3,5,6")
	fs.Parse(args)
	mainFlagsTests := []struct {
		fieldName string
		wantValue string
		gotValue  string
	}{
		{
			fieldName: "adminPassword",
			wantValue: "pass123",
			gotValue:  mainFlags.adminPassword,
		},
		{
			fieldName: "applicationName",
			wantValue: "cool_app",
			gotValue:  mainFlags.applicationName,
		},
		{
			fieldName: "dataSourceName",
			wantValue: "user:pass@host/db",
			gotValue:  mainFlags.dataSourceName,
		},
		{
			fieldName: "port",
			wantValue: "8080",
			gotValue:  mainFlags.port,
		},
		{
			fieldName: "playerTypesCsv",
			wantValue: "2,3,5,6",
			gotValue:  mainFlags.playerTypesCsv,
		},
	}
	if len(mainFlagsTests) != fs.NFlag() {
		t.Errorf("wanted %v flags to be loaded; got %v", len(mainFlagsTests), fs.NFlag())
	}
	for _, test := range mainFlagsTests {
		if test.gotValue != test.wantValue {
			t.Errorf("wanted %v for flag %v, but got %v", test.wantValue, test.fieldName, test.gotValue)
		}
	}
}

func TestFlagUsage(t *testing.T) {
	fs := flag.NewFlagSet("TestFlagUsage", flag.PanicOnError)
	var flag1 string
	fs.StringVar(&flag1, "FLAG#1", "DEFAULT_VALUE_74", "a flag to ensure the usage and default is printed")
	fs.Usage = func() { flagUsage(fs) }
	b := new(bytes.Buffer)
	fs.SetOutput(b)
	fs.Usage()
	gotUsage := b.String()
	if len(gotUsage) == 0 {
		t.Error("expected help message to be written")
	}
	if !strings.Contains(gotUsage, "FLAG#1") {
		t.Error("flag defaults not included in help message")
	}
}
