package ingo

import (
	"bytes"
	"flag"
	"os"
	"testing"
)

const (
	testfile = `
#comment=4
asse=4
assc:4
dup=4
dup=5
s=4
obs=4
obsdup=4
obsdup=5`
	wantSavedNil = `
# shorthand test
# (longhand) (default 3)
really-long-hand=3
`
	wantSavedEmpty = ``
	wantSavedObs   = `

# The following options are probably deprecated and not used currently!
obs=4
`
)

func TestParseConfig(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	comment := flag.Int("#comment", 3, "comment test")
	asse := flag.Int("asse", 3, "assignment character test for =")
	assc := flag.Int("assc", 3, "assignment character test for :")
	dup := flag.Int("dup", 3, "duplicate entries test")
	shorthand := flag.Int("shorthand", 3, "shorthand test")
	flag.IntVar(shorthand, "s", 3, "shorthand test (shorthand)")

	obsKeys := parseConfig(bytes.NewBufferString(testfile))

	if *comment != 3 {
		t.Errorf("`#comment` flag should not be populated")
	}
	if *asse != 4 {
		t.Errorf("assignment with `=`: (want: 4; got: %d)", *asse)
	}
	if *assc != 4 {
		t.Errorf("assignment with `:`: (want: 4; got: %d)", *assc)
	}
	if *dup != 5 {
		t.Errorf("the last occuring entry of duplicate flags from the file should be used")
	}
	if *shorthand != 4 {
		t.Errorf("shorthand assignment not working")
	}
	if obsKeys["obs"] != "4" {
		t.Errorf("obsolete key not parsed")
	}
	if obsKeys["obsdup"] != "5" {
		t.Errorf("the last occuring entry of duplicate obsolete flags from the file should be used")
	}
}

func TestSaveConfig(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	resWriter := new(bytes.Buffer)
	obsKeys := make(map[string]string)
	saveConfig(resWriter, obsKeys)
	got := resWriter.String()
	if got != wantSavedEmpty {
		t.Errorf("unexpected result:\nWANT:\n%s\n\nGOT:\n%s\n", wantSavedEmpty, got)
	}

	resWriter = new(bytes.Buffer)
	obsKeys["obs"] = "4"
	saveConfig(resWriter, obsKeys)
	got = resWriter.String()
	if got != wantSavedObs {
		t.Errorf("unexpected result:\nWANT:\n%s\n\nGOT:\n%s\n", wantSavedObs, got)
	}

	resWriter = new(bytes.Buffer)
	shorthand := flag.Int("shorthand", 3, "shorthand test")
	flag.IntVar(shorthand, "s", 3, "shorthand test (shorthand)")
	flag.IntVar(shorthand, "really-long-hand", 3, "shorthand test\n    \t(longhand)")
	saveConfig(resWriter, nil)
	got = resWriter.String()
	if got != wantSavedNil {
		t.Errorf("unexpected result:\nWANT:\n%s\n\nGOT:\n%s\n", wantSavedNil, got)
	}
}
