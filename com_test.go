package commandeer

import (
	"fmt"
	"os"
	"testing"

	"github.com/jaffee/commandeer/test"
	"github.com/spf13/pflag"
)

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
