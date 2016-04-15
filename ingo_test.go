package ingo

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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

var tmpFileName = ""

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

func TestParse(t *testing.T) {
	// pipe stderr to a temporary file
	oldErr := os.Stderr
	var err error
	if os.Stderr, err = ioutil.TempFile("", "ingo_test_err"); err != nil {
		t.Errorf("failed to redirect stderr to tempfile")
	}
	defer os.Remove(os.Stderr.Name())

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	oldFlag := flag.Int("old", 3, "old flag")
	openOrCreate = func(n string, l int, p os.FileMode) (*os.File, error) {
		f, err := ioutil.TempFile("", "ingo_testrc")
		if err == nil {
			tmpFileName = f.Name()
		}
		return f, err
	}
	if err := Parse("ingo_test"); err != nil {
		t.Fatalf("unexpected error occured: %v", err)
	}
	defer os.Remove(tmpFileName)

	if *oldFlag != 3 {
		t.Errorf("oldFlag: (want: %d; got: %d)", 3, *oldFlag)
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	newFlag := flag.Int("new", 3, "new flag")
	os.Setenv("INGO_TESTRC", tmpFileName)
	openOrCreate = os.OpenFile
	if err := Parse("ingo_test"); err != nil {
		t.Fatalf("unexpected error occured: %v", err)
	}

	if *newFlag != 3 {
		t.Errorf("newFlag: (want: %d; got: %d)", 3, *newFlag)
	}

	if err := Parse("ingo_test"); err == nil || err.Error() != "flags have been parsed already" {
		t.Errorf("expected Parse() to fail with `flags already parsed` error, but got: %v", err)
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	openOrCreate = func(n string, f int, p os.FileMode) (*os.File, error) {
		return nil, fmt.Errorf("expected")
	}
	if err := Parse("ingo_test"); err == nil || !strings.HasSuffix(err.Error(), "expected") {
		t.Errorf("expected Parse() to fail with `expected` error, but got: %v", err)
	}
	os.Stderr = oldErr
}
