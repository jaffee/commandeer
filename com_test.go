package commandeer

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jaffee/commandeer/test"
	"github.com/spf13/pflag"
)

func TestLoadEnv(t *testing.T) {
	fs := &flagSet{flag.NewFlagSet("", flag.ExitOnError)}

	// create initial instance with defaults
	mm := &test.SimpleMain{
		One:    "one",
		Two:    2,
		Three:  3,
		Four:   true,
		Five:   5,
		Six:    6,
		Seven:  7.0,
		Eight:  time.Millisecond * 8,
		Twelve: "blah",
	}

	// create flags with defaults by walking instance
	err := Flags(fs, mm)
	if err != nil {
		t.Fatalf("making Flags: %v", err)
	}

	// set up environment
	prefix := "COMMANDEER_"
	err = os.Setenv("COMMANDEER_ONE", "z")
	if err != nil {
		t.Fatalf("setting up environment for set: %v", err)
	}
	defer os.Unsetenv("COMMANDEER_ONE")

	// change values on instance by reading environment
	err = loadEnv(fs, prefix)
	if err != nil {
		t.Fatalf("loading env: %v", err)
	}

	if mm.One != "z" {
		t.Errorf("unexpected value for One: %s", mm.One)
	}

	// change values on instance by parsing command line
	err = fs.Parse([]string{"-two", "99", "-twelve", "haha"})
	if err != nil {
		t.Fatalf("parsing command line: %v", err)
	}

	if mm.Two != 99 {
		t.Fatalf("command line parsing didn't produce expected value 99, got %d", mm.Two)
	}
	// ensure that parsing command line doesn't vaules except one's that are specified.
	if mm.One != "z" {
		t.Errorf("unexpected value for One after command line parsing: %s", mm.One)
	}
	if mm.Twelve != "haha" {
		t.Errorf("unexpected value for Twelve after cmd line parsing: %s", mm.Twelve)
	}

	// simulate parsing a config file and setting values directly from it
	mm.Three = 33
	mm.Five = 55

	// reload environment
	err = os.Setenv("COMMANDEER_FIVE", "56")
	if err != nil {
		t.Fatalf("setting up environment for set: %v", err)
	}
	defer os.Unsetenv("COMMANDEER_FIVE")

	err = os.Setenv("COMMANDEER_TWO", "23")
	if err != nil {
		t.Fatalf("setting up environment for set: %v", err)
	}
	defer os.Unsetenv("COMMANDEER_TWO")

	err = loadEnv(fs, prefix)
	if err != nil {
		t.Fatalf("loading env: %v", err)
	}

	if mm.Five != 56 {
		t.Fatalf("env reload failed, five is %d", mm.Five)
	}
	if mm.Two != 23 {
		t.Fatalf("env reload should have clobbered two, but mm.Two is %d", mm.Two)
	}

	// re parse command line since env clobbered two value
	err = fs.Parse([]string{"-two", "99"})
	if err != nil {
		t.Fatalf("parsing command line: %v", err)
	}

	if mm.Two != 99 {
		t.Fatalf("command line reparsing didn't produce expected value 99, got %d", mm.Two)
	}

	// ensure that parsing command line doesn't vaules except one's that are specified.
	if mm.One != "z" {
		t.Errorf("unexpected value for One after command line parsing: %s", mm.One)
	}
}

func TestLoadArgsEnv(t *testing.T) {
	mm := test.NewSimpleMain()
	mustSetenv(t, "COMMANDEER_ONE", "envone")
	mustSetenv(t, "COMMANDEER_TWO", "32")

	fs := &flagSet{flag.NewFlagSet("", flag.ExitOnError)}
	err := LoadArgsEnv(fs, mm, []string{"-two=24", "-seven", "7.3", "-nine", "a,b,c"}, "COMMANDEER_", nil)
	if err != nil {
		t.Fatalf("LoadArgsEnv: %v", err)
	}

	if mm.One != "envone" {
		t.Errorf("unexpected value for One: %s", mm.One)
	}
	if mm.Two != 24 {
		t.Errorf("unexpected value for Two: %d", mm.Two)
	}

	if mm.Seven != 7.3 {
		t.Errorf("unexpected value for Seven: %f", mm.Seven)
	}

	if !reflect.DeepEqual(mm.Nine, []string{"a", "b", "c"}) {
		t.Errorf("wrong value for Nine: %v", mm.Nine)
	}
}

func mustSetenv(t *testing.T, key, val string) {
	err := os.Setenv(key, val)
	if err != nil {
		t.Fatalf("setting env '%s' to '%s': %v", key, val, err)
	}
}

func TestZeroStruct(t *testing.T) {
	fs := pflag.NewFlagSet("myset", pflag.ContinueOnError)
	mm := &test.MyMain{}
	err := Flags(fs, mm)
	if err != nil {
		t.Fatalf("setting flags with zero MyMain: %v", err)
	}
	err = fs.Parse([]string{"-h"})
	if err != nil && err != pflag.ErrHelp {
		t.Fatalf("parsing help flag: %v", err)
	}
}

func TestZeroStructStdlibFlag(t *testing.T) {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	mm := &test.SimpleMain{}
	err := Flags(fs, mm)
	if err != nil {
		t.Fatalf("setting flags with zero MyMain: %v", err)
	}

	err = fs.Parse([]string{"-h"})
	if err != nil && err != flag.ErrHelp {
		t.Fatalf("parsing help flag: %v", err)
	}
}

func TestNonStruct(t *testing.T) {
	var a int = 4
	err := Run(&a)
	if !strings.Contains(err.Error(), "value must be pointer to struct, but is pointer to") {
		t.Fatalf("pointer to int should have failed with different err: %v", err)
	}

}

func TestDowncaseAndDash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "a",
			expected: "a",
		},
		{
			input:    "A",
			expected: "a",
		},
		{
			input:    "aB",
			expected: "a-b",
		},
		{
			input:    "Ba",
			expected: "ba",
		},
		{
			input:    "HelloFriend",
			expected: "hello-friend",
		},
		{
			input:    "helloFriend",
			expected: "hello-friend",
		},
		{
			input:    "AAA",
			expected: "aaa",
		},
		{
			input:    "myURL",
			expected: "my-url",
		},
		{
			input:    "MyURL",
			expected: "my-url",
		},
		{
			input:    "URLFinder",
			expected: "url-finder",
		},
		{
			input:    "AaURLFinder",
			expected: "aa-url-finder",
		},
		{
			input:    "aURLFinder",
			expected: "a-url-finder",
		},
		{
			input:    "AURLFinder",
			expected: "aurl-finder", // NOTE: no easy way to handle this properly
		},
		{
			input:    "MyURLs",
			expected: "my-ur-ls", // NOTE: even worse
		},
		{
			input:    "KissAToad",
			expected: "kiss-a-toad",
		},
		{
			input:    "IAmAToad",
			expected: "i-am-a-toad",
		},
	}

	for i, tst := range tests {
		output := downcaseAndDash(tst.input)
		if output != tst.expected {
			t.Errorf("test: %d, '%v' is not '%v'", i, output, tst.expected)
		}
	}
}

type NonRunner struct {
	A int
}

func TestRun(t *testing.T) {
	tests := []struct {
		main interface{}
		err  string
	}{
		{
			main: &test.MyMain{},
			err:  "mymain error",
		},
		{
			main: &NonRunner{},
			err:  "called 'Run' with something which doesn't implement the 'Run() error' method.",
		},
		{
			main: test.MyMain{},
			err:  "calling Flags: value must be pointer to struct, but is struct",
		},
	}
	for i, tst := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			err := RunArgs(pflag.NewFlagSet("tstSet", pflag.ContinueOnError), tst.main, os.Args[1:])
			if err.Error() != tst.err {
				t.Fatalf("expected '%s', got '%s'", tst.err, err.Error())
			}
		})
	}
}

func TestRunMyMain(t *testing.T) {
	mm := test.NewMyMain()
	flags := pflag.NewFlagSet("tst", pflag.ContinueOnError)
	err := Flags(flags, mm)
	if err != nil {
		t.Fatalf("getting flags for MyMain: %v", err)
	}

	if f := flags.Lookup("thing"); f != nil {
		if f.DefValue != "Thing" {
			t.Fatalf("'thing' not defined properly")
		}
	} else {
		t.Fatalf("couldn't lookup 'thing'")
	}
	if f := flags.Lookup("wing"); f != nil {
		t.Fatalf("wing shouldn't be defined")
	}
	if f := flags.Lookup("bing"); f != nil {
		t.Fatalf("bing shouldn't be defined")
	}
	if f := flags.Lookup("-"); f != nil {
		t.Fatalf("- shouldn't be defined")
	}

	if f := flags.Lookup("ipmask"); f != nil {
		if f.DefValue != "01020300" {
			t.Fatalf("'ipmask' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'ipmask'")
	}
	if f := flags.Lookup("aip-net"); f != nil {
		if f.DefValue != "1.2.3.4/01020300" {
			t.Fatalf("'aip-net' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'aip-net'")
	}
	if f := flags.Lookup("aip"); f != nil {
		if f.DefValue != "1.2.3.4" {
			t.Fatalf("'aip' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'aip'")
	}
	if f := flags.Lookup("aip-slice"); f != nil {
		if f.DefValue != "[]" {
			t.Fatalf("'aip-slice' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'aip-slice'")
	}

	if f := flags.Lookup("a-bool"); f != nil {
		if f.DefValue != "false" {
			t.Fatalf("'abool' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-bool'")
	}
	if f := flags.Lookup("a-int"); f != nil {
		if f.DefValue != "1000000000" {
			t.Fatalf("'aint' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-int'")
	}
	if f := flags.Lookup("aint8"); f != nil {
		if f.DefValue != "-7" {
			t.Fatalf("'aint8' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'aint8'")
	}
	if f := flags.Lookup("anint16"); f != nil {
		if f.DefValue != "-30000" {
			t.Fatalf("'anint16' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'anint16'")
	}
	if f := flags.Lookup("aint32"); f != nil {
		if f.DefValue != "-2000000000" {
			t.Fatalf("'aint32' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'aint32'")
	}
	if f := flags.Lookup("a-int64"); f != nil {
		if f.DefValue != "-9999999999" {
			t.Fatalf("'aint64' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-int64'")
	}
	if f := flags.Lookup("a-float"); f != nil {
		if f.DefValue != "12.2" {
			t.Fatalf("'a-float' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'afloat'")
	}
	if f := flags.Lookup("afloat32"); f != nil {
		if f.DefValue != "-12.3" {
			t.Fatalf("'afloat32' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'afloat32'")
	}
	if f := flags.Lookup("a-uint"); f != nil {
		if f.DefValue != "9" {
			t.Fatalf("'a-uint' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'auint'")
	}
	if f := flags.Lookup("auint8"); f != nil {
		if f.DefValue != "10" {
			t.Fatalf("'auint8' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'auint8'")
	}
	if f := flags.Lookup("auint16"); f != nil {
		if f.DefValue != "11" {
			t.Fatalf("'auint16' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'auint16'")
	}
	if f := flags.Lookup("auint32"); f != nil {
		if f.DefValue != "12" {
			t.Fatalf("'auint32' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'auint32'")
	}
	if f := flags.Lookup("a-uint64"); f != nil {
		if f.DefValue != "13" {
			t.Fatalf("'a-uint64' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-uint64'")
	}
	if f := flags.Lookup("a-duration"); f != nil {
		if f.DefValue != "1s" {
			t.Fatalf("'a-duration' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-duration'")
	}

	if f := flags.Lookup("a-string-slice"); f != nil {
		if f.DefValue != "[hey,there]" {
			t.Fatalf("'a-string-slice' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-string-slice'")
	}
	if f := flags.Lookup("a-bool-slice"); f != nil {
		if f.DefValue != "[true,false]" {
			t.Fatalf("'a-bool-slice' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-bool-slice'")
	}
	if f := flags.Lookup("a-int-slice"); f != nil {
		if f.DefValue != "[9,-8,7]" {
			t.Fatalf("'a-int-slice' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-int-slice'")
	}
	if f := flags.Lookup("a-uint-slice"); f != nil {
		if f.DefValue != "[7,8,9]" {
			t.Fatalf("'a-uint-slice' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'a-uint-slice'")
	}

	if f := flags.Lookup("subthing.a-bool"); f != nil {
		if f.DefValue != "true" {
			t.Fatalf("'subthing.a-bool' not defined properly, got '%v'", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'subthing.a-bool'")
	}

}

func TestRunSimpleMain(t *testing.T) {
	mm := test.NewSimpleMain()
	flags := flag.NewFlagSet("tstsimplemain", flag.ContinueOnError)
	err := Flags(flags, mm)
	if err != nil {
		t.Fatalf("getting flags for MyMain: %v", err)
	}

	if f := flags.Lookup("one"); f != nil {
		if f.DefValue != "one" {
			t.Fatalf("wrong default value for 'one': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'one'")
	}
	if f := flags.Lookup("two"); f != nil {
		if f.DefValue != "2" {
			t.Fatalf("wrong default value for 'two': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'two'")
	}
	if f := flags.Lookup("three"); f != nil {
		if f.DefValue != "3" {
			t.Fatalf("wrong default value for 'three': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'three'")
	}
	if f := flags.Lookup("four"); f != nil {
		if f.DefValue != "true" {
			t.Fatalf("wrong default value for 'four': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'four'")
	}
	if f := flags.Lookup("five"); f != nil {
		if f.DefValue != "5" {
			t.Fatalf("wrong default value for 'five': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'five'")
	}
	if f := flags.Lookup("six"); f != nil {
		if f.DefValue != "6" {
			t.Fatalf("wrong default value for 'six': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'six'")
	}
	if f := flags.Lookup("seven"); f != nil {
		if f.DefValue != "7" {
			t.Fatalf("wrong default value for 'seven': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'seven'")
	}
	if f := flags.Lookup("eight"); f != nil {
		if f.DefValue != "8ms" {
			t.Fatalf("wrong default value for 'eight': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't lookup 'eight'")
	}
	if f := flags.Lookup("nine"); f != nil {
		if f.DefValue != "[9,nine]" {
			t.Fatalf("wrong default value for 'nine': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't look up 'nine'")
	}
	if f := flags.Lookup("ten"); f != nil {
		if f.DefValue != "1970-01-01T00:00:00Z" {
			t.Fatalf("wrong default value for 'ten': %v", f.DefValue)
		}
	} else {
		t.Fatalf("couldn't look up 'ten'")
	}
	if f := flags.Lookup("eleven"); f != nil {
		if f.DefValue != "11s" {
			t.Fatalf("wrong default value for 'eleven': %v", f.DefValue)
		}
	}
	if f := flags.Lookup("twelve"); f != nil {
		if f.DefValue != "twelve" {
			t.Fatalf("wrong default value for 'tweleve': %v", f.DefValue)
		}
	}
}
